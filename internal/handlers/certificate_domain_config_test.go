package handlers

import (
	"strings"
	"testing"
)

func TestParseCertificateDomainConfigDocumentSupportsYAML(t *testing.T) {
	content := `apiVersion: aiops/v1
kind: CertificateDomainList
domains:
  - domain: example.com
    port: 443
    fault_center_id: 1
    monitor_enabled: false
    check_interval_minutes: 360
    expiry_alert_days: 30
    severity: P2`
	document, err := parseCertificateDomainConfigDocument(content)
	if err != nil {
		t.Fatalf("parse YAML config: %v", err)
	}
	if len(document.Domains) != 1 || document.Domains[0].MonitorEnabled == nil || *document.Domains[0].MonitorEnabled {
		t.Fatalf("unexpected document: %#v", document)
	}
}

func TestParseCertificateDomainConfigDocumentRejectsJSON(t *testing.T) {
	_, err := parseCertificateDomainConfigDocument(`{"apiVersion":"aiops/v1","kind":"CertificateDomainList","domains":[]}`)
	if err == nil || !strings.Contains(err.Error(), "仅支持 YAML") {
		t.Fatalf("expected JSON to be rejected, got %v", err)
	}
}

func TestParseCertificateDomainConfigDocumentRejectsUnknownFields(t *testing.T) {
	_, err := parseCertificateDomainConfigDocument("apiVersion: aiops/v1\nkind: CertificateDomainList\nunknown: true\ndomains: []")
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("expected unknown field to be rejected, got %v", err)
	}
}
