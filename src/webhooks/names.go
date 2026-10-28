// Package webhooks provides parsing, emitting, and mock trigger dispatching of OpenAPI Webhooks and Callbacks.
package webhooks

import (
	"sort"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// CollectWebhookNames extracts all sorted unique webhook and callback names defined in an OpenAPI specification.
// Parameter oa is the OpenAPI specification to inspect, and it returns the sorted slice of names.
func CollectWebhookNames(oa *openapi.OpenAPI) []string {
	if oa == nil {
		return nil
	}
	seen := make(map[string]bool)
	for name := range oa.Webhooks {
		seen[name] = true
	}
	if oa.Components != nil {
		for name := range oa.Components.Callbacks {
			seen[name] = true
		}
	}
	for _, pi := range oa.Paths {
		for _, mo := range openapi.PathItemOperations(&pi) {
			for name := range mo.Operation.Callbacks {
				seen[name] = true
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasWebhooksOrCallbacks reports whether the OpenAPI specification defines any webhooks or callbacks.
// Parameter oa is the OpenAPI specification to inspect.
func HasWebhooksOrCallbacks(oa *openapi.OpenAPI) bool {
	return len(CollectWebhookNames(oa)) > 0
}
