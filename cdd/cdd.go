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

	"github.com/SamuelMarks/cdd-go/cdd/models"
	"github.com/SamuelMarks/cdd-go/src/callbacks"
	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/clients"
	"github.com/SamuelMarks/cdd-go/src/commands"
	"github.com/SamuelMarks/cdd-go/src/components"
	"github.com/SamuelMarks/cdd-go/src/externaldocs"
	"github.com/SamuelMarks/cdd-go/src/functions"
	"github.com/SamuelMarks/cdd-go/src/info"
	"github.com/SamuelMarks/cdd-go/src/ir"
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
	// GithubOrg is the optional GitHub organization or user name for module resolution.
	GithubOrg string
	// RepoName is the optional repository name override for module resolution.
	RepoName string
	// Topology selects "monorepo" or "polyrepo" module path layout.
	Topology string
}

// MCPToolEntry represents a generated Model Context Protocol (MCP) tool descriptor.
type MCPToolEntry struct {
	// Name is the MCP tool identifier (operationId or workflowId).
	Name string
	// Description is the human-readable summary of the MCP tool.
	Description string
}

// GenerateSDK generates the SDK based on the provided configuration.
// Parameter config holds the generation options, and it returns any generation error.
func GenerateSDK(config Config) error {
	in := config.InputPath
	if in == "" {
		in = config.InputDir
	}
	return FromOpenAPIWithConfig("to_sdk", in, config.OutputDir, config)
}

// ResolveModulePath computes the Go module path from the provided GitHub organization, repository name, and repository topology ("monorepo" or "polyrepo").
// Parameter githubOrg is the GitHub organization or user name, repoName is the optional repository name override, and topology selects "monorepo" or "polyrepo" layout.
func ResolveModulePath(githubOrg, repoName, topology string) string {
	org := strings.Trim(strings.TrimSpace(githubOrg), "/")
	repo := strings.Trim(strings.TrimSpace(repoName), "/")
	topo := strings.ToLower(strings.TrimSpace(topology))

	if org == "" && repo == "" {
		return "generated_sdk"
	}

	if org == "" {
		mod := repo
		if !strings.HasPrefix(repo, "github.com/") && strings.Contains(repo, "/") {
			mod = "github.com/" + repo
		}
		if topo == "monorepo" && !strings.HasSuffix(mod, "/go") {
			return mod + "/go"
		}
		return mod
	}

	if repo == "" {
		if topo == "monorepo" {
			return fmt.Sprintf("github.com/%s/%s-sdk/go", org, org)
		}
		return fmt.Sprintf("github.com/%s/%s-sdk-go", org, org)
	}

	mod := fmt.Sprintf("github.com/%s/%s", org, repo)
	if topo == "monorepo" && !strings.HasSuffix(mod, "/go") {
		return mod + "/go"
	}
	return mod
}

// ReadModulePath inspects go.mod in outDir and returns the declared module path, or "generated_sdk" if go.mod is absent or has no module directive.
// Parameter outDir is the target directory containing an optional go.mod file.
func ReadModulePath(outDir string) string {
	b, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		return "generated_sdk"
	}
	for _, line := range strings.Split(string(b), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
			if mod != "" {
				return mod
			}
		}
	}
	return "generated_sdk"
}

// loadAndNormalizeOpenAPI opens and parses an OpenAPI specification file.
// Parameter inputPath is the file path to the OpenAPI spec, and it returns the parsed OpenAPI document and error.
func loadAndNormalizeOpenAPI(inputPath string) (*openapi.OpenAPI, error) {
	if inputPath == "" {
		return nil, fmt.Errorf("input file or directory is required")
	}
	f, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	oa, err := openapi.Parse(f)
	if err != nil {
		return nil, err
	}
	oa, _ = ir.NormalizeOpenAPI(oa)
	fmt.Println("Parsed OpenAPI Version:", oa.OpenAPI)
	fmt.Println("API Title:", oa.Info.Title)
	return oa, nil
}

// emitSDKPackages generates the Go SDK models, client interfaces, and optional tests/mocks.
// Parameter oa is the OpenAPI document, outDir is the destination directory, and tests enables test/mock generation.
func emitSDKPackages(oa *openapi.OpenAPI, outDir string, tests bool) error {
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
	return nil
}

