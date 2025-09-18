package main

import (
	"os"

	"awesome-alert/cmd/app"
)

func main() {
	cmd := app.NewServerCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
