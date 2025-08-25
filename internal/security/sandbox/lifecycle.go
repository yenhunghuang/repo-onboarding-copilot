// Package sandbox provides advanced container lifecycle management
// with graceful termination, failure recovery, and comprehensive state management.
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// ContainerState represents the current state of a container
type ContainerState string

const (
	StateCreating    ContainerState = "creating"
	StateRunning     ContainerState = "running"
	StateStopping    ContainerState = "stopping"
	StateStopped     ContainerState = "stopped"
	StateTerminating ContainerState = "terminating"
	StateTerminated  ContainerState = "terminated"
	StateRecovering  ContainerState = "recovering"
	StateFailed      ContainerState = "failed"
	StateUnknown     ContainerState = "unknown"
)

// DockerClient interface defines the methods we need from Docker client
type DockerClient interface {
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerKill(ctx context.Context, containerID string, signal string) error
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
}

// LifecycleConfig represents container lifecycle management configuration
type LifecycleConfig struct {
	// Termination settings
	GracefulStopTimeout time.Duration `json:"graceful_stop_timeout"` // 30 seconds
	ForceKillTimeout    time.Duration `json:"force_kill_timeout"`    // 10 seconds
	CleanupTimeout      time.Duration `json:"cleanup_timeout"`       // 60 seconds

	// Recovery settings
	EnableFailureRecovery bool          `json:"enable_failure_recovery"` // true
	MaxRecoveryAttempts   int           `json:"max_recovery_attempts"`   // 3
	RecoveryBackoffBase   time.Duration `json:"recovery_backoff_base"`   // 5 seconds
	RecoveryBackoffMax    time.Duration `json:"recovery_backoff_max"`    // 5 minutes

	// Health check settings
	EnableHealthChecks  bool          `json:"enable_health_checks"`  // true
	HealthCheckInterval time.Duration `json:"health_check_interval"` // 30 seconds
	HealthCheckTimeout  time.Duration `json:"health_check_timeout"`  // 5 seconds
	HealthCheckRetries  int           `json:"health_check_retries"`  // 3

	// State persistence settings
	PersistState         bool          `json:"persist_state"`          // true
	StateCheckInterval   time.Duration `json:"state_check_interval"`   // 10 seconds
	StateRetentionPeriod time.Duration `json:"state_retention_period"` // 24 hours
}

// String returns a string representation of the LifecycleConfig
func (lc *LifecycleConfig) String() string {
	return fmt.Sprintf(
		"LifecycleConfig{GracefulStopTimeout: %v, ForceKillTimeout: %v, EnableFailureRecovery: %v, HealthCheckInterval: %v}",
		lc.GracefulStopTimeout, lc.ForceKillTimeout, lc.EnableFailureRecovery, lc.HealthCheckInterval,
	)
}

