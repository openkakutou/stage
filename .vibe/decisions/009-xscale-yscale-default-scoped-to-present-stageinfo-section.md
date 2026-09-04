---
date: 2026-09-04
status: accepted
---
# XScale/YScale's default-to-1 is scoped to a present "[StageInfo]" section, not applied when the section is absent entirely

**Context:** Backlog item 012 adds `BGdef.XScale`/`YScale`, which MUGEN/Ikemen GO themselves default to 1 (no scaling) when a stage author omits `xscale`/`yscale` from `[StageInfo]`. An existing test, `TestParse_EmptyInput_ReturnsZeroValueStageAndNoError`, already locks in that `Parse` on input with no recognized sections at all returns a fully zero-value `Stage` (`s.BGdef != (BGdef{})` fails the test) — and more generally, every other `BGdef` field (`ZOffset`, `ZoomOut`, `ZoomIn`, ...) only ever gets a real MUGEN-side default applied when its own section is actually present in the source.

**Decision:** The `xscale`/`yscale` default of 1 is applied the moment `Parse` enters a `[StageInfo]` section (before reading that section's own key=value lines), not unconditionally at the start of `Parse`. A file with a `[StageInfo]` section that omits `xscale`/`yscale` gets `BGdef.XScale`/`YScale` = 1/1; a file with no `[StageInfo]` section at all leaves them at the Go zero value (0/0), consistent with how the rest of `BGdef` behaves when its own section is missing.

**Reason:** Keeps the new fields consistent with every other `BGdef` field's existing behavior (a real-world default only applies once the section it lives in is actually read) and avoids silently rewriting the well-established, explicitly tested zero-value contract (`TestParse_EmptyInput_ReturnsZeroValueStageAndNoError`) for a scenario (a stage `.def` with no `[StageInfo]` section at all) that essentially never occurs in real files — `localcoord` alone makes the section all but mandatory in practice.

**Rejected alternatives:**
- **Defaulting XScale/YScale to 1 unconditionally at the top of `Parse`, regardless of whether `[StageInfo]` ever appears**: rejected — breaks the existing zero-value-on-no-recognized-sections guarantee for no real-world benefit, since every actual stage file in the local 58-file corpus already has a `[StageInfo]` section.
- **Defaulting `BGdef{}`'s own zero value to 1/1** (e.g. via a constructor): rejected — no other `BGdef` field carries a "real" MUGEN default in its Go zero value either (`ZoomOut`/`ZoomIn` zero to 0, not their own MUGEN defaults); introducing that only for `XScale`/`YScale` would be an inconsistent, one-off special case.
