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

// StageBoundaries is the x-range (and, for a model-based stage, z-range)
// characters are allowed to move within: it constrains characters, not the
// camera. See CameraBounds for the distinct, camera-facing concept (.def
// [PlayerInfo] section).
//
// No vertical (top/bottom, y-axis) movement bound is modeled: mainline
// MUGEN/Ikemen GO defines none — see
// .vibe/decisions/001-stage-boundaries-model-left-right-only.md. TopBound
// and BottomBound below are Ikemen GO's later, distinct z-axis (depth)
// extension for model-based stages, per that decision's own 2026-08-09
// update.
type StageBoundaries struct {
	// Left and Right are the minimum and maximum x-position a character
	// may move to (.def [PlayerInfo] "leftbound"/"rightbound").
	Left  int `json:"left"`
	Right int `json:"right"`
	// TopBound and BottomBound are the minimum and maximum z-position
	// (depth) a character may move to on a model-based stage (.def
	// [PlayerInfo] "topbound"/"botbound"), Ikemen GO's 3D stage extension.
	// Zero-valued (unused) for a traditional 2D stage. Named TopBound/
	// BottomBound rather than bare Top/Bottom to stay clearly distinct from
	// Scaling's TopZ/BottomZ, a different concept (perspective-scaling
	// anchor points, not a movement clamp) sourced from a different .def
	// section.
	TopBound    float64 `json:"topBound"`
	BottomBound float64 `json:"bottomBound"`
}

// PlayerStartZ holds each of up to 8 players' starting position on the
// depth (Z) axis (.def [PlayerInfo] "p1startz".."p8startz"), Ikemen GO's 3D
// stage extension. Unlike StageBoundaries, which is one pair of clamps
// shared by every character, this is a genuinely per-player value — kept as
// its own type rather than folded into StageBoundaries for that reason,
// even though both are sourced from the same [PlayerInfo] section.
//
// A stage with no z-axis players configured (a 2D stage, or a 3D stage that
// leaves a player at the engine default) leaves the corresponding field at
// its zero value, matching Ikemen GO's own behavior when a "pNstartz" key
// is omitted.
type PlayerStartZ struct {
	P1 int `json:"p1"`
	P2 int `json:"p2"`
	P3 int `json:"p3"`
	P4 int `json:"p4"`
	P5 int `json:"p5"`
	P6 int `json:"p6"`
	P7 int `json:"p7"`
	P8 int `json:"p8"`
}
