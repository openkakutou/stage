# stage

A read/write Go library for MUGEN/Ikemen GO stage (background) `.def` files — BGdef, BG elements/layers, camera bounds, and stage boundaries — built as part of the [OpenKakutou](https://github.com/openkakutou) project. No rendering dependency; compiles to WebAssembly.

<!-- vibe:begin:features -->
This project is in early-stage development — no functionality yet.

Planned:

- A data model for stage definitions: BGdef, BG elements/layers, camera bounds, and stage boundaries
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
No functionality is implemented yet — the parsing/serialization API will be documented here as it lands. For now the package only exposes its version:

```go
package main

import (
	"fmt"

	"github.com/openkakutou/stage"
)

func main() {
	fmt.Println(stage.Version)
}
```
<!-- vibe:end:usage -->

<!-- vibe:begin:docs-index -->
No additional documentation yet.
<!-- vibe:end:docs-index -->
