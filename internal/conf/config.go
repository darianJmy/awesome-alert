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

// SMTPConfig is shared by all email notification recipients.
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Security string `mapstructure:"security"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type Config struct {
	SMTP       SMTPConfig
	Log        LogConfig
	ServerPort string
	JWTSecret  string
	DBDSN      string
}

type fileConfig struct {
	SMTP   SMTPConfig `mapstructure:"smtp"`
	Log    LogConfig  `mapstructure:"log"`
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
	key  string
	name string
}{
	{"smtp.host", "APP_SMTP_HOST"},
	{"smtp.port", "APP_SMTP_PORT"},
	{"smtp.security", "APP_SMTP_SECURITY"},
	{"smtp.username", "APP_SMTP_USERNAME"},
	{"smtp.password", "APP_SMTP_PASSWORD"},
	{"smtp.from", "APP_SMTP_FROM"},
	{"log.level", "APP_LOG_LEVEL"},
	{"log.format", "APP_LOG_FORMAT"},
	{"log.file", "APP_LOG_FILE"},
	{"server.addr", "APP_SERVER_ADDR"},
	{"auth.jwt_secret", "APP_AUTH_JWT_SECRET"},
	{"database.dsn", "APP_DATABASE_DSN"},
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
		if err := v.BindEnv(binding.key, binding.name); err != nil {
			return Config{}, fmt.Errorf("bind environment %s: %w", binding.key, err)
		}
		if _, present := os.LookupEnv(binding.name); !present {
			if value, ok := envFile[binding.name]; ok {
				v.Set(binding.key, value)
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
		SMTP:       raw.SMTP,
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
