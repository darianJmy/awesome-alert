package handlers

import (
	"context"
	"encoding/json"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"guanlan/internal/database"
	"guanlan/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLLMGatewayProtocols(t *testing.T) {
	db := setupAIModelTestDB(t)
	for _, provider := range []string{"anthropic", "newapi", "openai"} {
		for _, suffix := range []string{"", "/v1"} {
			t.Run(provider+suffix, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					var body map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["model"] != "test-model" {
						t.Error("model missing")
					}
					if provider == "anthropic" {
						if r.URL.Path != "/proxy/v1/messages" || r.Header.Get("x-api-key") != "test-key" {
							t.Error("wrong Anthropic request")
						}
						w.Write([]byte(`{"id":"msg1","type":"message","role":"assistant","model":"test-model","content":[{"type":"text","text":"answer"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
					} else {
						if r.URL.Path != "/proxy/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer test-key" {
							t.Error("wrong gateway request")
						}
						messages := body["messages"].([]interface{})
						if messages[0].(map[string]interface{})["role"] != "system" {
							t.Error("system prompt missing")
						}
						w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"answer"},"finish_reason":"stop"}]}`))
					}
				}))
				defer server.Close()
				encrypted, err := encryptSecret("test-key")
				if err != nil {
					t.Fatal(err)
				}
				db.Exec("DELETE FROM ai_models")
				if err := db.Create(&models.AIModel{Name: "test", Provider: provider, BaseURL: server.URL + "/proxy" + suffix, Model: "test-model", SecretCipher: encrypted, Status: 1}).Error; err != nil {
					t.Fatal(err)
				}
				got, err := callLLM(context.Background(), "system", []anthropicMessage{{Role: "user", Content: "question"}})
				if err != nil || got != "answer" {
					t.Fatalf("got %q, %v", got, err)
				}
			})
		}
	}
}

func TestLLMDoesNotExposeProviderErrors(t *testing.T) {
	db := setupAIModelTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`secret-provider-key`))
	}))
	defer server.Close()
	encrypted, err := encryptSecret("test-key")
	if err != nil {
		t.Fatal(err)
	}
	db.Create(&models.AIModel{Name: "test", Provider: "newapi", BaseURL: server.URL, Model: "test", SecretCipher: encrypted, Status: 1})
	_, err = callLLM(context.Background(), "", nil)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := callLLM(ctx, "", nil); err == nil {
		t.Fatal("cancel ignored")
	}
}

func setupAIModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	conn, _ := db.DB()
	conn.SetMaxOpenConns(1)
	old := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = old; conn.Close() })
	if err := db.AutoMigrate(&models.AIModel{}); err != nil {
		t.Fatal(err)
	}
	return db
}
