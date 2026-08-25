package stage

import (
	"bytes"
	"fmt"
	"reflect"
)

// SerializeDef produces .def bytes for s, the write-path counterpart to
// Parse/ParseDocument, designed for a caller (e.g. a WASM host) that only
// ever holds a Stage as an edited in-memory value plus the original file's
// bytes — never a Document.
//
// original is the .def file's own previously loaded bytes — empty when s
// describes a brand new stage with no original file yet.
//
// When original is non-empty and s, once normalized the same way a JSON
// round trip already normalizes it (see normalizeStageForJSON), is
// unchanged from what parsing original itself produces, the original bytes
// are written back out verbatim — a byte-exact round trip, matching
// Document's existing guarantee. Otherwise (s was edited, or original is
// empty) fresh text is generated via Serialize, reflecting s's current
// values without preserving original's comments/ordering — see
// .vibe/decisions/005-wasm-entrypoint-mirrors-character-load-save-shape.md,
// mirroring character's own SerializeDef/.vibe/decisions/028.
//
// A malformed original returns a descriptive error (from ParseDocument)
// rather than silently falling back to a fresh serialize.
func SerializeDef(original []byte, s Stage) ([]byte, error) {
	edited := s
	normalizeStageForJSON(&edited)

	if len(original) > 0 {
		doc, err := ParseDocument(bytes.NewReader(original))
		if err != nil {
			return nil, fmt.Errorf("stage: parsing original stage definition for save: %w", err)
		}

		baseline := doc.Stage
		normalizeStageForJSON(&baseline)

		if reflect.DeepEqual(baseline, edited) {
			var buf bytes.Buffer
			if err := doc.Serialize(&buf); err != nil {
				return nil, fmt.Errorf("stage: writing unmodified stage definition: %w", err)
			}
			return buf.Bytes(), nil
		}
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("stage: serializing edited stage definition: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeStageForJSON replaces s.Elements' nil value with its non-nil
// empty equivalent, in place — needed so a caller round-tripping an
// unmodified Stage through JSON (where an absent/empty list becomes "[]",
// not "null") is never spuriously treated as an edit by SerializeDef,
// mirroring character's own normalizeCharacterInfoForJSON.
func normalizeStageForJSON(s *Stage) {
	if s.Elements == nil {
		s.Elements = []BGElement{}
	}
}
