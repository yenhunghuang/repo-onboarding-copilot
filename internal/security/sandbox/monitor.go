// Package sandbox provides comprehensive resource monitoring for secure container orchestration
// with real-time CPU and memory usage tracking and automatic termination triggers.
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// ResourceThresholds defines limits for resource consumption
type ResourceThresholds struct {
	// CPU limits (percentage)
	CPUWarningThreshold  float64 `json:"cpu_warning_threshold"`  // 70%
	CPUCriticalThreshold float64 `json:"cpu_critical_threshold"` // 90%

	// Memory limits (percentage of allocated memory)
	MemoryWarningThreshold  float64 `json:"memory_warning_threshold"`  // 80%
	MemoryCriticalThreshold float64 `json:"memory_critical_threshold"` // 95%

	// Time limits
	MaxExecutionTime time.Duration `json:"max_execution_time"` // 1 hour
	MonitorInterval  time.Duration `json:"monitor_interval"`   // 5 seconds

	// Process limits
	MaxProcessCount int `json:"max_process_count"` // 100
}

// ResourceUsage represents current resource consumption
type ResourceUsage struct {
	ContainerID string    `json:"container_id"`
	Timestamp   time.Time `json:"timestamp"`

	// CPU metrics
	CPUPercent float64 `json:"cpu_percent"`
	CPUUsage   string  `json:"cpu_usage"`

	// Memory metrics
	MemoryUsed    int64   `json:"memory_used"`  // bytes
	MemoryLimit   int64   `json:"memory_limit"` // bytes
	MemoryPercent float64 `json:"memory_percent"`

	// Process metrics
	ProcessCount int `json:"process_count"`

	// Network metrics
	NetworkRx int64 `json:"network_rx"` // bytes received
	NetworkTx int64 `json:"network_tx"` // bytes transmitted

	// Execution time
	ExecutionTime time.Duration `json:"execution_time"`
}

// ResourceMonitor tracks and enforces resource limits for containers
type ResourceMonitor struct {
	thresholds  *ResourceThresholds
	auditLogger *logger.Logger

	// Monitoring state
	activeMonitors map[string]*MonitorSession
	mu             sync.RWMutex

	// Statistics
	TotalViolations int64
	TerminatedCount int64
}

// MonitorSession represents an active container monitoring session
type MonitorSession struct {
	ContainerID    string
	StartTime      time.Time
	LastUsage      *ResourceUsage
	ViolationCount int

	// Control channels
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// ResourceViolation represents a resource limit violation
type ResourceViolation struct {
	ContainerID   string            `json:"container_id"`
	ViolationType string            `json:"violation_type"`
	Threshold     float64           `json:"threshold"`
	CurrentValue  float64           `json:"current_value"`
	Severity      ViolationSeverity `json:"severity"`
	Timestamp     time.Time         `json:"timestamp"`
	Action        string            `json:"action"`
}

// ViolationSeverity represents the severity of a resource violation
type ViolationSeverity string

const (
	SeverityWarning   ViolationSeverity = "warning"
	SeverityCritical  ViolationSeverity = "critical"
	SeverityEmergency ViolationSeverity = "emergency"
)

// NewResourceMonitor creates a new resource monitoring system
func NewResourceMonitor(auditLogger *logger.Logger) (*ResourceMonitor, error) {
	if auditLogger == nil {
		return nil, fmt.Errorf("audit logger cannot be nil")
	}

	thresholds := &ResourceThresholds{
		CPUWarningThreshold:     70.0,
		CPUCriticalThreshold:    90.0,
		MemoryWarningThreshold:  80.0,
		MemoryCriticalThreshold: 95.0,
		MaxExecutionTime:        1 * time.Hour,
		MonitorInterval:         5 * time.Second,
		MaxProcessCount:         100,
	}

	rm := &ResourceMonitor{
		thresholds:     thresholds,
		auditLogger:    auditLogger,
		activeMonitors: make(map[string]*MonitorSession),
	}

	auditLogger.WithFields(map[string]interface{}{
		"operation":                 "resource_monitor_created",
		"cpu_warning_threshold":     thresholds.CPUWarningThreshold,
		"cpu_critical_threshold":    thresholds.CPUCriticalThreshold,
		"memory_warning_threshold":  thresholds.MemoryWarningThreshold,
		"memory_critical_threshold": thresholds.MemoryCriticalThreshold,
		"max_execution_time":        thresholds.MaxExecutionTime.Seconds(),
		"monitor_interval":          thresholds.MonitorInterval.Seconds(),
		"timestamp":                 time.Now().Unix(),
	}).Info("Resource monitor initialized")

	return rm, nil
}

// StartMonitoring begins resource monitoring for a container
func (rm *ResourceMonitor) StartMonitoring(ctx context.Context, containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if already monitoring
	if _, exists := rm.activeMonitors[containerID]; exists {
		return fmt.Errorf("already monitoring container %s", containerID)
	}

	// Create monitoring session
	monitorCtx, cancel := context.WithCancel(ctx)
	session := &MonitorSession{
		ContainerID: containerID,
		StartTime:   time.Now(),
		ctx:         monitorCtx,
		cancel:      cancel,
		done:        make(chan struct{}),
	}

	rm.activeMonitors[containerID] = session

	rm.auditLogger.WithFields(map[string]interface{}{
		"operation":    "resource_monitoring_start",
		"container_id": containerID,
		"start_time":   session.StartTime.Unix(),
		"timestamp":    time.Now().Unix(),
	}).Info("Resource monitoring started")

	// Start monitoring goroutine
	go rm.monitorContainer(session)

	return nil
}

// monitorContainer runs the monitoring loop for a specific container
func (rm *ResourceMonitor) monitorContainer(session *MonitorSession) {
	defer close(session.done)

	ticker := time.NewTicker(rm.thresholds.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-session.ctx.Done():
			rm.auditLogger.WithFields(map[string]interface{}{
				"operation":    "resource_monitoring_cancelled",
				"container_id": session.ContainerID,
				"duration":     time.Since(session.StartTime).Seconds(),
				"timestamp":    time.Now().Unix(),
			}).Info("Resource monitoring cancelled")
			return

		case <-ticker.C:
			usage, err := rm.collectResourceUsage(session.ContainerID)
			if err != nil {
				rm.auditLogger.WithFields(map[string]interface{}{
					"operation":    "resource_collection_error",
					"container_id": session.ContainerID,
					"error":        err.Error(),
					"timestamp":    time.Now().Unix(),
				}).Error("Failed to collect resource usage")
				continue
			}

			session.LastUsage = usage

			// Check for violations
			violations := rm.checkResourceViolations(usage)
			if len(violations) > 0 {
				session.ViolationCount++
				rm.TotalViolations += int64(len(violations))

				for _, violation := range violations {
					rm.handleResourceViolation(violation)
				}
			}
		}
	}
}

