---
status: done
depends_on: [014]
---
# Serialize BGCtrl and keep byte-exact Document round trips

## Description
Write BGCtrl data back out in the fresh-write path and make the format-preserving Document handle them so editors can save.

## Acceptance Criteria
- [ ] Fresh write emits valid BGCtrl blocks
- [ ] Unmodified real stage round-trips byte-exact
- [ ] Edited controller changes only its own lines
