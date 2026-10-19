# WASM Support

The `cdd-go` compiler fully supports compiling to WebAssembly out of the box using Go's native toolchain (`GOOS=wasip1 GOARCH=wasm` for WASI runtimes and `GOOS=js GOARCH=wasm` for browsers and Node.js).

## Building

Run the Makefile target (builds `bin/cdd-go.wasm` with `GOOS=wasip1 GOARCH=wasm`):
```bash
make build_wasm
```

Or build directly for JavaScript / Browser environments (`GOOS=js GOARCH=wasm`):
```bash
GOOS=js GOARCH=wasm go build -o bin/cdd-go.wasm ./cmd/cdd-go
# or via the helper script:
go run ./scripts/build_wasm/build_wasm.go
```

## Execution

### WASI Runtime (`wasmtime` / `wasmer`)
```bash
wasmtime bin/cdd-go.wasm --help
```

### Node.js / Web Browser (`GOOS=js GOARCH=wasm`)
You can run the `GOOS=js GOARCH=wasm` WebAssembly binary in Node.js or a modern browser using Go's `wasm_exec.js`:
```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./bin/
node ./bin/wasm_exec.js ./bin/cdd-go.wasm --help
```

Since Go compiles directly to WASM and WASI without Emscripten, `cdd-go` does not require the `emsdk` toolchain for WebAssembly targets.
