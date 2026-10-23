// Package seeder provides AST emission for database seeding routines and fake entity factories.
package seeder

import (
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// resolveForeignKeyTarget checks if propName is a foreign key (e.g., "user_id" or "userId") matching a known schema.
// Parameter selfName is the owning schema name, propName is the property name, and schemas is the map of all schemas.
func resolveForeignKeyTarget(selfName, propName string, schemas map[string]openapi.Schema) string {
	lower := strings.ToLower(propName)
	var prefix string
	if strings.HasSuffix(lower, "_id") && len(lower) > 3 {
		prefix = lower[:len(lower)-3]
	} else if strings.HasSuffix(lower, "id") && len(lower) > 2 {
		prefix = lower[:len(lower)-2]
	}
	if prefix == "" {
		return ""
	}
	for candidate := range schemas {
		if candidate != selfName && strings.EqualFold(candidate, prefix) {
			return candidate
		}
	}
	return ""
}

// topologicalSortSchemas orders schemas so parent entities are seeded before dependent child entities.
// Parameter schemas is the map of OpenAPI schemas, and it returns the topologically sorted schema names and parent set.
func topologicalSortSchemas(schemas map[string]openapi.Schema) ([]string, map[string]bool) {
	allNames := make([]string, 0, len(schemas))
	for name := range schemas {
		allNames = append(allNames, name)
	}
	sort.Strings(allNames)

	deps := make(map[string][]string, len(schemas))
	hasDeps := make(map[string]bool, len(schemas))
	for _, name := range allNames {
		s := schemas[name]
		propNames := make([]string, 0, len(s.Properties))
		for p := range s.Properties {
			propNames = append(propNames, p)
		}
		sort.Strings(propNames)
		for _, p := range propNames {
			if target := resolveForeignKeyTarget(name, p, schemas); target != "" {
				deps[name] = append(deps[name], target)
				hasDeps[name] = true
			}
		}
	}

	visited := make(map[string]bool, len(schemas))
	ordered := make([]string, 0, len(schemas))
	var visit func(n string)
	visit = func(n string) {
		if visited[n] {
			return
		}
		visited[n] = true
		for _, d := range deps[n] {
			visit(d)
		}
		ordered = append(ordered, n)
	}
	for _, name := range allNames {
		visit(name)
	}
	return ordered, hasDeps
}

// seedCountForSchema returns the realistic seeding count for a schema (e.g., 100 for audit logs, 50 for child entities, 10 for parent entities).
// Parameter name is the schema name, and isChild indicates whether the schema depends on a parent schema.
func seedCountForSchema(name string, isChild bool) int {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "audit") || strings.Contains(lower, "log") {
		return 100
	}
	if isChild {
		return 50
	}
	return 10
}

// EmitSeeder generates a SeedDatabase function and entity pools using gofakeit.
// Parameter schemas is the map of OpenAPI schemas, and it returns the generated AST declarations.
func EmitSeeder(schemas map[string]openapi.Schema) []dst.Decl {
	decls := []dst.Decl{}

	schemaNames, hasDeps := topologicalSortSchemas(schemas)

	// Entity Pool struct
	poolFields := &dst.FieldList{List: []*dst.Field{}}
	for _, name := range schemaNames {
		f := &dst.Field{
			Names: []*dst.Ident{dst.NewIdent(name + "IDs")},
			Type:  &dst.ArrayType{Elt: dst.NewIdent("string")},
		}
		f.Decs.Start.Append(fmt.Sprintf("// %sIDs stores generated primary key IDs for %s entities.", name, name))
		poolFields.List = append(poolFields.List, f)
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
	seedFunc.Decs.Start.Append("// SeedDatabase generates a dependency graph of fake data and inserts it. Parameter db is the GORM database instance.")

	for _, name := range schemaNames {
		schema := schemas[name]
		loopStmts := []dst.Stmt{
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("item")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("Fake" + name), Args: []dst.Expr{dst.NewIdent("pool")}}},
			},
			&dst.IfStmt{
				Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.SelectorExpr{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("db"), Sel: dst.NewIdent("Create")}, Args: []dst.Expr{dst.NewIdent("item")}}, Sel: dst.NewIdent("Error")}}},
				Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("err")}}}},
			},
		}
		if idProp, hasID := schema.Properties["id"]; hasID && idProp.Type == "string" {
			loopStmts = append(loopStmts, &dst.AssignStmt{
				Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(name + "IDs")}},
				Tok: token.ASSIGN,
				Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("append"), Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(name + "IDs")}, &dst.SelectorExpr{X: dst.NewIdent("item"), Sel: dst.NewIdent("Id")}}}},
			})
		} else {
			loopStmts = append(loopStmts, &dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("_")},
				Tok: token.ASSIGN,
				Rhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(name + "IDs")}},
			})
		}

		count := seedCountForSchema(name, hasDeps[name])
		seedFunc.Body.List = append(seedFunc.Body.List, &dst.ForStmt{
			Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("i")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "0"}}},
			Cond: &dst.BinaryExpr{X: dst.NewIdent("i"), Op: token.LSS, Y: &dst.BasicLit{Kind: token.INT, Value: strconv.Itoa(count)}},
			Post: &dst.IncDecStmt{X: dst.NewIdent("i"), Tok: token.INC},
			Body: &dst.BlockStmt{
				List: loopStmts,
			},
		})

		// Add Fake Factory func for the schema
		decls = append(decls, emitFakeFactory(name, &schema, schemas))
	}

	seedFunc.Body.List = append(seedFunc.Body.List, &dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("nil")}})
	decls = append(decls, seedFunc)
	return decls
}

