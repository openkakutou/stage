---
status: in_progress
depends_on: [001]
---
# Parse Stage `.def`

## Description
Implement a text-format parser reading a MUGEN/Ikemen GO stage `.def` file into the data model defined in item 001. Stage `.def` files share the same `[SectionName]`/`key = value` text layout `character`'s `.def`/`.cns` parsers already handle, but with stage-specific sections (`[Info]`, `[Camera]`, `[PlayerInfo]`, `[Bound]`, `[StageInfo]`, `[Shadow]`, `[Reflection]`, `[Music]`, and one `[BG <name>]` section per background element/layer). The parser must be MUGEN- and Ikemen GO-compatible, tolerating unrecognized sections/keys the same way `character`'s `def.Parse`/`cns.Parse` skip what they don't recognize rather than aborting the read.

## Acceptance Criteria
- [ ] `[Info]`, `[Camera]`, `[Bound]`/`[StageInfo]`-equivalent boundary sections, and one or more `[BG ...]` sections all parse into the item 001 model
- [ ] Both MUGEN-style and Ikemen GO-style stage `.def` syntax variants parse correctly
- [ ] Unrecognized sections/keys are skipped rather than causing the parse to fail
- [ ] A malformed line (bad header, non-numeric value where a number is required) returns a descriptive, line-numbered error instead of crashing or silently producing wrong data
- [ ] An empty file produces an explicit empty result, not an error, matching `character`'s `.air`/`.cns` parser behavior on empty input

## Notes
None.
