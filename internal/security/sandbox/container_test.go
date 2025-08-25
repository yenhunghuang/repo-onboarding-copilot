package sandbox

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

func TestNewContainerOrchestrator(t *testing.T) {
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
			co, err := NewContainerOrchestrator(tt.auditLogger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, co)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, co)
				assert.NotNil(t, co.config)

				// Verify default configuration
				assert.Equal(t, 2, co.config.MemoryLimitGB)
				assert.Equal(t, "4.0", co.config.CPULimit)
				assert.Equal(t, "none", co.config.NetworkMode)
				assert.True(t, co.config.ReadOnly)
				assert.True(t, co.config.NoNewPrivs)
				assert.Equal(t, "alpine:latest", co.config.Image)
				assert.Equal(t, "/workspace", co.config.WorkDir)
				assert.Equal(t, 1*time.Hour, co.config.Timeout)
				assert.Equal(t, "1000:1000", co.config.User)
			}
		})
	}
}

func TestContainerConfig(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	// Test getting configuration
	config := co.GetConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 2, config.MemoryLimitGB)

	// Test setting new configuration
	newConfig := &ContainerConfig{
		MemoryLimitGB: 4,
		CPULimit:      "2.0",
		NetworkMode:   "bridge",
		ReadOnly:      false,
		NoNewPrivs:    false,
		Image:         "ubuntu:20.04",
		WorkDir:       "/app",
		Timeout:       30 * time.Minute,
		User:          "0:0",
	}

	co.SetConfig(newConfig)
	updatedConfig := co.GetConfig()
	assert.Equal(t, 4, updatedConfig.MemoryLimitGB)
	assert.Equal(t, "2.0", updatedConfig.CPULimit)
	assert.Equal(t, "bridge", updatedConfig.NetworkMode)
	assert.False(t, updatedConfig.ReadOnly)
	assert.False(t, updatedConfig.NoNewPrivs)
	assert.Equal(t, "ubuntu:20.04", updatedConfig.Image)
	assert.Equal(t, "/app", updatedConfig.WorkDir)
	assert.Equal(t, 30*time.Minute, updatedConfig.Timeout)
	assert.Equal(t, "0:0", updatedConfig.User)
}

func TestBuildDockerArgs(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	volumeMounts := map[string]string{
		"/host/path1": "/container/path1",
		"/host/path2": "/container/path2",
	}

	args := co.buildDockerArgs(volumeMounts)

	// Verify basic arguments
	assert.Contains(t, args, "run")
	assert.Contains(t, args, "--detach")
	assert.Contains(t, args, "--rm")
	assert.Contains(t, args, "--memory=2g")
	assert.Contains(t, args, "--cpus=4.0")
	assert.Contains(t, args, "--network=none")
	assert.Contains(t, args, "--user=65534:65534")
	assert.Contains(t, args, "--workdir=/workspace")

	// Verify security options
	assert.Contains(t, args, "--read-only")
	assert.Contains(t, args, "--security-opt")
	assert.Contains(t, args, "--cap-drop")
	assert.Contains(t, args, "--cap-add")
	assert.Contains(t, args, "--pids-limit")
	assert.Contains(t, args, "--ulimit")

	// Verify volume mounts
	assert.Contains(t, args, "-v")

	// Verify tmpfs mounts for read-only containers
	assert.Contains(t, args, "--tmpfs")

	// Verify image and command
	assert.Contains(t, args, "gcr.io/distroless/static-debian12")
	assert.Contains(t, args, "sleep")
	assert.Contains(t, args, "3600")
}

func TestBuildDockerArgsWithoutReadOnly(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	// Modify config to disable read-only
	config := co.GetConfig()
	config.ReadOnly = false
	config.NoNewPrivs = false
	co.SetConfig(config)

	volumeMounts := map[string]string{
		"/host/path": "/container/path",
	}

	args := co.buildDockerArgs(volumeMounts)

	// Should not contain read-only flags
	assert.NotContains(t, args, "--read-only")

	// Should not contain no-new-privileges if disabled
	noNewPrivsFound := false
	for i, arg := range args {
		if arg == "--security-opt" && i+1 < len(args) && args[i+1] == "no-new-privileges:true" {
			noNewPrivsFound = true
			break
		}
	}
	assert.False(t, noNewPrivsFound)
}

func TestSanitizeDockerOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal output",
			input:    "Container started successfully\nContainer ID: abc123",
			expected: "Container started successfully\nContainer ID: abc123",
		},
		{
			name:     "output with password",
			input:    "Error: invalid password provided\nContainer failed to start",
			expected: "[REDACTED SENSITIVE LINE]\nContainer failed to start",
		},
		{
			name:     "output with token",
			input:    "Using token: abc123\nAuthentication successful",
			expected: "[REDACTED SENSITIVE LINE]\nAuthentication successful",
		},
		{
			name:     "output with secret",
			input:    "Loading secret from file\nContainer running",
			expected: "[REDACTED SENSITIVE LINE]\nContainer running",
		},
		{
			name:     "output with key",
			input:    "SSH key loaded\nConnection established",
			expected: "[REDACTED SENSITIVE LINE]\nConnection established",
		},
		{
			name:     "multiple sensitive lines",
			input:    "Starting container\nPassword: 123\nToken: abc\nContainer ready",
			expected: "Starting container\n[REDACTED SENSITIVE LINE]\n[REDACTED SENSITIVE LINE]\nContainer ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeDockerOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal command",
			input:    "ls -la /workspace",
			expected: "ls -la /workspace",
		},
		{
			name:     "command with password",
			input:    "mysql -u user -ppassword123 database",
			expected: "[REDACTED SENSITIVE COMMAND]",
		},
		{
			name:     "command with token",
			input:    "curl -H 'Authorization: token abc123' api.github.com",
			expected: "[REDACTED SENSITIVE COMMAND]",
		},
		{
			name:     "command with secret",
			input:    "kubectl create secret generic mysecret",
			expected: "[REDACTED SENSITIVE COMMAND]",
		},
		{
			name:     "command with key",
			input:    "ssh-keygen -t rsa -f /tmp/key",
			expected: "[REDACTED SENSITIVE COMMAND]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeCommand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainerResult(t *testing.T) {
	result := &ContainerResult{
		ContainerID:   "abc123",
		ExitCode:      0,
		Output:        "Hello, World!",
		Error:         nil,
		ExecutionTime: 5 * time.Second,
		ResourceUsage: map[string]interface{}{
			"cpu_usage":    "50%",
			"memory_usage": "1GB",
		},
	}

	assert.Equal(t, "abc123", result.ContainerID)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "Hello, World!", result.Output)
	assert.Nil(t, result.Error)
	assert.Equal(t, 5*time.Second, result.ExecutionTime)
	assert.Equal(t, "50%", result.ResourceUsage["cpu_usage"])
	assert.Equal(t, "1GB", result.ResourceUsage["memory_usage"])
}

// Integration test - requires Docker to be available
func TestValidateDockerAvailability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	err = co.ValidateDockerAvailability()
	// This test might fail if Docker is not available, which is acceptable
	if err != nil {
		t.Logf("Docker validation failed (expected if Docker is not available): %v", err)
	}
}

func TestDefaultContainerConfigValues(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	config := co.GetConfig()

	// Verify all default values match security requirements
	assert.Equal(t, 2, config.MemoryLimitGB, "Memory limit should default to 2GB")
	assert.Equal(t, "4.0", config.CPULimit, "CPU limit should default to 4.0 cores")
	assert.Equal(t, "none", config.NetworkMode, "Network should be isolated by default")
	assert.True(t, config.ReadOnly, "Filesystem should be read-only by default")
	assert.True(t, config.NoNewPrivs, "Privilege escalation should be disabled by default")
	assert.Equal(t, "gcr.io/distroless/static-debian12", config.Image, "Should use minimal distroless image by default")
	assert.Equal(t, "/workspace", config.WorkDir, "Should use standard workspace directory")
	assert.Equal(t, 1*time.Hour, config.Timeout, "Should have 1-hour timeout by default")
	assert.Equal(t, "65534:65534", config.User, "Should run as nobody user by default")
}

