# cdd-go

[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml)
[![Test Coverage](https://img.shields.io/badge/Test_Coverage-91.2%25-green)](https://github.com/samuel/cdd-go)
[![Doc Coverage](https://img.shields.io/badge/Doc_Coverage-100.0%25-green)](https://github.com/samuel/cdd-go)

OpenAPI ↔ Go. Welcome to **cdd-go**, a code-generation and compilation tool bridging the gap between OpenAPI specifications and native `Go` source code. 

This toolset allows you to fluidly convert between your language's native constructs (like classes, structs, functions, routing, clients, and ORM models) and OpenAPI specifications, ensuring a single source of truth without sacrificing developer ergonomics.

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
cdd_go to_docs_json -i openapi.json
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
