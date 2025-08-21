// Package sandbox provides automatic cleanup and resource management capabilities
// with comprehensive cleanup orchestration, monitoring, and verification systems.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// CleanupOrchestrator manages automatic cleanup of resources with monitoring and verification
type CleanupOrchestrator struct {
	auditLogger     *logger.AuditLogger
	gitHandler      *GitHandler
	containerOrch   *ContainerOrchestrator
	cleanupTasks    []CleanupTask
	cleanupMutex    sync.RWMutex
	monitoringChan  chan CleanupResult
	verificationEnabled bool
}

// CleanupTask represents a cleanup operation
type CleanupTask struct {
	ID           string
	ResourceType ResourceType
	ResourceID   string
	Path         string
	ContainerID  string
	Priority     CleanupPriority
	CreatedAt    time.Time
	MaxAge       time.Duration
	Metadata     map[string]interface{}
}

// ResourceType represents different types of resources that can be cleaned up
type ResourceType string

const (
	ResourceTypeTempDir   ResourceType = "temp_directory"
	ResourceTypeContainer ResourceType = "container"
	ResourceTypeLogFile   ResourceType = "log_file"
	ResourceTypeCache     ResourceType = "cache"
)

// CleanupPriority represents the priority of cleanup operations
type CleanupPriority int

const (
	PriorityLow CleanupPriority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	Task     CleanupTask
	Success  bool
	Error    error
	Duration time.Duration
	Details  map[string]interface{}
}

// CleanupConfig represents cleanup configuration options
type CleanupConfig struct {
	EnableVerification     bool
	MaxConcurrentCleanups  int
	CleanupTimeout         time.Duration
	RetryAttempts          int
	RetryDelay             time.Duration
	MonitoringEnabled      bool
}

// NewCleanupOrchestrator creates a new cleanup orchestrator with secure defaults
func NewCleanupOrchestrator(auditLogger *logger.AuditLogger) (*CleanupOrchestrator, error) {
	if auditLogger == nil {
		return nil, fmt.Errorf("audit logger cannot be nil")
	}

	return &CleanupOrchestrator{
		auditLogger:         auditLogger,
		cleanupTasks:        make([]CleanupTask, 0),
		monitoringChan:      make(chan CleanupResult, 100),
		verificationEnabled: true,
	}, nil
}

// RegisterGitHandler registers a GitHandler for cleanup management
func (co *CleanupOrchestrator) RegisterGitHandler(gh *GitHandler) {
	co.gitHandler = gh
}

// RegisterContainerOrchestrator registers a ContainerOrchestrator for cleanup management
func (co *CleanupOrchestrator) RegisterContainerOrchestrator(containerOrch *ContainerOrchestrator) {
	co.containerOrch = containerOrch
}

// AddCleanupTask adds a cleanup task to the orchestrator
func (co *CleanupOrchestrator) AddCleanupTask(task CleanupTask) {
	co.cleanupMutex.Lock()
	defer co.cleanupMutex.Unlock()

	// Set default values if not provided
	if task.ID == "" {
		task.ID = generateCleanupTaskID(task.ResourceType, task.ResourceID)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.MaxAge == 0 {
		task.MaxAge = 24 * time.Hour // Default 24-hour retention
	}

	co.cleanupTasks = append(co.cleanupTasks, task)

	co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":     "cleanup_task_added",
		"task_id":       task.ID,
		"resource_type": string(task.ResourceType),
		"resource_id":   task.ResourceID,
		"priority":      int(task.Priority),
		"max_age":       task.MaxAge.Seconds(),
	})
}

