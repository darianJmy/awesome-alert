package handlers

import (
	"encoding/json"
	"testing"
)

func TestDashboardUpdateAcceptsSortOnly(t *testing.T) {
	var req UpdateDashboardReq
	if err := json.Unmarshal([]byte(`{"sort":2}`), &req); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	updates, err := dashboardUpdateFields(req)
	if err != nil {
		t.Fatalf("build updates: %v", err)
	}
	if len(updates) != 1 || updates["sort"] != 2 {
		t.Fatalf("unexpected updates: %#v", updates)
	}
}

func TestDashboardUpdateRejectsBlankRequiredFields(t *testing.T) {
	blank := "  "
	if _, err := dashboardUpdateFields(UpdateDashboardReq{Name: &blank}); err == nil {
		t.Fatal("expected blank name to be rejected")
	}
	if _, err := dashboardUpdateFields(UpdateDashboardReq{URL: &blank}); err == nil {
		t.Fatal("expected blank URL to be rejected")
	}
}

func TestDashboardProviderValidation(t *testing.T) {
	if provider, err := normalizeDashboardProvider(""); err != nil || provider != "grafana" {
		t.Fatalf("empty provider should default to grafana, got %q, %v", provider, err)
	}
	if provider, err := normalizeDashboardProvider("ALIYUN_SLS"); err != nil || provider != "aliyun_sls" {
		t.Fatalf("aliyun sls provider was not normalized: %q, %v", provider, err)
	}
	if _, err := normalizeDashboardProvider("unknown"); err == nil {
		t.Fatal("unsupported provider accepted")
	}
}
