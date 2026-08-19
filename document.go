package stage

import (
	"bytes"
	"fmt"
	"io"
)

// Document is the write-path counterpart to Parse/Serialize: it exists so a
// stage .def file can be round-tripped — parsed, then serialized back out —
// without losing the comments, section ordering, and unrecognized sections
// that the pure-data Stage model deliberately does not carry, mirroring
// character's own def.Document.
//
// Document.Stage is decoded the same way Parse's return value is, for
// convenient structured access to what was parsed — but Serialize does not
// read it back. As long as Document.Stage is left untouched, ParseDocument
// followed by Serialize reproduces the original source byte-for-byte,
// comments and all. Mutating Stage has no effect on Serialize's output:
// regenerating text from an edited Stage while still preserving unrelated
// comments/sections/ordering around the edit is a heavier per-line
// reconciliation this type does not attempt.
type Document struct {
	Stage Stage

	source []byte
}

// ParseDocument reads MUGEN/Ikemen GO stage .def text from r, decoding it
// the same way Parse does while also retaining the exact source bytes
// needed for a faithful round trip through Serialize.
func ParseDocument(r io.Reader) (*Document, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("stage: reading document source: %w", err)
	}

	s, err := Parse(bytes.NewReader(source))
	if err != nil {
		return nil, err
	}

	return &Document{Stage: s, source: source}, nil
}

// Serialize writes the Document's retained source back out to w verbatim,
// reproducing the exact text ParseDocument read — including comments,
// section ordering, unrecognized sections, and original line endings.
func (d *Document) Serialize(w io.Writer) error {
	if _, err := w.Write(d.source); err != nil {
		return fmt.Errorf("stage: writing document: %w", err)
	}
	return nil
}
