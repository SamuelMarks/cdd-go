package clients

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// EmitClientInterface generates a Go interface for an OpenAPI PathItem to be used as a client.
func EmitClientInterface(path string, pathItem *openapi.PathItem) (*dst.GenDecl, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("PathItem is nil")
	}

	interfaceName := "Client" + toCamelCase(path)
	iface := &dst.InterfaceType{
		Methods: &dst.FieldList{},
	}

	if pathItem.Get != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Get", pathItem.Get))
	}
	if pathItem.Post != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Post", pathItem.Post))
	}
	if pathItem.Put != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Put", pathItem.Put))
	}
	if pathItem.Delete != nil {
		iface.Methods.List = append(iface.Methods.List, emitMethodSignature("Delete", pathItem.Delete))
	}

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(interfaceName),
		Type: iface,
	}

	if pathItem.Summary != "" {
		ts.Decs.Start.Append(fmt.Sprintf("// %s", pathItem.Summary))
	}

	return &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}, nil
}

func emitMethodSignature(method string, op *openapi.Operation) *dst.Field {
	name := method
	if op.OperationID != "" {
		name = strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	}

	fType := &dst.FuncType{
		Params: &dst.FieldList{
			List: []*dst.Field{
				{
					Names: []*dst.Ident{dst.NewIdent("req")},
					Type: &dst.StarExpr{
						X: &dst.SelectorExpr{
							X:   dst.NewIdent("http"),
							Sel: dst.NewIdent("Request"),
						},
					},
				},
			},
		},
		Results: &dst.FieldList{
			List: []*dst.Field{
				{
					Type: &dst.StarExpr{
						X: &dst.SelectorExpr{
							X:   dst.NewIdent("http"),
							Sel: dst.NewIdent("Response"),
						},
					},
				},
				{
					Type: dst.NewIdent("error"),
				},
			},
		},
	}

	f := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  fType,
	}

	if op.Summary != "" {
		f.Decs.Start.Append(fmt.Sprintf("// %s", op.Summary))
	}

	return f
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if res == "" {
		res = "Root"
	}
	return res
}
