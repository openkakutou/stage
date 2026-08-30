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
assert(Array.isArray(stage?.elements) && stage.elements.length === 3, "stage has 3 BG elements");
assert(stage?.elements?.[1]?.type === "parallax", `second element is parallax (got: ${stage?.elements?.[1]?.type})`);
assert(stage?.elements?.[2]?.type === "anim", `third element is anim (got: ${stage?.elements?.[2]?.type})`);
assert(stage?.elements?.[2]?.actionNumber === 200, `third element's actionNumber is 200 (got: ${stage?.elements?.[2]?.actionNumber})`);

// --- load: the raw animation block referenced by the anim element is exposed ---
const torchAnim = stage?.animations?.["200"];
assert(torchAnim !== undefined, "stage.animations exposes the parsed action 200 block");
assert(Array.isArray(torchAnim?.frames) && torchAnim.frames.length === 2, `action 200 has 2 frames (got: ${torchAnim?.frames?.length})`);
assert(torchAnim?.frames?.[0]?.time === 10, `action 200's first frame lasts 10 ticks (got: ${torchAnim?.frames?.[0]?.time})`);

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

// --- resolveAnimationFrames: nominal — first frame at elapsed 0 ---
const torch = stage.animations["200"];
const nominalResult = globalThis.OpenKakutouStage.resolveAnimationFrames(
	JSON.stringify([{ animation: torch, elapsedTicks: 0 }]),
);
assert(nominalResult.error === null, `resolveAnimationFrames nominal reports no error (got: ${nominalResult.error})`);
assert(Array.isArray(nominalResult.sprites) && nominalResult.sprites.length === 1, "resolveAnimationFrames nominal returns 1 sprite");
assert(
	nominalResult.sprites[0].group === 0 && nominalResult.sprites[0].image === 0,
	`resolveAnimationFrames nominal resolves to frame 0's sprite (got: ${JSON.stringify(nominalResult.sprites[0])})`,
);

// --- resolveAnimationFrames: edge case — elapsed time past the loop point ---
// Action 200's total duration is 15 ticks (10 + 5), no explicit loop start
// (defaults to 0), so 15 + 4 = 19 ticks in must land back on frame 0.
const loopResult = globalThis.OpenKakutouStage.resolveAnimationFrames(
	JSON.stringify([{ animation: torch, elapsedTicks: 19 }]),
);
assert(loopResult.error === null, `resolveAnimationFrames loop-edge reports no error (got: ${loopResult.error})`);
assert(
	loopResult.sprites[0].group === 0 && loopResult.sprites[0].image === 0,
	`resolveAnimationFrames loop-edge wraps back to frame 0's sprite (got: ${JSON.stringify(loopResult.sprites[0])})`,
);

// --- resolveAnimationFrames: no matching BGAnimation for an element's ActionNumber ---
// Mirrors an anim element whose actionNumber has no parsed [Begin Action N]
// block at all — must resolve to the blank sentinel, not crash or error.
const noMatchResult = globalThis.OpenKakutouStage.resolveAnimationFrames(
	JSON.stringify([{ animation: null, elapsedTicks: 0 }]),
);
assert(noMatchResult.error === null, `resolveAnimationFrames no-match reports no error (got: ${noMatchResult.error})`);
assert(
	noMatchResult.sprites[0].group === -1 && noMatchResult.sprites[0].image === -1,
	`resolveAnimationFrames no-match resolves to the blank sentinel (got: ${JSON.stringify(noMatchResult.sprites[0])})`,
);

// --- resolveAnimationFrames: batching — several elements resolved in one call ---
const batchResult = globalThis.OpenKakutouStage.resolveAnimationFrames(
	JSON.stringify([
		{ animation: torch, elapsedTicks: 0 },
		{ animation: null, elapsedTicks: 0 },
		{ animation: torch, elapsedTicks: 12 },
	]),
);
assert(batchResult.error === null, `resolveAnimationFrames batch reports no error (got: ${batchResult.error})`);
assert(batchResult.sprites.length === 3, `resolveAnimationFrames batch returns 3 sprites (got: ${batchResult.sprites?.length})`);
assert(
	batchResult.sprites[0].group === 0 && batchResult.sprites[0].image === 0,
	"resolveAnimationFrames batch entry 0 (valid, elapsed 0) resolves to frame 0",
);
assert(
	batchResult.sprites[1].group === -1 && batchResult.sprites[1].image === -1,
	"resolveAnimationFrames batch entry 1 (no matching animation) resolves to the blank sentinel",
);
assert(
	batchResult.sprites[2].group === 0 && batchResult.sprites[2].image === 1,
	`resolveAnimationFrames batch entry 2 (valid, elapsed 12, into frame 1) resolves correctly (got: ${JSON.stringify(batchResult.sprites[2])})`,
);

// --- resolveAnimationFrames: malformed top-level JSON, must not crash ---
const badJSONResult = globalThis.OpenKakutouStage.resolveAnimationFrames("not json");
assert(badJSONResult.sprites === null, "resolveAnimationFrames malformed JSON: sprites is null");
assert(
	typeof badJSONResult.error === "string" && badJSONResult.error.length > 0,
	`resolveAnimationFrames malformed JSON: error is a non-empty string (got: ${badJSONResult.error})`,
);

