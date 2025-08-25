package ast

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceConfig defines performance optimization settings
type PerformanceConfig struct {
	// Worker pool settings
	MaxWorkers     int `json:"max_workers"`      // Maximum number of worker goroutines
	WorkerPoolSize int `json:"worker_pool_size"` // Channel buffer size for work distribution

	// Memory management
	MemoryLimitBytes    int64         `json:"memory_limit_bytes"`    // Hard memory limit (2GB default)
	MemoryCheckInterval time.Duration `json:"memory_check_interval"` // How often to check memory usage
	GCThreshold         float64       `json:"gc_threshold"`          // Memory percentage to trigger GC

	// File batching
	BatchSize       int `json:"batch_size"`       // Files per batch
	MaxBatches      int `json:"max_batches"`      // Maximum concurrent batches
	StreamThreshold int `json:"stream_threshold"` // File count threshold for streaming mode

	// Timeouts and limits
	ParseTimeout time.Duration `json:"parse_timeout"` // Per-file parsing timeout
	TotalTimeout time.Duration `json:"total_timeout"` // Overall analysis timeout

	// Progress tracking
	ProgressInterval time.Duration `json:"progress_interval"` // Progress reporting interval
	EnableMetrics    bool          `json:"enable_metrics"`    // Enable performance metrics collection
}

// DefaultPerformanceConfig returns performance configuration aligned with architecture requirements
func DefaultPerformanceConfig() PerformanceConfig {
	return PerformanceConfig{
		MaxWorkers:          runtime.GOMAXPROCS(0), // Use available CPU cores
		WorkerPoolSize:      1000,                  // Large buffer for work distribution
		MemoryLimitBytes:    2 << 30,               // 2GB hard limit per architecture
		MemoryCheckInterval: 1 * time.Second,       // Check memory every second
		GCThreshold:         0.8,                   // Trigger GC at 80% of limit
		BatchSize:           100,                   // Process 100 files per batch
		MaxBatches:          10,                    // Maximum 10 concurrent batches
		StreamThreshold:     1000,                  // Use streaming for >1000 files
		ParseTimeout:        30 * time.Second,      // 30 second per-file timeout
		TotalTimeout:        1 * time.Hour,         // 1 hour total limit per architecture
		ProgressInterval:    5 * time.Second,       // Report progress every 5 seconds
		EnableMetrics:       true,
	}
}

// PerformanceOptimizer manages high-performance parsing for large repositories
type PerformanceOptimizer struct {
	config         PerformanceConfig
	workerPool     chan WorkItem
	resultChan     chan ParseResult
	progressChan   chan ProgressUpdate
	metrics        *PerformanceMetrics
	memoryMonitor  *MemoryMonitor
	timeoutManager *TimeoutManager

	// State management
	totalFiles     int64
	processedFiles int64
	failedFiles    int64
	batchesActive  int64

	// Synchronization
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	// Runtime state
	startTime        time.Time
	lastProgressTime time.Time
}

// WorkItem represents a file parsing work unit
type WorkItem struct {
	FilePath  string
	Content   []byte
	BatchID   int
	Timestamp time.Time
	Priority  int // Priority level for processing order
	Metadata  map[string]interface{}
}

// ParseResult extends the existing ParseResult with performance metadata
type ParseResultWithPerf struct {
	*ParseResult
	ProcessingTime time.Duration `json:"processing_time"`
	MemoryUsage    int64         `json:"memory_usage"`
	WorkerID       int           `json:"worker_id"`
	BatchID        int           `json:"batch_id"`
	RetryCount     int           `json:"retry_count"`
}

// ProgressUpdate represents parsing progress information
type ProgressUpdate struct {
	TotalFiles         int64         `json:"total_files"`
	ProcessedFiles     int64         `json:"processed_files"`
	FailedFiles        int64         `json:"failed_files"`
	CurrentBatch       int           `json:"current_batch"`
	ActiveWorkers      int           `json:"active_workers"`
	MemoryUsage        int64         `json:"memory_usage"`
	EstimatedRemaining time.Duration `json:"estimated_remaining"`
	ThroughputFPS      float64       `json:"throughput_fps"` // Files per second
	Timestamp          time.Time     `json:"timestamp"`
}

