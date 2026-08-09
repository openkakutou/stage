package stage

// BGdef holds stage-level settings that apply to the whole stage rather
// than to any single BG element: which sprite sheet the stage's elements
// draw from, the coordinate space element positions are expressed in, the
// ground level, and the camera's zoom range.
//
// BGdef is distinct from CameraBounds: BGdef.ZoomOut/ZoomIn are the
// camera's scale range, while CameraBounds is the box the camera's
// position (not its zoom) is clamped to.
type BGdef struct {
	// SpriteFile is the path to this stage's sprite sheet (.sff) file
	// (.def [BGDef] "spr").
	SpriteFile string `json:"spriteFile"`
	// LocalCoordWidth and LocalCoordHeight define the coordinate space BG
	// element positions and this stage's other pixel-based fields are
	// expressed in (.def [StageInfo] "localcoord", stored as two ints
	// rather than one width,height pair for direct field access).
	LocalCoordWidth  int `json:"localCoordWidth"`
	LocalCoordHeight int `json:"localCoordHeight"`
	// ZOffset is the ground level's vertical distance from the top of the
	// stage's local coordinate space, in pixels (.def [StageInfo]
	// "zoffset").
	ZOffset int `json:"zOffset"`
	// ZoomOut is the camera's maximum zoom-out scale factor, e.g. 0.75
	// means the camera can shrink the view to 75% (.def [Camera]
	// "zoomout").
	ZoomOut float64 `json:"zoomOut"`
	// ZoomIn is the camera's maximum zoom-in scale factor, e.g. 1.5 means
	// the camera can enlarge the view to 150% (.def [Camera] "zoomin").
	ZoomIn float64 `json:"zoomIn"`
}