// ExecuteCleanup executes all pending cleanup tasks with priority ordering
func (co *CleanupOrchestrator) ExecuteCleanup(ctx context.Context) error {
	co.cleanupMutex.Lock()
	tasks := make([]CleanupTask, len(co.cleanupTasks))
	copy(tasks, co.cleanupTasks)
	co.cleanupMutex.Unlock()

	if len(tasks) == 0 {
		co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
			"operation": "cleanup_no_tasks",
			"message":   "No cleanup tasks to execute",
		})
		return nil
	}

	// Sort tasks by priority (highest first)
	sortCleanupTasksByPriority(tasks)

	co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":  "cleanup_execution_start",
		"task_count": len(tasks),
	})

	successCount := 0
	failureCount := 0

	for _, task := range tasks {
		result := co.executeCleanupTask(ctx, task)
		
		if result.Success {
			successCount++
			// Remove successful task from pending tasks
			co.removeCleanupTask(task.ID)
		} else {
			failureCount++
		}

		// Send result to monitoring channel if monitoring is enabled
		select {
		case co.monitoringChan <- result:
		default:
			// Channel is full, continue without blocking
		}
	}

	co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":      "cleanup_execution_complete",
		"success_count":  successCount,
		"failure_count":  failureCount,
		"total_tasks":    len(tasks),
	})

	if failureCount > 0 {
		return fmt.Errorf("cleanup completed with %d failures out of %d tasks", failureCount, len(tasks))
	}

	return nil
}

// executeCleanupTask executes a single cleanup task
func (co *CleanupOrchestrator) executeCleanupTask(ctx context.Context, task CleanupTask) CleanupResult {
	startTime := time.Now()
	
	result := CleanupResult{
		Task:    task,
		Details: make(map[string]interface{}),
	}

	co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":     "cleanup_task_start",
		"task_id":       task.ID,
		"resource_type": string(task.ResourceType),
		"resource_id":   task.ResourceID,
	})

	// Execute cleanup based on resource type
	switch task.ResourceType {
	case ResourceTypeTempDir:
		result.Error = co.cleanupTempDirectory(task.Path)
	case ResourceTypeContainer:
		result.Error = co.cleanupContainer(ctx, task.ContainerID)
	case ResourceTypeLogFile:
		result.Error = co.cleanupLogFile(task.Path)
	case ResourceTypeCache:
		result.Error = co.cleanupCache(task.Path)
	default:
		result.Error = fmt.Errorf("unsupported resource type: %s", task.ResourceType)
	}

	result.Duration = time.Since(startTime)
	result.Success = (result.Error == nil)

	// Perform verification if enabled
	if co.verificationEnabled && result.Success {
		verificationErr := co.verifyCleanup(task)
		if verificationErr != nil {
			result.Success = false
			result.Error = fmt.Errorf("cleanup verification failed: %w", verificationErr)
		}
	}

	// Log the result
	logLevel := "info"
	if !result.Success {
		logLevel = "error"
	}

	logFields := map[string]interface{}{
		"operation":     "cleanup_task_complete",
		"task_id":       task.ID,
		"resource_type": string(task.ResourceType),
		"resource_id":   task.ResourceID,
		"success":       result.Success,
		"duration":      result.Duration.Seconds(),
	}

	if result.Error != nil {
		logFields["error"] = result.Error.Error()
	}

	if logLevel == "error" {
		co.auditLogger.LogSecurityEvent(logger.SystemCleanup, logFields)
	} else {
		co.auditLogger.LogSecurityEvent(logger.SystemCleanup, logFields)
	}

	return result
}

// cleanupTempDirectory removes temporary directories with secure deletion
func (co *CleanupOrchestrator) cleanupTempDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Verify path is safe to delete (must be a temporary directory)
	if !isSafeTempPath(path) {
		return fmt.Errorf("path is not a safe temporary directory: %s", path)
	}

	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // Already cleaned up
	}

	// Remove directory and all contents
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", path, err)
	}

	return nil
}

// cleanupContainer stops and removes Docker containers
func (co *CleanupOrchestrator) cleanupContainer(ctx context.Context, containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	if co.containerOrch == nil {
		// Fallback to direct Docker commands if no orchestrator is registered
		return co.cleanupContainerDirect(ctx, containerID)
	}

	// Use the registered container orchestrator
	return co.containerOrch.StopContainer(ctx, containerID)
}

// cleanupContainerDirect uses direct Docker commands for container cleanup
func (co *CleanupOrchestrator) cleanupContainerDirect(ctx context.Context, containerID string) error {
	// Stop the container
	stopCmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	if _, err := stopCmd.CombinedOutput(); err != nil {
		// Container might already be stopped, continue with removal
	}

	// Remove the container
	rmCmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	if _, err := rmCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove container %s: %w", containerID, err)
	}

	return nil
}

