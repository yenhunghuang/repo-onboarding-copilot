package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditLogger(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)
	assert.NotNil(t, auditLogger)
	assert.NotNil(t, auditLogger.Logger)
	assert.False(t, auditLogger.enableFileLog)
	assert.Equal(t, "logs/audit", auditLogger.logRotation.LogDirectory)
	assert.Equal(t, int64(100*1024*1024), auditLogger.logRotation.MaxFileSize)
	assert.Equal(t, 10, auditLogger.logRotation.MaxFiles)
	assert.True(t, auditLogger.logRotation.RotateDaily)
}

func TestNewAuditLoggerWithFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	auditLogger, err := NewAuditLoggerWithFile(tempDir)
	require.NoError(t, err)
	defer auditLogger.Close()

	assert.NotNil(t, auditLogger)
	assert.True(t, auditLogger.enableFileLog)
	assert.Equal(t, tempDir, auditLogger.logRotation.LogDirectory)
	assert.NotNil(t, auditLogger.auditFile)

	// Verify log directory was created
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)
}

func TestLogSecurityEvent(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name   string
		event  SecurityEvent
		fields map[string]interface{}
	}{
		{
			name:  "git clone success",
			event: GitCloneSuccess,
			fields: map[string]interface{}{
				"repo_url": "https://github.com/test/repo.git",
				"duration": 5.2,
			},
		},
		{
			name:  "container failure",
			event: ContainerFail,
			fields: map[string]interface{}{
				"container_id": "abc123",
				"error":        "resource limit exceeded",
			},
		},
		{
			name:  "security scan",
			event: SecurityScan,
			fields: map[string]interface{}{
				"scan_type": "vulnerability",
				"findings":  3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies the method doesn't panic and handles different event types
			assert.NotPanics(t, func() {
				auditLogger.LogSecurityEvent(tt.event, tt.fields)
			})
		})
	}
}

func TestLogAccessPattern(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name      string
		repoURL   string
		access    string
		userAgent string
		success   bool
	}{
		{
			name:      "successful clone",
			repoURL:   "https://github.com/test/repo.git",
			access:    "clone",
			userAgent: "git/2.30.0",
			success:   true,
		},
		{
			name:      "failed clone with credentials",
			repoURL:   "https://user:pass@github.com/test/repo.git",
			access:    "clone",
			userAgent: "git/2.30.0",
			success:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				auditLogger.LogAccessPattern(tt.repoURL, tt.access, tt.userAgent, tt.success)
			})
		})
	}
}

func TestLogContainerActivity(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name        string
		containerID string
		action      string
		result      string
		metadata    map[string]interface{}
	}{
		{
			name:        "container create",
			containerID: "container123",
			action:      "create",
			result:      "success",
			metadata: map[string]interface{}{
				"image":  "alpine:latest",
				"memory": "2GB",
			},
		},
		{
			name:        "container exec",
			containerID: "container456",
			action:      "exec",
			result:      "success",
			metadata: map[string]interface{}{
				"command": "ls -la",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				auditLogger.LogContainerActivity(tt.containerID, tt.action, tt.result, tt.metadata)
			})
		})
	}
}

func TestLogCleanupActivity(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name         string
		resourceType string
		resourceID   string
		success      bool
		details      map[string]interface{}
	}{
		{
			name:         "successful cleanup",
			resourceType: "container",
			resourceID:   "container123",
			success:      true,
			details: map[string]interface{}{
				"cleanup_type": "automatic",
				"duration":     2.5,
			},
		},
		{
			name:         "failed cleanup",
			resourceType: "temp_directory",
			resourceID:   "/tmp/repo-sandbox-123",
			success:      false,
			details: map[string]interface{}{
				"error": "permission denied",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				auditLogger.LogCleanupActivity(tt.resourceType, tt.resourceID, tt.success, tt.details)
			})
		})
	}
}

