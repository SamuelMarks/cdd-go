package classes

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// ParseType parses a dst.TypeSpec into an OpenAPI Schema.
func ParseType(ts *dst.TypeSpec) (*openapi.Schema, error) {
	if ts == nil {
		return nil, fmt.Errorf("TypeSpec is nil")
	}

	schema := &openapi.Schema{}

	if len(ts.Decs.Start) > 0 {
		desc := ""
		for _, doc := range ts.Decs.Start {
			desc += strings.TrimSpace(strings.TrimPrefix(doc, "//")) + " "
		}
		schema.Description = strings.TrimSpace(desc)
	}

	switch t := ts.Type.(type) {
	case *dst.StructType:
		schema.Type = "object"
		schema.Properties = make(map[string]openapi.Schema)
		for _, field := range t.Fields.List {
			if len(field.Names) == 0 {
				continue // Skip embedded fields for now
			}
			propName := field.Names[0].Name
			if field.Tag != nil && field.Tag.Kind == token.STRING {
				tagVal := strings.Trim(field.Tag.Value, "`")
				structTag := reflect.StructTag(tagVal)
				if j := structTag.Get("json"); j != "" {
					parts := strings.Split(j, ",")
					if parts[0] != "" {
						propName = parts[0]
					}
				}
			}

			propSchema, err := ParseExpr(field.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to parse field %s: %w", field.Names[0].Name, err)
			}

			if len(field.Decs.Start) > 0 {
				desc := ""
				for _, doc := range field.Decs.Start {
					desc += strings.TrimSpace(strings.TrimPrefix(doc, "//")) + " "
				}
				propSchema.Description = strings.TrimSpace(desc)
			}

			schema.Properties[propName] = *propSchema
		}
	case *dst.ArrayType:
		schema.Type = "array"
		itemsSchema, err := ParseExpr(t.Elt)
		if err != nil {
			return nil, err
		}
		schema.Items = itemsSchema
	default:
		s, err := ParseExpr(t)
		if err != nil {
			return nil, err
		}
		schema.Type = s.Type
		schema.Ref = s.Ref
	}

	return schema, nil
}

// ParseExpr parses a dst.Expr into an OpenAPI Schema.
func ParseExpr(expr dst.Expr) (*openapi.Schema, error) {
	if expr == nil {
		return nil, fmt.Errorf("expr is nil")
	}

	switch e := expr.(type) {
	case *dst.Ident:
		switch e.Name {
		case "string":
			return &openapi.Schema{Type: "string"}, nil
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return &openapi.Schema{Type: "integer"}, nil
		case "float32", "float64":
			return &openapi.Schema{Type: "number"}, nil
		case "bool":
			return &openapi.Schema{Type: "boolean"}, nil
		case "interface{}":
			return &openapi.Schema{}, nil
		default:
			return &openapi.Schema{Ref: fmt.Sprintf("#/components/schemas/%s", e.Name)}, nil
		}
	case *dst.ArrayType:
		itemsSchema, err := ParseExpr(e.Elt)
		if err != nil {
			return nil, err
		}
		return &openapi.Schema{Type: "array", Items: itemsSchema}, nil
	case *dst.StarExpr:
		return ParseExpr(e.X)
	case *dst.StructType:
		return &openapi.Schema{Type: "object"}, nil
	default:
		return nil, fmt.Errorf("unsupported expr type: %T", expr)
	}
}
