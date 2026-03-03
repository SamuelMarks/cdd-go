# WASM Support

The `cdd-go` compiler fully supports compiling to WebAssembly out of the box using Go's native toolchain (`GOOS=js GOARCH=wasm`).

## Building

Run the Makefile target:
```bash
make build_wasm
```

This generates `bin/cdd-go.wasm`.

## Execution

You can run the generated WebAssembly file in an environment like Node.js:
```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./bin/
node ./bin/wasm_exec.js ./bin/cdd-go.wasm --help
```

Or you can use the same `wasm_exec.js` to initialize and run the binary natively inside a modern Web Browser.

Since Go compiles directly to WASM without Emscripten, `cdd-go` does not require the `emsdk` toolchain for WebAssembly targets.
