---
status: todo
depends_on: [009]
---
# Expose Animated BG Frame Resolution Via WASM

## Description
`OpenKakutouStage.load`'s JSON result (`cmd/wasm/main.go`) exposes a `BGElementAnim`'s raw `ActionNumber` but nothing that maps it to actual frame data — `BGAnimation`/`BGAnimFrame`/`ResolveAnimationFrame` (item 005) exist as Go API but cross no WASM boundary today. A consumer (`stage-viewer-web`'s own backlog item 005, currently blocked partly on this) needs either the parsed `BGAnimation` set included in the `load` result, or a new resolve-frame-at-time WASM call, to render animated BG playback for real. Found while attempting `stage-viewer-web#005`, not previously filed.

## Acceptance Criteria
- [ ] `OpenKakutouStage.load`'s JSON result exposes enough for a JS consumer to resolve which frame a `BGElementAnim`'s `ActionNumber` should show at a given elapsed time, mirroring `ResolveAnimationFrame`'s own semantics
- [ ] A `BGElementAnim` whose `ActionNumber` has no matching parsed `BGAnimation` (or whose frame data is malformed) is representable in the exposed result without crashing the WASM call
- [ ] Existing `load`/`save` behavior for non-`anim` elements is unchanged

## Notes
Depends on item `009` (`[Begin Action N]` parsing) — without it, `Parse` never populates real `BGAnimation` data to expose in the first place, regardless of what this item adds to the WASM boundary.
