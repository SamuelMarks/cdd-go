// Package daos provides generation of Data Access Objects (DAO) interfaces,
// stubs, and concrete implementations from OpenAPI schemas.
package daos

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// isValidDAOSchemaType checks whether an OpenAPI schema type is valid for DAO generation.
// Parameter t is the schema type string, and it returns true if valid.
func isValidDAOSchemaType(t string) bool {
	switch t {
	case "", "object", "string", "integer", "number", "boolean", "array":
		return true
	default:
		return false
	}
}

// modelExpr returns a qualified models.<name> AST selector expression.
// Parameter name is the model struct name, and it returns the selector expression.
func modelExpr(name string) dst.Expr {
	return &dst.SelectorExpr{
		X:   dst.NewIdent("models"),
		Sel: dst.NewIdent(name),
	}
}

// modelPtrExpr returns a pointer to a qualified *models.<name> AST expression.
// Parameter name is the model struct name, and it returns the star expression.
func modelPtrExpr(name string) dst.Expr {
	return &dst.StarExpr{
		X: modelExpr(name),
	}
}

// SchemaDAOBundle holds all emitted AST declarations for a single schema's DAO interface, stub, and GORM implementation.
type SchemaDAOBundle struct {
	// InterfaceDecl is the abstract <Name>DAO interface declaration.
	InterfaceDecl *dst.GenDecl
	// StubDecl is the <Name>StubDAO struct declaration.
	StubDecl *dst.GenDecl
	// StubMethods are the <Name>StubDAO method declarations.
	StubMethods []dst.Decl
	// ConcreteDecl is the <Name>GormDAO struct declaration.
	ConcreteDecl *dst.GenDecl
	// ConcreteMethods are the <Name>GormDAO method declarations.
	ConcreteMethods []dst.Decl
}

// EmitSchemaDAOs validates a schema once and emits its interface, stub DAO, and concrete GORM DAO declarations.
// Parameter name is the schema model name, schema is the OpenAPI Schema, and it returns the bundle of AST declarations and any error.
func EmitSchemaDAOs(name string, schema *openapi.Schema) (*SchemaDAOBundle, error) {
	ifaceDecl, err := EmitDAOInterface(name, schema)
	if err != nil {
		return nil, err
	}
	stubDecl, stubMethods, _ := EmitStubDAO(name, schema)
	concreteDecl, concreteMethods, _ := EmitConcreteDAO(name, schema)
	return &SchemaDAOBundle{
		InterfaceDecl:   ifaceDecl,
		StubDecl:        stubDecl,
		StubMethods:     stubMethods,
		ConcreteDecl:    concreteDecl,
		ConcreteMethods: concreteMethods,
	}, nil
}

// EmitModelsErrorsFile emits the AST for models/errors.go providing NewNotImplementedError for stub DAOs.
// It returns a complete dst.File for package models.
func EmitModelsErrorsFile() *dst.File {
	fn := &dst.FuncDecl{
		Name: dst.NewIdent("NewNotImplementedError"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{},
			Results: &dst.FieldList{
				List: []*dst.Field{{Type: dst.NewIdent("error")}},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.CallExpr{
							Fun:  &dst.SelectorExpr{X: dst.NewIdent("errors"), Sel: dst.NewIdent("New")},
							Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"not implemented"`}},
						},
					},
				},
			},
		},
	}
	fn.Decs.Start.Append("// NewNotImplementedError returns an error indicating a stubbed DAO operation is not implemented.")

	f := &dst.File{
		Name: dst.NewIdent("models"),
		Decls: []dst.Decl{
			&dst.GenDecl{
				Tok: token.IMPORT,
				Specs: []dst.Spec{
					&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"errors"`}},
				},
			},
			fn,
		},
	}
	f.Decs.Start.Append("// Package models provides generated domain data models and error constructors.")
	return f
}

