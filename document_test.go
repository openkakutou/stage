package stage

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseDocument_RealisticFixture_SerializeReproducesSourceByteForByte(t *testing.T) {
	src := `; Training Room stage definition
[Info] ; identifying information
name = "Training Room"
author = "Elecbyte"

[Music]
bgmusic = training.mp3

[BGDef]
spr = stage0.sff

[StageInfo]
localcoord = 320,240
zoffset = 220

[Camera]
boundleft = -180
boundright = 180
boundhigh = -240
boundlow = 0
zoomout = 0.75
zoomin = 1.5

[PlayerInfo]
leftbound = -1000
rightbound = 1000

[BG sky]
type = normal
spriteno = 0,0
layerno = 0
start = 0,0
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}

	if buf.String() != src {
		t.Errorf("expected byte-for-byte reproduction of source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}

func TestParseDocument_RealisticFixture_ExposesDecodedStage(t *testing.T) {
	src := `[BGDef]
spr = stage0.sff

[StageInfo]
localcoord = 320,240
zoffset = 220
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Stage.BGdef.SpriteFile != "stage0.sff" {
		t.Errorf("expected decoded BGdef.SpriteFile %q, got %q", "stage0.sff", doc.Stage.BGdef.SpriteFile)
	}
	if doc.Stage.BGdef.LocalCoordWidth != 320 || doc.Stage.BGdef.LocalCoordHeight != 240 {
		t.Errorf("expected decoded LocalCoordWidth/Height 320/240, got %d/%d", doc.Stage.BGdef.LocalCoordWidth, doc.Stage.BGdef.LocalCoordHeight)
	}
}

func TestParseDocument_EmptyInput_SerializeReproducesEmptySource(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestParseDocument_UnrecognizedSections_ArePreservedVerbatim(t *testing.T) {
	src := `[Info]
name = "Training Room"

[Shadow]
intensity = 128
color = 0,0,0

[Reflection]
type = 0

[BGDef]
spr = stage0.sff
`

	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected unrecognized sections to be preserved verbatim.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}

func TestParseDocument_MalformedSource_ReturnsError(t *testing.T) {
	src := `[BGDef
spr = stage0.sff
`
	_, err := ParseDocument(strings.NewReader(src))
	if err == nil {
		t.Fatal("expected an error for malformed source, got nil")
	}
}

func TestParseDocument_ReaderFailure_ReturnsError(t *testing.T) {
	_, err := ParseDocument(errorReader{})
	if err == nil {
		t.Fatal("expected an error from a failing reader, got nil")
	}
}

func TestDocument_MutatingStageAfterParse_DoesNotAffectSerializeOutput(t *testing.T) {
	src := `[BGDef]
spr = stage0.sff
`
	doc, err := ParseDocument(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc.Stage.BGdef.SpriteFile = "someone-else.sff"

	var buf bytes.Buffer
	if err := doc.Serialize(&buf); err != nil {
		t.Fatalf("unexpected error serializing: %v", err)
	}
	if buf.String() != src {
		t.Errorf("expected Serialize to ignore mutated Stage and reproduce original source.\n--- want ---\n%s\n--- got ---\n%s", src, buf.String())
	}
}

func TestDocument_WriterError_ReturnsError(t *testing.T) {
	doc, err := ParseDocument(strings.NewReader("[BGDef]\nspr = stage0.sff\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := doc.Serialize(errorWriter{}); err == nil {
		t.Fatal("expected an error from a failing writer, got nil")
	}
}
