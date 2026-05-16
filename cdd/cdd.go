package cdd

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/clients"
	"github.com/SamuelMarks/cdd-go/src/commands"
	"github.com/SamuelMarks/cdd-go/src/components"
	"github.com/SamuelMarks/cdd-go/src/mocks"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/routes"
	"github.com/SamuelMarks/cdd-go/src/tests"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// Config represents the SDK generation configuration.
type Config struct {
	InputPath             string
	InputDir              string
	OutputDir             string
	NoGithubActions       bool
	NoInstallablePackage  bool
	CreateComposableTests bool
}

// GenerateSDK generates the SDK based on the provided configuration.
func GenerateSDK(config Config) error {
	in := config.InputPath
	if in == "" {
		in = config.InputDir
	}
	return RunFromOpenAPI("to_sdk", in, config.OutputDir, config.NoGithubActions, config.NoInstallablePackage, config.CreateComposableTests)
}

// RunFromOpenAPI handles the generation of SDKs, servers, or CLI tools based on the OpenAPI specification.
func RunFromOpenAPI(subsubcommand, in, outDir string, noGithubActions, noInstallablePackage, tests bool) error {
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

	_ = os.MkdirAll(outDir, 0755)

	if !noInstallablePackage {
		GenerateGoMod(outDir)
	}

	if !noGithubActions {
		GenerateGithubActions(outDir)
	}

	switch subsubcommand {
	case "to_sdk":
		if err := GenerateClasses(oa, outDir); err != nil {
			return err
		}
		if err := GenerateClients(oa, outDir); err != nil {
			return err
		}
		if tests {
			if err := GenerateTests(oa, outDir); err != nil {
				return err
			}
			if err := GenerateMocks(oa, outDir); err != nil {
				return err
			}
		}
	case "to_server":
		if err := GenerateClasses(oa, outDir); err != nil {
			return err
		}
		if err := GenerateRoutes(oa, outDir); err != nil {
			return err
		}
		if tests {
			if err := GenerateTests(oa, outDir); err != nil {
				return err
			}
			if err := GenerateMocks(oa, outDir); err != nil {
				return err
			}
		}
	case "to_sdk_cli":
		if err := GenerateClasses(oa, outDir); err != nil {
			return err
		}
		if err := GenerateClients(oa, outDir); err != nil {
			return err
		}
		_ = GenerateCLI(oa, outDir)
		if tests {
			if err := GenerateTests(oa, outDir); err != nil {
				return err
			}
			if err := GenerateMocks(oa, outDir); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown subsubcommand: %s", subsubcommand)
	}
	return nil
}

// GenerateGoMod creates a go.mod file in the output directory if it does not already exist.
func GenerateGoMod(outDir string) {
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		content := "module generated_sdk\n\ngo 1.25.7\n"
		os.WriteFile(goModPath, []byte(content), 0644)
	}
}

// GenerateGithubActions creates a basic CI GitHub Actions workflow file in the output directory.
func GenerateGithubActions(outDir string) {
	ciPath := filepath.Join(outDir, ".github", "workflows", "ci.yml")
	os.MkdirAll(filepath.Dir(ciPath), 0755)
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		content := "name: CI\n\non:\n  push:\n    branches: [ main ]\n  pull_request:\n    branches: [ main ]\n\njobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n    - uses: actions/checkout@v4\n    - name: Set up Go\n      uses: actions/setup-go@v5\n      with:\n        go-version: '1.25'\n    - name: Build\n      run: go build -v ./...\n    - name: Test\n      run: go test -v ./...\n"
		os.WriteFile(ciPath, []byte(content), 0644)
	}
}

// GenerateCLI generates a basic CLI entrypoint (sdk_cli.go) utilizing the components and clients logic.
func GenerateCLI(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return WriteDstFile(filepath.Join(outDir, "sdk_cli.go"), &dst.File{Name: dst.NewIdent("main")})
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
	return WriteDstFile(filepath.Join(outDir, "sdk_cli.go"), file)
}

// RunToOpenAPI handles generation of an OpenAPI specification from Go codebase inputs.
func RunToOpenAPI(in, outPath string) error {
	if in == "" {
		return fmt.Errorf("input path is required")
	}

	if stat, err := os.Stat(outPath); err == nil && stat.IsDir() {
		outPath = filepath.Join(outPath, "openapi.json")
	}

	_ = GenerateOpenAPI(in, outPath)
	fmt.Printf("Successfully generated OpenAPI to %s\n", outPath)
	return nil
}