// EmitDAOInterface generates the abstract DAO interface for a given schema.
// It defines standard CRUD operations: Create, Get, Update, Delete, List.
// Parameter name is the schema model name, schema is the OpenAPI Schema, and it returns the AST declaration and any error.
func EmitDAOInterface(name string, schema *openapi.Schema) (*dst.GenDecl, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	if name == "" || !isValidDAOSchemaType(schema.Type) {
		return nil, fmt.Errorf("invalid schema %q with type %q for DAO interface", name, schema.Type)
	}

	interfaceName := name + "DAO"

	fields := &dst.FieldList{
		List: []*dst.Field{
			// Create(ctx context.Context, item *models.Model) error
			{
				Names: []*dst.Ident{dst.NewIdent("Create")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelPtrExpr(name)},
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
			// Get(ctx context.Context, id string) (*models.Model, error)
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
							{Type: modelPtrExpr(name)},
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
			// Update(ctx context.Context, id string, item *models.Model) error
			{
				Names: []*dst.Ident{dst.NewIdent("Update")},
				Type: &dst.FuncType{
					Params: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("ctx")}, Type: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Context")}},
							{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
							{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelPtrExpr(name)},
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
			// List(ctx context.Context) ([]*models.Model, error)
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
							{Type: &dst.ArrayType{Elt: modelPtrExpr(name)}},
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
// Parameter name is the schema model name, schema is the OpenAPI Schema, and it returns the struct declaration, method declarations, and any error.
func EmitStubDAO(name string, schema *openapi.Schema) (*dst.GenDecl, []dst.Decl, error) {
	if schema == nil {
		return nil, nil, fmt.Errorf("schema is nil")
	}
	if name == "" || !isValidDAOSchemaType(schema.Type) {
		return nil, nil, fmt.Errorf("invalid schema %q with type %q for stub DAO", name, schema.Type)
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

// emitStubMethod generates a stub CRUD method declaration for a stub DAO.
// Parameter receiver is the stub struct name, method is the CRUD method name, model is the schema name, hasItem indicates item param, and returnsModel indicates model return.
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
		params.List = append(params.List, &dst.Field{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelPtrExpr(model)})
	}

	results := &dst.FieldList{List: []*dst.Field{}}
	if returnsModel {
		results.List = append(results.List, &dst.Field{Type: modelPtrExpr(model)})
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

// emitStubListMethod generates the stub List method declaration for a stub DAO.
// Parameter receiver is the stub struct name, model is the schema name, and it returns the AST function declaration.
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
					{Type: &dst.ArrayType{Elt: modelPtrExpr(model)}},
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
// Parameter name is the schema model name, schema is the OpenAPI Schema, and it returns the struct declaration, method declarations, and any error.
func EmitConcreteDAO(name string, schema *openapi.Schema) (*dst.GenDecl, []dst.Decl, error) {
	if schema == nil {
		return nil, nil, fmt.Errorf("schema is nil")
	}
	if name == "" || !isValidDAOSchemaType(schema.Type) {
		return nil, nil, fmt.Errorf("invalid schema %q with type %q for concrete DAO", name, schema.Type)
	}

	concreteName := name + "GormDAO"

	dbField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("DB")},
		Type:  &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}},
		Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`json:\"-\"`"},
	}
	dbField.Decs.Start.Append("// DB is the underlying GORM database connection.")

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(concreteName),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{dbField},
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

// emitConcreteCreateMethod generates the GORM Create method declaration for a concrete DAO.
// Parameter receiver is the concrete DAO struct name, model is the schema name, and it returns the AST function declaration.
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
					{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelPtrExpr(model)},
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

// emitConcreteGetMethod generates the GORM Get method declaration for a concrete DAO.
// Parameter receiver is the concrete DAO struct name, model is the schema name, and it returns the AST function declaration.
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
				List: []*dst.Field{{Type: modelPtrExpr(model)}, {Type: dst.NewIdent("error")}},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.DeclStmt{
					Decl: &dst.GenDecl{
						Tok: token.VAR,
						Specs: []dst.Spec{
							&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelExpr(model)},
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

// emitConcreteUpdateMethod generates the GORM Update method declaration for a concrete DAO.
// Parameter receiver is the concrete DAO struct name, model is the schema name, and it returns the AST function declaration.
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
					{Names: []*dst.Ident{dst.NewIdent("item")}, Type: modelPtrExpr(model)},
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

// emitConcreteDeleteMethod generates the GORM Delete method declaration for a concrete DAO.
// Parameter receiver is the concrete DAO struct name, model is the schema name, and it returns the AST function declaration.
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
									&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: modelExpr(model)}},
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

// emitConcreteListMethod generates the GORM List method declaration for a concrete DAO.
// Parameter receiver is the concrete DAO struct name, model is the schema name, and it returns the AST function declaration.
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
					{Type: &dst.ArrayType{Elt: modelPtrExpr(model)}},
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
							&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("items")}, Type: &dst.ArrayType{Elt: modelPtrExpr(model)}},
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
