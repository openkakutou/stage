---
status: done
depends_on: [006]
---
# Expose Sprite Pixel Resolution Via WASM

## Description
The WASM entrypoint's `load()` call only carries a BG element's sprite *reference* (a `group`/`image` pair) through its JSON contract — never decoded pixel data — so a browser consumer currently has no WASM-side way to actually render a layer's sprite. Mirror the sibling `character` repo's own `resolveSprites` WASM global: a batched, stateless call resolving `(group, image)` requests against a loaded `.sff` sprite sheet into real pixel buffers, built on the external `sff` module's `ResolveSpritePixels` the same way `character`'s own `resolveSprites` already is.

## Acceptance Criteria
- [ ] `globalThis.OpenKakutouStage.resolveSprites(sffBytes, requests, overrideBytes)` resolves an array of `[group, image]` requests against a loaded `.sff` sprite sheet into one `{ pixels, width, height, error }` result per request, in order
- [ ] A request naming a sprite absent from the sheet reports a descriptive, distinguishable error rather than a silent placeholder
- [ ] An optional external palette override (`.act` bytes) recolors the resolved pixels, the same way `character`'s `resolveSprites` supports it
- [ ] Malformed `sffBytes` never throws or crashes the WASM runtime — every request in the batch reports an error instead

## Notes
Follow-up to item 006, which deliberately scoped the WASM entrypoint to `load`/`save` (the `Stage` data model itself) and left sprite pixel resolution out — see `docs/wasm.md`'s "Not yet exposed" section and `.vibe/decisions/005-wasm-entrypoint-mirrors-character-load-save-shape.md`. `stage`'s own `SpriteResolver` (Go-side) already resolves a `SpriteRef` to an `sff.Sprite` (metadata only); this item is specifically about exposing pixel *data* through the WASM boundary, mirroring `character/cmd/wasm`'s `resolveSprites` global closely enough that `stage-viewer-web` can reuse the same JS-side integration pattern `character-viewer-web` already established.
