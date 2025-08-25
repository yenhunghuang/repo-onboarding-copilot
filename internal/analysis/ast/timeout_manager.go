package ast

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// TimeoutManager handles parsing timeouts and provides timeout statistics
type TimeoutManager struct {
	config PerformanceConfig

	// Active timers tracking
	activeTimers map[string]*TimeoutContext
	timerMutex   sync.RWMutex

	// Statistics
	totalTimeouts int64
	fileTimeouts  int64
	batchTimeouts int64
	timeoutEvents int64

	// Timeout categorization
	shortTimeouts  int64 // < 5 seconds
	mediumTimeouts int64 // 5-30 seconds
	longTimeouts   int64 // > 30 seconds

	// Performance tracking
	averageParseTime  time.Duration
	longestParseTime  time.Duration
	shortestParseTime time.Duration
	timeoutHistogram  map[time.Duration]int64

	// State
	startTime time.Time
	mu        sync.RWMutex
}

// TimeoutContext represents a timeout-managed operation
type TimeoutContext struct {
	ID          string
	FilePath    string
	StartTime   time.Time
	Timeout     time.Duration
	Context     context.Context
	CancelFunc  context.CancelFunc
	TimedOut    bool
	CompletedAt *time.Time
}

// TimeoutStats provides comprehensive timeout statistics
type TimeoutStats struct {
	TotalOperations int64   `json:"total_operations"`
	TotalTimeouts   int64   `json:"total_timeouts"`
	FileTimeouts    int64   `json:"file_timeouts"`
	BatchTimeouts   int64   `json:"batch_timeouts"`
	TimeoutRate     float64 `json:"timeout_rate"`

	// Timeout distribution
	ShortTimeouts  int64 `json:"short_timeouts"`  // < 5s
	MediumTimeouts int64 `json:"medium_timeouts"` // 5-30s
	LongTimeouts   int64 `json:"long_timeouts"`   // > 30s

	// Performance metrics
	AverageParseTime  time.Duration `json:"average_parse_time"`
	LongestParseTime  time.Duration `json:"longest_parse_time"`
	ShortestParseTime time.Duration `json:"shortest_parse_time"`

	// Current state
	ActiveOperations int           `json:"active_operations"`
	LongestRunning   time.Duration `json:"longest_running"`
}

// TimeoutEvent represents a timeout occurrence
type TimeoutEvent struct {
	ID          string        `json:"id"`
	FilePath    string        `json:"file_path"`
	StartTime   time.Time     `json:"start_time"`
	TimeoutTime time.Time     `json:"timeout_time"`
	Duration    time.Duration `json:"duration"`
	TimeoutType string        `json:"timeout_type"` // "file", "batch", "total"
	Context     string        `json:"context"`
}

// NewTimeoutManager creates a new timeout manager
func NewTimeoutManager(config PerformanceConfig) *TimeoutManager {
	return &TimeoutManager{
		config:            config,
		activeTimers:      make(map[string]*TimeoutContext),
		timeoutHistogram:  make(map[time.Duration]int64),
		startTime:         time.Now(),
		shortestParseTime: time.Hour, // Initialize to high value
	}
}

// CreateFileTimeout creates a timeout context for file parsing
func (tm *TimeoutManager) CreateFileTimeout(filePath string, baseCtx context.Context) (context.Context, context.CancelFunc, string) {
	id := tm.generateTimeoutID(filePath)

	ctx, cancel := context.WithTimeout(baseCtx, tm.config.ParseTimeout)

	timeoutCtx := &TimeoutContext{
		ID:         id,
		FilePath:   filePath,
		StartTime:  time.Now(),
		Timeout:    tm.config.ParseTimeout,
		Context:    ctx,
		CancelFunc: cancel,
		TimedOut:   false,
	}

	tm.timerMutex.Lock()
	tm.activeTimers[id] = timeoutCtx
	tm.timerMutex.Unlock()

	// Monitor timeout in background
	go tm.monitorTimeout(id)

	return ctx, func() {
		tm.completeOperation(id, false)
		cancel()
	}, id
}

// CreateBatchTimeout creates a timeout context for batch processing
func (tm *TimeoutManager) CreateBatchTimeout(batchID int, baseCtx context.Context) (context.Context, context.CancelFunc, string) {
	id := tm.generateBatchTimeoutID(batchID)

	// Batch timeout is longer than individual file timeout
	batchTimeout := tm.config.ParseTimeout * 10
	ctx, cancel := context.WithTimeout(baseCtx, batchTimeout)

	timeoutCtx := &TimeoutContext{
		ID:         id,
		FilePath:   "",
		StartTime:  time.Now(),
		Timeout:    batchTimeout,
		Context:    ctx,
		CancelFunc: cancel,
		TimedOut:   false,
	}

	tm.timerMutex.Lock()
	tm.activeTimers[id] = timeoutCtx
	tm.timerMutex.Unlock()

	// Monitor timeout in background
	go tm.monitorTimeout(id)

	return ctx, func() {
		tm.completeOperation(id, false)
		cancel()
	}, id
}

