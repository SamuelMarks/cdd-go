package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaResourcesStructs(t *testing.T) {
	// Test Resource
	res := Resource{URI: "file://test", Name: "TestRes"}
	b, _ := json.Marshal(res)
	if !strings.Contains(string(b), `"file://test"`) || !strings.Contains(string(b), `"TestRes"`) {
		t.Errorf("expected uri and name")
	}

	// Test ResourceTemplate
	resTmpl := ResourceTemplate{URITemplate: "file://{path}", Name: "Tmpl"}
	b, _ = json.Marshal(resTmpl)
	if !strings.Contains(string(b), `"file://{path}"`) {
		t.Errorf("expected uriTemplate")
	}

	// Test ResourceReference
	resRef := ResourceReference{Type: "ref/resource", URI: "file://ref"}
	b, _ = json.Marshal(resRef)
	if !strings.Contains(string(b), `"ref/resource"`) {
		t.Errorf("expected ref/resource type")
	}

	// Test ReadResourceRequest
	req := ReadResourceRequest{Method: "resources/read"}
	req.Params.URI = "file://read"
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"resources/read"`) || !strings.Contains(string(b), `"file://read"`) {
		t.Errorf("expected method and uri")
	}

	// Test ReadResourceResult
	resRead := ReadResourceResult{}
	resRead.Contents = append(resRead.Contents, TextResourceContents{URI: "file://read", Text: "data"})
	b, _ = json.Marshal(resRead)
	if !strings.Contains(string(b), `"data"`) {
		t.Errorf("expected content")
	}

	// Test ListResourcesRequest
	listReq := ListResourcesRequest{Method: "resources/list"}
	b, _ = json.Marshal(listReq)
	if !strings.Contains(string(b), `"resources/list"`) {
		t.Errorf("expected list method")
	}

	// Test ListResourcesResult
	listRes := ListResourcesResult{Resources: []Resource{res}}
	b, _ = json.Marshal(listRes)
	if !strings.Contains(string(b), `"TestRes"`) {
		t.Errorf("expected resource in result")
	}

	// Test ListResourceTemplatesRequest
	tmplReq := ListResourceTemplatesRequest{Method: "resources/templates/list"}
	b, _ = json.Marshal(tmplReq)
	if !strings.Contains(string(b), `"resources/templates/list"`) {
		t.Errorf("expected templates/list method")
	}

	// Test ListResourceTemplatesResult
	tmplRes := ListResourceTemplatesResult{ResourceTemplates: []ResourceTemplate{resTmpl}}
	b, _ = json.Marshal(tmplRes)
	if !strings.Contains(string(b), `"Tmpl"`) {
		t.Errorf("expected template in result")
	}

	// Test ResourceListChangedNotification
	listChangedNot := ResourceListChangedNotification{Method: "notifications/resources/list_changed"}
	b, _ = json.Marshal(listChangedNot)
	if !strings.Contains(string(b), `"notifications/resources/list_changed"`) {
		t.Errorf("expected list_changed notification")
	}

	// Test ResourceUpdatedNotification
	updatedNot := ResourceUpdatedNotification{Method: "notifications/resources/updated"}
	updatedNot.Params.URI = "file://updated"
	b, _ = json.Marshal(updatedNot)
	if !strings.Contains(string(b), `"file://updated"`) {
		t.Errorf("expected updated notification with uri")
	}

	// Test SubscribeRequest
	subReq := SubscribeRequest{Method: "resources/subscribe"}
	subReq.Params.URI = "file://sub"
	b, _ = json.Marshal(subReq)
	if !strings.Contains(string(b), `"resources/subscribe"`) || !strings.Contains(string(b), `"file://sub"`) {
		t.Errorf("expected subscribe request")
	}

	// Test UnsubscribeRequest
	unsubReq := UnsubscribeRequest{Method: "resources/unsubscribe"}
	unsubReq.Params.URI = "file://unsub"
	b, _ = json.Marshal(unsubReq)
	if !strings.Contains(string(b), `"resources/unsubscribe"`) || !strings.Contains(string(b), `"file://unsub"`) {
		t.Errorf("expected unsubscribe request")
	}
}
