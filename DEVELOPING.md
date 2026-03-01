# Developing `cdd-go`

Welcome to `cdd-go`! This guide covers the workflows and requirements needed to extend or maintain this package safely.

## Environment Setup
3. **Pre-commit Hooks**: We use `pre-commit` to ensure code quality and coverage shields are up to date. Install it using `pip install pre-commit` (or your package manager) and run `pre-commit install`.

1. **Prerequisites**: Ensure you have Go 1.21 or higher installed.
2. **Download modules**:
   ```bash
   go mod download
   ```
   *(Note: The project heavily relies on a specific fork of DST: `github.com/SamuelMarks/dst@upgrade-ghactions-whitespace-new-syntax-support` for accurate AST/whitespace preservation)*.

## Principles

When adding code to this project, adhere strictly to the following requirements:
- **No YAML:** Keep parsers locked to `encoding/json`. We use standard library packages (`net/http`) wherever feasible.
- **AST Fidelity:** Do not use `go/ast` or text-based regex replacements. Use the `dst` package to manipulate trees. This ensures that when a user edits a generated `openapi.json` file back to `.go` code, we do not destroy their manual inline comments.
- **Modularity:** Maintain the split of files (`parse.go`, `emit.go`, `parse_test.go`, `emit_test.go`). When adding new mapping mechanisms, define whether you are targeting a docstring, a mock, a class, or a route and place it accordingly.

## Testing & Coverage

The project enforces **100% test coverage** on all components within the `src/` modules.

To run the test suite and verify coverage locally:
```bash
# Run tests with coverage output
go test -coverprofile=coverage.out ./...

# View coverage percentage per function
go tool cover -func=coverage.out
```

To generate a graphical view of which lines are covered:
```bash
go tool cover -html=coverage.out -o coverage.html
```
Open `coverage.html` in your browser. Ensure no newly added files drag the total underneath 100%.

## Modifying the OpenAPI Spec

If expanding support for OpenAPI 3.2.0 capabilities:
1. First, model the JSON definition inside `src/openapi/types.go`.
2. Add the corresponding evaluation switch blocks inside `src/classes/emit.go` (if creating a Go Type) or `src/classes/parse.go` (if parsing an existing Go struct).
3. Ensure corresponding AST tests represent standard structure permutations in `parse_test.go`.
