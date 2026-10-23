package config

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfigCommandLine(t *testing.T) {
	t.Chdir(t.TempDir())
	previousArgs, previousConfig := os.Args, AppConfig
	t.Cleanup(func() { os.Args, AppConfig = previousArgs, previousConfig })
	path := filepath.Join(t.TempDir(), "custom.yaml")
	if err := os.WriteFile(path, []byte("server:\n  addr: ':18081'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// Environment values are intentionally higher priority than the selected file.
	t.Setenv("APP_SERVER_ADDR", ":18082")
	os.Args = []string{"guanlan", "--conf", path}
	cfg, err := InitConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != ":18082" || AppConfig != cfg {
		t.Fatalf("configuration not published correctly: %#v", cfg)
	}
	for _, args := range [][]string{{"--help"}, {"--unknown"}, {"--conf"}, {"unexpected"}} {
		os.Args = append([]string{"guanlan"}, args...)
		_, err := InitConfig()
		if err == nil {
			t.Fatalf("expected error for %v", args)
		}
		if args[0] == "--help" && !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected help sentinel: %v", err)
		}
		if AppConfig != cfg {
			t.Fatal("failed argument parsing changed shared configuration")
		}
	}
}
