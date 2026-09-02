package stage

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stageCorpusDirEnv names the environment variable a caller sets to point
// at a local, machine-specific corpus of real .def stage files to scan —
// see .vibe/fixture-sources.md's "Local real-stage corpus" section. Never
// hardcode that path itself here: the corpus scan below is entirely
// skipped when this variable is unset, so this package's normal test run
// (CI included) never depends on it. Mirrors the identical convention the
// sibling `sff` repo's own corpus_compat_test.go established.
const stageCorpusDirEnv = "STAGE_CORPUS_DIR"

// TestCorpusCompat_RealDefFiles_ParseSuccessRate is the fixture-driven
// compatibility scan backlog item 007 asks for: every real, unmodified
// .def stage file under STAGE_CORPUS_DIR is parsed, then checked for two
// separate round-trip guarantees an unmodified real file must satisfy —
// Document's byte-exact preservation and SerializeDef's byte-exact save of
// an unedited stage. Any failure in any of the three checks fails the test
// loudly instead of being silently ignored, so a corpus file that hits a
// gap is caught here rather than shipping quietly broken.
//
// Skipped by default: this depends on a local, machine-specific corpus
// (see .vibe/fixture-sources.md) that is never vendored into this repo or
// available in CI.
func TestCorpusCompat_RealDefFiles_ParseSuccessRate(t *testing.T) {
	corpusDir := os.Getenv(stageCorpusDirEnv)
	if corpusDir == "" {
		t.Skipf("%s not set — skipping real-file corpus compatibility scan (see .vibe/fixture-sources.md)", stageCorpusDirEnv)
	}

	var files []string
	err := filepath.WalkDir(corpusDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".def") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking corpus directory %q: %v", corpusDir, err)
	}
	if len(files) == 0 {
		t.Fatalf("no .def files found under %q", corpusDir)
	}

	var parseFailures, docFailures, serializeFailures []string
	parsed := 0

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			parseFailures = append(parseFailures, fmt.Sprintf("%s: reading: %v", path, err))
			continue
		}

		s, err := Parse(bytes.NewReader(data))
		if err != nil {
			parseFailures = append(parseFailures, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		parsed++

		doc, err := ParseDocument(bytes.NewReader(data))
		if err != nil {
			docFailures = append(docFailures, fmt.Sprintf("%s: %v", path, err))
		} else {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				docFailures = append(docFailures, fmt.Sprintf("%s: serializing: %v", path, err))
			} else if !bytes.Equal(buf.Bytes(), data) {
				docFailures = append(docFailures, fmt.Sprintf("%s: not byte-exact (got %d bytes, want %d)", path, buf.Len(), len(data)))
			}
		}

		out, err := SerializeDef(data, s)
		if err != nil {
			serializeFailures = append(serializeFailures, fmt.Sprintf("%s: %v", path, err))
		} else if !bytes.Equal(out, data) {
			serializeFailures = append(serializeFailures, fmt.Sprintf("%s: not byte-exact on an unmodified save (got %d bytes, want %d)", path, len(out), len(data)))
		}
	}

	successRate := 0.0
	if len(files) > 0 {
		successRate = 100 * float64(parsed) / float64(len(files))
	}
	t.Logf("corpus scan: %d files under %s — %d parsed (%.1f%%), %d parse failures, %d Document round-trip failures, %d SerializeDef round-trip failures",
		len(files), corpusDir, parsed, successRate, len(parseFailures), len(docFailures), len(serializeFailures))

	report := func(label string, failures []string) {
		if len(failures) == 0 {
			return
		}
		max := len(failures)
		if max > 20 {
			max = 20
		}
		t.Errorf("%d %s (showing up to %d):\n%s", len(failures), label, max, strings.Join(failures[:max], "\n"))
	}
	report("file(s) failed to parse with an undocumented error", parseFailures)
	report("file(s) failed the Document byte-exact round trip", docFailures)
	report("file(s) failed the SerializeDef byte-exact round trip on an unmodified save", serializeFailures)
}
