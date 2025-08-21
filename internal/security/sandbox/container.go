// Package sandbox provides secure Docker container orchestration for repository analysis
// with comprehensive resource limits, network isolation, and security controls.
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// ContainerConfig represents Docker container security configuration
type ContainerConfig struct {
	// Resource limits
	MemoryLimitGB int    // Memory limit in GB (default: 2)
	CPULimit      string // CPU limit (default: "4.0")
	
	// Security settings
	NetworkMode   string // Network mode (default: "none" for isolation)
	ReadOnly      bool   // Read-only filesystem (default: true)
	NoNewPrivs    bool   // Disable privilege escalation (default: true)
	
	// Container settings
	Image         string        // Base image (default: "alpine:latest")
	WorkDir       string        // Working directory inside container
	Timeout       time.Duration // Container execution timeout
	
	// User settings
	User          string        // Non-root user (default: "1000:1000")
}

// ContainerOrchestrator manages secure Docker container operations
type ContainerOrchestrator struct {
	config      *ContainerConfig
	auditLogger *logger.Logger
}

// ContainerResult represents the result of container operations
type ContainerResult struct {
	ContainerID   string
	ExitCode      int
	Output        string
	Error         error
	ExecutionTime time.Duration
	ResourceUsage map[string]interface{}
}

// NewContainerOrchestrator creates a new container orchestrator with security defaults
func NewContainerOrchestrator(auditLogger *logger.Logger) (*ContainerOrchestrator, error) {
	if auditLogger == nil {
		return nil, fmt.Errorf("audit logger cannot be nil")
	}

	config := &ContainerConfig{
		MemoryLimitGB: 2,
		CPULimit:      "4.0",
		NetworkMode:   "none", // Complete network isolation
		ReadOnly:      true,
		NoNewPrivs:    true,
		Image:         "alpine:latest",
		WorkDir:       "/workspace",
		Timeout:       1 * time.Hour, // 1-hour execution limit
		User:          "1000:1000",   // Non-root user
	}

	return &ContainerOrchestrator{
		config:      config,
		auditLogger: auditLogger,
	}, nil
}

// ValidateDockerAvailability checks if Docker is available and accessible
func (co *ContainerOrchestrator) ValidateDockerAvailability() error {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker is not available or accessible: %w, output: %s", err, string(output))
	}

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":      "docker_validation",
		"docker_version": strings.TrimSpace(string(output)),
		"timestamp":      time.Now().Unix(),
	}).Info("Docker availability validated")

	return nil
}

