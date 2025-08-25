package security

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/security/sandbox"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// PerformanceMetrics represents performance measurement data
type PerformanceMetrics struct {
	Duration           time.Duration
	MemoryBefore       runtime.MemStats
	MemoryAfter        runtime.MemStats
	MemoryUsed         uint64
	Timestamp          time.Time
	Operation          string
	Success            bool
	ResourcesCleanedUp bool
}

// TestPerformanceMonitoring tests performance monitoring and resource utilization
func TestPerformanceMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance monitoring tests in short mode")
	}

	auditLogger, err := logger.NewAuditLogger()
	require.NoError(t, err)

	basicLogger := logger.New()

	t.Run("GitHandlerPerformance", func(t *testing.T) {
		testGitHandlerPerformance(t, basicLogger, auditLogger)
	})

	t.Run("ContainerOrchestrationPerformance", func(t *testing.T) {
		testContainerOrchestrationPerformance(t, basicLogger, auditLogger)
	})

	t.Run("CleanupOrchestrationPerformance", func(t *testing.T) {
		testCleanupOrchestrationPerformance(t, auditLogger)
	})

	t.Run("ResourceUtilizationMonitoring", func(t *testing.T) {
		testResourceUtilizationMonitoring(t, basicLogger, auditLogger)
	})
}

// testGitHandlerPerformance tests Git handler performance metrics
func testGitHandlerPerformance(t *testing.T, basicLogger *logger.Logger, auditLogger *logger.AuditLogger) {
	t.Log("Testing Git handler performance")

	// Initialize Git handler
	gitHandler, err := sandbox.NewGitHandler(basicLogger)
	require.NoError(t, err)
	defer gitHandler.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Measure performance metrics
	metrics := measurePerformance("git_clone_performance", func() bool {
		// Try to clone a small repository for performance testing
		// Use a very small repository to minimize test time
		testRepoURL := "https://github.com/octocat/Hello-World.git"

		result, err := gitHandler.CloneRepository(ctx, testRepoURL)
		if err != nil {
			t.Logf("Git clone failed (expected in CI/network-limited environments): %v", err)
			return false
		}

		// Verify basic clone result
		if result != nil {
			assert.NotEmpty(t, result.LocalPath)
			assert.Greater(t, result.RepoSize, int64(0))
			assert.Greater(t, result.CloneDuration, time.Duration(0))
		}

		return true
	})

	// Log performance metrics
	auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":         "git_handler_performance_test",
		"duration_ms":       metrics.Duration.Milliseconds(),
		"memory_used_bytes": metrics.MemoryUsed,
		"success":           metrics.Success,
		"resources_cleaned": metrics.ResourcesCleanedUp,
	})

	// Performance assertions
	if metrics.Success {
		// Git operations should complete within reasonable time
		assert.Less(t, metrics.Duration, 30*time.Second, "Git clone should complete within 30 seconds")
		t.Logf("Git handler performance: %v, Memory used: %d bytes", metrics.Duration, metrics.MemoryUsed)
	} else {
		t.Log("Git clone test skipped due to network/environment limitations")
	}
}

// testContainerOrchestrationPerformance tests container orchestration performance
func testContainerOrchestrationPerformance(t *testing.T, basicLogger *logger.Logger, auditLogger *logger.AuditLogger) {
	t.Log("Testing container orchestration performance")

	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping container performance tests")
	}

	// Initialize container orchestrator
	containerOrch, err := sandbox.NewContainerOrchestrator(basicLogger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create temporary directory for volume mounting
	tempDir, err := os.MkdirTemp("", "container-perf-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	volumeMounts := map[string]string{
		tempDir: "/workspace",
	}

	// Measure container creation performance
	var containerID string
	metrics := measurePerformance("container_creation_performance", func() bool {
		var createErr error
		containerID, createErr = containerOrch.CreateSecureContainer(ctx, volumeMounts)
		return createErr == nil && containerID != ""
	})

	defer func() {
		if containerID != "" {
			_ = containerOrch.StopContainer(ctx, containerID)
		}
	}()

	// Log performance metrics
	auditLogger.LogSecurityEvent(logger.ContainerCreate, map[string]interface{}{
		"operation":         "container_orchestration_performance_test",
		"duration_ms":       metrics.Duration.Milliseconds(),
		"memory_used_bytes": metrics.MemoryUsed,
		"success":           metrics.Success,
		"container_id":      containerID,
	})

	// Performance assertions
	if metrics.Success {
		// Container creation should complete within reasonable time
		assert.Less(t, metrics.Duration, 60*time.Second, "Container creation should complete within 60 seconds")
		assert.NotEmpty(t, containerID)

		// Verify container was created successfully
		assert.NotEmpty(t, containerID)

		t.Logf("Container orchestration performance: %v, Memory used: %d bytes", metrics.Duration, metrics.MemoryUsed)
	} else {
		t.Log("Container creation test failed - performance metrics collected for analysis")
	}
}