// collectResourceUsage gathers current resource usage statistics for a container
func (rm *ResourceMonitor) collectResourceUsage(containerID string) (*ResourceUsage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Execute docker stats command to get resource usage
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format",
		"table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.PIDs}}", containerID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to collect resource usage: %w, output: %s", err, string(output))
	}

	return rm.parseResourceUsage(containerID, string(output))
}

// parseResourceUsage parses Docker stats output into ResourceUsage struct
func (rm *ResourceMonitor) parseResourceUsage(containerID, output string) (*ResourceUsage, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid docker stats output format")
	}

	// Parse the data line (skip header)
	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return nil, fmt.Errorf("insufficient fields in docker stats output")
	}

	usage := &ResourceUsage{
		ContainerID: containerID,
		Timestamp:   time.Now(),
	}

	// Parse CPU percentage
	cpuStr := strings.TrimSuffix(fields[1], "%")
	if cpuPercent, err := strconv.ParseFloat(cpuStr, 64); err == nil {
		usage.CPUPercent = cpuPercent
	}

	// Parse memory usage (e.g., "1.5GiB / 2GiB")
	memUsage := fields[2]
	if memParts := strings.Split(memUsage, " / "); len(memParts) == 2 {
		usage.MemoryUsed = rm.parseMemorySize(memParts[0])
		usage.MemoryLimit = rm.parseMemorySize(memParts[1])
	}

	// Parse memory percentage
	memPercStr := strings.TrimSuffix(fields[3], "%")
	if memPercent, err := strconv.ParseFloat(memPercStr, 64); err == nil {
		usage.MemoryPercent = memPercent
	}

	// Parse process count
	if pids, err := strconv.Atoi(fields[5]); err == nil {
		usage.ProcessCount = pids
	}

	return usage, nil
}

// parseMemorySize converts memory size string to bytes
func (rm *ResourceMonitor) parseMemorySize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0
	}

	// Extract numeric part and unit
	var value float64
	var unit string

	if _, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit); err != nil {
		return 0
	}

	// Convert to bytes based on unit
	switch strings.ToUpper(unit) {
	case "B":
		return int64(value)
	case "KIB", "KB":
		return int64(value * 1024)
	case "MIB", "MB":
		return int64(value * 1024 * 1024)
	case "GIB", "GB":
		return int64(value * 1024 * 1024 * 1024)
	case "TIB", "TB":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}

