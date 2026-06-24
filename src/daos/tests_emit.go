package daos

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitFactoryTests generates tests for the DAO Factory.
func EmitFactoryTests(schemas map[string]openapi.Schema) (*dst.GenDecl, []dst.Decl) {
	// A simple test func
	f := &dst.FuncDecl{
		Name: dst.NewIdent("TestDAOFactory"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("t")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("testing"), Sel: dst.NewIdent("T")}}},
				},
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("fStub")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("NewDAOFactory"), Args: []dst.Expr{dst.NewIdent("nil"), dst.NewIdent("false")}}}},
			},
		},
	}

	for name := range schemas {
		f.Body.List = append(f.Body.List, &dst.IfStmt{
			Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("_, ok")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.TypeAssertExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("fStub"), Sel: dst.NewIdent(name)}, Args: []dst.Expr{}}, Type: &dst.StarExpr{X: dst.NewIdent(name + "StubDAO")}}}},
			Cond: &dst.UnaryExpr{Op: token.NOT, X: dst.NewIdent("ok")},
			Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Error")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"Expected %sStubDAO"`, name)}}}}}},
		})
	}

	return nil, []dst.Decl{f}
}

// EmitConcreteDAOTests generates tests for Concrete DAOs.
func EmitConcreteDAOTests(schemas map[string]openapi.Schema) []dst.Decl {
	var decls []dst.Decl

	for name := range schemas {
		f := &dst.FuncDecl{
			Name: dst.NewIdent("Test" + name + "ConcreteDAO"),
			Type: &dst.FuncType{
				Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("t")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("testing"), Sel: dst.NewIdent("T")}}}}},
			},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("db"), dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Open")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("sqlite"), Sel: dst.NewIdent("Open")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"file::memory:?cache=shared"`}}}, &dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("Config")}}}}}}},
					&dst.IfStmt{Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")}, Args: []dst.Expr{dst.NewIdent("err")}}}}}},
					&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("db"), Sel: dst.NewIdent("AutoMigrate")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("models." + name + "{}")}}}},

					&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("dao")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: dst.NewIdent(name + "GormDAO"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("DB"), Value: dst.NewIdent("db")}}}}}},

					// Create
					&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("item")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: dst.NewIdent("models." + name)}}}},
					&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.ASSIGN, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("dao"), Sel: dst.NewIdent("Create")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Background")}, Args: []dst.Expr{}}, dst.NewIdent("item")}}}},
					&dst.IfStmt{Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")}, Args: []dst.Expr{dst.NewIdent("err")}}}}}},

					// List
					&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("items"), dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("dao"), Sel: dst.NewIdent("List")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("context"), Sel: dst.NewIdent("Background")}, Args: []dst.Expr{}}}}}},
					&dst.IfStmt{Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Fatal")}, Args: []dst.Expr{dst.NewIdent("err")}}}}}},
					&dst.IfStmt{Cond: &dst.BinaryExpr{X: &dst.CallExpr{Fun: dst.NewIdent("len"), Args: []dst.Expr{dst.NewIdent("items")}}, Op: token.EQL, Y: &dst.BasicLit{Kind: token.INT, Value: "0"}}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("t"), Sel: dst.NewIdent("Error")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"expected items"`}}}}}}},
				},
			},
		}
		decls = append(decls, f)
	}

	return decls
}
