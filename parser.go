package stage

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parse reads MUGEN/Ikemen GO stage .def text from r and returns the Stage
// it describes.
//
// Only sections this data model has a place for are recognized:
// "[BGDef]" (BGdef.SpriteFile), "[StageInfo]" (BGdef's local coordinate
// space and ground level), "[Camera]" (CameraBounds plus BGdef's zoom
// range), "[PlayerInfo]" (StageBoundaries), and one "[BG <name>]" section
// per background element (matched case-insensitively). Any other section —
// including "[Info]", "[Bound]", "[Shadow]", "[Reflection]", "[Music]" —
// carries nothing this model represents, so its lines are skipped without
// validation, the same way def.Parse/cns.Parse in the character repo skip
// unrecognized sections rather than aborting the read. Within a recognized
// section, an unrecognized key is likewise ignored, and a content line that
// isn't a valid "key = value" pair is ignored rather than erroring — real
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

	flushCurrent := func() {
		if current != nil {
			elements = append(elements, *current)
			current = nil
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

			raw := strings.TrimSpace(line[1 : len(line)-1])
			if name, ok := bgElementName(raw); ok {
				current = &BGElement{Name: name}
				currentSection = "bg"
				continue
			}
			currentSection = strings.ToLower(raw)
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
		case "bgdef":
			if strings.EqualFold(key, "spr") {
				stage.BGdef.SpriteFile = value
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
				n, err := strconv.Atoi(strings.TrimSpace(value))
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
			switch {
			case strings.EqualFold(key, "leftbound"):
				n, err := strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					return Stage{}, fmt.Errorf("stage: line %d: invalid leftbound %q: %w", lineNumber, value, err)
				}
				stage.StageBoundaries.Left = n
			case strings.EqualFold(key, "rightbound"):
				n, err := strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					return Stage{}, fmt.Errorf("stage: line %d: invalid rightbound %q: %w", lineNumber, value, err)
				}
				stage.StageBoundaries.Right = n
			}
		case "bg":
			if current == nil {
				continue
			}
			if err := parseBGElementKey(current, key, value, lineNumber); err != nil {
				return Stage{}, err
			}
		}
		// Any other section — including "info" and anything unrecognized —
		// carries nothing this model represents; its key=value lines are
		// ignored.
	}

	flushCurrent()

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
		x, y, err := parseIntPair(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid tile %q: %w", lineNumber, value, err)
		}
		el.TileX, el.TileY = x, y
	case strings.EqualFold(key, "tilespacing"):
		x, y, err := parseIntPair(value)
		if err != nil {
			return fmt.Errorf("stage: line %d: invalid tilespacing %q: %w", lineNumber, value, err)
		}
		el.TileSpacingX, el.TileSpacingY = x, y
	}
	return nil
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
