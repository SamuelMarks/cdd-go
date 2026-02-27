# Project Architecture

`cdd-go` is designed as a bidirectional parser/emitter facilitating the translation between an OpenAPI 3.2.0 JSON specification and idiomatic Go source code. The project is separated into a CLI entrypoint and modular library packages mapped by functional logic rather than standard Go layouts.

## Directory Structure

The repository is organized to map OpenAPI semantics directly to granular Go concepts:

*   **`cmd/cdd_go/`**: The Command-Line Interface (CLI). Orchestrates the pipeline for translating files and managing user inputs (`--direction`, `--in`, `--out`).
*   **`src/`**: The core library handling structural translations.
    *   **`openapi/`**: Baseline structs directly mapping the OpenAPI 3.2.0 Specification. Provides raw unmarshalling/marshalling capabilities via standard `encoding/json`.
    *   **`classes/`**: Translates `openapi.Schema` objects into Go `*dst.StructType` nodes (and vice-versa). Handles Go-specific struct tags (`json:"name"`).
    *   **`routes/`**: Maps `openapi.PathItem` endpoints to Go interface definitions (`*dst.InterfaceType`) using standard `net/http` parameter signatures (`w http.ResponseWriter, r *http.Request`).
    *   **`functions/`**: Manages isolated `openapi.Operation` mapping to raw `*dst.FuncDecl` elements. Useful for implementing deeper logic layers beneath routes.
    *   **`mocks/`**: Translates OpenAPI `Examples` into Go `*dst.ValueSpec` variable declarations (string-wrapped JSON variables).
    *   **`docstrings/`**: Utility layer translating OpenAPI `Summary` and `Description` text strictly into `dst.Decorations` (whitespace-sensitive `//` Go comments).
    *   **`tests/`**: Scaffolds test structures (`*dst.FuncDecl`) mapping `openapi.Operation` targets to Go standard `*testing.T` signatures.

## Parsing Strategy (AST over Regex)

To ensure high-fidelity round-tripping (`openapi-to-language` -> edit -> `language-to-openapi`) without losing developer-authored formatting or comments:
1.  We rely strictly on standard `encoding/json` (skipping YAML) to keep specs deterministic.
2.  We use `github.com/SamuelMarks/dst` (Decorated Syntax Tree) instead of the standard `go/ast`. This fork/package retains whitespace, newlines, and comments explicitly mapped to nodes, allowing us to safely mutate specific structures without regenerating the entire file and losing custom developer logic.
