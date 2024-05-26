package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger wraps the standard log.Logger to add convenience methods.
type Logger struct {
	*log.Logger
}

// Config defines the structure for log configuration settings.
type Config struct {
	LogDir  string // Directory to save log files
	LogFile string // Log file name
}

// InitLogger initializes and returns a logger with the given configuration.
func InitLogger(cfg Config) (*Logger, error) {
	// Ensure the log directory exists.
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, err
	}

	// Create the full path to the log file.
	logFilePath := filepath.Join(cfg.LogDir, cfg.LogFile)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// Create the logger without any flags to prevent automatic timestamp and file info.
	logger := log.New(logFile, "", 0)
	return &Logger{Logger: logger}, nil
}

// DefaultConfig provides a default logging configuration.
func DefaultConfig() Config {
	return Config{
		LogDir:  "logs",
		LogFile: "voxctl.log",
	}
}

// formatLogMessage formats the log message with a newline before the level-specific message.
func formatLogMessage(level, msg string) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	_, file, line, ok := runtime.Caller(2) // Change from 3 to 2
	if ok {
		return fmt.Sprintf("%s %s:%d:\n%s - %s", timestamp, file, line, level, msg)
	}
	return fmt.Sprintf("%s:\n%s - %s", timestamp, level, msg)
}

// Info logs an informational message.
func (l *Logger) Info(msg string) {
	l.Output(2, formatLogMessage(" - INFO", msg))
}

// Infof logs an informational message with formatting.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(2, formatLogMessage(" - INFO", fmt.Sprintf(format, v...)))
}

// Warning logs a warning message.
func (l *Logger) Warning(msg string) {
	l.Output(2, formatLogMessage(" - WARNING", msg))
}

// Warningf logs a warning message with formatting.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Output(2, formatLogMessage(" - WARNING", fmt.Sprintf(format, v...)))
}

// Error logs an error message.
func (l *Logger) Error(msg string) {
	l.Output(2, formatLogMessage(" - ERROR", msg))
}

// Errorf logs an error message with formatting.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(2, formatLogMessage(" - ERROR", fmt.Sprintf(format, v...)))
}
