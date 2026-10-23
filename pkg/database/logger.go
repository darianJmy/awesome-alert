package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

const slowQueryThreshold = 200 * time.Millisecond

type gormLogger struct {
	log   *slog.Logger
	level gormlog.LogLevel
}

func Logger(log *slog.Logger) gormlog.Interface {
	return gormLogger{log: log, level: gormlog.Info}
}

func (l gormLogger) LogMode(level gormlog.LogLevel) gormlog.Interface { l.level = level; return l }
func (l gormLogger) Info(ctx context.Context, message string, args ...interface{}) {
	if l.level >= gormlog.Info {
		l.log.InfoContext(ctx, fmt.Sprintf(message, args...))
	}
}
func (l gormLogger) Warn(ctx context.Context, message string, args ...interface{}) {
	if l.level >= gormlog.Warn {
		l.log.WarnContext(ctx, fmt.Sprintf(message, args...))
	}
}
func (l gormLogger) Error(ctx context.Context, message string, args ...interface{}) {
	if l.level >= gormlog.Error {
		l.log.ErrorContext(ctx, fmt.Sprintf(message, args...))
	}
}

// GORM calls ParamsFilter before interpolating bound values into the trace SQL.
func (l gormLogger) ParamsFilter(_ context.Context, sql string, _ ...interface{}) (string, []interface{}) {
	return sql, nil
}

func (l gormLogger) Trace(ctx context.Context, begin time.Time, query func() (string, int64), err error) {
	if l.level == gormlog.Silent {
		return
	}
	elapsed := time.Since(begin)
	level, message := slog.LevelDebug, "database query"
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.level >= gormlog.Error {
		level, message = slog.LevelError, "database query failed"
	} else if elapsed >= slowQueryThreshold && l.level >= gormlog.Warn {
		level, message = slog.LevelWarn, "slow database query"
	} else if l.level < gormlog.Info {
		return
	}
	if !l.log.Enabled(ctx, level) {
		return
	}
	sql, rows := query()
	attrs := []any{"duration_ms", float64(elapsed.Microseconds()) / 1000, "rows", rows, "sql", sql}
	if level == slog.LevelError {
		attrs = append(attrs, "error", err)
	}
	l.log.Log(ctx, level, message, attrs...)
}
