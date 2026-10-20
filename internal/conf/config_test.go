package conf

import (
	"os"
	"path/filepath"
	"testing"
)

func isolate(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, binding := range bindings {
		for _, name := range binding.names {
			t.Setenv(name, "")
			if err := os.Unsetenv(name); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func writeConfig(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultsAndOptionalLocalFile(t *testing.T) {
	isolate(t)
	cfg, err := InitConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != DefaultConfig() {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	writeConfig(t, "conf/config.yaml", "server:\n  addr: ':9000'\nai:\n  model: local-model\n")
	cfg, err = InitConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != ":9000" || cfg.AnthropicModel != "local-model" || cfg.DBDSN != DefaultConfig().DBDSN {
		t.Fatalf("local config did not merge with defaults: %+v", cfg)
	}
}

func TestExplicitConfigAndEnvironmentPrecedence(t *testing.T) {
	isolate(t)
	writeConfig(t, "conf/config.yaml", "server:\n  addr: ':9000'\n")
	path := filepath.Join(t.TempDir(), "custom.yaml")
	writeConfig(t, path, "server:\n  addr: ':9001'\nauth:\n  jwt_secret: file-secret\ndatabase:\n  dsn: file-dsn\nai:\n  base_url: https://example.com\n  api_key: file-key\n  model: file-model\n")
	cfg, err := InitConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != ":9001" || cfg.JWTSecret != "file-secret" || cfg.DBDSN != "file-dsn" || cfg.AnthropicBaseURL != "https://example.com" || cfg.AnthropicAPIKey != "file-key" || cfg.AnthropicModel != "file-model" {
		t.Fatalf("explicit config not loaded: %+v", cfg)
	}
	writeConfig(t, ".env", "APP_SERVER_ADDR=:9002\nAPP_AI_MODEL=dotenv-model\nAPP_DATABASE_DSN=dotenv-dsn\n")
	t.Setenv("SERVER_PORT", ":9003")
	t.Setenv("DB_DSN", "legacy-dsn")
	t.Setenv("APP_DATABASE_DSN", "prefixed-dsn")
	t.Setenv("APP_AI_API_KEY", "")
	cfg, err = InitConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerPort != ":9003" || cfg.DBDSN != "prefixed-dsn" || cfg.AnthropicModel != "dotenv-model" || cfg.AnthropicAPIKey != "" {
		t.Fatalf("incorrect precedence: %+v", cfg)
	}
	if _, ok := os.LookupEnv("APP_AI_MODEL"); ok {
		t.Fatal(".env must not mutate process environment")
	}
}

func TestLegacyEnvironmentBindings(t *testing.T) {
	isolate(t)
	t.Setenv("SERVER_PORT", "127.0.0.1:9090")
	t.Setenv("JWT_SECRET", "legacy-secret")
	t.Setenv("DB_DSN", "legacy-dsn")
	t.Setenv("ANTHROPIC_BASE_URL", "https://legacy.example.com")
	t.Setenv("ANTHROPIC_API_KEY", "legacy-key")
	t.Setenv("ANTHROPIC_MODEL", "legacy-model")
	cfg, err := InitConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != (Config{Log: DefaultConfig().Log, ServerPort: "127.0.0.1:9090", JWTSecret: "legacy-secret", DBDSN: "legacy-dsn", AnthropicBaseURL: "https://legacy.example.com", AnthropicAPIKey: "legacy-key", AnthropicModel: "legacy-model"}) {
		t.Fatalf("legacy binding failed: %+v", cfg)
	}
}

func TestConfigErrors(t *testing.T) {
	for _, test := range []struct{ name, data string }{
		{"log-level", "log:\n  level: invalid"}, {"log-format", "log:\n  format: invalid"},
		{"missing", ""}, {"malformed", "server: ["}, {"type", "server:\n  addr: [x, y]"},
		{"address", "server:\n  addr: invalid"}, {"port", "server:\n  addr: ':70000'"},
		{"dsn", "database:\n  dsn: ''"}, {"secret", "auth:\n  jwt_secret: ''"},
	} {
		t.Run(test.name, func(t *testing.T) {
			isolate(t)
			if test.name != "missing" {
				writeConfig(t, "bad.yaml", test.data)
			}
			if _, err := InitConfig("bad.yaml"); err == nil {
				t.Fatal("expected config error")
			}
		})
	}
}

func TestRepeatedLoadsDoNotRetainOverrides(t *testing.T) {
	isolate(t)
	writeConfig(t, "custom.yaml", "server:\n  addr: ':9090'")
	if _, err := InitConfig("custom.yaml"); err != nil {
		t.Fatal(err)
	}
	cfg, err := InitConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg != DefaultConfig() {
		t.Fatal("configuration leaked between loads")
	}
}

func TestLogConfigEnvironmentOverridesFile(t *testing.T) {
	isolate(t)
	writeConfig(t, "conf/config.yaml", "log:\n  level: warn\n  format: text\n  file: logs/file.log\n")
	t.Setenv("APP_LOG_LEVEL", "DEBUG")
	t.Setenv("APP_LOG_FORMAT", "JSON")
	t.Setenv("APP_LOG_FILE", "logs/env.log")
	cfg, err := InitConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Log.Level != "debug" || cfg.Log.Format != "json" || cfg.Log.File != "logs/env.log" {
		t.Fatalf("incorrect logging config: %+v", cfg.Log)
	}
}
