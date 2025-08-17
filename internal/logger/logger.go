package logger

import (
	"io"
	"log/slog"
)

var logLevel slog.Level

func NewLogger(logLevel slog.Level, w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler)
}

func LogLevel() slog.Level {
	return logLevel
}
