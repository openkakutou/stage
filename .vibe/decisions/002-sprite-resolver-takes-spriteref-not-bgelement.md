---
date: 2026-08-20
status: accepted
---
# `SpriteResolver.Resolve` takes a `SpriteRef`, not a whole `BGElement`

**Context:** Backlog item 004 integrates the external `sff` module to resolve a BG element's sprite reference, mirroring the shape `character`'s `air.SpriteResolver` established (`Resolve(air.Frame) (sff.Sprite, error)`). Unlike `air.Frame`, which always carries a `(Group, Image)` pair (using a negative-value sentinel for "no sprite"), `BGElement.Sprite` is only meaningful for `BGElementNormal`/`BGElementParallax` — a `BGElementAnim` element has no `SpriteRef` of its own at all, it plays an `.air` animation action via `ActionNumber` instead (that resolution is backlog item 005's job, not this one).

**Decision:** `SpriteResolver.Resolve` takes a `SpriteRef` value directly, not a `BGElement`. Callers pass `element.Sprite` for a `BGElementNormal`/`BGElementParallax` element; item 005's animation-frame resolution will produce its own `SpriteRef` (from the resolved `.air` frame) and pass that through the same `Resolve` call.

**Reason:** Accepting a whole `BGElement` would force this resolver to either silently ignore `BGElementAnim` elements' irrelevant `Sprite` zero value or branch on `Type` — a parsing/domain concern this pure resolver has no reason to know about. Keeping the input to exactly the `(Group, Image)` pair keeps `SpriteResolver` reusable by both the direct-sprite case (this item) and the future animation-frame case (item 005) without either one adapting to the other's shape.

**Rejected alternatives:**
- *`Resolve(BGElement) (sff.Sprite, error)`, mirroring `air.SpriteResolver.Resolve(Frame)` exactly* — rejected: `air.Frame` always carries a sprite reference (with a defined blank sentinel); `BGElement` does not, so the same shape would leak a `Type`-branching concern into the resolver.
- *A blank-reference sentinel on `SpriteRef`, mirroring `air.Frame.IsBlank()`* — rejected: no MUGEN/Ikemen GO `.def` convention establishes a "no sprite" sentinel for a BG element's `spriteno` the way `.air` frames use negative values; inventing one here would be unfounded.
