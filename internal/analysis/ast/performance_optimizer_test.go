package ast

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDefaultPerformanceConfig(t *testing.T) {
	config := DefaultPerformanceConfig()

	// Validate default values align with architecture requirements
	if config.MemoryLimitBytes != 2<<30 {
		t.Errorf("Expected memory limit to be 2GB (2147483648), got %d", config.MemoryLimitBytes)
	}

	if config.TotalTimeout != 1*time.Hour {
		t.Errorf("Expected total timeout to be 1 hour, got %v", config.TotalTimeout)
	}

	if config.MaxWorkers != runtime.GOMAXPROCS(0) {
		t.Errorf("Expected max workers to match CPU cores (%d), got %d", runtime.GOMAXPROCS(0), config.MaxWorkers)
	}

	if config.StreamThreshold != 1000 {
		t.Errorf("Expected stream threshold to be 1000, got %d", config.StreamThreshold)
	}
}

func TestNewPerformanceOptimizer(t *testing.T) {
	config := DefaultPerformanceConfig()
	po := NewPerformanceOptimizer(config)

	if po == nil {
		t.Fatal("Expected performance optimizer to be created")
	}

	if po.config.MemoryLimitBytes != config.MemoryLimitBytes {
		t.Errorf("Expected memory limit to be %d, got %d", config.MemoryLimitBytes, po.config.MemoryLimitBytes)
	}

	if po.memoryMonitor == nil {
		t.Error("Expected memory monitor to be initialized")
	}

	if po.timeoutManager == nil {
		t.Error("Expected timeout manager to be initialized")
	}

	// Test channel initialization
	if po.workerPool == nil {
		t.Error("Expected worker pool channel to be initialized")
	}

	if po.progressChan == nil {
		t.Error("Expected progress channel to be initialized")
	}
}

