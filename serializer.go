package stage

import (
	"fmt"
	"io"
)

// Serialize writes s to w as MUGEN/Ikemen GO stage .def text: a "[BGDef]"
// section, a "[StageInfo]" section, a "[Camera]" section, a "[PlayerInfo]"
// section, then one "[BG <name>]" section per element in slice order.
//
// This is a first-pass write path: it does not attempt a byte-exact
// round-trip of any original file's formatting, comments, or unrecognized
// sections (a separate, format-preserving concern — see Document) — it only
// guarantees valid, readable output that Parse reads back into an
// equivalent Stage. BGdef.SpriteFile is omitted from "[BGDef]" when empty,
// matching character's def.Serialize convention for optional string
// fields; every other field is numeric, so it is always written, since a
// zero value there (e.g. ZOffset 0, an element's Delta 0,0) is meaningful
// data rather than an "unset" sentinel. A BGElement's Sprite is written
// only for BGElementNormal/BGElementParallax and its ActionNumber only for
// BGElementAnim, matching how Parse itself only populates the field the
// element's Type actually uses.
func Serialize(w io.Writer, s Stage) error {
	if err := writeBGDefSection(w, s.BGdef); err != nil {
		return err
	}
	if err := writeStageInfoSection(w, s.BGdef); err != nil {
		return err
	}
	if err := writeCameraSection(w, s); err != nil {
		return err
	}
	if err := writePlayerInfoSection(w, s.StageBoundaries); err != nil {
		return err
	}
	for _, el := range s.Elements {
		if err := writeBGElementSection(w, el); err != nil {
			return err
		}
	}
	return nil
}

func writeBGDefSection(w io.Writer, bgdef BGdef) error {
	if _, err := fmt.Fprintln(w, "[BGDef]"); err != nil {
		return fmt.Errorf("stage: writing [BGDef] header: %w", err)
	}
	if bgdef.SpriteFile != "" {
		if _, err := fmt.Fprintf(w, "spr = %s\n", bgdef.SpriteFile); err != nil {
			return fmt.Errorf("stage: writing spr: %w", err)
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("stage: writing section separator: %w", err)
	}
	return nil
}

func writeStageInfoSection(w io.Writer, bgdef BGdef) error {
	if _, err := fmt.Fprintln(w, "[StageInfo]"); err != nil {
		return fmt.Errorf("stage: writing [StageInfo] header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "localcoord = %d,%d\n", bgdef.LocalCoordWidth, bgdef.LocalCoordHeight); err != nil {
		return fmt.Errorf("stage: writing localcoord: %w", err)
	}
	if _, err := fmt.Fprintf(w, "zoffset = %d\n", bgdef.ZOffset); err != nil {
		return fmt.Errorf("stage: writing zoffset: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("stage: writing section separator: %w", err)
	}
	return nil
}

func writeCameraSection(w io.Writer, s Stage) error {
	if _, err := fmt.Fprintln(w, "[Camera]"); err != nil {
		return fmt.Errorf("stage: writing [Camera] header: %w", err)
	}
	fields := []struct {
		key   string
		value int
	}{
		{"boundleft", s.CameraBounds.Left},
		{"boundright", s.CameraBounds.Right},
		{"boundhigh", s.CameraBounds.High},
		{"boundlow", s.CameraBounds.Low},
	}
	for _, f := range fields {
		if _, err := fmt.Fprintf(w, "%s = %d\n", f.key, f.value); err != nil {
			return fmt.Errorf("stage: writing %s: %w", f.key, err)
		}
	}
	if _, err := fmt.Fprintf(w, "zoomout = %s\n", formatStageFloat(s.BGdef.ZoomOut)); err != nil {
		return fmt.Errorf("stage: writing zoomout: %w", err)
	}
	if _, err := fmt.Fprintf(w, "zoomin = %s\n", formatStageFloat(s.BGdef.ZoomIn)); err != nil {
		return fmt.Errorf("stage: writing zoomin: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("stage: writing section separator: %w", err)
	}
	return nil
}

func writePlayerInfoSection(w io.Writer, bounds StageBoundaries) error {
	if _, err := fmt.Fprintln(w, "[PlayerInfo]"); err != nil {
		return fmt.Errorf("stage: writing [PlayerInfo] header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "leftbound = %d\n", bounds.Left); err != nil {
		return fmt.Errorf("stage: writing leftbound: %w", err)
	}
	if _, err := fmt.Fprintf(w, "rightbound = %d\n", bounds.Right); err != nil {
		return fmt.Errorf("stage: writing rightbound: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("stage: writing section separator: %w", err)
	}
	return nil
}

func writeBGElementSection(w io.Writer, el BGElement) error {
	if _, err := fmt.Fprintf(w, "[BG %s]\n", el.Name); err != nil {
		return fmt.Errorf("stage: writing [BG %s] header: %w", el.Name, err)
	}
	if _, err := fmt.Fprintf(w, "type = %s\n", el.Type); err != nil {
		return fmt.Errorf("stage: writing type: %w", err)
	}
	if el.Type == BGElementAnim {
		if _, err := fmt.Fprintf(w, "actionno = %d\n", el.ActionNumber); err != nil {
			return fmt.Errorf("stage: writing actionno: %w", err)
		}
	} else {
		if _, err := fmt.Fprintf(w, "spriteno = %d,%d\n", el.Sprite.Group, el.Sprite.Image); err != nil {
			return fmt.Errorf("stage: writing spriteno: %w", err)
		}
	}
	if _, err := fmt.Fprintf(w, "layerno = %d\n", el.LayerNo); err != nil {
		return fmt.Errorf("stage: writing layerno: %w", err)
	}
	if _, err := fmt.Fprintf(w, "start = %d,%d\n", el.StartX, el.StartY); err != nil {
		return fmt.Errorf("stage: writing start: %w", err)
	}
	if _, err := fmt.Fprintf(w, "delta = %s,%s\n", formatStageFloat(el.DeltaX), formatStageFloat(el.DeltaY)); err != nil {
		return fmt.Errorf("stage: writing delta: %w", err)
	}
	if _, err := fmt.Fprintf(w, "tile = %d,%d\n", el.TileX, el.TileY); err != nil {
		return fmt.Errorf("stage: writing tile: %w", err)
	}
	if _, err := fmt.Fprintf(w, "tilespacing = %d,%d\n", el.TileSpacingX, el.TileSpacingY); err != nil {
		return fmt.Errorf("stage: writing tilespacing: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("stage: writing section separator: %w", err)
	}
	return nil
}

// formatStageFloat formats f without a trailing ".0" for whole numbers
// (matching common .def authoring style, e.g. "1" rather than "1.0") while
// still producing a value strconv.ParseFloat (via parseFloatPair) accepts
// unchanged.
func formatStageFloat(f float64) string {
	return fmt.Sprintf("%g", f)
}
