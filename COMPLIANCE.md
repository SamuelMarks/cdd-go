# Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and MCP 1.0.0 Compliance

`cdd-go` achieves 100% compliance with all major Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and Model Context Protocol (MCP 1.0.0) concepts.

Currently Supported:
- Schemas (`string`, `integer`, `number`, `boolean`, `array`, `object`, `null`)
- Operations (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`, `HEAD`, `TRACE`)
- `operationId`
- Path definitions
- Components -> Schemas, Parameters, Responses, RequestBodies, Headers, Examples, Links, Callbacks, SecuritySchemes, PathItems, MediaTypes
- `patternProperties`
- `$dynamicRef` and other advanced JSON Schema 2020-12 references
- `webhooks`
- Request Bodies
- Responses and HTTP status codes
- Full Security Schemes (`apiKey`, `http`, `mutualTLS`, `oauth2`, `openIdConnect`)
- Parameter serialization rules
- Reference Object (`$ref`)
- OAuth Flows (`implicit`, `password`, `clientCredentials`, `authorizationCode`, `deviceAuthorization`)
- Security Requirement Object
- Links / Callbacks
- Info Object & Tags / ExternalDocs
- Server Object & Server Variables
- Parameter Object
- Arazzo 1.1.0 Workflows, Steps, Runtime Expressions (`$url`, `$method`, `$statusCode`, `$request.*`, `$response.*`, `$message.*`, `$inputs`, `$outputs`, `$steps`, `$workflows`, `$sourceDescriptions`, `$components`, `$self`), Selectors (`simple`, `regex`, `jsonpath`, `xpath`), Success/Failure Actions (`end`, `goto`, `retry`), and Payload Replacements
- Model Context Protocol (MCP 1.0.0) stdio server, SSE/HTTP endpoints, JSON-RPC 2.0 transport, Tools, Resources, Prompts, Roots, Sampling, Logging, and AST Resource Querying

Full compliance achieved. Detailed feature-by-feature conformance tables are maintained in [`conformance/`](conformance/):
- **Swagger 2.0**: [`client-sdk.md`](conformance/swagger-2.0/client-sdk.md), [`client-sdk-cli.md`](conformance/swagger-2.0/client-sdk-cli.md), [`servers.md`](conformance/swagger-2.0/servers.md)
- **OpenAPI 3.2.0**: [`client-sdk.md`](conformance/openapi-3.2.0/client-sdk.md), [`client-sdk-cli.md`](conformance/openapi-3.2.0/client-sdk-cli.md), [`servers.md`](conformance/openapi-3.2.0/servers.md)
- **Arazzo 1.1.0**: [`client-sdk.md`](conformance/arazzo-1.1.0/client-sdk.md), [`client-sdk-cli.md`](conformance/arazzo-1.1.0/client-sdk-cli.md), [`servers.md`](conformance/arazzo-1.1.0/servers.md)
- **MCP 1.0.0**: [`client-sdk.md`](conformance/mcp-1.0.0/client-sdk.md), [`client-sdk-cli.md`](conformance/mcp-1.0.0/client-sdk-cli.md), [`servers.md`](conformance/mcp-1.0.0/servers.md)

## Support Matrix

| Specification Concept       | Status |
| --------------------------- | ------ |
| Info Object                 | ✅     |
| Server Object               | ✅     |
| Components Object           | ✅     |
| Paths Object                | ✅     |
| Path Item Object            | ✅     |
| Operation Object            | ✅     |
| Parameter Object            | ✅     |
| Request Body Object         | ✅     |
| Responses Object            | ✅     |
| Responses / Schema Object   | ✅     |
| Reference Object (`$ref`)   | ✅     |
| OAuth Flows                 | ✅     |
| Security Requirement Object | ✅     |
| Links / Callbacks           | ✅     |
| Webhooks                    | ✅     |
| Arazzo 1.1.0 Workflows      | ✅     |
| Arazzo 1.1.0 Steps & DAGs   | ✅     |
| Arazzo 1.1.0 Expressions    | ✅     |
| MCP 1.0.0 Server & Adapters | ✅     |
