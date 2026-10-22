package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"guanlan/internal/models"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIModelManagementAndSelection(t *testing.T) {
	db := setupAIModelTestDB(t)
	if err := db.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"INSERT INTO users (id,username,password,status) VALUES (1,'admin','unused',1),(2,'user','unused',1)",
		"INSERT INTO roles (id,name,code,status) VALUES (1,'admin','admin',1)",
		"INSERT INTO permissions (id,name,code,status) VALUES (1,'list','system-setting:list',1),(2,'edit','system-setting:edit',1)",
		"INSERT INTO user_roles (user_id,role_id) VALUES (1,1)",
		"INSERT INTO role_permissions (role_id,permission_id) VALUES (1,1),(1,2)",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	userID := uint(1)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("user_id", userID); c.Next() })
	r.POST("/models", SaveAIModel)
	r.PUT("/models/:id", SaveAIModel)
	r.DELETE("/models/:id", DeleteAIModel)
	r.GET("/models", ListAIModels)
	r.GET("/choices", ListChatModels)
	request := func(method, path string, body map[string]interface{}) *httptest.ResponseRecorder {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	body := map[string]interface{}{"name": "primary", "provider": "newapi", "base_url": "https://model.example.com/v1", "model": "model-one", "api_key": "private-key", "status": 1}
	w := request("POST", "/models", body)
	if w.Code != 200 || strings.Contains(w.Body.String(), "private-key") {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var first models.AIModel
	db.First(&first)
	if first.SecretCipher == "private-key" || first.SecretCipher == "" {
		t.Fatal("secret stored in plaintext")
	}
	cfg, err := aiModelConfig(first)
	if err != nil || cfg.APIKey != "private-key" {
		t.Fatal("cannot decrypt saved key")
	}
	body["api_key"] = ""
	body["model"] = "model-two"
	if w = request("PUT", "/models/1", body); w.Code != 200 {
		t.Fatalf("update: %s", w.Body)
	}
	db.First(&first)
	if first.SecretCipher == "" || first.Model != "model-two" {
		t.Fatal("update lost key/model")
	}
	body["base_url"] = "https://other.example.com"
	if w = request("PUT", "/models/1", body); w.Code != 400 {
		t.Fatal("endpoint change reused secret")
	}
	userID = 2
	if w = request("GET", "/models", nil); w.Code != 403 {
		t.Fatal("non-admin read config")
	}
	if w = request("POST", "/models", body); w.Code != 403 {
		t.Fatal("non-admin changed config")
	}
	w = request("GET", "/choices", nil)
	if w.Code != 200 || strings.Contains(w.Body.String(), "base_url") || strings.Contains(w.Body.String(), "private-key") || strings.Contains(w.Body.String(), "secret_cipher") {
		t.Fatalf("unsafe choices: %s", w.Body)
	}
	userID = 1
	body["base_url"] = "https://model.example.com/v1"
	body["status"] = 0
	if w = request("PUT", "/models/1", body); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if _, err := resolveAIModel(httptest.NewRequest("GET", "/", nil).Context(), 1); err == nil {
		t.Fatal("disabled model accepted")
	}
	if w = request("GET", "/choices", nil); !strings.Contains(w.Body.String(), `"data":[]`) {
		t.Fatal("disabled choice exposed")
	}
	if w = request("DELETE", "/models/1", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
}
