package sandbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

func TestNewResourceMonitor(t *testing.T) {
	tests := []struct {
		name        string
		auditLogger *logger.Logger
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid logger",
			auditLogger: logger.New(),
			wantErr:     false,
		},
		{
			name:        "nil logger",
			auditLogger: nil,
			wantErr:     true,
			errMsg:      "audit logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm, err := NewResourceMonitor(tt.auditLogger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, rm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rm)

				// Verify default thresholds
				assert.Equal(t, 70.0, rm.thresholds.CPUWarningThreshold)
				assert.Equal(t, 90.0, rm.thresholds.CPUCriticalThreshold)
				assert.Equal(t, 80.0, rm.thresholds.MemoryWarningThreshold)
				assert.Equal(t, 95.0, rm.thresholds.MemoryCriticalThreshold)
				assert.Equal(t, 1*time.Hour, rm.thresholds.MaxExecutionTime)
				assert.Equal(t, 5*time.Second, rm.thresholds.MonitorInterval)
				assert.Equal(t, 100, rm.thresholds.MaxProcessCount)
			}
		})
	}
}

func TestParseMemorySize(t *testing.T) {
	auditLogger := logger.New()
	rm, err := NewResourceMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		sizeStr  string
		expected int64
	}{
		{
			name:     "bytes",
			sizeStr:  "1024B",
			expected: 1024,
		},
		{
			name:     "kilobytes",
			sizeStr:  "1KiB",
			expected: 1024,
		},
		{
			name:     "megabytes",
			sizeStr:  "1.5MiB",
			expected: 1572864, // 1.5 * 1024 * 1024
		},
		{
			name:     "gigabytes",
			sizeStr:  "2GiB",
			expected: 2147483648, // 2 * 1024^3
		},
		{
			name:     "empty string",
			sizeStr:  "",
			expected: 0,
		},
		{
			name:     "invalid format",
			sizeStr:  "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rm.parseMemorySize(tt.sizeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckResourceViolations(t *testing.T) {
	auditLogger := logger.New()
	rm, err := NewResourceMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name               string
		usage              *ResourceUsage
		expectedViolations int
		expectedSeverity   []ViolationSeverity
	}{
		{
			name: "no violations",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    50.0,
				MemoryPercent: 60.0,
				ProcessCount:  50,
				Timestamp:     time.Now(),
			},
			expectedViolations: 0,
		},
		{
			name: "CPU warning",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    75.0, // Above warning threshold (70%)
				MemoryPercent: 60.0,
				ProcessCount:  50,
				Timestamp:     time.Now(),
			},
			expectedViolations: 1,
			expectedSeverity:   []ViolationSeverity{SeverityWarning},
		},
		{
			name: "CPU critical",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    95.0, // Above critical threshold (90%)
				MemoryPercent: 60.0,
				ProcessCount:  50,
				Timestamp:     time.Now(),
			},
			expectedViolations: 1,
			expectedSeverity:   []ViolationSeverity{SeverityCritical},
		},
		{
			name: "Memory critical",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    50.0,
				MemoryPercent: 96.0, // Above critical threshold (95%)
				ProcessCount:  50,
				Timestamp:     time.Now(),
			},
			expectedViolations: 1,
			expectedSeverity:   []ViolationSeverity{SeverityCritical},
		},
		{
			name: "Process limit exceeded",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    50.0,
				MemoryPercent: 60.0,
				ProcessCount:  150, // Above limit (100)
				Timestamp:     time.Now(),
			},
			expectedViolations: 1,
			expectedSeverity:   []ViolationSeverity{SeverityCritical},
		},
		{
			name: "Multiple violations",
			usage: &ResourceUsage{
				ContainerID:   "test-container",
				CPUPercent:    95.0, // Critical
				MemoryPercent: 85.0, // Warning
				ProcessCount:  150,  // Critical
				Timestamp:     time.Now(),
			},
			expectedViolations: 3,
			expectedSeverity:   []ViolationSeverity{SeverityCritical, SeverityWarning, SeverityCritical},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rm.checkResourceViolations(tt.usage)
			assert.Len(t, violations, tt.expectedViolations)

			for i, violation := range violations {
				if i < len(tt.expectedSeverity) {
					assert.Equal(t, tt.expectedSeverity[i], violation.Severity)
				}
				assert.Equal(t, tt.usage.ContainerID, violation.ContainerID)
				assert.Equal(t, tt.usage.Timestamp, violation.Timestamp)
			}
		})
	}
}

func TestValidateThresholds(t *testing.T) {
	auditLogger := logger.New()
	rm, err := NewResourceMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name       string
		thresholds *ResourceThresholds
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid thresholds",
			thresholds: &ResourceThresholds{
				CPUWarningThreshold:     70.0,
				CPUCriticalThreshold:    90.0,
				MemoryWarningThreshold:  80.0,
				MemoryCriticalThreshold: 95.0,
				MaxExecutionTime:        1 * time.Hour,
				MonitorInterval:         5 * time.Second,
				MaxProcessCount:         100,
			},
			wantErr: false,
		},
		{
			name: "invalid CPU warning threshold",
			thresholds: &ResourceThresholds{
				CPUWarningThreshold:     -10.0,
				CPUCriticalThreshold:    90.0,
				MemoryWarningThreshold:  80.0,
				MemoryCriticalThreshold: 95.0,
				MaxExecutionTime:        1 * time.Hour,
				MonitorInterval:         5 * time.Second,
				MaxProcessCount:         100,
			},
			wantErr: true,
			errMsg:  "invalid CPU warning threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rm.SetResourceThresholds(tt.thresholds)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
