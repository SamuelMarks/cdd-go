# WASM Integration

The `cdd-go` compiler can be compiled to WebAssembly (WASM), allowing it to run natively within modern web browsers or WASM runtimes without requiring a full Go environment.

## Building WASM

To build the WASM binary, run:

```bash
make build_wasm
```

This will output `bin/cdd-go.wasm`.

## Usage in the Browser

You can execute the WASM binary within a JavaScript environment using the standard Go WASM executor script:

```html
<script src="wasm_exec.js"></script>
<script>
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch("cdd-go.wasm"), go.importObject).then((result) => {
    go.run(result.instance);
  });
</script>
```

Currently, this allows seamless CLI generation and API transformation completely within a user's browser for an intuitive playground experience.
