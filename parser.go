package stage

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// actionHeaderPattern matches a "[Begin Action N]" line and captures the
// action number -- same shape character/air's own actionHeaderPattern
// uses, since the underlying file syntax is identical (this repo cannot
// depend on character, so the pattern is duplicated rather than shared).
var actionHeaderPattern = regexp.MustCompile(`(?i)^\[\s*begin\s+action\s+(-?\d+)\s*\]`)

// actionHeaderAttemptPattern recognizes any bracket line that starts with
// the "begin" keyword, whether or not it goes on to match
// actionHeaderPattern -- used to tell a malformed "[Begin Action N]"
// header apart from a genuinely unrelated bracket section.
var actionHeaderAttemptPattern = regexp.MustCompile(`(?i)^\[\s*begin(\s|\]|$)`)

// Parse reads MUGEN/Ikemen GO stage .def text from r and returns the Stage
// it describes.
//
// Only sections this data model has a place for are recognized:
// "[Info]" (Stage.Name/Author), "[BGDef]" (BGdef.SpriteFile plus, for a
// model-based stage, its ModelFile), "[StageInfo]" (BGdef's local
// coordinate space and ground level), "[Camera]" (CameraBounds plus
// BGdef's zoom range and, for a model-based stage, its Near/Far/FOV/
// YShift), "[PlayerInfo]" (StageBoundaries, plus its z-axis extension and
// PlayerStartZ for a model-based stage), "[Model]" and "[Scaling]"
// (Ikemen GO's 3D stage extension, see the roadmap's .vibe/decisions/014),
// one "[BG <name>]" section per background element (matched
// case-insensitively), and one "[Begin Action N]" section per animated BG
// element's frame sequence (Stage.Animations, keyed by action number,
// resolved independently of element order -- see .vibe/decisions/006 and
// item 009). Any other section — including "[Bound]",
// "[Shadow]", "[Reflection]", "[Music]" — carries nothing this model
// represents, so its lines are skipped without validation, the same way
// def.Parse/cns.Parse in the character repo skip unrecognized sections
// rather than aborting the read. Within a recognized section, an
// unrecognized key is likewise ignored, and a content line that isn't a
// valid "key = value" pair is ignored rather than erroring — real
// MUGEN/Ikemen engines tolerate both. Comment lines (';', whole-line or
// trailing) are stripped before parsing, and values may optionally be
// wrapped in double quotes, which are removed.
//
// A "[BG <name>]" section missing its "type" key defaults that element to
// BGElementNormal, matching MUGEN's own behavior for a missing type.
//
// An empty input returns a zero-value Stage and a nil error. A malformed
// section header (missing closing bracket) still returns a descriptive
// error naming the offending line, as does a reader that fails outright,
// a key requiring a number whose value isn't one, or a key requiring a
// comma-separated pair whose value isn't a valid pair.
func Parse(r io.Reader) (Stage, error) {
	scanner := bufio.NewScanner(r)

	var stage Stage
	var elements []BGElement
	var current *BGElement
	var currentSection string
	var currentAnimation *BGAnimation
	var currentActionNumber int

	flushCurrent := func() {
		if current != nil {
			elements = append(elements, *current)
			current = nil
		}
	}
	flushCurrentAnimation := func() {
		if currentAnimation != nil {
			if stage.Animations == nil {
				stage.Animations = make(map[int]BGAnimation)
			}
			stage.Animations[currentActionNumber] = *currentAnimation
			currentAnimation = nil
		}
	}

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := stripStageComment(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return Stage{}, fmt.Errorf("stage: line %d: malformed section header %q", lineNumber, line)
			}
			flushCurrent()
			flushCurrentAnimation()

			raw := strings.TrimSpace(line[1 : len(line)-1])
			if name, ok := bgElementName(raw); ok {
				current = &BGElement{Name: name}
				currentSection = "bg"
				continue
			}
			if m := actionHeaderPattern.FindStringSubmatch(line); m != nil {
				number, err := strconv.Atoi(m[1])
				if err != nil {
					return Stage{}, fmt.Errorf("stage: line %d: invalid action number %q: %w", lineNumber, m[1], err)
				}
				currentAnimation = &BGAnimation{}
				currentActionNumber = number
				currentSection = "beginaction"
				continue
			}
			if actionHeaderAttemptPattern.MatchString(line) {
				return Stage{}, fmt.Errorf("stage: line %d: malformed action header %q", lineNumber, line)
			}
			currentSection = strings.ToLower(raw)
			continue
		}

		if currentSection == "beginaction" {
			if currentAnimation == nil {
				continue
			}
			if strings.EqualFold(line, "loopstart") {
				currentAnimation.LoopStart = len(currentAnimation.Frames)
				continue
			}
			frame, err := parseBGAnimFrameLine(line)
			if err != nil {
				return Stage{}, fmt.Errorf("stage: line %d: %w", lineNumber, err)
			}
			currentAnimation.Frames = append(currentAnimation.Frames, frame)
			continue
		}

		key, value, ok := parseStageKeyValueLine(line)
		if !ok {
			// A content line inside a recognized section that isn't a valid
			// key=value pair (a decorative separator, a truncated leftover
			// key) is ignored, the same tolerance def.Parse/cns.Parse apply.
			continue
		}

		switch currentSection {
		case "info":
			switch strings.ToLower(key) {
			case "name":
				stage.Name = value
			case "author":
				stage.Author = value
			}
		case "bgdef":
			switch {
			case strings.EqualFold(key, "spr"):
				stage.BGdef.SpriteFile = value
			case strings.EqualFold(key, "model"):
				stage.BGdef.ModelFile = value
			}
		case "stageinfo":
			switch {
			case strings.EqualFold(key, "localcoord"):
				w, h, err := parseIntPair(value)
				if err != nil {
					return Stage{}, fmt.Errorf("stage: line %d: invalid localcoord %q: %w", lineNumber, value, err)
				}
				stage.BGdef.LocalCoordWidth, stage.BGdef.LocalCoordHeight = w, h
			case strings.EqualFold(key, "zoffset"):
				n, err := parseIntTolerant(value)
				if err != nil {
					return Stage{}, fmt.Errorf("stage: line %d: invalid zoffset %q: %w", lineNumber, value, err)
				}
				stage.BGdef.ZOffset = n
			}
		case "camera":
			if err := parseCameraKey(&stage, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		case "playerinfo":
			if err := parsePlayerInfoKey(&stage, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		case "bg":
			if current == nil {
				continue
			}
			if err := parseBGElementKey(current, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		case "model":
			if err := parseModelKey(&stage.Model, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		case "scaling":
			if err := parseScalingKey(&stage.Scaling, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		}
		// Any other, unrecognized section carries nothing this model
		// represents; its key=value lines are ignored.
	}

	flushCurrent()
	flushCurrentAnimation()

	if err := scanner.Err(); err != nil {
		return Stage{}, fmt.Errorf("stage: reading stage definition source: %w", err)
	}

	// Every BG element defaults to BGElementNormal unless its "type" key
	// said otherwise — applying that default is this parser's job, not the
	// pure-data model's (see BGElementType's own doc comment).
	for i := range elements {
		if elements[i].Type == "" {
			elements[i].Type = BGElementNormal
		}
	}
	stage.Elements = elements

	return stage, nil
}

// bgElementName reports whether raw (a section header's trimmed inner text,
// e.g. "BG floor") names a BG element section, returning the element's name
// with the "BG" keyword's original casing and surrounding whitespace
// stripped. The keyword match is case-insensitive; the name itself keeps
// whatever casing the file used.
func bgElementName(raw string) (name string, ok bool) {
	if len(raw) < 3 {
		return "", false
	}
	if !strings.EqualFold(raw[:2], "bg") {
		return "", false
	}
	if raw[2] != ' ' && raw[2] != '\t' {
		return "", false
	}
	name = strings.TrimSpace(raw[3:])
	if name == "" {
		return "", false
	}
	return name, true
}

// parseCameraKey applies a single "[Camera]" section key=value pair to
// stage's CameraBounds and BGdef.ZoomOut/ZoomIn fields.
func parseCameraKey(stage *Stage, key, value string, lineNumber int) error {
	switch {
	case strings.EqualFold(key, "boundleft"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid boundleft %q: %w", lineNumber, value, err)
		}
		stage.CameraBounds.Left = n
	case strings.EqualFold(key, "boundright"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid boundright %q: %w", lineNumber, value, err)
		}
		stage.CameraBounds.Right = n
	case strings.EqualFold(key, "boundhigh"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid boundhigh %q: %w", lineNumber, value, err)
		}
		stage.CameraBounds.High = n
	case strings.EqualFold(key, "boundlow"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid boundlow %q: %w", lineNumber, value, err)
		}
		stage.CameraBounds.Low = n
	case strings.EqualFold(key, "zoomout"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid zoomout %q: %w", lineNumber, value, err)
		}
		stage.BGdef.ZoomOut = f
	case strings.EqualFold(key, "zoomin"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid zoomin %q: %w", lineNumber, value, err)
		}
		stage.BGdef.ZoomIn = f
	case strings.EqualFold(key, "near"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid near %q: %w", lineNumber, value, err)
		}
		stage.BGdef.Near = f
	case strings.EqualFold(key, "far"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid far %q: %w", lineNumber, value, err)
		}
		stage.BGdef.Far = f
	case strings.EqualFold(key, "fov"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid fov %q: %w", lineNumber, value, err)
		}
		stage.BGdef.FOV = f
	case strings.EqualFold(key, "yshift"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid yshift %q: %w", lineNumber, value, err)
		}
		stage.BGdef.YShift = f
	}
	return nil
}

// parsePlayerInfoKey applies a single "[PlayerInfo]" section key=value pair
// to stage's StageBoundaries and PlayerStartZ fields.
func parsePlayerInfoKey(stage *Stage, key, value string, lineNumber int) error {
	switch {
	case strings.EqualFold(key, "leftbound"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid leftbound %q: %w", lineNumber, value, err)
		}
		stage.StageBoundaries.Left = n
	case strings.EqualFold(key, "rightbound"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid rightbound %q: %w", lineNumber, value, err)
		}
		stage.StageBoundaries.Right = n
	case strings.EqualFold(key, "topbound"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid topbound %q: %w", lineNumber, value, err)
		}
		stage.StageBoundaries.TopBound = f
	case strings.EqualFold(key, "botbound"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid botbound %q: %w", lineNumber, value, err)
		}
		stage.StageBoundaries.BottomBound = f
	default:
		if field := playerStartZField(&stage.PlayerStartZ, key); field != nil {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return fmt.Errorf("stage: line %d: invalid %s %q: %w", lineNumber, key, value, err)
			}
			*field = n
		}
	}
	return nil
}

// playerStartZField returns a pointer to the field of p that key names
// ("p1startz" through "p8startz"), case-insensitively, or nil if key names
// none of them.
func playerStartZField(p *PlayerStartZ, key string) *int {
	fields := [8]*int{&p.P1, &p.P2, &p.P3, &p.P4, &p.P5, &p.P6, &p.P7, &p.P8}
	for i, f := range fields {
		if strings.EqualFold(key, fmt.Sprintf("p%dstartz", i+1)) {
			return f
		}
	}
	return nil
}

// parseModelKey applies a single "[Model]" section key=value pair to m.
func parseModelKey(m *Model, key, value string, lineNumber int) error {
	switch {
	case strings.EqualFold(key, "offset"):
		x, y, z, err := parseFloatTriple(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid offset %q: %w", lineNumber, value, err)
		}
		m.OffsetX, m.OffsetY, m.OffsetZ = x, y, z
	case strings.EqualFold(key, "scale"):
		x, y, z, err := parseFloatTriple(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid scale %q: %w", lineNumber, value, err)
		}
		m.ScaleX, m.ScaleY, m.ScaleZ = x, y, z
	case strings.EqualFold(key, "environment"):
		m.Environment = value
	case strings.EqualFold(key, "environmentintensity"):
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid environmentintensity %q: %w", lineNumber, value, err)
		}
		m.EnvironmentIntensity = f
	}
	return nil
}

// parseScalingKey applies a single "[Scaling]" section key=value pair to sc.
func parseScalingKey(sc *Scaling, key, value string, lineNumber int) error {
	target := map[string]*float64{
		"depthtoscreen": &sc.DepthToScreen,
		"topz":          &sc.TopZ,
		"botz":          &sc.BottomZ,
		"topscale":      &sc.TopScale,
		"botscale":      &sc.BottomScale,
	}
	for k, field := range target {
		if strings.EqualFold(key, k) {
			f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			if err != nil {
				return fmt.Errorf("stage: line %d: invalid %s %q: %w", lineNumber, key, value, err)
			}
			*field = f
			return nil
		}
	}
	return nil
}

// parseBGElementKey applies a single "[BG <name>]" section key=value pair
// to the element currently being built.
func parseBGElementKey(el *BGElement, key, value string, lineNumber int) error {
	switch {
	case strings.EqualFold(key, "type"):
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "normal":
			el.Type = BGElementNormal
		case "parallax":
			el.Type = BGElementParallax
		case "anim":
			el.Type = BGElementAnim
		default:
			// An unrecognized type value is tolerated the same way an
			// unrecognized key is — real MUGEN/Ikemen engines fall back to
			// the normal default rather than rejecting the file.
			el.Type = BGElementNormal
		}
	case strings.EqualFold(key, "spriteno"):
		g, i, err := parseIntPair(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid spriteno %q: %w", lineNumber, value, err)
		}
		el.Sprite = SpriteRef{Group: g, Image: i}
	case strings.EqualFold(key, "actionno"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid actionno %q: %w", lineNumber, value, err)
		}
		el.ActionNumber = n
	case strings.EqualFold(key, "layerno"):
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid layerno %q: %w", lineNumber, value, err)
		}
		el.LayerNo = n
	case strings.EqualFold(key, "start"):
		x, y, err := parseIntPair(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid start %q: %w", lineNumber, value, err)
		}
		el.StartX, el.StartY = x, y
	case strings.EqualFold(key, "delta"):
		x, y, err := parseFloatPair(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid delta %q: %w", lineNumber, value, err)
		}
		el.DeltaX, el.DeltaY = x, y
	case strings.EqualFold(key, "tile"):
		x, y, err := parseIntPairOrSingle(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid tile %q: %w", lineNumber, value, err)
		}
		el.TileX, el.TileY = x, y
	case strings.EqualFold(key, "tilespacing"):
		x, y, err := parseIntPairOrSingle(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid tilespacing %q: %w", lineNumber, value, err)
		}
		el.TileSpacingX, el.TileSpacingY = x, y
	}
	return nil
}