// emitServerPackages generates the Go Gin server, GORM database/DAOs, middlewares, IdP, and optional tests/mocks/seeder.
// Parameter oa is the OpenAPI document, outDir is the destination directory, and tests enables test/mock/seeder generation.
func emitServerPackages(oa *openapi.OpenAPI, outDir string, tests bool) error {
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
	if err := GenerateWebhookTriggers(oa, outDir); err != nil {
		return err
	}
	if err := GenerateServerMain(oa, outDir, true, tests, true); err != nil {
		return err
	}
	_ = GenerateServerReadme(oa, outDir)
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
	return nil
}

// emitSDKCLIPackages generates the Go SDK, Cobra CLI commands, and optional tests/mocks.
// Parameter oa is the OpenAPI document, outDir is the destination directory, and tests enables test/mock generation.
func emitSDKCLIPackages(oa *openapi.OpenAPI, outDir string, tests bool) error {
	if err := emitSDKPackages(oa, outDir, false); err != nil {
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
	return nil
}

// FromOpenAPI generates code from an OpenAPI specification.
// Parameter subsubcommand selects the target ("to_sdk", "to_server", "to_sdk_cli"), in is the spec path, outDir is the destination directory, noGithubActions omits CI workflow files, noInstallablePackage omits go.mod, tests enables test generation, and mcp enables MCP bindings.
func FromOpenAPI(subsubcommand, in, outDir string, noGithubActions, noInstallablePackage, tests, mcp bool) error {
	return FromOpenAPIWithConfig(subsubcommand, in, outDir, Config{
		InputPath:             in,
		OutputDir:             outDir,
		NoGithubActions:       noGithubActions,
		NoInstallablePackage:  noInstallablePackage,
		CreateComposableTests: tests,
		Mcp:                   mcp,
	})
}

// FromOpenAPIWithConfig generates code from an OpenAPI specification using the full Config options.
// Parameter subsubcommand selects the target ("to_sdk", "to_server", "to_sdk_cli"), in is the spec path, outDir is the destination directory, and config holds module/CI/MCP options.
func FromOpenAPIWithConfig(subsubcommand, in, outDir string, config Config) error {
	oa, err := loadAndNormalizeOpenAPI(in)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(outDir, 0755)
	if !config.NoInstallablePackage {
		GenerateGoModWithConfig(outDir, config.GithubOrg, config.RepoName, config.Topology)
	}
	if !config.NoGithubActions {
		GenerateGithubActionsWithConfig(outDir, config.GithubOrg, config.RepoName, config.Topology)
	}

	switch subsubcommand {
	case "to_sdk":
		err = emitSDKPackages(oa, outDir, config.CreateComposableTests)
	case "to_server":
		err = emitServerPackages(oa, outDir, config.CreateComposableTests)
	case "to_sdk_cli":
		err = emitSDKCLIPackages(oa, outDir, config.CreateComposableTests)
	default:
		return models.NewDomainError(models.ErrKindInvalidArgument, 400, fmt.Sprintf("unknown subsubcommand: %s", subsubcommand), nil)
	}
	if err != nil {
		return err
	}

	if config.Mcp {
		if err := GenerateOpenAPIMCPTools(oa, outDir); err != nil {
			return err
		}
	}
	return nil
}

// GenerateOpenAPIMCPTools emits `mcp/mcp_tools.go` exposing OpenAPI operations as strongly-typed MCP tools.
// Parameter oa is the parsed OpenAPI document, outDir is the destination directory, and it returns any generation error.
func GenerateOpenAPIMCPTools(oa *openapi.OpenAPI, outDir string) error {
	var tools []MCPToolEntry
	if oa != nil && len(oa.Paths) > 0 {
		pathKeys := make([]string, 0, len(oa.Paths))
		for p := range oa.Paths {
			pathKeys = append(pathKeys, p)
		}
		sort.Strings(pathKeys)
		for _, p := range pathKeys {
			item := oa.Paths[p]
			for _, mo := range openapi.PathItemOperations(&item) {
				name := mo.Operation.OperationID
				if name == "" {
					name = strings.ToLower(mo.Method) + "_" + naming.SanitizePathToFilename(p)
				}
				desc := mo.Operation.Summary
				if desc == "" {
					desc = mo.Operation.Description
				}
				tools = append(tools, MCPToolEntry{
					Name:        name,
					Description: desc,
				})
			}
		}
	}
	return writeMCPToolsFile(outDir, tools)
}

// writeMCPToolsFile writes the `mcp/mcp_tools.go` package file containing MCPTool and GetMCPTools().
// Parameter outDir is the destination root directory, tools is the slice of MCP tool descriptors, and it returns any write error.
func writeMCPToolsFile(outDir string, tools []MCPToolEntry) error {
	mcpDir := filepath.Join(outDir, "mcp")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return err
	}

	nameField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("Name")},
		Type:  dst.NewIdent("string"),
		Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`" + `json:"name"` + "`"},
	}
	nameField.Decs.Start.Append("// Name is the MCP tool identifier.")
	descField := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("Description")},
		Type:  dst.NewIdent("string"),
		Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`" + `json:"description,omitempty"` + "`"},
	}
	descField.Decs.Start.Append("// Description is the human-readable tool description.")

	toolTypeSpec := &dst.TypeSpec{
		Name: dst.NewIdent("MCPTool"),
		Type: &dst.StructType{Fields: &dst.FieldList{List: []*dst.Field{nameField, descField}}},
	}
	toolTypeDecl := &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{toolTypeSpec}}
	toolTypeDecl.Decs.Start.Append("// MCPTool represents a Model Context Protocol (MCP) 1.0.0 tool definition.")

	var elts []dst.Expr
	for _, t := range tools {
		elts = append(elts, &dst.CompositeLit{
			Elts: []dst.Expr{
				&dst.KeyValueExpr{Key: dst.NewIdent("Name"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", t.Name)}},
				&dst.KeyValueExpr{Key: dst.NewIdent("Description"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", t.Description)}},
			},
		})
	}

	getToolsFn := &dst.FuncDecl{
		Name: dst.NewIdent("GetMCPTools"),
		Type: &dst.FuncType{
			Params:  &dst.FieldList{},
			Results: &dst.FieldList{List: []*dst.Field{{Type: &dst.ArrayType{Elt: dst.NewIdent("MCPTool")}}}},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ReturnStmt{
					Results: []dst.Expr{
						&dst.CompositeLit{
							Type: &dst.ArrayType{Elt: dst.NewIdent("MCPTool")},
							Elts: elts,
						},
					},
				},
			},
		},
	}
	getToolsFn.Decs.Start.Append("// GetMCPTools returns the registered Model Context Protocol tools for this module.")

	file := &dst.File{
		Name:  dst.NewIdent("mcp"),
		Decls: []dst.Decl{toolTypeDecl, getToolsFn},
	}
	file.Decs.Start.Append("// Package mcp provides generated Model Context Protocol (MCP) tool definitions.")
	return WriteDstFile(filepath.Join(mcpDir, "mcp_tools.go"), file)
}

// GenerateGoMod creates a go.mod and .goreleaser.yaml file in the output directory if they do not already exist.
// Parameter outDir is the destination directory path.
func GenerateGoMod(outDir string) {
	GenerateGoModWithConfig(outDir, "", "", "")
}

// GenerateGoModWithConfig creates a go.mod and .goreleaser.yaml file in the output directory using the resolved module path.
// Parameter outDir is the destination directory, githubOrg is the GitHub org, repoName is the repo name, and topology selects "monorepo" or "polyrepo".
func GenerateGoModWithConfig(outDir, githubOrg, repoName, topology string) {
	modulePath := ResolveModulePath(githubOrg, repoName, topology)
	goModPath := filepath.Join(outDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		content := fmt.Sprintf(`module %s

go 1.25.7

require (
	github.com/brianvoe/gofakeit/v6 v6.28.0
	github.com/gin-gonic/gin v1.10.0
	github.com/spf13/cobra v1.8.1
	gorm.io/driver/postgres v1.5.11
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)
`, modulePath)
		_ = os.WriteFile(goModPath, []byte(content), 0644)
	}

	goreleaserPath := filepath.Join(outDir, ".goreleaser.yaml")
	if _, err := os.Stat(goreleaserPath); os.IsNotExist(err) {
		goreleaserContent := `# GoReleaser configuration with GPG detached signature (gpg --armor --detach-sign)
version: 2
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
checksum:
  name_template: 'checksums.txt'
signs:
  - artifacts: checksum
    cmd: gpg
    output: true
    signature: "${artifact}.asc"
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--armor"
      - "--detach-sign"
      - "${artifact}"
`
		_ = os.WriteFile(goreleaserPath, []byte(goreleaserContent), 0644)
	}
}

// GenerateGithubActions creates a CI GitHub Actions workflow file in the output directory.
// Parameter outDir is the destination directory path.
func GenerateGithubActions(outDir string) {
	GenerateGithubActionsWithConfig(outDir, "", "", "")
}

// GenerateGithubActionsWithConfig creates a CI GitHub Actions workflow file supporting go test, GoReleaser GPG signing, proxy.golang.org warming, and gh-pages deployment.
// Parameter outDir is the destination directory, githubOrg is the GitHub org, repoName is the repo name, and topology selects "monorepo" or "polyrepo".
func GenerateGithubActionsWithConfig(outDir, githubOrg, repoName, topology string) {
	modulePath := ResolveModulePath(githubOrg, repoName, topology)
	ciPath := filepath.Join(outDir, ".github", "workflows", "ci.yml")
	_ = os.MkdirAll(filepath.Dir(ciPath), 0755)
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		content := fmt.Sprintf(`name: CI

on:
  push:
    branches: [ main, master ]
    tags:
      - 'v*'
      - 'go/v*'
  pull_request:
    branches: [ main, master ]

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v6
      with:
        fetch-depth: 0
    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.25'
    - name: Build
      run: go build -v ./...
    - name: Test
      run: go test -v ./...
    - name: Import GPG key
      if: startsWith(github.ref, 'refs/tags/v') || startsWith(github.ref, 'refs/tags/go/v')
      uses: crazy-max/ghaction-import-gpg@v6
      with:
        gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
        passphrase: ${{ secrets.GPG_PASSPHRASE }}
    - name: Run GoReleaser (gpg --armor --detach-sign)
      if: startsWith(github.ref, 'refs/tags/v') || startsWith(github.ref, 'refs/tags/go/v')
      uses: goreleaser/goreleaser-action@v6
      with:
        version: '~> v2'
        args: release --clean
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
    - name: Warm proxy.golang.org and pkg.go.dev index
      if: startsWith(github.ref, 'refs/tags/v') || startsWith(github.ref, 'refs/tags/go/v')
      run: |
        TAG_NAME="${GITHUB_REF#refs/tags/}"
        curl -fsSL "https://proxy.golang.org/%s/@v/${TAG_NAME##*/}.info" || true
    - name: Generate static godoc / pkgsite documentation
      run: |
        mkdir -p docs_site
        go doc -all ./... > docs_site/index.txt || true
    - name: Deploy documentation to gh-pages
      if: github.event_name == 'push' && (github.ref == 'refs/heads/main' || github.ref == 'refs/heads/master')
      uses: peaceiris/actions-gh-pages@v4
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./docs_site
        publish_branch: gh-pages
`, modulePath)
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
		for _, mo := range openapi.PathItemOperations(&item) {
			file.Decls = append(file.Decls, commands.Emit(path, mo.Method, mo.Operation))
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

// mergeComponentsFileIntoOpenAPI merges reusable components parsed from components.go into oa.Components.
// Parameter f is the parsed DST file, and oa is the target OpenAPI document.
func mergeComponentsFileIntoOpenAPI(f *dst.File, oa *openapi.OpenAPI) {
	parsedComp := components.Parse(f)
	if parsedComp == nil {
		return
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

// parseTypeDeclIntoOpenAPI extracts interfaces (Client/Handler) and struct/alias schemas from a TYPE declaration into oa.
// Parameter d is the TYPE AST declaration, and oa is the target OpenAPI document.
func parseTypeDeclIntoOpenAPI(d *dst.GenDecl, oa *openapi.OpenAPI) {
	for _, spec := range d.Specs {
		if ts, ok := spec.(*dst.TypeSpec); ok {
			if _, isIface := ts.Type.(*dst.InterfaceType); isIface {
				allDecs := append(dst.Decorations{}, d.Decs.Start...)
				allDecs = append(allDecs, ts.Decs.Start...)
				if strings.HasPrefix(ts.Name.Name, "Client") {
					if pi, err := clients.ParseClientInterface(ts); err == nil && pi != nil {
						name := strings.TrimPrefix(ts.Name.Name, "Client")
						path := extractPathComment(allDecs, name)
						oa.Paths[path] = *pi
					}
				} else {
					if pi, err := routes.ParseHandlerInterface(ts); err == nil && pi != nil {
						name := strings.TrimPrefix(ts.Name.Name, "Handler")
						path := extractPathComment(allDecs, name)
						oa.Paths[path] = *pi
					}
				}
			} else if ts.Name.Name != "OAuthFlow" && ts.Name.Name != "OAuthFlows" && ts.Name.Name != "MCPTool" {
				ts.Decs.Start = append(ts.Decs.Start, d.Decs.Start...)
				if schema, err := classes.ParseType(ts); err == nil && schema != nil {
					oa.Components.Schemas[ts.Name.Name] = *schema
				}
			}
		}
	}
}

// parseVarOrConstDeclIntoOpenAPI extracts Info, Servers, Tags, ExternalDocs, Webhooks, Callbacks, Links, Commands, and Examples from a VAR or CONST declaration.
// Parameter d is the VAR/CONST AST declaration, and oa is the target OpenAPI document.
func parseVarOrConstDeclIntoOpenAPI(d *dst.GenDecl, oa *openapi.OpenAPI) {
	if parsedInfo, err := info.Parse(d); err == nil && parsedInfo != nil {
		oa.Info = *parsedInfo
		return
	}
	if parsedServers, err := servers.Parse(d); err == nil && parsedServers != nil {
		oa.Servers = parsedServers
		return
	}
	if parsedTags, err := tags.Parse(d); err == nil && parsedTags != nil {
		oa.Tags = parsedTags
		return
	}
	if parsedExtDocs, err := externaldocs.Parse(d); err == nil && parsedExtDocs != nil {
		oa.ExternalDocs = parsedExtDocs
		return
	}
	if whName, parsedWh := webhooks.Parse(d); parsedWh != nil && whName != "" {
		if oa.Webhooks == nil {
			oa.Webhooks = make(map[string]openapi.PathItem)
		}
		for k, v := range parsedWh {
			oa.Webhooks[k] = v
		}
		return
	}
	if cbName, parsedCb := callbacks.Parse(d); parsedCb != nil && cbName != "" {
		if oa.Components.Callbacks == nil {
			oa.Components.Callbacks = make(map[string]openapi.Callback)
		}
		oa.Components.Callbacks[cbName] = *parsedCb
		return
	}
	if linkName, parsedLink := links.Parse(d); parsedLink != nil && linkName != "" {
		if oa.Components.Links == nil {
			oa.Components.Links = make(map[string]openapi.Link)
		}
		oa.Components.Links[linkName] = *parsedLink
		return
	}
	if method, path, op := commands.Parse(d); op != nil && method != "" && path != "" {
		pi := oa.Paths[path]
		pi.SetOperation(method, op)
		oa.Paths[path] = pi
		return
	}
	for _, spec := range d.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			vs.Decs.Start = append(vs.Decs.Start, d.Decs.Start...)
			if ex, err := mocks.ParseExample(vs); err == nil && ex != nil {
				oa.Components.Examples[vs.Names[0].Name] = *ex
			}
		}
	}
}

// parseFuncDeclIntoOpenAPI extracts standalone operation functions annotated with // Path: <path> using src/functions.Parse.
// Parameter fn is the AST function declaration, and oa is the target OpenAPI document.
func parseFuncDeclIntoOpenAPI(fn *dst.FuncDecl, oa *openapi.OpenAPI) {
	if fn.Recv != nil {
		return
	}
	var path, method string
	for _, c := range fn.Decs.Start {
		line := strings.TrimSpace(strings.TrimPrefix(c, "//"))
		if strings.HasPrefix(line, "Path:") {
			path = strings.TrimSpace(strings.TrimPrefix(line, "Path:"))
		} else if strings.HasPrefix(line, "METHOD:") {
			method = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "METHOD:")))
		}
	}
	if path == "" {
		return
	}
	if op, err := functions.ParseOperation(fn); err == nil && op != nil {
		if method == "" {
			method = "post"
		}
		pi := oa.Paths[path]
		pi.SetOperation(method, op)
		oa.Paths[path] = pi
	}
}

// parseASTFileIntoOpenAPI processes all declarations in file and merges extracted OpenAPI constructs into oa.
// Parameter file is the parsed Go AST file, isComponentsFile indicates whether the file is components.go, and oa is the target OpenAPI document.
func parseASTFileIntoOpenAPI(file *dst.File, isComponentsFile bool, oa *openapi.OpenAPI) {
	if isComponentsFile {
		mergeComponentsFileIntoOpenAPI(file, oa)
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			switch d.Tok {
			case token.TYPE:
				parseTypeDeclIntoOpenAPI(d, oa)
			case token.VAR, token.CONST:
				parseVarOrConstDeclIntoOpenAPI(d, oa)
			}
		case *dst.FuncDecl:
			parseFuncDeclIntoOpenAPI(d, oa)
		}
	}
}

