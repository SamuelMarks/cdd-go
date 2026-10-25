// Package oauthflows provides parsing and emitting of OpenAPI OAuthFlows objects to and from Go AST composite literals.
package oauthflows

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitTypes returns the Go AST type declarations for OAuthFlow and OAuthFlows.
func EmitTypes() []dst.Decl {
	flowDecl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("OAuthFlow"),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("AuthorizationURL")}, Type: dst.NewIdent("string"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"authorizationUrl,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("DeviceAuthorizationURL")}, Type: dst.NewIdent("string"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"deviceAuthorizationUrl,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("TokenURL")}, Type: dst.NewIdent("string"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"tokenUrl,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("RefreshURL")}, Type: dst.NewIdent("string"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"refreshUrl,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("Scopes")}, Type: &dst.MapType{Key: dst.NewIdent("string"), Value: dst.NewIdent("string")}, Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"scopes,omitempty\"`"}},
						},
					},
				},
			},
		},
	}
	flowsDecl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("OAuthFlows"),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{
							{Names: []*dst.Ident{dst.NewIdent("Implicit")}, Type: dst.NewIdent("OAuthFlow"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"implicit,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("Password")}, Type: dst.NewIdent("OAuthFlow"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"password,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("ClientCredentials")}, Type: dst.NewIdent("OAuthFlow"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"clientCredentials,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("AuthorizationCode")}, Type: dst.NewIdent("OAuthFlow"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"authorizationCode,omitempty\"`"}},
							{Names: []*dst.Ident{dst.NewIdent("DeviceAuthorization")}, Type: dst.NewIdent("OAuthFlow"), Tag: &dst.BasicLit{Kind: token.STRING, Value: "`json:\"deviceAuthorization,omitempty\"`"}},
						},
					},
				},
			},
		},
	}
	return []dst.Decl{flowDecl, flowsDecl}
}

// Emit formats an OpenAPI OAuthFlows object into a struct representation.
func Emit(flows *openapi.OAuthFlows) *dst.CompositeLit {
	if flows == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: dst.NewIdent("OAuthFlows"),
		Elts: []dst.Expr{},
	}

	hasContent := false

	if flows.Implicit != nil {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Implicit"),
			Value: emitFlow(flows.Implicit),
		})
		hasContent = true
	}
	if flows.Password != nil {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Password"),
			Value: emitFlow(flows.Password),
		})
		hasContent = true
	}
	if flows.ClientCredentials != nil {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("ClientCredentials"),
			Value: emitFlow(flows.ClientCredentials),
		})
		hasContent = true
	}
	if flows.AuthorizationCode != nil {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("AuthorizationCode"),
			Value: emitFlow(flows.AuthorizationCode),
		})
		hasContent = true
	}
	if flows.DeviceAuthorization != nil {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("DeviceAuthorization"),
			Value: emitFlow(flows.DeviceAuthorization),
		})
		hasContent = true
	}

	if !hasContent {
		return nil
	}

	return cl
}

// emitFlow formats a single OpenAPI OAuthFlow object into a dst.CompositeLit.
func emitFlow(flow *openapi.OAuthFlow) *dst.CompositeLit {
	cl := &dst.CompositeLit{
		Type: dst.NewIdent("OAuthFlow"),
		Elts: []dst.Expr{},
	}

	if flow.AuthorizationURL != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("AuthorizationURL"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flow.AuthorizationURL)},
		})
	}
	if flow.DeviceAuthorizationURL != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("DeviceAuthorizationURL"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flow.DeviceAuthorizationURL)},
		})
	}
	if flow.TokenURL != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("TokenURL"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flow.TokenURL)},
		})
	}
	if flow.RefreshURL != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("RefreshURL"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", flow.RefreshURL)},
		})
	}

	if len(flow.Scopes) > 0 {
		scopesMap := &dst.CompositeLit{
			Type: &dst.MapType{
				Key:   dst.NewIdent("string"),
				Value: dst.NewIdent("string"),
			},
			Elts: []dst.Expr{},
		}
		for k, v := range flow.Scopes {
			scopesMap.Elts = append(scopesMap.Elts, &dst.KeyValueExpr{
				Key:   &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", k)},
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", v)},
			})
		}
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Scopes"),
			Value: scopesMap,
		})
	}

	return cl
}
