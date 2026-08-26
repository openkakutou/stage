package stage

import (
	"errors"
	"strings"
	"testing"
)

func TestSerialize_FullStage_RoundTripsThroughParseWithEquivalentStructure(t *testing.T) {
	s := Stage{
		Name:   "Training Room",
		Author: "Elecbyte",
		BGdef: BGdef{
			SpriteFile:       "stage0.sff",
			LocalCoordWidth:  320,
			LocalCoordHeight: 240,
			ZOffset:          220,
			ZoomOut:          0.75,
			ZoomIn:           1.5,
		},
		Elements: []BGElement{
			{
				Name:    "sky",
				Type:    BGElementNormal,
				Sprite:  SpriteRef{Group: 0, Image: 0},
				LayerNo: 0,
				StartX:  0, StartY: 0,
			},
			{
				Name:    "cloud",
				Type:    BGElementParallax,
				Sprite:  SpriteRef{Group: 1, Image: 0},
				LayerNo: 0,
				StartX:  10, StartY: 20,
				DeltaX: 0.5, DeltaY: 0.8,
				TileX: 1, TileY: 0,
				TileSpacingX: 50, TileSpacingY: 0,
			},
			{
				Name:         "torch",
				Type:         BGElementAnim,
				ActionNumber: 10,
				LayerNo:      1,
			},
		},
		CameraBounds:    CameraBounds{Left: -180, Right: 180, High: -240, Low: 0},
		StageBoundaries: StageBoundaries{Left: -1000, Right: 1000},
	}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	if got.Name != s.Name || got.Author != s.Author {
		t.Errorf("Name/Author: expected %q/%q, got %q/%q", s.Name, s.Author, got.Name, got.Author)
	}
	if got.BGdef != s.BGdef {
		t.Errorf("BGdef: expected %+v, got %+v", s.BGdef, got.BGdef)
	}
	if got.CameraBounds != s.CameraBounds {
		t.Errorf("CameraBounds: expected %+v, got %+v", s.CameraBounds, got.CameraBounds)
	}
	if got.StageBoundaries != s.StageBoundaries {
		t.Errorf("StageBoundaries: expected %+v, got %+v", s.StageBoundaries, got.StageBoundaries)
	}
	if len(got.Elements) != len(s.Elements) {
		t.Fatalf("Elements: expected %d entries, got %d: %+v", len(s.Elements), len(got.Elements), got.Elements)
	}
	for i, want := range s.Elements {
		if got.Elements[i] != want {
			t.Errorf("Elements[%d]: expected %+v, got %+v", i, want, got.Elements[i])
		}
	}
}

func TestSerialize_MinimalStage_OmitsEmptySpriteFileAndRoundTrips(t *testing.T) {
	s := Stage{
		BGdef: BGdef{LocalCoordWidth: 320, LocalCoordHeight: 240},
	}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	if strings.Contains(strings.ToLower(output), "spr =") {
		t.Errorf("expected output to omit unset SpriteFile, got:\n%s", output)
	}

	got, err := Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, output)
	}
	if got.BGdef.SpriteFile != "" {
		t.Errorf("expected empty SpriteFile, got %q", got.BGdef.SpriteFile)
	}
	if got.BGdef.LocalCoordWidth != 320 || got.BGdef.LocalCoordHeight != 240 {
		t.Errorf("expected LocalCoordWidth/Height 320/240, got %d/%d", got.BGdef.LocalCoordWidth, got.BGdef.LocalCoordHeight)
	}
	if len(got.Elements) != 0 {
		t.Errorf("expected no Elements, got %+v", got.Elements)
	}
}

func TestSerialize_EmptyNameAndAuthor_OmitsInfoLinesButStillWritesSection(t *testing.T) {
	s := Stage{BGdef: BGdef{LocalCoordWidth: 320, LocalCoordHeight: 240}}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "[Info]") {
		t.Errorf("expected the [Info] header to still be written, got:\n%s", output)
	}
	if strings.Contains(strings.ToLower(output), "name =") || strings.Contains(strings.ToLower(output), "author =") {
		t.Errorf("expected no name/author lines when both are empty, got:\n%s", output)
	}

	got, err := Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, output)
	}
	if got.Name != "" || got.Author != "" {
		t.Errorf("expected empty Name/Author, got %q/%q", got.Name, got.Author)
	}
}