// generateTimeoutID creates a unique ID for timeout tracking
func (tm *TimeoutManager) generateTimeoutID(filePath string) string {
	return "file_" + filePath + "_" + time.Now().Format("20060102_150405.000000")
}

// generateBatchTimeoutID creates a unique ID for batch timeout tracking
func (tm *TimeoutManager) generateBatchTimeoutID(batchID int) string {
	return "batch_" + string(rune(batchID)) + "_" + time.Now().Format("20060102_150405.000000")
}

// monitorTimeout monitors a timeout context and handles timeout events
func (tm *TimeoutManager) monitorTimeout(id string) {
	tm.timerMutex.RLock()
	timeoutCtx, exists := tm.activeTimers[id]
	tm.timerMutex.RUnlock()

	if !exists {
		return
	}

	// Wait for context completion or timeout
	<-timeoutCtx.Context.Done()

	// Check if it was a timeout
	if timeoutCtx.Context.Err() == context.DeadlineExceeded {
		tm.handleTimeout(id, timeoutCtx)
	}
}

// handleTimeout processes a timeout event
func (tm *TimeoutManager) handleTimeout(id string, timeoutCtx *TimeoutContext) {
	now := time.Now()
	duration := now.Sub(timeoutCtx.StartTime)

	// Mark as timed out
	timeoutCtx.TimedOut = true
	completedAt := now
	timeoutCtx.CompletedAt = &completedAt

	// Update statistics
	atomic.AddInt64(&tm.totalTimeouts, 1)

	// Categorize timeout
	if timeoutCtx.FilePath != "" {
		atomic.AddInt64(&tm.fileTimeouts, 1)
	} else {
		atomic.AddInt64(&tm.batchTimeouts, 1)
	}

	// Categorize by duration
	tm.categorizeTimeout(duration)

	// Update histogram
	tm.updateTimeoutHistogram(duration)

	// Create timeout event (for logging/monitoring)
	event := TimeoutEvent{
		ID:          id,
		FilePath:    timeoutCtx.FilePath,
		StartTime:   timeoutCtx.StartTime,
		TimeoutTime: now,
		Duration:    duration,
		TimeoutType: tm.getTimeoutType(timeoutCtx),
		Context:     "AST parsing operation exceeded timeout limit",
	}

	// Log timeout event (in production, this would go to audit logging)
	tm.logTimeoutEvent(event)
}

// completeOperation marks an operation as completed successfully
func (tm *TimeoutManager) completeOperation(id string, timedOut bool) {
	tm.timerMutex.Lock()
	defer tm.timerMutex.Unlock()

	timeoutCtx, exists := tm.activeTimers[id]
	if !exists {
		return
	}

	now := time.Now()
	duration := now.Sub(timeoutCtx.StartTime)

	if !timedOut {
		timeoutCtx.CompletedAt = &now

		// Update performance statistics
		tm.updatePerformanceStats(duration)
	}

	// Remove from active timers
	delete(tm.activeTimers, id)
}

// categorizeTimeout categorizes timeout by duration
func (tm *TimeoutManager) categorizeTimeout(duration time.Duration) {
	switch {
	case duration < 5*time.Second:
		atomic.AddInt64(&tm.shortTimeouts, 1)
	case duration < 30*time.Second:
		atomic.AddInt64(&tm.mediumTimeouts, 1)
	default:
		atomic.AddInt64(&tm.longTimeouts, 1)
	}
}

// updateTimeoutHistogram updates the timeout duration histogram
func (tm *TimeoutManager) updateTimeoutHistogram(duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Round to nearest second for histogram, ensure minimum 1 second
	rounded := duration.Round(time.Second)
	if rounded <= 0 {
		rounded = 1 * time.Second
	}
	tm.timeoutHistogram[rounded]++
}

// updatePerformanceStats updates parsing performance statistics
func (tm *TimeoutManager) updatePerformanceStats(duration time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Update average parse time
	if tm.averageParseTime == 0 {
		tm.averageParseTime = duration
	} else {
		tm.averageParseTime = (tm.averageParseTime + duration) / 2
	}

	// Update longest parse time
	if duration > tm.longestParseTime {
		tm.longestParseTime = duration
	}

	// Update shortest parse time
	if duration < tm.shortestParseTime {
		tm.shortestParseTime = duration
	}
}

// getTimeoutType determines the type of timeout based on context
func (tm *TimeoutManager) getTimeoutType(timeoutCtx *TimeoutContext) string {
	if timeoutCtx.FilePath != "" {
		return "file"
	}
	return "batch"
}

// logTimeoutEvent logs a timeout event (placeholder for audit logging integration)
func (tm *TimeoutManager) logTimeoutEvent(event TimeoutEvent) {
	// In production, this would integrate with pkg/logger/audit_logger.go
	// For now, we just track the event internally
}

