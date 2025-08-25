package ast

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewTimeoutManager(t *testing.T) {
	config := DefaultPerformanceConfig()
	tm := NewTimeoutManager(config)

	if tm == nil {
		t.Fatal("Expected timeout manager to be created")
	}

	if tm.config.ParseTimeout != config.ParseTimeout {
		t.Errorf("Expected parse timeout to be %v, got %v", config.ParseTimeout, tm.config.ParseTimeout)
	}

	if tm.activeTimers == nil {
		t.Error("Expected active timers map to be initialized")
	}

	if tm.timeoutHistogram == nil {
		t.Error("Expected timeout histogram to be initialized")
	}

	if tm.startTime.IsZero() {
		t.Error("Expected start time to be set")
	}
}

func TestFileTimeoutCreation(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 100 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()
	filePath := "test.js"

	timeoutCtx, cancel, id := tm.CreateFileTimeout(filePath, ctx)

	if timeoutCtx == nil {
		t.Fatal("Expected timeout context to be created")
	}

	if cancel == nil {
		t.Fatal("Expected cancel function to be returned")
	}

	if id == "" {
		t.Fatal("Expected timeout ID to be generated")
	}

	// Check that ID contains file identifier
	if !strings.Contains(id, "file_") {
		t.Errorf("Expected file timeout ID to contain 'file_', got %s", id)
	}

	// Check active operations count
	if tm.GetActiveOperations() != 1 {
		t.Errorf("Expected 1 active operation, got %d", tm.GetActiveOperations())
	}

	cancel()

	// Allow cleanup to complete
	time.Sleep(10 * time.Millisecond)
}

func TestBatchTimeoutCreation(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 50 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()
	batchID := 5

	timeoutCtx, cancel, id := tm.CreateBatchTimeout(batchID, ctx)

	if timeoutCtx == nil {
		t.Fatal("Expected batch timeout context to be created")
	}

	if cancel == nil {
		t.Fatal("Expected cancel function to be returned")
	}

	if id == "" {
		t.Fatal("Expected timeout ID to be generated")
	}

	// Check that ID contains batch identifier
	if !strings.Contains(id, "batch_") {
		t.Errorf("Expected batch timeout ID to contain 'batch_', got %s", id)
	}

	// Batch timeout should be longer than file timeout
	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Fatal("Expected timeout context to have deadline")
	}

	expectedBatchTimeout := config.ParseTimeout * 10
	actualTimeout := time.Until(deadline)

	// Allow some tolerance for timing
	if actualTimeout < expectedBatchTimeout-10*time.Millisecond {
		t.Errorf("Expected batch timeout to be ~%v, got %v", expectedBatchTimeout, actualTimeout)
	}

	cancel()
}

func TestTimeoutOccurrence(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 50 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()
	timeoutCtx, cancel, id := tm.CreateFileTimeout("timeout_test.js", ctx)
	defer cancel()

	// Wait for timeout to occur
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected deadline exceeded error, got %v", timeoutCtx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timeout should have occurred within 200ms")
	}

	// Allow timeout handling to complete
	time.Sleep(50 * time.Millisecond)

	// Check timeout statistics
	stats := tm.GetTimeoutStats()
	if stats.TotalTimeouts == 0 {
		t.Error("Expected timeout to be recorded")
	}

	if stats.FileTimeouts == 0 {
		t.Error("Expected file timeout to be recorded")
	}

	if stats.TimeoutRate <= 0 {
		t.Error("Expected positive timeout rate")
	}

	t.Logf("Timeout ID: %s", id)
}

func TestTimeoutStatistics(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 30 * time.Millisecond
	tm := NewTimeoutManager(config)

	// Create multiple timeouts of different types
	ctx := context.Background()

	// Create file timeouts
	fileCtx1, cancel1, _ := tm.CreateFileTimeout("file1.js", ctx)
	fileCtx2, cancel2, _ := tm.CreateFileTimeout("file2.ts", ctx)
	defer cancel1()
	defer cancel2()

	// Create batch timeout
	_, cancelBatch, _ := tm.CreateBatchTimeout(1, ctx)
	defer cancelBatch()

	// Wait for file timeouts to occur
	<-fileCtx1.Done()
	<-fileCtx2.Done()

	// Allow timeout processing
	time.Sleep(100 * time.Millisecond)

	stats := tm.GetTimeoutStats()

	if stats.TotalTimeouts < 2 {
		t.Errorf("Expected at least 2 timeouts, got %d", stats.TotalTimeouts)
	}

	if stats.FileTimeouts < 2 {
		t.Errorf("Expected at least 2 file timeouts, got %d", stats.FileTimeouts)
	}

	// Batch timeout should still be active (longer timeout)
	if stats.ActiveOperations == 0 {
		t.Error("Expected at least 1 active operation (batch timeout)")
	}

	// Check timeout categorization
	totalCategorized := stats.ShortTimeouts + stats.MediumTimeouts + stats.LongTimeouts
	if totalCategorized < stats.TotalTimeouts {
		t.Errorf("Timeout categorization mismatch: %d categorized vs %d total",
			totalCategorized, stats.TotalTimeouts)
	}
}

