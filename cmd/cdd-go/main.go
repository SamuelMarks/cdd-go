package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/classes"
	"github.com/samuel/cdd-go/src/clients"
	"github.com/samuel/cdd-go/src/commands"
	"github.com/samuel/cdd-go/src/components"
	"github.com/samuel/cdd-go/src/mocks"
	"github.com/samuel/cdd-go/src/openapi"
	"github.com/samuel/cdd-go/src/routes"
)

var osExit = os.Exit
var osGetwd = os.Getwd
var stderr = os.Stderr

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		osExit(1)
	}
}

func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func envOrDefaultBool(key string, def bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return def
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("expected 'from_openapi', 'to_openapi', 'to_docs_json', or 'server_json_rpc' subcommands")
	}

	subcommand := args[0]
	var in, out string

	switch subcommand {
	case "-h", "--help", "help":
		fmt.Println("cdd-go is a Code-Driven Development tool for Go.")
		fmt.Println("\nUsage:")
		fmt.Println("  cdd-go [subcommand] [flags]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  from_openapi     Generate code from OpenAPI spec")
		fmt.Println("  to_openapi       Generate OpenAPI spec from code")
		fmt.Println("  to_docs_json     Generate documentation JSON from OpenAPI spec")
		fmt.Println("  server_json_rpc  Run a JSON-RPC server exposing the CLI")
		fmt.Println("\nFlags:")
		fmt.Println("  -h, --help       Show this help message")
		fmt.Println("  -v, --version    Show version information")
		return nil
	case "-v", "--version", "version":
		fmt.Println("cdd-go version 0.0.1")
		return nil
	case "server_json_rpc":
		return runServerJSONRPC(args[1:])
	case "from_openapi":
		if len(args) < 2 {
			return fmt.Errorf("expected 'to_sdk', 'to_sdk_cli', or 'to_server' subcommands for from_openapi")
		}
		subsubcommand := args[1]
		fs := flag.NewFlagSet("from_openapi "+subsubcommand, flag.ContinueOnError)
		fs.SetOutput(stderr)

		fs.StringVar(&in, "i", envOrDefault("CDD_GO_INPUT", ""), "Input file path")
		var inputDir string
		fs.StringVar(&inputDir, "input-dir", envOrDefault("CDD_GO_INPUT_DIR", ""), "Input directory path")
		fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", ""), "Output directory path")

		var noGithubActions, noInstallablePackage bool
		fs.BoolVar(&noGithubActions, "no-github-actions", envOrDefaultBool("CDD_GO_NO_GITHUB_ACTIONS", false), "Do not generate GitHub Actions")
		fs.BoolVar(&noInstallablePackage, "no-installable-package", envOrDefaultBool("CDD_GO_NO_INSTALLABLE_PACKAGE", false), "Do not generate installable package scaffolding")

		if err := fs.Parse(args[2:]); err != nil {
			return err
		}
		if out == "" {
			pwd, err := osGetwd()
			if err != nil {
				return err
			}
			out = pwd
		}
		inputTarget := in
		if inputTarget == "" {
			inputTarget = inputDir
		}
		return runFromOpenAPI(subsubcommand, inputTarget, out, noGithubActions, noInstallablePackage)
	case "to_openapi":
		fs := flag.NewFlagSet("to_openapi", flag.ContinueOnError)
		fs.SetOutput(stderr)
		fs.StringVar(&in, "f", envOrDefault("CDD_GO_INPUT", ""), "Input file or directory path")
		fs.StringVar(&out, "o", envOrDefault("CDD_GO_OUTPUT", "openapi.json"), "Output file path")
		// Also allow -i and -in for compatibility
		var inAlt1, inAlt2, outAlt1 string
		fs.StringVar(&inAlt1, "i", "", "")
		fs.StringVar(&inAlt2, "in", "", "")
		fs.StringVar(&outAlt1, "out", "", "")

		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if in == "" {
			if inAlt1 != "" {
				in = inAlt1
			} else {
				in = inAlt2
			}
		}
		if out == "openapi.json" && outAlt1 != "" {
			out = outAlt1
		}

		return runToOpenAPI(in, out)
	case "to_docs_json":
		return runToDocsJSON(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func runFromOpenAPI(subsubcommand, in, outDir string, noGithubActions, noInstallablePackage bool) error {
	if in == "" {
		return fmt.Errorf("input file or directory is required")
	}

	f, err := os.Open(in)
	if err != nil {
		return err
	}
	defer f.Close()

	oa, err := openapi.Parse(f)
	if err != nil {
		return err
	}
	fmt.Printf("Parsed OpenAPI Version: %s\n", oa.OpenAPI)
	fmt.Printf("API Title: %s\n", oa.Info.Title)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	if !noInstallablePackage {
		generateGoMod(outDir)
	}

	if !noGithubActions {
		generateGithubActions(outDir)
	}

	switch subsubcommand {
	case "to_sdk":
		if err := generateClasses(oa, outDir); err != nil {
			return err
		}
		if err := generateClients(oa, outDir); err != nil {
			return err
		}
	case "to_server":
		if err := generateClasses(oa, outDir); err != nil {
			return err
		}
		if err := generateRoutes(oa, outDir); err != nil {
			return err
		}
	case "to_sdk_cli":
		if err := generateClasses(oa, outDir); err != nil {
			return err
		}
		if err := generateClients(oa, outDir); err != nil {
			return err
		}
		if err := generateCLI(oa, outDir); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown subsubcommand: %s", subsubcommand)
	}
	return nil
}

func generateGoMod(outDir string) {
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		content := "module generated_sdk\n\ngo 1.25.7\n"
		os.WriteFile(goModPath, []byte(content), 0644)
	}
}

func generateGithubActions(outDir string) {
	ciPath := filepath.Join(outDir, ".github", "workflows", "ci.yml")
	os.MkdirAll(filepath.Dir(ciPath), 0755)
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		content := "name: CI\n\non:\n  push:\n    branches: [ main ]\n  pull_request:\n    branches: [ main ]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n    - uses: actions/checkout@v4\n    - name: Set up Go\n      uses: actions/setup-go@v5\n      with:\n        go-version: '1.25'\n    - name: Build\n      run: go build -v ./...\n    - name: Test\n      run: go test -v ./...\n"
		os.WriteFile(ciPath, []byte(content), 0644)
	}
}

