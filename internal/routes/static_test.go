package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestEmbeddedStaticRoutes(t *testing.T) {
	files := fstest.MapFS{
		"index.html":    {Data: []byte("<html>application</html>")},
		"assets/app.js": {Data: []byte("console.log('app')")},
	}
	router := gin.New()
	router.GET("/api/health", func(context *gin.Context) {
		context.String(http.StatusOK, "healthy")
	})
	if err := RegisterStatic(router, files); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		method string
		path   string
		status int
		body   string
	}{
		{http.MethodGet, "/", http.StatusOK, "<html>application</html>"},
		{http.MethodGet, "/index.html", http.StatusOK, "<html>application</html>"},
		{http.MethodGet, "/settings/users?tab=roles", http.StatusOK, "<html>application</html>"},
		{http.MethodHead, "/settings/users", http.StatusOK, ""},
		{http.MethodGet, "/assets/app.js", http.StatusOK, "console.log('app')"},
		{http.MethodHead, "/assets/app.js", http.StatusOK, ""},
		{http.MethodGet, "/assets/missing.js", http.StatusNotFound, ""},
		{http.MethodGet, "/assets/missing", http.StatusNotFound, ""},
		{http.MethodGet, "/favicon.ico", http.StatusNotFound, ""},
		{http.MethodGet, "/api", http.StatusNotFound, ""},
		{http.MethodGet, "/api/missing", http.StatusNotFound, ""},
		{http.MethodPost, "/login", http.StatusNotFound, ""},
		{http.MethodGet, "/api/health", http.StatusOK, "healthy"},
	} {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
			if response.Code != test.status || response.Body.String() != test.body {
				t.Fatalf("got status %d, body %q; want status %d, body %q", response.Code, response.Body.String(), test.status, test.body)
			}
			if strings.Contains(test.body, "<html>") && !strings.HasPrefix(response.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("unexpected HTML content type: %s", response.Header().Get("Content-Type"))
			}
		})
	}
}

func TestEmbeddedStaticRequiresIndex(t *testing.T) {
	if err := RegisterStatic(gin.New(), fstest.MapFS{}); err == nil {
		t.Fatal("expected missing index.html to fail registration")
	}
}
