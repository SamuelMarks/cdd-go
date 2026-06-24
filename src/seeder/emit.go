package seeder

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitSeeder generates a SeedDatabase function and entity pools using gofakeit.
func EmitSeeder(schemas map[string]openapi.Schema) []dst.Decl {
	decls := []dst.Decl{}

	// Entity Pool struct
	poolFields := &dst.FieldList{List: []*dst.Field{}}
	for name := range schemas {
		poolFields.List = append(poolFields.List, &dst.Field{
			Names: []*dst.Ident{dst.NewIdent(name + "IDs")},
			Type:  &dst.ArrayType{Elt: dst.NewIdent("string")},
		})
	}
	poolStruct := &dst.TypeSpec{
		Name: dst.NewIdent("EntityPool"),
		Type: &dst.StructType{Fields: poolFields},
	}
	poolStruct.Decs.Start.Append("// EntityPool caches IDs of generated records to maintain referential integrity.")
	decls = append(decls, &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{poolStruct}})

	// SeedDatabase func
	seedFunc := &dst.FuncDecl{
		Name: dst.NewIdent("SeedDatabase"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("db")}, Type: &dst.StarExpr{X: &dst.SelectorExpr{X: dst.NewIdent("gorm"), Sel: dst.NewIdent("DB")}}},
				},
			},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("error")}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Seed")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "0"}}}},
				&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("pool")}, Type: dst.NewIdent("EntityPool")}}}},
			},
		},
	}
	seedFunc.Decs.Start.Append("// SeedDatabase generates a dependency graph of fake data and inserts it.")

	for name, schema := range schemas {
		seedFunc.Body.List = append(seedFunc.Body.List, &dst.ForStmt{
			Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("i")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "0"}}},
			Cond: &dst.BinaryExpr{X: dst.NewIdent("i"), Op: token.LSS, Y: &dst.BasicLit{Kind: token.INT, Value: "10"}},
			Post: &dst.IncDecStmt{X: dst.NewIdent("i"), Tok: token.INC},
			Body: &dst.BlockStmt{
				List: []dst.Stmt{
					&dst.AssignStmt{
						Lhs: []dst.Expr{dst.NewIdent("item")},
						Tok: token.DEFINE,
						Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("Fake" + name), Args: []dst.Expr{dst.NewIdent("pool")}}},
					},
					&dst.IfStmt{
						Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("db"), Sel: dst.NewIdent("Create")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("item")}}}, Sel: dst.NewIdent("Error")}}},
						Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
						Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("err")}}}},
					},
					// We assume models have an ID string field for pool caching
					&dst.AssignStmt{
						Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(name + "IDs")}},
						Tok: token.ASSIGN,
						Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("append"), Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(name + "IDs")}, &dst.SelectorExpr{X: dst.NewIdent("item"), Sel: dst.NewIdent("Id")}}}},
					},
				},
			},
		})

		// Add Fake Factory func for the schema
		decls = append(decls, emitFakeFactory(name, &schema))
	}

	seedFunc.Body.List = append(seedFunc.Body.List, &dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}})
	decls = append(decls, seedFunc)
	return decls
}

func emitFakeFactory(name string, schema *openapi.Schema) *dst.FuncDecl {
	f := &dst.FuncDecl{
		Name: dst.NewIdent("Fake" + name),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("pool")}, Type: dst.NewIdent("EntityPool")}}},
			Results: &dst.FieldList{List: []*dst.Field{{Type: dst.NewIdent("models." + name)}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{}},
	}
	f.Decs.Start.Append(fmt.Sprintf("// Fake%s generates a random %s.", name, name))

	compLit := &dst.CompositeLit{
		Type: dst.NewIdent("models." + name),
		Elts: []dst.Expr{},
	}

	for propName, propSchema := range schema.Properties {
		goName := exportedName(propName)
		var fakeExpr dst.Expr

		if goName == "Id" || goName == "ID" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("UUID")}}
		} else if propSchema.Type == "string" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Word")}}
		} else if propSchema.Type == "integer" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Number")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "1"}, &dst.BasicLit{Kind: token.INT, Value: "100"}}}
		} else if propSchema.Type == "boolean" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Bool")}}
		} else {
			// default fallback
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Word")}}
		}

		compLit.Elts = append(compLit.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent(goName),
			Value: fakeExpr,
		})
	}

	f.Body.List = append(f.Body.List, &dst.ReturnStmt{Results: []dst.Expr{compLit}})
	return f
}

func exportedName(name string) string {
	if name == "" {
		return ""
	}
	if strings.ToLower(name) == "id" {
		return "Id"
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
