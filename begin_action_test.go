package stage

import (
	"bytes"
	"strings"
	"testing"
)

func TestParse_BeginActionBlock_PopulatesAnimationsMap(t *testing.T) {
	src := `[Begin Action 200]
0,0, 0,0, 5
0,1, 5,-2, 4
0,2, 0,0, 3
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	anim, ok := s.Animations[200]
	if !ok {
		t.Fatalf("expected Animations[200] to be present, got %v", s.Animations)
	}
	want := []BGAnimFrame{
		{Sprite: SpriteRef{Group: 0, Image: 0}, Time: 5},
		{Sprite: SpriteRef{Group: 0, Image: 1}, Time: 4},
		{Sprite: SpriteRef{Group: 0, Image: 2}, Time: 3},
	}
	if len(anim.Frames) != len(want) {
		t.Fatalf("expected %d frames, got %d: %+v", len(want), len(anim.Frames), anim.Frames)
	}
	for i, f := range anim.Frames {
		if f != want[i] {
			t.Errorf("frame %d = %+v, want %+v", i, f, want[i])
		}
	}
}

func TestParse_LoopstartMarker_SetsLoopStartToFrameIndex(t *testing.T) {
	src := `[Begin Action 5]
0,0, 0,0, 1
Loopstart
0,1, 0,0, 2
0,2, 0,0, 3
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	anim := s.Animations[5]
	if anim.LoopStart != 1 {
		t.Errorf("expected LoopStart 1, got %d", anim.LoopStart)
	}
	if len(anim.Frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(anim.Frames))
	}
}

func TestParse_LoopstartCaseInsensitive_IsRecognized(t *testing.T) {
	src := `[Begin Action 5]
0,0, 0,0, 1
LOOPSTART
0,1, 0,0, 2
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Animations[5].LoopStart != 1 {
		t.Errorf("expected LoopStart 1, got %d", s.Animations[5].LoopStart)
	}
}

func TestParse_MultipleBeginActionBlocks_InterleavedWithBGElements_AreAllParsed(t *testing.T) {
	// Real MUGEN/Ikemen files can group [Begin Action N] blocks anywhere in
	// the file, not necessarily right after the [BG name] element that
	// references them.
	src := `[BG torch]
type = anim
actionno = 10
layerno = 1

[Begin Action 20]
1,0, 0,0, 6

[BG candle]
type = anim
actionno = 20
layerno = 1

[Begin Action 10]
2,0, 0,0, 8
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Elements) != 2 {
		t.Fatalf("expected 2 BG elements, got %d", len(s.Elements))
	}
	torch, candle := s.Elements[0], s.Elements[1]
	if torch.Name != "torch" || torch.ActionNumber != 10 {
		t.Errorf("torch = %+v, want name torch actionNumber 10", torch)
	}
	if candle.Name != "candle" || candle.ActionNumber != 20 {
		t.Errorf("candle = %+v, want name candle actionNumber 20", candle)
	}

	anim10, ok := s.Animations[torch.ActionNumber]
	if !ok || len(anim10.Frames) != 1 || anim10.Frames[0].Sprite != (SpriteRef{Group: 2, Image: 0}) {
		t.Errorf("Animations[10] = %+v, ok=%v, want a single {2,0} frame", anim10, ok)
	}
	anim20, ok := s.Animations[candle.ActionNumber]
	if !ok || len(anim20.Frames) != 1 || anim20.Frames[0].Sprite != (SpriteRef{Group: 1, Image: 0}) {
		t.Errorf("Animations[20] = %+v, ok=%v, want a single {1,0} frame", anim20, ok)
	}

	// End-to-end: resolving torch's currently-visible sprite via the
	// already-shipped ResolveAnimationFrame against the newly-parsed data.
	sprite := ResolveAnimationFrame(anim10, 0)
	if sprite != (SpriteRef{Group: 2, Image: 0}) {
		t.Errorf("ResolveAnimationFrame at tick 0 = %+v, want {2,0}", sprite)
	}
}