func TestSanitizeFields(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "normal fields",
			input: map[string]interface{}{
				"repo_url": "https://github.com/test/repo.git",
				"duration": 5.2,
				"success":  true,
			},
			expected: map[string]interface{}{
				"repo_url": "https://github.com/test/repo.git",
				"duration": 5.2,
				"success":  true,
			},
		},
		{
			name: "sensitive key names",
			input: map[string]interface{}{
				"password":     "secret123",
				"api_token":    "abc123",
				"secret_key":   "xyz789",
				"normal_field": "normal_value",
			},
			expected: map[string]interface{}{
				"password":     "[REDACTED]",
				"api_token":    "[REDACTED]",
				"secret_key":   "[REDACTED]",
				"normal_field": "normal_value",
			},
		},
		{
			name: "sensitive string values",
			input: map[string]interface{}{
				"command":    "mysql -u user -ppassword123",
				"url":        "https://api.github.com/token/abc123",
				"normal_cmd": "ls -la",
			},
			expected: map[string]interface{}{
				"command":    "[REDACTED SENSITIVE CONTENT]",
				"url":        "[REDACTED SENSITIVE CONTENT]",
				"normal_cmd": "ls -la",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditLogger.sanitizeFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeStringValue(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal string",
			input:    "https://github.com/test/repo.git",
			expected: "https://github.com/test/repo.git",
		},
		{
			name:     "string with password",
			input:    "mysql -u user -ppassword123",
			expected: "[REDACTED SENSITIVE CONTENT]",
		},
		{
			name:     "string with token",
			input:    "curl -H 'Authorization: token abc123'",
			expected: "[REDACTED SENSITIVE CONTENT]",
		},
		{
			name:     "string with secret",
			input:    "export SECRET=mysecret123",
			expected: "[REDACTED SENSITIVE CONTENT]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditLogger.sanitizeStringValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "https with credentials",
			input:    "https://user:pass@github.com/owner/repo.git",
			expected: "https://[REDACTED]@github.com/owner/repo.git",
		},
		{
			name:     "http with credentials",
			input:    "http://user:token@gitlab.com/owner/repo.git",
			expected: "http://[REDACTED]@gitlab.com/owner/repo.git",
		},
		{
			name:     "no credentials",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo.git",
		},
		{
			name:     "ssh url",
			input:    "git@github.com:owner/repo.git",
			expected: "git@github.com:owner/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auditLogger.sanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityEventTypes(t *testing.T) {
	// Test that all security event constants are properly defined
	events := []SecurityEvent{
		GitCloneStart, GitCloneSuccess, GitCloneFailure,
		ContainerCreate, ContainerStart, ContainerStop, ContainerExec, ContainerFail,
		SecurityScan, AccessViolation, AuthFailure,
		SystemCleanup, ResourceLimit, ValidationFail,
	}

	for _, event := range events {
		assert.NotEmpty(t, string(event))
	}
}

func TestLogRotationConfig(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	// Test default configuration
	config := auditLogger.logRotation
	assert.Equal(t, int64(100*1024*1024), config.MaxFileSize)
	assert.Equal(t, 10, config.MaxFiles)
	assert.True(t, config.RotateDaily)
	assert.Equal(t, "logs/audit", config.LogDirectory)

	// Test updating configuration
	newConfig := LogRotationConfig{
		MaxFileSize:  50 * 1024 * 1024, // 50MB
		MaxFiles:     5,
		RotateDaily:  false,
		LogDirectory: "/custom/log/path",
	}

	auditLogger.SetLogRotation(newConfig)
	updatedConfig := auditLogger.logRotation

	assert.Equal(t, int64(50*1024*1024), updatedConfig.MaxFileSize)
	assert.Equal(t, 5, updatedConfig.MaxFiles)
	assert.False(t, updatedConfig.RotateDaily)
	assert.Equal(t, "/custom/log/path", updatedConfig.LogDirectory)
}

func TestGetLogDirectory(t *testing.T) {
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	assert.Equal(t, "logs/audit", auditLogger.GetLogDirectory())

	// Test with custom directory
	customDir := "/tmp/custom-audit"
	auditLogger.logRotation.LogDirectory = customDir
	assert.Equal(t, customDir, auditLogger.GetLogDirectory())
}

func TestAuditLoggerClose(t *testing.T) {
	// Test closing logger without file
	auditLogger, err := NewAuditLogger()
	require.NoError(t, err)

	err = auditLogger.Close()
	assert.NoError(t, err)

	// Test closing logger with file
	tempDir, err := os.MkdirTemp("", "audit-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	auditLoggerWithFile, err := NewAuditLoggerWithFile(tempDir)
	require.NoError(t, err)

	err = auditLoggerWithFile.Close()
	assert.NoError(t, err)
}

func TestFileLoggingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping file logging integration test in short mode")
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "audit-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create audit logger with file output
	auditLogger, err := NewAuditLoggerWithFile(tempDir)
	require.NoError(t, err)
	defer auditLogger.Close()

	// Log some events
	auditLogger.LogSecurityEvent(GitCloneStart, map[string]interface{}{
		"repo_url": "https://github.com/test/repo.git",
	})

	auditLogger.LogContainerActivity("container123", "create", "success", map[string]interface{}{
		"image": "alpine:latest",
	})

	// Verify log file was created
	files, err := filepath.Glob(filepath.Join(tempDir, "audit-*.log"))
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify file is not empty
	fileInfo, err := os.Stat(files[0])
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))
}