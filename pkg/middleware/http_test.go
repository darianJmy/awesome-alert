package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func capture(t *testing.T, level slog.Level) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previous := slog.Default()
	previousWriter, previousFlags := log.Writer(), log.Flags()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: level})))
	t.Cleanup(func() { slog.SetDefault(previous); log.SetOutput(previousWriter); log.SetFlags(previousFlags) })
	return &output
}

func TestHTTPAccessAndRecovery(t *testing.T) {
	output := capture(t, slog.LevelDebug)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Access(), Recovery())
	router.GET("/items/:id", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/panic", func(c *gin.Context) { panic("test panic") })
	request := httptest.NewRequest(http.MethodGet, "/items/123?token=secret-query", nil)
	request.Header.Set("Authorization", "Bearer secret-header")
	router.ServeHTTP(httptest.NewRecorder(), request)
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["path"] != "/items/:id" || record["status"] != float64(200) {
		t.Fatalf("unexpected access log: %v", record)
	}
	if strings.Contains(output.String(), "secret-") {
		t.Fatal("HTTP secrets leaked")
	}
	output.Reset()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if response.Code != 500 || !strings.Contains(output.String(), `"msg":"http panic"`) || !strings.Contains(output.String(), `"status":500`) {
		t.Fatalf("panic recovery failed: %d %s", response.Code, output.String())
	}
}
