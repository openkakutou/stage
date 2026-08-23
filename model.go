package stage

// Model is a 3D stage's model-based placement and lighting settings (.def
// "[Model]" section), Ikemen GO's 3D stage extension (see the roadmap's
// .vibe/decisions/014). The model file itself is referenced separately by
// BGdef.ModelFile (.def "[BGDef]" "model"), mirroring how BGdef.SpriteFile
// already holds the 2D sprite sheet path — verified against Ikemen GO's own
// source, "[Model]" carries only placement/lighting settings for whichever
// file BGdef.ModelFile names, never the file path itself.
//
// Real Ikemen GO stage .def data confirms exactly one "[Model]" section per
// stage — unlike a BG element, it is not a repeatable "[Model name]"
// section keyed by a mesh name.
type Model struct {
	// OffsetX, OffsetY, and OffsetZ place the model's origin in the 3D
	// scene (.def "[Model]" "offset").
	OffsetX float64 `json:"offsetX"`
	OffsetY float64 `json:"offsetY"`
	OffsetZ float64 `json:"offsetZ"`
	// ScaleX, ScaleY, and ScaleZ scale the model on each axis (.def
	// "[Model]" "scale").
	ScaleX float64 `json:"scaleX"`
	ScaleY float64 `json:"scaleY"`
	ScaleZ float64 `json:"scaleZ"`
	// Environment is the path to an ".hdr" file used for the model's
	// image-based lighting (.def "[Model]" "environment").
	Environment string `json:"environment"`
	// EnvironmentIntensity scales how strongly Environment's lighting
	// affects the model (.def "[Model]" "environmentintensity").
	EnvironmentIntensity float64 `json:"environmentIntensity"`
}
