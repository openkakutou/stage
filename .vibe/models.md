# Data models

## Stage
Root aggregate for a MUGEN/Ikemen GO stage. Zero-value valid (nil `Elements` ranges safely).

| Field | Type | Notes |
|---|---|---|
| Name | string | Display name, from `.def` `[Info]` "name" — empty if absent |
| Author | string | Author/creator, from `.def` `[Info]` "author" — empty if absent |
| BGdef | BGdef | Stage-level settings |
| Elements | []BGElement | BG elements/layers, in `.def` file order |
| Animations | map[int]BGAnimation | Every `[Begin Action N]` block, keyed by action number — nil when the stage declares none |
| CameraBounds | CameraBounds | Camera's own scroll clamp |
| StageBoundaries | StageBoundaries | Character x-movement clamp (and z-movement clamp, 3D-only) |
| Model | Model | 3D model placement/lighting (Ikemen GO extension, zero-valued for a 2D stage) |
| Scaling | Scaling | 3D perspective-scaling settings (Ikemen GO extension, zero-valued for a 2D stage) |
| PlayerStartZ | PlayerStartZ | Each player's starting depth position (Ikemen GO extension, zero-valued for a 2D stage) |
Defined in: `stage.go`

## BGdef
Stage-level settings sourced from `.def` `[BGDef]`/`[StageInfo]`/`[Camera]`.

| Field | Type | `.def` origin |
|---|---|---|
| SpriteFile | string | `[BGDef]` `spr` |
| LocalCoordWidth | int | `[StageInfo]` `localcoord` (1st value) |
| LocalCoordHeight | int | `[StageInfo]` `localcoord` (2nd value) |
| ZOffset | int | `[StageInfo]` `zoffset` |
| XScale, YScale | float64 | `[StageInfo]` `xscale`/`yscale` — default to 1 when `[StageInfo]` is present but omits them, stay at the Go zero value if the section is absent entirely |
| ZoomOut | float64 | `[Camera]` `zoomout` |
| ZoomIn | float64 | `[Camera]` `zoomin` |
| ModelFile | string | `[BGDef]` `model` (Ikemen GO 3D extension, empty for a 2D stage) |
| Near, Far, FOV, YShift | float64 | `[Camera]` `near`/`far`/`fov`/`yshift` (3D-only) |
Defined in: `bgdef.go`

## BGElement
One `[BG element_name]` section. `Type` selects which subset of fields is meaningful (Normal/Parallax use `Sprite`; Anim uses `ActionNumber`).

| Field | Type | `.def` origin |
|---|---|---|
| Name | string | `[BG <name>]` section header |
| Type | BGElementType | `type` |
| Sprite | SpriteRef | `spriteno` |
| ActionNumber | int | `actionno` |
| LayerNo | int | `layerno` |
| StartX, StartY | int | `start` |
| DeltaX, DeltaY | float64 | `delta` |
| TileX, TileY | int | `tile` |
| TileSpacingX, TileSpacingY | int | `tilespacing` |
Defined in: `bg_element.go`

## BGElementType
String enum matching literal `.def` "type" tokens: `BGElementNormal` ("normal"), `BGElementParallax` ("parallax"), `BGElementAnim` ("anim"). Zero value ("") is not a valid token — defaulting a missing key to "normal" is a parser concern, not encoded here.
Defined in: `bg_element.go`

## SpriteRef
Sprite address within the stage's sprite sheet.

| Field | Type | `.def` origin |
|---|---|---|
| Group | int | `spriteno` (1st value) |
| Image | int | `spriteno` (2nd value) |
Defined in: `bg_element.go`

## CameraBounds
The box the camera's own position is clamped to. Distinct type from StageBoundaries — see `.vibe/decisions/001-stage-boundaries-model-left-right-only.md`.

| Field | Type | `.def` origin |
|---|---|---|
| Left, Right | int | `[Camera]` `boundleft`/`boundright` |
| High, Low | int | `[Camera]` `boundhigh`/`boundlow` |
Defined in: `bounds.go`

## StageBoundaries
The x-range (and, 3D-only, z-range) characters may move within. No vertical (y-axis) fields — mainline MUGEN/Ikemen GO defines none.

| Field | Type | `.def` origin |
|---|---|---|
| Left, Right | int | `[PlayerInfo]` `leftbound`/`rightbound` |
| TopBound, BottomBound | float64 | `[PlayerInfo]` `topbound`/`botbound` (Ikemen GO 3D extension) |
Defined in: `bounds.go`

## Model
A model-based stage's 3D placement and lighting settings (Ikemen GO extension). The model file path itself is `BGdef.ModelFile`, not a field here.

| Field | Type | `.def` origin |
|---|---|---|
| OffsetX, OffsetY, OffsetZ | float64 | `[Model]` `offset` |
| ScaleX, ScaleY, ScaleZ | float64 | `[Model]` `scale` |
| Environment | string | `[Model]` `environment` |
| EnvironmentIntensity | float64 | `[Model]` `environmentintensity` |
Defined in: `model.go`

## Scaling
A model-based stage's 3D perspective-scaling settings (Ikemen GO extension): how on-screen size and Y offset change with depth.

| Field | Type | `.def` origin |
|---|---|---|
| DepthToScreen | float64 | `[Scaling]` `depthtoscreen` |
| TopZ, BottomZ | float64 | `[Scaling]` `topz`/`botz` |
| TopScale, BottomScale | float64 | `[Scaling]` `topscale`/`botscale` |
Defined in: `scaling.go`

## PlayerStartZ
Each of up to 8 players' starting depth (Z) position on a model-based stage (Ikemen GO extension). Kept separate from `StageBoundaries` since it is per-player, not stage-wide.

| Field | Type | `.def` origin |
|---|---|---|
| P1..P8 | int | `[PlayerInfo]` `p1startz`..`p8startz` |
Defined in: `bounds.go`

## BGAnimation
An animated element's (`BGElementAnim`) frame sequence — the `[Begin Action N]` block `ActionNumber` refers to. Mirrors `character`'s `air.Animation`. Populated by `Parse` into `Stage.Animations`, keyed by action number — see `.vibe/decisions/004-bg-animation-model-and-parallax-formula.md` and `.vibe/decisions/006-begin-action-parsing-frame-fields-and-serialize-scope.md`.

| Field | Type | Notes |
|---|---|---|
| Frames | []BGAnimFrame | Ordered frame sequence |
| LoopStart | int | Index playback loops back to after the full sequence plays once (0 = loop the whole sequence) |
Defined in: `animation.go`

## BGAnimFrame
One displayed frame within a `BGAnimation`.

| Field | Type | Notes |
|---|---|---|
| Sprite | SpriteRef | Which sprite this frame displays |
| Time | int | Ticks to hold this frame before advancing |
Defined in: `animation.go`

## Document
Write-path counterpart to `Parse`/`Serialize`: retains the exact source bytes `ParseDocument` read alongside the decoded data, so `Serialize` can reproduce an unmodified file byte-for-byte (comments, section ordering, unrecognized content included). Mutating `Stage` after parsing has no effect on `Serialize`'s output.

| Field | Type | Notes |
|---|---|---|
| Stage | Stage | Decoded the same way `Parse`'s return value is |
| source | []byte | Unexported; the raw bytes `Serialize` writes back verbatim |
Defined in: `document.go`
