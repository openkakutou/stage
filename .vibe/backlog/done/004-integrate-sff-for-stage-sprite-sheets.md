---
status: done
depends_on: [001]
---
# Integrate `sff` For Stage Sprite Sheets

## Description
Stage BG elements reference sprite sheets in the `.sff` format (the same format `character` sprites use), via a `spr` file path declared in the stage `.def`'s `[Info]` section plus per-BG-element `(group, image)` references, mirroring how `air.Frame` references sprites in `character`. Integrate the external `github.com/openkakutou/sff` module (never `character`, per the roadmap's decision `007`) to resolve those references to actual sprite data — following the same "depend on the external module directly" approach `character`'s own backlog item 035 takes for its migration off the internal package. This is the cross-repo dependency the roadmap decision `009` calls out explicitly.

## Acceptance Criteria
- [ ] `go.mod` depends on `github.com/openkakutou/sff` (a released, tagged version — see Notes)
- [ ] A BG element's sprite reference resolves to an actual `sff.Sprite`/pixel data via a loaded `.sff` file, working with either `.sff` version the same way `character`'s `air.NewSpriteResolver` does
- [ ] A BG element referencing a sprite that doesn't exist in the loaded sprite sheet returns a descriptive error, not a silent zero value
- [ ] No `.sff` parsing/decoding logic is reimplemented in this repo — everything routes through `sff`

## Notes
Cross-repo blocker: needs `sff`'s own backlog item 001 (extraction/migration of `.sff` parse/serialize/palette code from `character`) tagged and released before this item can start, since `sff` currently has no functionality yet. Not expressible via `depends_on` frontmatter (same limitation noted in `character`'s item 035 — that only tracks same-repo item numbers).
