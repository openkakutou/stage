package stage

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_MultiSectionSample_ProducesExpectedStage(t *testing.T) {
	src := `[Info]
name = "Training Room"
author = "Elecbyte"

[Camera]
boundleft = -180
boundright = 180
boundhigh = -240
boundlow = 0
zoomout = 0.75
zoomin = 1.5

[PlayerInfo]
leftbound = -1000
rightbound = 1000

[StageInfo]
localcoord = 320,240
zoffset = 220

[BGDef]
spr = stage0.sff

[BG sky]
type = normal
spriteno = 0,0
layerno = 0
start = 0,0

[BG cloud]
type = parallax
spriteno = 1,0
delta = 0.5,0.8
start = 10,20

[BG torch]
type = anim
actionno = 10
layerno = 1
`

	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.BGdef.SpriteFile != "stage0.sff" {
		t.Errorf("expected BGdef.SpriteFile %q, got %q", "stage0.sff", s.BGdef.SpriteFile)
	}
	if s.BGdef.LocalCoordWidth != 320 || s.BGdef.LocalCoordHeight != 240 {
		t.Errorf("expected LocalCoordWidth/Height 320/240, got %d/%d", s.BGdef.LocalCoordWidth, s.BGdef.LocalCoordHeight)
	}
	if s.BGdef.ZOffset != 220 {
		t.Errorf("expected ZOffset 220, got %d", s.BGdef.ZOffset)
	}
	if s.BGdef.ZoomOut != 0.75 || s.BGdef.ZoomIn != 1.5 {
		t.Errorf("expected ZoomOut/ZoomIn 0.75/1.5, got %v/%v", s.BGdef.ZoomOut, s.BGdef.ZoomIn)
	}

	wantCamera := CameraBounds{Left: -180, Right: 180, High: -240, Low: 0}
	if s.CameraBounds != wantCamera {
		t.Errorf("expected CameraBounds %+v, got %+v", wantCamera, s.CameraBounds)
	}

	wantBoundaries := StageBoundaries{Left: -1000, Right: 1000}
	if s.StageBoundaries != wantBoundaries {
		t.Errorf("expected StageBoundaries %+v, got %+v", wantBoundaries, s.StageBoundaries)
	}

	if len(s.Elements) != 3 {
		t.Fatalf("expected 3 Elements, got %d: %+v", len(s.Elements), s.Elements)
	}

	sky := s.Elements[0]
	if sky.Name != "sky" || sky.Type != BGElementNormal || sky.Sprite != (SpriteRef{Group: 0, Image: 0}) || sky.LayerNo != 0 {
		t.Errorf("expected sky element {normal, (0,0), layer 0}, got %+v", sky)
	}

	cloud := s.Elements[1]
	if cloud.Name != "cloud" || cloud.Type != BGElementParallax {
		t.Errorf("expected cloud element to be parallax, got %+v", cloud)
	}
	if cloud.DeltaX != 0.5 || cloud.DeltaY != 0.8 {
		t.Errorf("expected cloud DeltaX/DeltaY 0.5/0.8, got %v/%v", cloud.DeltaX, cloud.DeltaY)
	}
	if cloud.StartX != 10 || cloud.StartY != 20 {
		t.Errorf("expected cloud StartX/StartY 10/20, got %d/%d", cloud.StartX, cloud.StartY)
	}

	torch := s.Elements[2]
	if torch.Name != "torch" || torch.Type != BGElementAnim || torch.ActionNumber != 10 || torch.LayerNo != 1 {
		t.Errorf("expected torch element {anim, action 10, layer 1}, got %+v", torch)
	}
}

