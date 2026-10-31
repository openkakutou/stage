---
status: todo
---
# Fuzz the stage .def parser

## Description
Add Go fuzz targets for Parse and ParseDocument, bounded memory on pathological input.

## Acceptance Criteria
- [ ] Fuzz targets for Parse and ParseDocument in CI
- [ ] Findings fixed with regression seeds
- [ ] Huge tile/repeat values rejected or capped