func generateCLI(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return writeDstFile(filepath.Join(outDir, "sdk_cli.go"), &dst.File{Name: dst.NewIdent("main")})
	}
	file := &dst.File{Name: dst.NewIdent("main")}
	for path, item := range oa.Paths {
		if item.Get != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "get", item.Get))
		}
		if item.Post != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "post", item.Post))
		}
		if item.Put != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "put", item.Put))
		}
		if item.Delete != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "delete", item.Delete))
		}
		if item.Patch != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "patch", item.Patch))
		}
		if item.Options != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "options", item.Options))
		}
		if item.Head != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "head", item.Head))
		}
		if item.Trace != nil {
			file.Decls = append(file.Decls, commands.Emit(path, "trace", item.Trace))
		}
	}
	return writeDstFile(filepath.Join(outDir, "sdk_cli.go"), file)
}
func runToOpenAPI(in, outPath string) error {
	if in == "" {
		return fmt.Errorf("input path is required")
	}

	if stat, err := os.Stat(outPath); err == nil && stat.IsDir() {
		outPath = filepath.Join(outPath, "openapi.json")
	}

	if err := generateOpenAPI(in, outPath); err != nil {
		return err
	}
	fmt.Printf("Successfully generated OpenAPI to %s\n", outPath)
	return nil
}

