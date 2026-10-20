package routes

import (
	"testing"

	"github.com/dave/dst"
)

func TestParseHandlerInterface(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("HandlerUsersId"),
		Type: &dst.InterfaceType{
			Methods: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("GetUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
						Decs: dst.FieldDecorations{
							NodeDecs: dst.NodeDecs{
								Start: dst.Decorations{"// Get a user"},
							},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("PostUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("PutUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("DeleteUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("PatchUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						// Embedded field
						Type: dst.NewIdent("EmbeddedHandler"),
					},
				},
			},
		},
		Decs: dst.TypeSpecDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// User endpoints"},
			},
		},
	}

	pathItem, err := ParseHandlerInterface(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pathItem.Summary != "User endpoints" {
		t.Errorf("expected User endpoints, got %s", pathItem.Summary)
	}

	if pathItem.Get == nil || pathItem.Get.OperationID != "getUser" {
		t.Errorf("expected getUser operation")
	}

	if pathItem.Get.Summary != "Get a user" {
		t.Errorf("expected Get a user summary")
	}

	if pathItem.Post == nil || pathItem.Post.OperationID != "postUser" {
		t.Errorf("expected postUser operation")
	}

	if pathItem.Put == nil || pathItem.Put.OperationID != "putUser" {
		t.Errorf("expected putUser operation")
	}

	if pathItem.Delete == nil || pathItem.Delete.OperationID != "deleteUser" {
		t.Errorf("expected deleteUser operation")
	}

	if pathItem.Patch == nil || pathItem.Patch.OperationID != "patchUser" {
		t.Errorf("expected patchUser operation")
	}
}

func TestParseHandlerInterfaceNil(t *testing.T) {
	_, err := ParseHandlerInterface(nil)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseHandlerInterfaceNotInterface(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("Struct"),
		Type: &dst.StructType{},
	}
	_, err := ParseHandlerInterface(ts)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseHandlerInterfaceExtraVerbs(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("HandlerExtra"),
		Type: &dst.InterfaceType{
			Methods: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("OptionsUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("HeadUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("TraceUser")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("CustomAction")},
						Type: &dst.FuncType{
							Params:  &dst.FieldList{},
							Results: &dst.FieldList{},
						},
					},
				},
			},
		},
	}
	pathItem, err := ParseHandlerInterface(ts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if pathItem.Options == nil {
		t.Errorf("expected Options")
	}
	if pathItem.Head == nil {
		t.Errorf("expected Head")
	}
	if pathItem.Trace == nil {
		t.Errorf("expected Trace")
	}
}

func TestParseHandlerInterfaceMCPExt(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("HandlerTest"),
		Type: &dst.InterfaceType{
			Methods: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("HandleMcpSse")},
						Type:  &dst.FuncType{},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("HandleMcpMessage")},
						Type:  &dst.FuncType{},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("WithMcpAuth")},
						Type:  &dst.FuncType{},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("MapMcpToolToRoute")},
						Type:  &dst.FuncType{},
					},
				},
			},
		},
	}

	pi, err := ParseHandlerInterface(ts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if pi.Extensions == nil {
		t.Fatalf("expected extensions for MCP methods")
	}
	ext := pi.Extensions["x-mcp-server"]
	if ext["sse"] != true || ext["message"] != true || ext["auth"] != true || ext["proxy"] != true {
		t.Errorf("expected MCP server extensions to be set, got %v", ext)
	}
}

func TestParseHandlerInterfacePathAndMethodComment(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("HandlerUsers"),
		Decs: dst.TypeSpecDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// Path: /users", "// Users summary"},
			},
		},
		Type: &dst.InterfaceType{
			Methods: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("CustomOp")},
						Type:  &dst.FuncType{},
						Decs: dst.FieldDecorations{
							NodeDecs: dst.NodeDecs{
								Start: dst.Decorations{"// METHOD: POST", "// Custom post"},
							},
						},
					},
				},
			},
		},
	}
	pi, err := ParseHandlerInterface(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pi.Summary != "Users summary" {
		t.Errorf("expected Users summary, got %q", pi.Summary)
	}
	if pi.Post == nil || pi.Post.OperationID != "customOp" || pi.Post.Summary != "Custom post" {
		t.Errorf("expected customOp POST operation, got %+v", pi.Post)
	}
}