// ContainerLifecycle represents a container's lifecycle information
type ContainerLifecycle struct {
	ContainerID   string         `json:"container_id"`
	CurrentState  ContainerState `json:"current_state"`
	PreviousState ContainerState `json:"previous_state"`

	// Timestamps
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	StoppedAt       *time.Time `json:"stopped_at,omitempty"`
	TerminatedAt    *time.Time `json:"terminated_at,omitempty"`
	LastStateChange time.Time  `json:"last_state_change"`

	// Recovery information
	FailureCount     int        `json:"failure_count"`
	RecoveryAttempts int        `json:"recovery_attempts"`
	LastFailure      *time.Time `json:"last_failure,omitempty"`
	LastRecovery     *time.Time `json:"last_recovery,omitempty"`

	// Health information
	HealthStatus              string     `json:"health_status"`
	LastHealthCheck           *time.Time `json:"last_health_check,omitempty"`
	ConsecutiveHealthFailures int        `json:"consecutive_health_failures"`
	HealthCheckCount          int        `json:"health_check_count"`

	// Monitoring
	healthMonitorActive bool          `json:"health_monitor_active"`
	stopChan            chan struct{} `json:"-"`

	// Metadata
	ExitCode     *int                   `json:"exit_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ContainerLifecycleStats provides statistics for a specific container
type ContainerLifecycleStats struct {
	ContainerID           string         `json:"container_id"`
	CurrentState          ContainerState `json:"current_state"`
	CreatedAt             time.Time      `json:"created_at"`
	LastStateChange       time.Time      `json:"last_state_change"`
	HealthCheckCount      int            `json:"health_check_count"`
	RecoveryAttempts      int            `json:"recovery_attempts"`
	IsHealthMonitorActive bool           `json:"is_health_monitor_active"`
}

// String returns a string representation of ContainerLifecycleStats
func (cls *ContainerLifecycleStats) String() string {
	return fmt.Sprintf("ContainerStats{ID:%s, State:%s, HealthChecks:%d, RecoveryAttempts:%d, MonitorActive:%v}",
		cls.ContainerID, cls.CurrentState, cls.HealthCheckCount, cls.RecoveryAttempts, cls.IsHealthMonitorActive)
}

// LifecycleManagerStats provides overall manager statistics
type LifecycleManagerStats struct {
	TotalContainers      int `json:"total_containers"`
	SuccessfulStarts     int `json:"successful_starts"`
	FailedStarts         int `json:"failed_starts"`
	SuccessfulStops      int `json:"successful_stops"`
	FailedStops          int `json:"failed_stops"`
	RecoverySuccessful   int `json:"recovery_successful"`
	RecoveryFailed       int `json:"recovery_failed"`
	ActiveHealthMonitors int `json:"active_health_monitors"`
}

// String returns a string representation of LifecycleManagerStats
func (lms *LifecycleManagerStats) String() string {
	return fmt.Sprintf("ManagerStats{Containers:%d, Starts:%d/%d, Stops:%d/%d, Recovery:%d/%d, Monitors:%d}",
		lms.TotalContainers, lms.SuccessfulStarts, lms.FailedStarts,
		lms.SuccessfulStops, lms.FailedStops, lms.RecoverySuccessful, lms.RecoveryFailed,
		lms.ActiveHealthMonitors)
}

// LifecycleManager manages advanced container lifecycle operations
type LifecycleManager struct {
	config       *LifecycleConfig
	auditLogger  *logger.Logger
	dockerClient DockerClient

	// State tracking
	containers map[string]*ContainerLifecycle
	mu         sync.RWMutex

	// Statistics
	SuccessfulStarts   int
	FailedStarts       int
	SuccessfulStops    int
	FailedStops        int
	RecoverySuccessful int
	RecoveryFailed     int

	// Background operations
	healthCheckCtx     context.Context
	healthCheckCancel  context.CancelFunc
	stateMonitorCtx    context.Context
	stateMonitorCancel context.CancelFunc
}

// NewLifecycleManager creates a new container lifecycle manager
func NewLifecycleManager(config *LifecycleConfig, auditLogger *logger.Logger, dockerClient DockerClient) *LifecycleManager {
	if auditLogger == nil {
		auditLogger = logger.New()
	}
	if config == nil {
		config = &LifecycleConfig{
			GracefulStopTimeout:   30 * time.Second,
			ForceKillTimeout:      10 * time.Second,
			HealthCheckInterval:   5 * time.Second,
			EnableFailureRecovery: true,
		}
	}

	healthCheckCtx, healthCheckCancel := context.WithCancel(context.Background())
	stateMonitorCtx, stateMonitorCancel := context.WithCancel(context.Background())

	lm := &LifecycleManager{
		config:             config,
		auditLogger:        auditLogger,
		dockerClient:       dockerClient,
		containers:         make(map[string]*ContainerLifecycle),
		healthCheckCtx:     healthCheckCtx,
		healthCheckCancel:  healthCheckCancel,
		stateMonitorCtx:    stateMonitorCtx,
		stateMonitorCancel: stateMonitorCancel,
	}

	auditLogger.WithFields(map[string]interface{}{
		"operation":               "lifecycle_manager_created",
		"graceful_stop_timeout":   config.GracefulStopTimeout.Seconds(),
		"force_kill_timeout":      config.ForceKillTimeout.Seconds(),
		"enable_failure_recovery": config.EnableFailureRecovery,
		"max_recovery_attempts":   config.MaxRecoveryAttempts,
		"enable_health_checks":    config.EnableHealthChecks,
		"health_check_interval":   config.HealthCheckInterval.Seconds(),
		"timestamp":               time.Now().Unix(),
	}).Info("Container lifecycle manager initialized")

	return lm
}

// RegisterContainer registers a new container for lifecycle management
func (lm *LifecycleManager) RegisterContainer(containerID string, metadata map[string]interface{}) error {
	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if already registered
	if _, exists := lm.containers[containerID]; exists {
		return fmt.Errorf("container %s is already registered", containerID)
	}

	now := time.Now()
	lifecycle := &ContainerLifecycle{
		ContainerID:     containerID,
		CurrentState:    StateCreating,
		PreviousState:   StateUnknown,
		CreatedAt:       now,
		LastStateChange: now,
		HealthStatus:    "unknown",
		Metadata:        metadata,
		stopChan:        make(chan struct{}),
	}

	lm.containers[containerID] = lifecycle

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":    "container_registered",
		"container_id": containerID,
		"state":        string(lifecycle.CurrentState),
		"timestamp":    now.Unix(),
	}).Info("Container registered for lifecycle management")

	return nil
}

// UpdateContainerState updates the state of a managed container
func (lm *LifecycleManager) UpdateContainerState(containerID string, newState ContainerState) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lifecycle, exists := lm.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s is not registered", containerID)
	}

	previousState := lifecycle.CurrentState
	lifecycle.PreviousState = previousState
	lifecycle.CurrentState = newState
	lifecycle.LastStateChange = time.Now()

	// Update state-specific timestamps
	switch newState {
	case StateRunning:
		now := time.Now()
		lifecycle.StartedAt = &now
	case StateStopped:
		now := time.Now()
		lifecycle.StoppedAt = &now
	case StateTerminated:
		now := time.Now()
		lifecycle.TerminatedAt = &now
	case StateFailed:
		now := time.Now()
		lifecycle.LastFailure = &now
		lifecycle.FailureCount++
	}

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "container_state_changed",
		"container_id":   containerID,
		"previous_state": string(previousState),
		"current_state":  string(newState),
		"failure_count":  lifecycle.FailureCount,
		"timestamp":      lifecycle.LastStateChange.Unix(),
	}).Info("Container state updated")

	return nil
}

// GracefulStopContainer performs a graceful stop of a container with fallback to force termination
func (lm *LifecycleManager) GracefulStopContainer(ctx context.Context, containerID string) error {
	startTime := time.Now()

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":    "graceful_stop_start",
		"container_id": containerID,
		"timestamp":    startTime.Unix(),
	}).Info("Starting graceful container stop")

	// Update state to stopping
	if err := lm.UpdateContainerState(containerID, StateStopping); err != nil {
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "state_update_failed",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Warn("Failed to update container state to stopping")
	}

	// Attempt graceful stop
	stopCtx, cancel := context.WithTimeout(ctx, lm.config.GracefulStopTimeout)
	defer cancel()

	stopCmd := exec.CommandContext(stopCtx, "docker", "stop", containerID)
	if err := stopCmd.Run(); err != nil {
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "graceful_stop_failed",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Warn("Graceful stop failed, attempting force termination")

		// Fallback to force termination
		return lm.ForceTerminateContainer(containerID)
	}

	// Update state to stopped
	if err := lm.UpdateContainerState(containerID, StateStopped); err == nil {
		lm.SuccessfulStops++
	}

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":     "graceful_stop_success",
		"container_id":  containerID,
		"stop_duration": time.Since(startTime).Seconds(),
		"timestamp":     time.Now().Unix(),
	}).Info("Container stopped gracefully")

	return nil
}

// ForceTerminateContainer forcibly terminates a container when graceful stop fails
func (lm *LifecycleManager) ForceTerminateContainer(containerID string) error {
	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	startTime := time.Now()

	// Update state to terminating
	if err := lm.UpdateContainerState(containerID, StateTerminating); err != nil {
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "force_terminate_state_error",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Error("Failed to update container state to terminating")
	}

	// Attempt force kill
	killCtx, cancel := context.WithTimeout(context.Background(), lm.config.ForceKillTimeout)
	defer cancel()

	err := lm.dockerClient.ContainerKill(killCtx, containerID, "SIGKILL")
	if err != nil {
		lm.FailedStops++
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "force_terminate_failed",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Error("Failed to force terminate container")
		return fmt.Errorf("failed to force terminate container %s: %w", containerID, err)
	}

	// Wait for container to be removed
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer waitCancel()

	statusCh, errCh := lm.dockerClient.ContainerWait(waitCtx, containerID, container.WaitConditionRemoved)
	select {
	case <-statusCh:
		// Container removed successfully
	case waitErr := <-errCh:
		if !strings.Contains(waitErr.Error(), "No such container") {
			lm.auditLogger.WithFields(map[string]interface{}{
				"operation":    "force_terminate_wait_error",
				"container_id": containerID,
				"error":        waitErr.Error(),
				"timestamp":    time.Now().Unix(),
			}).Warn("Error waiting for container removal after force kill")
		}
	case <-waitCtx.Done():
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "force_terminate_wait_timeout",
			"container_id": containerID,
			"timeout":      10,
			"timestamp":    time.Now().Unix(),
		}).Warn("Timeout waiting for container removal after force kill")
	}

	// Update state to terminated
	if err := lm.UpdateContainerState(containerID, StateTerminated); err == nil {
		lm.SuccessfulStops++
	}

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":          "force_terminate_success",
		"container_id":       containerID,
		"terminate_duration": time.Since(startTime).Seconds(),
		"timestamp":          time.Now().Unix(),
	}).Info("Container force terminated successfully")

	return nil
}

// AttemptRecovery attempts to recover a failed container
func (lm *LifecycleManager) AttemptRecovery(containerID string) error {
	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	startTime := time.Now()

	// Update state to recovering
	if err := lm.UpdateContainerState(containerID, StateRecovering); err != nil {
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "recovery_state_error",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Error("Failed to update container state to recovering")
	}

	// Get container information
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	containerInfo, err := lm.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		lm.RecoveryFailed++
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "recovery_inspect_failed",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Error("Failed to inspect container for recovery")
		return fmt.Errorf("failed to inspect container for recovery: %w", err)
	}

	// Attempt to start the container if it's stopped
	if !containerInfo.State.Running {
		if err := lm.dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
			lm.RecoveryFailed++
			lm.auditLogger.WithFields(map[string]interface{}{
				"operation":    "recovery_start_failed",
				"container_id": containerID,
				"error":        err.Error(),
				"timestamp":    time.Now().Unix(),
			}).Error("Failed to start container during recovery")
			return fmt.Errorf("failed to start container during recovery: %w", err)
		}
	}

	// Wait for container to be healthy
	if err := lm.waitForHealthy(containerID, lm.config.HealthCheckInterval*5); err != nil {
		lm.RecoveryFailed++
		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "recovery_health_failed",
			"container_id": containerID,
			"error":        err.Error(),
			"timestamp":    time.Now().Unix(),
		}).Error("Container failed health check during recovery")
		return fmt.Errorf("container failed health check during recovery: %w", err)
	}

	// Update state to running
	if err := lm.UpdateContainerState(containerID, StateRunning); err == nil {
		lm.RecoverySuccessful++
	}

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":         "recovery_success",
		"container_id":      containerID,
		"recovery_duration": time.Since(startTime).Seconds(),
		"timestamp":         time.Now().Unix(),
	}).Info("Container recovery completed successfully")

	return nil
}

// StartHealthMonitoring starts continuous health monitoring for a container
func (lm *LifecycleManager) StartHealthMonitoring(containerID string) error {
	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.Lock()
	lifecycle, exists := lm.containers[containerID]
	if !exists {
		lm.mu.Unlock()
		return fmt.Errorf("container %s not found in lifecycle manager", containerID)
	}

	// Prevent duplicate monitoring
	if lifecycle.healthMonitorActive {
		lm.mu.Unlock()
		return nil
	}
	lifecycle.healthMonitorActive = true
	lm.mu.Unlock()

	go func() {
		ticker := time.NewTicker(lm.config.HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := lm.performHealthCheck(containerID); err != nil {
					lm.auditLogger.WithFields(map[string]interface{}{
						"operation":    "health_check_failed",
						"container_id": containerID,
						"error":        err.Error(),
						"timestamp":    time.Now().Unix(),
					}).Error("Health check failed")

					// Attempt recovery if configured
					if lm.config.EnableFailureRecovery {
						if recoveryErr := lm.AttemptRecovery(containerID); recoveryErr != nil {
							lm.auditLogger.WithFields(map[string]interface{}{
								"operation":    "auto_recovery_failed",
								"container_id": containerID,
								"error":        recoveryErr.Error(),
								"timestamp":    time.Now().Unix(),
							}).Error("Automatic recovery failed")
						}
					}
				}
			case <-lifecycle.stopChan:
				lm.mu.Lock()
				lifecycle.healthMonitorActive = false
				lm.mu.Unlock()
				return
			}
		}
	}()

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":    "health_monitoring_started",
		"container_id": containerID,
		"interval":     lm.config.HealthCheckInterval.Seconds(),
		"timestamp":    time.Now().Unix(),
	}).Info("Health monitoring started for container")

	return nil
}

// StopHealthMonitoring stops health monitoring for a container
func (lm *LifecycleManager) StopHealthMonitoring(containerID string) error {
	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()

	lifecycle, exists := lm.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found in lifecycle manager", containerID)
	}

	if lifecycle.healthMonitorActive {
		close(lifecycle.stopChan)
		lifecycle.healthMonitorActive = false

		lm.auditLogger.WithFields(map[string]interface{}{
			"operation":    "health_monitoring_stopped",
			"container_id": containerID,
			"timestamp":    time.Now().Unix(),
		}).Info("Health monitoring stopped for container")
	}

	return nil
}

// performHealthCheck performs a single health check on a container
func (lm *LifecycleManager) performHealthCheck(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if container is still running
	containerInfo, err := lm.dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	if !containerInfo.State.Running {
		return fmt.Errorf("container is not running (state: %s)", containerInfo.State.Status)
	}

	// Check container health if health check is configured
	if containerInfo.State.Health != nil {
		switch containerInfo.State.Health.Status {
		case "unhealthy":
			return fmt.Errorf("container health status is unhealthy")
		case "starting":
			// Health check is still starting, not an error
			return nil
		case "healthy":
			// All good
			return nil
		}
	}

	return nil
}

// waitForHealthy waits for a container to become healthy
func (lm *LifecycleManager) waitForHealthy(containerID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container to become healthy")
		case <-ticker.C:
			if err := lm.performHealthCheck(containerID); err == nil {
				return nil
			}
			// Continue checking
		}
	}
}

// GetContainerStats returns lifecycle statistics for a container
func (lm *LifecycleManager) GetContainerStats(containerID string) (*ContainerLifecycleStats, error) {
	if err := lm.validateContainerID(containerID); err != nil {
		return nil, fmt.Errorf("invalid container ID: %w", err)
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lifecycle, exists := lm.containers[containerID]
	if !exists {
		return nil, fmt.Errorf("container %s not found in lifecycle manager", containerID)
	}

	return &ContainerLifecycleStats{
		ContainerID:           containerID,
		CurrentState:          lifecycle.CurrentState,
		CreatedAt:             lifecycle.CreatedAt,
		LastStateChange:       lifecycle.LastStateChange,
		HealthCheckCount:      lifecycle.HealthCheckCount,
		RecoveryAttempts:      lifecycle.RecoveryAttempts,
		IsHealthMonitorActive: lifecycle.healthMonitorActive,
	}, nil
}

// GetManagerStats returns overall manager statistics
func (lm *LifecycleManager) GetManagerStats() *LifecycleManagerStats {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return &LifecycleManagerStats{
		TotalContainers:      len(lm.containers),
		SuccessfulStarts:     lm.SuccessfulStarts,
		FailedStarts:         lm.FailedStarts,
		SuccessfulStops:      lm.SuccessfulStops,
		FailedStops:          lm.FailedStops,
		RecoverySuccessful:   lm.RecoverySuccessful,
		RecoveryFailed:       lm.RecoveryFailed,
		ActiveHealthMonitors: lm.countActiveHealthMonitors(),
	}
}

// countActiveHealthMonitors counts currently active health monitors
func (lm *LifecycleManager) countActiveHealthMonitors() int {
	count := 0
	for _, lifecycle := range lm.containers {
		if lifecycle.healthMonitorActive {
			count++
		}
	}
	return count
}

// UnregisterContainer removes a container from lifecycle management
func (lm *LifecycleManager) UnregisterContainer(containerID string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if err := lm.validateContainerID(containerID); err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}

	lifecycle, exists := lm.containers[containerID]
	if !exists {
		return fmt.Errorf("container %s not found in lifecycle manager", containerID)
	}

	// Stop health monitoring if active
	if lifecycle.healthMonitorActive {
		close(lifecycle.stopChan)
		lifecycle.healthMonitorActive = false
	}

	// Remove from container map
	delete(lm.containers, containerID)

	lm.auditLogger.WithFields(map[string]interface{}{
		"operation":     "container_unregistered",
		"container_id":  containerID,
		"final_state":   lifecycle.CurrentState,
		"managed_for":   time.Since(lifecycle.CreatedAt).String(),
		"health_checks": lifecycle.HealthCheckCount,
	}).Info("Container unregistered from lifecycle management")

	return nil
}

// validateContainerID validates a container ID format
func (lm *LifecycleManager) validateContainerID(containerID string) error {
	if containerID == "" {
		return fmt.Errorf("container ID cannot be empty")
	}
	if len(containerID) < 12 {
		return fmt.Errorf("container ID must be at least 12 characters")
	}
	return nil
}
