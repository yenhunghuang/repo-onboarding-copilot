package ast

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryMonitor provides memory usage tracking and management for large repository parsing
type MemoryMonitor struct {
	config PerformanceConfig

	// Current state
	currentUsage  int64 // Current memory usage in bytes
	peakUsage     int64 // Peak memory usage observed
	baselineUsage int64 // Memory usage at startup

	// Historical data
	samples     []MemorySample // Recent memory usage samples
	sampleIndex int            // Circular buffer index
	maxSamples  int            // Maximum samples to retain

	// Statistics
	gcTriggers     int64 // Number of manual GC triggers
	pressureEvents int64 // Memory pressure events
	averageUsage   int64 // Running average memory usage

	// Control
	monitoring int32         // Atomic flag for monitoring state
	stopChan   chan struct{} // Channel to stop monitoring
	ticker     *time.Ticker  // Ticker for periodic checks

	// Synchronization
	mu sync.RWMutex // Protects samples and calculations
}

// MemorySample represents a point-in-time memory measurement
type MemorySample struct {
	Timestamp  time.Time `json:"timestamp"`
	AllocBytes int64     `json:"alloc_bytes"` // Currently allocated bytes
	TotalAlloc int64     `json:"total_alloc"` // Total allocated bytes (cumulative)
	SysBytes   int64     `json:"sys_bytes"`   // System memory obtained from OS
	GCCycles   uint32    `json:"gc_cycles"`   // Number of completed GC cycles
	NextGC     int64     `json:"next_gc"`     // Target heap size for next GC
	LastGC     int64     `json:"last_gc"`     // Time of last GC (nanoseconds since Unix epoch)
}

// MemoryPressureLevel represents different levels of memory pressure
type MemoryPressureLevel int

const (
	PressureLow MemoryPressureLevel = iota
	PressureModerate
	PressureHigh
	PressureCritical
)

