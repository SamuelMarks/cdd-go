// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"flag"
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/webhooks"
	"github.com/dave/dst"
)

// ServerCLIConfig holds parsed CLI flags and resolved DAO/database state for the generated server entrypoint.
type ServerCLIConfig struct {
	// StrictValidation enables strict OpenAPI request schema validation middleware.
	StrictValidation bool
	// EnforceAuth enables mock JWT/Bearer authentication middleware.
	EnforceAuth bool
	// StartAuthServer enables the integrated Identity Provider (IdP) authentication endpoints.
	StartAuthServer bool
	// Ephemeral triggers Concrete DAOs and overrides DATABASE_URL with a throwaway in-memory database.
	Ephemeral bool
	// Seed runs the fake data seeder on startup (requires a concrete DB connection).
	Seed bool
	// DatabaseURL is the configured database connection string from the environment.
	DatabaseURL string
	// UseConcreteDAOs indicates whether Concrete DAOs (true) or Stub DAOs (false) should be resolved.
	UseConcreteDAOs bool
}

// ParseServerFlags parses server CLI flags and resolves the orthogonal multi-tier DAO and database states.
// Parameter args is the CLI argument slice, dbURL is the DATABASE_URL environment value, and it returns the resolved ServerCLIConfig and any parse error.
func ParseServerFlags(args []string, dbURL string) (ServerCLIConfig, error) {
	var cfg ServerCLIConfig
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&cfg.StrictValidation, "strict-validation", false, "Enforce strict OpenAPI schema validation")
	fs.BoolVar(&cfg.EnforceAuth, "enforce-auth", false, "Enforce mock JWT/Bearer authentication")
	fs.BoolVar(&cfg.StartAuthServer, "start-auth-server", true, "Run the integrated Identity Provider (IdP) authentication endpoints")
	fs.BoolVar(&cfg.Ephemeral, "ephemeral", false, "Triggers the Concrete DAOs and overrides DATABASE_URL with a throwaway database")
	fs.BoolVar(&cfg.Seed, "seed", false, "Runs the fake data seeder on startup (requires a concrete DB connection)")
	if err := fs.Parse(args); err != nil {
		return ServerCLIConfig{}, err
	}
	cfg.DatabaseURL = dbURL
	cfg.UseConcreteDAOs = cfg.Ephemeral || strings.TrimSpace(dbURL) != ""
	return cfg, nil
}