// PerformanceMetrics tracks parsing performance statistics
type PerformanceMetrics struct {
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	TotalDuration time.Duration `json:"total_duration"`

	// File processing stats
	TotalFilesProcessed int64 `json:"total_files_processed"`
	SuccessfulFiles     int64 `json:"successful_files"`
	FailedFiles         int64 `json:"failed_files"`
	PartialFiles        int64 `json:"partial_files"`

	// Performance metrics
	AverageProcessingTime time.Duration `json:"avg_processing_time"`
	ThroughputFPS         float64       `json:"throughput_fps"`
	PeakMemoryUsage       int64         `json:"peak_memory_usage"`
	AverageMemoryUsage    int64         `json:"avg_memory_usage"`

	// Worker pool efficiency
	WorkerUtilization float64 `json:"worker_utilization"`
	BatchEfficiency   float64 `json:"batch_efficiency"`

	// Resource management
	GCTriggers           int64 `json:"gc_triggers"`
	MemoryPressureEvents int64 `json:"memory_pressure_events"`
	TimeoutEvents        int64 `json:"timeout_events"`
}

// Note: MemoryMonitor and TimeoutManager are defined in separate files

// NewPerformanceOptimizer creates an optimized parser for large repositories
func NewPerformanceOptimizer(config PerformanceConfig) *PerformanceOptimizer {
	ctx, cancel := context.WithTimeout(context.Background(), config.TotalTimeout)

	po := &PerformanceOptimizer{
		config:         config,
		workerPool:     make(chan WorkItem, config.WorkerPoolSize),
		resultChan:     make(chan ParseResult, config.WorkerPoolSize),
		progressChan:   make(chan ProgressUpdate, 100),
		ctx:            ctx,
		cancel:         cancel,
		metrics:        &PerformanceMetrics{},
		memoryMonitor:  NewMemoryMonitor(config),
		timeoutManager: NewTimeoutManager(config),
		startTime:      time.Now(),
	}

	return po
}

// ParseRepositoryOptimized performs high-performance repository parsing
func (po *PerformanceOptimizer) ParseRepositoryOptimized(ctx context.Context, files []string, parser *Parser) (*PerformanceMetrics, map[string]*ParseResult, error) {
	po.mu.Lock()
	po.totalFiles = int64(len(files))
	po.startTime = time.Now()
	po.metrics.StartTime = po.startTime
	po.mu.Unlock()

	// Start memory monitoring
	if err := po.memoryMonitor.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start memory monitor: %w", err)
	}
	defer po.memoryMonitor.Stop()

	// Start worker pool
	results := make(map[string]*ParseResult)
	var resultMu sync.RWMutex

	if err := po.startWorkerPool(parser, &results, &resultMu); err != nil {
		return nil, nil, fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Start progress monitoring
	go po.monitorProgress()

	// Process files in batches or stream based on threshold
	if len(files) > po.config.StreamThreshold {
		err := po.processFilesStreaming(ctx, files)
		if err != nil {
			return po.finalize(), results, err
		}
	} else {
		err := po.processFilesBatched(ctx, files)
		if err != nil {
			return po.finalize(), results, err
		}
	}

	// Wait for all work to complete
	po.wg.Wait()

	return po.finalize(), results, nil
}

// startWorkerPool initializes and starts the worker goroutines
func (po *PerformanceOptimizer) startWorkerPool(parser *Parser, results *map[string]*ParseResult, resultMu *sync.RWMutex) error {
	for i := 0; i < po.config.MaxWorkers; i++ {
		po.wg.Add(1)
		go po.worker(i, parser, results, resultMu)
	}
	return nil
}

// worker processes work items with performance monitoring and error handling
func (po *PerformanceOptimizer) worker(workerID int, parser *Parser, results *map[string]*ParseResult, resultMu *sync.RWMutex) {
	defer po.wg.Done()

	// Create dedicated parser for this worker (thread safety)
	workerParser, err := NewParserWithConfig(ErrorConfig{
		MaxErrors:          50,
		ErrorThreshold:     0.3,
		EnableRecovery:     true,
		EnablePartialParse: true,
		LogLevel:           "warn",
	})
	if err != nil {
		atomic.AddInt64(&po.failedFiles, 1)
		return
	}
	defer workerParser.Close()

	for {
		select {
		case <-po.ctx.Done():
			return
		case workItem, ok := <-po.workerPool:
			if !ok {
				return
			}

			po.processWorkItem(workerID, workItem, workerParser, results, resultMu)
		}
	}
}

