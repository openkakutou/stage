package stage

// Scaling holds Ikemen GO's 3D perspective-scaling settings (.def
// "[Scaling]" section): how a character's on-screen size changes with depth
// (Z) position, and how Z position translates to on-screen Y offset. Only
// meaningful alongside a model-based stage (BGdef.ModelFile) — a 2D stage
// leaves this at its zero value.
type Scaling struct {
	// DepthToScreen determines how a player's Z position affects their Y
	// offset on screen: 1 (the default) means 1 pixel in Z space equates to
	// 1 pixel of vertical screen offset (.def "[Scaling]" "depthtoscreen").
	DepthToScreen float64 `json:"depthToScreen"`
	// TopZ and BottomZ are the Z-space reference points TopScale and
	// BottomScale apply at (.def "[Scaling]" "topz"/"botz"). Distinct from
	// StageBoundaries.TopBound/BottomBound, which clamp character movement
	// rather than anchor a perspective-scaling range — see that type's own
	// doc comment.
	TopZ    float64 `json:"topZ"`
	BottomZ float64 `json:"bottomZ"`
	// TopScale and BottomScale are the on-screen scale factors applied at
	// TopZ and BottomZ respectively, interpolated in between (.def
	// "[Scaling]" "topscale"/"botscale").
	TopScale    float64 `json:"topScale"`
	BottomScale float64 `json:"bottomScale"`
}