// parseBGAnimFrameLine parses one "[Begin Action N]" frame line: the same
// underlying ".air"-syntax "group,image,x,y,time[,flip[,blend]]" shape
// character/air's own frame lines use (this repo cannot depend on
// character, so this is a small, local reimplementation covering only what
// BGAnimFrame models). x/y/flip/blend, if present, are validated as part
// of the minimum-field-count shape but not stored -- see
// .vibe/decisions/006.
func parseBGAnimFrameLine(line string) (BGAnimFrame, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		return BGAnimFrame{}, fmt.Errorf("malformed frame line %q: expected at least 5 comma-separated fields", line)
	}
	group, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil {
		return BGAnimFrame{}, fmt.Errorf("malformed frame line %q: invalid group: %w", line, err)
	}
	image, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil {
		return BGAnimFrame{}, fmt.Errorf("malformed frame line %q: invalid image: %w", line, err)
	}
	// A blank time field (backlog item 007's corpus scan found a real file,
	// "XX'GARAGE'XX", with one) defaults to 0 rather than erroring — a
	// real-world authoring slip, the same "absent numeric value reads as
	// its zero value" tolerance already applied elsewhere in this parser.
	timeStr := strings.TrimSpace(fields[4])
	var timeVal int
	if timeStr != "" {
		timeVal, err = strconv.Atoi(timeStr)
		if err != nil {
			return BGAnimFrame{}, fmt.Errorf("malformed frame line %q: invalid time: %w", line, err)
		}
	}
	return BGAnimFrame{Sprite: SpriteRef{Group: group, Image: image}, Time: timeVal}, nil
}

