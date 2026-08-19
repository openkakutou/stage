# Module: root
**Role:** Package `stage` at the module root — the library version constant, the stage read-path data model (`Stage`, `BGdef`, `BGElement` and its `BGElementType`/`SpriteRef` support types, `CameraBounds`, `StageBoundaries`), the `.def` read-path parser (`Parse`), the `.def` write path (a fresh-write `Serialize` plus a format-preserving `Document`/`ParseDocument` pair, mirroring `character`'s own `def` package split), and now sprite resolution (`SpriteResolver`) against the external `sff` module. BG element frame/animation resolution will land here as the backlog (item 005) is worked.
**Files:** `version.go`, `bgdef.go`, `bg_element.go`, `bounds.go`, `stage.go`, `parser.go`, `serializer.go`, `document.go`, `sprite_resolver.go`
**Exports:**
- `Version` (string constant)
- `Stage` — root aggregate: `BGdef`, `Elements []BGElement`, `CameraBounds`, `StageBoundaries`
- `BGdef` — stage-level settings (`SpriteFile`, `LocalCoordWidth`/`LocalCoordHeight`, `ZOffset`, `ZoomOut`/`ZoomIn`)
- `BGElement` — one BG layer (`Name`, `Type`, `Sprite`, `ActionNumber`, `LayerNo`, `StartX`/`StartY`, `DeltaX`/`DeltaY`, `TileX`/`TileY`, `TileSpacingX`/`TileSpacingY`)
- `BGElementType` (string enum) with constants `BGElementNormal`, `BGElementParallax`, `BGElementAnim`
- `SpriteRef` — `{Group, Image int}` sprite address
- `CameraBounds` — `{Left, Right, High, Low int}`, the camera's own scroll clamp
- `StageBoundaries` — `{Left, Right int}`, character x-movement clamp (distinct type from `CameraBounds` — see `.vibe/decisions/001`)
- `Parse(r io.Reader) (Stage, error)` — reads `.def` stage text into a `Stage`; recognizes `[BGDef]`/`[StageInfo]`/`[Camera]`/`[PlayerInfo]`/`[BG <name>]`, tolerates any other section/key, defaults a BG element with no `type` key to `BGElementNormal`, and returns a line-numbered error for a malformed section header or a non-numeric value on a key that requires one (see `docs/api.md`)
- `Serialize(w io.Writer, s Stage) error` — writes a `Stage` to `w` as fresh `.def` text (`[BGDef]`/`[StageInfo]`/`[Camera]`/`[PlayerInfo]` then one `[BG <name>]` per element), re-parseable by `Parse` into an equivalent `Stage`; not a byte-exact round trip of any original file (see `Document`)
- `Document`/`ParseDocument(r io.Reader) (*Document, error)` — the format-preserving write path: `Document.Stage` holds the decoded data, `(*Document).Serialize(w io.Writer) error` reproduces the exact source `ParseDocument` read, byte-for-byte, as long as `Stage` is left unmodified
- `SpriteResolver`/`NewSpriteResolver(groups []sff.SpriteGroup) *SpriteResolver` — indexes sprite groups loaded via `sff.Load` by `(Group, Image)`; `(*SpriteResolver).Resolve(ref SpriteRef) (sff.Sprite, error)` looks one up, returning a descriptive error rather than a zero value for an unmatched reference, version-agnostic the same way `character`'s `air.SpriteResolver` is (see `.vibe/decisions/002`)
**Depends on:** `github.com/openkakutou/sff` (sprite sheet loading and resolution, arrived with backlog item 004)
