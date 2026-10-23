package database

import (
	"bytes"
	"context"
	"errors"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"log"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func capture(t *testing.T, level slog.Level) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previous := slog.Default()
	previousWriter, previousFlags := log.Writer(), log.Flags()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: level})))
	t.Cleanup(func() { slog.SetDefault(previous); log.SetOutput(previousWriter); log.SetFlags(previousFlags) })
	return &output
}

func TestGORMUsesParameterizedLogs(t *testing.T) {
	output := capture(t, slog.LevelDebug)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: Logger(slog.Default())})
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	defer pool.Close()
	if err := db.Exec("CREATE TABLE credentials (value TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO credentials VALUES (?)", "secret-password").Error; err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "secret-password") || !strings.Contains(output.String(), "INSERT INTO credentials VALUES (?)") {
		t.Fatalf("SQL parameters were not hidden: %s", output.String())
	}
	output.Reset()
	adapter := Logger(slog.Default())
	adapter.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) { return "SELECT ?", 1 }, nil)
	if !strings.Contains(output.String(), `"level":"WARN"`) {
		t.Fatal("slow query not logged as warning")
	}
	output.Reset()
	adapter.Trace(context.Background(), time.Now(), func() (string, int64) { return "SELECT ?", 0 }, errors.New("test failure"))
	if !strings.Contains(output.String(), `"level":"ERROR"`) {
		t.Fatal("query error not logged as error")
	}
	output.Reset()
	adapter.LogMode(gormlog.Silent).Trace(context.Background(), time.Now(), func() (string, int64) { t.Fatal("silent logger evaluated query"); return "", 0 }, errors.New("test failure"))
	if output.Len() != 0 {
		t.Fatal("silent logger emitted records")
	}
}
