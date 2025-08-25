package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

func TestNewCleanupOrchestrator(t *testing.T) {
	tests := []struct {
		name        string
		auditLogger *logger.AuditLogger
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid logger",
			auditLogger: createTestAuditLogger(t),
			wantErr:     false,
		},
		{
			name:        "nil logger",
			auditLogger: nil,
			wantErr:     true,
			errMsg:      "audit logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co, err := NewCleanupOrchestrator(tt.auditLogger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, co)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, co)
				assert.NotNil(t, co.auditLogger)
				assert.NotNil(t, co.cleanupTasks)
				assert.NotNil(t, co.monitoringChan)
				assert.True(t, co.verificationEnabled)
			}
		})
	}
}

func TestAddCleanupTask(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	task := CleanupTask{
		ResourceType: ResourceTypeTempDir,
		ResourceID:   "test-resource",
		Path:         "/tmp/test-dir",
		Priority:     PriorityMedium,
	}

	co.AddCleanupTask(task)

	pendingTasks := co.GetPendingTasks()
	assert.Len(t, pendingTasks, 1)
	assert.Equal(t, ResourceTypeTempDir, pendingTasks[0].ResourceType)
	assert.Equal(t, "test-resource", pendingTasks[0].ResourceID)
	assert.Equal(t, "/tmp/test-dir", pendingTasks[0].Path)
	assert.Equal(t, PriorityMedium, pendingTasks[0].Priority)
	assert.NotEmpty(t, pendingTasks[0].ID)
	assert.False(t, pendingTasks[0].CreatedAt.IsZero())
	assert.Equal(t, 24*time.Hour, pendingTasks[0].MaxAge)
}

func TestAddCleanupTaskDefaults(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	task := CleanupTask{
		ResourceType: ResourceTypeContainer,
		ResourceID:   "container123",
		ContainerID:  "abc123",
		// ID, CreatedAt, and MaxAge not set - should get defaults
	}

	co.AddCleanupTask(task)

	pendingTasks := co.GetPendingTasks()
	require.Len(t, pendingTasks, 1)

	addedTask := pendingTasks[0]
	assert.NotEmpty(t, addedTask.ID)
	assert.False(t, addedTask.CreatedAt.IsZero())
	assert.Equal(t, 24*time.Hour, addedTask.MaxAge)
}

func TestCleanupTempDirectory(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleanup-test-*")
	require.NoError(t, err)

	// Create a test file in the directory
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(tempDir)
	require.NoError(t, err)

	// Test cleanup
	err = co.cleanupTempDirectory(tempDir)
	assert.NoError(t, err)

	// Verify directory is removed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))
}

func TestCleanupTempDirectoryValidation(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name:    "unsafe path",
			path:    "/etc/passwd",
			wantErr: true,
			errMsg:  "path is not a safe temporary directory",
		},
		{
			name:    "non-existent safe path",
			path:    filepath.Join(os.TempDir(), "non-existent-dir"),
			wantErr: false, // Should not error for non-existent paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := co.cleanupTempDirectory(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteCleanupTask(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cleanup-task-test-*")
	require.NoError(t, err)

	task := CleanupTask{
		ID:           "test-task-1",
		ResourceType: ResourceTypeTempDir,
		ResourceID:   "test-resource",
		Path:         tempDir,
		Priority:     PriorityMedium,
		CreatedAt:    time.Now(),
		MaxAge:       time.Hour,
	}

	ctx := context.Background()
	result := co.executeCleanupTask(ctx, task)

	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.Equal(t, task, result.Task)

	// Verify directory was actually removed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))
}

