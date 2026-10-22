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

func TestScheduledTaskEndpointsRemoved(t *testing.T) {
	paths := map[string]bool{}
	for _, route := range SetupRouter().Routes() {
		paths[route.Path] = true
		for _, prefix := range []string{"/api/scheduled-tasks", "/api/scheduled-task-executions", "/api/cron-help"} {
			if strings.HasPrefix(route.Path, prefix) {
				t.Fatalf("retired endpoint still registered: %s", route.Path)
			}
		}
	}
	for _, path := range []string{"/api/inspection-items", "/api/instant-probe", "/api/scripts", "/api/automation-tasks"} {
		if !paths[path] {
			t.Fatalf("required endpoint missing: %s", path)
		}
	}
}

func TestNonPrometheusDatasourceEndpointsRemoved(t *testing.T) {
	for _, route := range SetupRouter().Routes() {
		if strings.HasPrefix(route.Path, "/api/k8s/") || strings.Contains(route.Path, "/kafka/") {
			t.Fatalf("retired datasource endpoint still registered: %s", route.Path)
		}
	}
}