// --- resolveAnimationFrames: wrong argument count, must not crash the module ---
const argCountAnimResult = globalThis.OpenKakutouStage.resolveAnimationFrames();
assert(argCountAnimResult.sprites === null, "resolveAnimationFrames missing argument: sprites is null");
assert(typeof argCountAnimResult.error === "string" && argCountAnimResult.error.length > 0, "resolveAnimationFrames missing argument: error is a non-empty string");

// The module must still respond correctly after the resolveAnimationFrames calls too.
const afterAnimResult = globalThis.OpenKakutouStage.load(defBytes);
assert(afterAnimResult.error === null, "module still works after the resolveAnimationFrames calls");

// --- resolveSprites: batched sprite pixel resolution (item 010) ---
// v1-basic.sff carries exactly one real sprite, at (group 0, image 0) —
// copied from character's own cmd/wasm/testdata, the same tiny real-file
// fixture that repo's own resolveSprites smoke test already exercises.
const sffBytes = toUint8Array("cmd/wasm/testdata/v1-basic.sff");

// --- nominal batch: one real sprite, one nonexistent (group, image) ---
const spritesResult = globalThis.OpenKakutouStage.resolveSprites(sffBytes, [[0, 0], [999, 999]], null);
assert(Array.isArray(spritesResult) && spritesResult.length === 2, "resolveSprites returns one result per request");

const [foundSprite, notFoundSprite] = spritesResult;
assert(foundSprite.error === null, `resolveSprites: real sprite reports no error (got: ${foundSprite.error})`);
assert(foundSprite.pixels instanceof Uint8Array, "resolveSprites: real sprite returns a pixel buffer");
assert(foundSprite.pixels.length === foundSprite.width * foundSprite.height * 4, "resolveSprites: pixel buffer length is width*height*4 (RGBA)");
assert(foundSprite.width > 0 && foundSprite.height > 0, `resolveSprites: real sprite has positive dimensions (got: ${foundSprite.width}x${foundSprite.height})`);

assert(notFoundSprite.pixels === null, "resolveSprites: nonexistent sprite returns null pixels");
assert(notFoundSprite.width === 0 && notFoundSprite.height === 0, "resolveSprites: nonexistent sprite reports 0x0 dimensions");
assert(
	typeof notFoundSprite.error === "string" && notFoundSprite.error.startsWith("sprite not found: "),
	`resolveSprites: nonexistent sprite error is distinguishable (got: ${notFoundSprite.error})`,
);

// --- external palette override recolors the sprite ---
const spriteActBytes = toUint8Array("cmd/wasm/testdata/cyclops-v1-palette1.act");
const overriddenSpriteResult = globalThis.OpenKakutouStage.resolveSprites(sffBytes, [[0, 0]], spriteActBytes);
assert(overriddenSpriteResult[0].error === null, `resolveSprites: override reports no error (got: ${overriddenSpriteResult[0].error})`);
const spriteColorsDiffer = overriddenSpriteResult[0].pixels.some((b, i) => b !== foundSprite.pixels[i]);
assert(spriteColorsDiffer, "resolveSprites: external palette override changes the resolved colors");

// --- undefined and null overrideBytes are equivalent to "no override" ---
const undefinedOverrideSpriteResult = globalThis.OpenKakutouStage.resolveSprites(sffBytes, [[0, 0]], undefined);
assert(
	undefinedOverrideSpriteResult[0].pixels.every((b, i) => b === foundSprite.pixels[i]),
	"resolveSprites: undefined overrideBytes matches the sprite's own palette",
);

// --- an explicitly empty overrideBytes is an error, not a silent fallback ---
const emptyOverrideSpriteResult = globalThis.OpenKakutouStage.resolveSprites(sffBytes, [[0, 0]], new Uint8Array(0));
assert(emptyOverrideSpriteResult[0].pixels === null, "resolveSprites: empty overrideBytes returns null pixels");
assert(
	typeof emptyOverrideSpriteResult[0].error === "string" && emptyOverrideSpriteResult[0].error.length > 0,
	"resolveSprites: empty overrideBytes reports an error",
);

// --- malformed sffBytes: no throw, every request in the batch reports an error ---
const malformedSpriteBatchResult = globalThis.OpenKakutouStage.resolveSprites(new TextEncoder().encode("garbage"), [[0, 0]], null);
assert(malformedSpriteBatchResult[0].pixels === null, "resolveSprites: malformed sffBytes returns null pixels");
assert(
	typeof malformedSpriteBatchResult[0].error === "string" && malformedSpriteBatchResult[0].error.length > 0,
	"resolveSprites: malformed sffBytes reports an error",
);

// The module must still respond correctly after resolveSprites errors too.
const afterResolveSpritesErrorResult = globalThis.OpenKakutouStage.load(defBytes);
assert(afterResolveSpritesErrorResult.error === null, "module still works after a prior resolveSprites error");

if (process.exitCode) {
	console.error("\nsmoke test FAILED");
} else {
	console.log("\nsmoke test passed");
}
process.exit(process.exitCode ?? 0);
