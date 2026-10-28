# WASM Support

The `cdd-go` compiler fully supports compiling to WebAssembly out of the box using Go's native toolchain (`GOOS=wasip1 GOARCH=wasm` for WASI runtimes and `GOOS=js GOARCH=wasm` for browsers and Node.js).

## Building

Run the Makefile target (builds `bin/cdd-go.wasm` with `GOOS=wasip1 GOARCH=wasm`, `-trimpath`, `-ldflags="-s -w"`, `GOEXPERIMENT=nojsonv2`, and `wasm-opt -Oz` if installed):
```bash
make build_wasm
```

Or build directly for JavaScript / Browser environments (`GOOS=js GOARCH=wasm`):
```bash
GOEXPERIMENT=nojsonv2 GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o bin/cdd-go.wasm ./cmd/cdd-go
# or via the helper script (automatically runs wasm-opt -Oz if available):
go run ./scripts/build_wasm/build_wasm.go
```

## Binary Size Optimization

`cdd-go` applies several techniques to reduce the uncompressed WebAssembly binary size by ~50% (from ~18 MB down to ~9.1 MB uncompressed, and ~2.3 MB gzipped / ~1.7 MB brotli):

1. **HTTP / TLS Build-Tag Decoupling (`!js && !wasip1 && !wasm`):** Excludes `net/http`, `crypto/tls`, `crypto/x509`, and TCP server listener code (`serve_json_rpc` and live HTTP webhook dispatching) from WASM targets while retaining full OpenAPI/Arazzo AST parsing, code generation, and stdio MCP support (~7.0 MB saved).
2. **`GOEXPERIMENT=nojsonv2`:** Prevents Go 1.25+ from linking both the legacy `encoding/json` v1 and experimental `encoding/json/v2` + `jsontext` engines simultaneously (~1.0 MB saved).
3. **Linker & Path Stripping (`-trimpath -ldflags="-s -w"`):** Strips DWARF debugging sections, symbol tables, and host filesystem paths (~0.35 MB saved).
4. **Binaryen (`wasm-opt -Oz`):** Optimizes WASM instructions with `--enable-bulk-memory-opt --enable-nontrapping-float-to-int --enable-sign-ext --enable-mutable-globals --strip-debug --strip-dwarf --strip-producers` (~0.75 MB saved).

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
