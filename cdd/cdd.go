// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/callbacks"
	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/clients"
	"github.com/SamuelMarks/cdd-go/src/commands"
	"github.com/SamuelMarks/cdd-go/src/components"
	"github.com/SamuelMarks/cdd-go/src/externaldocs"
	"github.com/SamuelMarks/cdd-go/src/info"
	"github.com/SamuelMarks/cdd-go/src/links"
	"github.com/SamuelMarks/cdd-go/src/mocks"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/routes"
	"github.com/SamuelMarks/cdd-go/src/servers"
	"github.com/SamuelMarks/cdd-go/src/tags"
	"github.com/SamuelMarks/cdd-go/src/tests"
	"github.com/SamuelMarks/cdd-go/src/webhooks"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// Config represents the SDK generation configuration.
type Config struct {
	// InputPath is the path to a single OpenAPI specification file.
	InputPath string
	// InputDir is the directory containing OpenAPI specifications.
	InputDir string
	// OutputDir is the target directory for generated code.
	OutputDir string
	// NoGithubActions disables generating GitHub Actions CI workflows.
	NoGithubActions bool
	// NoInstallablePackage disables generating go.mod scaffolding.
	NoInstallablePackage bool
	// CreateComposableTests enables generating test suites and mocks.
	CreateComposableTests bool
	// Mcp enables generating Model Context Protocol (MCP) bindings.
	Mcp bool
}

// GenerateSDK generates the SDK based on the provided configuration.
// Parameter config holds the generation options, and it returns any generation error.
func GenerateSDK(config Config) error {
	in := config.InputPath
	if in == "" {
		in = config.InputDir
	}
	return FromOpenAPI("to_sdk", in, config.OutputDir, config.NoGithubActions, config.NoInstallablePackage, config.CreateComposableTests, config.Mcp)
}

// FromOpenAPI generates code from an OpenAPI specification.
// Parameter subsubcommand selects the target ("to_sdk", "to_server", "to_sdk_cli"), in is the spec path, outDir is the destination directory, noGithubActions omits CI workflow files, noInstallablePackage omits go.mod, tests enables test generation, and mcp enables MCP bindings.
func FromOpenAPI(subsubcommand, in, outDir string, noGithubActions, noInstallablePackage, tests, mcp bool) error {
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
	fmt.Println("Parsed OpenAPI Version:", oa.OpenAPI)
	fmt.Println("API Title:", oa.Info.Title)

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
		if err := GenerateMiddlewares(outDir); err != nil {
			return err
		}
		if err := GenerateIdP(outDir); err != nil {
			return err
		}
		if err := GenerateDatabase(oa, outDir); err != nil {
			return err
		}
		if err := GenerateDAOs(oa, outDir); err != nil {
			return err
		}
		if err := GenerateRoutes(oa, outDir); err != nil {
			return err
		}
		if err := GenerateServerMain(oa, outDir, true, tests, true); err != nil {
			return err
		}
		if tests {
			if err := GenerateTests(oa, outDir); err != nil {
				return err
			}
			if err := GenerateMocks(oa, outDir); err != nil {
				return err
			}
			if err := GenerateSeeder(oa, outDir); err != nil {
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
		if err := GenerateCLI(oa, outDir); err != nil {
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
	default:
		return fmt.Errorf("unknown subsubcommand: %s", subsubcommand)
	}
	return nil
}

// GenerateGoMod creates a go.mod file in the output directory if it does not already exist.
// Parameter outDir is the destination directory path.
func GenerateGoMod(outDir string) {
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		content := `module generated_sdk

go 1.25.7
`
		_ = os.WriteFile(goModPath, []byte(content), 0644)
	}
}

// GenerateGithubActions creates a basic CI GitHub Actions workflow file in the output directory.
// Parameter outDir is the destination directory path.
func GenerateGithubActions(outDir string) {
	ciPath := filepath.Join(outDir, ".github", "workflows", "ci.yml")
	_ = os.MkdirAll(filepath.Dir(ciPath), 0755)
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		content := `name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v6
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    - name: Build
      run: go build -v ./...
    - name: Test
      run: go test -v ./...
`
		_ = os.WriteFile(ciPath, []byte(content), 0644)
	}
}

// GenerateCLI generates a basic CLI entrypoint (sdk_cli.go) utilizing the components and clients logic.
// Parameter oa is the parsed OpenAPI specification, outDir is the output directory, and it returns any write error.
func GenerateCLI(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return WriteDstFile(filepath.Join(outDir, "sdk_cli.go"), &dst.File{Name: dst.NewIdent("main")})
	}
	file := &dst.File{Name: dst.NewIdent("main")}
	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := oa.Paths[path]
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
	file.Decls = append(file.Decls, commands.EmitMcpCmd())
	return WriteDstFile(filepath.Join(outDir, "sdk_cli.go"), file)
}