// cleanupLogFile removes old log files
func (co *CleanupOrchestrator) cleanupLogFile(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove log file %s: %w", path, err)
	}

	return nil
}

// cleanupCache removes cache files and directories
func (co *CleanupOrchestrator) cleanupCache(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove cache %s: %w", path, err)
	}

	return nil
}

// verifyCleanup verifies that cleanup was successful
func (co *CleanupOrchestrator) verifyCleanup(task CleanupTask) error {
	switch task.ResourceType {
	case ResourceTypeTempDir, ResourceTypeLogFile, ResourceTypeCache:
		if _, err := os.Stat(task.Path); !os.IsNotExist(err) {
			return fmt.Errorf("resource still exists after cleanup: %s", task.Path)
		}
	case ResourceTypeContainer:
		// Verify container is removed
		cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("id=%s", task.ContainerID), "--format", "{{.ID}}")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to verify container cleanup: %w", err)
		}
		if strings.TrimSpace(string(output)) != "" {
			return fmt.Errorf("container still exists after cleanup: %s", task.ContainerID)
		}
	}
	return nil
}

// CleanupOnExit performs cleanup on application exit with defer patterns
func (co *CleanupOrchestrator) CleanupOnExit() {
	defer func() {
		if r := recover(); r != nil {
			co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
				"operation": "cleanup_panic_recovery",
				"error":     fmt.Sprintf("panic during cleanup: %v", r),
			})
		}
	}()

	ctx := context.Background()
	if err := co.ExecuteCleanup(ctx); err != nil {
		co.auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
			"operation": "cleanup_on_exit_failure",
			"error":     err.Error(),
		})
	}

	// Close monitoring channel
	close(co.monitoringChan)
}

// GetPendingTasks returns the current list of pending cleanup tasks
func (co *CleanupOrchestrator) GetPendingTasks() []CleanupTask {
	co.cleanupMutex.RLock()
	defer co.cleanupMutex.RUnlock()

	tasks := make([]CleanupTask, len(co.cleanupTasks))
	copy(tasks, co.cleanupTasks)
	return tasks
}

// removeCleanupTask removes a task from the pending list
func (co *CleanupOrchestrator) removeCleanupTask(taskID string) {
	co.cleanupMutex.Lock()
	defer co.cleanupMutex.Unlock()

	for i, task := range co.cleanupTasks {
		if task.ID == taskID {
			co.cleanupTasks = append(co.cleanupTasks[:i], co.cleanupTasks[i+1:]...)
			break
		}
	}
}

// generateCleanupTaskID generates a unique ID for a cleanup task
func generateCleanupTaskID(resourceType ResourceType, resourceID string) string {
	return fmt.Sprintf("%s-%s-%d", string(resourceType), resourceID, time.Now().UnixNano())
}

// sortCleanupTasksByPriority sorts cleanup tasks by priority (highest first)
func sortCleanupTasksByPriority(tasks []CleanupTask) {
	for i := 0; i < len(tasks)-1; i++ {
		for j := 0; j < len(tasks)-i-1; j++ {
			if tasks[j].Priority < tasks[j+1].Priority {
				tasks[j], tasks[j+1] = tasks[j+1], tasks[j]
			}
		}
	}
}

// isSafeTempPath checks if a path is safe to delete (must be a temporary directory)
func isSafeTempPath(path string) bool {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check if path is within safe temporary directories
	tempDir := os.TempDir()
	safePrefixes := []string{
		tempDir,
		"/tmp/repo-sandbox",
		"/var/tmp/repo-sandbox",
	}

	for _, prefix := range safePrefixes {
		if strings.HasPrefix(absPath, prefix) {
			return true
		}
	}

	return false
}

// MonitorCleanup returns a channel for monitoring cleanup results
func (co *CleanupOrchestrator) MonitorCleanup() <-chan CleanupResult {
	return co.monitoringChan
}

// SetVerificationEnabled enables or disables cleanup verification
func (co *CleanupOrchestrator) SetVerificationEnabled(enabled bool) {
	co.verificationEnabled = enabled
}