---
status: todo
depends_on: [003, 005]
---
# WASM Entrypoint And Release Pipeline

## Description
`stage-viewer-web` and `stage-editor` need to load and manipulate stages in the browser without a Go toolchain, the same way `character-viewer-web` does for characters. Add a build-tag-gated (`//go:build js && wasm`) `cmd/wasm/` entrypoint exposing this repo's read (and, once available, write) surface as JS-callable global functions, verified by a Node.js smoke test since `syscall/js` code can't run under the plain Go toolchain — mirroring `character`'s `cmd/wasm/` pattern exactly. Also add the GitHub Actions release workflow that builds and publishes `stage.wasm` + `wasm_exec.js` as downloadable assets on every tagged release, mirroring `character`'s own release pipeline.

## Acceptance Criteria
- [ ] `GOOS=js GOARCH=wasm go build ./cmd/wasm` produces a working `stage.wasm`
- [ ] A Node.js smoke test loads the built module and exercises at least one exported function against a real fixture, the same verification approach `character`'s `cmd/wasm` uses
- [ ] Exported functions follow the same JSON-in/JSON-out contract shape as `character`'s WASM exports (e.g. `LoadBytes`-equivalent for a stage `.def`)
- [ ] A GitHub Actions workflow publishes `stage.wasm` + `wasm_exec.js` as release assets on every tag
- [ ] Malformed input to a WASM export returns a JSON error field rather than throwing or crashing the WASM runtime

## Notes
None.
