# testdata

Real test fixtures for this repo's `.def` stage parser (backlog item 007).

## Source

Both files are vendored, unmodified real-world stage `.def` files from a
local Ikemen GO frontend install's `stages/` directory (see
`.vibe/fixture-sources.md`'s "Local real-stage corpus" section for where
that corpus lives and how to point the corpus-scan test at it). Unlike
`sff`'s own `testdata/files/*.sff`, these are not trimmed down — a real
stage `.def` is already small plain text (under 100 lines), so the whole
file is vendored as-is rather than extracting a minimal subset.

## `mugen-2d-stage.def`

`The_Great_Cave_Offensive.def` — a real MUGEN 1.1 stage (`mugenversion =
1.1`, no Ikemen GO 3D extension: `BGdef.ModelFile` is empty). Exercises a
real-world tolerance shape this item's own corpus scan found and fixed:
`zoffset = 555.0` — a decimal value for what's declared as an integer
field.

## `ikemen-go-3d-model-stage.def`

`ggxrd-neonewyork.def` ("Guilty Gear Xrd - Neo New York") — a real Ikemen
GO 3D model-based stage: `[Model]` section, `BGdef.ModelFile`
(`ggxrd-neonewyork.glb`), a 3D camera `fov`, and a `[Begin Action 9000]`
stage-portrait animation block composing correctly alongside the 3D-only
sections.

## `mugen-nondefault-scale-stage.def`

`Dengeki_Subway.def` ("Dengeki_Military Subway", `mugenversion = 1.1`, no
Ikemen GO 3D extension) — a real stage authoring hi-res BG sprite art
larger than its `localcoord` and relying on `[StageInfo]`'s `xscale =
.35` / `yscale = .35` to scale it down at draw time (backlog item 012).
The leading-dot decimal shape (`.35`, no leading zero) is exactly how the
real file writes it. Also exercises a `[Begin Action 1]` block alongside
a plain 2D `[StageInfo]` (no 3D extension), and a lowercase `[BGdef]`
section header.

## Regenerating / adding a new fixture

There is no dedicated trimming tool the way `sff`'s `testdata/gen` is —
copy the real `.def` file directly from a local corpus (see
`.vibe/fixture-sources.md`) into this directory. Never invent or hand-edit
the copied content; if a scenario needs bytes a real file doesn't already
have, that scenario belongs in a hand-built synthetic fixture (inline in
the relevant `_test.go` file) instead, not a fabricated "real" one here.