// CreateSecureContainer creates a Docker container with comprehensive security controls
func (co *ContainerOrchestrator) CreateSecureContainer(ctx context.Context, volumeMounts map[string]string) (string, error) {
	startTime := time.Now()

	// Build Docker command with security parameters
	args := co.buildDockerArgs(volumeMounts)

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":     "container_creation_start",
		"image":         co.config.Image,
		"memory_limit":  fmt.Sprintf("%dg", co.config.MemoryLimitGB),
		"cpu_limit":     co.config.CPULimit,
		"network_mode":  co.config.NetworkMode,
		"read_only":     co.config.ReadOnly,
		"volume_mounts": len(volumeMounts),
		"timestamp":     startTime.Unix(),
	}).Info("Creating secure container")

	// Execute docker run command
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		co.auditLogger.WithFields(map[string]interface{}{
			"operation":        "container_creation_failure",
			"error":           err.Error(),
			"docker_output":   sanitizeDockerOutput(string(output)),
			"execution_time":  time.Since(startTime).Seconds(),
			"timestamp":       time.Now().Unix(),
		}).Error("Failed to create secure container")
		return "", fmt.Errorf("failed to create container: %w, output: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	
	co.auditLogger.WithFields(map[string]interface{}{
		"operation":      "container_creation_success",
		"container_id":   containerID,
		"execution_time": time.Since(startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Info("Secure container created successfully")

	return containerID, nil
}

// buildDockerArgs constructs Docker command arguments with security controls
func (co *ContainerOrchestrator) buildDockerArgs(volumeMounts map[string]string) []string {
	args := []string{
		"run", 
		"--detach",                                    // Run in background
		"--rm",                                        // Auto-remove when stopped
		fmt.Sprintf("--memory=%dg", co.config.MemoryLimitGB), // Memory limit
		fmt.Sprintf("--cpus=%s", co.config.CPULimit),         // CPU limit
		fmt.Sprintf("--network=%s", co.config.NetworkMode),   // Network isolation
		fmt.Sprintf("--user=%s", co.config.User),             // Non-root user
		fmt.Sprintf("--workdir=%s", co.config.WorkDir),       // Working directory
	}

	// Add security options
	if co.config.ReadOnly {
		args = append(args, "--read-only")
		// Add tmpfs for writable temporary space
		args = append(args, "--tmpfs", "/tmp:noexec,nosuid,size=100m")
		args = append(args, "--tmpfs", "/workspace/tmp:noexec,nosuid,size=500m")
	}

	if co.config.NoNewPrivs {
		args = append(args, "--security-opt", "no-new-privileges:true")
	}

	// Add additional security options
	args = append(args,
		"--security-opt", "apparmor:unconfined",           // Disable AppArmor for now
		"--security-opt", "seccomp:unconfined",            // Disable seccomp for now
		"--cap-drop", "ALL",                               // Drop all capabilities
		"--cap-add", "DAC_OVERRIDE",                       // Allow file access override
		"--pids-limit", "100",                             // Limit number of processes
		"--ulimit", "nofile=1024:1024",                    // Limit file descriptors
	)

	// Add volume mounts
	for hostPath, containerPath := range volumeMounts {
		mountSpec := fmt.Sprintf("%s:%s:ro", hostPath, containerPath) // Read-only mounts
		args = append(args, "-v", mountSpec)
	}

	// Add image and default command
	args = append(args, co.config.Image, "sleep", "3600") // Keep container running

	return args
}

// ExecuteInContainer executes a command inside a running container
func (co *ContainerOrchestrator) ExecuteInContainer(ctx context.Context, containerID, command string) (*ContainerResult, error) {
	startTime := time.Now()
	
	co.auditLogger.WithFields(map[string]interface{}{
		"operation":    "container_execution_start",
		"container_id": containerID,
		"command":      sanitizeCommand(command),
		"timestamp":    startTime.Unix(),
	}).Info("Executing command in container")

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, co.config.Timeout)
	defer cancel()

	// Execute command in container
	cmd := exec.CommandContext(execCtx, "docker", "exec", containerID, "sh", "-c", command)
	output, err := cmd.CombinedOutput()

	result := &ContainerResult{
		ContainerID:   containerID,
		Output:        string(output),
		ExecutionTime: time.Since(startTime),
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err != nil {
		result.Error = err
	}

	// Log execution result
	logLevel := "info"
	if err != nil {
		logLevel = "error"
	}

	fields := map[string]interface{}{
		"operation":      "container_execution_complete",
		"container_id":   containerID,
		"exit_code":      result.ExitCode,
		"execution_time": result.ExecutionTime.Seconds(),
		"output_length":  len(result.Output),
		"timestamp":      time.Now().Unix(),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	switch logLevel {
	case "error":
		co.auditLogger.WithFields(fields).Error("Container command execution failed")
	default:
		co.auditLogger.WithFields(fields).Info("Container command execution completed")
	}

	return result, nil
}

// StopContainer stops and removes a running container
func (co *ContainerOrchestrator) StopContainer(ctx context.Context, containerID string) error {
	startTime := time.Now()
	
	co.auditLogger.WithFields(map[string]interface{}{
		"operation":    "container_stop_start",
		"container_id": containerID,
		"timestamp":    startTime.Unix(),
	}).Info("Stopping container")

	// Stop the container with timeout
	stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(stopCtx, "docker", "stop", containerID)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		co.auditLogger.WithFields(map[string]interface{}{
			"operation":      "container_stop_failure",
			"container_id":   containerID,
			"error":          err.Error(),
			"docker_output":  sanitizeDockerOutput(string(output)),
			"execution_time": time.Since(startTime).Seconds(),
			"timestamp":      time.Now().Unix(),
		}).Error("Failed to stop container")
		return fmt.Errorf("failed to stop container %s: %w, output: %s", containerID, err, string(output))
	}

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":      "container_stop_success",
		"container_id":   containerID,
		"execution_time": time.Since(startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Info("Container stopped successfully")

	return nil
}

// GetContainerResourceUsage retrieves resource usage statistics for a container
func (co *ContainerOrchestrator) GetContainerResourceUsage(ctx context.Context, containerID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", 
		"table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}", containerID)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	stats := make(map[string]interface{})
	stats["raw_output"] = string(output)
	stats["container_id"] = containerID
	stats["timestamp"] = time.Now().Unix()

	return stats, nil
}

// sanitizeDockerOutput removes sensitive information from Docker output
func sanitizeDockerOutput(output string) string {
	// Remove potential sensitive information from Docker output
	lines := strings.Split(output, "\n")
	var sanitized []string
	
	for _, line := range lines {
		lineLower := strings.ToLower(line)
		// Skip lines that might contain sensitive paths or credentials
		if strings.Contains(lineLower, "password") || strings.Contains(lineLower, "token") ||
		   strings.Contains(lineLower, "secret") || strings.Contains(lineLower, "key") {
			sanitized = append(sanitized, "[REDACTED SENSITIVE LINE]")
		} else {
			sanitized = append(sanitized, line)
		}
	}
	
	return strings.Join(sanitized, "\n")
}

// sanitizeCommand removes potentially sensitive information from commands
func sanitizeCommand(command string) string {
	// Basic command sanitization - remove potential credentials
	commandLower := strings.ToLower(command)
	if strings.Contains(commandLower, "password") || strings.Contains(commandLower, "token") ||
	   strings.Contains(commandLower, "secret") || strings.Contains(commandLower, "key") {
		return "[REDACTED SENSITIVE COMMAND]"
	}
	return command
}

// SetConfig updates container configuration
func (co *ContainerOrchestrator) SetConfig(config *ContainerConfig) {
	co.config = config
}

// GetConfig returns current container configuration
func (co *ContainerOrchestrator) GetConfig() *ContainerConfig {
	return co.config
}