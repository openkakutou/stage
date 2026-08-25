#!/usr/bin/env node
// smoke.mjs is a Node.js verification harness for the WASM entrypoint built
// from this directory (see main.go) — it exercises the module the same way
// a browser consumer would (fetch/instantiate the .wasm, call the exposed
// global function, read back the result), without requiring an actual
// browser. It is not part of `go test` — syscall/js glue cannot run under
// the plain Go toolchain — and doubles as a minimal usage example for a JS
// consumer. Mirrors character's own cmd/wasm/smoke.mjs.
//
// Usage: node cmd/wasm/smoke.mjs [path/to/stage.wasm]
// (defaults to ./stage.wasm, relative to the repo root)

import { execSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..", "..");
const wasmPath = path.resolve(process.argv[2] || path.join(repoRoot, "stage.wasm"));

const goroot = execSync("go env GOROOT").toString().trim();
const wasmExecPath = path.join(goroot, "lib", "wasm", "wasm_exec.js");

// wasm_exec.js defines a global `Go` constructor; importing it for its
// side effect is the same pattern used to load it in a browser <script> tag.
await import(`file://${wasmExecPath}`);

const go = new globalThis.Go();
const { instance } = await WebAssembly.instantiate(readFileSync(wasmPath), go.importObject);
go.run(instance); // does not return: keeps the Go runtime (and its registered functions) alive

function toUint8Array(relativePath) {
	return new Uint8Array(readFileSync(path.join(repoRoot, relativePath)));
}

function assert(condition, message) {
	if (!condition) {
		console.error(`FAIL: ${message}`);
		process.exitCode = 1;
	} else {
		console.log(`ok - ${message}`);
	}
}

const defBytes = toUint8Array("cmd/wasm/testdata/sample.def");

// --- load: nominal path ---
const okResult = globalThis.OpenKakutouStage.load(defBytes);
assert(okResult.error === null, `nominal load reports no error (got: ${okResult.error})`);
assert(typeof okResult.stage === "string", "nominal load returns a stage JSON string");

const stage = JSON.parse(okResult.stage ?? "null");
assert(stage?.bgDef?.spriteFile === "stage0.sff", `stage spriteFile is "stage0.sff" (got: ${stage?.bgDef?.spriteFile})`);
assert(stage?.bgDef?.zOffset === 220, `stage zOffset is 220 (got: ${stage?.bgDef?.zOffset})`);
assert(Array.isArray(stage?.elements) && stage.elements.length === 2, "stage has 2 BG elements");
assert(stage?.elements?.[1]?.type === "parallax", `second element is parallax (got: ${stage?.elements?.[1]?.type})`);

// --- load: error path — malformed .def bytes, must not throw ---
const errResult = globalThis.OpenKakutouStage.load(new TextEncoder().encode("[BGDef\nspr = x\n"));
assert(errResult.stage === null, "malformed .def: stage is null");
assert(typeof errResult.error === "string" && errResult.error.length > 0, `malformed .def: error is a non-empty string (got: ${errResult.error})`);

// --- load: wrong argument count, must not crash the module ---
const argCountResult = globalThis.OpenKakutouStage.load();
assert(argCountResult.stage === null, "missing argument: stage is null");
assert(typeof argCountResult.error === "string" && argCountResult.error.length > 0, "missing argument: error is a non-empty string");

// The module must still respond correctly after an error.
const afterErrorResult = globalThis.OpenKakutouStage.load(defBytes);
assert(afterErrorResult.error === null, "module still works after a prior load error");

// --- save: no-edits round trip is byte-exact ---
const saveNoEditsResult = globalThis.OpenKakutouStage.save(defBytes, JSON.stringify(stage));
assert(saveNoEditsResult.error === null, `save (no edits) reports no error (got: ${saveNoEditsResult.error})`);
assert(saveNoEditsResult.bytes instanceof Uint8Array, "save (no edits) returns a Uint8Array");
assert(
	new TextDecoder().decode(saveNoEditsResult.bytes) === new TextDecoder().decode(defBytes),
	"save (no edits) is byte-identical to the original .def",
);

// --- save: an edit is reflected in the output ---
const editedStage = JSON.parse(JSON.stringify(stage));
editedStage.bgDef.zOffset = 999;
const saveEditedResult = globalThis.OpenKakutouStage.save(defBytes, JSON.stringify(editedStage));
assert(saveEditedResult.error === null, `save (edited) reports no error (got: ${saveEditedResult.error})`);
const savedText = new TextDecoder().decode(saveEditedResult.bytes);
assert(savedText.includes("zoffset = 999"), "save (edited) output contains the edited zoffset");

// --- save: brand new stage (empty original) ---
const newStage = { bgDef: { spriteFile: "new.sff", localCoordWidth: 320, localCoordHeight: 240 }, elements: [], cameraBounds: {}, stageBoundaries: {}, model: {}, scaling: {}, playerStartZ: {} };
const saveNewResult = globalThis.OpenKakutouStage.save(new Uint8Array(0), JSON.stringify(newStage));
assert(saveNewResult.error === null, `save (new stage) reports no error (got: ${saveNewResult.error})`);
assert(new TextDecoder().decode(saveNewResult.bytes).includes("new.sff"), "save (new stage) output contains the new spriteFile");

// --- save: malformed JSON returns a descriptive error instead of throwing ---
const saveBadJSONResult = globalThis.OpenKakutouStage.save(defBytes, "not json");
assert(saveBadJSONResult.bytes === null, "save (malformed JSON) returns null bytes");
assert(typeof saveBadJSONResult.error === "string" && saveBadJSONResult.error.length > 0, "save (malformed JSON) reports an error");

// The module must still respond correctly after the save calls above too.
const afterSaveResult = globalThis.OpenKakutouStage.load(defBytes);
assert(afterSaveResult.error === null, "module still works after the save calls");

if (process.exitCode) {
	console.error("\nsmoke test FAILED");
} else {
	console.log("\nsmoke test passed");
}
process.exit(process.exitCode ?? 0);
