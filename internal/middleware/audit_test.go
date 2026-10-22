package middleware

import (
	"net/http"
	"testing"
)

func TestAuditActionSkipsListsAndRecordsMeaningfulOperations(t *testing.T) {
	tests := []struct {
		method string
		path   string
		action string
		ok     bool
	}{
		{http.MethodGet, "/api/users", "", false},
		{http.MethodGet, "/api/scripts/:id", "查看", true},
		{http.MethodGet, "/api/certificates/:id/download", "下载", true},
		{http.MethodPost, "/api/alert-rules", "创建", true},
		{http.MethodPost, "/api/alert-rules/:id/run", "执行", true},
		{http.MethodPut, "/api/mcp-clients/:id/status", "启停", true},
		{http.MethodPost, "/api/mcp", "调用", true},
		{http.MethodDelete, "/api/fault-centers/:id", "删除", true},
	}
	for _, test := range tests {
		action, ok := auditAction(test.method, test.path)
		if action != test.action || ok != test.ok {
			t.Fatalf("auditAction(%s, %s) = %q, %v; want %q, %v", test.method, test.path, action, ok, test.action, test.ok)
		}
	}
}

func TestSanitizeAuditMapRedactsNestedSecrets(t *testing.T) {
	value := sanitizeAuditMap(map[string]interface{}{
		"name": "production", "api_key": "model-secret", "password": "plain-text", "auth_config": map[string]interface{}{"token": "abc", "endpoint": "https://example.com"},
	})
	if value["api_key"] != "***" || value["password"] != "***" || value["auth_config"] != "***" {
		t.Fatalf("expected sensitive values to be redacted, got %#v", value)
	}
	if value["name"] != "production" {
		t.Fatalf("expected ordinary values to remain visible, got %#v", value)
	}
}

func TestAuditResponseResultUsesApplicationCode(t *testing.T) {
	result, message := auditResponseResult(http.StatusOK, []byte(`{"code":400,"message":"名称已存在"}`))
	if result != "失败" || message != "名称已存在" {
		t.Fatalf("unexpected application failure result: %q, %q", result, message)
	}
	result, _ = auditResponseResult(http.StatusOK, []byte(`{"code":0,"message":"成功"}`))
	if result != "成功" {
		t.Fatalf("expected successful result, got %q", result)
	}
}
