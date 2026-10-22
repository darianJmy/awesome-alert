package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	aicore "guanlan/internal/ai"
	"guanlan/internal/ai/provider"
	"guanlan/internal/database"
	"guanlan/internal/models"
)

func datasourceToolDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	previous := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previous; conn.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.Datasource{}, &models.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		`INSERT INTO users(id,username,password,status) VALUES(1,'creator','unused',1),(2,'reader','unused',1)`,
		`INSERT INTO roles(id,name,code,status) VALUES(1,'creator','creator',1),(2,'reader','reader',1)`,
		`INSERT INTO permissions(id,name,code,status) VALUES(1,'list','datasource:list',1),(2,'create','datasource:create',1),(3,'run','automation-task:run',1)`,
		`INSERT INTO user_roles(user_id,role_id) VALUES(1,1),(2,2)`,
		`INSERT INTO role_permissions(role_id,permission_id) VALUES(1,1),(1,2),(1,3),(2,1)`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestDatasourceMCPCreatePermissionsAndValidation(t *testing.T) {
	db := datasourceToolDB(t)
	registry := aicore.NewRegistry()
	registerAssistantMCPTools(context.Background(), registry, 1)
	tool, ok := registry.Get("create_datasource")
	if !ok || tool.Spec.Risk != aicore.RiskWriteLow {
		t.Fatal("creation not registered")
	}
	for _, field := range []string{"name", "type", "url", "auth_type"} {
		if !tool.Spec.ParamsSchema[field].Required {
			t.Fatalf("%s not required", field)
		}
	}
	if _, ok := registry.Get("run_automation_task"); ok {
		t.Fatal("unrelated write tool exposed")
	}
	reader := aicore.NewRegistry()
	registerAssistantMCPTools(context.Background(), reader, 2)
	if _, ok := reader.Get("create_datasource"); ok {
		t.Fatal("reader received create tool")
	}
	args := map[string]interface{}{"name": "test metrics", "type": "prometheus", "url": "http://metrics.example:9090", "auth_type": "none"}
	if _, err := executeDatasourceMCPTool(context.Background(), 2, "create_datasource", args); err == nil {
		t.Fatal("reader created datasource")
	}
	for _, change := range []map[string]interface{}{
		{"url": ""}, {"url": "http://user:secret@example.com"}, {"type": "mysql"}, {"auth_type": "basic"}, {"auth_config": "secret"}, {"user_id": 1}, {"url": 123},
	} {
		bad := map[string]interface{}{}
		for k, v := range args {
			bad[k] = v
		}
		for k, v := range change {
			bad[k] = v
		}
		if _, err := executeDatasourceMCPTool(context.Background(), 1, "create_datasource", bad); err == nil {
			t.Fatal("invalid arguments accepted", change)
		}
	}
	result, err := executeDatasourceMCPTool(context.Background(), 1, "create_datasource", args)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), `"id":1`) || strings.Contains(string(raw), "metrics.example") {
		t.Fatalf("invalid projection %s", raw)
	}
	var ds models.Datasource
	if err := db.First(&ds).Error; err != nil || ds.URL != "http://metrics.example:9090" || ds.Status != 1 || ds.Capabilities != "metrics" {
		t.Fatalf("create did not persist: %+v %v", ds, err)
	}
	var audit models.AuditLog
	if err := db.First(&audit).Error; err != nil || audit.UserID != 1 || audit.ResourceID != "1" {
		t.Fatalf("missing audit: %+v %v", audit, err)
	}
	if _, err := executeDatasourceMCPTool(context.Background(), 1, "create_datasource", args); err == nil {
		t.Fatal("duplicate creation accepted")
	}
	db.Model(&ds).Updates(map[string]interface{}{"auth_type": "bearer", "auth_config": `{"token":"hidden-secret"}`})
	queried, err := executeDatasourceMCPTool(context.Background(), 2, "query_datasources", map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(queried)
	if strings.Contains(string(raw), "hidden-secret") || strings.Contains(string(raw), "metrics.example") {
		t.Fatal("query leaked connection settings")
	}
	// The remote MCP adapter uses the same implementation and permission checks.
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/mcp", nil)
	c.Set("user_id", uint(2))
	params, _ := json.Marshal(map[string]interface{}{"name": "query_datasources", "arguments": map[string]interface{}{}})
	mcpCallTool(c, mcpRequest{ID: json.RawMessage(`1`), Params: params})
	if !strings.Contains(recorder.Body.String(), "test metrics") {
		t.Fatal(recorder.Body.String())
	}
	db.Exec("UPDATE users SET status=0 WHERE id=1")
	args["name"] = "after revocation"
	if result := tool.Handler(context.Background(), aicore.ToolCall{Tool: "create_datasource", Params: args}); result.OK {
		t.Fatal("revoked user can still create")
	}
	var count int64
	db.Model(&models.Datasource{}).Count(&count)
	if count != 1 {
		t.Fatalf("unexpected datasource count %d", count)
	}
}

