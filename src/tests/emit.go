// Package tests provides parsing and emitting of Go integration test functions to and from OpenAPI operations.
package tests

import (
	"encoding/json"
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"

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
	if param.In == "path" && strings.HasSuffix(strings.ToLower(param.Name), "id") {
		return "1"
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

// resolveTestOperationMethodName computes the Go test function name (e.g., TestGetUser) for an operation.
// Parameter path is the URL path, method is the HTTP verb, and op is the OpenAPI Operation.
func resolveTestOperationMethodName(path, method string, op *openapi.Operation) string {
	if op.OperationID != "" {
		return "Test" + strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	}
	pathCamel := naming.ToCamelCasePath(path)
	return "Test" + strings.ToUpper(method[:1]) + method[1:] + pathCamel
}

// buildTestURLAndPathParams interpolates path parameters and appends required query parameters to build the quoted URL path literal.
// Parameter path is the URL path template, op is the OpenAPI Operation, and schemas is the component schemas map.
func buildTestURLAndPathParams(path string, op *openapi.Operation, schemas map[string]openapi.Schema) string {
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
	if len(queryParams) > 0 {
		return `"` + pathFilled + "?" + strings.Join(queryParams, "&") + `"`
	}
	return `"` + pathFilled + `"`
}

// buildTestRequestBodyAST determines whether the operation has a request body, resolves its content type, and builds the reader AST expression.
// Parameter method is the HTTP verb, op is the OpenAPI Operation, and schemas is the component schemas map.
func buildTestRequestBodyAST(method string, op *openapi.Operation, schemas map[string]openapi.Schema) (bool, dst.Expr, string) {
	hasBody := op.RequestBody != nil || method == "post" || method == "put" || method == "patch"
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

	if !hasBody {
		return false, dst.NewIdent("nil"), contentType
	}

	rawPayload := SynthesizeJSONPayload(bodySchema, isArray, schemas)
	jsonStr := strconv.Quote(rawPayload)

	switch contentType {
	case "application/x-www-form-urlencoded":
		jsonStr = `"name=dummy&status=available"`
	case "multipart/form-data":
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

	bodyExpr := &dst.CallExpr{
		Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
		Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: jsonStr}},
	}
	return true, bodyExpr, contentType
}

// buildTestHTTPRequestStmts constructs the AST statements that initialize baseURL, create the HTTP request, execute it, and assert the response.
// Parameter method is the HTTP verb, urlStr is the quoted URL path, hasBody indicates body presence, bodyExpr is the body reader AST, contentType is the MIME type, and defaultBaseURL is the fallback server URL.
func buildTestHTTPRequestStmts(method, urlStr string, hasBody bool, bodyExpr dst.Expr, contentType string, defaultBaseURL string) []dst.Stmt {
	if defaultBaseURL == "" {
		defaultBaseURL = "http://localhost:8080"
	}
	stmts := []dst.Stmt{
		&dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("baseURL")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun:  &dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Getenv")},
					Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"BASE_URL"`}},
				},
			},
		},
		&dst.IfStmt{
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
						Rhs: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(defaultBaseURL)}},
					},
				},
			},
		},
		&dst.AssignStmt{
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
		},
		buildIfErrFatalStmt("err"),
	}

	if hasBody {
		stmts = append(stmts, buildReqHeaderStmt("Content-Type", contentType))
	}
	stmts = append(stmts,
		buildReqHeaderStmt("Accept", "application/json"),
		buildReqHeaderStmt("api_key", "special-key"),
		buildReqHeaderStmt("Authorization", "Bearer special-key"),
		&dst.AssignStmt{
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
		},
		&dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun:  &dst.SelectorExpr{X: dst.NewIdent("client"), Sel: dst.NewIdent("Do")},
					Args: []dst.Expr{dst.NewIdent("req")},
				},
			},
		},
		buildIfErrFatalStmt("err"),
		&dst.DeferStmt{
			Call: &dst.CallExpr{
				Fun: &dst.SelectorExpr{
					X:   &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")},
					Sel: dst.NewIdent("Close"),
				},
			},
		},
		&dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("bodyBytes"), dst.NewIdent("errRead")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun:  &dst.SelectorExpr{X: dst.NewIdent("io"), Sel: dst.NewIdent("ReadAll")},
					Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")}},
				},
			},
		},
	)
	stmts = append(stmts, buildResponseAssertionStmts()...)
	return stmts
}

// buildReqHeaderStmt constructs a `req.Header.Set(key, val)` AST statement.
// Parameter key is the HTTP header name, and val is the header value.
func buildReqHeaderStmt(key, val string) dst.Stmt {
	return &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Header")},
				Sel: dst.NewIdent("Set"),
			},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(key)},
				&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote(val)},
			},
		},
	}
}

// buildIfErrFatalStmt constructs an `if <errVar> != nil { t.Fatal(<errVar>) }` AST statement.
// Parameter errVar is the error variable identifier.
func buildIfErrFatalStmt(errVar string) dst.Stmt {
	return &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent(errVar),
			Op: token.NEQ,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun:  &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")},
						Args: []dst.Expr{dst.NewIdent(errVar)},
					},
				},
			},
		},
	}
}

// buildResponseAssertionStmts constructs the status code and JSON body assertion AST statements.
// It returns the slice of AST assertion statements.
func buildResponseAssertionStmts() []dst.Stmt {
	return []dst.Stmt{
		&dst.IfStmt{
			Cond: &dst.BinaryExpr{
				X:  &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
				Op: token.GEQ,
				Y:  &dst.BasicLit{Kind: token.INT, Value: "400"},
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
		},
		buildIfErrFatalStmt("errRead"),
		&dst.IfStmt{
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
									Type:  dst.NewIdent("json.RawMessage"),
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
					buildIfErrFatalStmt("errJSON"),
				},
			},
		},
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
	fd, _, err := EmitTestWithConfig(path, method, op, schemas, "http://localhost:8080")
	return fd, err
}

// EmitTestWithConfig generates a dst.FuncDecl for testing an OpenAPI operation with component schema resolution and a configurable default base URL.
// Parameter path is the URL path, method is the HTTP verb, op is the OpenAPI Operation, schemas is the component schemas map, and defaultBaseURL is the fallback server URL.
func EmitTestWithConfig(path string, method string, op *openapi.Operation, schemas map[string]openapi.Schema, defaultBaseURL string) (*dst.FuncDecl, bool, error) {
	if op == nil {
		return nil, false, fmt.Errorf("Operation is nil")
	}

	name := resolveTestOperationMethodName(path, method, op)
	urlStr := buildTestURLAndPathParams(path, op, schemas)
	hasBody, bodyExpr, contentType := buildTestRequestBodyAST(method, op, schemas)

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
		Body: &dst.BlockStmt{
			List: buildTestHTTPRequestStmts(method, urlStr, hasBody, bodyExpr, contentType, defaultBaseURL),
		},
	}

	if op.Summary != "" {
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the %s operation.", name, op.Summary))
	} else {
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the operation.", name))
	}

	return fd, hasBody, nil
}
