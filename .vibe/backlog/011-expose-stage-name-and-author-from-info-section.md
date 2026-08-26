---
status: todo
depends_on: []
---
# Expose Stage Name And Author From `[Info]` Section

## Description
The `Stage` data model has no field for a stage's own identifying metadata (name, author) at all — `[Info]` is explicitly one of the sections `Parse` recognizes and tolerates but discards without contributing any data (see item 002's own implementation note). This blocks any consumer from displaying "which stage is this" the way `character`'s own `Name`/`Author` fields (sourced from its `.def` `[Info]` section) already let `character-viewer-web` do. Add a small `Info` struct (or equivalent fields on `Stage`) capturing `[Info]`'s `name` and `author` keys, mirroring `character.Character`'s `Name`/`Author` shape and JSON tags.

## Acceptance Criteria
- [ ] `Stage` (or a nested struct) exposes `Name`/`Author` string fields, JSON-tagged the same lower-camelCase way every other field is
- [ ] `Parse` reads `[Info]`'s `name`/`author` keys into these fields; a missing key leaves the corresponding field at its zero value (empty string), not an error
- [ ] `Serialize`/`SerializeDef` write `[Info]` back out with these values, preserving the existing "only recognized data round-trips" guarantee
- [ ] A stage `.def` with no `[Info]` section at all still parses successfully, with both fields empty
- [ ] The WASM entrypoint's `load`/`save` JSON contract carries the new field(s) automatically (no `cmd/wasm` code change needed if this lands on `Stage` itself, same as every other model field)

## Notes
Discovered while implementing `stage-viewer-web`'s own backlog item 003 (Characteristics Panel), whose acceptance criteria require displaying "the stage's name and author" — currently impossible, since no field carries this data anywhere in the pipeline. `stage-viewer-web`'s item 003 is blocked on this landing and a `stage` release publishing it (WASM pin bump), per the org's version-pinning policy (roadmap `.vibe/decisions/016`).