func TestExecuteCleanup(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Create multiple temporary directories
	tempDir1, err := os.MkdirTemp("", "cleanup-1-*")
	require.NoError(t, err)
	tempDir2, err := os.MkdirTemp("", "cleanup-2-*")
	require.NoError(t, err)

	// Add cleanup tasks with different priorities
	task1 := CleanupTask{
		ResourceType: ResourceTypeTempDir,
		ResourceID:   "resource-1",
		Path:         tempDir1,
		Priority:     PriorityLow,
	}
	task2 := CleanupTask{
		ResourceType: ResourceTypeTempDir,
		ResourceID:   "resource-2",
		Path:         tempDir2,
		Priority:     PriorityHigh,
	}

	co.AddCleanupTask(task1)
	co.AddCleanupTask(task2)

	// Verify tasks were added
	pendingTasks := co.GetPendingTasks()
	assert.Len(t, pendingTasks, 2)

	// Execute cleanup
	ctx := context.Background()
	err = co.ExecuteCleanup(ctx)
	assert.NoError(t, err)

	// Verify all tasks were completed (removed from pending)
	pendingTasks = co.GetPendingTasks()
	assert.Len(t, pendingTasks, 0)

	// Verify directories were actually removed
	_, err = os.Stat(tempDir1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(tempDir2)
	assert.True(t, os.IsNotExist(err))
}

func TestExecuteCleanupNoTasks(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	ctx := context.Background()
	err = co.ExecuteCleanup(ctx)
	assert.NoError(t, err)
}

func TestVerifyCleanup(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Test directory verification
	tempDir, err := os.MkdirTemp("", "verify-test-*")
	require.NoError(t, err)

	task := CleanupTask{
		ResourceType: ResourceTypeTempDir,
		Path:         tempDir,
	}

	// Verification should fail since directory still exists
	err = co.verifyCleanup(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource still exists after cleanup")

	// Remove the directory and verify again
	os.RemoveAll(tempDir)
	err = co.verifyCleanup(task)
	assert.NoError(t, err)
}

func TestSortCleanupTasksByPriority(t *testing.T) {
	tasks := []CleanupTask{
		{ID: "1", Priority: PriorityLow},
		{ID: "2", Priority: PriorityCritical},
		{ID: "3", Priority: PriorityMedium},
		{ID: "4", Priority: PriorityHigh},
	}

	sortCleanupTasksByPriority(tasks)

	// Should be sorted by priority: Critical, High, Medium, Low
	assert.Equal(t, PriorityCritical, tasks[0].Priority)
	assert.Equal(t, "2", tasks[0].ID)
	assert.Equal(t, PriorityHigh, tasks[1].Priority)
	assert.Equal(t, "4", tasks[1].ID)
	assert.Equal(t, PriorityMedium, tasks[2].Priority)
	assert.Equal(t, "3", tasks[2].ID)
	assert.Equal(t, PriorityLow, tasks[3].Priority)
	assert.Equal(t, "1", tasks[3].ID)
}

func TestIsSafeTempPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "system temp directory",
			path:     filepath.Join(os.TempDir(), "test"),
			expected: true,
		},
		{
			name:     "repo sandbox temp",
			path:     "/tmp/repo-sandbox-123",
			expected: true,
		},
		{
			name:     "var tmp repo sandbox",
			path:     "/var/tmp/repo-sandbox-456",
			expected: true,
		},
		{
			name:     "unsafe system path",
			path:     "/etc/passwd",
			expected: false,
		},
		{
			name:     "unsafe home directory",
			path:     "/home/user/documents",
			expected: false,
		},
		{
			name:     "unsafe root directory",
			path:     "/",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSafeTempPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateCleanupTaskID(t *testing.T) {
	id1 := generateCleanupTaskID(ResourceTypeTempDir, "resource1")
	id2 := generateCleanupTaskID(ResourceTypeTempDir, "resource1")

	// IDs should be unique even for same resource
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "temp_directory-resource1")
	assert.Contains(t, id2, "temp_directory-resource1")
}

func TestResourceTypes(t *testing.T) {
	// Test that all resource type constants are properly defined
	types := []ResourceType{
		ResourceTypeTempDir,
		ResourceTypeContainer,
		ResourceTypeLogFile,
		ResourceTypeCache,
	}

	for _, resourceType := range types {
		assert.NotEmpty(t, string(resourceType))
	}
}

func TestCleanupPriorities(t *testing.T) {
	// Test priority ordering
	assert.True(t, PriorityLow < PriorityMedium)
	assert.True(t, PriorityMedium < PriorityHigh)
	assert.True(t, PriorityHigh < PriorityCritical)
}

func TestRegisterHandlers(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Test registering GitHandler (requires *logger.Logger, not *logger.AuditLogger)
	basicLogger := logger.New()
	gitHandler, err := NewGitHandler(basicLogger)
	require.NoError(t, err)
	defer gitHandler.Cleanup()

	co.RegisterGitHandler(gitHandler)
	assert.Equal(t, gitHandler, co.gitHandler)

	// Test registering ContainerOrchestrator (requires *logger.Logger, not *logger.AuditLogger)
	containerOrch, err := NewContainerOrchestrator(basicLogger)
	require.NoError(t, err)

	co.RegisterContainerOrchestrator(containerOrch)
	assert.Equal(t, containerOrch, co.containerOrch)
}

func TestSetVerificationEnabled(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Default should be enabled
	assert.True(t, co.verificationEnabled)

	// Test disabling
	co.SetVerificationEnabled(false)
	assert.False(t, co.verificationEnabled)

	// Test re-enabling
	co.SetVerificationEnabled(true)
	assert.True(t, co.verificationEnabled)
}

func TestMonitorCleanup(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Get monitoring channel
	monitorChan := co.MonitorCleanup()
	assert.NotNil(t, monitorChan)

	// Channel should be accessible and have the correct type
	assert.IsType(t, (<-chan CleanupResult)(nil), monitorChan)
}

func TestCleanupOnExit(t *testing.T) {
	co, err := NewCleanupOrchestrator(createTestAuditLogger(t))
	require.NoError(t, err)

	// Add a cleanup task
	tempDir, err := os.MkdirTemp("", "exit-cleanup-test-*")
	require.NoError(t, err)

	task := CleanupTask{
		ResourceType: ResourceTypeTempDir,
		ResourceID:   "exit-test",
		Path:         tempDir,
		Priority:     PriorityHigh,
	}

	co.AddCleanupTask(task)

	// Execute cleanup on exit
	co.CleanupOnExit()

	// Verify cleanup was performed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))

	// Verify monitoring channel is closed
	// The channel might have received cleanup results before closure, so we need to drain it
	monitorChan := co.MonitorCleanup()
	var closed bool
	for {
		select {
		case _, ok := <-monitorChan:
			if !ok {
				closed = true
				break
			}
			// Continue draining the channel
		case <-time.After(100 * time.Millisecond):
			// Timeout - channel might be open but empty
			break
		}
		if closed {
			break
		}
	}

	if !closed {
		// Try one more read to see if channel is closed
		_, ok := <-monitorChan
		assert.False(t, ok, "Monitoring channel should be closed")
	}
}

// Helper function to create a test audit logger
func createTestAuditLogger(t *testing.T) *logger.AuditLogger {
	auditLogger, err := logger.NewAuditLogger()
	require.NoError(t, err)
	return auditLogger
}