// GenerateOpenAPI analyzes the input path (file or directory) and generates an OpenAPI JSON specification.
// Parameter inputPath is the Go file or directory to analyze, outPath is the output OpenAPI file path, and it returns any error.
func GenerateOpenAPI(inputPath string, outPath string) error {
	oa := &openapi.OpenAPI{
		OpenAPI: "3.2.0",
		Info: openapi.Info{
			Title:   "Generated API",
			Version: "0.0.5",
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
		file, err := decorator.ParseFile(fset, fpath, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", fpath, err)
		}
		parseASTFileIntoOpenAPI(file, strings.HasSuffix(fpath, "components.go"), oa)
	}

	_ = os.MkdirAll(filepath.Dir(outPath), 0755)

	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	normOA, _ := ir.NormalizeOpenAPI(oa)
	return openapi.Emit(out, normOA)
}

// GenerateClasses generates Go structs based on the components/schemas defined in the OpenAPI spec.
// Parameter oa is the parsed OpenAPI specification, outDir is the destination directory, and it returns any generation error.
func GenerateClasses(oa *openapi.OpenAPI, outDir string) error {
	if oa.Components == nil {
		return nil
	}

	compWithoutSchemas := *oa.Components
	compWithoutSchemas.Schemas = nil
	compDecls := components.Emit(&compWithoutSchemas)

	if len(compDecls) == 0 && len(oa.Components.Schemas) == 0 {
		return nil
	}

	modelsDir := filepath.Join(outDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}

	if len(compDecls) > 0 {
		file := &dst.File{
			Name:  dst.NewIdent("models"),
			Decls: compDecls,
		}
		if err := WriteDstFile(filepath.Join(modelsDir, "components.go"), file); err != nil {
			return err
		}
	}

	if len(oa.Components.Schemas) == 0 {
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
	if len(oa.Paths) == 0 {
		return nil
	}

	routesDir := filepath.Join(outDir, "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return err
	}

	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	daosImport := `"` + ReadModulePath(outDir) + `/daos"`

	for _, path := range pathKeys {
		item := oa.Paths[path]
		decl, structDecls, err := routes.EmitRouteBundle(path, &item)
		if err != nil {
			return err
		}

		decls := []dst.Decl{
			&dst.GenDecl{
				Tok: token.IMPORT,
				Specs: []dst.Spec{
					&dst.ImportSpec{
						Path: &dst.BasicLit{Kind: token.STRING, Value: `"github.com/gin-gonic/gin"`},
					},
					&dst.ImportSpec{
						Path: &dst.BasicLit{Kind: token.STRING, Value: daosImport},
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

		if err := WriteDstFile(filepath.Join(routesDir, fileName+"_routes.go"), file); err != nil {
			return err
		}
	}
	return nil
}

// ensureFileDocCoverage populates missing GoDoc comments on package, type, field, method, variable/constant, and function declarations in file.
// Parameter file is the AST file to annotate prior to formatting.
func ensureFileDocCoverage(file *dst.File) {
	if len(file.Decs.Start) == 0 {
		file.Decs.Start.Append(fmt.Sprintf("// Package %s provides generated API code.", file.Name.Name))
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			if d.Tok == token.IMPORT {
				d.Decs.After = dst.EmptyLine
				continue
			}
			if d.Tok == token.TYPE || d.Tok == token.VAR || d.Tok == token.CONST {
				d.Decs.Before = dst.EmptyLine
				d.Decs.After = dst.EmptyLine
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *dst.TypeSpec:
						d.Decs.Start = append(d.Decs.Start, s.Decs.Start...)
						s.Decs.Start = nil
						if len(d.Decs.Start) == 0 {
							d.Decs.Start.Append(fmt.Sprintf("// %s represents the %s type.", s.Name.Name, s.Name.Name))
						}
						if st, ok := s.Type.(*dst.StructType); ok && st.Fields != nil {
							for _, field := range st.Fields.List {
								field.Decs.Before = dst.NewLine
								field.Decs.After = dst.NewLine
								if len(field.Names) > 0 && len(field.Decs.Start) == 0 && len(field.Decs.End) == 0 {
									field.Decs.Start.Append(fmt.Sprintf("// %s field.", field.Names[0].Name))
								}
							}
						} else if iface, ok := s.Type.(*dst.InterfaceType); ok && iface.Methods != nil {
							for _, method := range iface.Methods.List {
								method.Decs.Before = dst.NewLine
								method.Decs.After = dst.NewLine
								if len(method.Names) > 0 && len(method.Decs.Start) == 0 && len(method.Decs.End) == 0 {
									method.Decs.Start.Append(fmt.Sprintf("// %s performs the %s operation.", method.Names[0].Name, method.Names[0].Name))
								}
							}
						}
					case *dst.ValueSpec:
						if len(s.Names) > 0 {
							if len(d.Specs) == 1 {
								d.Decs.Start = append(d.Decs.Start, s.Decs.Start...)
								s.Decs.Start = nil
								if len(d.Decs.Start) == 0 {
									d.Decs.Start.Append(fmt.Sprintf("// %s holds the %s value.", s.Names[0].Name, s.Names[0].Name))
								}
							} else {
								s.Decs.Before = dst.NewLine
								s.Decs.After = dst.NewLine
								s.Decs.Start.Append(fmt.Sprintf("// %s holds the %s value.", s.Names[0].Name, s.Names[0].Name))
							}
						}
					}
				}
			}
		case *dst.FuncDecl:
			d.Decs.Before = dst.EmptyLine
			d.Decs.After = dst.EmptyLine
			if len(d.Decs.Start) == 0 && d.Name != nil {
				d.Decs.Start.Append(fmt.Sprintf("// %s executes the %s function.", d.Name.Name, d.Name.Name))
			}
		}
	}
}

// WriteDstFile is a utility that correctly formats and writes a dst.File to disk.
// Parameter path is the target file path, file is the AST file node, and it returns any write error.
func WriteDstFile(path string, file *dst.File) error {
	if file == nil || file.Name == nil || file.Name.Name == "" {
		return fmt.Errorf("dst file or package name is nil")
	}
	ensureFileDocCoverage(file)
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

	defaultBaseURL := "http://localhost:8080"
	if len(oa.Servers) > 0 && oa.Servers[0].URL != "" {
		defaultBaseURL = oa.Servers[0].URL
	}

	pathKeys := make([]string, 0, len(oa.Paths))
	for path := range oa.Paths {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		item := oa.Paths[path]
		ops := openapi.PathItemOperations(&item)
		if len(ops) == 0 {
			continue
		}

		var testDecls []dst.Decl
		needsStrings := false
		for _, mo := range ops {
			decl, hasBody, _ := tests.EmitTestWithConfig(path, mo.Method, mo.Operation, schemas, defaultBaseURL)
			testDecls = append(testDecls, decl)
			if hasBody {
				needsStrings = true
			}
		}

		importSpecs := []dst.Spec{
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"encoding/json"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"io"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"net/http"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"os"`}},
			&dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"testing"`}},
		}
		if needsStrings {
			importSpecs = append(importSpecs, &dst.ImportSpec{Path: &dst.BasicLit{Kind: token.STRING, Value: `"strings"`}})
		}

		file := &dst.File{
			Name: dst.NewIdent("tests"),
			Decls: append([]dst.Decl{
				&dst.GenDecl{
					Tok:   token.IMPORT,
					Specs: importSpecs,
				},
			}, testDecls...),
		}

		fileName := naming.SanitizePathToFilename(path)

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

// GetGeneratorResources returns available in-memory AST and Schema resources to the MCP.
// It returns the slice of resource URIs and any error.
func GetGeneratorResources() ([]string, error) {
	return []string{"ast://current", "schema://active"}, nil
}

// ExecuteGeneratorRouter processes tool calls natively within the generator core via MCP.
// Parameter toolName is the MCP tool name, args is the strongly-typed MCPToolParams, and it returns the status string and error.
func ExecuteGeneratorRouter(toolName string, args MCPToolParams) (string, error) {
	switch toolName {
	case "cdd_generate_sdk":
		return "SDK Generated", FromOpenAPI("to_sdk", args.Input, args.Output, false, false, false, false)
	case "cdd_sync_schema":
		return "Schema Synced", ToOpenAPI(args.Input, args.Output)
	case "cdd_generate_arazzo_sdk":
		return "Arazzo SDK Generated", FromArazzo("to_sdk", args.Input, args.Output, false, false, false)
	case "cdd_sync_arazzo":
		return "Arazzo Synced", ToArazzo(args.Input, args.Output)
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
