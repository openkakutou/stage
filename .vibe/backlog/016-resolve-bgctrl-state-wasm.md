---
status: todo
depends_on: [014]
---
# Resolve controller-driven layer state at a tick (Go + WASM)

## Description
Computing how controllers change each layer over time (visibility, offsets, animation swaps) belongs in the library, as frame resolution already does for animations. Add a pure function returning each layer's resolved state at tick N, exposed via WASM for viewer and game use.

## Acceptance Criteria
- [ ] Resolve(stage, tick) returns per-layer visibility/offset/anim overrides
- [ ] Looping and timed controllers behave per the MUGEN reference
- [ ] WASM export handles many layers in one call
- [ ] Real-stage fixture test

## Notes
Consumers: stage-viewer-web, mode-quick-versus (native).
