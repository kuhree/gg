package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LogConfig holds configuration for the logger
type LogConfig struct {
	// LogFile is the path to the log file
	LogFile string
	// MaxSize is the maximum size in megabytes before log rotation
	MaxSize int64
	// LogToStdout determines if logs should also go to stdout
	LogToStdout bool
	// Level sets the minimum log level
	Level slog.Level
}

var (
	Logger    *slog.Logger
	logFile   *os.File
	defaultConfig = LogConfig{
		LogFile:     "gg.log",
		MaxSize:     10, // 10MB
		LogToStdout: false,
		Level:       slog.LevelWarn,
	}
)

func init() {
	if err := SetupLogger(defaultConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

// SetupLogger configures the logger with the given config
func SetupLogger(config LogConfig) error {
	if err := ensureLogDir(config.LogFile); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Close existing log file if open
	if logFile != nil {
		logFile.Close()
	}

	var err error
	logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	writers := []io.Writer{logFile}
	if config.LogToStdout {
		writers = append(writers, os.Stdout)
	}

	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	handler := slog.NewTextHandler(io.MultiWriter(writers...), opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	// Start log rotation goroutine
	go monitorLogSize(config)

	return nil
}

// SetLogLevel sets the log level for the logger
func SetLogLevel(level slog.Level) error {
	return SetupLogger(LogConfig{
		LogFile:     defaultConfig.LogFile,
		MaxSize:     defaultConfig.MaxSize,
		LogToStdout: defaultConfig.LogToStdout,
		Level:       level,
	})
}

// Cleanup closes the log file
func Cleanup() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

func ensureLogDir(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, 0755)
}

func monitorLogSize(config LogConfig) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if logFile == nil {
			continue
		}

		info, err := logFile.Stat()
		if err != nil {
			Logger.Error("Failed to stat log file", "error", err)
			continue
		}

		if info.Size() > config.MaxSize*1024*1024 {
			rotateLog(config)
		}
	}
}

func rotateLog(config LogConfig) {
	// Close current log file
	logFile.Close()

	// Rotate the log file
	timestamp := time.Now().Format("20060102-150405")
	rotatedName := fmt.Sprintf("%s.%s", config.LogFile, timestamp)
	
	if err := os.Rename(config.LogFile, rotatedName); err != nil {
		Logger.Error("Failed to rotate log file", "error", err)
		return
	}

	// Setup logger with new file
	if err := SetupLogger(config); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logger after rotation: %v\n", err)
	}
}
