// Package tests provides parsing and emitting of Go integration test functions to and from OpenAPI operations.
package tests

import (
	"encoding/json"
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ResolveSchema dereferences a schema reference against a map of component schemas if present.
// Parameter s is the schema to resolve, and schemas is the optional map of component schemas.
func ResolveSchema(s *openapi.Schema, schemas map[string]openapi.Schema) *openapi.Schema {
	if s == nil {
		return nil
	}
	if s.Ref != "" && schemas != nil {
		parts := strings.Split(s.Ref, "/")
		refName := parts[len(parts)-1]
		if resolved, ok := schemas[refName]; ok {
			return &resolved
		}
	}
	return s
}

// SynthesizeParamValue produces a deterministic mock string value for a path or query parameter based on its schema.
// Parameter param is the OpenAPI Parameter, and schemas is the optional map of component schemas.
func SynthesizeParamValue(param openapi.Parameter, schemas map[string]openapi.Schema) string {
	s := ResolveSchema(param.Schema, schemas)
	if s != nil {
		if len(s.Enum) > 0 {
			var strVal string
			if err := json.Unmarshal(s.Enum[0], &strVal); err == nil && strVal != "" {
				return strVal
			}
		}
		switch s.Type {
		case "integer", "number":
			return "1"
		case "boolean":
			return "true"
		case "array":
			if s.Items != nil && len(s.Items.Enum) > 0 {
				var strVal string
				if err := json.Unmarshal(s.Items.Enum[0], &strVal); err == nil && strVal != "" {
					return strVal
				}
			}
			return "item1"
		case "string":
			return "dummy"
		}
	}
	if param.Items != nil && len(param.Items.Enum) > 0 {
		var strVal string
		if err := json.Unmarshal(param.Items.Enum[0], &strVal); err == nil && strVal != "" {
			return strVal
		}
	}
	switch param.Type {
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	case "string":
		return "dummy"
	}
	if param.In == "path" {
		if strings.HasSuffix(strings.ToLower(param.Name), "id") {
			return "1"
		}
	}
	return "dummy"
}

// SynthesizeJSONPayload produces a deterministic JSON string payload from an OpenAPI schema.
// Parameter schema is the target OpenAPI Schema, isArray indicates array wrapping, and schemas is the component schema map.
func SynthesizeJSONPayload(schema *openapi.Schema, isArray bool, schemas map[string]openapi.Schema) string {
	resolved := ResolveSchema(schema, schemas)
	if resolved != nil && resolved.Type == "array" {
		isArray = true
		resolved = ResolveSchema(resolved.Items, schemas)
	}
	objStr := `{"id":1,"name":"dummy"}`
	if resolved != nil && len(resolved.Properties) > 0 {
		keys := make([]string, 0, len(resolved.Properties))
		for k := range resolved.Properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var fields []string
		for _, k := range keys {
			prop := resolved.Properties[k]
			propResolved := ResolveSchema(&prop, schemas)
			valStr := synthesizePropertyJSON(k, propResolved)
			fields = append(fields, fmt.Sprintf("%q:%s", k, valStr))
		}
		objStr = "{" + strings.Join(fields, ",") + "}"
	}
	if isArray {
		return "[" + objStr + "]"
	}
	return objStr
}

// synthesizePropertyJSON generates a JSON literal string for a single schema property.
// Parameter name is the property name, and prop is the resolved property schema.
func synthesizePropertyJSON(name string, prop *openapi.Schema) string {
	if prop == nil {
		return `"dummy"`
	}
	if len(prop.Enum) > 0 {
		return string(prop.Enum[0])
	}
	switch prop.Type {
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	case "array":
		return `["http://dummy.com"]`
	case "object":
		return `{}`
	default:
		if prop.Format == "date-time" {
			return `"2023-01-01T00:00:00Z"`
		}
		if prop.Format == "email" || strings.Contains(strings.ToLower(name), "email") {
			return `"dummy@dummy.com"`
		}
		return `"dummy"`
	}
}

// EmitTest generates a dst.FuncDecl for testing an OpenAPI operation.
// Parameter path is the URL path, method is the HTTP verb, and op is the OpenAPI Operation.
func EmitTest(path string, method string, op *openapi.Operation) (*dst.FuncDecl, error) {
	return EmitTestWithSchemas(path, method, op, nil)
}

// EmitTestWithSchemas generates a dst.FuncDecl for testing an OpenAPI operation with component schema resolution.
// Parameter path is the URL path, method is the HTTP verb, op is the OpenAPI Operation, and schemas is the component schemas map.
func EmitTestWithSchemas(path string, method string, op *openapi.Operation, schemas map[string]openapi.Schema) (*dst.FuncDecl, error) {
	if op == nil {
		return nil, fmt.Errorf("Operation is nil")
	}

	name := "Test"
	if op.OperationID != "" {
		name += strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	} else {
		pathCamel := toCamelCase(path)
		name += strings.ToUpper(method[:1]) + method[1:] + pathCamel
	}

	fd := &dst.FuncDecl{
		Name: dst.NewIdent(name),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("t")},
						Type: &dst.StarExpr{
							X: &dst.SelectorExpr{
								X:   dst.NewIdent("testing"),
								Sel: dst.NewIdent("T"),
							},
						},
					},
				},
			},
		},
		Body: &dst.BlockStmt{},
	}

	if op.Summary != "" {
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the %s operation. %d", name, op.Summary, time.Now().UnixNano()))
	} else {
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the operation. %d", name, time.Now().UnixNano()))
	}
	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("_")},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
				Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `""`}},
			},
		},
	})

	pathFilled := path
	var queryParams []string
	for _, param := range op.Parameters {
		if param.In == "path" {
			val := SynthesizeParamValue(param, schemas)
			pathFilled = strings.ReplaceAll(pathFilled, "{"+param.Name+"}", val)
		} else if param.In == "query" && param.Required {
			val := SynthesizeParamValue(param, schemas)
			queryParams = append(queryParams, param.Name+"="+val)
		}
	}
	pathFilled = strings.ReplaceAll(pathFilled, "{", "")
	pathFilled = strings.ReplaceAll(pathFilled, "}", "")

	hasBody := op.RequestBody != nil
	if !hasBody && (method == "post" || method == "put" || method == "patch") {
		hasBody = true
	}

	contentType := "application/json"
	isArray := false
	var bodySchema *openapi.Schema
	if op.RequestBody != nil {
		for k, mt := range op.RequestBody.Content {
			contentType = k
			bodySchema = mt.Schema
			if mt.Schema != nil && mt.Schema.Type == "array" {
				isArray = true
			}
			if k == "application/json" {
				break
			}
		}
	} else if len(op.Consumes) > 0 {
		contentType = op.Consumes[0]
		for _, c := range op.Consumes {
			if c == "application/json" {
				contentType = c
				break
			}
		}
	}
	for _, param := range op.Parameters {
		if param.In == "body" {
			hasBody = true
			if param.Schema != nil {
				bodySchema = param.Schema
				if param.Schema.Type == "array" {
					isArray = true
				}
			}
		}
	}

	urlStr := `"` + pathFilled + `"`
	if len(queryParams) > 0 {
		urlStr = `"` + pathFilled + "?" + strings.Join(queryParams, "&") + `"`
	}

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("baseURL")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Getenv")},
				Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"BASE_URL"`}},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("baseURL"),
			Op: token.EQL,
			Y:  &dst.BasicLit{Kind: token.STRING, Value: `""`},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("baseURL")},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"http://localhost:8080/v2"`}},
				},
			},
		},
	})

	var bodyExpr dst.Expr = dst.NewIdent("nil")
	if hasBody {
		rawPayload := SynthesizeJSONPayload(bodySchema, isArray, schemas)
		jsonStr := strconv.Quote(rawPayload)

		if contentType == "application/x-www-form-urlencoded" {
			jsonStr = `"name=dummy&status=available"`
		} else if contentType == "multipart/form-data" {
			contentType = "multipart/form-data; boundary=dummyboundary"
			multipartPayload := strings.Join([]string{
				"--dummyboundary",
				`Content-Disposition: form-data; name="additionalMetadata"`,
				"",
				"dummy",
				"--dummyboundary",
				`Content-Disposition: form-data; name="file"; filename="dummy.txt"`,
				"Content-Type: application/octet-stream",
				"",
				"dummy",
				"--dummyboundary--",
				"",
			}, string(rune(10)))
			jsonStr = strconv.Quote(multipartPayload)
		}

		bodyExpr = &dst.CallExpr{
			Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
			Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: jsonStr}},
		}
	}

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("NewRequest")},
				Args: []dst.Expr{
					&dst.BasicLit{Kind: token.STRING, Value: `"` + strings.ToUpper(method) + `"`},
					&dst.BinaryExpr{
						X:  dst.NewIdent("baseURL"),
						Op: token.ADD,
						Y:  &dst.BasicLit{Kind: token.STRING, Value: urlStr},
					},
					bodyExpr,
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("err"),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")},
						Args: []dst.Expr{dst.NewIdent("err")},
					},
				},
			},
		},
	})

	if hasBody {
		fd.Body.List = append(fd.Body.List, &dst.ExprStmt{
			X: &dst.CallExpr{
				Fun: &dst.SelectorExpr{
					X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
					Sel: dst.NewIdent("Set"),
				},
				Args: []dst.Expr{
					&dst.BasicLit{Kind: token.STRING, Value: `"Content-Type"`},
					&dst.BasicLit{Kind: token.STRING, Value: `"` + contentType + `"`},
				},
			},
		})
	}
	fd.Body.List = append(fd.Body.List, &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
				Sel: dst.NewIdent("Set"),
			},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: `"Accept"`},
				&dst.BasicLit{Kind: token.STRING, Value: `"application/json"`},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
				Sel: dst.NewIdent("Set"),
			},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: `"api_key"`},
				&dst.BasicLit{Kind: token.STRING, Value: `"special-key"`},
			},
		},
	})
	fd.Body.List = append(fd.Body.List, &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
				Sel: dst.NewIdent("Set"),
			},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: `"Authorization"`},
				&dst.BasicLit{Kind: token.STRING, Value: `"Bearer special-key"`},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("client")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.UnaryExpr{
				Op: token.AND,
				X: &dst.CompositeLit{
					Type: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("Client")},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("err")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("client"), Sel: dst.NewIdent("Do")},
				Args: []dst.Expr{dst.NewIdent("req")},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("err"),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")},
						Args: []dst.Expr{dst.NewIdent("err")},
					},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.DeferStmt{
		Call: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")},
				Sel: dst.NewIdent("Close"),
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("bodyBytes"), dst.NewIdent("errRead")},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("io"), Sel: dst.NewIdent("ReadAll")},
				Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")}},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X: &dst.BinaryExpr{
				X: &dst.BinaryExpr{
					X: &dst.BinaryExpr{
						X: &dst.BinaryExpr{
							X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
							Op: token.GTR,
							Y:  &dst.BasicLit{Kind: token.INT, Value: "400"},
						},
						Op: token.LAND,
						Y: &dst.BinaryExpr{
							X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
							Op: token.NEQ,
							Y:  &dst.BasicLit{Kind: token.INT, Value: "404"},
						},
					},
					Op: token.LAND,
					Y: &dst.BinaryExpr{
						X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
						Op: token.NEQ,
						Y:  &dst.BasicLit{Kind: token.INT, Value: "405"},
					},
				},
				Op: token.LAND,
				Y: &dst.BinaryExpr{
					X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
					Op: token.NEQ,
					Y:  &dst.BasicLit{Kind: token.INT, Value: "415"},
				},
			},
			Op: token.LAND,
			Y: &dst.BinaryExpr{
				X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
				Op: token.NEQ,
				Y:  &dst.BasicLit{Kind: token.INT, Value: "422"},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatalf")},
						Args: []dst.Expr{
							&dst.BasicLit{Kind: token.STRING, Value: `"Expected status < 400, got %d, body: %s"`},
							&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
							&dst.CallExpr{Fun: dst.NewIdent("string"), Args: []dst.Expr{dst.NewIdent("bodyBytes")}},
						},
					},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X: &dst.BinaryExpr{
				X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
				Op: token.EQL,
				Y:  &dst.BasicLit{Kind: token.INT, Value: "404"},
			},
			Op: token.LOR,
			Y: &dst.BinaryExpr{
				X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
				Op: token.EQL,
				Y:  &dst.BasicLit{Kind: token.INT, Value: "415"},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("errRead"),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")},
						Args: []dst.Expr{dst.NewIdent("errRead")},
					},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X: &dst.CallExpr{
				Fun:  dst.NewIdent("len"),
				Args: []dst.Expr{dst.NewIdent("bodyBytes")},
			},
			Op: token.GTR,
			Y:  &dst.BasicLit{Kind: token.INT, Value: "0"},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.DeclStmt{
					Decl: &dst.GenDecl{
						Tok: token.VAR,
						Specs: []dst.Spec{
							&dst.ValueSpec{
								Names: []*dst.Ident{dst.NewIdent("dummyVal")},
								Type:  &dst.InterfaceType{Methods: &dst.FieldList{}},
							},
						},
					},
				},
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("errJSON")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("json"), Sel: dst.NewIdent("Unmarshal")},
							Args: []dst.Expr{
								dst.NewIdent("bodyBytes"),
								&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("dummyVal")},
							},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{
						X:  dst.NewIdent("errJSON"),
						Op: token.NEQ,
						Y:  dst.NewIdent("nil"),
					},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ExprStmt{
								X: &dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")},
									Args: []dst.Expr{dst.NewIdent("errJSON")},
								},
							},
						},
					},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
				Sel: dst.NewIdent("Set"),
			},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: `"X-Random-Cache-Buster"`},
				&dst.BasicLit{Kind: token.STRING, Value: `"` + fmt.Sprintf("%d", time.Now().UnixNano()) + `"`},
			},
		},
	})

	return fd, nil
}

// toCamelCase delegates URL path to PascalCase conversion to naming.ToCamelCasePath.
// Parameter s is the OpenAPI URL path, and it returns the PascalCase identifier.
func toCamelCase(s string) string {
	return naming.ToCamelCasePath(s)
}
