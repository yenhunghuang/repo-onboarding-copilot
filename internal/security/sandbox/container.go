// Package sandbox provides secure Docker container orchestration for repository analysis
// with comprehensive resource limits, network isolation, and security controls.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	NetworkMode string // Network mode (default: "none" for isolation)
	ReadOnly    bool   // Read-only filesystem (default: true)
	NoNewPrivs  bool   // Disable privilege escalation (default: true)

	// Enhanced security settings
	UserNSMode      string // User namespace mode for isolation
	SeccompProfile  string // Seccomp profile path for syscall filtering
	ApparmorProfile string // AppArmor profile name for additional restrictions
	BaseImageType   string // Base image type: "distroless", "scratch", "alpine"

	// Container settings
	Image   string        // Base image (default: "gcr.io/distroless/static-debian12")
	WorkDir string        // Working directory inside container
	Timeout time.Duration // Container execution timeout

	// User settings
	User   string // Non-root user (default: "65534:65534" - nobody user)
	UserNS bool   // Enable user namespace isolation
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

		// Enhanced security settings
		UserNSMode:      "host",           // Will be updated to isolated mode when supported
		SeccompProfile:  "default",        // Use Docker's default seccomp profile
		ApparmorProfile: "docker-default", // Use Docker's default AppArmor profile
		BaseImageType:   "distroless",

		// Distroless static image for minimal attack surface
		Image:   "gcr.io/distroless/static-debian12",
		WorkDir: "/workspace",
		Timeout: 1 * time.Hour, // 1-hour execution limit
		User:    "65534:65534", // nobody user for enhanced security
		UserNS:  true,          // Enable user namespace isolation
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
		// Categorize and provide helpful error messages
		errorCategory, helpMessage := co.categorizeContainerError(string(output), err)

		co.auditLogger.WithFields(map[string]interface{}{
			"operation":      "container_creation_failure",
			"error":          err.Error(),
			"docker_output":  sanitizeDockerOutput(string(output)),
			"error_category": errorCategory,
			"help_message":   helpMessage,
			"docker_args":    strings.Join(args, " "),
			"execution_time": time.Since(startTime).Seconds(),
			"timestamp":      time.Now().Unix(),
		}).Error("Failed to create secure container")

		return "", fmt.Errorf("failed to create container (%s): %w\nOutput: %s\nSuggestion: %s",
			errorCategory, err, string(output), helpMessage)
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
		"--detach", // Run in background
		"--rm",     // Auto-remove when stopped
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

	// Add enhanced security options
	args = append(args,
		"--cap-drop", "ALL", // Drop all capabilities
		"--cap-add", "DAC_OVERRIDE", // Allow file access override
		"--pids-limit", "100", // Limit number of processes
		"--ulimit", "nofile=1024:1024", // Limit file descriptors
	)

	// Add seccomp profile with environment-aware configuration
	if co.config.SeccompProfile != "unconfined" {
		seccompArg := co.resolveSeccompProfile()
		if seccompArg != "" {
			args = append(args, "--security-opt", seccompArg)
		}
	}

	// Add AppArmor profile
	if co.config.ApparmorProfile != "unconfined" {
		args = append(args, "--security-opt", fmt.Sprintf("apparmor=%s", co.config.ApparmorProfile))
	}

	// Add user namespace isolation if enabled
	if co.config.UserNS && co.config.UserNSMode != "host" {
		args = append(args, "--userns", co.config.UserNSMode)
	}

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

// SetBaseImage configures the appropriate base image based on security requirements
func (co *ContainerOrchestrator) SetBaseImage(imageType string) error {
	switch imageType {
	case "distroless":
		co.config.Image = "gcr.io/distroless/static-debian12"
		co.config.BaseImageType = "distroless"
	case "scratch":
		co.config.Image = "scratch"
		co.config.BaseImageType = "scratch"
	case "alpine":
		co.config.Image = "alpine:3.19"
		co.config.BaseImageType = "alpine"
	default:
		return fmt.Errorf("unsupported base image type: %s", imageType)
	}

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":  "base_image_configuration",
		"image_type": imageType,
		"image":      co.config.Image,
		"timestamp":  time.Now().Unix(),
	}).Info("Base image configured")

	return nil
}

// ValidateSecurityConfiguration validates the current security configuration
func (co *ContainerOrchestrator) ValidateSecurityConfiguration() error {
	issues := make([]string, 0)

	// Validate base image security
	if co.config.BaseImageType == "alpine" {
		co.auditLogger.WithFields(map[string]interface{}{
			"operation": "security_validation_warning",
			"issue":     "alpine_base_image_not_minimal",
			"timestamp": time.Now().Unix(),
		}).Warn("Using Alpine base image - consider distroless for enhanced security")
	}

	// Validate user configuration
	if co.config.User == "root" || co.config.User == "0:0" {
		issues = append(issues, "running as root user poses security risk")
	}

	// Validate security options
	if co.config.SeccompProfile == "unconfined" {
		issues = append(issues, "seccomp disabled - syscall filtering not active")
	}

	if co.config.ApparmorProfile == "unconfined" {
		issues = append(issues, "apparmor disabled - mandatory access controls not active")
	}

	if len(issues) > 0 {
		co.auditLogger.WithFields(map[string]interface{}{
			"operation":       "security_validation_issues",
			"security_issues": issues,
			"timestamp":       time.Now().Unix(),
		}).Error("Security configuration validation failed")

		return fmt.Errorf("security configuration issues found: %v", issues)
	}

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":        "security_validation_success",
		"base_image_type":  co.config.BaseImageType,
		"seccomp_profile":  co.config.SeccompProfile,
		"apparmor_profile": co.config.ApparmorProfile,
		"user_namespace":   co.config.UserNS,
		"timestamp":        time.Now().Unix(),
	}).Info("Security configuration validation passed")

	return nil
}

