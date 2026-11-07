# testdata

Test fixtures for this repo's `.def` stage parser (backlog item 007): real vendored files plus a few hand-written synthetic ones.

## Source

The real files below are vendored, unmodified real-world stage `.def` files from a
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

## `mugen-real-bgctrl-visible-stage.def`

`Pokemon Gym` by ExShadow, a real community stage copied unmodified from
the local Ikemen GO corpus (`.vibe/fixture-sources.md`'s "Local real-stage
corpus"). Real BG controllers (backlog item 016): 50 BG elements with `ID =
1..50`, and 50 `[BGCtrlDef Back]` groups (same name repeated) each with a
lowercase `ctrlid = N` and `looptime = 201`, holding two `type = Visible`
controllers (`time = 4(N-1)` value 1, `time = 4N` value 0). Used to pin
`ResolveBGCtrlState` against values derived by hand from the file text.

## Regenerating / adding a new fixture

There is no dedicated trimming tool the way `sff`'s `testdata/gen` is —
copy the real `.def` file directly from a local corpus (see
`.vibe/fixture-sources.md`) into this directory. Never invent or hand-edit
the copied content; if a scenario needs bytes a real file doesn't already
have, that scenario belongs in a hand-built synthetic fixture (inline in
the relevant `_test.go` file) instead, not a fabricated "real" one here.

## `mugen-bgctrl-stage.def`

`mugen-bgctrl-stage.def` and `mugen-bgctrl-targeting.def` are
**hand-written synthetic** fixtures, not real files. Unlike the real ones above, `mugen-bgctrl-stage.def` is a **hand-written synthetic** MUGEN-style
stage (not copied from a real corpus) exercising `[BGCtrlDef]` / `[BGCtrl]`
blocks: comments inside and between blocks, mixed-case headers and keys,
blank lines, a reused group name, a comma-list `time`, a quoted value and a
group with no controllers (backlog item 015).

## `mugen-bgctrl-targeting.def`

Also **hand-written synthetic**: BG `id`s, `ctrlID` targeting, group loops,
time lists and ignored controller types for `ResolveBGCtrlState` (backlog
item 016). Edge shapes the real file lacks (`enabled`, `posset`, `velset`,
`anim`, a `time` list) live here only.
