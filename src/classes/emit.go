// Package classes provides parsing and emitting of Go structs from/to OpenAPI schemas.
package classes

import (
	"fmt"
	"go/token"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitOptions configures struct tag and field emission for OpenAPI schemas.
type EmitOptions struct {
	// IncludeGormTags specifies whether gorm:"column:..." tags should be emitted on struct fields.
	IncludeGormTags bool
	// OmitEmpty specifies whether json:",omitempty" should be appended to struct field JSON tags.
	OmitEmpty bool
	// UsePascalCaseFieldNames specifies whether property names should be converted using ToPascalCase instead of ExportedName.
	UsePascalCaseFieldNames bool
}

// isValidSchemaType checks whether an OpenAPI schema type is empty or a valid JSON Schema / OpenAPI type.
// Parameter t is the schema type string, and it returns true if valid.
func isValidSchemaType(t string) bool {
	switch t {
	case "", "string", "integer", "number", "boolean", "array", "object", "null":
		return true
	default:
		return false
	}
}

// EmitType converts an OpenAPI Schema into a dst.TypeSpec with default GORM-enabled options.
// Parameter name is the Go type name, and schema is the OpenAPI Schema.
func EmitType(name string, schema *openapi.Schema) (*dst.TypeSpec, error) {
	return EmitTypeWithOptions(name, schema, EmitOptions{IncludeGormTags: true})
}

// EmitTypeWithOptions converts an OpenAPI Schema into a dst.TypeSpec using the provided EmitOptions.
// Parameter name is the Go type name, schema is the OpenAPI Schema, and opts configures tag formatting.
func EmitTypeWithOptions(name string, schema *openapi.Schema, opts EmitOptions) (*dst.TypeSpec, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("schema name is empty")
	}
	if !isValidSchemaType(schema.Type) {
		return nil, fmt.Errorf("unsupported schema type: %s", schema.Type)
	}

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(name),
	}

	if schema.Description != "" {
		ts.Decs.Start.Append(fmt.Sprintf("// %s", schema.Description))
	}

	switch schema.Type {
	case "object", "":
		if schema.AdditionalProperties != nil {
			valExpr := EmitTypeExpr(schema.AdditionalProperties)
			ts.Type = &dst.MapType{
				Key:   dst.NewIdent("string"),
				Value: valExpr,
			}
		} else {
			fields := &dst.FieldList{}
			propNames := make([]string, 0, len(schema.Properties))
			for propName := range schema.Properties {
				propNames = append(propNames, propName)
			}
			sort.Strings(propNames)

			for _, propName := range propNames {
				propSchema := schema.Properties[propName]
				fieldName := exportedName(propName)
				if opts.UsePascalCaseFieldNames {
					fieldName = naming.ToPascalCase(propName)
				}
				fieldType := EmitTypeExpr(&propSchema)

				tagStr := fmt.Sprintf("`json:%q`", propName)
				if opts.OmitEmpty && !opts.IncludeGormTags {
					tagStr = fmt.Sprintf("`"+`json:"%s,omitempty"`+"`", propName)
				} else if opts.IncludeGormTags {
					tagStr = fmt.Sprintf("`"+`json:%q gorm:"column:%s"`+"`", propName, propName)
				}

				field := &dst.Field{
					Names: []*dst.Ident{dst.NewIdent(fieldName)},
					Type:  fieldType,
					Tag: &dst.BasicLit{
						Kind:  token.STRING,
						Value: tagStr,
					},
				}

				if propSchema.Description != "" {
					field.Decs.Start.Append(fmt.Sprintf("// %s", propSchema.Description))
				}

				fields.List = append(fields.List, field)
			}
			ts.Type = &dst.StructType{
				Fields: fields,
			}
		}
	case "array":
		eltExpr := EmitTypeExpr(schema.Items)
		ts.Type = &dst.ArrayType{
			Elt: eltExpr,
		}
	default:
		ts.Type = EmitTypeExpr(schema)
	}

	return ts, nil
}

// EmitTypeExpr converts an OpenAPI Schema into a dst.Expr (for field types, array elements, etc.).
// Parameter schema is the OpenAPI Schema, and it returns the AST type expression.
func EmitTypeExpr(schema *openapi.Schema) dst.Expr {
	if schema == nil {
		return dst.NewIdent("interface{}")
	}

	if schema.Ref != "" {
		parts := strings.Split(schema.Ref, "/")
		refName := naming.ToPascalCase(parts[len(parts)-1])
		return dst.NewIdent(refName)
	}

	switch schema.Type {
	case "string":
		return dst.NewIdent("string")
	case "integer":
		return dst.NewIdent("int")
	case "number":
		return dst.NewIdent("float64")
	case "boolean":
		return dst.NewIdent("bool")
	case "array":
		return &dst.ArrayType{
			Elt: EmitTypeExpr(schema.Items),
		}
	case "object":
		if schema.AdditionalProperties != nil {
			return &dst.MapType{
				Key:   dst.NewIdent("string"),
				Value: EmitTypeExpr(schema.AdditionalProperties),
			}
		}
		return dst.NewIdent("interface{}")
	default:
		return dst.NewIdent("interface{}")
	}
}

// exportedName converts a json field name to an exported Go field name.
// Parameter name is the JSON field name, and it returns the exported Go identifier.
func exportedName(name string) string {
	return naming.ExportedName(name)
}