// ToOpenAPI generates an OpenAPI specification from source code.
// Parameter in is the source file or directory path, outPath is the output JSON path, and it returns any generation error.
func ToOpenAPI(in, outPath string) error {
	if in == "" {
		return fmt.Errorf("input path is required")
	}

	if stat, err := os.Stat(outPath); err == nil && stat.IsDir() {
		outPath = filepath.Join(outPath, "openapi.json")
	}

	if err := GenerateOpenAPI(in, outPath); err != nil {
		return err
	}
	fmt.Println("Successfully generated OpenAPI to", outPath)
	return nil
}

// extractPathComment extracts the "// Path: <path>" annotation from a TypeSpec or GenDecl if present.
// Parameter decs is the decoration slice, fallbackName is the stripped interface name, and it returns the reconstructed URL path.
func extractPathComment(decs dst.Decorations, fallbackName string) string {
	for _, doc := range decs {
		line := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
		if strings.HasPrefix(line, "Path:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Path:"))
		}
	}
	return "/" + strings.ToLower(fallbackName)
}

// GenerateOpenAPI analyzes the input path (file or directory) and generates an OpenAPI JSON specification.
// Parameter inputPath is the Go file or directory to analyze, outPath is the output OpenAPI file path, and it returns any error.
func GenerateOpenAPI(inputPath string, outPath string) error {
	oa := &openapi.OpenAPI{
		OpenAPI: "3.2.0",
		Info: openapi.Info{
			Title:   "Generated API",
			Version: "0.0.4",
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
								allDecs := append(dst.Decorations{}, d.Decs.Start...)
								allDecs = append(allDecs, ts.Decs.Start...)
								if strings.HasPrefix(ts.Name.Name, "Client") {
									pi, err := clients.ParseClientInterface(ts)
									if err == nil && pi != nil {
										name := strings.TrimPrefix(ts.Name.Name, "Client")
										path := extractPathComment(allDecs, name)
										oa.Paths[path] = *pi
									}
								} else {
									pi, err := routes.ParseHandlerInterface(ts)
									if err == nil && pi != nil {
										name := strings.TrimPrefix(ts.Name.Name, "Handler")
										path := extractPathComment(allDecs, name)
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
				} else if d.Tok == token.VAR || d.Tok == token.CONST {
					if parsedInfo, err := info.Parse(d); err == nil && parsedInfo != nil {
						oa.Info = *parsedInfo
						continue
					}
					if parsedServers, err := servers.Parse(d); err == nil && parsedServers != nil {
						oa.Servers = parsedServers
						continue
					}
					if parsedTags, err := tags.Parse(d); err == nil && parsedTags != nil {
						oa.Tags = parsedTags
						continue
					}
					if parsedExtDocs, err := externaldocs.Parse(d); err == nil && parsedExtDocs != nil {
						oa.ExternalDocs = parsedExtDocs
						continue
					}
					if whName, parsedWh := webhooks.Parse(d); parsedWh != nil && whName != "" {
						if oa.Webhooks == nil {
							oa.Webhooks = make(map[string]openapi.PathItem)
						}
						for k, v := range parsedWh {
							oa.Webhooks[k] = v
						}
						continue
					}
					if cbName, parsedCb := callbacks.Parse(d); parsedCb != nil && cbName != "" {
						if oa.Components.Callbacks == nil {
							oa.Components.Callbacks = make(map[string]openapi.Callback)
						}
						oa.Components.Callbacks[cbName] = *parsedCb
						continue
					}
					if linkName, parsedLink := links.Parse(d); parsedLink != nil && linkName != "" {
						if oa.Components.Links == nil {
							oa.Components.Links = make(map[string]openapi.Link)
						}
						oa.Components.Links[linkName] = *parsedLink
						continue
					}

					method, path, op := commands.Parse(d)
					if op != nil && method != "" && path != "" {
						if _, ok := oa.Paths[path]; !ok {
							oa.Paths[path] = openapi.PathItem{}
						}
						pi := oa.Paths[path]
						switch strings.ToLower(method) {
						case "get":
							pi.Get = op
						case "post":
							pi.Post = op
						case "put":
							pi.Put = op
						case "delete":
							pi.Delete = op
						case "patch":
							pi.Patch = op
						case "options":
							pi.Options = op
						case "head":
							pi.Head = op
						case "trace":
							pi.Trace = op
						}
						oa.Paths[path] = pi
						continue
					}

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
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
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
	schemaNames := make([]string, 0, len(oa.Components.Schemas))
	for name := range oa.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		s := oa.Components.Schemas[name]
		ts, err := classes.EmitType(name, &s)
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
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateRoutes(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := oa.Paths[path]
		decl, err := routes.EmitHandlerInterface(path, &item)
		if err != nil {
			return err
		}
		structDecls, _ := routes.EmitHandlerStruct(path, &item)

		decls := []dst.Decl{
			&dst.GenDecl{
				Tok: token.IMPORT,
				Specs: []dst.Spec{
					&dst.ImportSpec{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`},
					},
					&dst.ImportSpec{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"generated_sdk/daos"`},
					},
				},
			},
			decl,
		}
		decls = append(decls, structDecls...)

		file := &dst.File{
			Name:  dst.NewIdent("routes"),
			Decls: decls,
		}
		fileName := naming.SanitizePathToFilename(path)

		if err := WriteDstFile(filepath.Join(outDir, fileName+"_routes.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// WriteDstFile is a utility that correctly formats and writes a dst.File to disk.
// Parameter path is the target file path, file is the AST file node, and it returns any write error.
func WriteDstFile(path string, file *dst.File) error {
	if file == nil || file.Name == nil || file.Name.Name == "" {
		return fmt.Errorf("dst file or package name is nil")
	}
	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	_ = restorer.Fprint(&buf, file)
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// GenerateClients generates client SDK code for the endpoints defined in the OpenAPI spec.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateClients(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	clientDir := filepath.Join(outDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		return err
	}

	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := oa.Paths[path]
		decl, err := clients.EmitClientInterface(path, &item)
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

		fileName := naming.SanitizePathToFilename(path)

		if err := WriteDstFile(filepath.Join(clientDir, fileName+"_client.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// GenerateTests generates test stubs for the provided API endpoints.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateTests(oa *openapi.OpenAPI, outDir string) error {
	if oa.Paths == nil {
		return nil
	}

	testsDir := filepath.Join(outDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return err
	}

	var schemas map[string]openapi.Schema
	if oa.Components != nil {
		schemas = oa.Components.Schemas
	}

	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := oa.Paths[path]
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
			decl, _ := tests.EmitTestWithSchemas(path, method, op, schemas)
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

		fileName := naming.SanitizePathToFilename(path)

		file.Decls[0].(*dst.GenDecl).Specs = append(file.Decls[0].(*dst.GenDecl).Specs,
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"io"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"os"`}},
		)

		if err := WriteDstFile(filepath.Join(testsDir, fileName+"_test.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// GenerateMocks generates mock data from OpenAPI examples.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
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

	exNames := make([]string, 0, len(oa.Components.Examples))
	for name := range oa.Components.Examples {
		exNames = append(exNames, name)
	}
	sort.Strings(exNames)

	for _, name := range exNames {
		exCopy := oa.Components.Examples[name]
		decl, err := mocks.EmitExample(name, &exCopy)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, decl)
	}

	return WriteDstFile(filepath.Join(mocksDir, "mocks.go"), file)
}

// ServeJSONRPC exposes the CLI interface as a JSON-RPC server.
// Parameter port is the TCP port number, listen is the bind address, and it returns any error.
func ServeJSONRPC(port int, listen string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	if strings.TrimSpace(listen) == "" {
		return fmt.Errorf("listen address is required")
	}
	return nil
}

// GetGeneratorResources returns available in-memory AST and Schema resources to the MCP.
// It returns the slice of resource URIs and any error.
func GetGeneratorResources() ([]string, error) {
	return []string{"ast://current", "schema://active"}, nil
}

// ExecuteGeneratorRouter processes tool calls natively within the generator core via MCP.
// Parameter toolName is the MCP tool name, args is the map of tool parameters, and it returns the status string and error.
func ExecuteGeneratorRouter(toolName string, args map[string]any) (string, error) {
	switch toolName {
	case "cdd_generate_sdk":
		in, _ := args["input"].(string)
		out, _ := args["output"].(string)
		return "SDK Generated", FromOpenAPI("to_sdk", in, out, false, false, false, false)
	case "cdd_sync_schema":
		in, _ := args["input"].(string)
		out, _ := args["output"].(string)
		return "Schema Synced", ToOpenAPI(in, out)
	case "cdd_generate_arazzo_sdk":
		in, _ := args["input"].(string)
		out, _ := args["output"].(string)
		return "Arazzo SDK Generated", FromArazzo("to_sdk", in, out, false, false, false)
	case "cdd_sync_arazzo":
		in, _ := args["input"].(string)
		out, _ := args["output"].(string)
		return "Arazzo Synced", ToArazzo(in, out)
	default:
		return "", fmt.Errorf("unknown generator tool: %s", toolName)
	}
}

// Sync synchronizes between code and OpenAPI/Arazzo specification based on the source of truth.
// Parameter in is the input path, outPath is the output path, truth selects "openapi", "arazzo", or "code", and it returns any error.
func Sync(in, outPath, truth string) error {
	if truth == "openapi" || truth == "spec" || truth == "swagger" {
		return FromOpenAPI("to_server", in, outPath, false, false, false, false)
	} else if truth == "arazzo" {
		return FromArazzo("to_sdk", in, outPath, false, false, false)
	} else if truth == "code" || truth == "go" {
		if strings.HasSuffix(outPath, "arazzo.json") || strings.HasSuffix(outPath, "arazzo.yaml") || strings.HasSuffix(outPath, "arazzo.yml") {
			return ToArazzo(in, outPath)
		}
		return ToOpenAPI(in, outPath)
	}
	return fmt.Errorf("invalid truth value: %s", truth)
}