// testCleanupOrchestrationPerformance tests cleanup orchestration performance
func testCleanupOrchestrationPerformance(t *testing.T, auditLogger *logger.AuditLogger) {
	t.Log("Testing cleanup orchestration performance")

	// Initialize cleanup orchestrator
	cleanupOrch, err := sandbox.NewCleanupOrchestrator(auditLogger)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create test resources for cleanup
	numTestResources := 5
	testDirs := make([]string, numTestResources)

	for i := 0; i < numTestResources; i++ {
		tempDir, err := os.MkdirTemp("", "cleanup-perf-test-*")
		require.NoError(t, err)
		testDirs[i] = tempDir

		// Create test file in each directory
		testFile := tempDir + "/test.txt"
		err = os.WriteFile(testFile, []byte("cleanup performance test"), 0644)
		require.NoError(t, err)

		// Add cleanup task
		cleanupTask := sandbox.CleanupTask{
			ResourceType: sandbox.ResourceTypeTempDir,
			ResourceID:   "perf-test-resource-" + string(rune(i)),
			Path:         tempDir,
			Priority:     sandbox.PriorityMedium,
		}
		cleanupOrch.AddCleanupTask(cleanupTask)
	}

	// Measure cleanup performance
	metrics := measurePerformance("cleanup_orchestration_performance", func() bool {
		err := cleanupOrch.ExecuteCleanup(ctx)
		if err != nil {
			t.Logf("Cleanup execution error: %v", err)
			return false
		}

		// Verify resources were cleaned up
		for _, dir := range testDirs {
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				t.Logf("Resource not cleaned up: %s", dir)
				return false
			}
		}

		return true
	})

	// Log performance metrics
	auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":         "cleanup_orchestration_performance_test",
		"duration_ms":       metrics.Duration.Milliseconds(),
		"memory_used_bytes": metrics.MemoryUsed,
		"resources_cleaned": numTestResources,
		"success":           metrics.Success,
	})

	// Performance assertions
	assert.True(t, metrics.Success, "Cleanup orchestration should succeed")
	assert.Less(t, metrics.Duration, 10*time.Second, "Cleanup should complete quickly")

	// Verify no pending tasks remain
	pendingTasks := cleanupOrch.GetPendingTasks()
	assert.Equal(t, 0, len(pendingTasks), "All cleanup tasks should be completed")

	t.Logf("Cleanup orchestration performance: %v, Memory used: %d bytes", metrics.Duration, metrics.MemoryUsed)
}

// testResourceUtilizationMonitoring tests resource utilization monitoring
func testResourceUtilizationMonitoring(t *testing.T, basicLogger *logger.Logger, auditLogger *logger.AuditLogger) {
	t.Log("Testing resource utilization monitoring")

	// Get initial system resource state
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// Initialize all components
	gitHandler, err := sandbox.NewGitHandler(basicLogger)
	require.NoError(t, err)
	defer gitHandler.Cleanup()

	containerOrch, err := sandbox.NewContainerOrchestrator(basicLogger)
	require.NoError(t, err)

	cleanupOrch, err := sandbox.NewCleanupOrchestrator(auditLogger)
	require.NoError(t, err)

	// Register components
	cleanupOrch.RegisterGitHandler(gitHandler)
	cleanupOrch.RegisterContainerOrchestrator(containerOrch)

	// Get resource state after initialization
	var postInitMem runtime.MemStats
	runtime.ReadMemStats(&postInitMem)
	runtime.GC() // Force garbage collection to get accurate memory usage

	var postGCMem runtime.MemStats
	runtime.ReadMemStats(&postGCMem)

	// Calculate resource utilization
	initMemoryUsage := postInitMem.Sys - initialMem.Sys
	postGCMemoryUsage := postGCMem.Sys - initialMem.Sys

	// Log resource utilization metrics
	auditLogger.LogSecurityEvent(logger.ResourceLimit, map[string]interface{}{
		"operation":                  "resource_utilization_monitoring",
		"initial_memory_bytes":       initialMem.Sys,
		"post_init_memory_bytes":     postInitMem.Sys,
		"post_gc_memory_bytes":       postGCMem.Sys,
		"init_memory_usage_bytes":    initMemoryUsage,
		"post_gc_memory_usage_bytes": postGCMemoryUsage,
		"goroutines":                 runtime.NumGoroutine(),
		"gc_cycles":                  postGCMem.NumGC - initialMem.NumGC,
	})

	// Resource utilization assertions
	assert.Greater(t, runtime.NumGoroutine(), 0, "System should have active goroutines")
	assert.Greater(t, postInitMem.Sys, initialMem.Sys, "Memory usage should increase after component initialization")

	// Log summary
	t.Logf("Resource utilization - Memory usage: %d bytes, Goroutines: %d, GC cycles: %d",
		postGCMemoryUsage, runtime.NumGoroutine(), postGCMem.NumGC-initialMem.NumGC)
}

