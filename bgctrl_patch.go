package stage

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"
)

// bgctrlSection is the parser's view of one BGCtrlDef/BGCtrl section of the
// original text: where its key lines are.
type bgctrlSection struct {
	header     int            // line index of the header
	last       int            // line index of the last key line (header if none)
	typed      map[string]int // typed key (BGCtrl: type, ctrlid, time; BGCtrlDef: ctrlid, looptime) -> last line index
	isDef      bool           // BGCtrlDef section (only ctrlid/looptime are read) vs BGCtrl
	paramLines []int          // line index of each Params entry, in order
	paramKeys  []string
}

type bgctrlGroup struct {
	section bgctrlSection
	ctrls   []bgctrlSection
}

// scanBGCtrlSections walks lines the way Parse does (same comment
// stripping, header and key recognition) and returns the groups found.
func scanBGCtrlSections(lines []string) []bgctrlGroup {
	var groups []bgctrlGroup
	var cur *bgctrlSection
	for i, raw := range lines {
		line := stripStageComment(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			cur = nil
			if !strings.HasSuffix(line, "]") {
				continue
			}
			inner := strings.TrimSpace(line[1 : len(line)-1])
			if _, ok := bgElementName(inner); ok {
				continue
			}
			kind, _, ok := bgCtrlHeader(inner)
			if !ok {
				continue
			}
			sec := bgctrlSection{header: i, last: i, typed: map[string]int{}, isDef: kind == sectionBGCtrlDef}
			switch {
			case kind == sectionBGCtrlDef:
				groups = append(groups, bgctrlGroup{section: sec})
				cur = &groups[len(groups)-1].section
			case len(groups) > 0:
				g := &groups[len(groups)-1]
				g.ctrls = append(g.ctrls, sec)
				cur = &g.ctrls[len(g.ctrls)-1]
			}
			continue
		}
		if cur == nil {
			continue
		}
		key, value, ok := parseStageKeyValueLine(line)
		if !ok {
			continue
		}
		cur.last = i
		lower := strings.ToLower(key)
		if cur.isDef {
			if lower == "ctrlid" || lower == "looptime" {
				cur.typed[lower] = i
			}
			continue
		}
		switch lower {
		case "type", "ctrlid":
			cur.typed[lower] = i
		case "time":
			if !strings.Contains(value, ",") {
				cur.typed[lower] = i
				continue
			}
			fallthrough
		default:
			cur.paramLines = append(cur.paramLines, i)
			cur.paramKeys = append(cur.paramKeys, lower)
		}
	}
	return groups
}

// lineEdits collects the line-level changes applied to the original text.
type lineEdits struct {
	replace map[int]string // line index -> new value
	remove  map[int]bool
	after   map[int][]string // new "key = value" lines (no EOL) after a line
}

