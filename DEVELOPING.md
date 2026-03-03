# Developing `cdd-go`

## Setup

Ensure you have Go 1.25+ installed.

```bash
make install_deps
```

## Testing

Run tests and display coverage (requires 100% statement coverage for core packages):

```bash
make test
```

## Building

```bash
make build
make build_wasm
make build_docker
```

## Architecture Map

- `src/classes`: Parses and emits Go structs / models.
- `src/clients`: Parses and emits Go client SDK interfaces.
- `src/functions`: Parses and emits OpenAPI Operations.
- `src/routes`: Parses and emits Go `gin-gonic/gin` web server routes.
- `cmd/cdd-go`: The main CLI entrypoint.
