package stage

import (
	"fmt"
	"strconv"
	"strings"
)

// BGCtrlDef is one "[BGCtrlDef <name>]" group of a stage .def file: a named
// set of background controllers sharing a loop time, followed in the file by
// its "[BGCtrl <n>]" blocks. Names are reused in real files, so a controller
// belongs to the group that precedes it by position, never by name.
type BGCtrlDef struct {
	// Name is the text after "BGCtrlDef" in the section header (may be empty).
	Name string `json:"name"`
	// CtrlID is the group's "ctrlID" key; 0 when absent. A single integer
	// only: comma-separated lists are not supported.
	CtrlID int `json:"ctrlId"`
	// LoopTime is the group's "looptime" key; 0 when absent.
	LoopTime int `json:"loopTime"`
	// Controllers are the group's "[BGCtrl <n>]" blocks, in file order.
	Controllers []BGCtrl `json:"controllers"`
}

// BGCtrl is one "[BGCtrl <n>]" block of a stage .def file.
type BGCtrl struct {
	// Type is the block's "type" key, lowercased ("visible", "palfx", ...).
	// Unknown types are kept as-is, never dropped.
	Type string `json:"type"`
	// CtrlID is the block's "ctrlID" key; 0 when absent.
	CtrlID int `json:"ctrlId"`
	// Time is the block's "time" key; 0 when absent, or when written as a
	// comma-separated list (kept raw in Params under "time").
	Time int `json:"time"`
	// Params are every other key of the block, in file order.
	Params []BGCtrlParam `json:"params"`
}

// BGCtrlParam is one raw key=value pair of a "[BGCtrl <n>]" block other than
// type/time/ctrlID.
type BGCtrlParam struct {
	// Key is the key, lowercased.
	Key string `json:"key"`
	// Value is the raw value text, unquoted, not interpreted.
	Value string `json:"value"`
}

// Section kinds returned by bgCtrlHeader.
const (
	sectionBGCtrlDef = "bgctrldef"
	sectionBGCtrl    = "bgctrl"
)

// bgCtrlHeader reports whether raw (a section header's trimmed inner text)
// opens a "BGCtrlDef" or "BGCtrl" section, case-insensitively, returning the
// section kind (sectionBGCtrlDef or sectionBGCtrl) and the optional trailing
// name.
func bgCtrlHeader(raw string) (kind, name string, ok bool) {
	word := raw
	rest := ""
	if i := strings.IndexAny(raw, " \t"); i != -1 {
		word, rest = raw[:i], strings.TrimSpace(raw[i+1:])
	}
	switch strings.ToLower(word) {
	case sectionBGCtrlDef:
		return sectionBGCtrlDef, rest, true
	case sectionBGCtrl:
		return sectionBGCtrl, rest, true
	}
	return "", "", false
}

// parseBGCtrlInt strictly parses the integer value of a BGCtrl key, wrapping
// a failure in a line-numbered error.
func parseBGCtrlInt(key, value string, lineNumber int) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("stage: line %d: invalid %s %q: %w", lineNumber, key, value, err)
	}
	return n, nil
}

// parseBGCtrlDefKey applies one key=value pair of a "[BGCtrlDef]" section.
func parseBGCtrlDefKey(def *BGCtrlDef, key, value string, lineNumber int) error {
	lower := strings.ToLower(key)
	switch lower {
	case "looptime":
		n, err := parseBGCtrlInt(lower, value, lineNumber)
		if err != nil {
			return err
		}
		def.LoopTime = n
	case "ctrlid":
		n, err := parseBGCtrlInt(lower, value, lineNumber)
		if err != nil {
			return err
		}
		def.CtrlID = n
	}
	return nil
}

// parseBGCtrlKey applies one key=value pair of a "[BGCtrl]" section. A "time"
// written as Ikemen's comma-separated list is not interpreted: Time stays 0
// and the raw text is kept in Params under "time".
func parseBGCtrlKey(c *BGCtrl, key, value string, lineNumber int) error {
	lower := strings.ToLower(key)
	switch lower {
	case "type":
		c.Type = strings.ToLower(strings.TrimSpace(value))
	case "time":
		if strings.Contains(value, ",") {
			c.Params = append(c.Params, BGCtrlParam{Key: lower, Value: value})
			return nil
		}
		n, err := parseBGCtrlInt(lower, value, lineNumber)
		if err != nil {
			return err
		}
		c.Time = n
	case "ctrlid":
		n, err := parseBGCtrlInt(lower, value, lineNumber)
		if err != nil {
			return err
		}
		c.CtrlID = n
	default:
		c.Params = append(c.Params, BGCtrlParam{Key: lower, Value: value})
	}
	return nil
}