func TestSerialize_ZeroValueStage_ProducesValidReparseableOutput(t *testing.T) {
	var buf strings.Builder
	if err := Serialize(&buf, Stage{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	var zero Stage
	if got.BGdef != zero.BGdef {
		t.Errorf("BGdef: expected zero value, got %+v", got.BGdef)
	}
	if got.CameraBounds != zero.CameraBounds {
		t.Errorf("CameraBounds: expected zero value, got %+v", got.CameraBounds)
	}
	if got.StageBoundaries != zero.StageBoundaries {
		t.Errorf("StageBoundaries: expected zero value, got %+v", got.StageBoundaries)
	}
	if len(got.Elements) != 0 {
		t.Errorf("expected no Elements, got %+v", got.Elements)
	}
}

func TestSerialize_AnimElement_OmitsSpriteNoWritesActionNo(t *testing.T) {
	s := Stage{Elements: []BGElement{
		{Name: "torch", Type: BGElementAnim, ActionNumber: 42},
	}}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	if strings.Contains(strings.ToLower(output), "spriteno") {
		t.Errorf("expected an anim element to omit spriteno, got:\n%s", output)
	}
	if !strings.Contains(output, "actionno = 42") {
		t.Errorf("expected actionno = 42 in output, got:\n%s", output)
	}

	got, err := Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, output)
	}
	if len(got.Elements) != 1 || got.Elements[0].ActionNumber != 42 || got.Elements[0].Type != BGElementAnim {
		t.Errorf("expected one anim element with ActionNumber 42, got %+v", got.Elements)
	}
}

func TestSerialize_NormalElement_OmitsActionNoWritesSpriteNo(t *testing.T) {
	s := Stage{Elements: []BGElement{
		{Name: "sky", Type: BGElementNormal, Sprite: SpriteRef{Group: 3, Image: 1}},
	}}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	if strings.Contains(strings.ToLower(output), "actionno") {
		t.Errorf("expected a normal element to omit actionno, got:\n%s", output)
	}
	if !strings.Contains(output, "spriteno = 3,1") {
		t.Errorf("expected spriteno = 3,1 in output, got:\n%s", output)
	}
}

func TestSerialize_ModelStage_WritesThreeDSectionsAndRoundTrips(t *testing.T) {
	s := Stage{
		BGdef: BGdef{
			SpriteFile: "stage0.sff",
			ModelFile:  "stage3d.glb",
			Near:       1, Far: 10000, FOV: 40, YShift: 0.5,
		},
		Model: Model{
			OffsetX: 0, OffsetY: -0.25, OffsetZ: -1,
			ScaleX: 0.5, ScaleY: 0.5, ScaleZ: 0.5,
			Environment:          "stage.hdr",
			EnvironmentIntensity: 1.2,
		},
		Scaling:         Scaling{DepthToScreen: 0.5, TopZ: 0, BottomZ: 50, TopScale: 1, BottomScale: 1.2},
		StageBoundaries: StageBoundaries{Left: -1000, Right: 1000, TopBound: -50, BottomBound: 50},
		PlayerStartZ:    PlayerStartZ{P1: -10, P2: 10},
	}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "[Model]") {
		t.Errorf("expected a [Model] section in output, got:\n%s", output)
	}
	if !strings.Contains(output, "[Scaling]") {
		t.Errorf("expected a [Scaling] section in output, got:\n%s", output)
	}

	got, err := Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, output)
	}
	if got.BGdef != s.BGdef {
		t.Errorf("BGdef: expected %+v, got %+v", s.BGdef, got.BGdef)
	}
	if got.Model != s.Model {
		t.Errorf("Model: expected %+v, got %+v", s.Model, got.Model)
	}
	if got.Scaling != s.Scaling {
		t.Errorf("Scaling: expected %+v, got %+v", s.Scaling, got.Scaling)
	}
	if got.StageBoundaries != s.StageBoundaries {
		t.Errorf("StageBoundaries: expected %+v, got %+v", s.StageBoundaries, got.StageBoundaries)
	}
	if got.PlayerStartZ != s.PlayerStartZ {
		t.Errorf("PlayerStartZ: expected %+v, got %+v", s.PlayerStartZ, got.PlayerStartZ)
	}
}

func TestSerialize_NonModelStage_OmitsAllThreeDOutput(t *testing.T) {
	// A stage with no model file is a 2D stage: its output must stay
	// byte-for-byte what item 008 found before this feature existed, so an
	// existing 2D stage-editor save never gains an unrequested [Model]/
	// [Scaling] section or new Camera/PlayerInfo keys.
	s := Stage{
		BGdef:           BGdef{SpriteFile: "stage0.sff", LocalCoordWidth: 320, LocalCoordHeight: 240},
		CameraBounds:    CameraBounds{Left: -180, Right: 180},
		StageBoundaries: StageBoundaries{Left: -1000, Right: 1000},
	}

	var buf strings.Builder
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	for _, unwanted := range []string{"[Model]", "[Scaling]", "model =", "near =", "far =", "fov =", "yshift =", "topbound", "botbound", "startz"} {
		if strings.Contains(strings.ToLower(output), strings.ToLower(unwanted)) {
			t.Errorf("expected a non-model stage to omit %q from output, got:\n%s", unwanted, output)
		}
	}
}

func TestSerialize_WriterError_ReturnsError(t *testing.T) {
	s := Stage{BGdef: BGdef{SpriteFile: "stage0.sff"}}

	err := Serialize(errorWriter{}, s)
	if err == nil {
		t.Fatal("expected an error from a failing writer, got nil")
	}
}

// errorWriter is an io.Writer that always fails, used to exercise
// Serialize's write-error path.
type errorWriter struct{}

func (errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("boom")
}