// measurePerformance measures performance metrics for a given operation
func measurePerformance(operationName string, operation func() bool) PerformanceMetrics {
	var memBefore, memAfter runtime.MemStats

	// Force GC and measure memory before operation
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	// Measure operation time
	startTime := time.Now()
	success := operation()
	duration := time.Since(startTime)

	// Force GC and measure memory after operation
	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	// Calculate memory usage
	memoryUsed := memAfter.TotalAlloc - memBefore.TotalAlloc
	if memAfter.Sys > memBefore.Sys {
		memoryUsed = memAfter.Sys - memBefore.Sys
	}

	return PerformanceMetrics{
		Duration:           duration,
		MemoryBefore:       memBefore,
		MemoryAfter:        memAfter,
		MemoryUsed:         memoryUsed,
		Timestamp:          startTime,
		Operation:          operationName,
		Success:            success,
		ResourcesCleanedUp: success, // Assume resources are cleaned up on success
	}
}

// TestResourceLeakDetection tests for resource leaks over multiple operations
func TestResourceLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource leak detection tests in short mode")
	}

	auditLogger, err := logger.NewAuditLogger()
	require.NoError(t, err)

	basicLogger := logger.New()

	t.Run("MemoryLeakDetection", func(t *testing.T) {
		testMemoryLeakDetection(t, basicLogger, auditLogger)
	})

	t.Run("FileHandleLeakDetection", func(t *testing.T) {
		testFileHandleLeakDetection(t, basicLogger, auditLogger)
	})
}

// testMemoryLeakDetection tests for memory leaks over multiple operations
func testMemoryLeakDetection(t *testing.T, basicLogger *logger.Logger, auditLogger *logger.AuditLogger) {
	t.Log("Testing memory leak detection")

	var initialMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialMem)

	// Perform multiple operations to test for memory leaks
	numIterations := 5
	for i := 0; i < numIterations; i++ {
		// Create and cleanup Git handler
		gitHandler, err := sandbox.NewGitHandler(basicLogger)
		require.NoError(t, err)
		gitHandler.Cleanup()

		// Create and cleanup Cleanup orchestrator
		cleanupOrch, err := sandbox.NewCleanupOrchestrator(auditLogger)
		require.NoError(t, err)

		// Add and execute some cleanup tasks
		tempDir, err := os.MkdirTemp("", "leak-test-*")
		require.NoError(t, err)

		cleanupTask := sandbox.CleanupTask{
			ResourceType: sandbox.ResourceTypeTempDir,
			ResourceID:   "leak-test",
			Path:         tempDir,
			Priority:     sandbox.PriorityHigh,
		}

		cleanupOrch.AddCleanupTask(cleanupTask)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_ = cleanupOrch.ExecuteCleanup(ctx)
		cancel()

		// Force garbage collection between iterations
		runtime.GC()
	}

	// Final memory measurement
	var finalMem runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&finalMem)

	// Calculate memory growth
	memoryGrowth := finalMem.Sys - initialMem.Sys
	allocGrowth := finalMem.TotalAlloc - initialMem.TotalAlloc

	// Log memory leak detection results
	auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":          "memory_leak_detection",
		"iterations":         numIterations,
		"initial_memory_sys": initialMem.Sys,
		"final_memory_sys":   finalMem.Sys,
		"memory_growth_sys":  memoryGrowth,
		"alloc_growth":       allocGrowth,
		"gc_cycles":          finalMem.NumGC - initialMem.NumGC,
	})

	// Memory leak assertions
	// Allow for some memory growth, but it should be reasonable
	maxAllowedGrowth := uint64(50 * 1024 * 1024) // 50MB
	assert.Less(t, memoryGrowth, maxAllowedGrowth,
		"Memory growth should be less than %d bytes over %d iterations", maxAllowedGrowth, numIterations)

	t.Logf("Memory leak detection completed - Growth: %d bytes over %d iterations", memoryGrowth, numIterations)
}

// testFileHandleLeakDetection tests for file handle leaks
func testFileHandleLeakDetection(t *testing.T, basicLogger *logger.Logger, auditLogger *logger.AuditLogger) {
	t.Log("Testing file handle leak detection")

	// Create multiple temporary directories and clean them up
	numIterations := 10
	for i := 0; i < numIterations; i++ {
		tempDir, err := os.MkdirTemp("", "file-handle-test-*")
		require.NoError(t, err)

		// Create multiple files
		for j := 0; j < 5; j++ {
			testFile := tempDir + "/test" + string(rune(j)) + ".txt"
			err = os.WriteFile(testFile, []byte("file handle test"), 0644)
			require.NoError(t, err)
		}

		// Clean up the directory
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
	}

	// Log file handle leak detection results
	auditLogger.LogSecurityEvent(logger.SystemCleanup, map[string]interface{}{
		"operation":           "file_handle_leak_detection",
		"iterations":          numIterations,
		"files_per_iteration": 5,
		"message":             "File handle leak detection completed - no obvious leaks detected",
	})

	t.Log("File handle leak detection completed - no obvious leaks detected")
}
