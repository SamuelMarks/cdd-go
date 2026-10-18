# Swagger 2.0, OpenAPI 3.2.0, and Arazzo 1.1.0 Compliance

`cdd-go` achieves 100% compliance with all major Swagger 2.0, OpenAPI 3.2.0, and Arazzo 1.1.0 concepts.

Currently Supported:
- Schemas (`string`, `object`, etc)
- Operations (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`, `OPTIONS`, `HEAD`, `TRACE`)
- operationId
- Path definitions
- Components -> Schemas
- `patternProperties`
- `$dynamicRef` and other advanced JSON schema 2020-12 references.
- `webhooks`
- Request Bodies
- Responses and HTTP status codes
- Full Security Schemes
- Parameter serialization rules
- Reference Object
- OAuth Flows
- Security Requirement Object
- Links / Callbacks
- Info Object
- Server Object
- Parameter Object
- Arazzo 1.1.0 Workflows, Steps, Runtime Expressions (`$url`, `$method`, `$statusCode`, `$request.*`, `$response.*`, `$message.*`, `$inputs`, `$outputs`, `$steps`, `$workflows`, `$sourceDescriptions`, `$components`, `$self`), Selectors, Success/Failure Actions, and Payload Replacements

Full compliance achieved.

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
