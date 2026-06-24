// Package daos provides generation of Data Access Objects (DAO) interfaces,
// stubs, and concrete implementations from OpenAPI schemas.
package daos

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitDAOInterface generates the abstract DAO interface for a given schema.
// It defines standard CRUD operations: Create, Get, Update, Delete, List.
func EmitDAOInterface(name string, schema *openapi.Schema) (*dst.GenDecl, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	interfaceName := name + "DAO"

	fields := &dst.FieldList{
		List: []*dst.Field{
			// Create(ctx context.Context, item *Model) error
			{
				Names: []*dst.Ident{dst.NewIdent("Create")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("item")}, Type: &dst.StarExpr{X: dst.NewIdent(name)}},
						},
					},
					Results: &dst.FieldList{
						List: []*dst.Field{{Type: dst.NewIdent("error")}},
					},
				},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// Create inserts a new " + name + " record into the database."},
					},
				},
			},
			// Get(ctx context.Context, id string) (*Model, error)
			{
				Names: []*dst.Ident{dst.NewIdent("Get")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
						},
					},
					Results: &dst.FieldList{
						List: []*dst.Field{
							{Type: &dst.StarExpr{X: dst.NewIdent(name)}},
							{Type: dst.NewIdent("error")},
						},
					},
				},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// Get retrieves a " + name + " record by its ID."},
					},
				},
			},
			// Update(ctx context.Context, id string, item *Model) error
			{
				Names: []*dst.Ident{dst.NewIdent("Update")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
							{Names: []*dst.Ident{dst.NewIdent("item")}, Type: &dst.StarExpr{X: dst.NewIdent(name)}},
						},
					},
					Results: &dst.FieldList{
						List: []*dst.Field{{Type: dst.NewIdent("error")}},
					},
				},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// Update modifies an existing " + name + " record."},
					},
				},
			},
			// Delete(ctx context.Context, id string) error
			{
				Names: []*dst.Ident{dst.NewIdent("Delete")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
						},
					},
					Results: &dst.FieldList{
						List: []*dst.Field{{Type: dst.NewIdent("error")}},
					},
				},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// Delete removes a " + name + " record by its ID."},
					},
				},
			},
			// List(ctx context.Context) ([]*Model, error)
			{
				Names: []*dst.Ident{dst.NewIdent("List")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
						},
					},
					Results: &dst.FieldList{
						List: []*dst.Field{
							{Type: &dst.ArrayType{Elt: &dst.StarExpr{X: dst.NewIdent(name)}}},
							{Type: dst.NewIdent("error")},
						},
					},
				},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{"// List retrieves all " + name + " records."},
					},
				},
			},
		},
	}

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(interfaceName),
		Type: &dst.InterfaceType{
			Methods: fields,
		},
	}
	ts.Decs.Start.Append("// " + interfaceName + " defines the data access object interface for " + name + ".")

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}
	return decl, nil
}

// EmitStubDAO generates the stub implementation of the DAO interface for a given schema.
func EmitStubDAO(name string, schema *openapi.Schema) (*dst.GenDecl, []dst.Decl, error) {
	if schema == nil {
		return nil, nil, fmt.Errorf("schema is nil")
	}

	stubName := name + "StubDAO"

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(stubName),
		Type: &dst.StructType{
			Fields: &dst.FieldList{},
		},
	}
	ts.Decs.Start.Append("// " + stubName + " is the stub implementation of " + name + "DAO.")

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}

	methods := []dst.Decl{
		emitStubMethod(stubName, "Create", name, true, false),
		emitStubMethod(stubName, "Get", name, true, true),
		emitStubMethod(stubName, "Update", name, true, false),
		emitStubMethod(stubName, "Delete", name, false, false),
		emitStubListMethod(stubName, name),
	}

	return decl, methods, nil
}

func emitStubMethod(receiver string, method string, model string, hasItem bool, returnsModel bool) *dst.FuncDecl {
	params := &dst.FieldList{
		List: []*dst.Field{
			{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
		},
	}

	if method == "Get" || method == "Update" || method == "Delete" {
		params.List = append(params.List, &dst.Field{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")})
	}
	if method == "Create" || method == "Update" {
		params.List = append(params.List, &dst.Field{Names: []*dst.Ident{dst.NewIdent("item")}, Type: &dst.StarExpr{X: dst.NewIdent(model)}})
	}

	results := &dst.FieldList{List: []*dst.Field{}}
	if returnsModel {
		results.List = append(results.List, &dst.Field{Type: &dst.StarExpr{X: dst.NewIdent(model)}})
	}
	results.List = append(results.List, &dst.Field{Type: dst.NewIdent("error")})

	bodyList := []dst.Stmt{}
	errStmt := &dst.ReturnStmt{
		Results: []dst.Expr{
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("models"), Sel: dst.NewIdent("NewNotImplementedError")},
				Args: []dst.Expr{},
			},
		},
	}
	if returnsModel {
		errStmt.Results = []dst.Expr{
			dst.NewIdent("nil"),
			&dst.CallExpr{
				Fun:  &dst.SelectorExpr{X: dst.NewIdent("models"), Sel: dst.NewIdent("NewNotImplementedError")},
				Args: []dst.Expr{},
			},
		}
	}
	bodyList = append(bodyList, errStmt)

	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("s")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}},
			},
		},
		Name: dst.NewIdent(method),
		Type: &dst.FuncType{
			Params:  params,
			Results: results,
		},
		Body: &dst.BlockStmt{List: bodyList},
	}
	f.Decs.Start.Append("// " + method + " stubs the " + method + " operation.")
	return f
}

func emitStubListMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("s")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}},
			},
		},
		Name: dst.NewIdent("List"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.ArrayType{Elt: &dst.StarExpr{X: dst.NewIdent(model)}}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						dst.NewIdent("nil"),
						&dst.CallExpr{
							Fun:  &dst.SelectorExpr{X: dst.NewIdent("models"), Sel: dst.NewIdent("NewNotImplementedError")},
							Args: []dst.Expr{},
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// List stubs the List operation.")
	return f
}

// EmitConcreteDAO generates the GORM-backed implementation of the DAO interface for a given schema.
func EmitConcreteDAO(name string, schema *openapi.Schema) (*dst.GenDecl, []dst.Decl, error) {
	if schema == nil {
		return nil, nil, fmt.Errorf("schema is nil")
	}

	concreteName := name + "GormDAO"

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(concreteName),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("DB")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}, Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"-\"`"}},
				},
			},
		},
	}
	ts.Decs.Start.Append("// " + concreteName + " is the GORM implementation of " + name + "DAO.")

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}

	methods := []dst.Decl{
		emitConcreteCreateMethod(concreteName, name),
		emitConcreteGetMethod(concreteName, name),
		emitConcreteUpdateMethod(concreteName, name),
		emitConcreteDeleteMethod(concreteName, name),
		emitConcreteListMethod(concreteName, name),
	}

	return decl, methods, nil
}

func emitConcreteCreateMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("d")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}}},
		},
		Name: dst.NewIdent("Create"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
					{Names: []*dst.Ident{dst.NewIdent("item")}, Type: &dst.StarExpr{X: dst.NewIdent(model)}},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("error")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.SelectorExpr{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("d"), Sel: dst.NewIdent("DB")}, Sel: dst.NewIdent("WithContext")},
										Args: []dst.Expr{dst.NewIdent("ctx")},
									},
									Sel: dst.NewIdent("Create"),
								},
								Args: []dst.Expr{dst.NewIdent("item")},
							},
							Sel: dst.NewIdent("Error"),
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// Create implements the Create operation using GORM.")
	return f
}

func emitConcreteGetMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("d")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}}},
		},
		Name: dst.NewIdent("Get"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
					{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{{Type: &dst.StarExpr{X: dst.NewIdent(model)}}, {Type: dst.NewIdent("error")}},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.DeclStmt{
					Decl: &dst.GenDecl{
						Tok: token.VAR,
						Specs: []dst.Spec{
							&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("item")}, Type: dst.NewIdent(model)},
						},
					},
				},
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.SelectorExpr{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("d"), Sel: dst.NewIdent("DB")}, Sel: dst.NewIdent("WithContext")},
										Args: []dst.Expr{dst.NewIdent("ctx")},
									},
									Sel: dst.NewIdent("First"),
								},
								Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("item")}, dst.NewIdent("id")},
							},
							Sel: dst.NewIdent("Error"),
						},
					},
				},
				&dst.ReturnStmt{
					Results: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("item")}, dst.NewIdent("err")},
				},
			},
		},
	}
	f.Decs.Start.Append("// Get implements the Get operation using GORM.")
	return f
}

func emitConcreteUpdateMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("d")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}}},
		},
		Name: dst.NewIdent("Update"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
					{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("item")}, Type: &dst.StarExpr{X: dst.NewIdent(model)}},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("error")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.SelectorExpr{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("d"), Sel: dst.NewIdent("DB")}, Sel: dst.NewIdent("WithContext")},
										Args: []dst.Expr{dst.NewIdent("ctx")},
									},
									Sel: dst.NewIdent("Save"),
								},
								Args: []dst.Expr{dst.NewIdent("item")},
							},
							Sel: dst.NewIdent("Error"),
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// Update implements the Update operation using GORM.")
	return f
}

func emitConcreteDeleteMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("d")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}}},
		},
		Name: dst.NewIdent("Delete"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
					{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("error")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.SelectorExpr{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("d"), Sel: dst.NewIdent("DB")}, Sel: dst.NewIdent("WithContext")},
										Args: []dst.Expr{dst.NewIdent("ctx")},
									},
									Sel: dst.NewIdent("Delete"),
								},
								Args: []dst.Expr{
									&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent(model + "{}")},
									dst.NewIdent("id"),
								},
							},
							Sel: dst.NewIdent("Error"),
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// Delete implements the Delete operation using GORM.")
	return f
}

func emitConcreteListMethod(receiver string, model string) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Recv: &dst.FieldList{
			List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("d")}, Type: &dst.StarExpr{X: dst.NewIdent(receiver)}}},
		},
		Name: dst.NewIdent("List"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.ArrayType{Elt: &dst.StarExpr{X: dst.NewIdent(model)}}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.DeclStmt{
					Decl: &dst.GenDecl{
						Tok: token.VAR,
						Specs: []dst.Spec{
							&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("items")}, Type: &dst.ArrayType{Elt: &dst.StarExpr{X: dst.NewIdent(model)}}},
						},
					},
				},
				&dst.AssignStmt{
					Lhs: []dst.Expr{dst.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []dst.Expr{
						&dst.SelectorExpr{
							X: &dst.CallExpr{
								Fun: &dst.SelectorExpr{
									X: &dst.CallExpr{
										Fun:  &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("d"), Sel: dst.NewIdent("DB")}, Sel: dst.NewIdent("WithContext")},
										Args: []dst.Expr{dst.NewIdent("ctx")},
									},
									Sel: dst.NewIdent("Find"),
								},
								Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("items")}},
							},
							Sel: dst.NewIdent("Error"),
						},
					},
				},
				&dst.ReturnStmt{
					Results: []dst.Expr{dst.NewIdent("items"), dst.NewIdent("err")},
				},
			},
		},
	}
	f.Decs.Start.Append("// List implements the List operation using GORM.")
	return f
}
