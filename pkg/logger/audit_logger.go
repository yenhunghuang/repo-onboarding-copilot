// Package logger provides enhanced audit logging capabilities for security events
// with structured logging, sensitive data protection, and secure storage mechanisms.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AuditLogger provides security-focused logging with enhanced capabilities
type AuditLogger struct {
	*Logger
	logRotation   LogRotationConfig
	sensitiveKeys []string
	auditFile     *os.File
	enableFileLog bool
}

// LogRotationConfig represents log rotation configuration
type LogRotationConfig struct {
	MaxFileSize  int64  // Maximum file size in bytes before rotation
	MaxFiles     int    // Maximum number of log files to keep
	RotateDaily  bool   // Whether to rotate logs daily
	LogDirectory string // Directory to store audit logs
}

// SecurityEvent represents different types of security events
type SecurityEvent string

const (
	// Git operations
	GitCloneStart   SecurityEvent = "git_clone_start"
	GitCloneSuccess SecurityEvent = "git_clone_success"
	GitCloneFailure SecurityEvent = "git_clone_failure"

	// Container operations
	ContainerCreate SecurityEvent = "container_create"
	ContainerStart  SecurityEvent = "container_start"
	ContainerStop   SecurityEvent = "container_stop"
	ContainerExec   SecurityEvent = "container_exec"
	ContainerFail   SecurityEvent = "container_fail"

	// Security events
	SecurityScan    SecurityEvent = "security_scan"
	AccessViolation SecurityEvent = "access_violation"
	AuthFailure     SecurityEvent = "auth_failure"

	// System events
	SystemCleanup  SecurityEvent = "system_cleanup"
	ResourceLimit  SecurityEvent = "resource_limit"
	ValidationFail SecurityEvent = "validation_fail"
)

// NewAuditLogger creates a new audit logger with enhanced security features
func NewAuditLogger() (*AuditLogger, error) {
	baseLogger := New()

	// Default log rotation configuration
	rotationConfig := LogRotationConfig{
		MaxFileSize:  100 * 1024 * 1024, // 100MB
		MaxFiles:     10,
		RotateDaily:  true,
		LogDirectory: "logs/audit",
	}

	// Sensitive keys that should be redacted in logs
	sensitiveKeys := []string{
		"password", "token", "key", "secret", "credential",
		"auth", "bearer", "api_key", "access_token", "refresh_token",
	}

	auditLogger := &AuditLogger{
		Logger:        baseLogger,
		logRotation:   rotationConfig,
		sensitiveKeys: sensitiveKeys,
		enableFileLog: false,
	}

	return auditLogger, nil
}

// NewAuditLoggerWithFile creates an audit logger with file output enabled
func NewAuditLoggerWithFile(logDir string) (*AuditLogger, error) {
	auditLogger, err := NewAuditLogger()
	if err != nil {
		return nil, err
	}

	// Update log directory
	auditLogger.logRotation.LogDirectory = logDir

	// Enable file logging
	if err := auditLogger.enableFileLogging(); err != nil {
		return nil, fmt.Errorf("failed to enable file logging: %w", err)
	}

	return auditLogger, nil
}

// enableFileLogging sets up file-based audit logging
func (al *AuditLogger) enableFileLogging() error {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(al.logRotation.LogDirectory, 0750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create audit log file
	logPath := filepath.Join(al.logRotation.LogDirectory,
		fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02")))

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	al.auditFile = file
	al.enableFileLog = true

	// Set logger to write to both stdout and file
	al.Logger.SetOutput(file)

	return nil
}

// LogSecurityEvent logs a security event with structured fields
func (al *AuditLogger) LogSecurityEvent(event SecurityEvent, fields map[string]interface{}) {
	// Sanitize sensitive information
	sanitizedFields := al.sanitizeFields(fields)

	// Add standard audit fields
	auditFields := map[string]interface{}{
		"audit_event": string(event),
		"timestamp":   time.Now().Unix(),
		"source":      "repo-onboarding-copilot",
		"version":     "1.0.0",
	}

	// Merge with provided fields
	for k, v := range sanitizedFields {
		auditFields[k] = v
	}

	// Log based on event severity
	switch event {
	case GitCloneFailure, ContainerFail, AccessViolation, AuthFailure, ValidationFail:
		al.WithFields(auditFields).Error("Security event occurred")
	case ResourceLimit, SecurityScan:
		al.WithFields(auditFields).Warn("Security event occurred")
	default:
		al.WithFields(auditFields).Info("Security event occurred")
	}

	// Check for log rotation if file logging is enabled
	if al.enableFileLog {
		al.checkLogRotation()
	}
}

// LogAccessPattern logs repository access patterns for security monitoring
func (al *AuditLogger) LogAccessPattern(repoURL string, accessType string, userAgent string, success bool) {
	fields := map[string]interface{}{
		"repo_url":    al.sanitizeURL(repoURL),
		"access_type": accessType,
		"user_agent":  userAgent,
		"success":     success,
		"client_ip":   "localhost", // Could be enhanced to capture real IP
	}

	if success {
		al.LogSecurityEvent(GitCloneSuccess, fields)
	} else {
		al.LogSecurityEvent(GitCloneFailure, fields)
	}
}

// LogContainerActivity logs container lifecycle events
func (al *AuditLogger) LogContainerActivity(containerID string, action string, result string, metadata map[string]interface{}) {
	fields := map[string]interface{}{
		"container_id": containerID,
		"action":       action,
		"result":       result,
	}

	// Add metadata if provided
	for k, v := range metadata {
		fields[k] = v
	}

	var event SecurityEvent
	switch action {
	case "create":
		event = ContainerCreate
	case "start":
		event = ContainerStart
	case "stop":
		event = ContainerStop
	case "exec":
		event = ContainerExec
	default:
		event = ContainerFail
	}

	al.LogSecurityEvent(event, fields)
}

// LogCleanupActivity logs cleanup operations for audit trail
func (al *AuditLogger) LogCleanupActivity(resourceType string, resourceID string, success bool, details map[string]interface{}) {
	fields := map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"success":       success,
	}

	// Add additional details
	for k, v := range details {
		fields[k] = v
	}

	al.LogSecurityEvent(SystemCleanup, fields)
}

// sanitizeFields removes or redacts sensitive information from log fields
func (al *AuditLogger) sanitizeFields(fields map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range fields {
		keyLower := strings.ToLower(key)

		// Check if key contains sensitive information
		isSensitive := false
		for _, sensitiveKey := range al.sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "[REDACTED]"
		} else {
			// For string values, check content for sensitive patterns
			if str, ok := value.(string); ok {
				sanitized[key] = al.sanitizeStringValue(str)
			} else {
				sanitized[key] = value
			}
		}
	}

	return sanitized
}

