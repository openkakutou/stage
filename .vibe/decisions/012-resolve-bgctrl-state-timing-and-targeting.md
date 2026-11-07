---
date: 2026-11-07
status: accepted
---
# ResolveBGCtrlState folds controllers per group tick with ID targeting

**Context:** Backlog item 016 needs the per-element state (visible, enabled, offset, anim action) that BGCtrl controllers produce at a tick, for MUGEN 1.0/1.1 and Ikemen GO stages alike. `BGElement` carried no `id`, so `ctrlID` could not target anything.

**Decision:** `BGElement.ID` (json `id`, from the BG block's `id` key, written only when nonzero). `ResolveBGCtrlState(s, tick)` returns one `BGLayerState` per element, in `s.Elements` order:
1. Group tick `t = tick mod LoopTime` when `LoopTime > 0`, else `t = tick`; a negative tick is 0. Groups are positional (names repeat) and independent.
2. Effective target ID = controller `CtrlID` if nonzero, else group `CtrlID`. **Assumption:** effective ID 0 targets every element. Otherwise only elements with that `ID`.
3. Single `time` T: in effect once `t >= T`, state persists. The controllers in effect are folded in chronological order of when they last fired (a list window's start is shifted to the current loop period), file order breaking ties, so the latest one overrides earlier ones for the same attribute, wherever it sits in the file. Groups are folded in file order. Ordering by last firing is unverified against either engine (neither engine source was read); it is the natural reading of "the later event wins".
4. `time` list `start,end,loop` (kept raw in Params): in effect while `start <= t' < end`, `t' = t mod loop` if `loop > 0`; omitted end = no end.
5. `visible`/`enabled`/`anim` read `value` (nonzero = true; anim truncates to int). `posset` sets the base offset, `posadd` adds to it (`x`,`y`, missing = 0).
6. Velocity: per group, `velset` replaces and `veladd` adds to the velocity in effect at each tick; the offset gains the sum of the velocity over ticks `0..t-1` (so at `t = T` the displacement is 0, at `T+1` it is one velocity step), added on top of the base position. Velocity folds independently per group, from the controllers in effect at each tick in the same chronological order (rule 3). A `posset` restarts the position: only the displacement from the start of the latest `posset` in the group counts (`velset` x=1 from 0, `posset` x=0 at 50: offset 10 at tick 60). The loop part of a velocity time list is honoured (the window repeats every `loop` ticks; unverified). Shortcut: the number of integration segments grows with `t/loop` for a looped velocity list in a group without `looptime`.
7. Other types (palfx, animtrans, unknown) and unparsable values are ignored silently.
8. WASM `resolveBGCtrlState({stage, ticks})` resolves many ticks in one call, like decision `007`.

**Engine divergences** (no MUGEN or Ikemen GO documentation is vendored in this repo, so anything not visible in the fixtures is marked unverified):

| Behavior | MUGEN | Ikemen GO | Chosen |
|---|---|---|---|
| `time` as a list `start,end,loop` | documented in the 2002 MUGEN stage docs (community copy, see Sources) | Ikemen GO wiki shows list examples (`time = 0, 60`) | support both: single value and list |
| Controller type spelling `enabled` vs `enable` | unverified | unverified; the real fixture only shows `Visible` (no enable controller), so both spellings are accepted defensively | accept both |
| `looptime <= 0` | unverified | unverified | no loop (user-approved) |
| `ctrlID` list / multiple IDs | a list, up to 10 ids; a controller's list overrides the group's (2002 docs) | Ikemen GO adds `sctrlid` (an id for ModifyBGCtrl, not for targeting) | NOT modeled yet: the model holds one integer; follow-up backlog item 019 |
| Effective ID 0 | unverified | unverified | targets all elements (assumption) |
| End bound of a time list (inclusive vs exclusive) | unverified | unverified | exclusive (user-approved) |
| Velocity integration across loop wrap | unverified | unverified | restarts at 0 on each loop |

**Overflow:** a summed offset that overflows float64 is clamped to ±MaxFloat64 (NaN to 0), so the WASM JSON encoding never fails the whole call.

**Sources:** 2002 MUGEN stage docs, community copy (https://www.geocities.ws/samus1978/help/8.html); Ikemen GO wiki, Background features (https://github.com/ikemen-engine/Ikemen-GO/wiki/Background-features), which also notes MUGEN 1.1 ignores the BGCtrlDef looptime for `SinX`/`SinY` (not resolved here: `sinx`/`siny` are ignored types). Rows still marked unverified were not found in these sources.

**Real-file observation:** `testdata/mugen-real-bgctrl-visible-stage.def` (unmodified community stage, "Pokemon Gym") has 50 BG elements with `ID = 1..50` (key spelled `ID`), and 50 groups all named `Back` (the name repeats, so groups are positional) each with lowercase `ctrlid = N` and `looptime = 201` on the group, holding two controllers written `type = Visible` (capitalised, hence the lowercased `Type`) with `time = 4(N-1)` value 1 then `time = 4N` value 0 and a repeated `ctrlid = N`. It only uses the single-value `time`; no `enable`, `anim` controller or time list occurs. Tests pin hand-derived values from it.

**Reason:** Pure, deterministic, total, easy to test against hand-derived values; every unverified row is the simplest reading, easy to change in one place.

**Rejected alternatives:** Per-tick simulation (cost grows with tick); one global velocity across groups (groups loop independently, no single clock); modeling `ctrlID` lists (parser change outside this item).
