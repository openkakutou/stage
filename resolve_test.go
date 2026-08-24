package stage

import "testing"

func TestResolveParallaxPosition_MovesByDeltaTimesCameraOffset(t *testing.T) {
	// delta = 1,1: the element scrolls one pixel for every pixel of camera
	// movement, per MUGEN's own documented delta definition.
	element := BGElement{StartX: 100, StartY: 50, DeltaX: 1, DeltaY: 1}

	x, y := ResolveParallaxPosition(element, 30, 10)

	if x != 130 || y != 60 {
		t.Errorf("expected (130, 60), got (%v, %v)", x, y)
	}
}

func TestResolveParallaxPosition_ScalesByFractionalDeltaForDepthIllusion(t *testing.T) {
	// A far-away layer scrolls slower than the camera (delta below 1).
	element := BGElement{StartX: 0, StartY: 0, DeltaX: 0.5, DeltaY: 0.25}

	x, y := ResolveParallaxPosition(element, 40, 40)

	if x != 20 || y != 10 {
		t.Errorf("expected (20, 10), got (%v, %v)", x, y)
	}
}

func TestResolveParallaxPosition_ZeroDeltaStaysFixedOnScreenRegardlessOfCamera(t *testing.T) {
	// delta = 0,0: the element never scrolls, e.g. a distant sky layer.
	element := BGElement{StartX: 200, StartY: 80, DeltaX: 0, DeltaY: 0}

	x1, y1 := ResolveParallaxPosition(element, 0, 0)
	x2, y2 := ResolveParallaxPosition(element, 500, -500)

	if x1 != x2 || y1 != y2 {
		t.Errorf("expected the same position regardless of camera movement, got (%v,%v) and (%v,%v)", x1, y1, x2, y2)
	}
	if x1 != 200 || y1 != 80 {
		t.Errorf("expected (200, 80), got (%v, %v)", x1, y1)
	}
}

func TestResolveParallaxPosition_NegativeDeltaScrollsOppositeTheCamera(t *testing.T) {
	// The .def format allows a negative delta; it is not this function's job
	// to reject it, only to apply the documented formula.
	element := BGElement{StartX: 0, StartY: 0, DeltaX: -1, DeltaY: 0}

	x, _ := ResolveParallaxPosition(element, 15, 0)

	if x != -15 {
		t.Errorf("expected -15, got %v", x)
	}
}

func TestResolveAnimationFrame_ReturnsFirstFrameAtZeroElapsedTime(t *testing.T) {
	anim := BGAnimation{Frames: []BGAnimFrame{
		{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 10},
		{Sprite: SpriteRef{Group: 1, Image: 1}, Time: 10},
	}}

	got := ResolveAnimationFrame(anim, 0)

	want := SpriteRef{Group: 1, Image: 0}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestResolveAnimationFrame_AdvancesToNextFrameOncePriorFramesElapse(t *testing.T) {
	anim := BGAnimation{Frames: []BGAnimFrame{
		{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 10},
		{Sprite: SpriteRef{Group: 1, Image: 1}, Time: 5},
		{Sprite: SpriteRef{Group: 1, Image: 2}, Time: 20},
	}}

	got := ResolveAnimationFrame(anim, 12)

	want := SpriteRef{Group: 1, Image: 1}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestResolveAnimationFrame_LoopsBackToLoopStartAfterLastFrame(t *testing.T) {
	// Total duration of one full pass is 10+5+20=35 ticks. LoopStart=1 means
	// the loop segment replayed forever afterward is frames[1:] (5+20=25
	// ticks long), mirroring air.Animation's own LoopStart semantics.
	anim := BGAnimation{
		Frames: []BGAnimFrame{
			{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 10},
			{Sprite: SpriteRef{Group: 1, Image: 1}, Time: 5},
			{Sprite: SpriteRef{Group: 1, Image: 2}, Time: 20},
		},
		LoopStart: 1,
	}

	// 35 (first pass) + 3 into the loop segment lands back on frame 1.
	got := ResolveAnimationFrame(anim, 35+3)

	want := SpriteRef{Group: 1, Image: 1}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestResolveAnimationFrame_LoopsMultipleTimesForLargeElapsedTime(t *testing.T) {
	// LoopStart=0: the whole 15-tick sequence repeats. A large elapsed time
	// (7 full loops + 4 ticks) must still land on the correct frame.
	anim := BGAnimation{Frames: []BGAnimFrame{
		{Sprite: SpriteRef{Group: 2, Image: 0}, Time: 10},
		{Sprite: SpriteRef{Group: 2, Image: 1}, Time: 5},
	}}

	got := ResolveAnimationFrame(anim, 15*7+4)

	want := SpriteRef{Group: 2, Image: 0}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestResolveAnimationFrame_ReturnsBlankSentinelForEmptyFrameSequence(t *testing.T) {
	got := ResolveAnimationFrame(BGAnimation{}, 0)

	if !got.IsBlank() {
		t.Errorf("expected a blank sentinel for an empty frame sequence, got %+v", got)
	}
}

func TestResolveAnimationFrame_ReturnsBlankSentinelForOutOfRangeLoopStart(t *testing.T) {
	anim := BGAnimation{
		Frames:    []BGAnimFrame{{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 10}},
		LoopStart: 5,
	}

	got := ResolveAnimationFrame(anim, 0)

	if !got.IsBlank() {
		t.Errorf("expected a blank sentinel for an out-of-range LoopStart, got %+v", got)
	}
}

func TestResolveAnimationFrame_ReturnsBlankSentinelWhenEveryFrameHasNoDuration(t *testing.T) {
	anim := BGAnimation{Frames: []BGAnimFrame{
		{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 0},
		{Sprite: SpriteRef{Group: 1, Image: 1}, Time: -1},
	}}

	got := ResolveAnimationFrame(anim, 0)

	if !got.IsBlank() {
		t.Errorf("expected a blank sentinel when no frame has positive duration, got %+v", got)
	}
}

func TestResolveAnimationFrame_SkipsZeroDurationFramesInTheMiddleOfTheSequence(t *testing.T) {
	// A zero-Time frame is instantaneous — it is never the frame landed on
	// by elapsed-time resolution, but it must not break the walk over the
	// frames after it.
	anim := BGAnimation{Frames: []BGAnimFrame{
		{Sprite: SpriteRef{Group: 1, Image: 0}, Time: 10},
		{Sprite: SpriteRef{Group: 1, Image: 1}, Time: 0},
		{Sprite: SpriteRef{Group: 1, Image: 2}, Time: 10},
	}}

	got := ResolveAnimationFrame(anim, 15)

	want := SpriteRef{Group: 1, Image: 2}
	if got != want {
		t.Errorf("expected %+v, got %+v", want, got)
	}
}

func TestSpriteRef_IsBlank_ReportsNegativeGroupOrImage(t *testing.T) {
	cases := []struct {
		name string
		ref  SpriteRef
		want bool
	}{
		{"positive group and image", SpriteRef{Group: 1, Image: 2}, false},
		{"zero group and image", SpriteRef{Group: 0, Image: 0}, false},
		{"negative group", SpriteRef{Group: -1, Image: 0}, true},
		{"negative image", SpriteRef{Group: 0, Image: -1}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.ref.IsBlank(); got != c.want {
				t.Errorf("IsBlank() = %v, want %v", got, c.want)
			}
		})
	}
}