// sanitizeStringValue sanitizes string values to remove sensitive information
func (al *AuditLogger) sanitizeStringValue(value string) string {
	valueLower := strings.ToLower(value)

	// Check for common sensitive patterns
	for _, sensitiveKey := range al.sensitiveKeys {
		if strings.Contains(valueLower, sensitiveKey) {
			return "[REDACTED SENSITIVE CONTENT]"
		}
	}

	return value
}

// sanitizeURL removes credentials from URLs for secure logging
func (al *AuditLogger) sanitizeURL(url string) string {
	// Remove credentials from URLs
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			protocolPart := strings.Split(parts[0], "://")
			if len(protocolPart) == 2 {
				return protocolPart[0] + "://[REDACTED]@" + parts[1]
			}
		}
	}
	return url
}

// checkLogRotation checks if log rotation is needed and performs it
func (al *AuditLogger) checkLogRotation() {
	if !al.enableFileLog || al.auditFile == nil {
		return
	}

	// Get current file info
	fileInfo, err := al.auditFile.Stat()
	if err != nil {
		al.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to get audit log file info")
		return
	}

	// Check if rotation is needed based on file size
	if fileInfo.Size() >= al.logRotation.MaxFileSize {
		al.rotateLogFile()
	}

	// Check if daily rotation is needed
	if al.logRotation.RotateDaily {
		today := time.Now().Format("2006-01-02")
		if !strings.Contains(fileInfo.Name(), today) {
			al.rotateLogFile()
		}
	}
}

// rotateLogFile performs log file rotation
func (al *AuditLogger) rotateLogFile() {
	if al.auditFile != nil {
		al.auditFile.Close()
	}

	// Create new log file with current timestamp
	logPath := filepath.Join(al.logRotation.LogDirectory,
		fmt.Sprintf("audit-%s.log", time.Now().Format("2006-01-02-150405")))

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		al.Error("Failed to create new audit log file: " + err.Error())
		return
	}

	al.auditFile = file
	al.Logger.SetOutput(file)

	// Clean up old log files
	al.cleanupOldLogFiles()
}

// cleanupOldLogFiles removes old log files based on retention policy
func (al *AuditLogger) cleanupOldLogFiles() {
	files, err := filepath.Glob(filepath.Join(al.logRotation.LogDirectory, "audit-*.log"))
	if err != nil {
		al.Error("Failed to list audit log files: " + err.Error())
		return
	}

	// If we have more files than MaxFiles, remove the oldest
	if len(files) > al.logRotation.MaxFiles {
		// Sort files by modification time (oldest first)
		for i := 0; i < len(files)-al.logRotation.MaxFiles; i++ {
			if err := os.Remove(files[i]); err != nil {
				al.WithFields(map[string]interface{}{
					"file":  files[i],
					"error": err.Error(),
				}).Error("Failed to remove old audit log file")
			}
		}
	}
}

// Close properly closes the audit logger and its resources
func (al *AuditLogger) Close() error {
	if al.auditFile != nil {
		return al.auditFile.Close()
	}
	return nil
}

// GetLogDirectory returns the current log directory
func (al *AuditLogger) GetLogDirectory() string {
	return al.logRotation.LogDirectory
}

// SetLogRotation updates log rotation configuration
func (al *AuditLogger) SetLogRotation(config LogRotationConfig) {
	al.logRotation = config
}
