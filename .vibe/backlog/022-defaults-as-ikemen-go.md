---
status: todo
---
# Stage defaults as Ikemen GO sets them

## Description
Found by the audit of 2026-11-21 (high; `delta` checked against stage.go:176). A missing key reads 0 where Ikemen GO has a non-zero default, and a fresh save writes that 0 into the file:
- A `[BG]` element without `delta` gets 0,0 (`parser.go:126`); Ikemen GO's `newBackGround` sets 1,1 (stage.go:176). The element stays fixed on screen instead of scrolling, and `ResolveParallaxPosition`'s doc already claims 1,1.
- Only `tension` gets its default (`parser.go:164`). `zoomout`/`zoomin` (1, camera.go:88), `localcoord` (320,240, camera.go:79) and `leftbound`/`rightbound` (-1000/1000, stage.go:1147-1148) read 0. Saving a MUGEN 1.0 stage that omits them writes `localcoord = 0,0` and `zoomout = 0` (`serializer.go:250, 294-298`), which breaks it in Ikemen GO.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Each default above is applied when its section starts, citing its Ikemen GO line
- [ ] A stage without these keys parses to Ikemen GO's values, and a save does not write a 0 the source never had