// String returns string representation of memory pressure level
func (mpl MemoryPressureLevel) String() string {
	switch mpl {
	case PressureLow:
		return "low"
	case PressureModerate:
		return "moderate"
	case PressureHigh:
		return "high"
	case PressureCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// NewMemoryMonitor creates a new memory monitor with specified configuration
func NewMemoryMonitor(config PerformanceConfig) *MemoryMonitor {
	mm := &MemoryMonitor{
		config:     config,
		maxSamples: 100, // Keep last 100 samples for trend analysis
		samples:    make([]MemorySample, 100),
		stopChan:   make(chan struct{}),
	}

	// Record baseline memory usage
	mm.recordBaseline()

	return mm
}

// Start begins memory monitoring in a separate goroutine
func (mm *MemoryMonitor) Start() error {
	if !atomic.CompareAndSwapInt32(&mm.monitoring, 0, 1) {
		return nil // Already monitoring
	}

	mm.ticker = time.NewTicker(mm.config.MemoryCheckInterval)

	go mm.monitorLoop()

	return nil
}

// Stop stops memory monitoring gracefully
func (mm *MemoryMonitor) Stop() {
	if !atomic.CompareAndSwapInt32(&mm.monitoring, 1, 0) {
		return // Already stopped
	}

	if mm.ticker != nil {
		mm.ticker.Stop()
	}

	// Close channel safely
	select {
	case <-mm.stopChan:
		// Already closed
	default:
		close(mm.stopChan)
	}
}

// recordBaseline captures initial memory state for comparison
func (mm *MemoryMonitor) recordBaseline() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mm.baselineUsage = int64(memStats.Alloc)
	mm.currentUsage = mm.baselineUsage
	mm.peakUsage = mm.baselineUsage
	mm.averageUsage = mm.baselineUsage
}

// monitorLoop performs periodic memory monitoring
func (mm *MemoryMonitor) monitorLoop() {
	for {
		select {
		case <-mm.stopChan:
			return
		case <-mm.ticker.C:
			mm.checkMemoryUsage()
		}
	}
}

// checkMemoryUsage performs a memory usage check and takes action if needed
func (mm *MemoryMonitor) checkMemoryUsage() {
	sample := mm.takeSample()

	mm.mu.Lock()
	// Store sample in circular buffer
	mm.samples[mm.sampleIndex] = sample
	mm.sampleIndex = (mm.sampleIndex + 1) % mm.maxSamples

	// Update current and peak usage
	atomic.StoreInt64(&mm.currentUsage, sample.AllocBytes)

	if sample.AllocBytes > atomic.LoadInt64(&mm.peakUsage) {
		atomic.StoreInt64(&mm.peakUsage, sample.AllocBytes)
	}

	// Update running average
	mm.updateAverageUsage(sample.AllocBytes)
	mm.mu.Unlock()

	// Check for memory pressure and take action
	pressureLevel := mm.calculatePressureLevel(sample.AllocBytes)
	mm.handleMemoryPressure(pressureLevel, sample)
}

// takeSample captures current memory statistics
func (mm *MemoryMonitor) takeSample() MemorySample {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return MemorySample{
		Timestamp:  time.Now(),
		AllocBytes: int64(memStats.Alloc),
		TotalAlloc: int64(memStats.TotalAlloc),
		SysBytes:   int64(memStats.Sys),
		GCCycles:   memStats.NumGC,
		NextGC:     int64(memStats.NextGC),
		LastGC:     int64(memStats.LastGC),
	}
}

// updateAverageUsage maintains a running average of memory usage
func (mm *MemoryMonitor) updateAverageUsage(currentUsage int64) {
	// Simple exponential moving average with alpha = 0.1
	oldAverage := atomic.LoadInt64(&mm.averageUsage)
	newAverage := int64(0.9*float64(oldAverage) + 0.1*float64(currentUsage))
	atomic.StoreInt64(&mm.averageUsage, newAverage)
}

// calculatePressureLevel determines current memory pressure based on usage
func (mm *MemoryMonitor) calculatePressureLevel(currentUsage int64) MemoryPressureLevel {
	usagePercent := float64(currentUsage) / float64(mm.config.MemoryLimitBytes)

	switch {
	case usagePercent < 0.5:
		return PressureLow
	case usagePercent < 0.7:
		return PressureModerate
	case usagePercent < 0.85:
		return PressureHigh
	default:
		return PressureCritical
	}
}

// handleMemoryPressure takes appropriate action based on memory pressure level
func (mm *MemoryMonitor) handleMemoryPressure(level MemoryPressureLevel, sample MemorySample) {
	switch level {
	case PressureHigh:
		// Trigger garbage collection proactively
		mm.TriggerGC()
		atomic.AddInt64(&mm.pressureEvents, 1)

	case PressureCritical:
		// Aggressive memory management
		mm.TriggerGC()
		runtime.GC() // Force a second GC cycle
		atomic.AddInt64(&mm.pressureEvents, 1)

		// Log critical memory pressure (in production, this would log to audit system)
		// For now, we track the event count
	}
}

// TriggerGC forces garbage collection and updates statistics
func (mm *MemoryMonitor) TriggerGC() {
	runtime.GC()
	atomic.AddInt64(&mm.gcTriggers, 1)
}

// IsUnderPressure returns true if system is under significant memory pressure
func (mm *MemoryMonitor) IsUnderPressure() bool {
	currentUsage := atomic.LoadInt64(&mm.currentUsage)
	usagePercent := float64(currentUsage) / float64(mm.config.MemoryLimitBytes)
	return usagePercent >= mm.config.GCThreshold
}

// GetCurrentUsage returns current memory usage in bytes
func (mm *MemoryMonitor) GetCurrentUsage() int64 {
	return atomic.LoadInt64(&mm.currentUsage)
}

// GetPeakUsage returns peak memory usage observed during monitoring
func (mm *MemoryMonitor) GetPeakUsage() int64 {
	return atomic.LoadInt64(&mm.peakUsage)
}

// GetAverageUsage returns average memory usage over monitoring period
func (mm *MemoryMonitor) GetAverageUsage() int64 {
	return atomic.LoadInt64(&mm.averageUsage)
}

// GetUsagePercent returns current usage as percentage of limit
func (mm *MemoryMonitor) GetUsagePercent() float64 {
	currentUsage := atomic.LoadInt64(&mm.currentUsage)
	return float64(currentUsage) / float64(mm.config.MemoryLimitBytes) * 100.0
}

// GetGCTriggers returns number of manual GC triggers
func (mm *MemoryMonitor) GetGCTriggers() int64 {
	return atomic.LoadInt64(&mm.gcTriggers)
}

// GetPressureEvents returns number of memory pressure events
func (mm *MemoryMonitor) GetPressureEvents() int64 {
	return atomic.LoadInt64(&mm.pressureEvents)
}

// GetMemoryTrend analyzes recent memory usage trend
func (mm *MemoryMonitor) GetMemoryTrend() MemoryTrend {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	if mm.sampleIndex < 10 {
		return MemoryTrend{Direction: "stable", Confidence: 0.0}
	}

	// Analyze last 10 samples for trend
	recent := make([]int64, 10)
	for i := 0; i < 10; i++ {
		idx := (mm.sampleIndex - 10 + i + mm.maxSamples) % mm.maxSamples
		recent[i] = mm.samples[idx].AllocBytes
	}

	return mm.calculateTrend(recent)
}

// MemoryTrend represents memory usage trend analysis
type MemoryTrend struct {
	Direction  string  `json:"direction"`  // "increasing", "decreasing", "stable"
	Rate       float64 `json:"rate"`       // Bytes per second change rate
	Confidence float64 `json:"confidence"` // Confidence level (0.0-1.0)
}

// calculateTrend performs linear regression on memory samples to determine trend
func (mm *MemoryMonitor) calculateTrend(samples []int64) MemoryTrend {
	if len(samples) < 2 {
		return MemoryTrend{Direction: "stable", Confidence: 0.0}
	}

	n := float64(len(samples))

	// Calculate linear regression slope
	var sumX, sumY, sumXY, sumXX float64
	for i, sample := range samples {
		x := float64(i)
		y := float64(sample)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Linear regression: y = mx + b, we want slope m
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return MemoryTrend{Direction: "stable", Confidence: 0.0}
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	// Convert slope to bytes per second (assuming samples are taken at MemoryCheckInterval)
	slopePerSecond := slope / mm.config.MemoryCheckInterval.Seconds()

	// Determine direction and confidence
	var direction string
	absSlope := slope
	if absSlope < 0 {
		absSlope = -absSlope
	}

	// Confidence based on slope magnitude relative to average usage
	averageUsage := sumY / n
	relativeSlope := absSlope / averageUsage
	confidence := relativeSlope * 10.0 // Scale factor
	if confidence > 1.0 {
		confidence = 1.0
	}

	if slope > averageUsage*0.01 { // Increasing if slope > 1% of average
		direction = "increasing"
	} else if slope < -averageUsage*0.01 { // Decreasing if slope < -1% of average
		direction = "decreasing"
	} else {
		direction = "stable"
		confidence = 1.0 - confidence // Invert confidence for stable trend
	}

	return MemoryTrend{
		Direction:  direction,
		Rate:       slopePerSecond,
		Confidence: confidence,
	}
}

// GetLatestSamples returns the most recent memory samples for analysis
func (mm *MemoryMonitor) GetLatestSamples(count int) []MemorySample {
	if count <= 0 || count > mm.maxSamples {
		count = mm.maxSamples
	}

	mm.mu.RLock()
	defer mm.mu.RUnlock()

	var samples []MemorySample
	for i := 0; i < count; i++ {
		idx := (mm.sampleIndex - count + i + mm.maxSamples) % mm.maxSamples
		sample := mm.samples[idx]
		
		// Only include initialized samples (non-zero timestamp)
		if !sample.Timestamp.IsZero() {
			samples = append(samples, sample)
		}
	}

	return samples
}

// GetMemoryStats returns comprehensive memory statistics
func (mm *MemoryMonitor) GetMemoryStats() MemoryStats {
	return MemoryStats{
		CurrentUsage:    mm.GetCurrentUsage(),
		PeakUsage:       mm.GetPeakUsage(),
		AverageUsage:    mm.GetAverageUsage(),
		BaselineUsage:   mm.baselineUsage,
		LimitBytes:      mm.config.MemoryLimitBytes,
		UsagePercent:    mm.GetUsagePercent(),
		GCTriggers:      mm.GetGCTriggers(),
		PressureEvents:  mm.GetPressureEvents(),
		Trend:           mm.GetMemoryTrend(),
		IsUnderPressure: mm.IsUnderPressure(),
	}
}

// MemoryStats provides comprehensive memory usage statistics
type MemoryStats struct {
	CurrentUsage    int64       `json:"current_usage"`
	PeakUsage       int64       `json:"peak_usage"`
	AverageUsage    int64       `json:"average_usage"`
	BaselineUsage   int64       `json:"baseline_usage"`
	LimitBytes      int64       `json:"limit_bytes"`
	UsagePercent    float64     `json:"usage_percent"`
	GCTriggers      int64       `json:"gc_triggers"`
	PressureEvents  int64       `json:"pressure_events"`
	Trend           MemoryTrend `json:"trend"`
	IsUnderPressure bool        `json:"is_under_pressure"`
}