func splitLines(s string) []string {
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func lineEOL(line string) string {
	switch {
	case strings.HasSuffix(line, "\r\n"):
		return "\r\n"
	case strings.HasSuffix(line, "\n"):
		return "\n"
	}
	return ""
}

// replaceValue swaps the value of a "key = value ; comment" line, keeping
// the key text, the spacing around the value, the comment and the EOL.
func replaceValue(line, value string) string {
	eol := lineEOL(line)
	body := strings.TrimSuffix(line, eol)
	comment := ""
	if i := strings.IndexByte(body, ';'); i != -1 {
		body, comment = body[:i], body[i:]
	}
	eq := strings.IndexByte(body, '=')
	region := body[eq+1:]
	trimmed := strings.TrimSpace(region)
	lead := region[:len(region)-len(strings.TrimLeft(region, " \t"))]
	trail := ""
	if trimmed != "" {
		trail = region[len(lead)+len(trimmed):]
	}
	if lead == "" {
		lead = " "
	}
	return body[:eq+1] + lead + formatBGCtrlValue(value) + trail + comment + eol
}

type kv struct{ key, value string }

// patchSection records the edits turning one section's baseline key values
// into the edited ones. typed lists (key, baseline, edited) triples;
// params are matched baseline-to-edited by key and occurrence.
func (e *lineEdits) patchSection(sec bgctrlSection, typed [][3]string, base, edited []BGCtrlParam) {
	var added []kv
	for _, t := range typed {
		if t[1] == t[2] {
			continue
		}
		if i, ok := sec.typed[strings.ToLower(t[0])]; ok {
			e.replace[i] = t[2]
		} else {
			added = append(added, kv{t[0], t[2]})
		}
	}
	used := make([]bool, len(edited))
	for bi, bp := range base {
		match := -1
		for ei, ep := range edited {
			if !used[ei] && ep.Key == bp.Key {
				match = ei
				break
			}
		}
		line := sec.paramLines[bi]
		switch {
		case match == -1:
			e.remove[line] = true
		default:
			used[match] = true
			if edited[match].Value != bp.Value {
				e.replace[line] = edited[match].Value
			}
		}
	}
	for ei, ep := range edited {
		if !used[ei] {
			added = append(added, kv{ep.Key, ep.Value})
		}
	}
	for _, a := range added {
		e.after[sec.last] = append(e.after[sec.last], a.key+" = "+formatBGCtrlValue(a.value))
	}
}

func hasReplacement(m map[int]string, i int) bool {
	_, ok := m[i]
	return ok
}

func (e *lineEdits) apply(lines []string, crlf bool) string {
	var b strings.Builder
	for i, line := range lines {
		switch {
		case e.remove[i]:
			// line dropped
		case hasReplacement(e.replace, i):
			b.WriteString(replaceValue(line, e.replace[i]))
		default:
			b.WriteString(line)
		}
		extra := e.after[i]
		if len(extra) == 0 {
			continue
		}
		eol := lineEOL(line)
		if eol == "" {
			eol = "\n"
			if crlf {
				eol = "\r\n"
			}
			b.WriteString(eol)
		}
		for _, l := range extra {
			b.WriteString(l + eol)
		}
		if lineEOL(line) == "" {
			// the original last line had no EOL: do not add a final one
			s := b.String()
			b.Reset()
			b.WriteString(strings.TrimSuffix(s, eol))
		}
	}
	return b.String()
}

// patchBGCtrlDefs edits original in place so it reflects edited's
// BGCtrlDefs, provided edited differs from baseline (both normalized) only
// there, with unchanged group and per-group controller counts. ok is false
// when the edit cannot be patched safely, including when the patched text
// does not parse back to edited.
func patchBGCtrlDefs(original []byte, baseline, edited Stage) ([]byte, bool) {
	b, e := baseline, edited
	b.BGCtrlDefs, e.BGCtrlDefs = nil, nil
	if !reflect.DeepEqual(b, e) || len(baseline.BGCtrlDefs) != len(edited.BGCtrlDefs) {
		return nil, false
	}

	lines := splitLines(string(original))
	groups := scanBGCtrlSections(lines)
	if len(groups) != len(baseline.BGCtrlDefs) {
		return nil, false
	}

	edits := &lineEdits{replace: map[int]string{}, remove: map[int]bool{}, after: map[int][]string{}}
	for gi, bg := range baseline.BGCtrlDefs {
		eg := edited.BGCtrlDefs[gi]
		g := groups[gi]
		if bg.Name != eg.Name || len(bg.Controllers) != len(eg.Controllers) || len(g.ctrls) != len(bg.Controllers) {
			return nil, false
		}
		edits.patchSection(g.section, [][3]string{
			{"ctrlID", strconv.Itoa(bg.CtrlID), strconv.Itoa(eg.CtrlID)},
			{"looptime", strconv.Itoa(bg.LoopTime), strconv.Itoa(eg.LoopTime)},
		}, nil, nil)
		for ci, bc := range bg.Controllers {
			ec := eg.Controllers[ci]
			if len(g.ctrls[ci].paramLines) != len(bc.Params) {
				return nil, false
			}
			edits.patchSection(g.ctrls[ci], [][3]string{
				{"type", bc.Type, ec.Type},
				{"ctrlID", strconv.Itoa(bc.CtrlID), strconv.Itoa(ec.CtrlID)},
				{"time", strconv.Itoa(bc.Time), strconv.Itoa(ec.Time)},
			}, bc.Params, ec.Params)
		}
	}

	patched := []byte(edits.apply(lines, bytes.Contains(original, []byte("\r\n"))))
	got, err := Parse(bytes.NewReader(patched))
	if err != nil {
		return nil, false
	}
	normalizeStageForJSON(&got)
	if !reflect.DeepEqual(got, edited) {
		return nil, false
	}
	return patched, true
}
