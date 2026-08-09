---
date: 2026-08-09
status: accepted
---
# StageBoundaries models only Left/Right, not Top/Bottom

**Context:** Backlog item 001 ("Define Stage Data Model") describes stage
boundaries loosely as "left/right/top/bottom edges characters and the
camera are constrained to", but its own acceptance criteria define the
concept more precisely as "camera scroll limits vs. character movement
limits" — two distinct things. Verified against the real MUGEN/Ikemen GO
`.def` format (elecbyte MUGEN 1.1 docs, Ikemen GO stage wiki): the
`[Camera]` section's `boundleft`/`boundright`/`boundhigh`/`boundlow` clamp
only the camera's own scroll position (all 4 edges), while the
`[PlayerInfo]` section's `leftbound`/`rightbound` clamp only where
characters may move on the x-axis — there is no mainline vertical
(top/bottom) bound on character movement.

**Decision:** `CameraBounds` models all four edges (Left, Right, High, Low,
from `[Camera]`). `StageBoundaries` models only Left/Right (from
`[PlayerInfo]` `leftbound`/`rightbound`) — no Top/Bottom fields, since no
such character-movement key exists in the format this repo targets
(mainline MUGEN/Ikemen GO stage `.def`).

**Reason:** Accuracy matters more here than literal adherence to the
backlog's looser prose: item 001's own acceptance criteria require every
field to cite an unambiguous `.def` key origin so the future parser (item
002) has a clean mapping. Inventing a Top/Bottom character-bound pair with
no corresponding `.def` key would violate that criterion and mislead the
parser's design. If Ikemen GO's nightly Z-axis `topbound`/`botbound`
extension (a different, depth-axis concept per its own docs) ever needs
modeling, that is a separate, later decision — not a same-axis twin of
Left/Right.

**Rejected alternatives:**
- *Add Top/Bottom fields to StageBoundaries anyway, left unmapped to any
  `.def` key* — rejected: violates the "every field cites its `.def` key
  origin" acceptance criterion, and risks the future parser inventing
  behavior for a key that does not exist.
- *Merge CameraBounds and StageBoundaries into one four-edge type* —
  rejected: contradicts the acceptance criterion that the two remain
  distinct concepts (scroll limits vs. movement limits), and conflates two
  independent `.def` sections (`[Camera]` vs `[PlayerInfo]`).

**Update (2026-08-09):** the deferred question above has been answered by
the roadmap's `.vibe/decisions/014-support-ikemen-go-3d-stages.md` —
Ikemen GO's `[PlayerInfo]` `topbound`/`botbound` (Z-axis character movement
bound, verified real) is now scoped for this repo's backlog item `008`.
This decision's own Left/Right-only ruling still stands as written above;
014 only adds the previously-deferred Z axis alongside it, as a distinct
concept from this decision's rejected Top/Bottom (which would have been a
same-axis Y/vertical twin of Left/Right — Z is a different axis entirely).