// GetActiveOperations returns count of currently active operations
func (tm *TimeoutManager) GetActiveOperations() int {
	tm.timerMutex.RLock()
	defer tm.timerMutex.RUnlock()

	return len(tm.activeTimers)
}

// GetLongestRunning returns duration of longest currently running operation
func (tm *TimeoutManager) GetLongestRunning() time.Duration {
	tm.timerMutex.RLock()
	defer tm.timerMutex.RUnlock()

	var longest time.Duration
	now := time.Now()

	for _, timeoutCtx := range tm.activeTimers {
		duration := now.Sub(timeoutCtx.StartTime)
		if duration > longest {
			longest = duration
		}
	}

	return longest
}

// GetTimeoutStats returns comprehensive timeout statistics
func (tm *TimeoutManager) GetTimeoutStats() TimeoutStats {
	totalOps := atomic.LoadInt64(&tm.totalTimeouts) + int64(tm.GetActiveOperations())
	timeouts := atomic.LoadInt64(&tm.totalTimeouts)

	var timeoutRate float64
	if totalOps > 0 {
		timeoutRate = float64(timeouts) / float64(totalOps)
	}

	tm.mu.RLock()
	avgParseTime := tm.averageParseTime
	longestParseTime := tm.longestParseTime
	shortestParseTime := tm.shortestParseTime
	tm.mu.RUnlock()

	return TimeoutStats{
		TotalOperations:   totalOps,
		TotalTimeouts:     timeouts,
		FileTimeouts:      atomic.LoadInt64(&tm.fileTimeouts),
		BatchTimeouts:     atomic.LoadInt64(&tm.batchTimeouts),
		TimeoutRate:       timeoutRate,
		ShortTimeouts:     atomic.LoadInt64(&tm.shortTimeouts),
		MediumTimeouts:    atomic.LoadInt64(&tm.mediumTimeouts),
		LongTimeouts:      atomic.LoadInt64(&tm.longTimeouts),
		AverageParseTime:  avgParseTime,
		LongestParseTime:  longestParseTime,
		ShortestParseTime: shortestParseTime,
		ActiveOperations:  tm.GetActiveOperations(),
		LongestRunning:    tm.GetLongestRunning(),
	}
}

// GetTimeoutHistogram returns timeout duration histogram
func (tm *TimeoutManager) GetTimeoutHistogram() map[time.Duration]int64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	histogram := make(map[time.Duration]int64)
	for duration, count := range tm.timeoutHistogram {
		histogram[duration] = count
	}

	return histogram
}

// GetActiveTimers returns information about currently active timers
func (tm *TimeoutManager) GetActiveTimers() []ActiveTimer {
	tm.timerMutex.RLock()
	defer tm.timerMutex.RUnlock()

	timers := make([]ActiveTimer, 0, len(tm.activeTimers))
	now := time.Now()

	for id, timeoutCtx := range tm.activeTimers {
		timer := ActiveTimer{
			ID:                id,
			FilePath:          timeoutCtx.FilePath,
			StartTime:         timeoutCtx.StartTime,
			ConfiguredTimeout: timeoutCtx.Timeout,
			ElapsedTime:       now.Sub(timeoutCtx.StartTime),
			RemainingTime:     timeoutCtx.Timeout - now.Sub(timeoutCtx.StartTime),
		}

		if timer.RemainingTime < 0 {
			timer.RemainingTime = 0
		}

		timers = append(timers, timer)
	}

	return timers
}

// ActiveTimer represents an active timeout timer
type ActiveTimer struct {
	ID                string        `json:"id"`
	FilePath          string        `json:"file_path"`
	StartTime         time.Time     `json:"start_time"`
	ConfiguredTimeout time.Duration `json:"configured_timeout"`
	ElapsedTime       time.Duration `json:"elapsed_time"`
	RemainingTime     time.Duration `json:"remaining_time"`
}

// GetTimeoutEvents returns the total number of timeout events
func (tm *TimeoutManager) GetTimeoutEvents() int64 {
	return atomic.LoadInt64(&tm.timeoutEvents)
}

// IsHealthy returns true if timeout manager is operating within expected parameters
func (tm *TimeoutManager) IsHealthy() bool {
	stats := tm.GetTimeoutStats()

	// Consider healthy if timeout rate is under 10%
	return stats.TimeoutRate < 0.10
}

// Stop gracefully stops the timeout manager
func (tm *TimeoutManager) Stop() {
	tm.timerMutex.Lock()
	defer tm.timerMutex.Unlock()

	// Cancel all active timers
	for id, timeoutCtx := range tm.activeTimers {
		if timeoutCtx.CancelFunc != nil {
			timeoutCtx.CancelFunc()
		}
		delete(tm.activeTimers, id)
	}
}
