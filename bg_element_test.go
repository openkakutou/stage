package stage

import "testing"

func TestBGElementType_Constants_MatchDefFormatTokens(t *testing.T) {
	// The exact string values matter: the future .def parser (backlog item
	// 002) will map the file's literal "type" token onto these constants,
	// so a wrong string here is a silent parser bug waiting to happen.
	cases := []struct {
		name string
		typ  BGElementType
		want string
	}{
		{"normal", BGElementNormal, "normal"},
		{"parallax", BGElementParallax, "parallax"},
		{"anim", BGElementAnim, "anim"},
	}
	for _, c := range cases {
		if string(c.typ) != c.want {
			t.Errorf("%s: expected BGElementType %q, got %q", c.name, c.want, string(c.typ))
		}
	}

	// The three constants must be pairwise distinct, otherwise a switch on
	// BGElement.Type could not actually distinguish the three kinds the
	// acceptance criteria require.
	if BGElementNormal == BGElementParallax || BGElementNormal == BGElementAnim || BGElementParallax == BGElementAnim {
		t.Errorf("expected BGElementNormal, BGElementParallax and BGElementAnim to be pairwise distinct, got %q, %q, %q", BGElementNormal, BGElementParallax, BGElementAnim)
	}
}

func TestBGElement_ZeroValue_AllFieldsZero(t *testing.T) {
	var e BGElement

	if e.Name != "" {
		t.Errorf("expected zero-value BGElement to have empty Name, got %q", e.Name)
	}
	if e.Type != "" {
		t.Errorf("expected zero-value BGElement to have empty Type, got %q", e.Type)
	}
	if e.Sprite != (SpriteRef{}) {
		t.Errorf("expected zero-value BGElement to have zero SpriteRef, got %+v", e.Sprite)
	}
	if e.ActionNumber != 0 {
		t.Errorf("expected zero-value BGElement to have ActionNumber 0, got %d", e.ActionNumber)
	}
	if e.LayerNo != 0 {
		t.Errorf("expected zero-value BGElement to have LayerNo 0, got %d", e.LayerNo)
	}
	if e.StartX != 0 || e.StartY != 0 {
		t.Errorf("expected zero-value BGElement to have StartX=0, StartY=0, got %d,%d", e.StartX, e.StartY)
	}
	if e.DeltaX != 0 || e.DeltaY != 0 {
		t.Errorf("expected zero-value BGElement to have DeltaX=0, DeltaY=0, got %v,%v", e.DeltaX, e.DeltaY)
	}
	if e.TileX != 0 || e.TileY != 0 {
		t.Errorf("expected zero-value BGElement to have TileX=0, TileY=0, got %d,%d", e.TileX, e.TileY)
	}
	if e.TileSpacingX != 0 || e.TileSpacingY != 0 {
		t.Errorf("expected zero-value BGElement to have TileSpacingX=0, TileSpacingY=0, got %d,%d", e.TileSpacingX, e.TileSpacingY)
	}
}

func TestBGElement_NormalType_WithValues_PreservesAssignedFieldValues(t *testing.T) {
	e := BGElement{
		Name:         "floor",
		Type:         BGElementNormal,
		Sprite:       SpriteRef{Group: 10, Image: 0},
		LayerNo:      0,
		StartX:       -100,
		StartY:       20,
		DeltaX:       1,
		DeltaY:       1,
		TileX:        1,
		TileY:        0,
		TileSpacingX: 320,
		TileSpacingY: 0,
	}

	if e.Type != BGElementNormal {
		t.Errorf("expected Type BGElementNormal, got %q", e.Type)
	}
	if e.Sprite.Group != 10 || e.Sprite.Image != 0 {
		t.Errorf("expected Sprite {10,0}, got %+v", e.Sprite)
	}
	if e.StartX != -100 || e.StartY != 20 {
		t.Errorf("expected Start (-100,20), got (%d,%d)", e.StartX, e.StartY)
	}
	if e.DeltaX != 1 || e.DeltaY != 1 {
		t.Errorf("expected Delta (1,1), got (%v,%v)", e.DeltaX, e.DeltaY)
	}
	if e.TileX != 1 || e.TileSpacingX != 320 {
		t.Errorf("expected TileX=1, TileSpacingX=320, got TileX=%d, TileSpacingX=%d", e.TileX, e.TileSpacingX)
	}
}

func TestBGElement_ParallaxType_WithFractionalDelta_PreservesValue(t *testing.T) {
	// Parallax elements are the reason Delta is a float, not an int:
	// depth is expressed via non-integer scroll ratios like 0.5,0.7.
	e := BGElement{
		Type:   BGElementParallax,
		Sprite: SpriteRef{Group: 20, Image: 3},
		DeltaX: 0.5,
		DeltaY: 0.7,
	}

	if e.Type != BGElementParallax {
		t.Errorf("expected Type BGElementParallax, got %q", e.Type)
	}
	if e.DeltaX != 0.5 || e.DeltaY != 0.7 {
		t.Errorf("expected Delta (0.5,0.7), got (%v,%v)", e.DeltaX, e.DeltaY)
	}
}

func TestBGElement_AnimType_UsesActionNumberInsteadOfSprite(t *testing.T) {
	// An "anim" element is driven by an .air action number, not a static
	// sprite reference — Sprite is left at its zero value.
	e := BGElement{
		Type:         BGElementAnim,
		ActionNumber: 200,
		LayerNo:      1,
	}

	if e.Type != BGElementAnim {
		t.Errorf("expected Type BGElementAnim, got %q", e.Type)
	}
	if e.ActionNumber != 200 {
		t.Errorf("expected ActionNumber 200, got %d", e.ActionNumber)
	}
	if e.Sprite != (SpriteRef{}) {
		t.Errorf("expected zero SpriteRef for an anim element, got %+v", e.Sprite)
	}
	if e.LayerNo != 1 {
		t.Errorf("expected LayerNo 1 (drawn in front of characters), got %d", e.LayerNo)
	}
}
