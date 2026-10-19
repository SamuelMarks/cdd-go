# Usage

Use the `cdd-go` CLI binary or programmatic `github.com/SamuelMarks/cdd-go/cdd` package to compile OpenAPI 3.2.0, Swagger 2.0, and Arazzo 1.1.0 specifications to and from Go code.

## CLI Usage

```bash
# Display help and version
cdd-go --help
cdd-go --version

# Generate a Client SDK from an OpenAPI spec (with tests, mocks, and MCP adapters)
cdd-go from_openapi to_sdk -i spec.json -o ./client --tests --mcp

# Generate a Cobra CLI Tool from an OpenAPI spec
cdd-go from_openapi to_sdk_cli -i spec.json -o ./cli --tests --mcp

# Generate a Gin-Gonic Web Server (with GORM DAOs, SQLite/Postgres auto-migration, and Faker seeder)
cdd-go from_openapi to_server -i spec.json -o ./server --tests --mcp

# Batch generate from a directory of OpenAPI specifications
cdd-go from_openapi to_sdk -d ./specs -o ./client

# Extract an OpenAPI 3.2.0 Spec from a Go Package or File
cdd-go to_openapi -i ./my-package -o ./openapi.json

# Generate a Workflow Client SDK from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_sdk -i arazzo.json -o ./workflow_sdk --tests

# Generate a Workflow CLI Tool (reusing the SDK) from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_sdk_cli -i arazzo.json -o ./workflow_cli --tests

# Generate a Workflow Server & E2E Test Harness from an Arazzo 1.1.0 spec
cdd-go from_arazzo to_server -i arazzo.json -o ./workflow_server --tests

# Extract an Arazzo 1.1.0 Spec from Go Workflow Code (SDK, CLI, or Server)
cdd-go to_arazzo -i ./workflow_sdk -o ./arazzo.json

# Synchronize between Go source code and OpenAPI / Arazzo specifications in-place
cdd-go sync -i ./server -o ./openapi.json --truth code
cdd-go sync -i ./openapi.json -o ./server --truth openapi
cdd-go sync -i ./arazzo.json -o ./workflow_sdk --truth arazzo

# Generate documentation-optimized JSON with Go code snippets
cdd-go to_docs_json -i spec.json -o ./docs.json --no-imports --no-wrapping

# Expose the compiler as a JSON-RPC 2.0 HTTP server
cdd-go serve_json_rpc --listen 127.0.0.1 --port 8080

# Run the compiler as a Model Context Protocol (MCP) server over stdio
cdd-go mcp
```

## Running Generated Servers (`to_server`)

Servers generated via `cdd-go from_openapi to_server` include a runnable `cmd/server/main.go` entrypoint with built-in flags for DAO persistence modes, fake data seeding, and request validation:

```bash
# Run in stub DAO mode (returns 501 Not Implemented by default)
go run ./cmd/server -mode stub

# Run with an ephemeral in-memory SQLite database and auto-migrate GORM schemas
go run ./cmd/server -mode ephemeral

# Run with ephemeral SQLite, auto-migrate, and seed relational fake data via gofakeit
go run ./cmd/server -mode ephemeral -seed -validate

# Run against PostgreSQL using concrete GORM DAOs
DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
  go run ./cmd/server -mode concrete -seed -validate
```

## Programmatic Go SDK Usage

```go
package main

import (
	"log"

	"github.com/SamuelMarks/cdd-go/cdd"
)

func main() {
	// Generate a Gin server with tests and MCP adapters
	if err := cdd.FromOpenAPI("to_server", "spec.json", "./server", false, false, true, true); err != nil {
		log.Fatal(err)
	}

	// Extract OpenAPI 3.2.0 from Go source code
	if err := cdd.ToOpenAPI("./server", "openapi.json"); err != nil {
		log.Fatal(err)
	}

	// Generate and extract Arazzo 1.1.0 workflows
	if err := cdd.FromArazzo("to_sdk", "arazzo.json", "./workflow_sdk", false, false, true); err != nil {
		log.Fatal(err)
	}
	if err := cdd.ToArazzo("./workflow_sdk", "extracted.arazzo.json"); err != nil {
		log.Fatal(err)
	}

	// Bidirectional sync
	if err := cdd.Sync("./server", "openapi.json", "code"); err != nil {
		log.Fatal(err)
	}
}
```