// resolveSeccompProfile resolves the appropriate seccomp profile based on environment
func (co *ContainerOrchestrator) resolveSeccompProfile() string {
	if co.config.SeccompProfile == "default" {
		// In CI/DinD environments, use explicit profile path
		if co.isDockerInDocker() {
			profilePath := co.findSeccompProfile()
			if profilePath != "" {
				co.auditLogger.WithFields(map[string]interface{}{
					"operation":    "seccomp_profile_resolution",
					"environment":  "docker-in-docker",
					"profile_path": profilePath,
					"timestamp":    time.Now().Unix(),
				}).Info("Using explicit seccomp profile for DinD environment")
				return fmt.Sprintf("seccomp=%s", profilePath)
			}

			// Fallback: disable seccomp in DinD if profile not found
			co.auditLogger.WithFields(map[string]interface{}{
				"operation":   "seccomp_profile_fallback",
				"environment": "docker-in-docker",
				"action":      "unconfined_fallback",
				"timestamp":   time.Now().Unix(),
			}).Warn("Seccomp profile not found in DinD - falling back to unconfined mode")
			return "seccomp=unconfined"
		}

		// In native Docker environments, use default
		return "seccomp=default"
	}

	// Use explicit profile path
	return fmt.Sprintf("seccomp=%s", co.config.SeccompProfile)
}

// isDockerInDocker detects if running in Docker-in-Docker environment
func (co *ContainerOrchestrator) isDockerInDocker() bool {
	// Check for CI environment variables
	if os.Getenv("GITHUB_ACTIONS") == "true" ||
		os.Getenv("CI") == "true" ||
		os.Getenv("DOCKER_HOST") != "" {
		return true
	}

	// Check if we're running inside a container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check for DinD-specific indicators
	if os.Getenv("DOCKER_TLS_VERIFY") == "1" {
		return true
	}

	return false
}

// findSeccompProfile locates available seccomp profiles
func (co *ContainerOrchestrator) findSeccompProfile() string {
	// Define potential profile locations
	candidatePaths := []string{
		"./configs/security/seccomp-analysis.json",          // Local development
		"/workspace/configs/security/seccomp-analysis.json", // Container workspace
		"/app/configs/security/seccomp-analysis.json",       // App directory
		"/tmp/seccomp-profile.json",                         // Temp location for CI
	}

	for _, path := range candidatePaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				co.auditLogger.WithFields(map[string]interface{}{
					"operation":    "seccomp_profile_found",
					"profile_path": absPath,
					"timestamp":    time.Now().Unix(),
				}).Info("Found seccomp profile")
				return absPath
			}
		}
	}

	co.auditLogger.WithFields(map[string]interface{}{
		"operation":      "seccomp_profile_search",
		"searched_paths": candidatePaths,
		"timestamp":      time.Now().Unix(),
	}).Warn("No seccomp profile found in expected locations")

	return ""
}

// categorizeContainerError analyzes container creation errors and provides helpful messages
func (co *ContainerOrchestrator) categorizeContainerError(output string, err error) (category string, helpMessage string) {
	outputLower := strings.ToLower(output)
	errLower := strings.ToLower(err.Error())

	// Seccomp profile errors
	if strings.Contains(outputLower, "seccomp profile") && strings.Contains(outputLower, "failed") {
		return "seccomp_profile_error",
			"Seccomp profile not found. In DinD environments, ensure the profile is accessible or disable seccomp temporarily."
	}

	// Docker daemon connectivity issues
	if strings.Contains(errLower, "cannot connect to the docker daemon") ||
		strings.Contains(outputLower, "connection refused") {
		return "docker_daemon_error",
			"Docker daemon is not accessible. Verify Docker is running and accessible at " + os.Getenv("DOCKER_HOST")
	}

	// Image pull errors
	if strings.Contains(outputLower, "unable to find image") ||
		strings.Contains(outputLower, "pull access denied") {
		return "image_pull_error",
			"Container image is not available. Try: docker pull " + co.config.Image
	}

	// Resource limit errors
	if strings.Contains(outputLower, "memory") && strings.Contains(outputLower, "limit") {
		return "resource_limit_error",
			"Memory limit too restrictive. Current limit: " + fmt.Sprintf("%dGB", co.config.MemoryLimitGB)
	}

	// Permission/privilege errors
	if strings.Contains(outputLower, "permission denied") ||
		strings.Contains(outputLower, "operation not permitted") {
		return "permission_error",
			"Docker permissions issue. In CI: ensure privileged mode. Local: check Docker group membership."
	}

	// Network isolation issues
	if strings.Contains(outputLower, "network") && strings.Contains(outputLower, "none") {
		return "network_isolation_error",
			"Network isolation mode 'none' may not be supported in this environment."
	}

	// AppArmor profile errors
	if strings.Contains(outputLower, "apparmor") {
		return "apparmor_profile_error",
			"AppArmor profile not available. This is common in containers - profile will be disabled."
	}

	// User namespace errors
	if strings.Contains(outputLower, "user namespace") || strings.Contains(outputLower, "userns") {
		return "user_namespace_error",
			"User namespace isolation not supported in this Docker environment."
	}

	// Generic exit code errors
	if strings.Contains(errLower, "exit status 125") {
		return "docker_argument_error",
			"Invalid Docker arguments. Check container configuration and Docker version compatibility."
	}

	// Default case
	return "unknown_error",
		"Unexpected container creation failure. Check Docker daemon status and container configuration."
}
