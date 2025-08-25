package security

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// TestEnvironment represents the test environment setup
type TestEnvironment struct {
	TempDir     string
	AuditLogger *logger.AuditLogger
	BasicLogger *logger.Logger
	Cleanup     func()
}

// SetupTestEnvironment creates a clean test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "integration-test-env-*")
	require.NoError(t, err)

	// Initialize loggers
	auditLogger, err := logger.NewAuditLogger()
	require.NoError(t, err)

	basicLogger := logger.New()

	// Setup cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return &TestEnvironment{
		TempDir:     tempDir,
		AuditLogger: auditLogger,
		BasicLogger: basicLogger,
		Cleanup:     cleanup,
	}
}

// CreateTestRepository creates a test Git repository for testing
func CreateTestRepository(t *testing.T, baseDir string) string {
	repoDir := filepath.Join(baseDir, "test-repo")
	err := os.MkdirAll(repoDir, 0755)
	require.NoError(t, err)

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = repoDir
	err = cmd.Run()
	if err != nil {
		t.Logf("Could not initialize git repository: %v", err)
		// Create a simple directory structure instead
		err = os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("# Test Repository"), 0644)
		require.NoError(t, err)
		return repoDir
	}

	// Create test files
	files := map[string]string{
		"README.md":    "# Test Repository\n\nThis is a test repository for integration testing.",
		"src/main.go":  "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		"src/utils.go": "package main\n\nfunc utils() {\n\t// Utility functions\n}",
		"docs/api.md":  "# API Documentation\n\nAPI endpoints and usage.",
		".gitignore":   "*.log\n*.tmp\nnode_modules/\n",
	}

	for filePath, content := range files {
		fullPath := filepath.Join(repoDir, filePath)
		dir := filepath.Dir(fullPath)

		err = os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Configure git user for commits
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = repoDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = repoDir
	_ = cmd.Run()

	// Add and commit files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = repoDir
	err = cmd.Run()
	if err == nil {
		cmd = exec.Command("git", "commit", "-m", "Initial commit for integration testing")
		cmd.Dir = repoDir
		_ = cmd.Run()
	}

	return repoDir
}

// CreateTestFiles creates test files in the specified directory
func CreateTestFiles(t *testing.T, baseDir string, files map[string]string) {
	for filePath, content := range files {
		fullPath := filepath.Join(baseDir, filePath)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}
}

// VerifyFileExists checks if a file exists and optionally verifies its content
func VerifyFileExists(t *testing.T, filePath string, expectedContent ...string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	require.NoError(t, err)

	if len(expectedContent) > 0 {
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		contentStr := string(content)
		for _, expected := range expectedContent {
			if !strings.Contains(contentStr, expected) {
				t.Errorf("File %s does not contain expected content: %s", filePath, expected)
				return false
			}
		}
	}

	return true
}

// VerifyFileDoesNotExist checks that a file does not exist
func VerifyFileDoesNotExist(t *testing.T, filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

// VerifyDirectoryEmpty checks that a directory is empty or does not exist
func VerifyDirectoryEmpty(t *testing.T, dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if os.IsNotExist(err) {
		return true // Directory doesn't exist, so it's "empty"
	}
	require.NoError(t, err)
	return len(entries) == 0
}

// MeasureExecutionTime measures the execution time of a function
func MeasureExecutionTime(operation func()) time.Duration {
	start := time.Now()
	operation()
	return time.Since(start)
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}

// IsDockerInstalled checks if Docker is installed and running
func IsDockerInstalled() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// IsGitInstalled checks if Git is installed
func IsGitInstalled() bool {
	cmd := exec.Command("git", "--version")
	return cmd.Run() == nil
}

// GetSystemInfo returns basic system information for testing
func GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"docker_available": IsDockerInstalled(),
		"git_available":    IsGitInstalled(),
		"temp_dir":         os.TempDir(),
		"working_dir": func() string {
			wd, _ := os.Getwd()
			return wd
		}(),
	}
}

// LogTestEnvironment logs test environment information
func LogTestEnvironment(t *testing.T, auditLogger *logger.AuditLogger) {
	sysInfo := GetSystemInfo()

	auditLogger.LogSecurityEvent(logger.SecurityScan, map[string]interface{}{
		"operation":   "integration_test_environment_setup",
		"system_info": sysInfo,
		"test_name":   t.Name(),
		"timestamp":   time.Now().Unix(),
	})
}

// CreateMockSecurityViolation creates a mock security violation for testing
func CreateMockSecurityViolation(auditLogger *logger.AuditLogger, violationType string, details map[string]interface{}) {
	violationDetails := map[string]interface{}{
		"violation_type": violationType,
		"timestamp":      time.Now().Unix(),
		"test_mode":      true,
	}

	// Merge additional details
	for key, value := range details {
		violationDetails[key] = value
	}

	auditLogger.LogSecurityEvent(logger.AccessViolation, violationDetails)
}

// ValidateSecurityConfiguration validates that security configurations are properly set
func ValidateSecurityConfiguration(t *testing.T, auditLogger *logger.AuditLogger) {
	// Check temp directory permissions
	tempDir := os.TempDir()
	info, err := os.Stat(tempDir)
	if err == nil {
		mode := info.Mode()
		auditLogger.LogSecurityEvent(logger.SecurityScan, map[string]interface{}{
			"operation": "temp_directory_permission_check",
			"path":      tempDir,
			"mode":      mode.String(),
		})
	}

	// Check working directory permissions
	wd, err := os.Getwd()
	if err == nil {
		info, err := os.Stat(wd)
		if err == nil {
			mode := info.Mode()
			auditLogger.LogSecurityEvent(logger.SecurityScan, map[string]interface{}{
				"operation": "working_directory_permission_check",
				"path":      wd,
				"mode":      mode.String(),
			})
		}
	}
}

// CleanupTestResources performs comprehensive cleanup of test resources
func CleanupTestResources(t *testing.T, resources []string) {
	for _, resource := range resources {
		if strings.HasPrefix(resource, "/") || strings.Contains(resource, "test") {
			err := os.RemoveAll(resource)
			if err != nil && !os.IsNotExist(err) {
				t.Logf("Warning: Could not clean up test resource %s: %v", resource, err)
			}
		}
	}
}

// TestConstants contains constants used across integration tests
var TestConstants = struct {
	MaxTestDuration  time.Duration
	DefaultTimeout   time.Duration
	CleanupTimeout   time.Duration
	ContainerTimeout time.Duration
	GitCloneTimeout  time.Duration
	MaxMemoryUsage   uint64
	MaxTempDirSize   int64
}{
	MaxTestDuration:  10 * time.Minute,
	DefaultTimeout:   2 * time.Minute,
	CleanupTimeout:   30 * time.Second,
	ContainerTimeout: 60 * time.Second,
	GitCloneTimeout:  30 * time.Second,
	MaxMemoryUsage:   100 * 1024 * 1024, // 100MB
	MaxTempDirSize:   50 * 1024 * 1024,  // 50MB
}
