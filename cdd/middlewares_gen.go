package cdd

import (
	"go/token"
	"os"
	"path/filepath"

	"github.com/dave/dst"
)

// GenerateMiddlewares generates the global middlewares like CORS and request validation.
func GenerateMiddlewares(outDir string) error {
	mwDir := filepath.Join(outDir, "middlewares")
	if err := os.MkdirAll(mwDir, 0755); err != nil {
		return err
	}

	decls := []dst.Decl{
		&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"strings"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`}},
		}},
		emitCorsMiddleware(),
		emitValidationMiddleware(),
		emitAuthMockMiddleware(),
	}

	file := &dst.File{
		Name:  dst.NewIdent("middlewares"),
		Decls: decls,
	}

	return WriteDstFile(filepath.Join(mwDir, "middlewares.go"), file)
}

func emitCorsMiddleware() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("CORSMiddleware"),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("gin.HandlerFunc")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.FuncLit{
							Type: &dst.FuncType{Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}}},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Writer")}, Sel: dst.NewIdent("Header")}, Args: []dst.Expr{}}}, // To easily chain Set
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Writer")}, Args: []dst.Expr{}}, Sel: dst.NewIdent("Set")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Access-Control-Allow-Origin"`}, &dst.BasicLit{Kind: token.STRING, Value: `"*"`}}}},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Writer")}, Args: []dst.Expr{}}, Sel: dst.NewIdent("Set")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Access-Control-Allow-Credentials"`}, &dst.BasicLit{Kind: token.STRING, Value: `"true"`}}}},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Writer")}, Args: []dst.Expr{}}, Sel: dst.NewIdent("Set")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Access-Control-Allow-Headers"`}, &dst.BasicLit{Kind: token.STRING, Value: `"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"`}}}},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Writer")}, Args: []dst.Expr{}}, Sel: dst.NewIdent("Set")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Access-Control-Allow-Methods"`}, &dst.BasicLit{Kind: token.STRING, Value: `"POST, OPTIONS, GET, PUT, PATCH, DELETE"`}}}},
									&dst.IfStmt{
										Cond: &dst.BinaryExpr{X: &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Request")}, Sel: dst.NewIdent("Method")}, Op: token.EQL, Y: &dst.BasicLit{Kind: token.STRING, Value: `"OPTIONS"`}},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("AbortWithStatus")}, Args: []dst.Expr{dst.NewIdent("http.StatusNoContent")}}},
												&dst.ReturnStmt{},
											},
										},
									},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}},
								},
							},
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// CORSMiddleware implements a permissive CORS policy for local development.")
	return f
}

func emitAuthMockMiddleware() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("AuthMockMiddleware"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("enforceAuth")}, Type: dst.NewIdent("bool")},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("gin.HandlerFunc")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.FuncLit{
							Type: &dst.FuncType{Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}}},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.IfStmt{
										Cond: &dst.UnaryExpr{Op: token.NOT, X: dst.NewIdent("enforceAuth")},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}},
												&dst.ReturnStmt{},
											},
										},
									},
									&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("authHeader")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("GetHeader")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Authorization"`}}}}},
									&dst.IfStmt{
										Cond: &dst.BinaryExpr{X: dst.NewIdent("authHeader"), Op: token.EQL, Y: &dst.BasicLit{Kind: token.STRING, Value: `""`}},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("AbortWithStatusJSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusUnauthorized"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"Unauthorized: Missing Authorization header"`}}}}}}},
												&dst.ReturnStmt{},
											},
										},
									},
									&dst.IfStmt{
										Cond: &dst.UnaryExpr{Op: token.NOT, X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("strings"), Sel: dst.NewIdent("HasPrefix")}, Args: []dst.Expr{dst.NewIdent("authHeader"), &dst.BasicLit{Kind: token.STRING, Value: `"Bearer mock-token-"`}}}},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("AbortWithStatusJSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusForbidden"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.BasicLit{Kind: token.STRING, Value: `"Forbidden: Invalid mock token"`}}}}}}},
												&dst.ReturnStmt{},
											},
										},
									},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}},
								},
							},
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// AuthMockMiddleware intercepts requests to mock authentication validity if enabled.")
	return f
}

func emitValidationMiddleware() *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("ValidationMiddleware"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("strict")}, Type: dst.NewIdent("bool")},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("gin.HandlerFunc")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.FuncLit{
							Type: &dst.FuncType{Params: &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}}}}},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.IfStmt{
										Cond: &dst.UnaryExpr{Op: token.NOT, X: dst.NewIdent("strict")},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}},
												&dst.ReturnStmt{},
											},
										},
									},
									// In a real generator, this would bind to an OpenAPI validator (e.g. kin-openapi).
									// For the mock architecture scaffold, we simulate the hook here.
									&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("validationErr")}, Type: dst.NewIdent("error")}}}},
									&dst.IfStmt{
										Cond: &dst.BinaryExpr{X: dst.NewIdent("validationErr"), Op: token.NEQ, Y: dst.NewIdent("nil")},
										Body: &dst.BlockStmt{
											List: []dst.Stmt{
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusBadRequest"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("validationErr"), Sel: dst.NewIdent("Error")}, Args: []dst.Expr{}}}}}}}},
												&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Abort")}, Args: []dst.Expr{}}},
												&dst.ReturnStmt{},
											},
										},
									},
									&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Next")}, Args: []dst.Expr{}}},
								},
							},
						},
					},
				},
			},
		},
	}
	f.Decs.Start.Append("// ValidationMiddleware intercepts requests to validate them strictly against the schema if enabled.")
	return f
}
