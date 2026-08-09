---
status: in_progress
---
# Define Stage Data Model

## Description
Before any parsing can happen, this repo needs a pure-data model for what a MUGEN/Ikemen GO stage actually is, mirroring the "data model before parser" order `character` followed for `.air`/`.sff`/`.def`/`.cns`. This covers `BGdef` (stage-level settings: localcoord, camera zoom bounds, etc.), BG elements/layers (each with its own type — normal, parallax, anim — position, sprite reference, tiling, layer number), camera bounds (the playable/scrollable area), and stage boundaries (left/right/top/bottom edges characters and the camera are constrained to). This model is read-path only: it must not carry any parsing- or write-path-specific concerns (comment/ordering preservation), per the same design constraint `character` enforces on its own read-path types.

## Acceptance Criteria
- [ ] A `Stage` (or equivalent) root type composes `BGdef` settings, a list of BG elements/layers, camera bounds, and stage boundaries
- [ ] BG element type distinguishes at least normal, parallax, and animated backgrounds, with the fields each needs (sprite reference, position, tiling/delta, layer number)
- [ ] Camera bounds and stage boundaries are modeled as distinct concepts (camera scroll limits vs. character movement limits), not conflated into one
- [ ] A zero-value `Stage` is a valid, usable empty value (no nil-pointer panics on first use) — same expectation `character`'s read-path types meet
- [ ] Every field has a doc comment stating its MUGEN/Ikemen GO `.def` key origin, so the future parser (item 002) has an unambiguous mapping

## Notes
None.
