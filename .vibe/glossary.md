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
The range characters are allowed to move within during a match — distinct from Camera bounds, which clamp the camera instead. A left/right (horizontal) range is always defined; a model-based stage additionally defines a top/bottom (depth) range. No vertical (up/down) counterpart exists in the format.
**Do not confuse with:** Camera bounds, Perspective scaling.
_Sources: `bounds.go`, `.vibe/decisions/001-stage-boundaries-model-left-right-only.md`_

## Local coordinate space
The pixel coordinate system a stage's positions (BG element placement, ground level) are expressed in, independent of the actual resolution the stage is rendered at.
_Sources: `bgdef.go`_

## Model-based stage
A stage rendered from a 3D model instead of 2D sprite layers — Ikemen GO's 3D stage extension. Identified by whether the stage references a model file at all; a stage with none is a traditional 2D (sprite-based) stage and every 3D-only setting stays unused.
**Do not confuse with:** BG element, which is a 2D stage's own layer concept.
_Sources: `bgdef.go`, `model.go`_

## Perspective scaling
How a character's on-screen size and vertical screen offset change with their depth (Z) position on a model-based stage — the illusion that a character farther from the camera appears smaller and higher on screen.
**Do not confuse with:** Stage boundaries, which clamp where a character may move, not how it is drawn once there.
_Sources: `scaling.go`_
