---
status: done
depends_on: [003]
---
# Fixture-Driven Compatibility Testing Against Real MUGEN/Ikemen Stage Files

## Description
Extend the real-file compatibility testing practice `character` established (and `sff` is replicating in its own item 005) to this repo's `.def` stage parser: validate parse/serialize round trips against a corpus of real, unmodified MUGEN 1.0/1.1 and Ikemen GO stage files, not just hand-built synthetic fixtures — `character`'s own history (see its `.vibe/fixture-sources.md`) shows synthetic-only fixtures repeatedly missed real-file edge cases (e.g. two separate `.sff` v1 decoding bugs items 028/034 only caught this way).

## Acceptance Criteria
- [x] A documented, gitignored-path fixture corpus is used to validate parse success rate across real stage `.def` files (`STAGE_CORPUS_DIR`-gated `TestCorpusCompat_RealDefFiles_ParseSuccessRate`, mirroring `sff`'s own `corpus_compat_test.go`/`SFF_CORPUS_DIR` convention — see `.vibe/fixture-sources.md`)
- [x] Trimmed real fixtures that exercise scenarios synthetic test data doesn't cover are vendored into `testdata/` (`mugen-2d-stage.def`, `ikemen-go-3d-model-stage.def` — see `testdata/README.md`)
- [x] Any real-file parse failures are triaged and either fixed or explicitly documented as a known gap, never silently ignored — all 3 real files found by the scan were root-caused and fixed (4 real-world tolerance gaps total, one file needed two fixes); see `.vibe/decisions/008`
- [x] At least one MUGEN-only and one Ikemen-GO-only stage `.def` syntax variant is represented in the vendored corpus — `mugen-2d-stage.def` (MUGEN 1.1, no 3D) and `ikemen-go-3d-model-stage.def` (Ikemen GO 3D model-based stage)

## Notes
Mirrors `character`'s `.vibe/fixture-sources.md` practice — read that file for the exact convention to replicate: local, machine-specific corpus paths (e.g. a full Ikemen GO frontend install's `stages/` directory) must never be hardcoded into source, tests, or committed config; only the trimmed, vendored `testdata/` result is committed.
