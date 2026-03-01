# OpenAPI Compliance Report

The `cdd-go` parser and schema definitions adhere to **OpenAPI Specification Version 3.2.0**.

## Supported Capabilities

### OpenAPI Data Structures (`src/openapi/types.go`)
- **Metadata Objects**: `Info`, `Contact`, `License`, `ExternalDocs`.
- **Server Objects**: Full mapping of `Server`, `ServerVariable`.
- **Pathing**: `Paths`, `PathItem`, `Operation` mapping.
- **Component Registries**: Models `Components` covering `Schemas`, `Responses`, `Parameters`, `Examples`, `RequestBodies`, `Headers`, `SecuritySchemes`, `Links`, `Callbacks`, `PathItems`, and `MediaTypes`.
- **Security**: Basic array mappings for `SecurityRequirement` and extensive OAuth flow definitions (`OAuthFlows`, `OAuthFlow`).

### Structural Translation Support (`openapi-to-language` / `language-to-openapi`)
- **Types**: Deep struct translation. Supports scalar types (`string`, `integer`, `number`, `boolean`) and lists/slices (`array`), and dynamic hash maps (`additionalProperties`).
- **References**: Evaluates `$ref` identifiers locally as struct dependencies.
- **REST Paths**: Detects and scaffolds HTTP method structs matching Go's HTTP verb expectations (supports GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, TRACE).
- **Comments**: Native support for OpenAPI `summary` and `description` string keys transforming bidirectionally into standard multi-line Go docstrings.
- **Examples**: Captures raw JSON arrays/objects natively into explicit raw string instances avoiding unmarshalling complexity.

## Current Limitations & Ongoing Work
- **Complex Objects:** Subschemas referencing nested arrays inside nested objects (`deepObject` resolution) requires recursive definition bridging.
- **Parameter Locations:** Advanced handling of HTTP parameters within `query`, `header`, or `cookie` constraints is currently modeled in definitions but may not emit complex routing boilerplate beyond standard `http.ResponseWriter`.
- **Response Validation:** Emitted tests mock empty signature endpoints; advanced runtime validators evaluating the schemas live against `net/http` traffic remain a future implementation.
