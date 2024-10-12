package utils

import (
	"io"
	"log/slog"
	"os"
)

var Logger *slog.Logger

func init() {
	SetLogLevel(slog.LevelWarn)
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level slog.Level) {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	multiWriter := io.MultiWriter(logFile)

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(multiWriter, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}
