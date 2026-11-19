package stage

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestParse_RealTrainingFixture_CameraTensionStartAndBGScale pins the values
// of the real quick-versus-bootstrap Training stage (testdata/README.md):
// "tension = 400", "startx/starty = 0", "zoffsetlink = 3" pointing at
// "[BG Floor]" (id 3, start 0,640, yscale 1), and the BG elements'
// scalestart/xscale/scaledelta lines.
func TestParse_RealTrainingFixture_CameraTensionStartAndBGScale(t *testing.T) {
	s, err := Parse(bytes.NewReader(readTestdataFile(t, "ikemen-training-stage.def")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.BGdef.Tension != 400 {
		t.Errorf("Tension: got %d, want 400", s.BGdef.Tension)
	}
	if s.BGdef.CameraStartX != 0 || s.BGdef.CameraStartY != 0 {
		t.Errorf("CameraStart: got %d,%d, want 0,0", s.BGdef.CameraStartX, s.BGdef.CameraStartY)
	}
	if s.BGdef.ZOffsetLink == nil || *s.BGdef.ZOffsetLink != 3 {
		t.Errorf("ZOffsetLink: got %v, want 3", s.BGdef.ZOffsetLink)
	}
	if s.BGdef.ZOffset != 0 {
		t.Errorf("ZOffset (as written): got %d, want 0", s.BGdef.ZOffset)
	}
	if s.BGdef.ResolvedZOffset != 640 {
		t.Errorf("ResolvedZOffset: got %d, want 640 (0 + Floor's start y 640 * yscale 1)", s.BGdef.ResolvedZOffset)
	}
	byName := map[string]BGElement{}
	for _, el := range s.Elements {
		byName[el.Name] = el
	}
	floor := byName["Floor"]
	if floor.ScaleStartX != 2.4 || floor.ScaleStartY != 2.4 {
		t.Errorf("Floor scalestart: got %v,%v, want 2.4,2.4", floor.ScaleStartX, floor.ScaleStartY)
	}
	if floor.XScaleTop != 0.307 || floor.XScaleBottom != 1.55186 {
		t.Errorf("Floor xscale: got %v,%v, want 0.307,1.55186", floor.XScaleTop, floor.XScaleBottom)
	}
	if floor.ScaleDeltaX != 0 || floor.ScaleDeltaY != 0 {
		t.Errorf("Floor scaledelta: got %v,%v, want default 0,0", floor.ScaleDeltaX, floor.ScaleDeltaY)
	}
	left := byName["Wall_Left"]
	if left.ScaleDeltaX != 0.00312 || left.ScaleDeltaY != 0 {
		t.Errorf("Wall_Left scaledelta: got %v,%v, want 0.00312,0", left.ScaleDeltaX, left.ScaleDeltaY)
	}
	if left.XScaleTop != 1 || left.XScaleBottom != 1 {
		t.Errorf("Wall_Left xscale: got %v,%v, want default 1,1", left.XScaleTop, left.XScaleBottom)
	}
}

// TestParse_CameraAndBGScaleKeysAbsent_UseIkemenDefaults checks Ikemen GO's
// defaults: tension 50 (camera.go:85), startx/starty 0, scalestart 1,1,
// xscale 1,1, scaledelta 0,0 (stage.go:176-181), no zoffsetlink (-1,
// stage.go:1151) so ResolvedZOffset equals ZOffset.
func TestParse_CameraAndBGScaleKeysAbsent_UseIkemenDefaults(t *testing.T) {
	src := "[StageInfo]\nzoffset = 200\n[Camera]\nboundleft = -10\n[BG a]\ntype = normal\nid = 0\nstart = 0,50\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.BGdef.Tension != 50 {
		t.Errorf("Tension: got %d, want 50", s.BGdef.Tension)
	}
	if s.BGdef.ZOffsetLink != nil {
		t.Errorf("ZOffsetLink: got %v, want nil", *s.BGdef.ZOffsetLink)
	}
	if s.BGdef.ResolvedZOffset != 200 {
		t.Errorf("ResolvedZOffset: got %d, want 200 (an element with id 0 must not be linked)", s.BGdef.ResolvedZOffset)
	}
	el := s.Elements[0]
	if el.ScaleStartX != 1 || el.ScaleStartY != 1 || el.XScaleTop != 1 || el.XScaleBottom != 1 {
		t.Errorf("scale defaults: got scalestart %v,%v xscale %v,%v, want 1,1 and 1,1", el.ScaleStartX, el.ScaleStartY, el.XScaleTop, el.XScaleBottom)
	}
}

// TestParse_ZOffsetLink_AddsFirstMatchingElementStartYTimesYScale follows
// stage.go:1749-1761: the first element whose id matches adds int(start.y *
// yscale) to zoffset; a later element with the same id is ignored.
func TestParse_ZOffsetLink_AddsFirstMatchingElementStartYTimesYScale(t *testing.T) {
	src := "[StageInfo]\nzoffset = 10\nzoffsetlink = 7\nyscale = 0.5\n" +
		"[BG a]\nid = 7\nstart = 0,101\n[BG b]\nid = 7\nstart = 0,900\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.BGdef.ResolvedZOffset != 60 {
		t.Errorf("ResolvedZOffset: got %d, want 60 (10 + int(101*0.5))", s.BGdef.ResolvedZOffset)
	}
	if s.BGdef.ZOffset != 10 {
		t.Errorf("ZOffset must keep the written value: got %d, want 10", s.BGdef.ZOffset)
	}
}

