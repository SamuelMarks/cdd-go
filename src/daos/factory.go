// Package daos provides generation of Data Access Objects (DAO) interfaces, stubs, and concrete implementations from OpenAPI schemas.
package daos

import (
	"go/token"
	"sort"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitFactoryInterface generates an interface that defines how to retrieve DAOs.
// Parameter schemas is the map of OpenAPI schemas, and it returns the generated AST interface declaration.
func EmitFactoryInterface(schemas map[string]openapi.Schema) *dst.GenDecl {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := &dst.FieldList{List: []*dst.Field{}}
	for _, name := range names {
		fields.List = append(fields.List, &dst.Field{
			Names: []*dst.Ident{dst.NewIdent(name)},
			Type: &dst.FuncType{
				Params:  &dst.FieldList{},
				Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent(name + "DAO")}}},
			},
			Decs: dst.FieldDecorations{
				NodeDecs: dst.NodeDecs{Start: dst.Decorations{"// " + name + " returns the " + name + " DAO."}},
			},
		})
	}
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("Factory"),
		Type: &dst.InterfaceType{Methods: fields},
	}
	ts.Decs.Start.Append("// Factory provides dependency injection for DAOs.")
	return &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{ts}}
}

// EmitConcreteFactory generates the implementation of the Factory.
func EmitConcreteFactory(schemas map[string]openapi.Schema) (*dst.GenDecl, []dst.Decl) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("DAOFactory"),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("db")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}},
					{Names: []*dst.Ident{dst.NewIdent("ephemeral")}, Type: dst.NewIdent("bool")},
				},
			},
		},
	}
	ts.Decs.Start.Append("// DAOFactory is the concrete implementation of Factory.")

	decl := &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{ts}}

	f1 := &dst.FuncDecl{
		Name: dst.NewIdent("NewDAOFactory"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("db")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}},
					{Names: []*dst.Ident{dst.NewIdent("ephemeral")}, Type: dst.NewIdent("bool")},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.StarExpr{X: dst.NewIdent("DAOFactory")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.UnaryExpr{
							Op: token.AND,
							X: &dst.CompositeLit{
								Type: dst.NewIdent("DAOFactory"),
								Elts: []dst.Expr{
									&dst.KeyValueExpr{Key: dst.NewIdent("db"), Value: dst.NewIdent("db")},
									&dst.KeyValueExpr{Key: dst.NewIdent("ephemeral"), Value: dst.NewIdent("ephemeral")},
								},
							},
						},
					},
				},
			},
		},
	}
	f1.Decs.Start.Append("// NewDAOFactory creates a new DAOFactory.")
	methods := []dst.Decl{f1}

	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		f := &dst.FuncDecl{
			Recv: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("f")}, Type: &dst.StarExpr{X: dst.NewIdent("DAOFactory")}}}},
			Name: dst.NewIdent(name),
			Type: &dst.FuncType{
				Params:  &dst.FieldList{},
				Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent(name + "DAO")}}},
			},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.IfStmt{
						Cond: &dst.BinaryExpr{
							X:  &dst.SelectorExpr{X: dst.NewIdent("f"), Sel: dst.NewIdent("db")},
							Op: token.NEQ,
							Y:  dst.NewIdent("nil"),
						},
						Body: &dst.BlockStmt{
							List: []dst.Stmt{
								&dst.ReturnStmt{
									Results: []dst.Expr{
										&dst.UnaryExpr{
											Op: token.AND,
											X: &dst.CompositeLit{
												Type: dst.NewIdent(name + "GormDAO"),
												Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("DB"), Value: &dst.SelectorExpr{X: dst.NewIdent("f"), Sel: dst.NewIdent("db")}}},
											},
										},
									},
								},
							},
						},
					},
					&dst.ReturnStmt{
						Results: []dst.Expr{
							&dst.UnaryExpr{
								Op: token.AND,
								X:  &dst.CompositeLit{Type: dst.NewIdent(name + "StubDAO")},
							},
						},
					},
				},
			},
		}
		f.Decs.Start.Append("// " + name + " returns the appropriate " + name + " DAO.")
		methods = append(methods, f)
	}
	return decl, methods
}
