# cdd-go

[![License](https://img.shields.io/badge/license-Apache--2.0%20OR%20MIT-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CI](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml/badge.svg)](https://github.com/SamuelMarks/cdd-go/actions/workflows/ci.yml)
![Test Coverage](https://img.shields.io/badge/Test%20Coverage-100.0%25-brightgreen.svg)
![Doc Coverage](https://img.shields.io/badge/Doc%20Coverage-100.0%25-brightgreen.svg)

OpenAPI ↔ Go. This is one compiler in a suite, all focussed on the same task: Compiler Driven Development (CDD).

Each compiler is written in its target language, is whitespace and comment sensitive, and has both an SDK and CLI.

The CLI—at a minimum—has:

- `cdd-go --help`
- `cdd-go --version`
- `cdd-go from_openapi to_sdk_cli -i spec.json`
- `cdd-go from_openapi to_sdk -i spec.json`
- `cdd-go from_openapi to_server -i spec.json`
- `cdd-go to_openapi -f path/to/code`
- `cdd-go to_docs_json --no-imports --no-wrapping -i spec.json`
- `cdd-go serve_json_rpc --port 8080 --listen 0.0.0.0`

The goal of this project is to enable rapid application development without tradeoffs. Tradeoffs of Protocol Buffers / Thrift etc. are an untouchable "generated" directory and package, compile-time and/or runtime overhead. Tradeoffs of Java or JavaScript for everything are: overhead in hardware access, offline mode, ML inefficiency, and more. And neither of these alternative approaches are truly integrated into your target system, test frameworks, and bigger abstractions you build in your app. Tradeoffs in CDD are code duplication (but CDD handles the synchronisation for you).

## 🚀 Capabilities

The `cdd-go` compiler leverages a unified architecture to support various facets of API and code lifecycle management.

- **Compilation**:
    - **OpenAPI → `Go`**: Generate idiomatic native models, network routes, client SDKs, and boilerplate directly from OpenAPI (`.json` / `.yaml`) specifications.
    - **`Go` → OpenAPI**: Statically parse existing `Go` source code and emit compliant OpenAPI specifications.
- **AST-Driven & Safe**: Employs static analysis instead of unsafe dynamic execution or reflection, allowing it to safely parse and emit code even for incomplete or un-compilable project states.
- **Seamless Sync**: Keep your docs, tests, database, clients, and routing in perfect harmony. Update your code, and generate the docs; or update the docs, and generate the code.

## 📦 Installation & Build

### Native Tooling

```bash
go build ./...
go test ./...
```

### Makefile / make.bat

You can also use the included cross-platform Makefiles to fetch dependencies, build, and test:

```bash
# Install dependencies
make deps

# Build the project
make build

# Run tests
make test
```

## 🛠 Usage

### Command Line Interface

```bash
# Generate Go models from an OpenAPI spec
cdd-go from_openapi to_sdk -i spec.json -o src/models

# Generate an OpenAPI spec from your Go code
cdd-go to_openapi -f src/models -o openapi.json
```

### Programmatic SDK / Library

```go
package main

import (
	"fmt"
	"github.com/SamuelMarks/cdd-go/cdd"
)

func main() {
	config := cdd.Config{InputPath: "spec.json", OutputDir: "src/models"}
	cdd.GenerateSDK(config)
	fmt.Println("SDK generation complete.")
}
```

## 🏗 Supported Conversions for Go

*(The boxes below reflect the features supported by this specific `cdd-go` implementation)*

| Features | Parse (From) | Emit (To) |
| --- | --- | --- |
| OpenAPI 3.2.0 | ✅ | ✅ |
| API Client SDK | ✅ | ✅ |
| API Client CLI | ✅ | ✅ |
| Server Routes / Endpoints | ✅ | ✅ |
| ORM / DB Schema | ✅ | ✅ |
| Mocks + Tests | [ ] | [ ] |
| Model Context Protocol (MCP) | [ ] | [ ] |

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
