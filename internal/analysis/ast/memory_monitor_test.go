package ast

import (
	"testing"
	"time"
)

func TestNewMemoryMonitor(t *testing.T) {
	config := DefaultPerformanceConfig()
	mm := NewMemoryMonitor(config)

	if mm == nil {
		t.Fatal("Expected memory monitor to be created")
	}

	if mm.config.MemoryLimitBytes != config.MemoryLimitBytes {
		t.Errorf("Expected memory limit to be %d, got %d", config.MemoryLimitBytes, mm.config.MemoryLimitBytes)
	}

	if mm.baselineUsage <= 0 {
		t.Error("Expected positive baseline memory usage")
	}

	if mm.currentUsage <= 0 {
		t.Error("Expected positive current memory usage")
	}

	if mm.maxSamples <= 0 {
		t.Error("Expected positive max samples")
	}
}

func TestMemoryMonitorStartStop(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MemoryCheckInterval = 50 * time.Millisecond
	mm := NewMemoryMonitor(config)

	// Test starting monitoring
	err := mm.Start()
	if err != nil {
		t.Fatalf("Failed to start memory monitor: %v", err)
	}

	// Allow monitoring to run
	time.Sleep(200 * time.Millisecond)

	// Check that monitoring is active
	if mm.monitoring == 0 {
		t.Error("Expected monitoring to be active")
	}

	// Test stopping monitoring
	mm.Stop()

	// Allow stop to complete
	time.Sleep(100 * time.Millisecond)

	if mm.monitoring != 0 {
		t.Error("Expected monitoring to be stopped")
	}

	// Test double start (should be safe)
	err = mm.Start()
	if err != nil {
		t.Fatalf("Failed to restart memory monitor: %v", err)
	}
	defer mm.Stop()

	// Test double stop (should be safe)
	mm.Stop()
	mm.Stop()
}

func TestMemoryUsageTracking(t *testing.T) {
	config := DefaultPerformanceConfig()
	mm := NewMemoryMonitor(config)

	// Test initial usage values
	currentUsage := mm.GetCurrentUsage()
	if currentUsage <= 0 {
		t.Error("Expected positive current usage")
	}

	peakUsage := mm.GetPeakUsage()
	if peakUsage <= 0 {
		t.Error("Expected positive peak usage")
	}

	averageUsage := mm.GetAverageUsage()
	if averageUsage <= 0 {
		t.Error("Expected positive average usage")
	}

	// Test usage percentage calculation
	usagePercent := mm.GetUsagePercent()
	if usagePercent <= 0 || usagePercent > 100 {
		t.Errorf("Expected usage percentage between 0-100, got %f", usagePercent)
	}
}

func TestMemoryPressureDetection(t *testing.T) {
	config := DefaultPerformanceConfig()
	// Set a very low memory limit to easily test pressure detection
	config.MemoryLimitBytes = 100 * 1024 * 1024 // 100MB
	config.GCThreshold = 0.05                   // 5% threshold for testing

	mm := NewMemoryMonitor(config)

	// Initially should not be under pressure (unless system is really low on memory)
	initialPressure := mm.IsUnderPressure()

	// Test pressure level calculation with different usage levels
	testCases := []struct {
		usage    int64
		expected MemoryPressureLevel
	}{
		{10 * 1024 * 1024, PressureLow},      // 10MB (10%)
		{60 * 1024 * 1024, PressureModerate}, // 60MB (60%)
		{75 * 1024 * 1024, PressureHigh},     // 75MB (75%) - within threshold
		{95 * 1024 * 1024, PressureCritical}, // 95MB (95%)
	}

	for _, tc := range testCases {
		level := mm.calculatePressureLevel(tc.usage)
		if level != tc.expected {
			t.Errorf("Usage %d bytes: expected pressure level %s, got %s",
				tc.usage, tc.expected.String(), level.String())
		}
	}

	t.Logf("Initial pressure state: %v", initialPressure)
}

func TestMemoryStats(t *testing.T) {
	config := DefaultPerformanceConfig()
	mm := NewMemoryMonitor(config)

	stats := mm.GetMemoryStats()

	// Validate all stats fields
	if stats.CurrentUsage <= 0 {
		t.Error("Expected positive current usage in stats")
	}

	if stats.PeakUsage <= 0 {
		t.Error("Expected positive peak usage in stats")
	}

	if stats.AverageUsage <= 0 {
		t.Error("Expected positive average usage in stats")
	}

	if stats.BaselineUsage <= 0 {
		t.Error("Expected positive baseline usage in stats")
	}

	if stats.LimitBytes != config.MemoryLimitBytes {
		t.Errorf("Expected limit bytes to be %d, got %d", config.MemoryLimitBytes, stats.LimitBytes)
	}

	if stats.UsagePercent <= 0 || stats.UsagePercent > 100 {
		t.Errorf("Expected usage percentage between 0-100, got %f", stats.UsagePercent)
	}

	// GC triggers and pressure events should be non-negative
	if stats.GCTriggers < 0 {
		t.Errorf("Expected non-negative GC triggers, got %d", stats.GCTriggers)
	}

	if stats.PressureEvents < 0 {
		t.Errorf("Expected non-negative pressure events, got %d", stats.PressureEvents)
	}
}

