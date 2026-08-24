---
status: in_progress
depends_on: [004]
---
# Resolve BG Element Frames/Animation

## Description
Stage backgrounds aren't always static: a parallax BG element scrolls at a delta relative to the camera (giving depth), and an animated BG element cycles through a sequence of sprite frames over time (declared the same way `character`'s `.air` actions declare frame sequences and timing, but embedded inline in the stage `.def`'s `[BG ...]` sections rather than a separate file). This item resolves both: computing a parallax element's effective screen position from the camera position and its configured scroll deltas, and resolving an animated element's currently-visible sprite from elapsed time/tick count, the same kind of playback-state resolution `air.SpriteResolver` performs for character animations.

## Acceptance Criteria
- [ ] A parallax BG element's on-screen position is computed correctly from camera position and its `delta`-style scroll configuration
- [ ] An animated BG element resolves to the correct sprite for a given elapsed time/tick, including looping back to the start after its last frame
- [ ] A static (non-parallax, non-animated) BG element is unaffected — same resolved position/sprite regardless of camera movement or elapsed time
- [ ] An animated element with a malformed or empty frame sequence resolves to a defined fallback (e.g. no sprite shown) instead of panicking or returning an out-of-bounds frame

## Notes
None.