func TestParse_NoBeginActionBlocks_LeavesAnimationsNil(t *testing.T) {
	src := "[Info]\nname = \"Training Room\"\n"
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Animations != nil {
		t.Errorf("expected Animations to stay nil, got %v", s.Animations)
	}
}

func TestParse_UnrelatedBracketLineInsideAnimationParsing_IsSkippedNotErrored(t *testing.T) {
	src := `[Begin Action 1]
0,0, 0,0, 1

[SomeUnrelatedSection]
whatever = 1

[Begin Action 2]
0,1, 0,0, 2
`
	s, err := Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Animations) != 2 {
		t.Fatalf("expected 2 animations, got %d: %v", len(s.Animations), s.Animations)
	}
}

func TestParse_MalformedActionHeader_ReturnsLineNumberedError(t *testing.T) {
	src := "[Begin Action]\n0,0, 0,0, 1\n"
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to mention line 1, got: %v", err)
	}
}

func TestParse_MalformedFrameLine_TooFewFields_ReturnsLineNumberedError(t *testing.T) {
	src := "[Begin Action 1]\n0,0,0\n"
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got: %v", err)
	}
}

func TestParse_MalformedFrameLine_NonNumericGroup_ReturnsLineNumberedError(t *testing.T) {
	src := "[Begin Action 1]\nabc,0, 0,0, 5\n"
	_, err := Parse(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Errorf("expected error to mention line 2, got: %v", err)
	}
}

func TestSerialize_WritesBeginActionBlocksForReferencedElements_ReparsingToEquivalentResult(t *testing.T) {
	s := Stage{
		Elements: []BGElement{
			{Name: "torch", Type: BGElementAnim, ActionNumber: 10, LayerNo: 1},
		},
		Animations: map[int]BGAnimation{
			10: {
				Frames: []BGAnimFrame{
					{Sprite: SpriteRef{Group: 2, Image: 0}, Time: 8},
					{Sprite: SpriteRef{Group: 2, Image: 1}, Time: 4},
				},
				LoopStart: 1,
			},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reparsed, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v\noutput:\n%s", err, buf.String())
	}

	got, ok := reparsed.Animations[10]
	if !ok {
		t.Fatalf("expected Animations[10] in re-parsed result, got %v", reparsed.Animations)
	}
	if len(got.Frames) != 2 || got.Frames[0].Sprite != (SpriteRef{Group: 2, Image: 0}) || got.Frames[0].Time != 8 {
		t.Errorf("re-parsed frame 0 = %+v, want {2,0} time 8", got.Frames)
	}
	if got.LoopStart != 1 {
		t.Errorf("re-parsed LoopStart = %d, want 1", got.LoopStart)
	}
}

func TestSerialize_SkipsAnActionNumber_NotResolvableInAnimationsMap(t *testing.T) {
	s := Stage{
		Elements: []BGElement{
			{Name: "torch", Type: BGElementAnim, ActionNumber: 999, LayerNo: 1},
		},
		// No entry for 999 in Animations at all.
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(buf.String(), "Begin Action") {
		t.Errorf("expected no [Begin Action] block for an unresolvable reference, got:\n%s", buf.String())
	}
}

func TestSerialize_WritesEachReferencedActionNumberOnlyOnce(t *testing.T) {
	s := Stage{
		Elements: []BGElement{
			{Name: "torch", Type: BGElementAnim, ActionNumber: 10, LayerNo: 1},
			{Name: "torch2", Type: BGElementAnim, ActionNumber: 10, LayerNo: 1},
		},
		Animations: map[int]BGAnimation{
			10: {Frames: []BGAnimFrame{{Sprite: SpriteRef{Group: 0, Image: 0}, Time: 1}}},
		},
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := strings.Count(buf.String(), "[Begin Action 10]")
	if count != 1 {
		t.Errorf("expected exactly 1 [Begin Action 10] block, got %d:\n%s", count, buf.String())
	}
}
