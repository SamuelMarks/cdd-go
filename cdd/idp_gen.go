// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/dave/dst"
)

// GenerateIdP generates a generic Integrated Identity Provider for testing environments.
func GenerateIdP(outDir string) error {
	idpDir := filepath.Join(outDir, "idp")
	if err := os.MkdirAll(idpDir, 0755); err != nil {
		return err
	}

	decls := []dst.Decl{
		&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`}},
		}},
		emitIdpRegisterHandler(),
		emitIdpLoginHandler(),
		emitIdpRefreshHandler(),
		emitIdpLogoutHandler(),
	}

	file := &dst.File{
		Name:  dst.NewIdent("idp"),
		Decls: decls,
	}
	file.Decs.Start.Append("// Package idp provides an integrated mock Identity Provider (IdP) for authentication endpoints.")

	return WriteDstFile(filepath.Join(idpDir, "idp_handlers.go"), file)
}

// emitIdpRegisterHandler generates the AST for the mock HandleRegister Gin endpoint.
func emitIdpRegisterHandler() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("HandleRegister"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusOK"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"message"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"User registered successfully"`}}}}}}},
			},
		},
	}
	f.Decs.Start.Append("// HandleRegister mocks a generic user registration endpoint.")
	return f
}

// emitIdpLoginHandler generates the AST for the mock HandleLogin Gin endpoint.
func emitIdpLoginHandler() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("HandleLogin"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusOK"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"token"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"Bearer mock-token-123"`}}}}}}},
			},
		},
	}
	f.Decs.Start.Append("// HandleLogin mocks a generic login endpoint returning a mock JWT token.")
	return f
}

// emitIdpRefreshHandler generates the AST for the mock HandleRefresh Gin endpoint.
func emitIdpRefreshHandler() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("HandleRefresh"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusOK"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"token"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"Bearer mock-token-refreshed"`}}}}}}},
			},
		},
	}
	f.Decs.Start.Append("// HandleRefresh mocks a token refresh endpoint returning a refreshed mock JWT token.")
	return f
}

// emitIdpLogoutHandler generates the AST for the mock HandleLogout Gin endpoint.
func emitIdpLogoutHandler() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("HandleLogout"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusOK"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"message"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"User logged out successfully"`}}}}}}},
			},
		},
	}
	f.Decs.Start.Append("// HandleLogout mocks a user logout endpoint.")
	return f
}