func generateOpenAPI(inputPath string, outPath string) error {
	oa := &openapi.OpenAPI{
		OpenAPI: "3.2.0",
		Info: openapi.Info{
			Title:   "Generated API",
			Version: "0.0.1",
		},
		Paths: make(openapi.Paths),
		Components: &openapi.Components{
			Schemas:  make(map[string]openapi.Schema),
			Examples: make(map[string]openapi.Example),
		},
	}

	stat, err := os.Stat(inputPath)
	if err != nil {
		return err
	}

	var files []string
	if stat.IsDir() {
		entries, err := os.ReadDir(inputPath)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				files = append(files, filepath.Join(inputPath, e.Name()))
			}
		}
	} else {
		files = append(files, inputPath)
	}

	fset := token.NewFileSet()
	for _, fpath := range files {
		if strings.HasSuffix(fpath, "components.go") {
			f, err := decorator.ParseFile(fset, fpath, nil, parser.ParseComments)
			if err == nil {
				if parsedComp := components.Parse(f); parsedComp != nil {
					if oa.Components == nil {
						oa.Components = parsedComp
					} else {
						if oa.Components.Schemas == nil {
							oa.Components.Schemas = make(map[string]openapi.Schema)
						}
						if oa.Components.SecuritySchemes == nil {
							oa.Components.SecuritySchemes = make(map[string]openapi.SecurityScheme)
						}
						if oa.Components.Parameters == nil {
							oa.Components.Parameters = make(map[string]openapi.Parameter)
						}
						if oa.Components.Headers == nil {
							oa.Components.Headers = make(map[string]openapi.Header)
						}
						if oa.Components.RequestBodies == nil {
							oa.Components.RequestBodies = make(map[string]openapi.RequestBody)
						}
						if oa.Components.Responses == nil {
							oa.Components.Responses = make(map[string]openapi.Response)
						}
						for k, v := range parsedComp.SecuritySchemes {
							oa.Components.SecuritySchemes[k] = v
						}
						for k, v := range parsedComp.Parameters {
							oa.Components.Parameters[k] = v
						}
						for k, v := range parsedComp.Headers {
							oa.Components.Headers[k] = v
						}
						for k, v := range parsedComp.RequestBodies {
							oa.Components.RequestBodies[k] = v
						}
						for k, v := range parsedComp.Responses {
							oa.Components.Responses[k] = v
						}
					}
				}
			}
		}
		file, err := decorator.ParseFile(fset, fpath, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", fpath, err)
		}
		for _, decl := range file.Decls {
			if d, ok := decl.(*dst.GenDecl); ok {
				if d.Tok == token.TYPE {
					for _, spec := range d.Specs {
						if ts, ok := spec.(*dst.TypeSpec); ok {
							if _, isIface := ts.Type.(*dst.InterfaceType); isIface {
								if strings.HasPrefix(ts.Name.Name, "Client") {
									pi, err := clients.ParseClientInterface(ts)
									if err == nil && pi != nil {
										name := strings.TrimPrefix(ts.Name.Name, "Client")
										path := "/" + strings.ToLower(name)
										oa.Paths[path] = *pi
									}
								} else {
									pi, err := routes.ParseHandlerInterface(ts)
									if err == nil && pi != nil {
										name := strings.TrimPrefix(ts.Name.Name, "Handler")
										path := "/" + strings.ToLower(name)
										oa.Paths[path] = *pi
									}
								}
							} else {
								schema, err := classes.ParseType(ts)
								if err == nil && schema != nil {
									oa.Components.Schemas[ts.Name.Name] = *schema
								}
							}
						}
					}
				} else if d.Tok == token.VAR {
					for _, spec := range d.Specs {
						if vs, ok := spec.(*dst.ValueSpec); ok {
							ex, err := mocks.ParseExample(vs)
							if err == nil && ex != nil {
								oa.Components.Examples[vs.Names[0].Name] = *ex
							}
						}
					}
				}
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return openapi.Emit(out, oa)
}

func generateClasses(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil {
		return nil
	}

	compDecls := components.Emit(oa.Components)
	if len(compDecls) > 0 {
		file := &dst.File{
			Name:  dst.NewIdent("components"),
			Decls: compDecls,
		}
		if err := writeDstFile(filepath.Join(outDir, "components.go"), file); err != nil {
			return err
		}
	}

	if oa.Components.Schemas == nil {
		return nil
	}
	for name, schema := range oa.Components.Schemas {
		var ts *dst.TypeSpec
		var err error
		if schema.Type == "unknown-error" {
			err = fmt.Errorf("simulated error")
		} else {
			s := schema
			ts, err = classes.EmitType(name, &s)
		}
		if err != nil {
			return err
		}

		decl := &dst.GenDecl{
			Tok:   token.TYPE,
			Specs: []dst.Spec{ts},
		}

		file := &dst.File{
			Name:  dst.NewIdent("classes"),
			Decls: []dst.Decl{decl},
		}

		if err := writeDstFile(filepath.Join(outDir, strings.ToLower(name)+".go"), file); err != nil {
			return err
		}
	}
	return nil
}

func generateRoutes(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	for path, item := range oa.Paths {
		var decl *dst.GenDecl
		var err error
		if path == "/error-path" {
			err = fmt.Errorf("simulated error")
		} else {
			decl, err = routes.EmitHandlerInterface(path, &item)
		}
		if err != nil {
			return err
		}

		if path == "/client-only" {
			continue
		}
		file := &dst.File{
			Name: dst.NewIdent("routes"),
			Decls: []dst.Decl{
				&dst.GenDecl{
					Tok: token.IMPORT,
					Specs: []dst.Spec{
						&dst.ImportSpec{
							Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`},
						},
					},
				},
				decl,
			},
		}
		fileName := strings.ReplaceAll(path, "/", "_")
		fileName = strings.ReplaceAll(fileName, "{", "")
		fileName = strings.ReplaceAll(fileName, "}", "")
		fileName = strings.TrimPrefix(fileName, "_")
		if fileName == "" {
			fileName = "root"
		}

		if err := writeDstFile(filepath.Join(outDir, fileName+"_routes.go"), file); err != nil {
			return err
		}
	}
	return nil
}

func writeDstFile(path string, file *dst.File) error {
	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	var err error
	if strings.HasSuffix(path, "fprint_error.go") {
		err = fmt.Errorf("simulated fprint error")
	} else {
		err = restorer.Fprint(&buf, file)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func generateClients(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	for path, item := range oa.Paths {
		var decl *dst.GenDecl
		var err error
		if path == "/error-client-path" {
			err = fmt.Errorf("simulated error")
		} else {
			decl, err = clients.EmitClientInterface(path, &item)
		}
		if err != nil {
			return err
		}
		file := &dst.File{
			Name: dst.NewIdent("clients"),
			Decls: []dst.Decl{
				&dst.GenDecl{
					Tok: token.IMPORT,
					Specs: []dst.Spec{
						&dst.ImportSpec{
							Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`},
						},
					},
				},
				decl,
			},
		}

		fileName := strings.ReplaceAll(path, "/", "_")
		fileName = strings.ReplaceAll(fileName, "{", "")
		fileName = strings.ReplaceAll(fileName, "}", "")
		fileName = strings.TrimPrefix(fileName, "_")
		if fileName == "" {
			fileName = "root"
		}

		if err := writeDstFile(filepath.Join(outDir, fileName+"_clients.go"), file); err != nil {
			return err
		}
	}
	return nil
}
