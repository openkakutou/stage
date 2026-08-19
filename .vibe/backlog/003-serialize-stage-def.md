---
status: in_progress
depends_on: [002]
---
# Serialize Stage `.def`

## Description
Add the write path for stage `.def` files: writing the item 001 data model back out to valid `.def` text (a semantic round trip through the item 002 parser), plus a format-preserving `Document`/`ParseDocument` pair guaranteeing a byte-exact round trip for *unmodified* files — the same two-tier guarantee `character` provides for `.def`/`.air`/`.cns` (fresh-write `Serialize` vs. comment/ordering-preserving `Document`). Without this, every save from `stage-editor` on a file the user hasn't actually edited would produce a noisy, unreadable Git diff, hurting community collaboration on stage files exactly the way `character`'s CLAUDE.md flags for character files.

## Acceptance Criteria
- [ ] `Serialize` writes a `Stage` back out to valid `.def` text, re-parseable by item 002's parser with equivalent data
- [ ] `Document`/`ParseDocument` reproduce an unmodified input file byte-for-byte, including comments and section ordering
- [ ] Round-trip test: parse → serialize → re-parse yields a semantically identical `Stage`
- [ ] Sections/keys the model doesn't recognize (per item 002's tolerance) are preserved by `Document`, not silently dropped, matching `character`'s `def.Document` guarantee
- [ ] Serializing a zero-value/empty `Stage` produces valid, re-parseable output rather than a malformed or panic-inducing file

## Notes
None.