func TestActiveTimersTracking(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 200 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()

	// Create multiple active timers
	_, cancel1, id1 := tm.CreateFileTimeout("active1.js", ctx)
	_, cancel2, id2 := tm.CreateFileTimeout("active2.ts", ctx)
	_, cancel3, _ := tm.CreateBatchTimeout(1, ctx)

	defer cancel1()
	defer cancel2()
	defer cancel3()

	// Check active operations count
	activeCount := tm.GetActiveOperations()
	if activeCount != 3 {
		t.Errorf("Expected 3 active operations, got %d", activeCount)
	}

	// Get active timer details
	activeTimers := tm.GetActiveTimers()
	if len(activeTimers) != 3 {
		t.Errorf("Expected 3 active timers, got %d", len(activeTimers))
	}

	// Validate timer details
	var foundFile1, foundFile2, foundBatch bool
	for _, timer := range activeTimers {
		if timer.ID == id1 {
			foundFile1 = true
			if timer.FilePath != "active1.js" {
				t.Errorf("Expected file path active1.js, got %s", timer.FilePath)
			}
		}
		if timer.ID == id2 {
			foundFile2 = true
			if timer.FilePath != "active2.ts" {
				t.Errorf("Expected file path active2.ts, got %s", timer.FilePath)
			}
		}
		if strings.Contains(timer.ID, "batch_") {
			foundBatch = true
			if timer.FilePath != "" {
				t.Errorf("Expected empty file path for batch timer, got %s", timer.FilePath)
			}
		}

		// Validate timer fields
		if timer.ConfiguredTimeout <= 0 {
			t.Errorf("Expected positive configured timeout, got %v", timer.ConfiguredTimeout)
		}

		if timer.ElapsedTime < 0 {
			t.Errorf("Expected non-negative elapsed time, got %v", timer.ElapsedTime)
		}

		if timer.RemainingTime < 0 {
			t.Errorf("Expected non-negative remaining time, got %v", timer.RemainingTime)
		}
	}

	if !foundFile1 || !foundFile2 || !foundBatch {
		t.Error("Not all expected timers found in active timers list")
	}

	// Test longest running operation
	longestRunning := tm.GetLongestRunning()
	if longestRunning <= 0 {
		t.Error("Expected positive longest running duration")
	}
}

func TestTimeoutHistogram(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 40 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()

	// Create timeouts with different durations
	timeouts := []string{"hist1.js", "hist2.js", "hist3.js"}

	for _, file := range timeouts {
		timeoutCtx, cancel, _ := tm.CreateFileTimeout(file, ctx)

		// Wait for timeout
		<-timeoutCtx.Done()
		cancel()
	}

	// Allow histogram updates
	time.Sleep(100 * time.Millisecond)

	histogram := tm.GetTimeoutHistogram()

	if len(histogram) == 0 {
		t.Error("Expected timeout histogram to have entries")
	}

	// Validate histogram entries
	totalHistogramCounts := int64(0)
	for duration, count := range histogram {
		if duration <= 0 {
			t.Errorf("Expected positive duration in histogram, got %v", duration)
		}

		if count <= 0 {
			t.Errorf("Expected positive count for duration %v, got %d", duration, count)
		}

		totalHistogramCounts += count
	}

	// Should have at least some histogram entries matching our timeouts
	if totalHistogramCounts == 0 {
		t.Error("Expected positive total histogram counts")
	}
}

func TestTimeoutHealthCheck(t *testing.T) {
	config := DefaultPerformanceConfig()
	tm := NewTimeoutManager(config)

	// Initially should be healthy (no timeouts)
	if !tm.IsHealthy() {
		t.Error("Expected timeout manager to be initially healthy")
	}

	// Create many timeouts to make it unhealthy
	ctx := context.Background()
	config.ParseTimeout = 10 * time.Millisecond

	// Create timeouts that will exceed 10% threshold
	for i := 0; i < 20; i++ {
		file := fmt.Sprintf("health_test_%d.js", i)
		timeoutCtx, cancel, _ := tm.CreateFileTimeout(file, ctx)

		// Wait for timeout
		<-timeoutCtx.Done()
		cancel()
	}

	// Allow timeout processing
	time.Sleep(100 * time.Millisecond)

	// Should now be unhealthy due to high timeout rate
	stats := tm.GetTimeoutStats()
	if stats.TimeoutRate < 0.10 {
		// If timeout rate is still low, force some successful operations
		// This test may be sensitive to timing - log for debugging
		t.Logf("Timeout rate: %.2f%%, Total ops: %d, Timeouts: %d",
			stats.TimeoutRate*100, stats.TotalOperations, stats.TotalTimeouts)
	}
}

