package stage

// BGElementType is a BG element's rendering behavior, matching the literal
// tokens used by the .def BG element "type" key.
//
// The zero value ("") is not a valid .def token: MUGEN treats a missing
// "type" key as BGElementNormal, but applying that default is a parsing
// concern (backlog item 002), not something this pure-data model encodes
// silently — mirroring how air.Frame's fields carry no implicit defaults
// of their own either.
type BGElementType string

const (
	// BGElementNormal is a static sprite background element (.def BG
	// element "type" = "normal").
	BGElementNormal BGElementType = "normal"
	// BGElementParallax is a background element that scales and scrolls
	// to simulate depth (.def BG element "type" = "parallax").
	BGElementParallax BGElementType = "parallax"
	// BGElementAnim is a background element driven by an .air animation
	// action rather than a single static sprite (.def BG element "type" =
	// "anim").
	BGElementAnim BGElementType = "anim"
)

// SpriteRef identifies a sprite within the stage's sprite sheet by its
// (group, image) pair (.def BG element "spriteno").
type SpriteRef struct {
	Group int `json:"group"`
	Image int `json:"image"`
}

// BGElement is a single background element (a "[BG element_name]" section
// of a .def file): what to draw, where, on which layer, and how it scrolls
// and tiles.
type BGElement struct {
	// Name is this element's section name, e.g. the "floor" in a
	// "[BG floor]" header (.def "[BG <name>]" section header).
	Name string `json:"name"`
	// Type selects how this element is drawn and animated (.def BG
	// element "type").
	Type BGElementType `json:"type"`
	// Sprite identifies the static sprite shown for a BGElementNormal or
	// BGElementParallax element (.def BG element "spriteno"). Left at its
	// zero value when Type is BGElementAnim, which uses ActionNumber
	// instead.
	Sprite SpriteRef `json:"sprite"`
	// ActionNumber is the .air animation action number this element plays
	// when Type is BGElementAnim (.def BG element "actionno"). Unused for
	// other element types.
	ActionNumber int `json:"actionNumber"`
	// LayerNo selects draw order relative to characters: 0 draws behind
	// characters, 1 draws in front of them (.def BG element "layerno").
	LayerNo int `json:"layerNo"`
	// StartX and StartY are this element's starting position, in the
	// stage's local coordinate units (.def BG element "start").
	StartX int `json:"startX"`
	StartY int `json:"startY"`
	// DeltaX and DeltaY are the scroll ratio applied per unit of camera
	// movement: 1,1 scrolls at the same speed as the camera, values below
	// 1 scroll slower to simulate distance (a parallax "delta"). Values
	// are fractional (.def BG element "delta").
	DeltaX float64 `json:"deltaX"`
	DeltaY float64 `json:"deltaY"`
	// TileX and TileY control tiling of the element's sprite on each
	// axis: 0 disables tiling on that axis, 1 tiles infinitely, and any
	// value above 1 repeats the tile that many times (.def BG element
	// "tile").
	TileX int `json:"tileX"`
	TileY int `json:"tileY"`
	// TileSpacingX and TileSpacingY are the pixel gap between repeated
	// tiles on each axis, used only when tiling is enabled (.def BG
	// element "tilespacing").
	TileSpacingX int `json:"tileSpacingX"`
	TileSpacingY int `json:"tileSpacingY"`
}
