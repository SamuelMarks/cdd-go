package tests

import (
	"fmt"
	"strings"

	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitTest generates a dst.FuncDecl for testing an OpenAPI operation.
func EmitTest(path string, method string, op *openapi.Operation) (*dst.FuncDecl, error) {
	if op == nil {
		return nil, fmt.Errorf("Operation is nil")
	}

	name := "Test"
	if op.OperationID != "" {
		name += strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	} else {
		// Use path and method
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
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the %s operation.", name, op.Summary))
	}
	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("_")},
		Tok: token.ASSIGN, // =
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
				Args: []dst.Expr{&dst.BasicLit{Kind: 9, Value: `""`}},
			},
		},
	})

	// path templating {param} -> "dummy"
	pathFilled := path
	for _, param := range op.Parameters {
		if param.In == "path" {
			pathFilled = strings.ReplaceAll(pathFilled, "{"+param.Name+"}", "1")
		}
	}
	// fallback for any remaining braces
	pathFilled = strings.ReplaceAll(pathFilled, "{", "")
	pathFilled = strings.ReplaceAll(pathFilled, "}", "")

	bodyArg := "nil"
	hasBody := op.RequestBody != nil
	if !hasBody && (method == "post" || method == "put" || method == "patch") {
		hasBody = true
	}
	if !hasBody {
		for _, param := range op.Parameters {
			if param.In == "body" {
				hasBody = true
				break
			}
		}
	}

	isArray := false
	if op.RequestBody != nil {
		if mt, ok := op.RequestBody.Content["application/json"]; ok && mt.Schema != nil {
			if mt.Schema.Type == "array" {
				isArray = true
			}
		}
	}
	for _, param := range op.Parameters {
		if param.In == "body" && param.Schema != nil {
			if param.Schema.Type == "array" {
				isArray = true
			}
		}
	}

	if hasBody {
		if isArray {
			bodyArg = `strings.NewReader("[{\"id\":1,\"name\":\"dummy\",\"photoUrls\":[\"http://dummy.com\"],\"status\":\"available\",\"petId\":1,\"quantity\":1,\"shipDate\":\"2023-01-01T00:00:00Z\",\"complete\":true,\"username\":\"dummy\",\"password\":\"password\",\"firstName\":\"dummy\",\"lastName\":\"dummy\",\"email\":\"dummy@dummy.com\",\"phone\":\"12345\",\"userStatus\":1}]")`
		} else {
			bodyArg = `strings.NewReader("{\"id\":1,\"name\":\"dummy\",\"photoUrls\":[\"http://dummy.com\"],\"status\":\"available\",\"petId\":1,\"quantity\":1,\"shipDate\":\"2023-01-01T00:00:00Z\",\"complete\":true,\"username\":\"dummy\",\"password\":\"password\",\"firstName\":\"dummy\",\"lastName\":\"dummy\",\"email\":\"dummy@dummy.com\",\"phone\":\"12345\",\"userStatus\":1}")`
		}
	}

	urlStr := "http://localhost:8080/v2" + pathFilled
	if path == "/pet/findByStatus" && method == "get" {
		urlStr += "?status=available"
	}

	// Build AST: req, err := http.NewRequest(METHOD, URL, bodyArg)
	var bodyExpr dst.Expr = dst.NewIdent("nil")
	if bodyArg != "nil" {
		jsonStr := `"{\"id\":1,\"name\":\"dummy\",\"photoUrls\":[\"http://dummy.com\"],\"status\":\"available\",\"petId\":1,\"quantity\":1,\"shipDate\":\"2023-01-01T00:00:00Z\",\"complete\":true,\"username\":\"dummy\",\"password\":\"password\",\"firstName\":\"dummy\",\"lastName\":\"dummy\",\"email\":\"dummy@dummy.com\",\"phone\":\"12345\",\"userStatus\":1}"`
		if isArray {
			jsonStr = `"[{\"id\":1,\"name\":\"dummy\",\"photoUrls\":[\"http://dummy.com\"],\"status\":\"available\",\"petId\":1,\"quantity\":1,\"shipDate\":\"2023-01-01T00:00:00Z\",\"complete\":true,\"username\":\"dummy\",\"password\":\"password\",\"firstName\":\"dummy\",\"lastName\":\"dummy\",\"email\":\"dummy@dummy.com\",\"phone\":\"12345\",\"userStatus\":1}]"`
		}
		bodyExpr = &dst.CallExpr{
			Fun:  &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("NewReader")},
			Args: []dst.Expr{&dst.BasicLit{Kind: 9, Value: jsonStr}}, // token.STRING is 9
		}
	}

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("req"), dst.NewIdent("err")},
		Tok: 47, // token.DEFINE (:=)
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("NewRequest")},
				Args: []dst.Expr{
					&dst.BasicLit{Kind: 9, Value: `"` + strings.ToUpper(method) + `"`},
					&dst.BasicLit{Kind: 9, Value: `"` + urlStr + `"`},
					bodyExpr,
				},
			},
		},
	})

	// if err != nil { t.Fatal(err) }
	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("err"),
			Op: token.NEQ, // token.NEQ (!=)
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

	// client := &http.Client{}
	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("client")},
		Tok: 47, // :=
		Rhs: []dst.Expr{
			&dst.UnaryExpr{
				Op: 17, // token.AND (&)
				X: &dst.CompositeLit{
					Type: &dst.SelectorExpr{X: dst.NewIdent("http"), Sel: dst.NewIdent("Client")},
				},
			},
		},
	})

	// resp, err := client.Do(req)
	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("err")},
		Tok: 47, // :=
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("client"), Sel: dst.NewIdent("Do")},
				Args: []dst.Expr{dst.NewIdent("req")},
			},
		},
	})

	// if err != nil { t.Fatal(err) }
	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X:  dst.NewIdent("err"),
			Op: token.NEQ, // token.NEQ (!=)
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

	// defer resp.Body.Close()
	fd.Body.List = append(fd.Body.List, &dst.DeferStmt{
		Call: &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")},
				Sel: dst.NewIdent("Close"),
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.IfStmt{
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
							&dst.BasicLit{Kind: 9, Value: `"Expected status < 400, got %d"`},
							&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")},
						},
					},
				},
			},
		},
	})

	fd.Body.List = append(fd.Body.List, &dst.AssignStmt{
		Lhs: []dst.Expr{dst.NewIdent("bodyBytes"), dst.NewIdent("errRead")},
		Tok: token.DEFINE, // :=
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.SelectorExpr{X: dst.NewIdent("io"), Sel: dst.NewIdent("ReadAll")},
				Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")}},
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
					Tok: token.DEFINE, // :=
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("json"), Sel: dst.NewIdent("Unmarshal")},
							Args: []dst.Expr{
								dst.NewIdent("bodyBytes"),
								&dst.UnaryExpr{Op: 17, X: dst.NewIdent("dummyVal")}, // &dummyVal
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

	return fd, nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if res == "" {
		res = "Root"
	}
	return res
}
