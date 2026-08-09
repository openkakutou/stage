# Ubiquitous Language

## BG element
A single background layer of a stage — a static sprite, a depth-scrolling (parallax) layer, or an `.air`-animated layer — drawn at a given position and draw order relative to the characters. Modeled by `BGElement`, distinguished by its `Type`.
**Do not confuse with:** Parallax, which is one specific *kind* of BG element, not the general concept.
_Sources: `bg_element.go`_

## Parallax
A BG element whose scroll speed is a fraction of the camera's own movement, so it appears to sit farther away than layers that scroll at full speed — the classic depth illusion in a 2D background. Expressed as the element's `DeltaX`/`DeltaY` scroll ratio.
_Sources: `bg_element.go`_

## Camera bounds
The box the camera's own scroll position is clamped to as it follows the characters — distinct from Stage boundaries, which clamp the characters instead.
**Do not confuse with:** Stage boundaries.
_Sources: `bounds.go`_

## Stage boundaries
The horizontal range characters are allowed to move within during a match — distinct from Camera bounds, which clamp the camera instead. Only a left/right range is defined; the format has no vertical counterpart.
**Do not confuse with:** Camera bounds.
_Sources: `bounds.go`, `.vibe/decisions/001-stage-boundaries-model-left-right-only.md`_

## Local coordinate space
The pixel coordinate system a stage's positions (BG element placement, ground level) are expressed in, independent of the actual resolution the stage is rendered at.
_Sources: `bgdef.go`_
