package handlers

import (
	"strings"
	"testing"
)

func TestParseInspectionTaskConfigDocumentSupportsYAML(t *testing.T) {
	yamlContent := `apiVersion: aiops/v1
kind: InspectionTaskList
tasks:
  - name: HTTP 健康检查
    type: http
    target: https://example.com/health
    config:
      execution_nodes:
        - node_type: direct
          node_id: 0
    interval_seconds: 60
    failure_threshold: 3
    recovery_threshold: 2
    severity: P1
    fault_center_id: 1
    status: 1`
	document, err := parseInspectionTaskConfigDocument(yamlContent)
	if err != nil {
		t.Fatalf("parse YAML config: %v", err)
	}
	if len(document.Tasks) != 1 || document.Tasks[0].Name != "HTTP 健康检查" {
		t.Fatalf("unexpected YAML document: %#v", document)
	}
}

func TestParseInspectionTaskConfigDocumentRejectsJSON(t *testing.T) {
	content := `{"apiVersion":"aiops/v1","kind":"InspectionTaskList","tasks":[]}`
	_, err := parseInspectionTaskConfigDocument(content)
	if err == nil || !strings.Contains(err.Error(), "仅支持 YAML") {
		t.Fatalf("expected JSON config to be rejected, got %v", err)
	}
}

func TestParseInspectionTaskConfigDocumentRejectsUnknownFields(t *testing.T) {
	_, err := parseInspectionTaskConfigDocument("apiVersion: aiops/v1\nkind: InspectionTaskList\nunknown: true\ntasks: []")
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("expected unknown YAML field to be rejected, got %v", err)
	}
}

func TestInspectionTaskConfigDocumentRejectsWrongKindAndEmptyTasks(t *testing.T) {
	for _, content := range []string{
		"apiVersion: aiops/v2\nkind: InspectionTaskList\ntasks: []",
		"apiVersion: aiops/v1\nkind: InspectionTaskList\ntasks: []",
	} {
		if _, err := parseInspectionTaskConfigDocument(content); err == nil {
			t.Fatalf("expected invalid document to be rejected: %s", content)
		}
	}
}
