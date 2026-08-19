package stage

import (
	"strings"
	"testing"

	"github.com/openkakutou/sff"
)

// sampleSpriteGroups builds a small, multi-group sprite collection with
// distinguishing metadata (Width/Height/AxisX/AxisY/Palette) per sprite, so
// a test can tell a wrong (but plausible) lookup apart from the correct one.
func sampleSpriteGroups() []sff.SpriteGroup {
	return []sff.SpriteGroup{
		{
			Index: 0,
			Sprites: []sff.Sprite{
				{Group: 0, Image: 0, Width: 320, Height: 240, AxisX: 0, AxisY: 0, Palette: 1},
				{Group: 0, Image: 1, Width: 322, Height: 242, AxisX: 1, AxisY: 1, Palette: 1},
			},
		},
		{
			Index: 5,
			Sprites: []sff.Sprite{
				{Group: 5, Image: 0, Width: 64, Height: 64, AxisX: 32, AxisY: 64, Palette: 2},
			},
		},
	}
}

func TestSpriteResolver_Resolve_ReturnsMatchingSpriteForKnownReference(t *testing.T) {
	resolver := NewSpriteResolver(sampleSpriteGroups())

	got, err := resolver.Resolve(SpriteRef{Group: 5, Image: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sff.Sprite{Group: 5, Image: 0, Width: 64, Height: 64, AxisX: 32, AxisY: 64, Palette: 2}
	if got != want {
		t.Errorf("expected sprite %+v, got %+v", want, got)
	}
}

func TestSpriteResolver_Resolve_DistinguishesGroupFromImageIndex(t *testing.T) {
	// Group 0 has images 0 and 1; group 5 has image 0. A reference to
	// (group 5, image 1) must NOT resolve to (group 0, image 1) or to
	// (group 5, image 0) even though each half of the key matches something.
	resolver := NewSpriteResolver(sampleSpriteGroups())

	_, err := resolver.Resolve(SpriteRef{Group: 5, Image: 1})
	if err == nil {
		t.Fatal("expected an error for (group 5, image 1), got nil")
	}
}

func TestSpriteResolver_Resolve_ReturnsDescriptiveErrorOnMissingSprite(t *testing.T) {
	resolver := NewSpriteResolver(sampleSpriteGroups())

	_, err := resolver.Resolve(SpriteRef{Group: 99, Image: 7})
	if err == nil {
		t.Fatal("expected an error for a non-existent (group, image) reference, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "99") || !strings.Contains(msg, "7") {
		t.Errorf("expected error to mention the missing (group, image) = (99, 7), got: %q", msg)
	}
}

func TestSpriteResolver_Resolve_ReturnsErrorWhenNoSpriteGroupsLoaded(t *testing.T) {
	resolver := NewSpriteResolver(nil)

	_, err := resolver.Resolve(SpriteRef{Group: 0, Image: 0})
	if err == nil {
		t.Fatal("expected an error when resolving against an empty sprite collection, got nil")
	}
}

func TestSpriteResolver_Resolve_WorksTheSameRegardlessOfSFFVersionOrigin(t *testing.T) {
	// sff.Sprite/sff.SpriteGroup are already version-agnostic (populated
	// identically regardless of whether the source file was .sff v1 or v2),
	// so the resolver needs no version-specific branching. This test
	// exercises the resolver against a sprite using a v2-only concept (a
	// non-zero Palette bank index unrelated to v1's shared-palette flag) to
	// document that no special-casing is required.
	groups := []sff.SpriteGroup{
		{Index: 3, Sprites: []sff.Sprite{{Group: 3, Image: 2, Width: 10, Height: 10, Palette: 4}}},
	}
	resolver := NewSpriteResolver(groups)

	got, err := resolver.Resolve(SpriteRef{Group: 3, Image: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := sff.Sprite{Group: 3, Image: 2, Width: 10, Height: 10, Palette: 4}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}
