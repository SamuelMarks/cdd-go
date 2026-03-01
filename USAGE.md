# Usage

## CLI Reference

`cdd-go` serves as the core utility for Code-Driven Development in Go environments.

### Help & Version

```bash
cdd-go --help
cdd-go --version
```

### Parsing Go to OpenAPI

```bash
cdd-go to_openapi -f ./src -o openapi.json
```

### Emitting OpenAPI to Go SDKs & Server Stubs

```bash
# Emit a Server Framework (Routes & Handlers)
cdd-go from_openapi to_server -i openapi.json -o ./generated/

# Emit a Client SDK
cdd-go from_openapi to_sdk -i openapi.json -o ./generated/

# Emit a Client SDK along with a CLI
cdd-go from_openapi to_sdk_cli -i openapi.json -o ./generated/
```

### Generate documentation JSON

```bash
cdd-go to_docs_json --no-imports --no-wrapping -i openapi.json -o docs.json
```

### Server JSON RPC

```bash
cdd-go serve_json_rpc --port 8082 --listen 0.0.0.0
```