// processWorkItem handles individual file parsing with timeout and memory management
func (po *PerformanceOptimizer) processWorkItem(workerID int, item WorkItem, parser *Parser, results *map[string]*ParseResult, resultMu *sync.RWMutex) {
	startTime := time.Now()
	defer func() {
		atomic.AddInt64(&po.processedFiles, 1)

		// Update metrics
		processingTime := time.Since(startTime)
		po.mu.Lock()
		if po.metrics.AverageProcessingTime == 0 {
			po.metrics.AverageProcessingTime = processingTime
		} else {
			po.metrics.AverageProcessingTime = (po.metrics.AverageProcessingTime + processingTime) / 2
		}
		po.mu.Unlock()
	}()

	// Check memory pressure before processing
	if po.memoryMonitor.IsUnderPressure() {
		po.memoryMonitor.TriggerGC()
		// Brief pause to allow GC to run
		time.Sleep(100 * time.Millisecond)
	}

	// Create file-specific timeout context
	fileCtx, cancel := context.WithTimeout(po.ctx, po.config.ParseTimeout)
	defer cancel()

	// Parse the file with timeout
	result, err := parser.ParseFile(fileCtx, item.FilePath, item.Content)
	if err != nil {
		atomic.AddInt64(&po.failedFiles, 1)

		// Check if it was a timeout
		if fileCtx.Err() == context.DeadlineExceeded {
			// Timeout events are tracked internally by the timeout manager
		}
		return
	}

	// Store result safely
	resultMu.Lock()
	(*results)[item.FilePath] = result
	resultMu.Unlock()

	// Update success metrics
	atomic.AddInt64(&po.metrics.SuccessfulFiles, 1)
}

// processFilesBatched handles file processing in manageable batches
func (po *PerformanceOptimizer) processFilesBatched(ctx context.Context, files []string) error {
	batchSize := po.config.BatchSize
	batchCount := (len(files) + batchSize - 1) / batchSize // Ceiling division

	batchSemaphore := make(chan struct{}, po.config.MaxBatches)
	var batchWg sync.WaitGroup

	for i := 0; i < batchCount; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[start:end]
		batchID := i

		batchWg.Add(1)
		go func(batchFiles []string, id int) {
			defer batchWg.Done()

			// Acquire batch semaphore
			batchSemaphore <- struct{}{}
			defer func() { <-batchSemaphore }()

			atomic.AddInt64(&po.batchesActive, 1)
			defer atomic.AddInt64(&po.batchesActive, -1)

			po.processBatch(ctx, batchFiles, id)
		}(batch, batchID)
	}

	batchWg.Wait()
	close(po.workerPool) // Signal workers to stop

	return nil
}

// processFilesStreaming handles large file sets with streaming approach
func (po *PerformanceOptimizer) processFilesStreaming(ctx context.Context, files []string) error {
	// For very large repositories, stream files to avoid memory pressure
	fileChan := make(chan string, po.config.WorkerPoolSize)

	// Producer goroutine
	go func() {
		defer close(fileChan)
		for _, file := range files {
			select {
			case <-ctx.Done():
				return
			case fileChan <- file:
			}
		}
	}()

	// Consumer goroutines
	batchID := 0
	for file := range fileChan {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		content, err := po.readFileWithSizeCheck(file)
		if err != nil {
			atomic.AddInt64(&po.failedFiles, 1)
			continue
		}

		workItem := WorkItem{
			FilePath:  file,
			Content:   content,
			BatchID:   batchID,
			Timestamp: time.Now(),
			Priority:  1,
			Metadata:  make(map[string]interface{}),
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case po.workerPool <- workItem:
		}
	}

	close(po.workerPool) // Signal workers to stop
	return nil
}

// processBatch handles a single batch of files
func (po *PerformanceOptimizer) processBatch(ctx context.Context, files []string, batchID int) {
	for _, file := range files {
		select {
		case <-ctx.Done():
			return
		default:
		}

		content, err := po.readFileWithSizeCheck(file)
		if err != nil {
			atomic.AddInt64(&po.failedFiles, 1)
			continue
		}

		workItem := WorkItem{
			FilePath:  file,
			Content:   content,
			BatchID:   batchID,
			Timestamp: time.Now(),
			Priority:  1,
			Metadata:  make(map[string]interface{}),
		}

		select {
		case <-ctx.Done():
			return
		case po.workerPool <- workItem:
		}
	}
}

