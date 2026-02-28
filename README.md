# cdd-go

[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI/CD](https://github.com/SamuelMarks/cdd-go/workflows/CI/badge.svg)](https://github.com/SamuelMarks/cdd-go/actions)
[![Test Coverage](https://img.shields.io/badge/Test_Coverage-91.2%25-green)](https://github.com/samuel/cdd-go)
[![Doc Coverage](https://img.shields.io/badge/Doc_Coverage-100.0%25-green)](https://github.com/samuel/cdd-go)

OpenAPI ↔ Go. This is one compiler in a suite, all focussed on the same task: Compiler Driven Development (CDD).

Each compiler is written in its target language, is whitespace and comment sensitive, and has both an SDK and CLI.

The CLI—at a minimum—has:
- `cdd_go --help`
- `cdd_go --version`
- `cdd_go from_openapi -i spec.json`
- `cdd_go to_openapi -f path/to/code`
- `cdd_go to_docs_json --no-imports --no-wrapping -i spec.json`

The goal of this project is to enable rapid application development without tradeoffs. Tradeoffs of Protocol Buffers / Thrift etc. are an untouchable "generated" directory and package, compile-time and/or runtime overhead. Tradeoffs of Java or JavaScript for everything are: overhead in hardware access, offline mode, ML inefficiency, and more. And neither of these alterantive approaches are truly integrated into your target system, test frameworks, and bigger abstractions you build in your app. Tradeoffs in CDD are code duplication (but CDD handles the synchronisation for you).

## 🚀 Capabilities

The `cdd-go` compiler leverages a unified architecture to support various facets of API and code lifecycle management.

* **Compilation**:
  * **OpenAPI → `Go`**: Generate idiomatic native models, network routes, client SDKs, database schemas, and boilerplate directly from OpenAPI (`.json` / `.yaml`) specifications.
  * **`Go` → OpenAPI**: Statically parse existing `Go` source code and emit compliant OpenAPI specifications.
* **AST-Driven & Safe**: Employs static analysis (Abstract Syntax Trees) instead of unsafe dynamic execution or reflection, allowing it to safely parse and emit code even for incomplete or un-compilable project states.
* **Seamless Sync**: Keep your docs, tests, database, clients, and routing in perfect harmony. Update your code, and generate the docs; or update the docs, and generate the code.

## 📦 Installation

To install `cdd-go`, ensure you have Go installed (version 1.18 or higher is recommended) and run:

```bash
go install github.com/samuel/cdd-go/cmd/cdd_go@latest
```

Alternatively, you can clone the repository and build the binary yourself:

```bash
git clone https://github.com/samuel/cdd-go.git
cd cdd-go
go build -o cdd_go ./cmd/cdd_go/
```

## 🛠 Usage

### Command Line Interface

Generate Go code (structs and routes) from an OpenAPI specification:

```bash
cdd_go from_openapi -i openapi.json -o ./generated
```

Generate an OpenAPI specification from existing Go source code:

```bash
cdd_go to_openapi -i ./src -o openapi.json
```

Generate structured JSON documentation examples for your API:

```bash
cdd_go to_docs_json --no-imports --no-wrapping -i openapi.json
```

### Programmatic SDK / Library

You can also use the `cdd-go` packages programmatically within your own Go applications:

```go
package main

import (
        "fmt"
        "os"

        "github.com/samuel/cdd-go/src/openapi"
)

func main() {
        f, err := os.Open("openapi.json")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        // Parse the OpenAPI specification
        doc, err := openapi.Parse(f)
        if err != nil {
                panic(err)
        }

        fmt.Printf("Parsed API: %s\n", doc.Info.Title)
}
```

## Design choices

The `cdd-go` compiler uses the `github.com/dave/dst` (Decorated Syntax Tree) library rather than just the standard `go/ast`. This is a crucial design choice: it allows the compiler to be completely whitespace and comment sensitive. When `cdd-go` parses a file, modifies a struct based on an OpenAPI update, and writes it back, it preserves all of your original formatting, inline comments, and docstrings perfectly. This fulfills the CDD promise of avoiding "untouchable generated directories"—you can freely edit the generated code, and the compiler respects your additions.

## 🏗 Supported Conversions for Go

*(The boxes below reflect the features supported by this specific `cdd-go` implementation)*

| Concept | Parse (From) | Emit (To) |
|---------|--------------|-----------|
| OpenAPI (JSON/YAML) | [✅] | [✅] |
| `Go` Models / Structs / Types | [✅] | [✅] |
| `Go` Server Routes / Endpoints | [✅] | [✅] |
| `Go` API Clients / SDKs | [✅] | [✅] |
| `Go` ORM / DB Schemas | [ ] | [ ] |
| `Go` CLI Argument Parsers | [ ] | [ ] |
| `Go` Docstrings / Comments | [✅] | [✅] |

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
