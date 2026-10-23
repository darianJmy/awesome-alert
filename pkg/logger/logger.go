// Package logger configures application logging and adapters for HTTP and GORM.
package logger

import (
	"fmt"
	"io"
	stdlog "log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	File   string `mapstructure:"file"`
}

// New installs the shared logger before other Wire providers initialize.
// An optional file receives the same records as stderr. Cleanup restores the
// previous logger before closing the file, so startup failures can still log.
func New(cfg Config) (*slog.Logger, func(), error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(cfg.Level))); err != nil {
		return nil, nil, fmt.Errorf("invalid log.level: %w", err)
	}
	if cfg.Format != "text" && cfg.Format != "json" {
		return nil, nil, fmt.Errorf("log.format must be text or json")
	}
	output := io.Writer(os.Stderr)
	var file *os.File
	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0700); err != nil {
			return nil, nil, fmt.Errorf("create log directory: %w", err)
		}
		var err error
		file, err = os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		output = io.MultiWriter(os.Stderr, file)
	}
	log := slog.New(newHandler(output, cfg.Format, level))
	previous := slog.Default()
	previousWriter, previousFlags := stdlog.Writer(), stdlog.Flags()
	slog.SetDefault(log) // Also routes the standard library log package here.
	cleanup := func() {
		slog.SetDefault(previous)
		stdlog.SetOutput(previousWriter)
		stdlog.SetFlags(previousFlags)
		if file != nil {
			_ = file.Close()
		}
	}
	return log, cleanup, nil
}

func newHandler(output io.Writer, format string, level slog.Level) slog.Handler {
	options := &slog.HandlerOptions{Level: level}
	if format == "json" {
		return slog.NewJSONHandler(output, options)
	}
	return slog.NewTextHandler(output, options)
}

func Debug(message string, args ...any) { slog.Debug(message, args...) }
func Info(message string, args ...any)  { slog.Info(message, args...) }
func Warn(message string, args ...any)  { slog.Warn(message, args...) }
func Error(message string, args ...any) { slog.Error(message, args...) }
