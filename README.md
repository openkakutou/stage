# stage

A read/write Go library for MUGEN/Ikemen GO stage (background) `.def` files — BGdef, BG elements/layers, camera bounds, and stage boundaries — built as part of the [OpenKakutou](https://github.com/openkakutou) project. No rendering dependency; compiles to WebAssembly.

<!-- vibe:begin:features -->
This project is in early-stage development.

Available now:

- A data model for stage definitions: stage-level settings, BG elements/layers (static, parallax, and animated backgrounds), camera scroll limits, and character movement limits
- Reading real MUGEN and Ikemen GO stage `.def` files into that model, including every background layer and Ikemen GO's tiling extension, with unrecognized content tolerated and malformed files reported with a clear, line-numbered error
- Writing stage `.def` files back out: a fresh-write path for a stage built or edited in memory, and a format-preserving path that reproduces an unmodified file's comments, section ordering, and unrecognized content byte-for-byte instead of overwriting them
- Resolving a BG element's sprite reference to real pixel data from a loaded sprite sheet, via [`sff`](https://github.com/openkakutou/sff), regardless of which `.sff` file version it came from
- Ikemen GO's 3D model-based stage extension: a stage can reference a 3D model with its own placement, scaling, and image-based lighting, plus depth-based (Z-axis) camera perspective, character movement limits, and per-player starting positions — a stage that doesn't use any of this reads and writes exactly as before
- Computing where a scrolling background layer should appear on screen as the camera moves (parallax depth), and which frame an animated background layer should currently show at a given moment in time, including looping the animation once it finishes

Planned:

- Reading an animated background layer's frame sequence out of a stage file
- A WebAssembly build so web apps can load a stage without a Go toolchain
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
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the package's public functions — `Parse`, `Serialize`, `Document`/`ParseDocument` — and how they behave
- [docs/data-model.md](docs/data-model.md) — the stage data types (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) and the exact `.def` file section/key each field maps to
<!-- vibe:end:docs-index -->
