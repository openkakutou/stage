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
// When s differs from that baseline only in BGCtrlDefs, with the same
// number of groups and the same number of controllers in each, the original
// bytes are patched in place instead: only the lines of the changed
// groups/controllers are altered (a changed value is replaced on its own
// line, a new key is appended after the section's last key line, a removed
// key's line is deleted), so comments, ordering, line endings and every
// untouched section stay byte-identical. If the edit cannot be patched
// safely (or the patched text would not parse back to s), the fresh
// Serialize path above is used. Document is unchanged: mutating
// Document.Stage still has no effect on Document.Serialize.
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

		if patched, ok := patchBGCtrlDefs(original, baseline, edited); ok {
			return patched, nil
		}
	}

	var buf bytes.Buffer
	if err := Serialize(&buf, edited); err != nil {
		return nil, fmt.Errorf("stage: serializing edited stage definition: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeStageForJSON returns a deep copy of s in which every nil slice
// or map reachable from it is replaced by an empty non-nil one. The rule is
// generic (reflect-based), so no field list is kept here: a new collection
// field is covered automatically. It is needed so a caller round-tripping
// an unmodified Stage through JSON (where an absent/empty list becomes
// "[]", not "null") is never spuriously treated as an edit by SerializeDef,
// mirroring character's own normalizeCharacterInfoForJSON. s itself, and
// the arrays it shares, are never written to.
func normalizeStageForJSON(s *Stage) {
	out := reflect.New(reflect.TypeOf(*s)).Elem()
	normalizeValue(out, reflect.ValueOf(*s))
	*s = out.Interface().(Stage)
}

func normalizeValue(dst, src reflect.Value) {
	switch src.Kind() {
	case reflect.Slice:
		n := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
		for i := 0; i < src.Len(); i++ {
			normalizeValue(n.Index(i), src.Index(i))
		}
		dst.Set(n)
	case reflect.Map:
		n := reflect.MakeMapWithSize(src.Type(), src.Len())
		iter := src.MapRange()
		for iter.Next() {
			v := reflect.New(src.Type().Elem()).Elem()
			normalizeValue(v, iter.Value())
			n.SetMapIndex(iter.Key(), v)
		}
		dst.Set(n)
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if !dst.Field(i).CanSet() {
				continue
			}
			normalizeValue(dst.Field(i), src.Field(i))
		}
	case reflect.Ptr:
		if src.IsNil() {
			return
		}
		n := reflect.New(src.Type().Elem())
		normalizeValue(n.Elem(), src.Elem())
		dst.Set(n)
	default:
		dst.Set(src)
	}
}
