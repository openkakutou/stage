//go:build js && wasm

// Command wasm is the WASM entrypoint for the stage library: thin
// syscall/js glue exposing stage.Parse (via load) and stage.SerializeDef
// (via save) to a browser (or any JS host) as global functions, so a
// consumer can load and edit a MUGEN/Ikemen stage without a Go toolchain
// of its own.
//
// It carries no logic beyond argument conversion, calling into the root
// stage package, and marshaling results to JSON — all real behavior lives
// in that package, which is unit-tested independently of this file (see
// .vibe/decisions/005-wasm-entrypoint-mirrors-character-load-save-shape.md).
// This file's own behavior is instead verified by smoke.mjs, a Node.js
// script that loads the built module the way a real JS consumer would —
// syscall/js code cannot run under the plain `go test` toolchain. Mirrors
// character's own cmd/wasm/main.go.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"syscall/js"

	"github.com/openkakutou/sff"
	stage "github.com/openkakutou/stage"
)

func main() {
	globalName := "OpenKakutouStage"
	js.Global().Set(globalName, js.ValueOf(map[string]any{
		"load":                   js.FuncOf(load),
		"save":                   js.FuncOf(save),
		"resolveAnimationFrames": js.FuncOf(resolveAnimationFrames),
		"resolveSprites":         js.FuncOf(resolveSprites),
	}))

	// Registering js.FuncOf callbacks does not keep the Go runtime alive on
	// its own; block forever so OpenKakutouStage.load/save keep working for
	// the lifetime of the page.
	select {}
}

// load is OpenKakutouStage.load(defBytes) as seen from JS: defBytes is a
// Uint8Array (or any JS value js.CopyBytesToGo accepts) holding a stage
// .def file's raw bytes. It always returns a JS object shaped
// { stage: string|null, error: string|null } — exactly one of the two
// fields is non-null — never throws and never lets an internal panic
// escape to the JS caller.
func load(this js.Value, args []js.Value) any {
	defer func() {
		// A panic here would otherwise propagate out of the js.Func
		// callback and tear down the whole page's WASM instance; recover
		// is this boundary's own responsibility, not Parse's (which
		// already returns descriptive errors for every malformed-input
		// path it knows about).
		recover()
	}()

	if len(args) != 1 {
		return loadResult(nil, fmt.Errorf("OpenKakutouStage.load: expected 1 argument (defBytes), got %d", len(args)))
	}

	defBytes, err := bytesFromJS(args[0])
	if err != nil {
		return loadResult(nil, fmt.Errorf("OpenKakutouStage.load: defBytes: %w", err))
	}

	s, err := stage.Parse(bytes.NewReader(defBytes))
	if err != nil {
		return loadResult(nil, err)
	}

	data, err := json.Marshal(s)
	if err != nil {
		return loadResult(nil, fmt.Errorf("OpenKakutouStage.load: encoding result as JSON: %w", err))
	}

	return loadResult(data, nil)
}

// bytesFromJS copies a JS Uint8Array-like value into a Go []byte via
// js.CopyBytesToGo, the standard syscall/js conversion. It returns a
// descriptive error instead of panicking if v is not a byte-array-like
// value (e.g. undefined, or missing a numeric "length").
func bytesFromJS(v js.Value) (b []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			b, err = nil, fmt.Errorf("expected a byte array, got %v (%v)", v, r)
		}
	}()

	length := v.Get("length").Int()
	buf := make([]byte, length)
	js.CopyBytesToGo(buf, v)
	return buf, nil
}

// loadResult builds load's { stage, error } JS return shape. Exactly one
// field is ever non-null.
func loadResult(stageJSON []byte, err error) map[string]any {
	if err != nil {
		return map[string]any{"stage": nil, "error": err.Error()}
	}
	return map[string]any{"stage": string(stageJSON), "error": nil}
}

// save is OpenKakutouStage.save(originalDefBytes, editedStageJSON) as seen
// from JS: originalDefBytes is a Uint8Array holding the .def file's
// previously loaded bytes (an empty array for a brand new stage with no
// original file yet), editedStageJSON is a JSON string matching load's own
// "stage" field shape — the caller's current in-memory representation,
// edited or not. Returns { bytes, error }: on success bytes is a
// Uint8Array of the serialized .def file — byte-exact to originalDefBytes
// when editedStageJSON describes no real change, freshly generated text
// otherwise (see stage.SerializeDef) — and error is null; on failure bytes
// is null and error is a descriptive string. Never throws.
func save(this js.Value, args []js.Value) any {
	defer func() {
		// See load's identical recover() — a panic here would otherwise
		// tear down the whole page's WASM instance.
		recover()
	}()

	if len(args) != 2 {
		return saveResult(nil, fmt.Errorf("OpenKakutouStage.save: expected 2 arguments (originalDefBytes, editedStageJSON), got %d", len(args)))
	}

	original, err := bytesFromJS(args[0])
	if err != nil {
		return saveResult(nil, fmt.Errorf("OpenKakutouStage.save: originalDefBytes: %w", err))
	}

	var s stage.Stage
	if err := json.Unmarshal([]byte(args[1].String()), &s); err != nil {
		return saveResult(nil, fmt.Errorf("OpenKakutouStage.save: editedStageJSON: %w", err))
	}

	out, err := stage.SerializeDef(original, s)
	if err != nil {
		return saveResult(nil, err)
	}
	return saveResult(out, nil)
}

