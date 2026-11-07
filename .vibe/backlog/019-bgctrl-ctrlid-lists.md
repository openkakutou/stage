---
status: todo
---
# Support ctrlID lists and sinx/siny in BG controllers

## Description
MUGEN's `ctrlID` on a `[BGCtrlDef]` and on a `[BGCtrl]` is a list of up to 10 ids; the model and parser hold a single integer (a list fails to parse). Model the list, keep byte-exact round trips, and make `ResolveBGCtrlState` target elements whose `id` is in the list. Also resolve `sinx`/`siny` controllers (MUGEN 1.1 ignores the group looptime for them; Ikemen GO documents the same quirk).

## Acceptance Criteria
- [ ] ctrlID lists parse, serialize and round-trip byte-exact
- [ ] ResolveBGCtrlState targets by list membership
- [ ] sinx/siny resolved, divergence recorded in an ADR
- [ ] Real-stage fixture test

## Notes
Found while grounding stage#016 against the MUGEN/Ikemen GO docs. Must stay compatible with both MUGEN and Ikemen GO.
