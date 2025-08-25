package sandbox

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// MockDockerClient implements a mock Docker client for testing
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerKill(ctx context.Context, containerID string, signal string) error {
	args := m.Called(ctx, containerID, signal)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(types.ContainerJSON), args.Error(1)
}

func (m *MockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	args := m.Called(ctx, containerID, condition)
	return args.Get(0).(<-chan container.WaitResponse), args.Get(1).(<-chan error)
}

// Test helper functions
func createTestLifecycleManager() *LifecycleManager {
	mockLogger := logger.New()
	mockDocker := &MockDockerClient{}

	config := &LifecycleConfig{
		GracefulStopTimeout:   30 * time.Second,
		ForceKillTimeout:      10 * time.Second,
		HealthCheckInterval:   5 * time.Second,
		EnableFailureRecovery: true,
	}

	return NewLifecycleManager(config, mockLogger, mockDocker)
}

func createMockContainerJSON(running bool, healthy bool) types.ContainerJSON {
	state := &types.ContainerState{
		Running: running,
		Status:  "running",
	}

	if healthy {
		state.Health = &types.Health{
			Status: "healthy",
		}
	}

	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    "test-container-123456789012",
			State: state,
		},
	}
}

func TestNewLifecycleManager(t *testing.T) {
	mockLogger := logger.New()
	mockDocker := &MockDockerClient{}

	config := &LifecycleConfig{
		GracefulStopTimeout:   30 * time.Second,
		ForceKillTimeout:      10 * time.Second,
		HealthCheckInterval:   5 * time.Second,
		EnableFailureRecovery: true,
	}

	lm := NewLifecycleManager(config, mockLogger, mockDocker)

	assert.NotNil(t, lm)
	assert.Equal(t, config, lm.config)
	assert.Equal(t, mockLogger, lm.auditLogger)
	assert.Equal(t, mockDocker, lm.dockerClient)
	assert.NotNil(t, lm.containers)
	assert.Equal(t, 0, len(lm.containers))
}

func TestRegisterContainer(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Check if container is registered
	lm.mu.RLock()
	lifecycle, exists := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, StateCreating, lifecycle.CurrentState)
	assert.NotZero(t, lifecycle.CreatedAt)
	assert.NotNil(t, lifecycle.stopChan)
	assert.False(t, lifecycle.healthMonitorActive)
	assert.Equal(t, 0, lifecycle.HealthCheckCount)
	assert.Equal(t, 0, lifecycle.RecoveryAttempts)
}

