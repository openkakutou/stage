package stage

import "testing"

func TestStage_ZeroValue_IsUsableWithNoElements(t *testing.T) {
	var s Stage

	if s.Elements != nil {
		t.Errorf("expected zero-value Stage to have nil Elements, got %v", s.Elements)
	}
	if len(s.Elements) != 0 {
		t.Errorf("expected zero-value Stage to have 0 Elements, got %d", len(s.Elements))
	}
	if s.BGdef != (BGdef{}) {
		t.Errorf("expected zero-value Stage to have zero BGdef, got %+v", s.BGdef)
	}
	if s.CameraBounds != (CameraBounds{}) {
		t.Errorf("expected zero-value Stage to have zero CameraBounds, got %+v", s.CameraBounds)
	}
	if s.StageBoundaries != (StageBoundaries{}) {
		t.Errorf("expected zero-value Stage to have zero StageBoundaries, got %+v", s.StageBoundaries)
	}

	// Ranging over a nil Elements slice must not panic — a zero-value
	// Stage has to be usable on first access, not just constructible.
	for range s.Elements {
		t.Fatal("expected no iterations over a nil Elements slice")
	}
}

func TestStage_WithValues_ComposesAllFourConcepts(t *testing.T) {
	s := Stage{
		BGdef: BGdef{
			SpriteFile:       "stage0.sff",
			LocalCoordWidth:  320,
			LocalCoordHeight: 240,
		},
		Elements: []BGElement{
			{Name: "sky", Type: BGElementNormal, Sprite: SpriteRef{Group: 0, Image: 0}, LayerNo: 0},
			{Name: "cloud", Type: BGElementParallax, Sprite: SpriteRef{Group: 1, Image: 0}, DeltaX: 0.5, DeltaY: 0.5},
			{Name: "torch", Type: BGElementAnim, ActionNumber: 10, LayerNo: 1},
		},
		CameraBounds:    CameraBounds{Left: -180, Right: 180, High: -240, Low: 0},
		StageBoundaries: StageBoundaries{Left: -1000, Right: 1000},
	}

	if s.BGdef.SpriteFile != "stage0.sff" {
		t.Errorf("expected BGdef.SpriteFile %q, got %q", "stage0.sff", s.BGdef.SpriteFile)
	}
	if len(s.Elements) != 3 {
		t.Fatalf("expected 3 Elements, got %d", len(s.Elements))
	}
	if s.Elements[0].Type != BGElementNormal || s.Elements[1].Type != BGElementParallax || s.Elements[2].Type != BGElementAnim {
		t.Errorf("expected Elements in order [normal, parallax, anim], got [%q, %q, %q]", s.Elements[0].Type, s.Elements[1].Type, s.Elements[2].Type)
	}
	if s.CameraBounds.Left != -180 {
		t.Errorf("expected CameraBounds.Left -180, got %d", s.CameraBounds.Left)
	}
	if s.StageBoundaries.Left != -1000 {
		t.Errorf("expected StageBoundaries.Left -1000, got %d", s.StageBoundaries.Left)
	}
	// CameraBounds and StageBoundaries must stay independently settable —
	// the acceptance criterion this guards is "distinct concepts, not
	// conflated into one": a bug that aliased them would make this fail.
	if s.CameraBounds.Left == s.StageBoundaries.Left {
		t.Errorf("expected CameraBounds.Left (%d) and StageBoundaries.Left (%d) to be independently set, got equal values by coincidence of a shared field", s.CameraBounds.Left, s.StageBoundaries.Left)
	}
}

func TestStage_EmptyElementsSlice_DistinctFromNilButBothLenZero(t *testing.T) {
	s := Stage{Elements: []BGElement{}}

	if s.Elements == nil {
		t.Errorf("expected an explicitly empty slice to stay non-nil")
	}
	if len(s.Elements) != 0 {
		t.Errorf("expected 0 Elements, got %d", len(s.Elements))
	}
}
