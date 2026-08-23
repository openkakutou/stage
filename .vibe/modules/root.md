# Module: root
**Role:** Package `stage` at the module root — the library version constant, the stage read-path data model (`Stage`, `BGdef`, `BGElement` and its `BGElementType`/`SpriteRef` support types, `CameraBounds`, `StageBoundaries`, and Ikemen GO's 3D stage extension: `Model`, `Scaling`, `PlayerStartZ`), the `.def` read-path parser (`Parse`), the `.def` write path (a fresh-write `Serialize` plus a format-preserving `Document`/`ParseDocument` pair, mirroring `character`'s own `def` package split), and sprite resolution (`SpriteResolver`) against the external `sff` module. BG element frame/animation resolution will land here as the backlog (item 005) is worked.
**Files:** `version.go`, `bgdef.go`, `bg_element.go`, `bounds.go`, `model.go`, `scaling.go`, `stage.go`, `parser.go`, `serializer.go`, `document.go`, `sprite_resolver.go`
**Exports:**
- `Version` (string constant)
- `Stage` — root aggregate: `BGdef`, `Elements []BGElement`, `CameraBounds`, `StageBoundaries`, `Model`, `Scaling`, `PlayerStartZ`
- `BGdef` — stage-level settings (`SpriteFile`, `LocalCoordWidth`/`LocalCoordHeight`, `ZOffset`, `ZoomOut`/`ZoomIn`, plus 3D-only `ModelFile`, `Near`/`Far`/`FOV`/`YShift`)
- `BGElement` — one BG layer (`Name`, `Type`, `Sprite`, `ActionNumber`, `LayerNo`, `StartX`/`StartY`, `DeltaX`/`DeltaY`, `TileX`/`TileY`, `TileSpacingX`/`TileSpacingY`)
- `BGElementType` (string enum) with constants `BGElementNormal`, `BGElementParallax`, `BGElementAnim`
- `SpriteRef` — `{Group, Image int}` sprite address
- `CameraBounds` — `{Left, Right, High, Low int}`, the camera's own scroll clamp
- `StageBoundaries` — `{Left, Right int, TopBound, BottomBound float64}`, character x-movement clamp plus (3D-only) its z-axis extension (distinct type from `CameraBounds` — see `.vibe/decisions/001`)
- `Model` — `{OffsetX/Y/Z, ScaleX/Y/Z, EnvironmentIntensity float64, Environment string}`, a model-based stage's 3D placement/lighting settings (Ikemen GO extension)
- `Scaling` — `{DepthToScreen, TopZ, BottomZ, TopScale, BottomScale float64}`, a model-based stage's 3D perspective-scaling settings
- `PlayerStartZ` — `{P1..P8 int}`, each player's starting depth (Z) position on a model-based stage, kept separate from `StageBoundaries` since it is per-player, not stage-wide
- `Parse(r io.Reader) (Stage, error)` — reads `.def` stage text into a `Stage`; recognizes `[BGDef]`/`[StageInfo]`/`[Camera]`/`[PlayerInfo]`/`[Model]`/`[Scaling]`/`[BG <name>]`, tolerates any other section/key, defaults a BG element with no `type` key to `BGElementNormal`, and returns a line-numbered error for a malformed section header or a non-numeric (or non-`a,b`/`a,b,c`) value on a key that requires one (see `docs/api.md`)
- `Serialize(w io.Writer, s Stage) error` — writes a `Stage` to `w` as fresh `.def` text (`[BGDef]`/`[StageInfo]`/`[Camera]`/`[PlayerInfo]` then one `[BG <name>]` per element, plus `[Model]`/`[Scaling]` and the 3D-only `Camera`/`PlayerInfo` keys only when `BGdef.ModelFile` is set), re-parseable by `Parse` into an equivalent `Stage`; not a byte-exact round trip of any original file (see `Document`)
- `Document`/`ParseDocument(r io.Reader) (*Document, error)` — the format-preserving write path: `Document.Stage` holds the decoded data, `(*Document).Serialize(w io.Writer) error` reproduces the exact source `ParseDocument` read, byte-for-byte, as long as `Stage` is left unmodified
- `SpriteResolver`/`NewSpriteResolver(groups []sff.SpriteGroup) *SpriteResolver` — indexes sprite groups loaded via `sff.Load` by `(Group, Image)`; `(*SpriteResolver).Resolve(ref SpriteRef) (sff.Sprite, error)` looks one up, returning a descriptive error rather than a zero value for an unmatched reference, version-agnostic the same way `character`'s `air.SpriteResolver` is (see `.vibe/decisions/002`)
**Depends on:** `github.com/openkakutou/sff` (sprite sheet loading and resolution, arrived with backlog item 004)
