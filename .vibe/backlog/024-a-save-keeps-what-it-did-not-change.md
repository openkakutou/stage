---
status: todo
---
# A save keeps what it did not change

## Description
Found by the audit of 2026-11-21 (medium):
- `actionno` and its `[Begin Action]` block are written only for `type = anim` (`serializer.go:393`, 444). Ikemen GO plays the action of any non-normal element, or one without `spriteno` (stage.go:312-316): saving a parallax element with `actionno` replaces its animation with `spriteno = 0,0`.
- `SerializeDef` compares the model that came through JSON with the parsed one (`save.go:53`). `json.Marshal` turns invalid UTF-8 into U+FFFD, so a Shift-JIS or Latin-1 stage never compares equal: saving it unchanged takes the fresh path, losing comments and unmodelled sections and mangling its names.

## Acceptance Criteria
- [ ] An element of any type keeps its `actionno` and action block on save
- [ ] Saving a Shift-JIS stage unchanged returns its bytes unchanged
