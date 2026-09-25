---
date: 2026-09-26
status: accepted
---
# Background music path lives on `Stage`, not `BGdef`

**Context:** Backlog item 013 adds the `.def` `[Music]` section's `bgmusic` file path to the data model. `BGdef` already holds two other file-reference fields, `SpriteFile` and `ModelFile`, both sourced from `[BGDef]`'s own keys (`spr`/`model` — see decision 003).

**Decision:** The new field is `Stage.MusicFile string` (`json:"musicFile"`), a top-level field on `Stage` documented as sourced from `[Music]`'s `bgmusic` key — mirroring how `Stage.Name`/`Stage.Author` already live directly on `Stage`, sourced from their own `[Info]` section, rather than being folded into `BGdef`.

**Reason:** `BGdef`'s own doc comment and every field it currently holds (`SpriteFile`, `ModelFile`, `LocalCoordWidth`/`Height`, `ZOffset`, `ZoomOut`/`ZoomIn`, `XScale`/`YScale`, the 3D camera settings) trace back to `[BGDef]`, `[StageInfo]`, or `[Camera]` — never a section of its own. `[Music]` is its own top-level `.def` section, structurally identical in that sense to `[Info]`, so the field belongs alongside `Name`/`Author` rather than inside `BGdef`, keeping `BGdef`'s existing "sourced from `[BGDef]`/`[StageInfo]`/`[Camera]`" scope accurate without a special case.

**Rejected alternatives:**
- **Add `MusicFile` to `BGdef`, alongside `SpriteFile`/`ModelFile`:** rejected — `SpriteFile`/`ModelFile` are both literally sourced from `[BGDef]`'s own keys; `bgmusic` is not, so placing it there would misrepresent its `.def` origin and break `BGdef`'s single-section-family scope for no benefit (unlike decision 003's `ModelFile`, there is no real-engine evidence the key actually lives under `[BGDef]`).