// parseIntPair parses a "a,b" comma-separated pair of integers.
func parseIntPair(value string) (a, b int, err error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected a \"a,b\" pair")
	}
	a, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	b, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

// parseIntTolerant parses value as an integer, falling back to parsing it
// as a float and rounding to the nearest integer when the plain-integer
// parse fails. A real-file shape (backlog item 007's corpus scan, `zoffset
// = 555.0`): some stage authors write an integer-valued field with a
// redundant decimal point. Genuinely non-numeric input still errors —
// this only accepts an otherwise-valid number, never garbage.
func parseIntTolerant(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if n, err := strconv.Atoi(trimmed); err == nil {
		return n, nil
	}
	f, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, err
	}
	return int(math.Round(f)), nil
}

// parseIntPairOrSingle parses a "a,b" comma-separated pair of integers, the
// same shape parseIntPair does — except a bare single value (no comma) is
// also accepted and applied to both a and b. A real-file shape (backlog
// item 007's corpus scan, `tile = 1` on a BG element): some stage authors
// write a single value for a symmetric property instead of the documented
// pair, matching the "a,b" pair's own convention when both axes are equal.
func parseIntPairOrSingle(value string) (a, b int, err error) {
	if !strings.Contains(value, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, 0, err
		}
		return n, n, nil
	}
	return parseIntPair(value)
}

