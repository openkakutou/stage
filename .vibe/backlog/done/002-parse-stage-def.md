---
status: done
depends_on: [001]
---
# Parse Stage `.def`

## Description
Implement a text-format parser reading a MUGEN/Ikemen GO stage `.def` file into the data model defined in item 001. Stage `.def` files share the same `[SectionName]`/`key = value` text layout `character`'s `.def`/`.cns` parsers already handle, but with stage-specific sections (`[Info]`, `[Camera]`, `[PlayerInfo]`, `[Bound]`, `[StageInfo]`, `[Shadow]`, `[Reflection]`, `[Music]`, and one `[BG <name>]` section per background element/layer). The parser must be MUGEN- and Ikemen GO-compatible, tolerating unrecognized sections/keys the same way `character`'s `def.Parse`/`cns.Parse` skip what they don't recognize rather than aborting the read.

## Acceptance Criteria
- [x] `[Info]`, `[Camera]`, `[Bound]`/`[StageInfo]`-equivalent boundary sections, and one or more `[BG ...]` sections all parse into the item 001 model
- [x] Both MUGEN-style and Ikemen GO-style stage `.def` syntax variants parse correctly
- [x] Unrecognized sections/keys are skipped rather than causing the parse to fail
- [x] A malformed line (bad header, non-numeric value where a number is required) returns a descriptive, line-numbered error instead of crashing or silently producing wrong data
- [x] An empty file produces an explicit empty result, not an error, matching `character`'s `.air`/`.cns` parser behavior on empty input

## Notes
None.

## Implementation note (2026-08-18)
`[Info]` and `[Bound]` have no corresponding field on the `Stage` model (per `.vibe/decisions/001`, boundary data comes from `[PlayerInfo]`, not a literal `[Bound]` section) — both are tolerated the same way any other unmapped section is: recognized without erroring, contributing no data. See `docs/api.md` for the exact section-to-field mapping implemented.
