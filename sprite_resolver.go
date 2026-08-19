package stage

import (
	"fmt"

	"github.com/openkakutou/sff"
)

// SpriteResolver resolves a BG element's SpriteRef against a loaded set of
// sprite groups. sff.Sprite/sff.SpriteGroup are the same version-agnostic
// shape regardless of whether they came from a .sff v1 or v2 file, so
// Resolve needs no version-specific branching.
//
// This is the cross-repo integration point between stage and sff: per
// CLAUDE.md, a BG element's sprite reference is meaningless without a
// sprite collection to resolve it against. Mirrors the same shape
// character's air.SpriteResolver established for .air Frame references —
// see .vibe/decisions/002-sprite-resolver-takes-spriteref-not-bgelement.md
// for why this resolver takes a SpriteRef directly rather than a whole
// BGElement.
type SpriteResolver struct {
	sprites map[SpriteRef]sff.Sprite
}

// NewSpriteResolver indexes the given sprite groups by (Group, Image) so
// Resolve can look up any reference without rescanning groups on every
// call. Passing nil or an empty slice is valid and produces a resolver for
// which every Resolve call fails.
func NewSpriteResolver(groups []sff.SpriteGroup) *SpriteResolver {
	sprites := make(map[SpriteRef]sff.Sprite)
	for _, group := range groups {
		for _, sprite := range group.Sprites {
			sprites[SpriteRef{Group: sprite.Group, Image: sprite.Image}] = sprite
		}
	}
	return &SpriteResolver{sprites: sprites}
}

// Resolve returns the Sprite that ref's (Group, Image) reference names. It
// returns a descriptive error, rather than a zero Sprite, when no sprite
// with a matching (Group, Image) exists in the resolver — a missing
// reference must fail explicitly rather than silently rendering blank.
func (r *SpriteResolver) Resolve(ref SpriteRef) (sff.Sprite, error) {
	sprite, ok := r.sprites[ref]
	if !ok {
		return sff.Sprite{}, fmt.Errorf("stage: BG element references sprite (group %d, image %d), which was not found in the loaded sprite groups", ref.Group, ref.Image)
	}
	return sprite, nil
}
