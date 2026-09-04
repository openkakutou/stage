package stage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// readTestdataFile reads a real, vendored fixture from testdata/ — see
// testdata/README.md for provenance. Trimmed real stage .def files
// (backlog item 007), not hand-built synthetic data: real-world authoring
// habits synthetic fixtures don't reproduce (comment placement, spacing,
// section ordering, Ikemen GO's own 3D extension syntax).
func readTestdataFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata fixture %s: %v", name, err)
	}
	return data
}

// TestParse_RealMugenOnlyFixture_DecodesCorrectly exercises Parse against a
// real, unmodified MUGEN 1.1 stage (no Ikemen GO 3D extension) — see
// testdata/README.md for provenance.
func TestParse_RealMugenOnlyFixture_DecodesCorrectly(t *testing.T) {
	data := readTestdataFile(t, "mugen-2d-stage.def")

	s, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Name != "The_Great_Cave_Offensive" {
		t.Errorf("expected Name %q, got %q", "The_Great_Cave_Offensive", s.Name)
	}
	if s.BGdef.ModelFile != "" {
		t.Errorf("expected a 2D-only stage to have no ModelFile, got %q", s.BGdef.ModelFile)
	}
	// The real file writes "zoffset = 555.0" — a decimal value for what's
	// an integer field, one of this item's own real-file tolerance fixes.
	if s.BGdef.ZOffset != 555 {
		t.Errorf("expected ZOffset 555 (rounded from the real file's \"555.0\"), got %d", s.BGdef.ZOffset)
	}
	if len(s.Elements) != 2 {
		t.Fatalf("expected 2 BG elements, got %d", len(s.Elements))
	}
	// Both real elements write "tilespacing = 1, 0" — the documented pair
	// form, distinct from the single-value shorthand another real fixture
	// needed this item's own tolerance fix for (see
	// TestParse_TileSpacingSingleValue_AppliesToBothAxes in parser_test.go).
	if s.Elements[0].TileSpacingX != 1 || s.Elements[0].TileSpacingY != 0 {
		t.Errorf("expected first element's TileSpacingX/Y 1/0, got %d/%d", s.Elements[0].TileSpacingX, s.Elements[0].TileSpacingY)
	}
}

// TestParse_RealIkemenGoOnlyFixture_DecodesCorrectly exercises Parse
// against a real, unmodified Ikemen GO 3D model-based stage — see
// testdata/README.md for provenance.
func TestParse_RealIkemenGoOnlyFixture_DecodesCorrectly(t *testing.T) {
	data := readTestdataFile(t, "ikemen-go-3d-model-stage.def")

	s, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Name != "Guilty Gear Xrd - Neo New York" {
		t.Errorf("expected Name %q, got %q", "Guilty Gear Xrd - Neo New York", s.Name)
	}
	if s.BGdef.ModelFile != "ggxrd-neonewyork.glb" {
		t.Errorf("expected ModelFile %q, got %q", "ggxrd-neonewyork.glb", s.BGdef.ModelFile)
	}
	if s.Model.OffsetX != 0 || s.Model.OffsetY != -0.195 || s.Model.OffsetZ != -0.85 {
		t.Errorf("expected Model offset 0/-0.195/-0.85, got %v/%v/%v", s.Model.OffsetX, s.Model.OffsetY, s.Model.OffsetZ)
	}
	if s.Model.ScaleX != 1 || s.Model.ScaleY != 1 || s.Model.ScaleZ != 1 {
		t.Errorf("expected Model scale 1/1/1, got %v/%v/%v", s.Model.ScaleX, s.Model.ScaleY, s.Model.ScaleZ)
	}
	if s.BGdef.FOV != 30 {
		t.Errorf("expected FOV 30, got %v", s.BGdef.FOV)
	}
	// The real file also has a "[Begin Action 9000]" portrait animation
	// block, the same underlying syntax item 009's own animation support
	// covers — real-file confirmation it composes correctly with a
	// model-based stage's other 3D-only sections.
	if _, ok := s.Animations[9000]; !ok {
		t.Errorf("expected Animations to include action 9000 (stage portrait), got %v", s.Animations)
	}
}

// TestParse_RealNonDefaultScaleFixture_DecodesXScaleYScale exercises Parse
// against a real, unmodified stage that authors hi-res BG sprite art and
// relies on "[StageInfo]"'s xscale/yscale to scale it down at draw time
// (backlog item 012) — see testdata/README.md for provenance.
func TestParse_RealNonDefaultScaleFixture_DecodesXScaleYScale(t *testing.T) {
	data := readTestdataFile(t, "mugen-nondefault-scale-stage.def")

	s, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Name != "Dengeki_Military Subway" {
		t.Errorf("expected Name %q, got %q", "Dengeki_Military Subway", s.Name)
	}
	// The real file writes "xscale = .35" / "yscale = .35" -- the exact
	// real-world shape (leading-dot decimal, no leading zero) this item's
	// bug report was filed against.
	if s.BGdef.XScale != 0.35 || s.BGdef.YScale != 0.35 {
		t.Errorf("expected XScale/YScale 0.35/0.35, got %v/%v", s.BGdef.XScale, s.BGdef.YScale)
	}
	if len(s.Elements) != 11 {
		t.Fatalf("expected 11 BG elements, got %d", len(s.Elements))
	}
	if _, ok := s.Animations[1]; !ok {
		t.Errorf("expected Animations to include action 1, got %v", s.Animations)
	}
}

// TestDocument_RealFixtures_RoundTripByteExact confirms both real, vendored
// fixtures satisfy the same byte-exact-on-unmodified-content guarantee
// Document promises for hand-built synthetic fixtures — a real file's own
// comment placement, blank lines, and section ordering are exactly the
// kind of shape synthetic test data doesn't reliably exercise.
func TestDocument_RealFixtures_RoundTripByteExact(t *testing.T) {
	for _, name := range []string{"mugen-2d-stage.def", "ikemen-go-3d-model-stage.def", "mugen-nondefault-scale-stage.def"} {
		t.Run(name, func(t *testing.T) {
			data := readTestdataFile(t, name)

			doc, err := ParseDocument(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("ParseDocument: %v", err)
			}
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				t.Fatalf("Serialize: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), data) {
				t.Errorf("expected byte-exact reproduction of %s (got %d bytes, want %d)", name, buf.Len(), len(data))
			}
		})
	}
}

// TestSerializeDef_RealFixtures_UnmodifiedSaveIsByteExact confirms
// SerializeDef's own byte-exact-when-unmodified guarantee (decision 005)
// holds for both real, vendored fixtures, not just hand-built ones.
func TestSerializeDef_RealFixtures_UnmodifiedSaveIsByteExact(t *testing.T) {
	for _, name := range []string{"mugen-2d-stage.def", "ikemen-go-3d-model-stage.def", "mugen-nondefault-scale-stage.def"} {
		t.Run(name, func(t *testing.T) {
			data := readTestdataFile(t, name)

			s, err := Parse(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			out, err := SerializeDef(data, s)
			if err != nil {
				t.Fatalf("SerializeDef: %v", err)
			}
			if !bytes.Equal(out, data) {
				t.Errorf("expected byte-exact unmodified save of %s (got %d bytes, want %d)", name, len(out), len(data))
			}
		})
	}
}
