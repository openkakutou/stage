---
status: todo
---
# Parse StageInfo xscale/yscale into BGdef

## Description
Real Ikemen GO stages (e.g. `Dengeki_Subway.def`, `xscale = .35` / `yscale = .35`) author BG sprite art larger than the stage's `localcoord` and rely on `[StageInfo]`'s `xscale`/`yscale` to scale it down at draw time. `parser.go`'s `"stageinfo"` case currently only captures `localcoord` and `zoffset` — the `xscale`/`yscale` keys are silently dropped, and `BGdef` has no field for them at all, so no consumer of this library can ever know a stage declared a non-default scale.

## Acceptance Criteria
- [ ] `BGdef` gets `XScale`/`YScale` float fields, defaulting to `1, 1` when the keys are absent from `[StageInfo]` — matching MUGEN/Ikemen's own default.
- [ ] `parser.go`'s `[StageInfo]` handling parses `xscale`/`yscale`, tolerant of the same real-world authoring quirks already handled for other numeric keys (see `.vibe/decisions/008-tolerant-parsing-for-real-file-authoring-habits.md`).
- [ ] `Document`/`SerializeDef` round-trip stays byte-exact on files that don't touch these keys, and correctly preserves/round-trips files that do.
- [ ] Verified against the local real-stage corpus (`STAGE_CORPUS_DIR`, see `.vibe/fixture-sources.md`): `Dengeki_Subway.def` (from `~/workspace/ikemen-quick-versus/stages/Dengeki_Subway/`) parses `xscale=.35, yscale=.35` correctly, and the full corpus scan (`TestCorpusCompat_RealDefFiles_ParseSuccessRate`) still passes 100%.
- [ ] `Dengeki_Subway.def` (or a trimmed fixture derived from it) is added as a vendored `testdata/` fixture covering a non-default `xscale`/`yscale`, mirroring how `mugen-2d-stage.def`/`ikemen-go-3d-model-stage.def` were sourced from the same corpus (see `testdata/README.md`).

## Notes
Found via a user-reported "no preview" bug in `stage-viewer-web` while testing this exact real file: without this field, every BG sprite in a hi-res stage like this one renders at ~8x its intended on-canvas area (1/0.35² ≈ 8.16x), filling/overflowing the visible canvas with a zoomed-in crop of one element while the rest land off-canvas — reading as a completely blank/broken preview.

This item only adds the data — it does not fix any rendering. It blocks a `stage-viewer-web` backlog item that applies the scale factor in that repo's own composition math once this field exists here.
