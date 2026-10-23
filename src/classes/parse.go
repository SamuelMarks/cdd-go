// Package classes provides parsing and emitting of Go structs from/to OpenAPI schemas.
package classes

import (
	"fmt"
	"go/token"
	"reflect"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseType parses a dst.TypeSpec into an OpenAPI Schema.
// Parameter ts is the AST type specification, and it returns the parsed OpenAPI Schema and error.
func ParseType(ts *dst.TypeSpec) (*openapi.Schema, error) {
	if ts == nil {
		return nil, fmt.Errorf("TypeSpec is nil")
	}

	schema := &openapi.Schema{}

	if len(ts.Decs.Start) > 0 {
		rawDoc := docstrings.Parse(ts.Decs.Start)
		var descLines []string
		for _, line := range strings.Split(rawDoc, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "@discriminator ") {
				schema.Discriminator = parseDiscriminatorDirective(strings.TrimPrefix(trimmed, "@discriminator "))
			} else if strings.HasPrefix(trimmed, "@xml ") {
				schema.XML = parseXMLDirective(strings.TrimPrefix(trimmed, "@xml "))
			} else if trimmed != "" {
				descLines = append(descLines, trimmed)
			}
		}
		schema.Description = strings.Join(descLines, " ")
	}

	switch t := ts.Type.(type) {
	case *dst.StructType:
		schema.Type = "object"
		props, err := parseStructFields(t)
		if err != nil {
			return nil, err
		}
		schema.Properties = props
	case *dst.ArrayType:
		schema.Type = "array"
		itemsSchema, err := ParseExpr(t.Elt)
		if err != nil {
			return nil, err
		}
		schema.Items = itemsSchema
	case *dst.MapType:
		schema.Type = "object"
		if keyIdent, ok := t.Key.(*dst.Ident); !ok || keyIdent.Name != "string" {
			return nil, fmt.Errorf("map keys must be strings")
		}
		valSchema, err := ParseExpr(t.Value)
		if err != nil {
			return nil, err
		}
		schema.AdditionalProperties = valSchema
	default:
		s, err := ParseExpr(t)
		if err != nil {
			return nil, err
		}
		schema.Type = s.Type
		schema.Format = s.Format
		schema.Ref = s.Ref
	}

	return schema, nil
}

// parseStructFields extracts OpenAPI Schema properties from a dst.StructType's field list.
// Parameter st is the StructType AST node, and it returns the property map and any error.
func parseStructFields(st *dst.StructType) (map[string]openapi.Schema, error) {
	props := make(map[string]openapi.Schema)
	if st == nil || st.Fields == nil {
		return props, nil
	}
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		propName := field.Names[0].Name
		var xmlObj *openapi.XML
		if field.Tag != nil && field.Tag.Kind == token.STRING {
			tagVal := strings.Trim(field.Tag.Value, "`")
			structTag := reflect.StructTag(tagVal)
			if j := structTag.Get("json"); j != "" {
				parts := strings.Split(j, ",")
				if parts[0] != "" && parts[0] != "-" {
					propName = parts[0]
				}
			}
			if x := structTag.Get("xml"); x != "" {
				xmlObj = parseXMLStructTag(x)
			}
		}

		propSchema, err := ParseExpr(field.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", field.Names[0].Name, err)
		}
		propSchema.XML = xmlObj

		if len(field.Decs.Start) > 0 {
			propSchema.Description = strings.ReplaceAll(docstrings.Parse(field.Decs.Start), "\n", " ")
		}

		props[propName] = *propSchema
	}
	return props, nil
}

// parseDiscriminatorDirective parses key=value pairs from a @discriminator docstring directive.
// Parameter rest is the directive body, and it returns the populated Discriminator object.
func parseDiscriminatorDirective(rest string) *openapi.Discriminator {
	d := &openapi.Discriminator{}
	for _, tokenStr := range strings.Fields(rest) {
		kv := strings.SplitN(tokenStr, "=", 2)
		if len(kv) == 2 {
			switch kv[0] {
			case "propertyName":
				d.PropertyName = kv[1]
			case "defaultMapping":
				d.DefaultMapping = kv[1]
			}
		}
	}
	return d
}