func TestDockerArgsSecurityCompliance(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	args := co.buildDockerArgs(map[string]string{})

	// Convert args to string for easier searching
	argsStr := strings.Join(args, " ")

	// Verify security compliance
	assert.Contains(t, argsStr, "--cap-drop ALL", "Should drop all capabilities")
	assert.Contains(t, argsStr, "--pids-limit 100", "Should limit number of processes")
	assert.Contains(t, argsStr, "--ulimit nofile=1024:1024", "Should limit file descriptors")
	assert.Contains(t, argsStr, "--network=none", "Should disable network access")
	assert.Contains(t, argsStr, "--user=65534:65534", "Should run as non-root user")
	assert.Contains(t, argsStr, "--read-only", "Should use read-only filesystem")
	assert.Contains(t, argsStr, "--tmpfs /tmp:noexec,nosuid", "Should provide secure temporary space")
}

func TestEnhancedSecurityConfiguration(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	config := co.GetConfig()

	// Verify enhanced security defaults
	assert.Equal(t, "distroless", config.BaseImageType, "Should default to distroless base image")
	assert.Equal(t, "gcr.io/distroless/static-debian12", config.Image, "Should use distroless static image")
	assert.Equal(t, "65534:65534", config.User, "Should use nobody user for enhanced security")
	assert.Equal(t, "default", config.SeccompProfile, "Should use default seccomp profile")
	assert.Equal(t, "docker-default", config.ApparmorProfile, "Should use docker-default AppArmor profile")
	assert.True(t, config.UserNS, "Should enable user namespace isolation")
}

func TestSetBaseImage(t *testing.T) {
	auditLogger := logger.New()
	co, err := NewContainerOrchestrator(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name          string
		imageType     string
		expectedImage string
		wantErr       bool
	}{
		{
			name:          "distroless image",
			imageType:     "distroless",
			expectedImage: "gcr.io/distroless/static-debian12",
			wantErr:       false,
		},
		{
			name:          "scratch image",
			imageType:     "scratch",
			expectedImage: "scratch",
			wantErr:       false,
		},
		{
			name:          "alpine image",
			imageType:     "alpine",
			expectedImage: "alpine:3.19",
			wantErr:       false,
		},
		{
			name:          "unsupported image",
			imageType:     "ubuntu",
			expectedImage: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := co.SetBaseImage(tt.imageType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported base image type")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedImage, co.GetConfig().Image)
				assert.Equal(t, tt.imageType, co.GetConfig().BaseImageType)
			}
		})
	}
}

func TestValidateSecurityConfiguration(t *testing.T) {
	auditLogger := logger.New()

	t.Run("secure configuration passes validation", func(t *testing.T) {
		co, err := NewContainerOrchestrator(auditLogger)
		require.NoError(t, err)

		err = co.ValidateSecurityConfiguration()
		assert.NoError(t, err, "Secure configuration should pass validation")
	})

	t.Run("insecure configuration fails validation", func(t *testing.T) {
		co, err := NewContainerOrchestrator(auditLogger)
		require.NoError(t, err)

		// Configure insecure settings
		config := co.GetConfig()
		config.User = "root"
		config.SeccompProfile = "unconfined"
		config.ApparmorProfile = "unconfined"
		co.SetConfig(config)

		err = co.ValidateSecurityConfiguration()
		assert.Error(t, err, "Insecure configuration should fail validation")
		assert.Contains(t, err.Error(), "running as root user poses security risk")
		assert.Contains(t, err.Error(), "seccomp disabled")
		assert.Contains(t, err.Error(), "apparmor disabled")
	})
}