// GenerateWebhookTriggers generates the administrative webhook trigger endpoint and HTTP dispatcher when webhooks or callbacks are defined.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateWebhookTriggers(oa *openapi.OpenAPI, outDir string) error {
	if !webhooks.HasWebhooksOrCallbacks(oa) {
		return nil
	}
	whDir := filepath.Join(outDir, "webhooks")
	if err := os.MkdirAll(whDir, 0755); err != nil {
		return err
	}

	reqTargetField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("TargetURL")},
		Type:  dst.NewIdent("string"),
		Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`" + `json:"target_url"` + "`"},
	}
	reqTargetField.Decs.Start.Append("// TargetURL is the registered target URL to dispatch the webhook payload to.")

	reqPayloadField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("Payload")},
		Type:  &dst.SelectorExpr{X: dst.NewIdent("json"), Sel: dst.NewIdent("RawMessage")},
		Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`" + `json:"payload,omitempty"` + "`"},
	}
	reqPayloadField.Decs.Start.Append("// Payload is the JSON body to dispatch to the target URL.")

	reqSpec := &dst.TypeSpec{
		Name: dst.NewIdent("TriggerRequest"),
		Type: &dst.StructType{Fields: &dst.FieldList{List: []*dst.Field{reqTargetField, reqPayloadField}}},
	}
	reqSpec.Decs.Start.Append("// TriggerRequest represents the request payload for triggering a mock webhook.")

	dispatchFn := &dst.FuncDecl{
		Name: dst.NewIdent("DispatchWebhook"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("targetURL")}, Type: dst.NewIdent("string")},
				{Names: []*dst.Ident{dst.NewIdent("payload")}, Type: dst.NewIdent("[]byte")},
			}},
			Results: &dst.FieldList{List: []*dst.Field{
				{Type: dst.NewIdent("int")},
				{Type: dst.NewIdent("error")},
			}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("client")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{Type: dst.NewIdent("http.Client")}}},
			},
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("resp"), dst.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.CallExpr{
					Fun: &dst.SelectorExpr{X: dst.NewIdent("client"), Sel: dst.NewIdent("Post")},
					Args: []dst.Expr{
						dst.NewIdent("targetURL"),
						&dst.BasicLit{Kind: token.STRING, Value: `"application/json"`},
						&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("bytes"), Sel: dst.NewIdent("NewReader")}, Args: []dst.Expr{dst.NewIdent("payload")}},
					},
				}},
			},
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{List: []dst.Stmt{
					&dst.ReturnStmt{Results: []dst.Expr{&dst.BasicLit{Kind: token.INT, Value: "0"}, dst.NewIdent("err")}},
				}},
			},
			&dst.DeferStmt{Call: &dst.CallExpr{Fun: &dst.SelectorExpr{X: &dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("Body")}, Sel: dst.NewIdent("Close")}}},
			&dst.ReturnStmt{Results: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("resp"), Sel: dst.NewIdent("StatusCode")}, dst.NewIdent("nil")}},
		}},
	}
	dispatchFn.Decs.Start.Append("// DispatchWebhook sends an outgoing webhook payload to targetURL using an HTTP client.")
	dispatchFn.Decs.Start.Append("// Parameter targetURL is the destination URL, payload is the raw JSON bytes, and it returns the HTTP status code and any error.")

	handleFn := &dst.FuncDecl{
		Name: dst.NewIdent("HandleTriggerWebhook"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{List: []*dst.Field{
				{Names: []*dst.Ident{dst.NewIdent("c")}, Type: &dst.StarExpr{X: dst.NewIdent("gin.Context")}},
			}},
		},
		Body: &dst.BlockStmt{List: []dst.Stmt{
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("webhookName")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("Param")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"webhook_name"`}}}},
			},
			&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("req")}, Type: dst.NewIdent("TriggerRequest")}}}},
			&dst.IfStmt{
				Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("ShouldBindJSON")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("req")}}}}},
				Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{List: []dst.Stmt{
					&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusBadRequest"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("err"), Sel: dst.NewIdent("Error")}, Args: []dst.Expr{}}}}}}}},
					&dst.ReturnStmt{},
				}},
			},
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("rawPayload"), dst.NewIdent("_")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("json"), Sel: dst.NewIdent("Marshal")}, Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("Payload")}}}},
			},
			&dst.AssignStmt{
				Lhs: []dst.Expr{dst.NewIdent("statusCode"), dst.NewIdent("err")},
				Tok: token.DEFINE,
				Rhs: []dst.Expr{&dst.CallExpr{Fun: dst.NewIdent("DispatchWebhook"), Args: []dst.Expr{&dst.SelectorExpr{X: dst.NewIdent("req"), Sel: dst.NewIdent("TargetURL")}, dst.NewIdent("rawPayload")}}},
			},
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")},
				Body: &dst.BlockStmt{List: []dst.Stmt{
					&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusBadGateway"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"error"`}, Value: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("err"), Sel: dst.NewIdent("Error")}, Args: []dst.Expr{}}}}}}}},
					&dst.ReturnStmt{},
				}},
			},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("c"), Sel: dst.NewIdent("JSON")}, Args: []dst.Expr{dst.NewIdent("http.StatusOK"), &dst.CompositeLit{Type: dst.NewIdent("gin.H"), Elts: []dst.Expr{
				&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"webhook"`}, Value: dst.NewIdent("webhookName")},
				&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"status_code"`}, Value: dst.NewIdent("statusCode")},
			}}}}},
		}},
	}
	handleFn.Decs.Start.Append("// HandleTriggerWebhook handles POST /_mock/trigger-webhook/:webhook_name administrative trigger requests.")
	handleFn.Decs.Start.Append("// Parameter c is the Gin HTTP context.")

	file := &dst.File{
		Name: dst.NewIdent("webhooks"),
		Decls: []dst.Decl{
			&dst.GenDecl{Tok: token.IMPORT, Specs: []dst.Spec{
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"bytes"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
				&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`}},
			}},
			&dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{reqSpec}},
			dispatchFn,
			handleFn,
		},
	}
	file.Decs.Start.Append("// Package webhooks provides administrative webhook trigger handlers and HTTP dispatchers.")

	return WriteDstFile(filepath.Join(whDir, "trigger.go"), file)
}

// GenerateServerReadme generates a README.md in the server project root documenting server modes and webhook triggers.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any write error.
func GenerateServerReadme(oa *openapi.OpenAPI, outDir string) error {
	title := "Generated CDD Mock & Production Server"
	if oa != nil && oa.Info.Title != "" {
		title = oa.Info.Title + " — CDD Server"
	}

	lines := []string{
		"# " + title,
		"",
		"## Decoupled CDD Server Modes",
		"",
		"The generated server supports four orthogonal runtime modes controlled by environment variables and CLI flags:",
		"",
		"- `start` (No DB configured): **Stub Mode**. Server runs using traditional scaffolds, endpoints return `NotImplementedError` (`501 Not Implemented`) or empty bodies.",
		"- `start` (With `DATABASE_URL`): **Production Mode**. Uses actual ORM interactions against a real PostgreSQL database.",
		"- `start --ephemeral`: **Sandbox Mode**. Uses actual ORM interactions against a fresh, throwaway in-memory SQLite database.",
		"- `start --ephemeral --seed`: **Full Mock Mode**. Ephemeral database, automatically populated with a localized fake data graph on startup.",
		"",
		"### CLI Flags",
		"",
		"- `--ephemeral`: Triggers the Concrete DAOs and overrides `DATABASE_URL` with a throwaway database.",
		"- `--seed`: Runs the fake data seeder on startup (requires a concrete DB connection).",
		"- `--strict-validation`: Enforce strict OpenAPI schema validation.",
		"- `--enforce-auth`: Enforce mock JWT/Bearer authentication.",
		"- `--start-auth-server`: Run the integrated Identity Provider (IdP) authentication endpoints (`/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`).",
		"",
		"## Administrative Webhook & Callback Trigger API",
		"",
		"When the OpenAPI specification defines `webhooks` or `callbacks`, the server exposes an administrative trigger endpoint:",
		"",
		"- `POST /_mock/trigger-webhook/{webhook_name}`",
		"",
		"Request body JSON:",
		"",
		"```json",
		"{",
		`  "target_url": "http://localhost:9000/webhook-receiver",`,
		`  "payload": {`,
		`    "event": "example"`,
		"  }",
		"}",
		"```",
	}

	if names := webhooks.CollectWebhookNames(oa); len(names) > 0 {
		lines = append(lines, "", "### Registered Webhooks & Callbacks", "")
		for _, name := range names {
			lines = append(lines, fmt.Sprintf("- `%s` (`POST /_mock/trigger-webhook/%s`)", name, name))
		}
	}
	lines = append(lines, "")

	return os.WriteFile(filepath.Join(outDir, "README.md"), []byte(strings.Join(lines, string(rune(10)))), 0644)
}

// GenerateServerMain generates the main entrypoint for the server.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, database enables DB setup, seed enables DB seeding, and idp enables mock IdP endpoints.
func GenerateServerMain(oa *openapi.OpenAPI, outDir string, database bool, seed bool, idp bool) error {
	mainDir := filepath.Join(outDir, "cmd", "server")
	if err := os.MkdirAll(mainDir, 0755); err != nil {
		return err
	}

	hasWebhooks := webhooks.HasWebhooksOrCallbacks(oa)

	bodyStmts := []dst.Stmt{
		&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("strictValidation")}, Type: dst.NewIdent("bool")}}}},
		&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("enforceAuth")}, Type: dst.NewIdent("bool")}}}},
		&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("startAuthServer")}, Type: dst.NewIdent("bool")}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("strictValidation")}, &dst.BasicLit{Kind: token.STRING, Value: `"strict-validation"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Enforce strict OpenAPI schema validation"`}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("enforceAuth")}, &dst.BasicLit{Kind: token.STRING, Value: `"enforce-auth"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Enforce mock JWT/Bearer authentication"`}}}},
		&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("startAuthServer")}, &dst.BasicLit{Kind: token.STRING, Value: `"start-auth-server"`}, dst.NewIdent("true"), &dst.BasicLit{Kind: token.STRING, Value: `"Run the integrated Identity Provider (IdP) authentication endpoints"`}}}},
	}

	if database {
		bodyStmts = append(bodyStmts,
			&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("ephemeral")}, Type: dst.NewIdent("bool")}}}},
			&dst.DeclStmt{Decl: &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("seedFlag")}, Type: dst.NewIdent("bool")}}}},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("ephemeral")}, &dst.BasicLit{Kind: token.STRING, Value: `"ephemeral"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Triggers the Concrete DAOs and overrides DATABASE_URL with a throwaway database"`}}}},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("flag"), Sel: dst.NewIdent("BoolVar")}, Args: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("seedFlag")}, &dst.BasicLit{Kind: token.STRING, Value: `"seed"`}, dst.NewIdent("false"), &dst.BasicLit{Kind: token.STRING, Value: `"Runs the fake data seeder on startup (requires a concrete DB connection)"`}}}},
		)
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
			&dst.IfStmt{
				Cond: dst.NewIdent("startAuthServer"),
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/register"`}, dst.NewIdent("idp.HandleRegister")}}},
						&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/login"`}, dst.NewIdent("idp.HandleLogin")}}},
						&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/refresh"`}, dst.NewIdent("idp.HandleRefresh")}}},
						&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/auth/logout"`}, dst.NewIdent("idp.HandleLogout")}}},
					},
				},
			},
		)
	} else {
		bodyStmts = append(bodyStmts,
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("_")}, Tok: token.ASSIGN, Rhs: []dst.Expr{dst.NewIdent("startAuthServer")}},
		)
	}

	if hasWebhooks {
		bodyStmts = append(bodyStmts,
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("r"), Sel: dst.NewIdent("POST")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"/_mock/trigger-webhook/:webhook_name"`}, dst.NewIdent("webhooks.HandleTriggerWebhook")}}},
		)
	}

	if database {
		bodyStmts = append(bodyStmts,
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("dbUrl")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("os"), Sel: dst.NewIdent("Getenv")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"DATABASE_URL"`}}}}},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("useConcreteDAOs")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.BinaryExpr{X: dst.NewIdent("ephemeral"), Op: token.LOR, Y: &dst.BinaryExpr{X: dst.NewIdent("dbUrl"), Op: token.NEQ, Y: &dst.BasicLit{Kind: token.STRING, Value: `""`}}}}},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("cfg")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CompositeLit{Type: dst.NewIdent("database.Config"), Elts: []dst.Expr{&dst.KeyValueExpr{Key: dst.NewIdent("DatabaseURL"), Value: dst.NewIdent("dbUrl")}, &dst.KeyValueExpr{Key: dst.NewIdent("Ephemeral"), Value: dst.NewIdent("ephemeral")}}}}},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("db"), dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("database"), Sel: dst.NewIdent("InitDB")}, Args: []dst.Expr{dst.NewIdent("cfg")}}}},
			&dst.IfStmt{Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Failed to init DB: %v"`}, dst.NewIdent("err")}}}}}},
		)

		dbBlockStmts := []dst.Stmt{
			&dst.IfStmt{Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("database"), Sel: dst.NewIdent("Migrate")}, Args: []dst.Expr{dst.NewIdent("db")}}}}, Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Migration failed: %v"`}, dst.NewIdent("err")}}}}}},
		}

		if seed {
			dbBlockStmts = append(dbBlockStmts,
				&dst.IfStmt{Cond: dst.NewIdent("seedFlag"), Body: &dst.BlockStmt{List: []dst.Stmt{&dst.IfStmt{Init: &dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("err")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("seeder"), Sel: dst.NewIdent("SeedDatabase")}, Args: []dst.Expr{dst.NewIdent("db")}}}}, Cond: &dst.BinaryExpr{X: dst.NewIdent("err"), Op: token.NEQ, Y: dst.NewIdent("nil")}, Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Fatalf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Seeding failed: %v"`}, dst.NewIdent("err")}}}}}}}}},
			)
		} else {
			dbBlockStmts = append(dbBlockStmts,
				&dst.IfStmt{Cond: dst.NewIdent("seedFlag"), Body: &dst.BlockStmt{List: []dst.Stmt{&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Println")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Seeder flag set"`}}}}}}},
			)
		}

		bodyStmts = append(bodyStmts,
			&dst.IfStmt{
				Cond: &dst.BinaryExpr{X: dst.NewIdent("useConcreteDAOs"), Op: token.LAND, Y: &dst.BinaryExpr{X: dst.NewIdent("db"), Op: token.NEQ, Y: dst.NewIdent("nil")}},
				Body: &dst.BlockStmt{List: dbBlockStmts},
			},
			&dst.AssignStmt{Lhs: []dst.Expr{dst.NewIdent("daoFactory")}, Tok: token.DEFINE, Rhs: []dst.Expr{&dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("daos"), Sel: dst.NewIdent("NewDAOFactory")}, Args: []dst.Expr{dst.NewIdent("db"), dst.NewIdent("ephemeral")}}}},
			&dst.ExprStmt{X: &dst.CallExpr{Fun: &dst.SelectorExpr{X: dst.NewIdent("log"), Sel: dst.NewIdent("Printf")}, Args: []dst.Expr{&dst.BasicLit{Kind: token.STRING, Value: `"Starting server... DAO Factory initialized: %v"`}, dst.NewIdent("daoFactory")}}},
		)
	}

	if oa != nil && len(oa.Paths) > 0 {
		pathKeys := make([]string, 0, len(oa.Paths))
		for path := range oa.Paths {
			pathKeys = append(pathKeys, path)
		}
		sort.Strings(pathKeys)

		daoArgName := "nil"
		if database {
			daoArgName = "daoFactory"
		}

		for _, path := range pathKeys {
			newHandlerName := "NewHandler" + naming.ToCamelCasePath(path)
			bodyStmts = append(bodyStmts, &dst.ExprStmt{
				X: &dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X: &dst.CallExpr{
							Fun:  &dst.SelectorExpr{X: dst.NewIdent("routes"), Sel: dst.NewIdent(newHandlerName)},
							Args: []dst.Expr{dst.NewIdent(daoArgName)},
						},
						Sel: dst.NewIdent("Register"),
					},
					Args: []dst.Expr{dst.NewIdent("r")},
				},
			})
		}
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

	if oa != nil && len(oa.Paths) > 0 {
		imports = append(imports, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/routes"`}})
	}

	if idp {
		imports = append(imports, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/idp"`}})
	}

	if hasWebhooks {
		imports = append(imports, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/webhooks"`}})
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
	file.Decs.Start.Append("// Package main provides the generated HTTP server entrypoint.")

	return WriteDstFile(filepath.Join(mainDir, "main.go"), file)
}