// parseXMLDirective parses key=value pairs from an @xml docstring directive.
// Parameter rest is the directive body, and it returns the populated XML object.
func parseXMLDirective(rest string) *openapi.XML {
	x := &openapi.XML{}
	for _, tokenStr := range strings.Fields(rest) {
		kv := strings.SplitN(tokenStr, "=", 2)
		if len(kv) == 2 {
			switch kv[0] {
			case "name":
				x.Name = kv[1]
			case "nodeType":
				x.NodeType = kv[1]
			}
		}
	}
	return x
}

// parseXMLStructTag parses a Go xml:"..." struct tag value into an OpenAPI XML object.
// Parameter tag is the raw xml tag value, and it returns the parsed XML configuration.
func parseXMLStructTag(tag string) *openapi.XML {
	parts := strings.Split(tag, ",")
	x := &openapi.XML{Name: parts[0], NodeType: "element"}
	for _, opt := range parts[1:] {
		switch opt {
		case "attr":
			x.Attribute = true
			x.NodeType = "attribute"
		case "cdata":
			x.NodeType = "cdata"
		case "chardata":
			x.NodeType = "text"
		}
	}
	return x
}

// ParseExpr parses a dst.Expr into an OpenAPI Schema.
// Parameter expr is the AST type expression, and it returns the parsed OpenAPI Schema and error.
func ParseExpr(expr dst.Expr) (*openapi.Schema, error) {
	if expr == nil {
		return nil, fmt.Errorf("expr is nil")
	}

	switch e := expr.(type) {
	case *dst.Ident:
		switch e.Name {
		case "string":
			return &openapi.Schema{Type: "string"}, nil
		case "int32":
			return &openapi.Schema{Type: "integer", Format: "int32"}, nil
		case "int64":
			return &openapi.Schema{Type: "integer", Format: "int64"}, nil
		case "int", "int8", "int16", "uint", "uint8", "uint16", "uint32", "uint64":
			return &openapi.Schema{Type: "integer"}, nil
		case "float32":
			return &openapi.Schema{Type: "number", Format: "float"}, nil
		case "float64":
			return &openapi.Schema{Type: "number"}, nil
		case "bool":
			return &openapi.Schema{Type: "boolean"}, nil
		case "interface{}", "any":
			return &openapi.Schema{}, nil
		default:
			return &openapi.Schema{Ref: fmt.Sprintf("#/components/schemas/%s", e.Name)}, nil
		}
	case *dst.SelectorExpr:
		if x, ok := e.X.(*dst.Ident); ok && x.Name == "time" && e.Sel.Name == "Time" {
			return &openapi.Schema{Type: "string", Format: "date-time"}, nil
		}
		return &openapi.Schema{}, nil
	case *dst.ArrayType:
		itemsSchema, err := ParseExpr(e.Elt)
		if err != nil {
			return nil, err
		}
		return &openapi.Schema{Type: "array", Items: itemsSchema}, nil
	case *dst.StarExpr:
		return ParseExpr(e.X)
	case *dst.StructType:
		props, err := parseStructFields(e)
		if err != nil {
			return nil, err
		}
		if len(props) > 0 {
			return &openapi.Schema{Type: "object", Properties: props}, nil
		}
		return &openapi.Schema{Type: "object"}, nil
	case *dst.MapType:
		if e.Key != nil {
			if keyIdent, ok := e.Key.(*dst.Ident); !ok || keyIdent.Name != "string" {
				return nil, fmt.Errorf("map keys must be strings")
			}
		}
		valSchema, err := ParseExpr(e.Value)
		if err != nil {
			return nil, err
		}
		return &openapi.Schema{Type: "object", AdditionalProperties: valSchema}, nil
	default:
		return nil, fmt.Errorf("unsupported expr type: %T", expr)
	}
}
