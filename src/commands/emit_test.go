package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitCommand(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "getUser",
		Summary:     "Get user",
		Description: "Gets a user by ID",
		Parameters: []openapi.Parameter{
			{Name: "id", In: "path", Schema: &openapi.Schema{Type: "string"}},
			{Name: "verbose", In: "query"},
		},
	}

	decl := Emit("/users/{id}", "get", op)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("commands"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "\n", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "// Method: GET") {
		t.Errorf("missing method comment")
	}
	if !strings.Contains(out, "// Path: /users/{id}") {
		t.Errorf("missing path comment")
	}
	if !strings.Contains(out, "var GetUserCmd = &cobra.Command{") {
		t.Errorf("missing command var, got %s", out)
	}
	if !strings.Contains(out, `Use: "get_user"`) {
		t.Errorf("missing use")
	}
	if !strings.Contains(out, `Short: "Get user"`) {
		t.Errorf("missing short")
	}
	if !strings.Contains(out, `Long: "Gets a user by ID"`) {
		t.Errorf("missing long")
	}
}

func TestEmitCommandNoOpID(t *testing.T) {
	op := &openapi.Operation{}
	decl := Emit("/users", "post", op)
	if decl == nil {
		t.Fatalf("expected decl")
	}
	vs := decl.Specs[0].(*dst.ValueSpec)
	if vs.Names[0].Name != "PostUsersCmd" {
		t.Errorf("expected PostUsersCmd, got %s", vs.Names[0].Name)
	}
}

func TestEmitCommandNil(t *testing.T) {
	if Emit("/", "get", nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestEmitMcpCmd(t *testing.T) {
	decl := EmitMcpCmd()
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("commands"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "\n", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "// McpCmd represents the MCP CLI subcommand and stdio transport bindings.") {
		t.Errorf("missing comment, got %s", out)
	}
	if !strings.Contains(out, "var McpCmd = &cobra.Command{") {
		t.Errorf("missing command var, got %s", out)
	}
	if !strings.Contains(out, `Use: "mcp"`) {
		t.Errorf("missing use")
	}
	if !strings.Contains(out, `cmd.SetIn(os.Stdin)`) {
		t.Errorf("missing cmd.SetIn(os.Stdin), got %s", out)
	}
	if !strings.Contains(out, `cmd.SetOut(os.Stdout)`) {
		t.Errorf("missing cmd.SetOut(os.Stdout), got %s", out)
	}
	if !strings.Contains(out, `cmd.Println("MCP server started on stdio")`) {
		t.Errorf("missing cmd.Println statement, got %s", out)
	}
}