func TestMemoryTrendAnalysis(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MemoryCheckInterval = 10 * time.Millisecond
	mm := NewMemoryMonitor(config)

	// Start monitoring to accumulate samples
	err := mm.Start()
	if err != nil {
		t.Fatalf("Failed to start memory monitor: %v", err)
	}
	defer mm.Stop()

	// Allow some samples to accumulate
	time.Sleep(150 * time.Millisecond)

	trend := mm.GetMemoryTrend()

	// Validate trend structure
	validDirections := map[string]bool{
		"increasing": true,
		"decreasing": true,
		"stable":     true,
	}

	if !validDirections[trend.Direction] {
		t.Errorf("Invalid trend direction: %s", trend.Direction)
	}

	if trend.Confidence < 0 || trend.Confidence > 1 {
		t.Errorf("Expected confidence between 0-1, got %f", trend.Confidence)
	}

	t.Logf("Memory trend: %s (confidence: %.2f, rate: %.2f bytes/sec)",
		trend.Direction, trend.Confidence, trend.Rate)
}

func TestMemorySampleCollection(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MemoryCheckInterval = 20 * time.Millisecond
	mm := NewMemoryMonitor(config)

	// Start monitoring
	err := mm.Start()
	if err != nil {
		t.Fatalf("Failed to start memory monitor: %v", err)
	}
	defer mm.Stop()

	// Allow samples to be collected
	time.Sleep(100 * time.Millisecond)

	// Get recent samples
	samples := mm.GetLatestSamples(5)

	if len(samples) == 0 {
		t.Error("Expected at least some memory samples")
	}

	// Validate sample structure
	for i, sample := range samples {
		if sample.AllocBytes <= 0 {
			t.Errorf("Sample %d: expected positive alloc bytes, got %d", i, sample.AllocBytes)
		}

		if sample.TotalAlloc <= 0 {
			t.Errorf("Sample %d: expected positive total alloc, got %d", i, sample.TotalAlloc)
		}

		if sample.SysBytes <= 0 {
			t.Errorf("Sample %d: expected positive sys bytes, got %d", i, sample.SysBytes)
		}

		if sample.Timestamp.IsZero() {
			t.Errorf("Sample %d: expected valid timestamp", i)
		}
	}
}

func TestMemoryGCTrigger(t *testing.T) {
	config := DefaultPerformanceConfig()
	mm := NewMemoryMonitor(config)

	initialTriggers := mm.GetGCTriggers()

	// Manually trigger GC
	mm.TriggerGC()

	afterTriggers := mm.GetGCTriggers()

	if afterTriggers != initialTriggers+1 {
		t.Errorf("Expected GC triggers to increase by 1, got %d -> %d", initialTriggers, afterTriggers)
	}

	// Test multiple triggers
	mm.TriggerGC()
	mm.TriggerGC()

	finalTriggers := mm.GetGCTriggers()
	if finalTriggers != initialTriggers+3 {
		t.Errorf("Expected GC triggers to increase by 3 total, got %d -> %d", initialTriggers, finalTriggers)
	}
}

func TestMemoryPressureLevelString(t *testing.T) {
	testCases := []struct {
		level    MemoryPressureLevel
		expected string
	}{
		{PressureLow, "low"},
		{PressureModerate, "moderate"},
		{PressureHigh, "high"},
		{PressureCritical, "critical"},
		{MemoryPressureLevel(999), "unknown"},
	}

	for _, tc := range testCases {
		result := tc.level.String()
		if result != tc.expected {
			t.Errorf("Level %d: expected %s, got %s", tc.level, tc.expected, result)
		}
	}
}

func TestMemoryMonitorConcurrentAccess(t *testing.T) {
	config := DefaultPerformanceConfig()
	config.MemoryCheckInterval = 5 * time.Millisecond
	mm := NewMemoryMonitor(config)

	err := mm.Start()
	if err != nil {
		t.Fatalf("Failed to start memory monitor: %v", err)
	}
	defer mm.Stop()

	// Run concurrent operations to test thread safety
	done := make(chan struct{}, 10)

	// Concurrent readers
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 10; j++ {
				_ = mm.GetCurrentUsage()
				_ = mm.GetPeakUsage()
				_ = mm.GetMemoryStats()
				_ = mm.IsUnderPressure()
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// Concurrent GC triggers
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 5; j++ {
				mm.TriggerGC()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}

	// Verify monitor still works after concurrent access
	finalStats := mm.GetMemoryStats()
	if finalStats.CurrentUsage <= 0 {
		t.Error("Memory monitor appears to be corrupted after concurrent access")
	}
}

func BenchmarkMemoryMonitor(b *testing.B) {
	config := DefaultPerformanceConfig()
	mm := NewMemoryMonitor(config)

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("GetCurrentUsage", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mm.GetCurrentUsage()
		}
	})

	b.Run("GetMemoryStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mm.GetMemoryStats()
		}
	})

	b.Run("IsUnderPressure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mm.IsUnderPressure()
		}
	})

	b.Run("TriggerGC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			mm.TriggerGC()
		}
	})
}
