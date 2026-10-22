package main

import (
	"embed"
	"errors"
	"flag"
	"guanlan/internal/app"
	"guanlan/pkg/logger"
	"os"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	a, cleanup, err := app.InitializeApp(staticFiles)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
	defer cleanup()
	if err := a.Run(); err != nil {
		logger.Error("server exited", "error", err)
		// os.Exit does not run deferred functions.
		cleanup()
		os.Exit(1)
	}
}
