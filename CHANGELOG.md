# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.11.1] - 2026-09-02

### Fixed

- Now uses the latest version of the underlying sprite sheet library, which fixes two real-file loading bugs: a sprite sheet's last sprite could fail to load entirely when it reused an earlier sprite's palette, and legitimate extreme-aspect-ratio sprites were rejected as implausibly large.

## [0.11.0] - 2026-09-02

### Added

- Consuming apps can now resolve a stage's background sprites into actual displayable pixels directly through the WASM build, for as many sprites as needed in a single call, with support for previewing an alternate color palette. A sprite that doesn't exist, or a corrupted sprite sheet, is reported with a clear error instead of breaking the preview.

### Fixed

- Loading is now validated against a large corpus of real MUGEN/Ikemen GO stage files, and three real-file loading bugs this uncovered are fixed: a background layer's tiling settings written as a single shorthand value, a stage's ground-level setting written with a decimal point, and a background animation frame with a blank timing value all now load correctly instead of being rejected.

## [0.10.0] - 2026-08-29

### Added

- Consuming apps can now resolve which frame an animated background element should currently show, for as many elements as needed in a single call, without reimplementing the frame-timing logic themselves.

## [0.9.0] - 2026-08-28

### Added

- Animated background elements now get their actual frame sequence read from the stage file, wherever in the file it's declared, instead of only carrying a reference nothing resolves. Saving a stage writes that frame data back out for every animated element that still uses it. A malformed frame line or animation block header shows a clear error naming the exact line instead of failing silently or corrupting the rest of the stage.

## [0.8.0] - 2026-08-26

### Added

- A stage's name and author (when the file sets them) are now read and kept when the file is loaded or saved.

## [0.7.0] - 2026-08-25

### Added

- A stage can now be loaded and saved from a web browser, without needing Go installed: a WebAssembly build of this library is published as a downloadable file on every release. Saving an unedited stage reproduces the original file exactly; saving an edited one writes out the changes.

## [0.6.0] - 2026-08-25

### Added

- Scrolling background layers now compute their real on-screen position from the camera's position and their configured scroll speed, giving parallax (depth) layers the correct offset as the camera moves. Animated background layers now resolve which sprite frame should currently be shown from elapsed time, including looping back once the sequence finishes, instead of only being able to show a single static image.

## [0.5.0] - 2026-08-23

### Added

- Stages can now use Ikemen GO's 3D model-based extension: a stage can reference a 3D model file with its own placement, scaling, and lighting settings, plus depth-based (Z-axis) camera perspective, character movement limits, and per-player starting positions. A stage that doesn't use any of this reads and writes exactly as before.

## [0.4.0] - 2026-08-20

### Added

- Stage background layers can now resolve their sprite reference to real pixel data from a loaded `.sff` sprite sheet, working with either the older or newer `.sff` file format transparently. A layer referencing a sprite missing from the sheet is reported with a clear error instead of silently showing nothing.

## [0.3.0] - 2026-08-19

### Added

- Stages can now be written back out as MUGEN/Ikemen GO `.def` text, ready for a `stage-editor` save: a fresh-write path producing valid, readable output equivalent to the original data, and a format-preserving path that reproduces an unmodified file's comments, section ordering, and unrecognized sections byte-for-byte instead of overwriting them.

## [0.2.0] - 2026-08-18

### Added

- MUGEN/Ikemen GO stage `.def` files can now be read into the stage data model: camera scroll limits, character movement limits, sprite sheet and coordinate-space settings, and every background layer (static, parallax, or animated), including Ikemen GO's tiling extension. Unrecognized sections and keys are tolerated rather than rejected, and a malformed file reports a clear, line-numbered error instead of crashing.

## [0.1.0] - 2026-08-09

### Added

- Defined the stage data model (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) that will represent MUGEN/Ikemen GO stage backgrounds — layers, camera scroll limits, and character movement limits — once `.def` stage file reading is implemented

[Unreleased]: https://github.com/openkakutou/stage/compare/v0.11.1...HEAD
[0.11.1]: https://github.com/openkakutou/stage/compare/v0.11.0...v0.11.1
[0.11.0]: https://github.com/openkakutou/stage/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/openkakutou/stage/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/openkakutou/stage/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/openkakutou/stage/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/openkakutou/stage/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/openkakutou/stage/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/openkakutou/stage/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/openkakutou/stage/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/openkakutou/stage/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/openkakutou/stage/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/openkakutou/stage/releases/tag/v0.1.0