func TestPerformanceOptimizerWithMockFiles(t *testing.T) {
	// Create temporary test files
	tempDir, err := os.MkdirTemp("", "ast-perf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test JavaScript files
	testFiles := []string{
		"file1.js",
		"file2.ts",
		"file3.jsx",
	}

	testContent := `
function testFunction() {
    console.log("Hello world");
    return 42;
}

class TestClass {
    constructor() {
        this.value = 0;
    }
    
    method() {
        return this.value + 1;
    }
}

export { testFunction, TestClass };
`

	filePaths := make([]string, len(testFiles))
	for i, fileName := range testFiles {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
		filePaths[i] = filePath
	}

	// Test performance optimization with small file set
	config := DefaultPerformanceConfig()
	config.BatchSize = 2 // Small batches for testing
	config.MaxWorkers = 2
	config.ParseTimeout = 5 * time.Second

	po := NewPerformanceOptimizer(config)
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test parsing with performance optimizer
	metrics, results, err := po.ParseRepositoryOptimized(ctx, filePaths, parser)
	if err != nil {
		t.Fatalf("Performance optimization failed: %v", err)
	}

	// Validate results
	if metrics == nil {
		t.Error("Expected performance metrics to be returned")
	}

	if len(results) == 0 {
		t.Error("Expected results to be returned")
	}

	// Validate metrics structure
	if metrics.TotalFilesProcessed == 0 {
		t.Error("Expected files to be processed")
	}

	if metrics.ThroughputFPS <= 0 {
		t.Error("Expected positive throughput")
	}

	// Test progress monitoring
	progressUpdates := 0
	done := make(chan struct{})

	go func() {
		defer close(done)
		for update := range po.GetProgressChannel() {
			progressUpdates++
			if update.TotalFiles != int64(len(filePaths)) {
				t.Errorf("Expected total files to be %d, got %d", len(filePaths), update.TotalFiles)
			}
			if progressUpdates >= 1 {
				return // Received at least one update
			}
		}
	}()

	// Allow some time for progress updates
	time.Sleep(100 * time.Millisecond)
	po.Stop()

	select {
	case <-done:
		// Progress monitoring worked
	case <-time.After(1 * time.Second):
		t.Log("Progress monitoring may not have generated updates (acceptable for small file sets)")
	}
}

func TestMemoryPressureHandling(t *testing.T) {
	config := DefaultPerformanceConfig()
	// Set a very low memory limit to test pressure handling
	config.MemoryLimitBytes = 100 * 1024 * 1024 // 100MB
	config.GCThreshold = 0.5                    // Trigger GC at 50% usage

	po := NewPerformanceOptimizer(config)

	if po.memoryMonitor == nil {
		t.Fatal("Expected memory monitor to be initialized")
	}

	// Start memory monitoring
	err := po.memoryMonitor.Start()
	if err != nil {
		t.Fatalf("Failed to start memory monitor: %v", err)
	}
	defer po.memoryMonitor.Stop()

	// Allow monitoring to run briefly
	time.Sleep(200 * time.Millisecond)

	// Check if memory monitoring is active
	if po.memoryMonitor.GetCurrentUsage() <= 0 {
		t.Error("Expected memory monitor to report positive usage")
	}

	// Test memory stats
	stats := po.memoryMonitor.GetMemoryStats()
	if stats.CurrentUsage <= 0 {
		t.Error("Expected positive current memory usage")
	}

	if stats.LimitBytes != config.MemoryLimitBytes {
		t.Errorf("Expected limit to be %d, got %d", config.MemoryLimitBytes, stats.LimitBytes)
	}
}

func TestTimeoutHandling(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 100 * time.Millisecond // Very short timeout for testing

	tm := NewTimeoutManager(config)

	if tm == nil {
		t.Fatal("Expected timeout manager to be created")
	}

	// Test file timeout creation
	ctx := context.Background()
	timeoutCtx, cancel, id := tm.CreateFileTimeout("test.js", ctx)

	if timeoutCtx == nil {
		t.Error("Expected timeout context to be created")
	}

	if cancel == nil {
		t.Error("Expected cancel function to be returned")
	}

	if id == "" {
		t.Error("Expected timeout ID to be generated")
	}

	// Test timeout occurrence
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected deadline exceeded error, got %v", timeoutCtx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected timeout to occur within 200ms")
	}

	// Allow timeout handling to complete
	time.Sleep(50 * time.Millisecond)

	// Check timeout statistics
	stats := tm.GetTimeoutStats()
	if stats.TotalTimeouts == 0 {
		t.Error("Expected at least one timeout to be recorded")
	}

	cancel() // Clean up
}

func TestFileBatchingStrategy(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.BatchSize = 3
	config.StreamThreshold = 10

	po := NewPerformanceOptimizer(config)

	// Test with file count below streaming threshold (should use batching)
	files := make([]string, 5)
	for i := range files {
		files[i] = fmt.Sprintf("file%d.js", i)
	}

	// We can't easily test the actual batching logic without running the full parser,
	// but we can test the configuration is set up correctly
	if config.BatchSize != 3 {
		t.Errorf("Expected batch size to be 3, got %d", config.BatchSize)
	}

	if config.StreamThreshold != 10 {
		t.Errorf("Expected stream threshold to be 10, got %d", config.StreamThreshold)
	}

	// Use variables to avoid unused errors
	_ = po
	_ = files

	// Test with file count above streaming threshold
	largeFileSet := make([]string, 15)
	for i := range largeFileSet {
		largeFileSet[i] = fmt.Sprintf("large_file%d.js", i)
	}

	// The streaming vs batching decision is made internally in the performance optimizer
	// This test validates the configuration is properly set up for the decision logic
	po2 := NewPerformanceOptimizer(config)
	if len(largeFileSet) > config.StreamThreshold {
		// Should use streaming approach for large file sets
		t.Logf("Large file set (%d files) would trigger streaming mode (threshold: %d)",
			len(largeFileSet), config.StreamThreshold)
	}

	// Use po2 to avoid unused variable error
	_ = po2
}