// saveResult builds save's { bytes, error } JS return shape. Exactly one
// field is ever non-null. bytes is transferred as a Uint8Array, since the
// caller's next step is typically offering it as a browser download.
func saveResult(fileBytes []byte, err error) map[string]any {
	if err != nil {
		return map[string]any{"bytes": nil, "error": err.Error()}
	}
	arr := js.Global().Get("Uint8Array").New(len(fileBytes))
	js.CopyBytesToJS(arr, fileBytes)
	return map[string]any{"bytes": arr, "error": nil}
}

// animationFrameRequest is one entry of resolveAnimationFrames's input
// array: the raw animation data for one BG element (typically read
// straight from a prior load result's stage.animations[actionNumber]) and
// how far into its playback that element currently is. Animation is a
// pointer so a JSON `null` (an ActionNumber with no matching parsed
// "[Begin Action N]" block) decodes to a nil value rather than a
// zero-value struct indistinguishable from "present but empty" — both
// resolve identically below, but keeping them distinct in the decoded
// request costs nothing and matches the JSON's own shape.
type animationFrameRequest struct {
	Animation    *stage.BGAnimation `json:"animation"`
	ElapsedTicks int                `json:"elapsedTicks"`
}

// resolveAnimationFrames is OpenKakutouStage.resolveAnimationFrames(requestsJSON)
// as seen from JS: requestsJSON is a JSON-encoded array of
// { animation, elapsedTicks } requests, one per animated BG element that
// needs resolving this rendered frame — a single call batches every
// element rather than one WASM call each, since each JS↔WASM crossing
// carries fixed overhead that would otherwise multiply by element count on
// every single animation frame (see
// .vibe/decisions/007-batched-resolve-animation-frames-wasm-call.md).
// A request's animation is the same BGAnimation shape load's own
// stage.animations[actionNumber] already exposes, or null when no block
// matches that element's ActionNumber.
//
// Returns { sprites, error }: on success sprites is an array of
// { group, image } sprite references, same order and length as the input
// requests — a request with a null/malformed animation resolves to the
// blank sentinel { group: -1, image: -1 } (mirrors stage.ResolveAnimationFrame's
// own "never panics, always returns a SpriteRef" contract) rather than
// failing that entry or the whole call. error is only set (sprites null)
// when requestsJSON itself cannot be parsed, or the argument count is
// wrong. Never throws.
func resolveAnimationFrames(this js.Value, args []js.Value) any {
	defer func() {
		// See load's identical recover() — a panic here would otherwise
		// tear down the whole page's WASM instance.
		recover()
	}()

	if len(args) != 1 {
		return resolveAnimationFramesResult(nil, fmt.Errorf("OpenKakutouStage.resolveAnimationFrames: expected 1 argument (requestsJSON), got %d", len(args)))
	}

	var requests []animationFrameRequest
	if err := json.Unmarshal([]byte(args[0].String()), &requests); err != nil {
		return resolveAnimationFramesResult(nil, fmt.Errorf("OpenKakutouStage.resolveAnimationFrames: requestsJSON: %w", err))
	}

	sprites := make([]map[string]any, len(requests))
	for i, req := range requests {
		var anim stage.BGAnimation
		if req.Animation != nil {
			anim = *req.Animation
		}
		sprite := stage.ResolveAnimationFrame(anim, req.ElapsedTicks)
		sprites[i] = map[string]any{"group": sprite.Group, "image": sprite.Image}
	}

	return resolveAnimationFramesResult(sprites, nil)
}

// resolveAnimationFramesResult builds resolveAnimationFrames's
// { sprites, error } JS return shape. Exactly one field is ever non-null.
func resolveAnimationFramesResult(sprites []map[string]any, err error) map[string]any {
	if err != nil {
		return map[string]any{"sprites": nil, "error": err.Error()}
	}
	result := make([]any, len(sprites))
	for i, s := range sprites {
		result[i] = s
	}
	return map[string]any{"sprites": result, "error": nil}
}

