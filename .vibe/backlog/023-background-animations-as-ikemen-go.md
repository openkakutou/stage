---
status: todo
---
# Background animations as Ikemen GO plays them

## Description
Found by the audit of 2026-11-21 (high and medium, to re-check against anim.go):
- A `Loopstart` after the last frame leaves `LoopStart = len(Frames)` (`parser.go:177`), and `ResolveAnimationFrame` then shows nothing (`animation.go:78`). Ikemen GO restores the loop start in effect before that frame (anim.go:383, 440-442).
- In `[Begin Action]`, any line that is not `loopstart` is a frame and a bad one aborts `Parse` (`parser.go:180`). Ikemen GO skips lines not starting with a digit or `-` (anim.go:47-49), reads Clsn and Interpolate lines, and accepts `loopstart` as a prefix (anim.go:395-437). A stage whose action was copied from an `.air` cannot be opened.
- A frame of time -1 is skipped (`animation.go:107`); in Ikemen GO it holds forever (anim.go:444-445).

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Each case above follows Ikemen GO, with a test; decision 004 is superseded where it differs
- [ ] A stage action with Clsn and Interpolate lines loads
