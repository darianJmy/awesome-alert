package database

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func OpenSQLite(dsn string, log *slog.Logger) (*gorm.DB, error) {
	if log == nil {
		log = slog.Default()
	}
	normalized, err := sqliteDSN(dsn)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(normalized), &gorm.Config{Logger: Logger(log)})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("configure database pool: %w", err)
	}
	// SQLite allows only one writer. Keep a single long-lived connection so
	// concurrent scheduler/request writes serialize and memory databases survive.
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	return db, nil
}

func sqliteDSN(dsn string) (string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" || strings.Contains(dsn, "@tcp(") || strings.Contains(dsn, "@unix(") || strings.Contains(dsn, "://") && !strings.HasPrefix(dsn, "file:") {
		return "", fmt.Errorf("database.dsn must be a SQLite file path or file: URI")
	}
	name, query, _ := strings.Cut(dsn, "?")
	params, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("parse SQLite DSN: %w", err)
	}
	path := name
	if strings.HasPrefix(name, "file:") {
		uri, err := url.Parse(name)
		if err != nil {
			return "", fmt.Errorf("parse SQLite file URI: %w", err)
		}
		if uri.Host != "" && uri.Host != "localhost" {
			return "", fmt.Errorf("SQLite file URI must be local")
		}
		path = uri.Path
		if uri.Opaque != "" {
			path, err = url.PathUnescape(uri.Opaque)
			if err != nil {
				return "", fmt.Errorf("parse SQLite file path: %w", err)
			}
		}
	}
	if path == "" {
		return "", fmt.Errorf("SQLite database path must not be empty")
	}
	if path != ":memory:" && params.Get("mode") != "memory" {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return "", fmt.Errorf("create SQLite database directory: %w", err)
		}
	}
	// Apply on every new connection, including one reopened by database/sql.
	params.Set("_txlock", "immediate")
	params.Add("_pragma", "busy_timeout(5000)")
	params.Add("_pragma", "journal_mode(WAL)")
	params.Add("_pragma", "foreign_keys(1)")
	return name + "?" + params.Encode(), nil
}