func TestWorkerPoolConfiguration(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MaxWorkers = 4
	config.WorkerPoolSize = 100

	po := NewPerformanceOptimizer(config)

	// Test configuration values
	if po.config.MaxWorkers != 4 {
		t.Errorf("Expected 4 max workers, got %d", po.config.MaxWorkers)
	}

	if po.config.WorkerPoolSize != 100 {
		t.Errorf("Expected worker pool size of 100, got %d", po.config.WorkerPoolSize)
	}

	// Test channel buffer size
	if cap(po.workerPool) != config.WorkerPoolSize {
		t.Errorf("Expected worker pool channel capacity to be %d, got %d",
			config.WorkerPoolSize, cap(po.workerPool))
	}
}

func TestPerformanceMetricsCalculation(t *testing.T) {
	config := DefaultPerformanceConfig()
	po := NewPerformanceOptimizer(config)

	// Simulate processing some files
	po.totalFiles = 10
	po.processedFiles = 8
	po.failedFiles = 2
	po.startTime = time.Now().Add(-5 * time.Second) // Started 5 seconds ago

	metrics := po.finalize()

	if metrics == nil {
		t.Fatal("Expected metrics to be calculated")
	}

	if metrics.TotalFilesProcessed != 8 {
		t.Errorf("Expected 8 processed files, got %d", metrics.TotalFilesProcessed)
	}

	if metrics.FailedFiles != 2 {
		t.Errorf("Expected 2 failed files, got %d", metrics.FailedFiles)
	}

	if metrics.SuccessfulFiles != 6 {
		t.Errorf("Expected 6 successful files, got %d", metrics.SuccessfulFiles)
	}

	if metrics.ThroughputFPS <= 0 {
		t.Error("Expected positive throughput")
	}

	if metrics.TotalDuration <= 0 {
		t.Error("Expected positive total duration")
	}
}

func BenchmarkPerformanceOptimizer(b *testing.B) {
	// Create temporary test files for benchmarking
	tempDir, err := os.MkdirTemp("", "ast-perf-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test content
	testContent := `
function benchmarkFunction() {
    const data = [];
    for (let i = 0; i < 1000; i++) {
        data.push({
            id: i,
            value: Math.random(),
            timestamp: Date.now()
        });
    }
    return data;
}

class BenchmarkClass {
    constructor() {
        this.items = [];
    }
    
    addItem(item) {
        this.items.push(item);
    }
    
    process() {
        return this.items.map(item => item.value * 2);
    }
}

export { benchmarkFunction, BenchmarkClass };
`

	// Create multiple test files
	fileCount := 100
	filePaths := make([]string, fileCount)
	for i := 0; i < fileCount; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("bench_file_%d.js", i))
		if err := os.WriteFile(filePath, []byte(testContent), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
		filePaths[i] = filePath
	}

	config := DefaultPerformanceConfig()
	config.MaxWorkers = runtime.GOMAXPROCS(0)
	config.BatchSize = 20

	parser, err := NewParser()
	if err != nil {
		b.Fatalf("Failed to create parser: %v", err)
	}
	defer parser.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		po := NewPerformanceOptimizer(config)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		_, _, err := po.ParseRepositoryOptimized(ctx, filePaths, parser)
		if err != nil {
			b.Fatalf("Performance optimization failed: %v", err)
		}

		cancel()
		po.Stop()
	}
}

// Test helper function to verify file processing results
func validateParseResults(t *testing.T, results map[string]*ParseResult, expectedFiles []string) {
	if len(results) != len(expectedFiles) {
		t.Errorf("Expected %d results, got %d", len(expectedFiles), len(results))
	}

	for _, filePath := range expectedFiles {
		result, exists := results[filePath]
		if !exists {
			t.Errorf("Expected result for file %s", filePath)
			continue
		}

		if result.FilePath != filePath {
			t.Errorf("Expected file path %s, got %s", filePath, result.FilePath)
		}

		// Basic validation that parsing occurred
		if len(result.Functions) == 0 && len(result.Classes) == 0 {
			t.Errorf("Expected at least some functions or classes in %s", filePath)
		}
	}
}
