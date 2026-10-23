package routes

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterStatic(router *gin.Engine, files fs.FS) error {
	index, err := fs.ReadFile(files, "index.html")
	if err != nil {
		return fmt.Errorf("read web index: %w", err)
	}
	fileServer := http.FileServer(http.FS(files))
	router.NoRoute(func(context *gin.Context) {
		requestPath := path.Clean("/" + context.Request.URL.Path)
		if (context.Request.Method != http.MethodGet && context.Request.Method != http.MethodHead) ||
			requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") {
			context.AbortWithStatus(http.StatusNotFound)
			return
		}
		name := strings.TrimPrefix(requestPath, "/")
		if name != "" && name != "index.html" {
			info, err := fs.Stat(files, name)
			if err == nil && !info.IsDir() {
				context.Status(http.StatusOK)
				fileServer.ServeHTTP(context.Writer, context.Request)
				return
			}
			if path.Ext(name) != "" || name == "assets" || strings.HasPrefix(name, "assets/") {
				context.AbortWithStatus(http.StatusNotFound)
				return
			}
		}
		context.Header("Cache-Control", "no-cache")
		context.Header("Content-Type", "text/html; charset=utf-8")
		context.Status(http.StatusOK)
		if context.Request.Method == http.MethodGet {
			_, _ = context.Writer.Write(index)
		}
	})
	return nil
}
