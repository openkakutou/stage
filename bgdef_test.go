package stage

import "testing"

func TestBGdef_ZeroValue_AllFieldsZero(t *testing.T) {
	var b BGdef

	if b.SpriteFile != "" {
		t.Errorf("expected zero-value BGdef to have empty SpriteFile, got %q", b.SpriteFile)
	}
	if b.LocalCoordWidth != 0 || b.LocalCoordHeight != 0 {
		t.Errorf("expected zero-value BGdef to have LocalCoordWidth=0, LocalCoordHeight=0, got %d,%d", b.LocalCoordWidth, b.LocalCoordHeight)
	}
	if b.ZOffset != 0 {
		t.Errorf("expected zero-value BGdef to have ZOffset 0, got %d", b.ZOffset)
	}
	if b.ZoomOut != 0 {
		t.Errorf("expected zero-value BGdef to have ZoomOut 0, got %v", b.ZoomOut)
	}
	if b.ZoomIn != 0 {
		t.Errorf("expected zero-value BGdef to have ZoomIn 0, got %v", b.ZoomIn)
	}
}

func TestBGdef_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	b := BGdef{
		SpriteFile:       "stage0.sff",
		LocalCoordWidth:  320,
		LocalCoordHeight: 240,
		ZOffset:          220,
		ZoomOut:          0.75,
		ZoomIn:           1.5,
	}

	if b.SpriteFile != "stage0.sff" {
		t.Errorf("expected SpriteFile %q, got %q", "stage0.sff", b.SpriteFile)
	}
	if b.LocalCoordWidth != 320 {
		t.Errorf("expected LocalCoordWidth 320, got %d", b.LocalCoordWidth)
	}
	if b.LocalCoordHeight != 240 {
		t.Errorf("expected LocalCoordHeight 240, got %d", b.LocalCoordHeight)
	}
	if b.ZOffset != 220 {
		t.Errorf("expected ZOffset 220, got %d", b.ZOffset)
	}
	if b.ZoomOut != 0.75 {
		t.Errorf("expected ZoomOut 0.75, got %v", b.ZoomOut)
	}
	if b.ZoomIn != 1.5 {
		t.Errorf("expected ZoomIn 1.5, got %v", b.ZoomIn)
	}
}
