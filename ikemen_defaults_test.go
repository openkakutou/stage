package stage

import (
	"bytes"
	"strings"
	"testing"
)

// A MUGEN 1.0 stage often omits these keys: Ikemen GO's newCamera
// (camera.go:79, 88), newStage (stage.go:1147-1148) and newBackGround
// (stage.go:176) give them non-zero defaults.
const minimalStage = "[Info]\nname = \"Minimal\"\n[Camera]\n[PlayerInfo]\n[StageInfo]\n[BGdef]\nspr = s.sff\n[BG floor]\ntype = normal\nspriteno = 0,0\n"

func TestParse_OmittedKeysTakeIkemenGOsDefaults(t *testing.T) {
	s, err := Parse(strings.NewReader(minimalStage))
	if err != nil {
		t.Fatal(err)
	}
	if s.BGdef.LocalCoordWidth != 320 || s.BGdef.LocalCoordHeight != 240 {
		t.Errorf("localcoord = %d,%d, want 320,240 (camera.go:79)", s.BGdef.LocalCoordWidth, s.BGdef.LocalCoordHeight)
	}
	if s.BGdef.ZoomOut != 1 || s.BGdef.ZoomIn != 1 {
		t.Errorf("zoomout, zoomin = %v, %v, want 1, 1 (camera.go:88)", s.BGdef.ZoomOut, s.BGdef.ZoomIn)
	}
	if s.StageBoundaries.Left != -1000 || s.StageBoundaries.Right != 1000 {
		t.Errorf("bounds = %d, %d, want -1000, 1000 (stage.go:1147-1148)", s.StageBoundaries.Left, s.StageBoundaries.Right)
	}
	floor := s.Elements[0]
	if floor.DeltaX != 1 || floor.DeltaY != 1 {
		t.Errorf("floor delta = %v,%v, want 1,1 (stage.go:176)", floor.DeltaX, floor.DeltaY)
	}
}

func TestParse_WrittenKeysOverrideTheDefaults(t *testing.T) {
	src := "[Camera]\nzoomout = 0.5\n[PlayerInfo]\nleftbound = -200\n[StageInfo]\nlocalcoord = 1280,720\n[BG a]\ntype = normal\nspriteno = 0,0\ndelta = 0.2,0.3\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if s.BGdef.ZoomOut != 0.5 || s.BGdef.ZoomIn != 1 || s.StageBoundaries.Left != -200 || s.StageBoundaries.Right != 1000 || s.BGdef.LocalCoordWidth != 1280 {
		t.Errorf("got zoomout %v zoomin %v bounds %d,%d localcoord %d", s.BGdef.ZoomOut, s.BGdef.ZoomIn, s.StageBoundaries.Left, s.StageBoundaries.Right, s.BGdef.LocalCoordWidth)
	}
	if a := s.Elements[0]; a.DeltaX != 0.2 || a.DeltaY != 0.3 {
		t.Errorf("delta = %v,%v, want 0.2,0.3", a.DeltaX, a.DeltaY)
	}
}

func TestSerialize_AStageWithoutTheKeysIsNotSavedWithZeros(t *testing.T) {
	s, err := Parse(strings.NewReader(minimalStage))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, bad := range []string{"localcoord = 0,0", "zoomout = 0\n", "zoomin = 0\n", "\ndelta = 0,0"} {
		if strings.Contains(out, bad) {
			t.Errorf("serialized stage contains %q:\n%s", bad, out)
		}
	}
}
