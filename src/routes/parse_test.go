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

	if pathItem.Get == nil || pathItem.Get.OperationID != "GetUser" {
		t.Errorf("expected GetUser operation")
	}

	if pathItem.Get.Summary != "Get a user" {
		t.Errorf("expected Get a user summary")
	}

	if pathItem.Post == nil || pathItem.Post.OperationID != "PostUser" {
		t.Errorf("expected PostUser operation")
	}

	if pathItem.Put == nil || pathItem.Put.OperationID != "PutUser" {
		t.Errorf("expected PutUser operation")
	}

	if pathItem.Delete == nil || pathItem.Delete.OperationID != "DeleteUser" {
		t.Errorf("expected DeleteUser operation")
	}

	if pathItem.Patch == nil || pathItem.Patch.OperationID != "PatchUser" {
		t.Errorf("expected PatchUser operation")
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
