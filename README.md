# stage

A read/write Go library for MUGEN/Ikemen GO stage (background) `.def` files — BGdef, BG elements/layers, camera bounds, and stage boundaries — built as part of the [OpenKakutou](https://github.com/openkakutou) project. No rendering dependency; compiles to WebAssembly.

<!-- vibe:begin:features -->
This project is in early-stage development — reading and writing real stage files isn't possible yet.

Defined so far:

- A data model for stage definitions: stage-level settings, BG elements/layers (static, parallax, and animated backgrounds), camera scroll limits, and character movement limits

Planned:

- Reading and writing stage `.def` files, MUGEN and Ikemen GO compatible
- Format-preserving round-trip serialization, the same guarantee `character` provides for its own `.def`/`.air`/`.cns` files
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
Reading and writing real `.def` stage files isn't implemented yet — that API will be documented here as it lands. For now, the package exposes its data model, which can already be built and inspected in memory:

```go
package main

import (
	"fmt"

	"github.com/openkakutou/stage"
)

func main() {
	s := stage.Stage{
		BGdef: stage.BGdef{SpriteFile: "stage0.sff", LocalCoordWidth: 320, LocalCoordHeight: 240},
		Elements: []stage.BGElement{
			{Name: "sky", Type: stage.BGElementNormal, Sprite: stage.SpriteRef{Group: 0, Image: 0}},
		},
		CameraBounds:    stage.CameraBounds{Left: -180, Right: 180, High: -240, Low: 0},
		StageBoundaries: stage.StageBoundaries{Left: -1000, Right: 1000},
	}
	fmt.Printf("%+v\n", s)
}
```
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
- [docs/data-model.md](docs/data-model.md) — the stage data types (`Stage`, `BGdef`, `BGElement`, `CameraBounds`, `StageBoundaries`) and the exact `.def` file section/key each field maps to
<!-- vibe:end:docs-index -->