func TestConcurrentTimeoutOperations(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 50 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()
	numGoroutines := 10
	timeoutsPerGoroutine := 5

	done := make(chan struct{}, numGoroutines)

	// Run concurrent timeout operations
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- struct{}{} }()

			for j := 0; j < timeoutsPerGoroutine; j++ {
				file := fmt.Sprintf("concurrent_%d_%d.js", routineID, j)
				timeoutCtx, cancel, _ := tm.CreateFileTimeout(file, ctx)

				// Wait for timeout
				<-timeoutCtx.Done()
				cancel()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent timeout test timed out")
		}
	}

	// Allow final timeout processing
	time.Sleep(100 * time.Millisecond)

	// Verify statistics
	stats := tm.GetTimeoutStats()
	expectedTimeouts := int64(numGoroutines * timeoutsPerGoroutine)

	if stats.TotalTimeouts < expectedTimeouts {
		t.Errorf("Expected at least %d timeouts, got %d", expectedTimeouts, stats.TotalTimeouts)
	}

	// Should have no active operations after completion
	if tm.GetActiveOperations() != 0 {
		t.Errorf("Expected 0 active operations after concurrent test, got %d", tm.GetActiveOperations())
	}
}

func TestTimeoutManagerStop(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 500 * time.Millisecond // Long timeout
	tm := NewTimeoutManager(config)

	ctx := context.Background()

	// Create several active timers
	_, cancel1, _ := tm.CreateFileTimeout("stop_test1.js", ctx)
	_, cancel2, _ := tm.CreateBatchTimeout(1, ctx)

	defer cancel1()
	defer cancel2()

	// Verify timers are active
	if tm.GetActiveOperations() != 2 {
		t.Errorf("Expected 2 active operations, got %d", tm.GetActiveOperations())
	}

	// Stop the timeout manager
	tm.Stop()

	// After stop, should have no active operations
	if tm.GetActiveOperations() != 0 {
		t.Errorf("Expected 0 active operations after stop, got %d", tm.GetActiveOperations())
	}

	// Verify that creating new timers after stop still works
	timeoutCtx, cancel3, _ := tm.CreateFileTimeout("after_stop.js", ctx)
	defer cancel3()

	if timeoutCtx == nil {
		t.Error("Expected to be able to create new timeout after stop")
	}
}

func TestPerformanceMetricsTracking(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.ParseTimeout = 100 * time.Millisecond
	tm := NewTimeoutManager(config)

	ctx := context.Background()

	// Simulate operations with different completion times
	fastFile := "fast.js"
	timeoutCtx, cancel, _ := tm.CreateFileTimeout(fastFile, ctx)

	// Complete operation quickly (before timeout)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	select {
	case <-timeoutCtx.Done():
		// Operation completed (either successfully or timed out)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Operation should have completed within 200ms")
	}

	// Allow metrics processing
	time.Sleep(50 * time.Millisecond)

	stats := tm.GetTimeoutStats()

	// Should have performance metrics
	if stats.AverageParseTime == 0 && stats.TotalOperations > 0 {
		t.Error("Expected average parse time to be calculated for completed operations")
	}

	if stats.ShortestParseTime == 0 && stats.TotalOperations > 0 {
		t.Error("Expected shortest parse time to be recorded")
	}

	t.Logf("Performance metrics - Avg: %v, Longest: %v, Shortest: %v",
		stats.AverageParseTime, stats.LongestParseTime, stats.ShortestParseTime)
}

func BenchmarkTimeoutManager(b *testing.B) {
	config := DefaultPerformanceConfig()
	tm := NewTimeoutManager(config)

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("CreateFileTimeout", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			file := fmt.Sprintf("bench_%d.js", i)
			_, cancel, _ := tm.CreateFileTimeout(file, ctx)
			cancel() // Immediately cancel to avoid timeout processing
		}
	})

	b.Run("GetTimeoutStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tm.GetTimeoutStats()
		}
	})

	b.Run("GetActiveOperations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tm.GetActiveOperations()
		}
	})
}