func TestRegisterContainerInvalidID(t *testing.T) {
	lm := createTestLifecycleManager()

	tests := []struct {
		name        string
		containerID string
		expectError bool
	}{
		{
			name:        "empty container ID",
			containerID: "",
			expectError: true,
		},
		{
			name:        "short container ID",
			containerID: "short",
			expectError: true,
		},
		{
			name:        "valid container ID",
			containerID: "valid-container-123456789012",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lm.RegisterContainer(tt.containerID, nil)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnregisterContainer(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Register container first
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Unregister container
	err = lm.UnregisterContainer(containerID)
	assert.NoError(t, err)

	// Check if container is removed
	lm.mu.RLock()
	_, exists := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.False(t, exists)
}

func TestUpdateContainerState(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Register container first
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Update state
	err = lm.UpdateContainerState(containerID, StateRunning)
	assert.NoError(t, err)

	// Check state
	lm.mu.RLock()
	lifecycle := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.Equal(t, StateRunning, lifecycle.CurrentState)
	assert.NotZero(t, lifecycle.LastStateChange)
}

func TestGracefulStopContainer(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Setup mock expectations
	mockDocker := lm.dockerClient.(*MockDockerClient)

	// Mock successful container stop
	stopOptions := container.StopOptions{
		Timeout: &[]int{int(lm.config.GracefulStopTimeout.Seconds())}[0],
	}
	mockDocker.On("ContainerStop", mock.AnythingOfType("*context.timerCtx"), containerID, stopOptions).Return(nil)

	// Mock container wait for removal
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	statusCh <- container.WaitResponse{StatusCode: 0}
	close(statusCh)
	close(errCh)

	mockDocker.On("ContainerWait", mock.AnythingOfType("*context.timerCtx"), containerID, container.WaitConditionRemoved).Return(
		(<-chan container.WaitResponse)(statusCh),
		(<-chan error)(errCh),
	)

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Update to running state
	err = lm.UpdateContainerState(containerID, StateRunning)
	assert.NoError(t, err)

	// Perform graceful stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = lm.GracefulStopContainer(ctx, containerID)
	assert.NoError(t, err)

	// Check state
	lm.mu.RLock()
	lifecycle := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.Equal(t, StateStopped, lifecycle.CurrentState)
	assert.Equal(t, 1, lm.SuccessfulStops)

	mockDocker.AssertExpectations(t)
}

func TestForceTerminateContainer(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Setup mock expectations
	mockDocker := lm.dockerClient.(*MockDockerClient)

	// Mock successful container kill
	mockDocker.On("ContainerKill", mock.AnythingOfType("*context.timerCtx"), containerID, "SIGKILL").Return(nil)

	// Mock container wait for removal
	statusCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	statusCh <- container.WaitResponse{StatusCode: 0}
	close(statusCh)
	close(errCh)

	mockDocker.On("ContainerWait", mock.AnythingOfType("*context.timerCtx"), containerID, container.WaitConditionRemoved).Return(
		(<-chan container.WaitResponse)(statusCh),
		(<-chan error)(errCh),
	)

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Perform force terminate
	err = lm.ForceTerminateContainer(containerID)
	assert.NoError(t, err)

	// Check state
	lm.mu.RLock()
	lifecycle := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.Equal(t, StateTerminated, lifecycle.CurrentState)
	assert.Equal(t, 1, lm.SuccessfulStops)

	mockDocker.AssertExpectations(t)
}

func TestAttemptRecovery(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Setup mock expectations
	mockDocker := lm.dockerClient.(*MockDockerClient)

	// Mock container inspect - container stopped
	stoppedContainer := createMockContainerJSON(false, false)
	mockDocker.On("ContainerInspect", mock.AnythingOfType("*context.timerCtx"), containerID).Return(stoppedContainer, nil).Once()

	// Mock successful container start
	mockDocker.On("ContainerStart", mock.AnythingOfType("*context.timerCtx"), containerID, types.ContainerStartOptions{}).Return(nil)

	// Mock container inspect for health check - running and healthy
	runningContainer := createMockContainerJSON(true, true)
	mockDocker.On("ContainerInspect", mock.AnythingOfType("*context.timerCtx"), containerID).Return(runningContainer, nil)

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Perform recovery
	err = lm.AttemptRecovery(containerID)
	assert.NoError(t, err)

	// Check state
	lm.mu.RLock()
	lifecycle := lm.containers[containerID]
	lm.mu.RUnlock()

	assert.Equal(t, StateRunning, lifecycle.CurrentState)
	assert.Equal(t, 1, lm.RecoverySuccessful)

	mockDocker.AssertExpectations(t)
}

func TestPerformHealthCheck(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	mockDocker := lm.dockerClient.(*MockDockerClient)

	tests := []struct {
		name        string
		container   types.ContainerJSON
		expectError bool
	}{
		{
			name:        "healthy running container",
			container:   createMockContainerJSON(true, true),
			expectError: false,
		},
		{
			name:        "running container without health check",
			container:   createMockContainerJSON(true, false),
			expectError: false,
		},
		{
			name: "unhealthy container",
			container: func() types.ContainerJSON {
				c := createMockContainerJSON(true, false)
				c.State.Health = &types.Health{Status: "unhealthy"}
				return c
			}(),
			expectError: true,
		},
		{
			name:        "stopped container",
			container:   createMockContainerJSON(false, false),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDocker.ExpectedCalls = nil // Reset mock
			mockDocker.On("ContainerInspect", mock.AnythingOfType("*context.timerCtx"), containerID).Return(tt.container, nil)

			err := lm.performHealthCheck(containerID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockDocker.AssertExpectations(t)
		})
	}
}

func TestStartHealthMonitoring(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Start health monitoring
	err = lm.StartHealthMonitoring(containerID)
	assert.NoError(t, err)

	// Check monitoring state
	lm.mu.RLock()
	lifecycle := lm.containers[containerID]
	isActive := lifecycle.healthMonitorActive
	lm.mu.RUnlock()

	assert.True(t, isActive)

	// Try to start monitoring again (should not error)
	err = lm.StartHealthMonitoring(containerID)
	assert.NoError(t, err)

	// Stop monitoring
	err = lm.StopHealthMonitoring(containerID)
	assert.NoError(t, err)

	// Give goroutine time to stop
	time.Sleep(100 * time.Millisecond)

	// Check monitoring stopped
	lm.mu.RLock()
	isActive = lm.containers[containerID].healthMonitorActive
	lm.mu.RUnlock()

	assert.False(t, isActive)
}

func TestStopHealthMonitoring(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Start monitoring first
	err = lm.StartHealthMonitoring(containerID)
	assert.NoError(t, err)

	// Stop monitoring
	err = lm.StopHealthMonitoring(containerID)
	assert.NoError(t, err)

	// Give goroutine time to stop
	time.Sleep(100 * time.Millisecond)

	// Check monitoring stopped
	lm.mu.RLock()
	isActive := lm.containers[containerID].healthMonitorActive
	lm.mu.RUnlock()

	assert.False(t, isActive)
}

func TestGetContainerStats(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	// Register container
	err := lm.RegisterContainer(containerID, nil)
	assert.NoError(t, err)

	// Get stats
	stats, err := lm.GetContainerStats(containerID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	assert.Equal(t, containerID, stats.ContainerID)
	assert.Equal(t, StateCreating, stats.CurrentState)
	assert.NotZero(t, stats.CreatedAt)
	assert.False(t, stats.IsHealthMonitorActive)
	assert.Equal(t, 0, stats.HealthCheckCount)
	assert.Equal(t, 0, stats.RecoveryAttempts)
}

func TestGetContainerStatsInvalidContainer(t *testing.T) {
	lm := createTestLifecycleManager()
	containerID := "nonexistent-container"

	stats, err := lm.GetContainerStats(containerID)
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "not found in lifecycle manager")
}

func TestGetManagerStats(t *testing.T) {
	lm := createTestLifecycleManager()

	// Register some containers
	containerIDs := []string{
		"test-container-123456789012",
		"test-container-234567890123",
		"test-container-345678901234",
	}

	for _, id := range containerIDs {
		err := lm.RegisterContainer(id, nil)
		assert.NoError(t, err)
	}

	// Update some stats
	lm.SuccessfulStarts = 2
	lm.FailedStarts = 1
	lm.SuccessfulStops = 1
	lm.FailedStops = 0
	lm.RecoverySuccessful = 1
	lm.RecoveryFailed = 0

	// Get manager stats
	stats := lm.GetManagerStats()
	assert.NotNil(t, stats)

	assert.Equal(t, 3, stats.TotalContainers)
	assert.Equal(t, 2, stats.SuccessfulStarts)
	assert.Equal(t, 1, stats.FailedStarts)
	assert.Equal(t, 1, stats.SuccessfulStops)
	assert.Equal(t, 0, stats.FailedStops)
	assert.Equal(t, 1, stats.RecoverySuccessful)
	assert.Equal(t, 0, stats.RecoveryFailed)
	assert.Equal(t, 0, stats.ActiveHealthMonitors)
}

func TestValidateContainerID(t *testing.T) {
	lm := createTestLifecycleManager()

	tests := []struct {
		name        string
		containerID string
		expectError bool
	}{
		{
			name:        "empty container ID",
			containerID: "",
			expectError: true,
		},
		{
			name:        "short container ID",
			containerID: "short",
			expectError: true,
		},
		{
			name:        "valid container ID",
			containerID: "valid-container-123456789012",
			expectError: false,
		},
		{
			name:        "long container ID",
			containerID: "very-long-container-id-123456789012345678901234567890",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lm.validateContainerID(tt.containerID)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLifecycleConfig_String(t *testing.T) {
	config := &LifecycleConfig{
		GracefulStopTimeout: 30 * time.Second,
		ForceKillTimeout:    10 * time.Second,
		HealthCheckInterval: 5 * time.Second,
		EnableFailureRecovery: true,
	}

	str := config.String()
	assert.NotEmpty(t, str)

	// Check if string contains expected fields
	assert.Contains(t, str, "GracefulStopTimeout")
	assert.Contains(t, str, "ForceKillTimeout")
	assert.Contains(t, str, "HealthCheckInterval")
	assert.Contains(t, str, "EnableFailureRecovery")
}

func TestContainerLifecycleStats_String(t *testing.T) {
	stats := &ContainerLifecycleStats{
		ContainerID:           "test-container-123456789012",
		CurrentState:          StateRunning,
		CreatedAt:             time.Now(),
		LastStateChange:       time.Now(),
		HealthCheckCount:      5,
		RecoveryAttempts:      1,
		IsHealthMonitorActive: true,
	}

	str := stats.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "test-container-123456789012")
	assert.Contains(t, str, "running")
	assert.Contains(t, str, "HealthCheckCount")
}

func TestLifecycleManagerStats_String(t *testing.T) {
	stats := &LifecycleManagerStats{
		TotalContainers:      3,
		SuccessfulStarts:     2,
		FailedStarts:         1,
		SuccessfulStops:      1,
		FailedStops:          0,
		RecoverySuccessful:   1,
		RecoveryFailed:       0,
		ActiveHealthMonitors: 2,
	}

	str := stats.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "TotalContainers")
	assert.Contains(t, str, "SuccessfulStarts")
	assert.Contains(t, str, "ActiveHealthMonitors")
}

func TestContainerState_String(t *testing.T) {
	tests := []struct {
		state    ContainerState
		expected string
	}{
		{StateCreating, "creating"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{StateTerminating, "terminating"},
		{StateTerminated, "terminated"},
		{StateRecovering, "recovering"},
		{StateFailed, "failed"},
		{StateUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.state))
		})
	}
}

// Benchmark tests
func BenchmarkRegisterContainer(b *testing.B) {
	lm := createTestLifecycleManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		containerID := fmt.Sprintf("test-container-%d", i)
		if len(containerID) < 12 {
			containerID = containerID + "123456789012"
		}
		_ = lm.RegisterContainer(containerID, map[string]interface{}{"test": true})
	}
}

func BenchmarkUpdateContainerState(b *testing.B) {
	lm := createTestLifecycleManager()
	containerID := "test-container-123456789012"

	_ = lm.RegisterContainer(containerID, map[string]interface{}{"test": true})

	states := []ContainerState{
		StateRunning, StateStopping, StateStopped, StateRunning,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := states[i%len(states)]
		_ = lm.UpdateContainerState(containerID, state)
	}
}