// resolveSprites is OpenKakutouStage.resolveSprites(sffBytes, requests,
// overrideBytes) as seen from JS: sffBytes is a Uint8Array holding a
// loaded .sff sprite sheet's raw bytes, requests is a JS array of
// [group, image] pairs to resolve against it, and overrideBytes is an
// optional Uint8Array of external .act palette bytes (undefined/null means
// "use the sprite's own palette"). Returns one { pixels, width, height,
// error } result per request, same order — batched in a single call
// rather than one per sprite, mirroring the sibling character repo's own
// resolveSprites (see .vibe/backlog/010, which this item follows). A
// request naming a sprite absent from the sheet, or malformed sffBytes,
// never throws or crashes the module — each affected entry reports a
// descriptive error instead. This function carries no logic of its own
// beyond argument conversion and batching: all real decoding lives in the
// external sff module's ResolveSpritePixels.
func resolveSprites(this js.Value, args []js.Value) any {
	defer func() {
		// See load's identical recover() — a panic here would otherwise
		// tear down the whole page's WASM instance.
		recover()
	}()

	if len(args) != 3 {
		return []any{resolveSpritesResult(nil, 0, 0, fmt.Errorf("OpenKakutouStage.resolveSprites: expected 3 arguments (sffBytes, requests, overrideBytes), got %d", len(args)))}
	}

	sffBytes, err := bytesFromJS(args[0])
	if err != nil {
		return []any{resolveSpritesResult(nil, 0, 0, fmt.Errorf("OpenKakutouStage.resolveSprites: sffBytes: %w", err))}
	}

	override, overrideErr := overridePaletteFromJS(args[2])

	n := args[1].Get("length").Int()
	r := bytes.NewReader(sffBytes)
	results := make([]any, n)
	for i := 0; i < n; i++ {
		if overrideErr != nil {
			results[i] = resolveSpritesResult(nil, 0, 0, fmt.Errorf("OpenKakutouStage.resolveSprites: overrideBytes: %w", overrideErr))
			continue
		}
		group, image, err := spriteRequestFromJS(args[1].Index(i))
		if err != nil {
			results[i] = resolveSpritesResult(nil, 0, 0, fmt.Errorf("OpenKakutouStage.resolveSprites: requests[%d]: %w", i, err))
			continue
		}
		pixels, width, height, err := sff.ResolveSpritePixels(r, group, image, override)
		results[i] = resolveSpritesResult(pixels, width, height, err)
	}
	return results
}

// overridePaletteFromJS decodes v as an external .act palette override, or
// returns (nil, nil) when v is JS undefined/null — the two values a
// caller uses to mean "no override, use the sprite's own palette". Any
// other value, including an empty Uint8Array, is decoded via
// sff.DecodeExternalPalette and its own validation (wrong size, etc.)
// surfaces as err, never a silent fallback to no override.
func overridePaletteFromJS(v js.Value) (*sff.Palette, error) {
	if v.IsUndefined() || v.IsNull() {
		return nil, nil
	}
	b, err := bytesFromJS(v)
	if err != nil {
		return nil, err
	}
	p, err := sff.DecodeExternalPalette(b)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// spriteRequestFromJS reads one [group, image] pair from requests[i] as
// seen from JS. Returns a descriptive error instead of panicking if v is
// not a length-2 array-like value.
func spriteRequestFromJS(v js.Value) (group, image int, err error) {
	defer func() {
		if r := recover(); r != nil {
			group, image, err = 0, 0, fmt.Errorf("expected a [group, image] pair, got %v (%v)", v, r)
		}
	}()

	if length := v.Get("length").Int(); length != 2 {
		return 0, 0, fmt.Errorf("expected a [group, image] pair (length 2), got length %d", length)
	}
	return v.Index(0).Int(), v.Index(1).Int(), nil
}

// resolveSpritesResult builds resolveSprites's { pixels, width, height,
// error } JS return shape for one resolved sprite. On error,
// pixels/width/height are explicitly nil/0/0 rather than left undefined.
func resolveSpritesResult(pixels []color.RGBA, width, height int, err error) map[string]any {
	if err != nil {
		return map[string]any{"pixels": nil, "width": 0, "height": 0, "error": err.Error()}
	}
	return map[string]any{"pixels": rgbaToJS(pixels), "width": width, "height": height, "error": nil}
}

// rgbaToJS converts a slice of decoded RGBA pixels into a flat JS
// Uint8Array (4 bytes per pixel, R/G/B/A in order) — the same transfer
// shape the sibling character repo's own resolveSprites uses.
func rgbaToJS(pixels []color.RGBA) js.Value {
	buf := make([]byte, len(pixels)*4)
	for i, p := range pixels {
		buf[i*4], buf[i*4+1], buf[i*4+2], buf[i*4+3] = p.R, p.G, p.B, p.A
	}
	arr := js.Global().Get("Uint8Array").New(len(buf))
	js.CopyBytesToJS(arr, buf)
	return arr
}
