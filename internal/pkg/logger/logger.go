package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for application use.
type Logger struct {
	*slog.Logger
}

// New creates a new structured logger.
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

// WithContext returns a new logger with context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{l.Logger.With(slog.Any("context", ctx))}
}

// WithError adds error to the logger.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.Logger.With(slog.Any("error", err))}
}

// WithFields adds multiple fields to the logger.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []any
	for k, v := range fields {
		args = append(args, slog.Any(k, v))
	}
	return &Logger{l.Logger.With(args...)}
}

// Global logger instance
var globalLogger *Logger

// Initialize sets up the global logger.
func Initialize(isDev bool) {
	globalLogger = New(isDev)
}

// Get returns the global logger.
func Get() *Logger {
	if globalLogger == nil {
		globalLogger = New(false) // default to production mode
	}
	return globalLogger
}
