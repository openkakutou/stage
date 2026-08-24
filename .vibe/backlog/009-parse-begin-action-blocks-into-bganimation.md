---
status: todo
depends_on: [005]
---
# Parse `[Begin Action N]` Blocks Into `BGAnimation`

## Description
An animated BG element (`BGElementAnim`) carries an `ActionNumber`, but nothing yet resolves it to real frame data: `Parse` doesn't read the stage `.def` file's `[Begin Action N]` blocks (the same `.air`-syntax format character animations use, confirmed via MUGEN's own stage documentation), and `BGAnimation`/`BGAnimFrame` (added by item 005) are currently only ever built by hand in memory. This item wires `[Begin Action N]` blocks into `Parse`/`Serialize`, indexed by action number, so a `BGElementAnim`'s `ActionNumber` can actually be looked up against real parsed data.

## Acceptance Criteria
- [ ] `Parse` reads one or more `[Begin Action N]` blocks from `.def` text into `BGAnimation` values, keyed by their action number
- [ ] A `BGElementAnim`'s `ActionNumber` can be resolved to the matching parsed `BGAnimation` (and, from there, to `ResolveAnimationFrame`)
- [ ] `Serialize` writes `[Begin Action N]` blocks back out for every action number referenced by an anim element, re-parseable by `Parse` into an equivalent result
- [ ] A malformed frame line or action header returns a descriptive, line-numbered error, matching `Parse`'s existing error conventions for other sections

## Notes
Follow-up to item 005 (`.vibe/decisions/004-bg-animation-model-and-parallax-formula.md`), which deliberately scoped the frame-sequence data model and resolution logic (`BGAnimation`, `ResolveAnimationFrame`) separately from this parsing/wiring work, mirroring the "data model and resolution before parser" order `character`'s own `air`/`cns` packages were built in. `[Begin Action N]` blocks can appear immediately after their referencing `[BG <name>]` element or grouped elsewhere in the file (per MUGEN's own stage docs), so this cannot assume a fixed position relative to the element that references it.
