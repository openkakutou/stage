---
date: 2026-08-29
status: accepted
---
# Expose animation frame resolution as one batched WASM call, not one call per element

**Context:** Backlog item 011 needs `OpenKakutouStage.load`'s existing `animations` data (item 009) to become actually usable by a JS consumer to resolve which sprite an animated BG element shows at a given elapsed time — `stage-viewer-web`'s playback preview will call this on every rendered frame, for every animated element on screen at once.

**Decision:** Add `OpenKakutouStage.resolveAnimationFrames(requestsJSON)`: a single call taking a JSON array of `{animation, elapsedTicks}` requests (one per animated element needing resolution this frame) and returning `{sprites: [...], error}` — a resolved sprite per request, same order. A request's `animation` is the raw `BGAnimation` object already available from `load`'s own `stage.animations[actionNumber]` (or `null` when no block matches that `ActionNumber`). The call is a thin `syscall/js` wrapper with no new domain logic — it decodes each request and delegates to the existing, already-tested `ResolveAnimationFrame`, which already resolves a missing/empty/malformed animation to the blank sprite sentinel without erroring.

**Reason:** A plan consultation with the realtime-rendering domain flagged that one WASM call per animated element per frame does not scale: each JS↔WASM crossing plus JSON marshal/unmarshal carries fixed overhead that multiplies by element count on every single rendered frame, and `GOOS=js GOARCH=wasm`'s single-threaded, cooperative GC makes per-call allocation spikes a real jank risk rather than free background work. Collapsing N per-frame calls into 1 is the largest, cheapest available win and requires no new Go logic. A further optimization (parse the animation once into a Go-side handle, resolve cheaply per frame after that) was considered and deliberately deferred: it adds real complexity (handle lifecycle, cross-call state) for a cost that only matters once profiling shows GC jank from re-unmarshaling the same small frame data every tick — not assumed up front.

**Rejected alternatives:**
- *One `resolveAnimationFrame` call per animated element, called once per element per frame* — rejected: scales linearly with animated-element count on the hot path, the exact risk the expert consultation flagged.
- *A `loadAnimation(json) → handle` / `resolveAnimationFrame(handle, elapsedTicks)` two-step API, parsing each animation once and resolving by handle thereafter* — rejected for now as premature: real complexity (handle lifecycle management across the WASM boundary) for a performance problem not yet measured; documented here as the next lever if profiling later shows it's needed.
- *Folding animation resolution into `load` itself (e.g. `load` accepting an elapsed-time argument)* — rejected: conflates the read path (parsing a stage once) with a per-frame playback concern that has nothing to do with parsing, and would force a full stage re-parse on every animation tick.
