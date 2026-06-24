package cdd

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// GenerateServerMain generates the main entrypoint for the server.
func GenerateServerMain(oa *openapi.OpenAPI, outDir string, database bool, seed bool, idp bool) error {
	mainDir := filepath.Join(outDir, "cmd", "server")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		return err
	}

	bodyStmts := []dst.Stmt{}

	bodyStmts = append(bodyStmts,
		&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("strictValidation")}, Type: dst.NewIdent("bool")}}}},
		&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("enforceAuth")}, Type: dst.NewIdent("bool")}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("strictValidation")}, &dst.BasicLit{Kind: token.STRING, Value: `"strict-validation"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Enforce strict OpenAPI schema validation"`}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("enforceAuth")}, &dst.BasicLit{Kind: token.STRING, Value: `"enforce-auth"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Enforce mock JWT/Bearer authentication"`}}}},
	)

	if database {
		bodyStmts = append(bodyStmts,
			&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("ephemeral")}, Type: dst.NewIdent("bool")}}}},
		)
		if seed {
			bodyStmts = append(bodyStmts, &dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("seed")}, Type: dst.NewIdent("bool")}}}})
		}

		bodyStmts = append(bodyStmts, &dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("ephemeral")}, &dst.BasicLit{Kind: token.STRING, Value: `"ephemeral"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Use a throwaway in-memory database"`}}}})
		if seed {
			bodyStmts = append(bodyStmts, &dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("seed")}, &dst.BasicLit{Kind: token.STRING, Value: `"seed"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Run the fake data seeder on startup"`}}}})
		}
	}

	bodyStmts = append(bodyStmts,
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("Parse")}, Args: []dst.Expr{}}},
		&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("r")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gin"), Sel: dst.NewIdent("Default")}, Args: []dst.Expr{}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("Use")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("middlewares"), Sel: dst.NewIdent("CORSMiddleware")}, Args: []dst.Expr{}}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("Use")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("middlewares"), Sel: dst.NewIdent("ValidationMiddleware")}, Args: []dst.Expr{dst.NewIdent("strictValidation")}}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("Use")}, Args: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("middlewares"), Sel: dst.NewIdent("AuthMockMiddleware")}, Args: []dst.Expr{dst.NewIdent("enforceAuth")}}}}},
	)

	if idp {
		bodyStmts = append(bodyStmts,
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/register"`}, dst.NewIdent("idp.HandleRegister")}}},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/login"`}, dst.NewIdent("idp.HandleLogin")}}},
		)
	}

	if database {
		bodyStmts = append(bodyStmts,
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("dbUrl")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Getenv")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"DATABASE_URL"`}}}}},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("cfg")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CompositeLit{Type: dst.NewIdent("database.Config"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("DatabaseURL"), Value: dst.NewIdent("dbUrl")}, &dst.KeyValueExpr{Key: dst.NewIdent("Ephemeral"), Value: dst.NewIdent("ephemeral")}}}}},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("db"), dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("database"), Sel: dst.NewIdent("InitDB")}, Args: []dst.Expr{dst.NewIdent("cfg")}}}},
			&dst.IfStmt{Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Failed to init DB: %v"`}, dst.NewIdent("err")}}}}}},
		)

		dbBlockStmts := []dst.Stmt{
			&dst.IfStmt{Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("database"), Sel: dst.NewIdent("Migrate")}, Args: []dst.Expr{dst.NewIdent("db")}}}}, Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Migration failed: %v"`}, dst.NewIdent("err")}}}}}},
		}

		if seed {
			dbBlockStmts = append(dbBlockStmts,
				&dst.IfStmt{Cond: dst.NewIdent("seed"), Body: &dst.BlockStmt{List: []dst.Stmt{&dst.IfStmt{Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("seeder"), Sel: dst.NewIdent("SeedDatabase")}, Args: []dst.Expr{dst.NewIdent("db")}}}}, Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Seeding failed: %v"`}, dst.NewIdent("err")}}}}}}}}},
			)
		}

		bodyStmts = append(bodyStmts,
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{X: dst.NewIdent("db"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{List: dbBlockStmts},
			},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("daoFactory")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("NewDAOFactory")}, Args: []dst.Expr{dst.NewIdent("db"), dst.NewIdent("ephemeral")}}}},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Printf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Starting server... DAO Factory initialized: %v"`}, dst.NewIdent("daoFactory")}}},
		)
	}

	bodyStmts = append(bodyStmts,
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Println")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Server listening on :8080"`}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("Run")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `":8080"`}}}},
	)

	mainFunc := &dst.FuncDecl{
		Name: dst.NewIdent("main"),
		Type: &dst.FuncType{Params: &dst.FieldList{}, Results: &dst.FieldList{}},
		Body: &dst.BlockStmt{List: bodyStmts},
	}
	mainFunc.Decs.Start.Append("// main is the server entrypoint.")

	imports := []dst.Spec{
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"flag"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"log"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"os"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`}},
		&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/middlewares"`}},
	}

	if idp {
		imports = append(imports, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/idp"`}})
	}

	if database {
		imports = append(imports,
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/database"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/daos"`}},
		)
		if seed {
			imports = append(imports, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/seeder"`}})
		}
	}

	file := &dst.File{
		Name: dst.NewIdent("main"),
		Decls: []dst.Decl{
			&dst.GenDecl{Tok: token.IMPORT, Specs: imports},
			mainFunc,
		},
	}

	return WriteDstFile(filepath.Join(mainDir, "main.go"), file)
}
