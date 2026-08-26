package stage

// Stage is a MUGEN/Ikemen GO stage (background): its name/author, BGdef
// settings, BG elements/layers, camera bounds, and stage boundaries — the
// read-path pure-data model for a stage .def file's content.
//
// This is the read-path surface — the stable vocabulary a library consumer
// (viewer, editor) works with. It carries no .def parsing, file I/O, or
// write-only (format-preservation) logic; per CLAUDE.md's read/write
// separation constraint, that lives elsewhere so importing this data model
// alone never pulls in write-only dependencies. No parser populates it yet
// (see backlog item 002).
//
// A zero-value Stage is valid and usable: Elements is nil (ranges over it
// safely) and every nested settings/bounds type is itself zero-value-safe.
type Stage struct {
	// Name is the stage's display name (.def [Info] "name"). Empty when
	// the source .def has no [Info] section, or omits the key.
	Name string `json:"name"`
	// Author is the stage's author/creator (.def [Info] "author"). Empty
	// when the source .def has no [Info] section, or omits the key.
	Author string `json:"author"`
	// BGdef holds this stage's stage-level settings (sprite sheet,
	// coordinate space, ground level, camera zoom range).
	BGdef BGdef `json:"bgDef"`
	// Elements are this stage's BG elements/layers, in .def file order.
	Elements []BGElement `json:"elements"`
	// CameraBounds is the box the camera's own position is clamped to.
	CameraBounds CameraBounds `json:"cameraBounds"`
	// StageBoundaries is the x-range (and, for a model-based stage,
	// z-range) characters are allowed to move within, distinct from
	// CameraBounds.
	StageBoundaries StageBoundaries `json:"stageBoundaries"`
	// Model is this stage's 3D model placement/lighting settings, Ikemen
	// GO's 3D stage extension. Zero-valued (unused) unless BGdef.ModelFile
	// names a model file.
	Model Model `json:"model"`
	// Scaling is this stage's 3D perspective-scaling settings. Zero-valued
	// (unused) unless BGdef.ModelFile names a model file.
	Scaling Scaling `json:"scaling"`
	// PlayerStartZ holds each player's starting depth (Z) position on a
	// model-based stage. Zero-valued (unused) unless BGdef.ModelFile names
	// a model file.
	PlayerStartZ PlayerStartZ `json:"playerStartZ"`
}
