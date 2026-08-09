---
status: todo
depends_on: [003]
---
# Model-Based BG Elements And 3D Stage Settings (Ikemen GO)

## Description
Extend the data model, parser, and serializer to cover Ikemen GO's 3D model-based stage extensions (see the roadmap's `.vibe/decisions/014`), the same format-preserving round-trip guarantee items 002/003 already provide for 2D stages. Verified `.def` additions to target: a `[Model]` section (model file reference, `Offset`/`Scale` placement, `Environment`/`EnvironmentIntensity` image-based lighting), 3D-specific `[Camera]` keys (`Near`, `Far`, `fov`, `YShift`) alongside the already-modeled 2D zoom keys, a new `[Scaling]` section (`DepthToScreen`, `topz`/`botz`, `topscale`/`botscale`), and a Z-axis extension of `StageBoundaries` (`[PlayerInfo]` `topbound`/`botbound`, per-player `Startz`). This repo only parses/serializes these as pure data — no model-file reading, no rendering, per its existing no-rendering-dependency constraint. A stage that doesn't use any of these sections must remain byte-for-byte and semantically identical to today's behavior.

## Acceptance Criteria
- [ ] `[Model]` section parses into a new model reference type: file path, `Offset` (X/Y/Z), `Scale` (X/Y/Z), `Environment` (`.hdr` path), `EnvironmentIntensity`
- [ ] Whether `[Model]` is a single stage-wide section or repeatable like `[BG name]` is confirmed against at least one real Ikemen GO 3D stage `.def` file before committing to either shape — not assumed
- [ ] `[Camera]`'s `Near`/`Far`/`fov`/`YShift` parse alongside the existing `ZoomOut`/`ZoomIn` fields, without disturbing 2D-only stages that omit them
- [ ] A new `[Scaling]` type models `DepthToScreen`, `topz`/`botz`, `topscale`/`botscale`
- [ ] `StageBoundaries` gains Z-axis fields sourced from `[PlayerInfo]` `topbound`/`botbound` (one value per stage, same cardinality as the existing `Left`/`Right`)
- [ ] Per-player `Startz` (P1-P8) is modeled in a separate type of its own (a per-player value, not a stage-wide one like `StageBoundaries` — do not bundle it into the same struct just because both are sourced from `[PlayerInfo]`)
- [ ] Field/doc-comment naming keeps the new Z-axis `StageBoundaries` bound and the new `[Scaling]` type's `topz`/`botz` clearly distinct — both are Z-related but are different concepts (movement clamp vs. perspective-scaling range); avoid reusing bare `Top`/`Bottom` identifiers across the two types, and update `bounds.go`'s existing doc comment (currently: "no vertical (top/bottom) movement bound... see decision 001") so it no longer reads as contradicted by the new Z fields — decision 001's Left/Right-only ruling is about the Y axis and still stands; Z is a separate axis, per decision 001's own 2026-08-09 update
- [ ] `Serialize`/`Document` round-trip the new sections/keys with the same guarantees item 003 already provides (fresh-write validity, comment/ordering-preserving byte-exact round trip for unmodified files)
- [ ] A stage `.def` with none of these sections parses/serializes identically to its pre-item-008 behavior (zero-value defaults, fully backward compatible)
- [ ] At least one real Ikemen GO 3D stage `.def` file is used to validate the field mapping, not invented from memory

## Notes
This repo never reads or decodes the referenced model/`.hdr` files themselves — that's `stage-viewer-web`/`stage-editor`'s job (their own backlog items, see roadmap decision `014`). Unlike `.sff`, the model format (glTF) is open and already tooled, so no dedicated parsing repo is needed here the way `sff` was extracted (decision `007`).
