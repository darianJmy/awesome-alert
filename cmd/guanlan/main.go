package main

import (
	"errors"
	"flag"
	"guanlan/internal/app"
	"guanlan/pkg/logger"
	"os"
)

func main() {
	a, cleanup, err := app.InitializeApp()
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
