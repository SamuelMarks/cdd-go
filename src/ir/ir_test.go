// Package ir provides the universal Intermediate Representation (IR) that decouples
// Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and Go AST emitters and parsers.
package ir

import (
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestFromOpenAPIAndToOpenAPI(t *testing.T) {
	if _, err := FromOpenAPI(nil); err == nil {
		t.Errorf("expected error for nil OpenAPI")
	}
	if _, err := ToOpenAPI(nil); err == nil {
		t.Errorf("expected error for nil IR Document")
	}
	if _, err := NormalizeOpenAPI(nil); err == nil {
		t.Errorf("expected error for nil NormalizeOpenAPI input")
	}

	swaggerDoc := &openapi.OpenAPI{
		Swagger:  "2.0",
		Host:     "petstore.swagger.io",
		BasePath: "v2",
		Info: openapi.Info{
			Title:   "Swagger Petstore",
			Version: "1.0.0",
		},
		Definitions: map[string]openapi.Schema{
			"Pet": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"category": {Ref: "#/definitions/Category"},
				},
				Items:                &openapi.Schema{Ref: "#/definitions/Tag"},
				AdditionalProperties: &openapi.Schema{Ref: "#/definitions/Meta"},
			},
		},
		SecurityDefinitions: map[string]openapi.SecurityScheme{
			"basicAuth": {Type: "basic"},
			"apiKey":    {Type: "apiKey", In: "header", Name: "X-API-Key"},
		},
		Parameters: map[string]openapi.Parameter{
			"limitParam": {Name: "limit", In: "query", Type: "integer"},
		},
		Responses: map[string]openapi.Response{
			"NotFound": {Description: "Entity not found"},
		},
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"Existing": {Type: "string"},
			},
			SecuritySchemes: map[string]openapi.SecurityScheme{
				"bearerAuth": {Type: "http", Scheme: "bearer"},
			},
		},
		Paths: openapi.Paths{
			"/pets": {
				Get: &openapi.Operation{
					OperationID: "listPets",
					Parameters: []openapi.Parameter{
						{Name: "limit", In: "query", Type: "integer", Format: "int32"},
						{Name: "tag", In: "query", Schema: &openapi.Schema{Ref: "#/definitions/Tag"}},
					},
				},
				Post: &openapi.Operation{
					OperationID: "createPet",
					Parameters: []openapi.Parameter{
						{
							Name:        "body",
							In:          "body",
							Description: "Pet to add",
							Required:    true,
							Schema:      &openapi.Schema{Ref: "#/definitions/Pet"},
						},
					},
				},
				Put: &openapi.Operation{
					OperationID: "updatePetRaw",
					Parameters: []openapi.Parameter{
						{Name: "body", In: "body"},
					},
				},
				Patch: &openapi.Operation{
					OperationID: "uploadForm",
					Consumes:    []string{"multipart/form-data"},
					Parameters: []openapi.Parameter{
						{Name: "note", In: "formData", Description: "Note text"},
						{Name: "meta", In: "formData", Schema: &openapi.Schema{Ref: "#/definitions/Meta"}},
					},
				},
				Delete: &openapi.Operation{
					OperationID: "deletePet",
				},
			},
		},
		Webhooks: map[string]openapi.PathItem{
			"newPet": {
				Post: &openapi.Operation{
					OperationID: "onNewPet",
				},
			},
		},
	}

	normalized, err := NormalizeOpenAPI(swaggerDoc)
	if err != nil {
		t.Fatalf("unexpected error normalizing Swagger doc: %v", err)
	}
	if normalized.OpenAPI != "3.2.0" {
		t.Errorf("expected OpenAPI 3.2.0, got %s", normalized.OpenAPI)
	}
	if len(normalized.Servers) != 1 || normalized.Servers[0].URL != "https://petstore.swagger.io/v2" {
		t.Errorf("expected normalized server URL https://petstore.swagger.io/v2, got %+v", normalized.Servers)
	}
	if normalized.Components == nil {
		t.Fatalf("expected non-nil Components")
	}
	petSchema := normalized.Components.Schemas["Pet"]
	if petSchema.Properties["category"].Ref != "#/components/schemas/Category" {
		t.Errorf("expected rewritten property $ref, got %s", petSchema.Properties["category"].Ref)
	}
	if petSchema.Items == nil || petSchema.Items.Ref != "#/components/schemas/Tag" {
		t.Errorf("expected rewritten items $ref, got %+v", petSchema.Items)
	}
	if petSchema.AdditionalProperties == nil || petSchema.AdditionalProperties.Ref != "#/components/schemas/Meta" {
		t.Errorf("expected rewritten additionalProperties $ref, got %+v", petSchema.AdditionalProperties)
	}
	if normalized.Components.SecuritySchemes["basicAuth"].Type != "http" || normalized.Components.SecuritySchemes["basicAuth"].Scheme != "basic" {
		t.Errorf("expected basicAuth normalized to http/basic, got %+v", normalized.Components.SecuritySchemes["basicAuth"])
	}
	if _, ok := normalized.Components.Parameters["limitParam"]; !ok {
		t.Errorf("expected top-level Swagger parameter migrated to components.parameters")
	}
	if _, ok := normalized.Components.Responses["NotFound"]; !ok {
		t.Errorf("expected top-level Swagger response migrated to components.responses")
	}

	petsItem := normalized.Paths["/pets"]
	if petsItem.Post == nil || petsItem.Post.RequestBody == nil {
		t.Fatalf("expected POST /pets in:body parameter converted to RequestBody")
	}
	if petsItem.Post.RequestBody.Content["application/json"].Schema.Ref != "#/components/schemas/Pet" {
		t.Errorf("expected Pet schema ref in RequestBody, got %+v", petsItem.Post.RequestBody)
	}
	if petsItem.Put == nil || petsItem.Put.RequestBody == nil {
		t.Fatalf("expected PUT /pets in:body without schema converted to object RequestBody")
	}
	if petsItem.Patch == nil || petsItem.Patch.RequestBody == nil {
		t.Fatalf("expected PATCH /pets formData parameters converted to RequestBody")
	}
	if _, ok := petsItem.Patch.RequestBody.Content["multipart/form-data"]; !ok {
		t.Errorf("expected multipart/form-data content in PATCH RequestBody, got %+v", petsItem.Patch.RequestBody.Content)
	}
	if len(petsItem.Get.Parameters) != 2 || petsItem.Get.Parameters[0].Schema == nil || petsItem.Get.Parameters[0].Schema.Type != "integer" {
		t.Errorf("expected GET query parameter type migrated to Schema, got %+v", petsItem.Get.Parameters)
	}

	// Test empty IR Document ToOpenAPI
	emptyOut, err := ToOpenAPI(&Document{})
	if err != nil || emptyOut.OpenAPI != "3.2.0" || emptyOut.Components != nil {
		t.Errorf("expected clean empty OpenAPI from empty Document, got %+v, %v", emptyOut, err)
	}

	// Test ToOpenAPI with Models and SecuritySchemes when Components maps are nil
	docWithMaps := &Document{
		Models:          map[string]openapi.Schema{"Item": {Type: "string"}},
		SecuritySchemes: map[string]openapi.SecurityScheme{"api": {Type: "apiKey"}},
	}
	outMaps, err := ToOpenAPI(docWithMaps)
	if err != nil || outMaps.Components == nil || len(outMaps.Components.Schemas) != 1 || len(outMaps.Components.SecuritySchemes) != 1 {
		t.Errorf("expected populated Components from Document maps, got %+v, %v", outMaps, err)
	}
}
