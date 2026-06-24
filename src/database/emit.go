package database

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitDatabaseConfig generates a struct for Database configuration.
func EmitDatabaseConfig() *dst.GenDecl {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("Config"),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("DatabaseURL")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Ephemeral")}, Type: dst.NewIdent("bool")},
				},
			},
		},
	}
	ts.Decs.Start.Append("// Config holds database connection configuration.")
	return &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{ts}}
}

// EmitInitDB generates the InitDB function to connect to Postgres or Ephemeral SQLite.
func EmitInitDB() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("InitDB"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("cfg")}, Type: dst.NewIdent("Config")},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.IfStmt{
					Cond: &dst.SelectorExpr{X: dst.NewIdent("cfg"), Sel: dst.NewIdent("Ephemeral")},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									&dst.CallExpr{
										Fun: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Open")},
										Args: []dst.Expr{
											&dst.CallExpr{
												Fun:  &dst.SelectorExpr{X: dst.NewIdent("sqlite"), Sel: dst.NewIdent("Open")},
												Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"file::memory:?cache=shared"`}},
											},
											&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Config")}}},
										},
									},
								},
							},
						},
					},
				},
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{X: &dst.SelectorExpr{X: dst.NewIdent("cfg"), Sel: dst.NewIdent("DatabaseURL")}, Op: token.EQL, Y: &dst.BasicLit{Kind: token.STRING, Value: `""`}},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{
							&dst.ReturnStmt{
								Results: []dst.Expr{
									dst.NewIdent("nil"),
									dst.NewIdent("nil"),
								},
							},
						},
					},
				},
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Open")},
							Args: []dst.Expr{
								&dst.CallExpr{
									Fun:  &dst.SelectorExpr{X: dst.NewIdent("postgres"), Sel: dst.NewIdent("Open")},
									Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("cfg"), Sel: dst.NewIdent("DatabaseURL")}},
								},
								&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Config")}}},
							},
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// InitDB creates a new database connection.")
	return f
}

// EmitMigrate generates the Migrate function to run AutoMigrate.
func EmitMigrate(schemas map[string]openapi.Schema) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("Migrate"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("db")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.IfStmt{
					Cond: &dst.BinaryExpr{X: dst.NewIdent("db"), Op: token.EQL, Y: dst.NewIdent("nil")},
					Body: &dst.BlockStmt{
						List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}}},
					},
				},
			},
		},
	}

	for name := range schemas {
		f.Body.List = append(f.Body.List, &dst.IfStmt{
			Init: &dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{
					&dst.CallExpr{
						Fun:  &dst.SelectorExpr{X: dst.NewIdent("db"), Sel: dst.NewIdent("AutoMigrate")},
						Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent(name + "{}")}},
					},
				},
			},
			Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("err")}}},
			},
		})
	}

	f.Body.List = append(f.Body.List, &dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}})
	f.Decs.Start.Append("// Migrate automatically migrates schemas for Concrete DAOs.")
	return f
}
