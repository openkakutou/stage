package stage

import (
	"strings"
	"testing"
)

func TestSerializeDef_NoEdits_ProducesByteIdenticalOutput(t *testing.T) {
	original := []byte(`[BGDef]
spr = stage0.sff

[StageInfo]
localcoord = 320,240
zoffset = 220

[Camera]
boundleft = -180
boundright = 180
boundhigh = -240
boundlow = 0

[PlayerInfo]
leftbound = -1000
rightbound = 1000

[BG sky]
type = normal
spriteno = 0,0
`)

	s, err := Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: Parse failed: %v", err)
	}

	out, err := SerializeDef(original, s)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}

	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output on no edits\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}

func TestSerializeDef_WithEdits_ReflectsEditedValues(t *testing.T) {
	original := []byte(`[BGDef]
spr = stage0.sff

[StageInfo]
localcoord = 320,240
zoffset = 220

[Camera]
boundleft = -180
boundright = 180
boundhigh = -240
boundlow = 0

[PlayerInfo]
leftbound = -1000
rightbound = 1000

[BG sky]
type = normal
spriteno = 0,0
`)

	s, err := Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: Parse failed: %v", err)
	}
	s.BGdef.ZOffset = 999

	out, err := SerializeDef(original, s)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}

	if string(out) == string(original) {
		t.Fatalf("expected output to differ from original once edited, got identical bytes")
	}

	roundTripped, err := Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.BGdef.ZOffset != 999 {
		t.Fatalf("expected edited ZOffset to survive round trip, got %d", roundTripped.BGdef.ZOffset)
	}
	if roundTripped.BGdef.SpriteFile != "stage0.sff" {
		t.Fatalf("expected untouched SpriteFile to survive round trip, got %q", roundTripped.BGdef.SpriteFile)
	}
}

func TestSerializeDef_EmptyOriginal_SerializesFreshForNewStage(t *testing.T) {
	s := Stage{
		BGdef: BGdef{SpriteFile: "new.sff", LocalCoordWidth: 320, LocalCoordHeight: 240},
		Elements: []BGElement{
			{Name: "sky", Type: BGElementNormal, Sprite: SpriteRef{Group: 0, Image: 0}},
		},
	}

	out, err := SerializeDef(nil, s)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}
	if len(out) == 0 {
		t.Fatalf("expected non-empty output for a brand new stage")
	}

	roundTripped, err := Parse(strings.NewReader(string(out)))
	if err != nil {
		t.Fatalf("re-parsing serialized output failed: %v", err)
	}
	if roundTripped.BGdef.SpriteFile != s.BGdef.SpriteFile {
		t.Fatalf("expected SpriteFile %q, got %q", s.BGdef.SpriteFile, roundTripped.BGdef.SpriteFile)
	}
}

func TestSerializeDef_MalformedOriginal_ReturnsDescriptiveError(t *testing.T) {
	malformed := []byte("[BGDef\nspr = stage0.sff\n")

	_, err := SerializeDef(malformed, Stage{})
	if err == nil {
		t.Fatalf("expected an error for a malformed original .def, got nil")
	}
	if !strings.Contains(err.Error(), "stage:") {
		t.Fatalf("expected error to mention the stage package's own diagnostic, got: %v", err)
	}
}

// TestSerializeDef_NoEdits_NilVsEmptySlicesDoNotCountAsEdited verifies that a
// caller round-tripping the baseline through JSON — where an absent/empty
// Elements slice becomes a non-nil empty slice instead of staying nil — is
// still treated as "no edits", the same normalization character's own
// SerializeDef already applies for its JSON contract.
func TestSerializeDef_NoEdits_NilVsEmptySlicesDoNotCountAsEdited(t *testing.T) {
	original := []byte(`[BGDef]
spr = stage0.sff

[StageInfo]
localcoord = 320,240
`)

	s, err := Parse(strings.NewReader(string(original)))
	if err != nil {
		t.Fatalf("test setup: Parse failed: %v", err)
	}
	// s.Elements is nil here (no [BG ...] sections in the source); simulate
	// what a JSON round trip through a JS caller would produce.
	s.Elements = []BGElement{}

	out, err := SerializeDef(original, s)
	if err != nil {
		t.Fatalf("SerializeDef returned an error: %v", err)
	}
	if string(out) != string(original) {
		t.Fatalf("expected byte-identical output when only nil-vs-empty-slice differs\n--- got ---\n%s\n--- want ---\n%s", out, original)
	}
}
