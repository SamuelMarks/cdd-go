# cdd-go

![Test Coverage](https://img.shields.io/badge/Test_Coverage-91.2%25-green) ![Doc Coverage](https://img.shields.io/badge/Doc_Coverage-100.0%25-green)

`cdd-go` is a modular, bidirectional translation tool designed to convert standard OpenAPI 3.2.0 JSON specifications into highly idiomatic, executable Go source code—and vice versa. 

Designed around AST manipulation using [`github.com/SamuelMarks/dst`](https://github.com/SamuelMarks/dst), it is comment and whitespace-sensitive. This means you can generate a Go server client from an OpenAPI spec, allow your engineers to add custom code and comments inside the generated Go structures, and safely regenerate an `openapi.json` spec *back* from those modified files without destroying logic.

## Key Features

- **Bidirectional CLI (`from_openapi` | `to_openapi`)**
- **100% Native Spec Integrity:** Relies exclusively on `encoding/json` and `net/http` to match industry standardization strictly. No YAML parsing ambiguities.
- **AST Driven:** Emits and parses code dynamically preserving docstrings (`//`).
- **Modularity:** Maps specific OpenAPI components to isolated internal `cdd-go` libraries (`classes`, `docstrings`, `functions`, `mocks`, `routes`, `tests`).

## Getting Started

Check out [USAGE.md](./USAGE.md) for quick-start CLI examples.
To read more about the tool's specific OpenAPI compliance mapping, see [COMPLIANCE.md](./COMPLIANCE.md).

## Contributing & Publishing
- Looking to extend the library? See [DEVELOPING.md](./DEVELOPING.md) for architecture and test-coverage expectations.
- Ready to host the output? Read [PUBLISH.md](./PUBLISH.md) and [PUBLISH_OUTPUT.md](./PUBLISH_OUTPUT.md) for guide on automating CI/CD releases to `proxy.golang.org`.
