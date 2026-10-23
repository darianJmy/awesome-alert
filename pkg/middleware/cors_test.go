package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSUsesCallerPolicy(t *testing.T) {
	router := gin.New()
	router.Use(CORS("https://example.com", "GET,OPTIONS", "Authorization"))
	reached := false
	router.GET("/", func(c *gin.Context) { reached = true; c.Status(200) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodOptions, "/", nil))
	if response.Code != 204 || reached {
		t.Fatal("preflight did not stop middleware chain")
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "https://example.com" || response.Header().Get("Access-Control-Allow-Headers") != "Authorization" {
		t.Fatal("caller policy not applied")
	}
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !reached {
		t.Fatal("normal request did not reach handler")
	}
}
