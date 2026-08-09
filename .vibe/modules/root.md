# Module: root
**Role:** Package `stage` at the module root — the library version constant plus the stage read-path data model: `Stage`, `BGdef`, `BGElement` (and its `BGElementType`/`SpriteRef` support types), `CameraBounds`, `StageBoundaries`. `.def` parse/serialize, `sff` sprite integration, and BG element frame/animation resolution will land here as the backlog (items 002-006) is worked.
**Files:** `version.go`, `bgdef.go`, `bg_element.go`, `bounds.go`, `stage.go`
**Exports:**
- `Version` (string constant)
- `Stage` — root aggregate: `BGdef`, `Elements []BGElement`, `CameraBounds`, `StageBoundaries`
- `BGdef` — stage-level settings (`SpriteFile`, `LocalCoordWidth`/`LocalCoordHeight`, `ZOffset`, `ZoomOut`/`ZoomIn`)
- `BGElement` — one BG layer (`Name`, `Type`, `Sprite`, `ActionNumber`, `LayerNo`, `StartX`/`StartY`, `DeltaX`/`DeltaY`, `TileX`/`TileY`, `TileSpacingX`/`TileSpacingY`)
- `BGElementType` (string enum) with constants `BGElementNormal`, `BGElementParallax`, `BGElementAnim`
- `SpriteRef` — `{Group, Image int}` sprite address
- `CameraBounds` — `{Left, Right, High, Low int}`, the camera's own scroll clamp
- `StageBoundaries` — `{Left, Right int}`, character x-movement clamp (distinct type from `CameraBounds` — see `.vibe/decisions/001`)
**Depends on:** nothing yet (stdlib-only for now; `github.com/openkakutou/sff` arrives with backlog item 004)
