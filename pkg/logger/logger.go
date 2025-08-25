// Package logger provides structured JSON logging capabilities
// with configurable log levels and proper error handling.
package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Logger represents the application logger with structured logging capabilities
type Logger struct {
	*logrus.Logger
}

// LogLevel represents available log levels
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// New creates a new structured logger instance
func New() *Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)

	return &Logger{Logger: log}
}

// NewWithLevel creates a new logger with specified level
func NewWithLevel(level LogLevel) *Logger {
	logger := New()
	logger.SetLogLevel(level)
	return logger
}

// SetLogLevel sets the logging level
func (l *Logger) SetLogLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		l.Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.Logger.SetLevel(logrus.ErrorLevel)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

// WithFields adds fields to log entry
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// ErrorWithExit logs error and exits with specified code
func (l *Logger) ErrorWithExit(msg string, code int) {
	l.Error(msg)
	os.Exit(code)
}

// ErrorfWithExit logs formatted error and exits with specified code
func (l *Logger) ErrorfWithExit(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// FatalError logs error and exits with code 1
func (l *Logger) FatalError(msg string) {
	l.ErrorWithExit(msg, 1)
}
