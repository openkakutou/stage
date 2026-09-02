---
date: 2026-09-02
status: accepted
---
# Parse tolerates three real-world authoring habits found by the corpus scan

**Context:** Backlog item 007 added a real-file compatibility scan (58 real stage `.def` files, mirroring the sibling `sff` repo's own practice). It found 3 files that failed to parse, each on a distinct real-world shape: a BG element's `tile`/`tilespacing` key given as a single bare value instead of the documented `"a,b"` pair (found on two different elements in one file, for both keys); `zoffset` written with a redundant decimal point (`555.0`) despite being an integer field; and a `[Begin Action N]` frame line's `time` field left blank instead of a number. All three files are real, working stages from an actual Ikemen GO frontend install — not corrupt or hand-crafted adversarial input.

**Decision:** `Parse` now tolerates all three shapes instead of erroring:
1. `tile`/`tilespacing` accept a single bare value (no comma), applied to both axes — `parseIntPairOrSingle`.
2. `zoffset` falls back to parsing as a float and rounding to the nearest integer when a plain integer parse fails — `parseIntTolerant`. A genuinely non-numeric value still errors.
3. A `[Begin Action N]` frame line's blank `time` field defaults to `0` instead of erroring.

Each is scoped exactly to where real-world evidence was found: `parseIntPairOrSingle` is wired into `tile`/`tilespacing` specifically (the two keys actually observed with this shape), not applied speculatively to every other `"a,b"`-pair field (`delta`, `start`, `width`, …) with no evidence any of them share the habit; `parseIntTolerant` is wired into `zoffset` specifically, likewise.

**Reason:** Matches this repo's own established tolerance philosophy for `Parse` (already documented: unrecognized sections/keys skipped, non-key-value lines ignored, comments stripped) and `character`'s own precedent of accepting real-world authoring quirks (e.g. boolean/numeric header fields holding raw MUGEN trigger expressions) rather than rejecting files real games and tools already load successfully. A stage `.def` compatibility library that can't parse a real, working stage because of a cosmetic decimal point or a shorthand notation would be strictly less useful than the reference engines it's trying to match.

**Rejected alternatives:**
- **Applying the pair-or-single tolerance to every numeric pair field, not just the two observed**: rejected — no evidence exists for `delta`/`start`/`width`/etc. sharing this habit; extending the tolerance there would be guessing, not fixing an observed gap, and could silently mask a genuinely malformed value in a field where a single bare number was never a real author's intent.
- **Rejecting these 3 files as permanent, documented gaps** (the same treatment item 007's sibling `sff` gave its own remaining, genuinely-corrupt real files): rejected — unlike those, all three shapes here have an unambiguous, safe, narrowly-scoped interpretation (apply to both axes; round a clearly integer-valued float; default a blank field to its zero value), so fixing them is strictly better than documenting them as unsupported.
