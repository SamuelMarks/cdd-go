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
cdd-go to_openapi -i ./my-package -o ./openapi.json

# Generate a Workflow Client SDK from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_sdk -i arazzo.json -o ./workflow_sdk

# Generate a Workflow CLI Tool (reusing the SDK) from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_sdk_cli -i arazzo.json -o ./workflow_cli

# Generate a Workflow Server & E2E Test Harness from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_server -i arazzo.json -o ./workflow_server

# Extract an Arazzo 1.1.0 Spec from Go Workflow Code
cdd-go to_arazzo -i ./workflow_sdk -o ./arazzo.json
```
