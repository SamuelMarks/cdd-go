# Developing `cdd-go`

## Setup

Ensure you have Go 1.25+ installed.

```bash
make deps
# or:
make install_deps
```

## Testing & Coverage

Run the test suite and verify 100% statement coverage and 100% AST documentation coverage across all core packages:

```bash
make test
go run ./scripts/doc_cover/doc_cover.go
```

## Building

```bash
make build
make build_wasm
make build_docker
```

## Architecture Map

- `cmd/cdd-go`: The main CLI entrypoint, flag parser, JSON-RPC 2.0 server (`serve_json_rpc`), and documentation snippet CLI (`to_docs_json`).
- `cdd`: High-level programmatic SDK orchestrating OpenAPI (`FromOpenAPI`, `ToOpenAPI`), Arazzo (`FromArazzo`, `ToArazzo`), in-place `dave/dst` AST merging (`Sync`), documentation generation (`ToDocsJSON`), DAO/Database/Seeder/IdP/Middleware scaffolding, and MCP stdio server (`ServeMCPStdio`).
- `src/openapi`: Swagger 2.0 & OpenAPI 3.2.0 parser, upgrader, and JSON/YAML emitter.
- `src/arazzo`: Arazzo 1.1.0 parser, validator, cycle-checking DAG analyzer, runtime expression evaluator, and JSON/YAML emitter.
- `src/workflows`: Arazzo 1.1.0 workflow emitters and parsers for Client SDKs (`to_sdk`), Cobra CLIs (`to_sdk_cli`), and Gin Server routes & E2E tests (`to_server`).
- `src/classes`: Parses and emits Go structs / GORM models.
- `src/clients`: Parses and emits Go `net/http` client SDK interfaces and MCP client adapters.
- `src/commands`: Parses and emits `spf13/cobra` CLI commands and MCP CLI bindings.
- `src/routes`: Parses and emits `gin-gonic/gin` HTTP handler interfaces, structs, and MCP SSE/message endpoints.
- `src/daos`, `src/database`, `src/seeder`: Generates multi-tiered DAOs (`StubDAO` / `GormDAO` / `DAOFactory`), SQLite/PostgreSQL auto-migrations, `go-sqlmock` DAO tests, and `gofakeit` entity seeders.
- `src/mcp`: Model Context Protocol (MCP 1.0.0) and JSON-RPC 2.0 schemas, lifecycle state machine, and AST resource handlers.
- `src/tests`, `src/mocks`: Synthesizes deterministic contract integration tests (`testing.T`) and OpenAPI Example mock variables.