// TestParse_ScaleStartSingleValue_SetsXOnly mirrors readF32ForStage
// (common.go:1192): a single value fills only the first component.
func TestParse_ScaleStartSingleValue_SetsXOnly(t *testing.T) {
	s, err := Parse(strings.NewReader("[BG a]\nscalestart = 2\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if el := s.Elements[0]; el.ScaleStartX != 2 || el.ScaleStartY != 1 {
		t.Errorf("scalestart: got %v,%v, want 2,1", el.ScaleStartX, el.ScaleStartY)
	}
}

func TestParse_MalformedCameraAndBGScaleValues_Error(t *testing.T) {
	cases := map[string]string{
		"tension":     "[Camera]\ntension = abc\n",
		"startx":      "[Camera]\nstartx = x\n",
		"starty":      "[Camera]\nstarty = y\n",
		"zoffsetlink": "[StageInfo]\nzoffsetlink = z\n",
		"scalestart":  "[BG a]\nscalestart = a,1\n",
		"scaledelta":  "[BG a]\nscaledelta = 1,b\n",
		"xscale":      "[BG a]\nxscale = q\n",
	}
	for key, src := range cases {
		_, err := Parse(strings.NewReader(src))
		if err == nil || !strings.Contains(err.Error(), "invalid "+key) {
			t.Errorf("%s: got error %v, want an \"invalid %s\" error", key, err, key)
		}
	}
}

// TestCameraAndBGScale_SurviveJSONAndSerializeRoundTrip covers the WASM JSON
// shape (cmd/wasm marshals Stage with encoding/json) and Serialize -> Parse.
func TestCameraAndBGScale_SurviveJSONAndSerializeRoundTrip(t *testing.T) {
	s, err := Parse(bytes.NewReader(readTestdataFile(t, "ikemen-training-stage.def")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, name := range []string{`"tension":400`, `"cameraStartX":0`, `"zOffsetLink":3`, `"resolvedZOffset":640`, `"scaleStartX":2.4`, `"scaleDeltaX":0.00312`, `"xScaleBottom":1.55186`} {
		if !bytes.Contains(data, []byte(name)) {
			t.Errorf("JSON lacks %s", name)
		}
	}
	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("serialize: %v", err)
	}
	back, err := Parse(&buf)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if back.BGdef.Tension != 400 || *back.BGdef.ZOffsetLink != 3 || back.BGdef.ResolvedZOffset != 640 || back.BGdef.ZOffset != 0 {
		t.Errorf("BGdef after round trip: %+v", back.BGdef)
	}
	for i, el := range s.Elements {
		b := back.Elements[i]
		if b.ScaleStartX != el.ScaleStartX || b.ScaleStartY != el.ScaleStartY || b.ScaleDeltaX != el.ScaleDeltaX || b.ScaleDeltaY != el.ScaleDeltaY || b.XScaleTop != el.XScaleTop || b.XScaleBottom != el.XScaleBottom {
			t.Errorf("element %s scale fields after round trip: got %+v, want %+v", el.Name, b, el)
		}
	}
}
