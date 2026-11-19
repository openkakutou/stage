---
status: done
---
# Camera Tension, Start and BG Scale

## Description
The stage library does not expose `[Camera]` `tension`, `startx`/`starty`, nor the BG elements' `scalestart`, `xscale` and `scaledelta`, and it does not resolve `zoffsetlink`. The native game (game#056) therefore uses Ikemen GO's default tension, draws scaled layers at the wrong size (Training's `scalestart = 2.4`), and falls back to a fixed ground line. Parse them as Ikemen GO does (stage.go, camera.go, commit `54d4148a`).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO compatibility (roadmap decision `038`).

## Acceptance Criteria
- [x] `[Camera]` tension, startx and starty, and each BG element's scalestart, xscale and scaledelta, are parsed with Ikemen GO's defaults, with the source cited
- [x] `zoffsetlink` is resolved to the linked BG element as Ikemen GO does (stage.go:1758)
- [x] The new fields survive the JSON shape the WebAssembly build returns

## Notes
Found by game#056 (2026-11-19). Stage used: quick-versus-bootstrap Training.
