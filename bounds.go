package stage

// CameraBounds is the box the camera's own position is clamped to as it
// scrolls: it constrains the camera, not characters. See StageBoundaries
// for the distinct, character-facing concept (.def [Camera] section).
//
// This deliberately models only the four scroll-clamp edges, not the
// camera's zoom range — zoom belongs to BGdef (.def [Camera]
// "zoomout"/"zoomin"), a separate stage-level setting rather than a bound
// on position.
type CameraBounds struct {
	// Left and Right are the minimum and maximum x-position the camera can
	// scroll to. Left is conventionally negative (.def [Camera]
	// "boundleft"/"boundright").
	Left  int `json:"left"`
	Right int `json:"right"`
	// High and Low are the minimum and maximum y-position the camera can
	// scroll to. High is conventionally negative or zero (.def [Camera]
	// "boundhigh"/"boundlow").
	High int `json:"high"`
	Low  int `json:"low"`
}

// StageBoundaries is the x-range characters are allowed to move within: it
// constrains characters, not the camera. See CameraBounds for the distinct,
// camera-facing concept (.def [PlayerInfo] section).
//
// Only a horizontal range is modeled: mainline MUGEN/Ikemen GO defines no
// vertical (top/bottom) movement bound for characters — see
// .vibe/decisions/001-stage-boundaries-model-left-right-only.md.
type StageBoundaries struct {
	// Left and Right are the minimum and maximum x-position a character
	// may move to (.def [PlayerInfo] "leftbound"/"rightbound").
	Left  int `json:"left"`
	Right int `json:"right"`
}
