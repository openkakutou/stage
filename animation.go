package stage

// BGAnimFrame is a single displayed frame within a BGAnimation: which
// sprite to show and how long (in ticks) to hold it before advancing,
// mirroring the ".air"-syntax "[Begin Action N]" block a stage .def
// declares an animated BG element's frames in (.def "actionno").
type BGAnimFrame struct {
	Sprite SpriteRef `json:"sprite"`
	Time   int       `json:"time"`
}

// BGAnimation is one "[Begin Action N]" block: an ordered sequence of
// BGAnimFrames plus the loop point frames return to once they have all
// played through once, mirroring `character`'s `air.Animation` — the two
// share the identical underlying file syntax, per MUGEN's own stage
// documentation.
//
// LoopStart is the index into Frames the animation loops back to once the
// full sequence has played once. Its zero value (0) matches the format's
// own default: an animation with no Loopstart marker loops back to its very
// first frame.
//
// This type is not yet populated by Parse: reading "[Begin Action N]"
// blocks out of stage .def text is tracked as a follow-up backlog item, see
// .vibe/decisions/004.
type BGAnimation struct {
	Frames    []BGAnimFrame `json:"frames"`
	LoopStart int           `json:"loopStart"`
}

// blankSpriteRef is the "no sprite" sentinel ResolveAnimationFrame falls
// back to for a malformed or empty frame sequence, mirroring the
// negative-group-or-image convention air.Frame.IsBlank() already
// establishes for the identical concept in character animations.
var blankSpriteRef = SpriteRef{Group: -1, Image: -1}

// IsBlank reports whether ref is the "no sprite shown" sentinel: a Group or
// Image below zero.
func (ref SpriteRef) IsBlank() bool {
	return ref.Group < 0 || ref.Image < 0
}

// ResolveParallaxPosition computes a BG element's effective on-screen
// position given the camera's current position, per MUGEN's own delta
// definition: "how many pixels the background element should scroll for
// each pixel of camera movement" — i.e. the element's position is offset
// from its configured start by cameraPosition scaled by its delta. A delta
// of 0 keeps the element fixed on screen regardless of camera movement; a
// delta of 1 scrolls it exactly with the camera; fractional deltas below 1
// scroll it more slowly than the camera for the classic depth illusion.
// This formula applies uniformly regardless of element.Type — a
// non-parallax element simply carries whatever Delta its .def declares
// (defaulting to 1,1 if unset by Parse).
func ResolveParallaxPosition(element BGElement, cameraX, cameraY float64) (x, y float64) {
	x = float64(element.StartX) + cameraX*element.DeltaX
	y = float64(element.StartY) + cameraY*element.DeltaY
	return x, y
}

// ResolveAnimationFrame returns the sprite an animated BG element should
// currently show, elapsedTicks after its animation started, honoring
// anim.LoopStart the same way air.Animation's own LoopStart does: once the
// full Frames sequence has played through once, playback loops back to
// index LoopStart (not necessarily the start) and keeps cycling
// Frames[LoopStart:] indefinitely.
//
// A frame with a zero or negative Time is instantaneous — it is skipped
// when walking elapsed time, but does not otherwise break resolution.
//
// ResolveAnimationFrame never panics or returns an out-of-bounds frame: an
// empty Frames sequence, an out-of-range LoopStart, or a sequence whose
// frames carry no positive Time at all (so it can never actually advance or
// loop) all resolve to the blank "no sprite" sentinel instead.
func ResolveAnimationFrame(anim BGAnimation, elapsedTicks int) SpriteRef {
	if len(anim.Frames) == 0 {
		return blankSpriteRef
	}
	if anim.LoopStart < 0 || anim.LoopStart >= len(anim.Frames) {
		return blankSpriteRef
	}

	firstPassDuration := totalDuration(anim.Frames)
	if firstPassDuration <= 0 {
		return blankSpriteRef
	}

	if elapsedTicks < 0 {
		elapsedTicks = 0
	}

	if elapsedTicks < firstPassDuration {
		return frameAtOffset(anim.Frames, elapsedTicks)
	}

	loopFrames := anim.Frames[anim.LoopStart:]
	loopDuration := totalDuration(loopFrames)
	if loopDuration <= 0 {
		return blankSpriteRef
	}

	loopElapsed := (elapsedTicks - firstPassDuration) % loopDuration
	return frameAtOffset(loopFrames, loopElapsed)
}

// totalDuration sums the positive Time of every frame in frames, treating a
// zero or negative Time as contributing nothing (an instantaneous frame).
func totalDuration(frames []BGAnimFrame) int {
	total := 0
	for _, f := range frames {
		if f.Time > 0 {
			total += f.Time
		}
	}
	return total
}

// frameAtOffset walks frames, accumulating each positive-Time frame's
// duration, and returns the sprite of the frame whose span contains offset.
// Callers must ensure frames has at least one frame with a positive Time
// and 0 <= offset < totalDuration(frames).
func frameAtOffset(frames []BGAnimFrame, offset int) SpriteRef {
	elapsed := 0
	for _, f := range frames {
		if f.Time <= 0 {
			continue
		}
		elapsed += f.Time
		if offset < elapsed {
			return f.Sprite
		}
	}
	// Unreachable given the caller contract above, but returning the last
	// frame's sprite rather than panicking keeps this defensive.
	return frames[len(frames)-1].Sprite
}