// parseFloatPair parses a "a,b" comma-separated pair of floats.
func parseFloatPair(value string) (a, b float64, err error) {
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected a \"a,b\" pair")
	}
	a, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, err
	}
	b, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

// parseFloatTriple parses a "a,b,c" comma-separated triple of floats, used
// by [Model]'s "offset" and "scale" keys.
func parseFloatTriple(value string) (a, b, c float64, err error) {
	parts := strings.SplitN(value, ",", 3)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("expected a \"a,b,c\" triple")
	}
	a, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	c, err = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return a, b, c, nil
}

// stripStageComment removes a ".def" comment from line — everything from
// the first ';' to the end of the line, whether the comment stands on its
// own line or trails after real content — and trims surrounding whitespace
// from what remains.
func stripStageComment(line string) string {
	if idx := strings.IndexByte(line, ';'); idx != -1 {
		line = line[:idx]
	}
	return strings.TrimSpace(line)
}

// parseStageKeyValueLine splits a "key = value" line on its first '=',
// trimming whitespace from both sides and removing a matching pair of
// surrounding double quotes from the value. ok is false if line has no '='
// or an empty key.
func parseStageKeyValueLine(line string) (key, value string, ok bool) {
	idx := strings.IndexByte(line, '=')
	if idx == -1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	value = unquoteStage(strings.TrimSpace(line[idx+1:]))
	return key, value, true
}

// unquoteStage removes a matching pair of surrounding double quotes from s,
// if present.
func unquoteStage(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
