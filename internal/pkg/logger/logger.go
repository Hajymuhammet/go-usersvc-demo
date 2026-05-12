package logger

import (
	"context"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(isDev bool) *Logger {
	var opts *slog.HandlerOptions
	if isDev {
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	} else {
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{slog.New(handler)}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{l.Logger.With(slog.Any("context", ctx))}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.Logger.With(slog.Any("error", err))}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []any
	for k, v := range fields {
		args = append(args, slog.Any(k, v))
	}
	return &Logger{l.Logger.With(args...)}
}

var globalLogger *Logger

func Initialize(isDev bool) {
	globalLogger = New(isDev)
}

func Get() *Logger {
	if globalLogger == nil {
		globalLogger = New(false)
	}
	return globalLogger
}
