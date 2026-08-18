# stage

A read/write Go library for MUGEN/Ikemen GO stage (background) `.def` files — BGdef, BG elements/layers, camera bounds, and stage boundaries — built as part of the [OpenKakutou](https://github.com/openkakutou) project. No rendering dependency; compiles to WebAssembly.

<!-- vibe:begin:features -->
This project is in early-stage development — writing/exporting stage files isn't possible yet.

Available now:

- A data model for stage definitions: stage-level settings, BG elements/layers (static, parallax, and animated backgrounds), camera scroll limits, and character movement limits
- Reading real MUGEN and Ikemen GO stage `.def` files into that model, including every background layer and Ikemen GO's tiling extension, with unrecognized content tolerated and malformed files reported with a clear, line-numbered error

Planned:

- Writing stage `.def` files back out, with format-preserving round-trip serialization — the same guarantee `character` provides for its own `.def`/`.air`/`.cns` files
- Resolving stage BG element sprite references against sprite sheets via [`sff`](https://github.com/openkakutou/sff)
- Resolving parallax scroll deltas and animated background playback state
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
Writing real `.def` stage files back out isn't implemented yet — that API will be documented here as it lands. Reading one, though, works today:

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
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/api.md](docs/api.md) — the package's public functions, starting with `Parse`, and how they behave
- [docs/data-model.md](docs/data-model.md) — the stage data types (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) and the exact `.def` file section/key each field maps to
<!-- vibe:end:docs-index -->