func TestParse_IkemenGoStyleWithTilingAndComments(t *testing.T) {
	// Ikemen GO stage .def commonly adds tile/tilespacing on BG elements and
	// tolerates trailing/whole-line comments the same way MUGEN itself does.
	src := `[Camera] ; scroll clamp
boundleft = -100 ; left edge
boundright = 100

[BG grid]
type = normal
spriteno = 2,3
tile = 1,0
tilespacing = 16,0
`

	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.CameraBounds.Left != -100 || s.CameraBounds.Right != 100 {
		t.Errorf("expected CameraBounds.Left/Right -100/100, got %d/%d", s.CameraBounds.Left, s.CameraBounds.Right)
	}
	if len(s.Elements) != 1 {
		t.Fatalf("expected 1 Element, got %d", len(s.Elements))
	}
	grid := s.Elements[0]
	if grid.Sprite != (SpriteRef{Group: 2, Image: 3}) {
		t.Errorf("expected grid Sprite {2,3}, got %+v", grid.Sprite)
	}
	if grid.TileX != 1 || grid.TileY != 0 {
		t.Errorf("expected TileX/TileY 1/0, got %d/%d", grid.TileX, grid.TileY)
	}
	if grid.TileSpacingX != 16 || grid.TileSpacingY != 0 {
		t.Errorf("expected TileSpacingX/TileSpacingY 16/0, got %d/%d", grid.TileSpacingX, grid.TileSpacingY)
	}
}

func TestParse_MissingTypeKey_DefaultsElementToNormal(t *testing.T) {
	src := `[BG plain]
spriteno = 4,0
`

	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Elements) != 1 {
		t.Fatalf("expected 1 Element, got %d", len(s.Elements))
	}
	if s.Elements[0].Type != BGElementNormal {
		t.Errorf("expected a missing type key to default to BGElementNormal, got %q", s.Elements[0].Type)
	}
}

func TestParse_UnrecognizedSectionsAndKeys_AreSkipped(t *testing.T) {
	src := `[Info]
name = "Some Stage"
does-not-exist = 1

[Music]
bgmusic = stage.ogg

[Shadow]
intensity = 128

[Camera]
boundleft = -50
notarealkey = whatever
boundright = 50
`

	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.CameraBounds.Left != -50 || s.CameraBounds.Right != 50 {
		t.Errorf("expected CameraBounds.Left/Right -50/50 despite surrounding unknown sections/keys, got %d/%d", s.CameraBounds.Left, s.CameraBounds.Right)
	}
}

func TestParse_EmptyInput_ReturnsZeroValueStageAndNoError(t *testing.T) {
	s, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Elements != nil {
		t.Errorf("expected nil Elements, got %v", s.Elements)
	}
	if s.BGdef != (BGdef{}) || s.CameraBounds != (CameraBounds{}) || s.StageBoundaries != (StageBoundaries{}) {
		t.Errorf("expected a zero-value Stage, got %+v", s)
	}
}

func TestParse_MalformedSectionHeader_ReturnsLineNumberedError(t *testing.T) {
	src := `[Camera]
boundleft = -10
[BG broken
type = normal
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a section header missing its closing bracket")
	}
	if !strings.Contains(err.Error(), "line 3") {
		t.Errorf("expected the error to name line 3, got: %v", err)
	}
}

func TestParse_NonNumericValueForIntegerKey_ReturnsLineNumberedError(t *testing.T) {
	src := `[Camera]
boundleft = notanumber
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a non-numeric value on an integer key")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected the error to name line 2, got: %v", err)
	}
}

func TestParse_NonNumericValueForFloatKey_ReturnsLineNumberedError(t *testing.T) {
	src := `[Camera]
zoomout = notafloat
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a non-numeric value on a float key")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected the error to name line 2, got: %v", err)
	}
}

func TestParse_MalformedCommaPairValue_ReturnsLineNumberedError(t *testing.T) {
	src := `[BG broken]
spriteno = notapair
`

	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for a spriteno value that isn't a group,image pair")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected the error to name line 2, got: %v", err)
	}
}

func TestParse_NonKeyValueLineInsideKnownSection_IsIgnored(t *testing.T) {
	src := `[Camera]
; a decorative separator below, not a key=value pair
----------------------
boundleft = -20
`

	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.CameraBounds.Left != -20 {
		t.Errorf("expected CameraBounds.Left -20, got %d", s.CameraBounds.Left)
	}
}

func TestParse_ReaderFailure_ReturnsWrappedError(t *testing.T) {
	_, err := Parse(errorReader{})
	if err == nil {
		t.Fatal("expected an error when the reader itself fails")
	}
}

// errorReader is an io.Reader that always fails, used to exercise Parse's
// scanner-error path.
type errorReader struct{}

func (errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("boom")
}
