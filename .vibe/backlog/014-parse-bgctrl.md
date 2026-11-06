---
status: in_progress
---
# Parse BGCtrlDef and BGCtrl controllers

## Description
Real MUGEN/Ikemen stages animate backgrounds with [BGCtrlDef]/[BGCtrl] controllers (visibility, position, velocity, sin/anim changes, timers). They are currently dropped as unrecognised content (a real-world fixture in stage-viewer-web already contains them). Add a typed model and parser.

## Acceptance Criteria
- [ ] BGCtrlDef and BGCtrl blocks parse into typed data (type, ctrlID, time, looptime, parameters)
- [ ] Unknown controller types are kept raw, not dropped
- [ ] Validated against the real stage corpus with a count of controllers per type