// checkResourceViolations evaluates current usage against thresholds
func (rm *ResourceMonitor) checkResourceViolations(usage *ResourceUsage) []ResourceViolation {
	var violations []ResourceViolation

	// Check CPU usage
	if usage.CPUPercent >= rm.thresholds.CPUCriticalThreshold {
		violations = append(violations, ResourceViolation{
			ContainerID:   usage.ContainerID,
			ViolationType: "cpu_critical",
			Threshold:     rm.thresholds.CPUCriticalThreshold,
			CurrentValue:  usage.CPUPercent,
			Severity:      SeverityCritical,
			Timestamp:     usage.Timestamp,
			Action:        "terminate_container",
		})
	} else if usage.CPUPercent >= rm.thresholds.CPUWarningThreshold {
		violations = append(violations, ResourceViolation{
			ContainerID:   usage.ContainerID,
			ViolationType: "cpu_warning",
			Threshold:     rm.thresholds.CPUWarningThreshold,
			CurrentValue:  usage.CPUPercent,
			Severity:      SeverityWarning,
			Timestamp:     usage.Timestamp,
			Action:        "log_warning",
		})
	}

	// Check memory usage
	if usage.MemoryPercent >= rm.thresholds.MemoryCriticalThreshold {
		violations = append(violations, ResourceViolation{
			ContainerID:   usage.ContainerID,
			ViolationType: "memory_critical",
			Threshold:     rm.thresholds.MemoryCriticalThreshold,
			CurrentValue:  usage.MemoryPercent,
			Severity:      SeverityCritical,
			Timestamp:     usage.Timestamp,
			Action:        "terminate_container",
		})
	} else if usage.MemoryPercent >= rm.thresholds.MemoryWarningThreshold {
		violations = append(violations, ResourceViolation{
			ContainerID:   usage.ContainerID,
			ViolationType: "memory_warning",
			Threshold:     rm.thresholds.MemoryWarningThreshold,
			CurrentValue:  usage.MemoryPercent,
			Severity:      SeverityWarning,
			Timestamp:     usage.Timestamp,
			Action:        "log_warning",
		})
	}

	// Check process count
	if usage.ProcessCount > rm.thresholds.MaxProcessCount {
		violations = append(violations, ResourceViolation{
			ContainerID:   usage.ContainerID,
			ViolationType: "process_limit",
			Threshold:     float64(rm.thresholds.MaxProcessCount),
			CurrentValue:  float64(usage.ProcessCount),
			Severity:      SeverityCritical,
			Timestamp:     usage.Timestamp,
			Action:        "terminate_container",
		})
	}

	// Check execution time
	if session := rm.getMonitorSession(usage.ContainerID); session != nil {
		executionTime := time.Since(session.StartTime)
		if executionTime > rm.thresholds.MaxExecutionTime {
			violations = append(violations, ResourceViolation{
				ContainerID:   usage.ContainerID,
				ViolationType: "execution_timeout",
				Threshold:     rm.thresholds.MaxExecutionTime.Seconds(),
				CurrentValue:  executionTime.Seconds(),
				Severity:      SeverityEmergency,
				Timestamp:     usage.Timestamp,
				Action:        "force_terminate_container",
			})
		}
	}

	return violations
}

// handleResourceViolation processes a resource violation and takes appropriate action
func (rm *ResourceMonitor) handleResourceViolation(violation ResourceViolation) {
	rm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "resource_violation",
		"container_id":   violation.ContainerID,
		"violation_type": violation.ViolationType,
		"severity":       string(violation.Severity),
		"threshold":      violation.Threshold,
		"current_value":  violation.CurrentValue,
		"action":         violation.Action,
		"timestamp":      violation.Timestamp.Unix(),
	}).Warn("Resource violation detected")

	switch violation.Action {
	case "terminate_container", "force_terminate_container":
		if err := rm.terminateContainer(violation.ContainerID, violation.Action == "force_terminate_container"); err != nil {
			rm.auditLogger.WithFields(map[string]interface{}{
				"operation":    "container_termination_failed",
				"container_id": violation.ContainerID,
				"error":        err.Error(),
				"timestamp":    time.Now().Unix(),
			}).Error("Failed to terminate container due to resource violation")
		} else {
			rm.TerminatedCount++
			rm.auditLogger.WithFields(map[string]interface{}{
				"operation":      "container_terminated",
				"container_id":   violation.ContainerID,
				"violation_type": violation.ViolationType,
				"timestamp":      time.Now().Unix(),
			}).Info("Container terminated due to resource violation")
		}
	case "log_warning":
		// Already logged above, no additional action needed
	}
}

// terminateContainer stops a container due to resource violation
func (rm *ResourceMonitor) terminateContainer(containerID string, forceKill bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if forceKill {
		cmd = exec.CommandContext(ctx, "docker", "kill", containerID)
	} else {
		cmd = exec.CommandContext(ctx, "docker", "stop", containerID)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to terminate container %s: %w", containerID, err)
	}

	return nil
}

