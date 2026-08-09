# Data models

## Stage
Root aggregate for a MUGEN/Ikemen GO stage. Zero-value valid (nil `Elements` ranges safely).

| Field | Type | Notes |
|---|---|---|
| BGdef | BGdef | Stage-level settings |
| Elements | []BGElement | BG elements/layers, in `.def` file order |
| CameraBounds | CameraBounds | Camera's own scroll clamp |
| StageBoundaries | StageBoundaries | Character x-movement clamp |
Defined in: `stage.go`

## BGdef
Stage-level settings sourced from `.def` `[BGDef]`/`[StageInfo]`/`[Camera]`.

| Field | Type | `.def` origin |
|---|---|---|
| SpriteFile | string | `[BGDef]` `spr` |
| LocalCoordWidth | int | `[StageInfo]` `localcoord` (1st value) |
| LocalCoordHeight | int | `[StageInfo]` `localcoord` (2nd value) |
| ZOffset | int | `[StageInfo]` `zoffset` |
| ZoomOut | float64 | `[Camera]` `zoomout` |
| ZoomIn | float64 | `[Camera]` `zoomin` |
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
The x-range characters may move within. No vertical fields — mainline MUGEN/Ikemen GO defines none.

| Field | Type | `.def` origin |
|---|---|---|
| Left, Right | int | `[PlayerInfo]` `leftbound`/`rightbound` |
Defined in: `bounds.go`