// emitFakeFactory generates a Fake<Name> function for a single schema.
// Parameter name is the model name, schema is the OpenAPI Schema, and allSchemas is the map of all schemas for foreign key resolution.
func emitFakeFactory(name string, schema *openapi.Schema, allSchemas map[string]openapi.Schema) *dst.FuncDecl {
	modelSel := &dst.SelectorExpr{
		X:   dst.NewIdent("models"),
		Sel: dst.NewIdent(name),
	}
	f := &dst.FuncDecl{
		Name: dst.NewIdent("Fake" + name),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{List: []*dst.Field{{Names: []*dst.Ident{dst.NewIdent("pool")}, Type: dst.NewIdent("EntityPool")}}},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.StarExpr{X: modelSel}}}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{}},
	}
	f.Decs.Start.Append(fmt.Sprintf("// Fake%s generates a random %s. Parameter pool holds cached entity IDs.", name, name))

	compLit := &dst.CompositeLit{
		Type: &dst.SelectorExpr{
			X:   dst.NewIdent("models"),
			Sel: dst.NewIdent(name),
		},
		Elts: []dst.Expr{},
	}

	propNames := make([]string, 0, len(schema.Properties))
	for propName := range schema.Properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	var fkAssignStmts []dst.Stmt

	for _, propName := range propNames {
		propSchema := schema.Properties[propName]
		goName := naming.ExportedName(propName)
		if goName == "" {
			continue
		}
		if propSchema.Ref != "" || propSchema.Type == "array" || propSchema.AdditionalProperties != nil || len(propSchema.Properties) > 0 {
			continue
		}
		var fakeExpr dst.Expr

		if (goName == "Id" || goName == "ID") && propSchema.Type == "string" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("UUID")}}
		} else if propSchema.Type == "string" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Word")}}
		} else if propSchema.Type == "integer" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Number")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "1"}, &dst.BasicLit{Kind: token.INT, Value: "100"}}}
		} else if propSchema.Type == "number" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Float64Range")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.FLOAT, Value: "1.0"}, &dst.BasicLit{Kind: token.FLOAT, Value: "100.0"}}}
		} else if propSchema.Type == "boolean" {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Bool")}}
		} else {
			fakeExpr = &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Word")}}
		}

		compLit.Elts = append(compLit.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent(goName),
			Value: fakeExpr,
		})

		if target := resolveForeignKeyTarget(name, propName, allSchemas); target != "" && propSchema.Type == "string" {
			poolFieldSel := &dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(target + "IDs")}
			fkAssignStmts = append(fkAssignStmts, &dst.IfStmt{
				Cond: &dst.BinaryExpr{
					X:  &dst.CallExpr{Fun: dst.NewIdent("len"), Args: []dst.Expr{poolFieldSel}},
					Op: token.GTR,
					Y:  &dst.BasicLit{Kind: token.INT, Value: "0"},
				},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.AssignStmt{
							Lhs: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("item"), Sel: dst.NewIdent(goName)}},
							Tok: token.ASSIGN,
							Rhs: []dst.Expr{
								&dst.IndexExpr{
									X: &dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(target + "IDs")},
									Index: &dst.CallExpr{
										Fun: &dst.SelectorExpr{X: dst.NewIdent("gofakeit"), Sel: dst.NewIdent("Number")},
										Args: []dst.Expr{
											&dst.BasicLit{Kind: token.INT, Value: "0"},
											&dst.BinaryExpr{
												X:  &dst.CallExpr{Fun: dst.NewIdent("len"), Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("pool"), Sel: dst.NewIdent(target + "IDs")}}},
												Op: token.SUB,
												Y:  &dst.BasicLit{Kind: token.INT, Value: "1"},
											},
										},
									},
								},
							},
						},
					},
				},
			})
		}
	}

	if len(fkAssignStmts) > 0 {
		f.Body.List = append(f.Body.List, &dst.AssignStmt{
			Lhs: []dst.Expr{dst.NewIdent("item")},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: compLit}},
		})
		f.Body.List = append(f.Body.List, fkAssignStmts...)
		f.Body.List = append(f.Body.List, &dst.ReturnStmt{Results: []dst.Expr{dst.NewIdent("item")}})
	} else {
		f.Body.List = append(f.Body.List, &dst.ReturnStmt{Results: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: compLit}}})
	}
	return f
}