// getMonitorSession safely retrieves a monitoring session
func (rm *ResourceMonitor) getMonitorSession(containerID string) *MonitorSession {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return rm.activeMonitors[containerID]
}

// StopMonitoring stops resource monitoring for a container
func (rm *ResourceMonitor) StopMonitoring(containerID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	session, exists := rm.activeMonitors[containerID]
	if !exists {
		return fmt.Errorf("no monitoring session found for container %s", containerID)
	}

	// Cancel monitoring
	session.cancel()

	// Wait for monitoring goroutine to finish
	select {
	case <-session.done:
	case <-time.After(10 * time.Second):
		rm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "monitoring_stop_timeout",
			"container_id": containerID,
			"timestamp":    time.Now().Unix(),
		}).Warn("Timeout waiting for monitoring session to stop")
	}

	// Remove from active monitors
	delete(rm.activeMonitors, containerID)

	rm.auditLogger.WithFields(map[string]interface{}{
		"operation":    "resource_monitoring_stopped",
		"container_id": containerID,
		"duration":     time.Since(session.StartTime).Seconds(),
		"violations":   session.ViolationCount,
		"timestamp":    time.Now().Unix(),
	}).Info("Resource monitoring stopped")

	return nil
}

// GetResourceStatistics returns current resource monitoring statistics
func (rm *ResourceMonitor) GetResourceStatistics() map[string]interface{} {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := map[string]interface{}{
		"active_monitors":  len(rm.activeMonitors),
		"total_violations": rm.TotalViolations,
		"terminated_count": rm.TerminatedCount,
		"thresholds":       rm.thresholds,
		"timestamp":        time.Now().Unix(),
	}

	// Add current usage for all active monitors
	activeUsage := make(map[string]*ResourceUsage)
	for id, session := range rm.activeMonitors {
		if session.LastUsage != nil {
			activeUsage[id] = session.LastUsage
		}
	}
	stats["active_usage"] = activeUsage

	return stats
}

// SetResourceThresholds updates the resource monitoring thresholds
func (rm *ResourceMonitor) SetResourceThresholds(thresholds *ResourceThresholds) error {
	if thresholds == nil {
		return fmt.Errorf("resource thresholds cannot be nil")
	}

	if err := rm.validateThresholds(thresholds); err != nil {
		return fmt.Errorf("invalid resource thresholds: %w", err)
	}

	rm.thresholds = thresholds

	rm.auditLogger.WithFields(map[string]interface{}{
		"operation":                 "resource_thresholds_updated",
		"cpu_warning_threshold":     thresholds.CPUWarningThreshold,
		"cpu_critical_threshold":    thresholds.CPUCriticalThreshold,
		"memory_warning_threshold":  thresholds.MemoryWarningThreshold,
		"memory_critical_threshold": thresholds.MemoryCriticalThreshold,
		"max_execution_time":        thresholds.MaxExecutionTime.Seconds(),
		"timestamp":                 time.Now().Unix(),
	}).Info("Resource monitoring thresholds updated")

	return nil
}

// validateThresholds validates resource threshold configuration
func (rm *ResourceMonitor) validateThresholds(thresholds *ResourceThresholds) error {
	if thresholds.CPUWarningThreshold <= 0 || thresholds.CPUWarningThreshold > 100 {
		return fmt.Errorf("invalid CPU warning threshold: %f", thresholds.CPUWarningThreshold)
	}

	if thresholds.CPUCriticalThreshold <= thresholds.CPUWarningThreshold || thresholds.CPUCriticalThreshold > 100 {
		return fmt.Errorf("invalid CPU critical threshold: %f", thresholds.CPUCriticalThreshold)
	}

	if thresholds.MemoryWarningThreshold <= 0 || thresholds.MemoryWarningThreshold > 100 {
		return fmt.Errorf("invalid memory warning threshold: %f", thresholds.MemoryWarningThreshold)
	}

	if thresholds.MemoryCriticalThreshold <= thresholds.MemoryWarningThreshold || thresholds.MemoryCriticalThreshold > 100 {
		return fmt.Errorf("invalid memory critical threshold: %f", thresholds.MemoryCriticalThreshold)
	}

	if thresholds.MaxExecutionTime <= 0 {
		return fmt.Errorf("invalid max execution time: %v", thresholds.MaxExecutionTime)
	}

	if thresholds.MonitorInterval <= 0 {
		return fmt.Errorf("invalid monitor interval: %v", thresholds.MonitorInterval)
	}

	if thresholds.MaxProcessCount <= 0 {
		return fmt.Errorf("invalid max process count: %d", thresholds.MaxProcessCount)
	}

	return nil
}
