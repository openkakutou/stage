---
status: done
---
# Expose Background Music Path

## Description
`[Music]` is a recognized `.def` section this library's parser currently skips outright (`parser.go:43`, `"carries nothing this model"`) — a stage's background-music file reference is read by nothing and exposed nowhere. Add it to the data model so consuming apps (starting with `mode-quick-versus`) can actually play a stage's music.

## Acceptance Criteria
- [x] The parsed stage data model carries the `[Music]` section's background-music file path (typically `.mp3`/`.ogg`, referenced relative to the stage's own folder, same convention as its sprite sheet reference)
- [x] A stage `.def` with no `[Music]` section (or an empty path) parses successfully with the field left empty, not an error
- [x] The WASM `load` JSON contract exposes this field the same way every other stage field already is
- [x] Round-trip serialization preserves an unedited `[Music]` path byte-for-byte, matching this library's existing format-preservation guarantee for untouched sections

## Notes
Raised while scoping org-wide audio support, see roadmap `.vibe/decisions/026`. No dependency on the new `snd` repo — this is a plain file path, not a decoded binary format. Feeds `mode-quick-versus#013` (match audio playback).
