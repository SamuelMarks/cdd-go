cdd-go
======

[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml)
![Test Coverage](https://img.shields.io/badge/Test%20Coverage-100.0%25-brightgreen.svg)
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
go install github.com/SamuelMarks/cdd-go/cmd/cdd-go@latest
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

The `cdd-go` project provides a robust, granular SDK for programmatically inspecting, modifying, and generating both OpenAPI specifications and Go AST (Abstract Syntax Tree) nodes.

#### Parsing and Emitting OpenAPI JSON

At the core is the `openapi` package, which lets you seamlessly ingest and output OpenAPI documents.

```go
package main

import (
	"os"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func main() {
	// Parse OpenAPI JSON to a structured *openapi.OpenAPI object
	f, _ := os.Open("openapi.json")
	defer f.Close()
	oa, err := openapi.Parse(f)
	if err != nil {
		panic(err)
	}

	// Manipulate the OpenAPI object
	oa.Info.Version = "2.0.0"

	// Emit the OpenAPI object back to JSON
	out, _ := os.Create("openapi_v2.json")
	defer out.Close()
	openapi.Emit(out, oa)
}
```

#### Working with Schemas (Models)

The `schemas` package handles bidirectional translation between Go `dst.GenDecl` (struct declarations) and OpenAPI `Schema` objects.

```go
package main

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/schemas"
)

func main() {
	// 1. Emit a Go struct from an OpenAPI Schema
	schema := &openapi.Schema{
		Type: "object",
		Description: "User profile information",
		Properties: map[string]openapi.Schema{
			"id":   {Type: "integer"},
			"name": {Type: "string", Description: "The user's full name"},
		},
	}
	
	decl := schemas.Emit("User", schema)
	
	// Print the generated Go struct
	fset := token.NewFileSet()
	decorator.Print(fset, decl)

	// 2. Parse a Go struct AST back into an OpenAPI Schema
	name, parsedSchema := schemas.Parse(decl)
	fmt.Printf("Parsed Schema '%s' with %d properties\n", name, len(parsedSchema.Properties))
}
```

#### Working with Routes (Interfaces)

The `routes` package maps OpenAPI `PathItem` objects representing endpoints into Go `dst.InterfaceType` nodes. This lets you generate standard Go interfaces for your server handlers, complete with docstrings.

```go
package main

import (
	"go/token"

	"github.com/dave/dst/decorator"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/routes"
)

func main() {
	pathItem := &openapi.PathItem{
		Summary: "User Management",
		Get: &openapi.Operation{
			OperationID: "GetUser",
			Summary:     "Retrieves a user by ID",
		},
		Post: &openapi.Operation{
			OperationID: "CreateUser",
			Summary:     "Creates a new user",
		},
	}

	// Emit an interface declaration for the route handlers
	decl, err := routes.EmitHandlerInterface("/users", pathItem)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	decorator.Print(fset, decl)
}
```

#### Working with Components

If you need to generate all the reusable components from an OpenAPI document (like security schemes, parameters, headers, and request bodies), the `components` package can emit an array of `dst.Decl`.

```go
package main

import (
	"go/token"

	"github.com/dave/dst/decorator"
	"github.com/SamuelMarks/cdd-go/src/components"
	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func main() {
	comp := &openapi.Components{
		Headers: map[string]openapi.Header{
			"Rate-Limit": {
				Description: "The number of allowed requests in the current period",
				Schema: &openapi.Schema{Type: "integer"},
			},
		},
	}

	// Returns an array of AST declarations for all components
	decls := components.Emit(comp)

	fset := token.NewFileSet()
	for _, decl := range decls {
		decorator.Print(fset, decl)
	}
}
```

### Granular Package Overview

The library is broken out into modules aligning to OpenAPI specifications and Go constructs:
- **`openapi`**: Top-level parsing/emitting and type definitions.
- **`schemas`**: OpenAPI `Schema` ↔ Go structs/types (`dst.GenDecl`).
- **`routes`**: OpenAPI `PathItem`/`Operation` ↔ Go interfaces (`dst.TypeSpec`).
- **`clients`**: Emits Go interfaces (`dst.TypeSpec`) for client SDK usage.
- **`components`**: OpenAPI `Components` ↔ Multiple Go declarations.
- **`docstrings`**: Extracts and emits Go comments aligned to OpenAPI descriptions/summaries.
- *(And many more for parameters, headers, servers, mocks, etc.)*

This enables deep integration for developers looking to incorporate OpenAPI-driven capabilities into their own Go-based code generators or toolchains.

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

## CLI Help

```
$ ./bin/cdd-go --help
cdd-go is a Code-Driven Development tool for Go.

Usage:
  cdd-go [subcommand] [flags]

Subcommands:
  from_openapi     Generate code from OpenAPI spec
  to_openapi       Generate OpenAPI spec from code
  to_docs_json     Generate documentation JSON from OpenAPI spec
  server_json_rpc  Run a JSON-RPC server exposing the CLI

Flags:
  -h, --help       Show this help message
  -v, --version    Show version information
```

### `from_openapi`

```
$ ./bin/cdd-go from_openapi --help
error: input file or directory is required
```

### `to_openapi`

```
$ ./bin/cdd-go to_openapi --help
Usage of to_openapi:
  -i string
    	Input file or directory path
  -o string
    	Output file path (default "openapi.json")
error: flag: help requested
```

### `to_docs_json`

```
$ ./bin/cdd-go to_docs_json --help
Usage of to_docs_json:
  -i string
    	Input file path
  -input string
    	Input file path
  -no-imports
    	Omit imports
  -no-wrapping
    	Omit wrapping
  -o string
    	Output file path
error: flag: help requested
```