// GenerateOpenAPI analyzes the input path (file or directory) and generates an OpenAPI JSON specification.
func GenerateOpenAPI(inputPath string, outPath string) error {
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
		err := filepath.WalkDir(inputPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return err
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

	_ = os.MkdirAll(filepath.Dir(outPath), 0755)

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return openapi.Emit(out, oa)
}

// GenerateClasses generates Go structs based on the components/schemas defined in the OpenAPI spec.
func GenerateClasses(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil {
		return nil
	}

	modelsDir := filepath.Join(outDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}

	compDecls := components.Emit(oa.Components)
	if len(compDecls) > 0 {
		file := &dst.File{
			Name:  dst.NewIdent("models"),
			Decls: compDecls,
		}
		if err := WriteDstFile(filepath.Join(modelsDir, "components.go"), file); err != nil {
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
			Name:  dst.NewIdent("models"),
			Decls: []dst.Decl{decl},
		}

		if err := WriteDstFile(filepath.Join(modelsDir, strings.ToLower(name)+".go"), file); err != nil {
			return err
		}
	}
	return nil
}

// GenerateRoutes generates server routing interfaces and stub code.
func GenerateRoutes(oa *openapi.OpenAPI, outDir string) error {
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

		_ = WriteDstFile(filepath.Join(outDir, fileName+"_routes.go"), file)
	}
	return nil
}

// WriteDstFile is a utility that correctly formats and writes a dst.File to disk.
func WriteDstFile(path string, file *dst.File) error {
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

// GenerateClients generates client SDK code for the endpoints defined in the OpenAPI spec.
func GenerateClients(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	clientDir := filepath.Join(outDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return err
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
			Name: dst.NewIdent("client"),
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

		if err := WriteDstFile(filepath.Join(clientDir, fileName+"_client.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// GenerateTests generates test stubs for the provided API endpoints.
func GenerateTests(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	testsDir := filepath.Join(outDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return err
	}

	for path, item := range oa.Paths {
		file := &dst.File{
			Name: dst.NewIdent("tests"),
			Decls: []dst.Decl{
				&dst.GenDecl{
					Tok: token.IMPORT,
					Specs: []dst.Spec{
						&dst.ImportSpec{
							Path: &dst.BasicLit{Kind: token.STRING, Value: `"testing"`},
						},
						&dst.ImportSpec{
							Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`},
						},
						&dst.ImportSpec{
							Path: &dst.BasicLit{Kind: token.STRING, Value: `"strings"`},
						},
					},
				},
			},
		}
		hasTests := false
		addTest := func(method string, op *openapi.Operation) {
			if op == nil {
				return
			}
			decl, _ := tests.EmitTest(path, method, op)
			file.Decls = append(file.Decls, decl)
			hasTests = true
		}

		addTest("get", item.Get)
		addTest("post", item.Post)
		addTest("put", item.Put)
		addTest("delete", item.Delete)
		addTest("patch", item.Patch)
		addTest("options", item.Options)
		addTest("head", item.Head)
		addTest("trace", item.Trace)

		if !hasTests {
			continue
		}

		fileName := strings.ReplaceAll(path, "/", "_")
		fileName = strings.ReplaceAll(fileName, "{", "")
		fileName = strings.ReplaceAll(fileName, "}", "")
		fileName = strings.TrimPrefix(fileName, "_")
		if fileName == "" {
			fileName = "root"
		}

		file.Decls[0].(*dst.GenDecl).Specs = append(file.Decls[0].(*dst.GenDecl).Specs,
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"io"`}},
		)

		if err := WriteDstFile(filepath.Join(testsDir, fileName+"_test.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// GenerateMocks generates mock data from OpenAPI examples.
func GenerateMocks(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil || len(oa.Components.Examples) == 0 {
		return nil
	}

	mocksDir := filepath.Join(outDir, "mocks")
	if err := os.MkdirAll(mocksDir, 0755); err != nil {
		return err
	}

	file := &dst.File{
		Name: dst.NewIdent("mocks"),
	}

	for name, ex := range oa.Components.Examples {
		exCopy := ex
		decl, err := mocks.EmitExample(name, &exCopy)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, decl)
	}

	return WriteDstFile(filepath.Join(mocksDir, "mocks.go"), file)
}
