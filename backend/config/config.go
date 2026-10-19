package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	ServerPort       string
	JWTSecret        string
	DBDSN            string
	AnthropicBaseURL string
	AnthropicAPIKey  string
	AnthropicModel   string
}

var AppConfig = loadConfig()

func loadConfig() Config {
	loadEnvFile(".env")
	loadEnvFile("backend/.env")

	return Config{
		ServerPort:       getenv("SERVER_PORT", ":8080"),
		JWTSecret:        getenv("JWT_SECRET", "aiops-jwt-secret-key-2024"),
		DBDSN:            getenv("DB_DSN", "root@tcp(127.0.0.1:3306)/aiops?charset=utf8mb4&parseTime=True&loc=Local"),
		AnthropicBaseURL: getenv("ANTHROPIC_BASE_URL", "https://dreame-claude.dreame.tech"),
		AnthropicAPIKey:  getenv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:   getenv("ANTHROPIC_MODEL", "deepseek-v4-pro"),
	}
}

func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