// readFileWithSizeCheck reads file content with memory-conscious size checking
func (po *PerformanceOptimizer) readFileWithSizeCheck(filePath string) ([]byte, error) {
	// Check file size before reading to prevent memory issues
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// Apply reasonable file size limits to prevent memory exhaustion
	maxFileSize := int64(50 * 1024 * 1024) // 50MB limit for individual files
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file %s exceeds size limit (%d bytes > %d bytes)", filePath, info.Size(), maxFileSize)
	}

	// Check memory pressure before reading large files
	if info.Size() > 10*1024*1024 && po.memoryMonitor.IsUnderPressure() {
		po.memoryMonitor.TriggerGC()
		time.Sleep(100 * time.Millisecond) // Brief pause for GC
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return content, nil
}

// monitorProgress tracks and reports parsing progress
func (po *PerformanceOptimizer) monitorProgress() {
	ticker := time.NewTicker(po.config.ProgressInterval)
	defer ticker.Stop()

	for {
		select {
		case <-po.ctx.Done():
			return
		case <-ticker.C:
			po.reportProgress()
		}
	}
}

// reportProgress generates and sends progress updates
func (po *PerformanceOptimizer) reportProgress() {
	now := time.Now()
	processed := atomic.LoadInt64(&po.processedFiles)
	failed := atomic.LoadInt64(&po.failedFiles)
	total := atomic.LoadInt64(&po.totalFiles)

	var throughput float64
	elapsed := now.Sub(po.startTime).Seconds()
	if elapsed > 0 {
		throughput = float64(processed) / elapsed
	}

	var estimatedRemaining time.Duration
	if throughput > 0 && total > processed {
		remaining := float64(total - processed)
		estimatedRemaining = time.Duration(remaining/throughput) * time.Second
	}

	update := ProgressUpdate{
		TotalFiles:         total,
		ProcessedFiles:     processed,
		FailedFiles:        failed,
		CurrentBatch:       int(atomic.LoadInt64(&po.batchesActive)),
		ActiveWorkers:      po.config.MaxWorkers,
		MemoryUsage:        po.memoryMonitor.GetCurrentUsage(),
		EstimatedRemaining: estimatedRemaining,
		ThroughputFPS:      throughput,
		Timestamp:          now,
	}

	// Send progress update (non-blocking)
	select {
	case po.progressChan <- update:
	default:
	}

	po.mu.Lock()
	po.lastProgressTime = now
	po.mu.Unlock()
}

// finalize completes parsing and generates final performance metrics
func (po *PerformanceOptimizer) finalize() *PerformanceMetrics {
	endTime := time.Now()

	po.mu.Lock()
	defer po.mu.Unlock()

	po.metrics.EndTime = endTime
	po.metrics.TotalDuration = endTime.Sub(po.startTime)
	po.metrics.TotalFilesProcessed = atomic.LoadInt64(&po.processedFiles)
	po.metrics.FailedFiles = atomic.LoadInt64(&po.failedFiles)
	po.metrics.SuccessfulFiles = po.metrics.TotalFilesProcessed - po.metrics.FailedFiles

	// Calculate throughput
	if po.metrics.TotalDuration.Seconds() > 0 {
		po.metrics.ThroughputFPS = float64(po.metrics.TotalFilesProcessed) / po.metrics.TotalDuration.Seconds()
	}

	// Memory metrics
	po.metrics.PeakMemoryUsage = po.memoryMonitor.GetPeakUsage()
	po.metrics.AverageMemoryUsage = po.memoryMonitor.GetAverageUsage()
	po.metrics.GCTriggers = po.memoryMonitor.GetGCTriggers()
	po.metrics.MemoryPressureEvents = po.memoryMonitor.GetPressureEvents()
	po.metrics.TimeoutEvents = po.timeoutManager.GetTimeoutEvents()

	// Calculate efficiency metrics
	po.metrics.WorkerUtilization = float64(po.metrics.TotalFilesProcessed) / float64(po.config.MaxWorkers*int(po.metrics.TotalDuration.Seconds()))

	return po.metrics
}

// GetProgressChannel returns channel for progress updates
func (po *PerformanceOptimizer) GetProgressChannel() <-chan ProgressUpdate {
	return po.progressChan
}

// Stop gracefully stops the performance optimizer
func (po *PerformanceOptimizer) Stop() {
	if po.cancel != nil {
		po.cancel()
	}
}
