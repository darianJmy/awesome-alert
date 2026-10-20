package handlers

import (
	"guanlan/internal/models"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestChatConversationScopeAlwaysFiltersUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("create dry-run database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	query := chatConversationScope(db.Model(&models.ChatConversation{}), 42).
		Where("id = ?", 7).
		Find(&models.ChatConversation{})
	statement := query.Statement.SQL.String()
	if !strings.Contains(statement, "user_id") {
		t.Fatalf("conversation query is not user-scoped: %s", statement)
	}
	if len(query.Statement.Vars) != 2 || query.Statement.Vars[0] != uint(42) || query.Statement.Vars[1] != 7 {
		t.Fatalf("unexpected conversation scope vars: %#v", query.Statement.Vars)
	}
}

func TestExtractDatasourceReq(t *testing.T) {
	req := extractDatasourceReq("帮我新增一个 Prometheus 数据源，名字叫生产 prometheus，地址是 http://10.0.0.1:9090")

	if req.Type != "prometheus" {
		t.Fatalf("expected prometheus type, got %q", req.Type)
	}
	if req.Name != "生产 prometheus" {
		t.Fatalf("expected full datasource name, got %q", req.Name)
	}
	if req.URL != "http://10.0.0.1:9090" {
		t.Fatalf("expected url without trailing punctuation, got %q", req.URL)
	}
	if req.AuthType != "none" {
		t.Fatalf("expected default auth type none, got %q", req.AuthType)
	}
}

func TestSelectPromDatasourcesByQuestionIgnoresSpaces(t *testing.T) {
	ds := []models.Datasource{
		{Name: "个人prometheus", Type: "prometheus", URL: "http://127.0.0.1:9090"},
		{Name: "生产prometheus", Type: "prometheus", URL: "http://10.0.0.1:9090"},
	}

	matched := selectPromDatasourcesByQuestion("用个人 prometheus 看一下系统性能", ds, "")
	if len(matched) != 1 {
		t.Fatalf("expected one datasource, got %d", len(matched))
	}
	if matched[0].Name != "个人prometheus" {
		t.Fatalf("expected personal prometheus, got %q", matched[0].Name)
	}
}

func TestMissingDatasourceFields(t *testing.T) {
	missing := missingDatasourceFields(CreateDatasourceWithTunnelReq{Type: "prometheus"})
	if len(missing) != 2 || missing[0] != "名称" || missing[1] != "地址" {
		t.Fatalf("unexpected missing fields: %#v", missing)
	}
}

func TestExtractDatasourceMemory(t *testing.T) {
	key, value := extractDatasourceMemory("记住默认 Prometheus 数据源是个人prometheus")
	if key != "default_prometheus_datasource" {
		t.Fatalf("unexpected memory key: %q", key)
	}
	if value != "个人prometheus" {
		t.Fatalf("unexpected memory value: %q", value)
	}
}
