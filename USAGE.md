# Usage

Use the `cdd-go` binary to generate Code from OpenAPI, or OpenAPI from Code.

```bash
# Display help
cdd-go --help

# Generate a Client SDK from an OpenAPI spec
cdd-go from_openapi to_sdk -i spec.json -o ./client

# Generate a CLI Tool from an OpenAPI spec
cdd-go from_openapi to_sdk_cli -i spec.json -o ./cli

# Generate a Gin-Gonic Web Server from an OpenAPI spec
cdd-go from_openapi to_server -i spec.json -o ./server

# Generate OpenAPI Spec from a Go Package
cdd-go to_openapi -f ./my-package -o ./openapi.json
```