func TestDatasourceMCPConnection(t *testing.T) {
	db := datasourceToolDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer hidden-secret" {
			t.Error("stored authentication not used")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1,"1"]}]}}`))
	}))
	defer server.Close()
	ds := models.Datasource{Name: "vm", Type: "victoriametrics", URL: server.URL, AuthType: "bearer", AuthConfig: `{"token":"hidden-secret"}`, Status: 1}
	if err := db.Create(&ds).Error; err != nil {
		t.Fatal(err)
	}
	result, err := executeDatasourceMCPTool(context.Background(), 2, "test_datasource_connection", map[string]interface{}{"datasource_id": ds.ID})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), `"connected":true`) || strings.Contains(string(raw), "hidden-secret") {
		t.Fatalf("invalid result: %s", raw)
	}
}

func TestDatasourceMCPAssistantRoundTrip(t *testing.T) {
	db := datasourceToolDB(t)
	registry := aicore.NewRegistry()
	registerAssistantMCPTools(context.Background(), registry, 1)
	rounds := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rounds++
		var request map[string]interface{}
		json.NewDecoder(r.Body).Decode(&request)
		if rounds == 1 {
			found := false
			for _, raw := range request["tools"].([]interface{}) {
				item := raw.(map[string]interface{})
				if item["function"].(map[string]interface{})["name"] == "create_datasource" {
					found = true
				}
			}
			if !found {
				t.Error("write tool not advertised to model")
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{map[string]interface{}{"message": map[string]interface{}{"role": "assistant", "tool_calls": []interface{}{map[string]interface{}{"id": "create_1", "type": "function", "function": map[string]interface{}{"name": "create_datasource", "arguments": `{"name":"prom-test","type":"prometheus","url":"http://metrics.example:9090","auth_type":"none"}`}}}}, "finish_reason": "tool_calls"}}})
			return
		}
		raw, _ := json.Marshal(request)
		if !strings.Contains(string(raw), "prom-test") {
			t.Error("creation result missing")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{map[string]interface{}{"message": map[string]interface{}{"role": "assistant", "content": "已创建 prom-test，ID 为 1。"}, "finish_reason": "stop"}}})
	}))
	defer server.Close()
	chatModel, err := provider.New(context.Background(), provider.Config{Provider: "newapi", BaseURL: server.URL, APIKey: "test-key", Model: "test"})
	if err != nil {
		t.Fatal(err)
	}
	orchestrator := aicore.Orchestrator{Registry: registry, Model: chatModel, Authorize: aiToolPolicy(1), MaxRounds: 3}
	response := orchestrator.Run(context.Background(), aicore.RunRequest{Question: "创建无认证直连 Prometheus 数据源 prom-test，地址 http://metrics.example:9090"})
	var count int64
	db.Model(&models.Datasource{}).Where("name = ?", "prom-test").Count(&count)
	if count != 1 || rounds != 2 || !strings.Contains(response.Answer, "已创建") {
		t.Fatalf("round trip failed: %+v rounds=%d count=%d", response, rounds, count)
	}
}
