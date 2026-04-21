package responses

import (
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"testing"
)

func TestEmitResponses(t *testing.T) {
	resps := openapi.Responses{
		"200": {
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{Ref: "#/components/schemas/User"},
				},
			},
			Headers: map[string]openapi.Header{
				"X-RateLimit": {
					Description: "Limit",
					Schema:      &openapi.Schema{Type: "integer"},
				},
			},
		},
	}

	exprs := Emit(resps)
	if len(exprs) != 3 {
		t.Fatalf("expected 3 returns, got %d", len(exprs))
	}

	if star, ok := exprs[0].(*dst.StarExpr); !ok {
		t.Errorf("expected StarExpr")
	} else if ident, ok := star.X.(*dst.Ident); !ok || ident.Name != "User" {
		t.Errorf("expected User")
	}

	if ident, ok := exprs[1].(*dst.Ident); !ok || ident.Name != "int" {
		t.Errorf("expected int for header, got %+v", exprs[1])
	}

	if ident, ok := exprs[2].(*dst.Ident); !ok || ident.Name != "error" {
		t.Errorf("expected error")
	}
}

func TestEmitResponsesNil(t *testing.T) {
	exprs := Emit(nil)
	if len(exprs) != 1 {
		t.Fatalf("expected 1 return")
	}
	if ident, ok := exprs[0].(*dst.Ident); !ok || ident.Name != "error" {
		t.Errorf("expected error")
	}
}
