cdd-go
======

[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI/CD](https://github.com/offscale/cdd-go/workflows/CI/badge.svg)](https://github.com/offscale/cdd-go/actions)
![Test Coverage](https://img.shields.io/badge/Test%20Coverage-99.6%25-brightgreen.svg)
![Doc Coverage](https://img.shields.io/badge/Doc%20Coverage-100.0%25-brightgreen.svg)

OpenAPI ↔ Go. This is one compiler in a suite, all focussed on the same task: Compiler Driven Development (CDD).

Each compiler is written in its target language, is whitespace and comment sensitive, and has both an SDK and CLI.

The CLI—at a minimum—has:
- `cdd-go --help`
- `cdd-go --version`
- `cdd-go from_openapi -i spec.json`
- `cdd-go to_openapi -f path/to/code`
- `cdd-go to_docs_json --no-imports --no-wrapping -i spec.json`

The goal of this project is to enable rapid application development without tradeoffs. Tradeoffs of Protocol Buffers / Thrift etc. are an untouchable "generated" directory and package, compile-time and/or runtime overhead. Tradeoffs of Java or JavaScript for everything are: overhead in hardware access, offline mode, ML inefficiency, and more. And neither of these alterantive approaches are truly integrated into your target system, test frameworks, and bigger abstractions you build in your app. Tradeoffs in CDD are code duplication (but CDD handles the synchronisation for you).

## 🚀 Capabilities

The `cdd-go` compiler leverages a unified architecture to support various facets of API and code lifecycle management.

* **Compilation**:
  * **OpenAPI → `Go`**: Generate idiomatic native models, network routes, client SDKs, database schemas, and boilerplate directly from OpenAPI (`.json` / `.yaml`) specifications.
  * **`Go` → OpenAPI**: Statically parse existing `Go` source code and emit compliant OpenAPI specifications.
* **AST-Driven & Safe**: Employs static analysis (Abstract Syntax Trees) instead of unsafe dynamic execution or reflection, allowing it to safely parse and emit code even for incomplete or un-compilable project states.
* **Seamless Sync**: Keep your docs, tests, database, clients, and routing in perfect harmony. Update your code, and generate the docs; or update the docs, and generate the code.

## 📦 Installation

Requires Go 1.25+.

```bash
go install github.com/samuel/cdd-go/cmd/cdd-go@latest
```

## 🛠 Usage

### Command Line Interface

```bash
# Generate Go SDK from OpenAPI
cdd-go from_openapi to_sdk -i openapi.json -o ./sdk

# Generate OpenAPI from Go source code
cdd-go to_openapi -f ./src -o openapi.json
```

### Programmatic SDK / Library

```go
package main

import (
	"os"
	"github.com/samuel/cdd-go/src/openapi"
)

func main() {
	f, _ := os.Open("openapi.json")
	defer f.Close()
	oa, _ := openapi.Parse(f)
	// Work with the AST representations
}
```

## Design choices

We chose `github.com/dave/dst` to manipulate Go source code because it provides a reliable, high-level abstraction over the standard `go/ast` tree while fully preserving comments and whitespace, allowing surgical code generation that feels hand-written. We rely on standard Go tools, `gin-gonic/gin` for generated REST routes, and `cobra` for the generated SDK CLIs to ensure familiar, idiomatic output.

## 🏗 Supported Conversions for Go

*(The boxes below reflect the features supported by this specific `cdd-go` implementation)*

| Concept | Parse (From) | Emit (To) |
|---------|--------------|-----------|
| WebAssembly (WASM) | ❌ | ✅ |
| OpenAPI (JSON/YAML) | ✅ | ✅ |
| `Go` Models / Structs / Types | ✅ | ✅ |
| `Go` Server Routes / Endpoints | ✅ | ✅ |
| `Go` API Clients / SDKs | ✅ | ✅ |
| `Go` ORM / DB Schemas | ✅ | ✅ |
| `Go` CLI Argument Parsers | ✅ | ✅ |
| `Go` Docstrings / Comments | ✅ | ✅ |

---

## License

Licensed under either of

- Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE) or <https://www.apache.org/licenses/LICENSE-2.0>)
- MIT license ([LICENSE-MIT](LICENSE-MIT) or <https://opensource.org/licenses/MIT>)

at your option.

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in the work by you, as defined in the Apache-2.0 license, shall be
dual licensed as above, without any additional terms or conditions.
