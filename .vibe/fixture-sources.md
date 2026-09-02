# Fixture sources for `.def` stage test data

Reference material and real-world files used to source or validate test
fixtures for this repo. Kept separate from the codebase index proper since
the local corpus below is a machine-specific resource unavailable in CI or
on other machines. Mirrors the sibling `sff` repo's own
`.vibe/fixture-sources.md` practice, adapted for this repo's own (text,
not binary) `.def` format — see backlog item 007.

## Local real-stage corpus (not referenced from code)

`~/workspace/ikemen-quick-versus/stages/` on the machine this repo is
usually developed on: a real Ikemen GO frontend install's `stages/`
directory — **58 real stage `.def` files** across a mix of classic MUGEN
ports and Ikemen GO-specific 3D model-based stages.

Available interactively for:

```sh
STAGE_CORPUS_DIR=~/workspace/ikemen-quick-versus/stages go test -run TestCorpusCompat -v .
```

`corpus_compat_test.go`'s `TestCorpusCompat_RealDefFiles_ParseSuccessRate`
is gated on the `STAGE_CORPUS_DIR` env var and skipped entirely when unset
— this repo's normal test run (CI included) never depends on this corpus
existing. Never hardcode this (or any other machine-specific corpus) path
into source, tests, or committed config.

## Corpus compatibility scan results (backlog item 007, 2026-09-02)

`TestCorpusCompat_RealDefFiles_ParseSuccessRate` scanned the full local
corpus above: **58 files**.

- **58 / 58 parsed successfully (100%).**
- **0 `Document` byte-exact round-trip failures.**
- **0 `SerializeDef` byte-exact-on-unmodified-save round-trip failures.**

Three real-file parse failures were found and fixed as part of this item
(4 real-world tolerance gaps across those 3 files — one file needed two
separate fixes):

| Shape | Files | Fix |
|---|---|---|
| A BG element's `tile`/`tilespacing` key given as a single bare value instead of the documented `"a,b"` pair | `Otherworldly Forest` (both keys, on two different elements) | Applied to both axes — `parseIntPairOrSingle` |
| `zoffset` written with a redundant decimal point (`555.0`) despite being an integer field | `The_Great_Cave_Offensive` | Falls back to parsing as a float and rounding — `parseIntTolerant` |
| A `[Begin Action N]` frame line's `time` field left blank instead of a number | `XX'GARAGE'XX` | Defaults to `0` |

All three are real-world authoring habits (not corrupt files) — every
affected file is a real, working stage in the corpus. See
`.vibe/decisions/008-tolerant-parsing-for-real-file-authoring-habits.md`
for the full reasoning. Two of the vendored `testdata/` fixtures
(`mugen-2d-stage.def`, `ikemen-go-3d-model-stage.def`) are real files from
this same corpus — see `testdata/README.md`.
