---
date: 2026-11-07
status: accepted
---
# SerializeDef patches the original bytes in place when only BGCtrl data was edited

**Context:** Backlog item 015 requires that editing a background controller changes only that controller's own lines, while decision `005` had `SerializeDef` fall back to a fresh `Serialize` on any edit and `Document` decline per-line reconciliation.
**Decision:** `SerializeDef` patches the original bytes line by line when the only difference from the parsed baseline is in `BGCtrlDefs` and the group and controller counts are unchanged. The patch is re-parsed and must deep-equal the edited `Stage`; otherwise, and for every other kind of edit, it falls back to the fresh `Serialize` exactly as before.
**Reason:** Editors need small Git diffs for controller edits, the one case the item names; the re-parse check makes the patch unable to emit semantically wrong text.
**Rejected alternatives:** General per-line reconciliation for every field (contradicts `005`, much larger); handling controller or group addition and removal in place (costlier, not required); no targeted edit (fails the third acceptance criterion).
