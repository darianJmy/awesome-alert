package conf

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"guanlan/pkg/logger"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

//go:embed conf/config.yaml
var configYaml []byte

type LogConfig = logger.Config

type Config struct {
	Log        LogConfig
	ServerPort string
	JWTSecret  string
	DBDSN      string
}

type fileConfig struct {
	Log    LogConfig `mapstructure:"log"`
	Server struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"server"`
	Auth struct {
		JWTSecret string `mapstructure:"jwt_secret"`
	} `mapstructure:"auth"`
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
}

var bindings = []struct {
	key   string
	names []string
}{
	{"log.level", []string{"APP_LOG_LEVEL"}},
	{"log.format", []string{"APP_LOG_FORMAT"}},
	{"log.file", []string{"APP_LOG_FILE"}},
	{"server.addr", []string{"APP_SERVER_ADDR", "SERVER_PORT"}},
	{"auth.jwt_secret", []string{"APP_AUTH_JWT_SECRET", "JWT_SECRET"}},
	{"database.dsn", []string{"APP_DATABASE_DSN", "DB_DSN"}},
}

func defaults() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(configYaml)); err != nil {
		return nil, fmt.Errorf("read embedded configuration: %w", err)
	}
	return v, nil
}

// DefaultConfig reads only embedded defaults, without inspecting files or the environment.
func DefaultConfig() Config {
	v, err := defaults()
	if err != nil {
		panic(err)
	}
	cfg, err := decode(v)
	if err != nil {
		panic(err)
	}
	return cfg
}

// InitConfig loads defaults < external YAML < .env < process environment.
// An empty path makes conf/config.yaml optional; an explicit path is required.
func InitConfig(path string) (Config, error) {
	v, err := defaults()
	if err != nil {
		return Config{}, err
	}
	explicit := path != ""
	if !explicit {
		path = "conf/config.yaml"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if explicit || !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read configuration %q: %w", path, err)
		}
	} else if err := v.MergeConfig(bytes.NewReader(data)); err != nil {
		return Config{}, fmt.Errorf("parse configuration %q: %w", path, err)
	}
	envFile, err := readEnvFile(".env")
	if err != nil {
		return Config{}, err
	}
	v.AllowEmptyEnv(true)
	for _, binding := range bindings {
		if err := v.BindEnv(append([]string{binding.key}, binding.names...)...); err != nil {
			return Config{}, fmt.Errorf("bind environment %s: %w", binding.key, err)
		}
		hasEnv := false
		for _, name := range binding.names {
			if _, ok := os.LookupEnv(name); ok {
				hasEnv = true
				break
			}
		}
		if !hasEnv {
			for _, name := range binding.names {
				if value, ok := envFile[name]; ok {
					v.Set(binding.key, value)
					break
				}
			}
		}
	}
	cfg, err := decode(v)
	if err != nil {
		return Config{}, fmt.Errorf("decode configuration: %w", err)
	}
	_, port, err := net.SplitHostPort(cfg.ServerPort)
	if err != nil {
		return Config{}, fmt.Errorf("server.addr must be host:port or :port")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return Config{}, fmt.Errorf("server.addr port must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.DBDSN) == "" {
		return Config{}, fmt.Errorf("database.dsn must not be empty")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, fmt.Errorf("auth.jwt_secret must not be empty")
	}
	cfg.Log.Level = strings.ToLower(strings.TrimSpace(cfg.Log.Level))
	cfg.Log.Format = strings.ToLower(strings.TrimSpace(cfg.Log.Format))
	switch cfg.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, fmt.Errorf("log.level must be debug, info, warn or error")
	}
	if cfg.Log.Format != "text" && cfg.Log.Format != "json" {
		return Config{}, fmt.Errorf("log.format must be text or json")
	}
	return cfg, nil
}

func decode(v *viper.Viper) (Config, error) {
	var raw fileConfig
	if err := v.Unmarshal(&raw); err != nil {
		return Config{}, err
	}
	return Config{
		Log:        raw.Log,
		ServerPort: raw.Server.Addr, JWTSecret: raw.Auth.JWTSecret, DBDSN: raw.Database.DSN,
	}, nil
}

// Parse the existing simple KEY=value format without changing process environment.
func readEnvFile(path string) (map[string]string, error) {
	values := make(map[string]string)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return values, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
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
		if key != "" {
			values[key] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	return values, nil
}
