// Package ir provides the universal Intermediate Representation (IR) that decouples
// Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and Go AST emitters and parsers.
package ir

import (
	"fmt"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// ToOpenAPI converts a canonical IR Document back into an OpenAPI 3.2.0 specification structure.
// Parameter doc is the source IR Document, and it returns the constructed OpenAPI 3.2.0 document and any error.
func ToOpenAPI(doc *Document) (*openapi.OpenAPI, error) {
	if doc == nil {
		return nil, fmt.Errorf("ir document is nil")
	}

	version := doc.Metadata.OpenAPIVersion
	if version == "" {
		version = "3.2.0"
	}

	comp := doc.Components
	if len(doc.Models) > 0 {
		if comp.Schemas == nil {
			comp.Schemas = make(map[string]openapi.Schema, len(doc.Models))
		}
		for k, v := range doc.Models {
			comp.Schemas[k] = v
		}
	}
	if len(doc.SecuritySchemes) > 0 {
		if comp.SecuritySchemes == nil {
			comp.SecuritySchemes = make(map[string]openapi.SecurityScheme, len(doc.SecuritySchemes))
		}
		for k, v := range doc.SecuritySchemes {
			comp.SecuritySchemes[k] = v
		}
	}

	var compPtr *openapi.Components
	if len(comp.Schemas) > 0 || len(comp.SecuritySchemes) > 0 || len(comp.Parameters) > 0 ||
		len(comp.Responses) > 0 || len(comp.Examples) > 0 || len(comp.RequestBodies) > 0 ||
		len(comp.Headers) > 0 || len(comp.Links) > 0 || len(comp.Callbacks) > 0 ||
		len(comp.PathItems) > 0 || len(comp.MediaTypes) > 0 {
		compPtr = &comp
	}

	var paths openapi.Paths
	if len(doc.Endpoints) > 0 {
		paths = make(openapi.Paths, len(doc.Endpoints))
		for k, v := range doc.Endpoints {
			paths[k] = v
		}
	}

	var webhooks map[string]openapi.PathItem
	if len(doc.Webhooks) > 0 {
		webhooks = make(map[string]openapi.PathItem, len(doc.Webhooks))
		for k, v := range doc.Webhooks {
			webhooks[k] = v
		}
	}

	return &openapi.OpenAPI{
		OpenAPI:           version,
		Self:              doc.Metadata.Self,
		Info:              doc.Metadata.Info,
		JSONSchemaDialect: doc.Metadata.JSONSchemaDialect,
		Servers:           doc.Metadata.Servers,
		Paths:             paths,
		Webhooks:          webhooks,
		Components:        compPtr,
		Security:          doc.Security,
		Tags:              doc.Metadata.Tags,
		ExternalDocs:      doc.Metadata.ExternalDocs,
	}, nil
}

// NormalizeOpenAPI normalizes a Swagger 2.0 or OpenAPI 3.x specification through the canonical IR layer.
// Parameter oa is the input OpenAPI document, and it returns the normalized OpenAPI 3.2.0 document and any error.
func NormalizeOpenAPI(oa *openapi.OpenAPI) (*openapi.OpenAPI, error) {
	irDoc, err := FromOpenAPI(oa)
	if err != nil {
		return nil, err
	}
	return ToOpenAPI(irDoc)
}
