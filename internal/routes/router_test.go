package routes

import (
	"strings"
	"testing"
)

func TestCertificateModuleRemovedAndSharedNodesRetained(t *testing.T) {
	paths := map[string]bool{}
	for _, route := range SetupRouter().Routes() {
		paths[route.Path] = true
		for _, prefix := range []string{"/api/certificates", "/api/certificate-domains", "/api/certificate-rotation-records"} {
			if strings.HasPrefix(route.Path, prefix) {
				t.Fatalf("retired endpoint still registered: %s", route.Path)
			}
		}
	}
	for _, path := range []string{"/api/certificate-executors", "/api/agents", "/api/inspection-items"} {
		if !paths[path] {
			t.Fatalf("shared capability missing: %s", path)
		}
	}
}
