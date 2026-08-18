# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- MUGEN/Ikemen GO stage `.def` files can now be read into the stage data model: camera scroll limits, character movement limits, sprite sheet and coordinate-space settings, and every background layer (static, parallax, or animated), including Ikemen GO's tiling extension. Unrecognized sections and keys are tolerated rather than rejected, and a malformed file reports a clear, line-numbered error instead of crashing.

## [0.1.0] - 2026-08-09

### Added

- Defined the stage data model (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) that will represent MUGEN/Ikemen GO stage backgrounds — layers, camera scroll limits, and character movement limits — once `.def` stage file reading is implemented

[Unreleased]: https://github.com/openkakutou/stage/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/openkakutou/stage/releases/tag/v0.1.0
