---
date: 2026-08-23
status: accepted
---
# 3D stage settings: model file lives in BGdef, not the [Model] section

**Context:** Backlog item 008 assumed the new model reference type would
carry the model's file path itself, e.g. "`[Model]` section parses into a
new model reference type: file path, Offset...". Verified against Ikemen
GO's actual source (`src/stage.go` on GitHub, 2026-08): the model's file
path is read from `[BGdef]`'s own `"model"` key (`sec.LoadFile("model",
...)`), mirroring how `spr` already holds the 2D sprite sheet path there —
`[Model]` itself only carries placement/lighting settings (`offset`,
`scale`, `environment`, `environmentintensity`) for whichever file
`[BGdef]`'s `model` key names. The same source also confirms `[Model]` is
read as a single stage-wide block, never a repeatable `[Model name]`
section, resolving the item's own open acceptance criterion. It further
shows `topbound`/`botbound` and every `[Scaling]` key are read as floats
(`ReadF32`), not integers, and `pNstartz` as an integer (`ReadI32`).

**Decision:** `BGdef` gains `ModelFile` (mirroring `SpriteFile`) plus the
3D camera settings (`Near`/`Far`/`FOV`/`YShift`, alongside the existing
`ZoomOut`/`ZoomIn`). A new `Model` type holds only placement/lighting
(`OffsetX/Y/Z`, `ScaleX/Y/Z`, `Environment`, `EnvironmentIntensity`). A new
`Scaling` type holds the `[Scaling]` section as float64 fields.
`StageBoundaries` gains `TopBound`/`BottomBound` (float64, from
`topbound`/`botbound`) — named to stay clearly distinct from `Scaling`'s
`TopZ`/`BottomZ`, per item 008's own naming caution. A new `PlayerStartZ`
type (P1..P8, int) holds the per-player `pNstartz` values, kept separate
from `StageBoundaries` since it is a per-player value, not a stage-wide
one. The fresh-write `Serialize` path only emits `[Model]`, `[Scaling]`,
the 3D `[Camera]` keys, and the Z-axis `[PlayerInfo]` keys when
`BGdef.ModelFile` is non-empty, so a 2D-only stage serializes byte-for-byte
as it did before this item, keeping the model's presence as the single
source of truth for whether a stage is 3D.

**Reason:** Matches decision 001's own precedent — accuracy against the
real format outweighs literal adherence to the backlog's own (in this case
slightly imprecise) prose, and every field must still cite an unambiguous
`.def` key origin. Gating fresh-write output on `ModelFile` is the
simplest rule that satisfies the item's own backward-compatibility
acceptance criterion without inventing a separate "is 3D" flag that could
drift out of sync with the model file's actual presence.

**Rejected alternatives:**
- *Carry the file path inside the `Model` type itself, as the item's prose
  originally assumed* — rejected: contradicts the real `[BGdef]` "model"
  key confirmed in Ikemen GO's own source; would also break the existing
  `SpriteFile`-in-`BGdef` precedent for where a stage's asset file
  references live.
- *A separate `Is3D`/`Model3D bool` flag gating serialization* —
  rejected: redundant with `ModelFile`'s own emptiness, a second source of
  truth that could disagree with it.
- *Model `topbound`/`botbound`/`[Scaling]` fields as `int`, matching
  `StageBoundaries.Left`/`Right`'s existing type* — rejected: the real
  engine reads them as floats (`ReadF32`); modeling them as `int` would
  silently truncate legitimate fractional values.
