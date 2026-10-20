// Package ir provides the universal Intermediate Representation (IR) that decouples
// Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and Go AST emitters and parsers.
package ir

import (
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// FromOpenAPI normalizes a parsed Swagger 2.0 or OpenAPI 3.x specification into a canonical IR Document.
// Parameter oa is the source OpenAPI/Swagger structure, and it returns the normalized IR Document and any error.
func FromOpenAPI(oa *openapi.OpenAPI) (*Document, error) {
	if oa == nil {
		return nil, fmt.Errorf("openapi document is nil")
	}

	version := oa.OpenAPI
	if version == "" {
		version = "3.2.0"
	}

	servers := make([]openapi.Server, len(oa.Servers))
	copy(servers, oa.Servers)
	if len(servers) == 0 && oa.Host != "" {
		schemes := oa.Schemes
		if len(schemes) == 0 {
			schemes = []string{"https"}
		}
		basePath := oa.BasePath
		if basePath != "" && !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		for _, scheme := range schemes {
			servers = append(servers, openapi.Server{
				URL: fmt.Sprintf("%s://%s%s", scheme, oa.Host, basePath),
			})
		}
	}

	doc := &Document{
		Metadata: Metadata{
			OpenAPIVersion:    version,
			Info:              oa.Info,
			Servers:           servers,
			Tags:              oa.Tags,
			ExternalDocs:      oa.ExternalDocs,
			Self:              oa.Self,
			JSONSchemaDialect: oa.JSONSchemaDialect,
		},
		Models:          make(map[string]openapi.Schema),
		Endpoints:       make(map[string]openapi.PathItem),
		SecuritySchemes: make(map[string]openapi.SecurityScheme),
		Security:        oa.Security,
		Webhooks:        make(map[string]openapi.PathItem),
	}

	if oa.Components != nil {
		doc.Components = *oa.Components
		for k, v := range oa.Components.Schemas {
			doc.Models[k] = normalizeSchemaRefs(v)
		}
		for k, v := range oa.Components.SecuritySchemes {
			doc.SecuritySchemes[k] = v
		}
	}

	for k, v := range oa.Definitions {
		doc.Models[k] = normalizeSchemaRefs(v)
	}
	for k, v := range oa.SecurityDefinitions {
		doc.SecuritySchemes[k] = normalizeSwaggerSecurityScheme(v)
	}
	if len(oa.Parameters) > 0 {
		if doc.Components.Parameters == nil {
			doc.Components.Parameters = make(map[string]openapi.Parameter)
		}
		for k, v := range oa.Parameters {
			doc.Components.Parameters[k] = v
		}
	}
	if len(oa.Responses) > 0 {
		if doc.Components.Responses == nil {
			doc.Components.Responses = make(map[string]openapi.Response)
		}
		for k, v := range oa.Responses {
			doc.Components.Responses[k] = v
		}
	}

	doc.Components.Schemas = doc.Models
	doc.Components.SecuritySchemes = doc.SecuritySchemes

	defaultConsumes := oa.Consumes
	if len(defaultConsumes) == 0 {
		defaultConsumes = []string{"application/json"}
	}

	for path, item := range oa.Paths {
		doc.Endpoints[path] = normalizePathItem(item, defaultConsumes)
	}
	for name, wh := range oa.Webhooks {
		doc.Webhooks[name] = normalizePathItem(wh, defaultConsumes)
	}

	return doc, nil
}

// normalizeSwaggerSecurityScheme converts Swagger 2.0 security definitions (e.g. "basic") into OpenAPI 3.2.0 SecuritySchemes.
// Parameter s is the input SecurityScheme, and it returns the normalized SecurityScheme.
func normalizeSwaggerSecurityScheme(s openapi.SecurityScheme) openapi.SecurityScheme {
	if s.Type == "basic" {
		s.Type = "http"
		s.Scheme = "basic"
	}
	return s
}

// normalizeSchemaRefs rewrites legacy Swagger 2.0 `#/definitions/` references to `#/components/schemas/`.
// Parameter s is the Schema to normalize, and it returns the updated Schema.
func normalizeSchemaRefs(s openapi.Schema) openapi.Schema {
	if strings.HasPrefix(s.Ref, "#/definitions/") {
		s.Ref = "#/components/schemas/" + strings.TrimPrefix(s.Ref, "#/definitions/")
	}
	if s.Items != nil {
		normItems := normalizeSchemaRefs(*s.Items)
		s.Items = &normItems
	}
	if s.AdditionalProperties != nil {
		normAdd := normalizeSchemaRefs(*s.AdditionalProperties)
		s.AdditionalProperties = &normAdd
	}
	if len(s.Properties) > 0 {
		props := make(map[string]openapi.Schema, len(s.Properties))
		for k, v := range s.Properties {
			props[k] = normalizeSchemaRefs(v)
		}
		s.Properties = props
	}
	return s
}

// normalizePathItem normalizes all operations on a PathItem, including Swagger 2.0 body/formData parameters.
// Parameter item is the source PathItem, defaultConsumes is the fallback MIME list, and it returns the normalized PathItem.
func normalizePathItem(item openapi.PathItem, defaultConsumes []string) openapi.PathItem {
	for _, mo := range openapi.PathItemOperations(&item) {
		normOp := normalizeOperation(*mo.Operation, defaultConsumes)
		item.SetOperation(mo.Method, &normOp)
	}
	return item
}

// normalizeOperation converts Swagger 2.0 `in: "body"` and `in: "formData"` parameters into an OpenAPI 3.2.0 RequestBody.
// Parameter op is the Operation to normalize, defaultConsumes is the fallback MIME list, and it returns the normalized Operation.
func normalizeOperation(op openapi.Operation, defaultConsumes []string) openapi.Operation {
	if len(op.Parameters) == 0 {
		return op
	}
	consumes := op.Consumes
	if len(consumes) == 0 {
		consumes = defaultConsumes
	}

	var keptParams []openapi.Parameter
	var formProps map[string]openapi.Schema
	for _, p := range op.Parameters {
		switch p.In {
		case "body":
			if op.RequestBody == nil {
				var bodySchema *openapi.Schema
				if p.Schema != nil {
					norm := normalizeSchemaRefs(*p.Schema)
					bodySchema = &norm
				} else {
					bodySchema = &openapi.Schema{Type: "object"}
				}
				content := make(map[string]openapi.MediaType, len(consumes))
				for _, mime := range consumes {
					content[mime] = openapi.MediaType{Schema: bodySchema}
				}
				op.RequestBody = &openapi.RequestBody{
					Description: p.Description,
					Required:    p.Required,
					Content:     content,
				}
			}
		case "formData":
			if formProps == nil {
				formProps = make(map[string]openapi.Schema)
			}
			propType := p.Type
			if propType == "" {
				propType = "string"
			}
			if p.Schema != nil {
				formProps[p.Name] = normalizeSchemaRefs(*p.Schema)
			} else {
				formProps[p.Name] = openapi.Schema{Type: propType, Description: p.Description}
			}
		default:
			if p.Schema == nil && p.Type != "" {
				p.Schema = &openapi.Schema{Type: p.Type, Format: p.Format}
				p.Type = ""
				p.Format = ""
			} else if p.Schema != nil {
				norm := normalizeSchemaRefs(*p.Schema)
				p.Schema = &norm
			}
			keptParams = append(keptParams, p)
		}
	}

	if len(formProps) > 0 && op.RequestBody == nil {
		mime := "application/x-www-form-urlencoded"
		for _, c := range consumes {
			if strings.Contains(c, "multipart") || strings.Contains(c, "form") {
				mime = c
				break
			}
		}
		op.RequestBody = &openapi.RequestBody{
			Content: map[string]openapi.MediaType{
				mime: {
					Schema: &openapi.Schema{
						Type:       "object",
						Properties: formProps,
					},
				},
			},
		}
	}

	op.Parameters = keptParams
	return op
}
