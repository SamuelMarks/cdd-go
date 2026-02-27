# Using `cdd-go`

The `cdd-go` project provides a central binary capable of parsing OpenAPI formats or scanning existing Go packages to extract OpenAPI configurations.

## Installation

```bash
# Install the CLI binary globally
go install github.com/samuel/cdd-go/cmd/cdd_go@latest
```

## OpenAPI to Language (Go Code Generation)

Given an existing `openapi.json` file, you can instruct `cdd-go` to read the structural definitions and emit standard Go components (like `net/http` route handler interfaces and struct maps).

```bash
# Generate Go structures from a local openapi.json file
cdd_go from_openapi -i ./openapi.json -o ./generated
```

### What happens?
- Components defined under `components.schemas` are evaluated. File names are matched to snake-cased struct models and exported safely (e.g. `User` -> `user.go`).
- Endpoints under `paths` are aggregated and evaluated into RESTful interfaces implementing `(w http.ResponseWriter, r *http.Request)`.

## Language to OpenAPI (Reverse Generation)

If developers have written Go code directly or modified the generated output, you can safely extract an entirely new `openapi.json` spec. `cdd-go` crawls the target directory matching standard interfaces, structs, tags, and inline comments.

```bash
# Generate an openapi.json from an entire Go module directory
cdd_go to_openapi -i ./my_project_pkg -o ./docs/openapi.json
```

### What happens?
- Crawls the `my_project_pkg` root directory looking for Go types.
- Evaluates `type XYZ struct` declarations, picking up exact `json:"xyz"` tags to emit standard JSON components back to `components.schemas`.
- Scans Interface signatures, mapping `Get` / `Post` method names directly to REST `Paths`.
- Parses any docstrings mapping directly back to `summary` and `description` objects.
