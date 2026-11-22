---
date: 2026-11-22
status: accepted
---
# BG animations are read and played as Ikemen GO does

**Context:** Backlog item 023 (audit of 2026-11-21). Decision 004 treated a frame of time -1 as instantaneous and kept a `Loopstart` past the last frame, so a static `…,-1` background and an action ending on `Loopstart` were invisible, and a Clsn or Interpolate line inside `[Begin Action]` failed the whole stage.

**Decision:** Follow Ikemen GO's `anim.go` (commit `54d4148a`): a last frame of time -1 holds forever after the timed frames play once (441-443, 931-946); a `Loopstart` past the last frame falls back to the loop start in effect when that frame was read (383, 440-442); `loopstart` matches as a prefix (395); a line not starting with a digit or `-` is skipped (47-49).

**Reason:** The org-wide MUST: stage behaviour matches Ikemen GO.

**Rejected alternatives:** Keeping decision 004's rules and documenting the divergence — it hides layers Ikemen GO shows.

Supersedes decision 004's rules on negative frame times and out-of-range loop starts; the rest of 004 stands.
