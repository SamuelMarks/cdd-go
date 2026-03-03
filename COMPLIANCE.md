# OpenAPI 3.2.0 Compliance

`cdd-go` currently partially implements OpenAPI 3.2.0.

Currently Supported:
- Schemas (`string`, `object`, etc)
- Operations (`GET`, `POST`, `PUT`, `DELETE`)
- operationId
- Path definitions
- Components -> Schemas

To Be Implemented:
- `patternProperties`
- `$dynamicRef` and other advanced JSON schema 2020-12 references.
- `webhooks`
- Request Bodies
- Responses and HTTP status codes
- Full Security Schemes
- Parameter serialization rules

Full compliance validation is a work in progress.
