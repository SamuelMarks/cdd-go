# WebAssembly (WASM) Support

`cdd-go` natively supports compilation to WebAssembly (WASM). This allows you to run `cdd-go` entirely within a modern web browser or in WASM environments (like unified CLI interfaces across multiple languages).

## Building WASM

You can build the WASM binary using the provided Makefiles:

**POSIX (Linux, macOS, etc.):**
```bash
make build_wasm
```

**Windows:**
```cmd
make.bat build_wasm
```

This will produce a `cdd-go.wasm` file in the `bin/` directory.

## Integrating

To run this WASM file in a browser or a Node.js environment, you need Go's `wasm_exec.js` which bridges the Go runtime with JavaScript.

```bash
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./bin/
```

You can then load and execute it using a standard WebAssembly pipeline.
