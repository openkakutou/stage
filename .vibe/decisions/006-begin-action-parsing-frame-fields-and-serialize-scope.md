---
date: 2026-08-28
status: accepted
---
# `[Begin Action N]` parsing: only group/image/time populate `BGAnimFrame`, stored in a new `Stage.Animations` map, `Serialize` writes only referenced+resolvable blocks

**Context:** Backlog item 009 wires `[Begin Action N]` blocks (decision 004 deliberately deferred this) into `Parse`/`Serialize`, keyed by action number so a `BGElementAnim`'s `ActionNumber` can be resolved against real data. The underlying frame-line syntax is identical to `.air`'s own (`group,image,x,y,time[,flip[,blend]]`, confirmed by `character/air`'s own parser and MUGEN's stage docs), but `BGAnimFrame` (decision 004) only models `Sprite`/`Time` — no `X`/`Y`/`Flip`/`Blend` fields exist on it, and this repo may never depend on `character` (an explicit, standing constraint) so its parser code cannot be imported or reused directly.

**Decision:**
1. A frame line's `group`/`image`/`time` fields (positions 0, 1, 4) populate `BGAnimFrame`; any `x`/`y`/`flip`/`blend` fields present are read past (required for the minimum-5-fields shape check) but not stored anywhere, matching `BGAnimFrame`'s own already-decided scope — the same "read-path model can't hold everything yet" precedent, not a new gap this item introduces.
2. Parsed animations are stored in a new `Stage.Animations map[int]BGAnimation`, keyed by action number, `nil` when the stage declares none (mirroring `Stage.Elements`'s own nil-when-empty convention). `[Begin Action N]` headers are recognized in the same top-level bracket-line dispatch `[BG name]` headers already use, independent of element ordering — a block can appear anywhere in the file, per MUGEN's own stage docs and this item's own Notes.
3. `Serialize` writes one `[Begin Action N]` block per **distinct** action number actually referenced by a `BGElementAnim` element (in first-reference order, not map-iteration order, which Go randomizes) whose number also resolves against `Stage.Animations` — a referenced-but-unresolvable number (no matching entry) is silently skipped, the same graceful-lenience instinct `buildDrawPlan`-style consumers elsewhere in this org already apply to a dangling reference, rather than erroring on data this item never guaranteed was internally consistent.
4. Frame-line parsing is hand-written locally (not extracted into a shared helper with `character/air`), reusing only the pattern (five-plus comma fields, `strconv.Atoi` per field, `Loopstart` marker) — the two packages live in different, must-stay-independent repos.

**Reason:** Points 1 and 4 follow directly from constraints already fixed by earlier decisions (`BGAnimFrame`'s shape in 004, the no-`character`-dependency rule in this repo's own `CLAUDE.md`) — there was no real alternative to evaluate. Point 2's map keeps the new data reachable from `Stage` without touching `BGElement`'s own already-parser-wired fields. Point 3's referenced-only, ordered write keeps `Serialize`'s output deterministic and matches this item's own acceptance criterion ("for every action number referenced by an anim element"), rather than dumping every parsed block regardless of whether anything still uses it.

**Rejected alternatives:**
- *Store `X`/`Y`/`Flip`/`Blend` on `BGAnimFrame` too, to preserve them* — rejected: reopens decision 004's already-settled type shape for data no acceptance criterion here asks to carry, and no BG-animation consumer (`ResolveAnimationFrame`, `stage-viewer-web`) needs it.
- *Serialize every entry in `Stage.Animations`, regardless of whether any element still references it* — rejected: doesn't match the acceptance criterion's literal "referenced by an anim element" wording, and would silently resurrect a stale/orphaned block after an edit removed its only reference.
