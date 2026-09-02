# stage

A read/write Go library for MUGEN/Ikemen GO stage (background) `.def` files — BGdef, BG elements/layers, camera bounds, and stage boundaries — built as part of the [OpenKakutou](https://github.com/openkakutou) project. No rendering dependency; compiles to WebAssembly.

<!-- vibe:begin:features -->
This project is in early-stage development.

Available now:

- A data model for stage definitions: name and author, stage-level settings, BG elements/layers (static, parallax, and animated backgrounds), camera scroll limits, and character movement limits
- Reading real MUGEN and Ikemen GO stage `.def` files into that model, including its name/author, every background layer, and Ikemen GO's tiling extension, with unrecognized content tolerated and malformed files reported with a clear, line-numbered error
- Writing stage `.def` files back out: a fresh-write path for a stage built or edited in memory, and a format-preserving path that reproduces an unmodified file's comments, section ordering, and unrecognized content byte-for-byte instead of overwriting them
- Resolving a BG element's sprite reference to real pixel data from a loaded sprite sheet, via [`sff`](https://github.com/openkakutou/sff), regardless of which `.sff` file version it came from
- Ikemen GO's 3D model-based stage extension: a stage can reference a 3D model with its own placement, scaling, and image-based lighting, plus depth-based (Z-axis) camera perspective, character movement limits, and per-player starting positions — a stage that doesn't use any of this reads and writes exactly as before
- Reading an animated background layer's actual frame sequence out of a stage file — wherever in the file it's declared — and computing where a scrolling background layer should appear on screen as the camera moves (parallax depth) and which frame an animated one should currently show at a given moment in time, including looping the animation once it finishes; saving a stage writes that frame data back out for every animated layer that still uses it
- Loading and saving a stage from a web browser, with no Go toolchain needed: a WebAssembly build of this library is published as a downloadable file on every release, so web apps like `stage-viewer-web` and `stage-editor` can use it directly
- Resolving which frame each animated background layer should currently show, for as many layers as needed in one call, directly from a web browser — no need to reimplement the frame-timing logic in JavaScript
- Resolving a stage's background sprites into actual displayable pixels directly from a web browser, for as many sprites as needed in one call, with support for previewing an alternate color palette — a missing sprite or a corrupted sprite sheet is reported with a clear error instead of breaking the preview
- Loading is validated against a large corpus of real MUGEN and Ikemen GO stage files, not just hand-built test data — real-world authoring habits like a shorthand tiling value, a decimal point on a whole-number setting, or a blank animation timing value are all read correctly instead of being rejected
<!-- vibe:end:features -->

<!-- vibe:begin:install -->
Requires [Go](https://go.dev/) 1.26 or later.

```sh
go get github.com/openkakutou/stage
```

Verify the install by importing the module in a Go file and running `go build`:

```go
import "github.com/openkakutou/stage"
```

To update to the latest version:

```sh
go get -u github.com/openkakutou/stage
```
<!-- vibe:end:install -->

<!-- vibe:begin:usage -->
Reading a `.def` stage file into the data model:

```go
package main

import (
	"fmt"
	"os"

	"github.com/openkakutou/stage"
)

func main() {
	f, err := os.Open("stage0.def")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s, err := stage.Parse(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", s)
}
```

The data model can also be built and inspected directly in memory, without going through `.def` text:

```go
s := stage.Stage{
	BGdef: stage.BGdef{SpriteFile: "stage0.sff", LocalCoordWidth: 320, LocalCoordHeight: 240},
	Elements: []stage.BGElement{
		{Name: "sky", Type: stage.BGElementNormal, Sprite: stage.SpriteRef{Group: 0, Image: 0}},
	},
	CameraBounds:    stage.CameraBounds{Left: -180, Right: 180, High: -240, Low: 0},
	StageBoundaries: stage.StageBoundaries{Left: -1000, Right: 1000},
}
```

Writing a `Stage` back out to `.def` text — for a stage built or edited in memory, without preserving any original file's formatting:

```go
f, err := os.Create("stage0.def")
if err != nil {
	panic(err)
}
defer f.Close()

if err := stage.Serialize(f, s); err != nil {
	panic(err)
}
```

Editing an existing file while preserving everything about it you didn't change — comments, section ordering, unrecognized sections — as long as the parsed `Stage` itself is left untouched:

```go
f, err := os.Open("stage0.def")
if err != nil {
	panic(err)
}
doc, err := stage.ParseDocument(f)
f.Close()
if err != nil {
	panic(err)
}

out, err := os.Create("stage0.def")
if err != nil {
	panic(err)
}
defer out.Close()

if err := doc.Serialize(out); err != nil {
	panic(err)
}
```

Resolving a BG element's sprite reference against sprite sheets loaded via [`sff`](https://github.com/openkakutou/sff):

```go
import "github.com/openkakutou/sff"

f, err := os.Open("stage0.sff")
if err != nil {
	panic(err)
}
defer f.Close()

groups, err := sff.Load(f)
if err != nil {
	panic(err)
}

resolver := stage.NewSpriteResolver(groups)
sprite, err := resolver.Resolve(s.Elements[0].Sprite)
if err != nil {
	panic(err)
}
fmt.Printf("%+v\n", sprite)
```

Computing a parallax layer's on-screen position as the camera scrolls, and which sprite an animated layer should currently show:

```go
element := s.Elements[0]
x, y := stage.ResolveParallaxPosition(element, cameraX, cameraY)

anim := stage.BGAnimation{
	Frames: []stage.BGAnimFrame{
		{Sprite: stage.SpriteRef{Group: 9, Image: 0}, Time: 10},
		{Sprite: stage.SpriteRef{Group: 9, Image: 1}, Time: 5},
	},
}
currentSprite := stage.ResolveAnimationFrame(anim, elapsedTicks)
```

### Loading a stage in a web browser (WebAssembly)

A web app with no Go toolchain of its own can load (and save) a stage too, using a pre-built WebAssembly module downloaded from a tagged release's assets (`stage.wasm` + `wasm_exec.js`):

```html
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("stage.wasm"), go.importObject)
    .then((result) => go.run(result.instance))
    .then(async () => {
      const defBytes = await fetch("stage0.def").then((r) => r.arrayBuffer()).then((buf) => new Uint8Array(buf));

      const result = globalThis.OpenKakutouStage.load(defBytes);
      if (result.error) {
        throw new Error(result.error);
      }

      const stage = JSON.parse(result.stage);
      console.log(`${stage.bgDef.spriteFile}: ${stage.elements.length} BG elements`);
    });
</script>
```

See [docs/wasm.md](docs/wasm.md) for the full JS API contract and how to build the module locally.
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the package's public functions — `Parse`, `Serialize`, `Document`/`ParseDocument`, `SerializeDef` — and how they behave
- [docs/data-model.md](docs/data-model.md) — the stage data types (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) and the exact `.def` file section/key each field maps to
- [docs/testing.md](docs/testing.md) — how the test suite is structured, including the real-file corpus compatibility scan and the vendored real-file fixtures
- [docs/wasm.md](docs/wasm.md) — the WebAssembly entrypoint's JS API, how to build it locally, and the release pipeline that publishes it
<!-- vibe:end:docs-index -->
