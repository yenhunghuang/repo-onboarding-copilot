package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := New()
	
	require.NotNil(t, logger)
	assert.IsType(t, &Logger{}, logger)
	assert.Equal(t, logrus.InfoLevel, logger.Logger.Level)
}

func TestNewWithLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         LogLevel
		expectedLevel logrus.Level
	}{
		{
			name:          "debug level",
			level:         DebugLevel,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "info level",
			level:         InfoLevel,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "warn level",
			level:         WarnLevel,
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "error level",
			level:         ErrorLevel,
			expectedLevel: logrus.ErrorLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewWithLevel(tt.level)
			
			require.NotNil(t, logger)
			assert.Equal(t, tt.expectedLevel, logger.Logger.Level)
		})
	}
}

func TestLogger_SetLogLevel(t *testing.T) {
	logger := New()
	
	tests := []struct {
		name          string
		level         LogLevel
		expectedLevel logrus.Level
	}{
		{
			name:          "set debug level",
			level:         DebugLevel,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "set info level",
			level:         InfoLevel,
			expectedLevel: logrus.InfoLevel,
		},
		{
			name:          "set warn level",
			level:         WarnLevel,
			expectedLevel: logrus.WarnLevel,
		},
		{
			name:          "set error level",
			level:         ErrorLevel,
			expectedLevel: logrus.ErrorLevel,
		},
		{
			name:          "set invalid level defaults to info",
			level:         LogLevel("invalid"),
			expectedLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLogLevel(tt.level)
			assert.Equal(t, tt.expectedLevel, logger.Logger.Level)
		})
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger := New()
	
	// Capture output
	var buf bytes.Buffer
	logger.Logger.SetOutput(&buf)
	
	fields := map[string]interface{}{
		"component": "test",
		"action":    "validation",
	}
	
	entry := logger.WithFields(fields)
	entry.Info("test message")
	
	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "test", logEntry["component"])
	assert.Equal(t, "validation", logEntry["action"])
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "info", logEntry["level"])
}

func TestLogger_LoggingOutput(t *testing.T) {
	logger := New()
	
	// Capture output
	var buf bytes.Buffer
	logger.Logger.SetOutput(&buf)
	
	testMessage := "test log message"
	logger.Info(testMessage)
	
	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, testMessage, logEntry["msg"])
	assert.Equal(t, "info", logEntry["level"])
	assert.Contains(t, logEntry, "time")
}