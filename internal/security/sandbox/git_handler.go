// Package sandbox provides secure Git repository cloning and sandboxing capabilities
// with comprehensive timeout, size validation, and audit logging.
package sandbox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// GitHandler manages secure Git repository operations with sandboxing
type GitHandler struct {
	CloneTimeout    time.Duration
	MaxRepoSize     int64 // in bytes
	TempDir         string
	AuditLogger     *logger.Logger
	tempDirCreated  bool
}

// GitCloneResult represents the result of a Git clone operation
type GitCloneResult struct {
	LocalPath    string
	RepoSize     int64
	CloneDuration time.Duration
	Success      bool
	Error        error
}

// NewGitHandler creates a new GitHandler with secure defaults
func NewGitHandler(auditLogger *logger.Logger) (*GitHandler, error) {
	if auditLogger == nil {
		return nil, fmt.Errorf("audit logger cannot be nil")
	}

	// Create secure temporary directory
	tempDir, err := createSecureTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create secure temp directory: %w", err)
	}

	return &GitHandler{
		CloneTimeout:   30 * time.Minute, // Default 30 minute timeout
		MaxRepoSize:    10 * 1024 * 1024 * 1024, // 10GB limit
		TempDir:        tempDir,
		AuditLogger:    auditLogger,
		tempDirCreated: true,
	}, nil
}

// createSecureTempDir creates a temporary directory with restricted permissions
func createSecureTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "repo-sandbox-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Set restrictive permissions (700 - owner read/write/execute only)
	if err := os.Chmod(tempDir, 0700); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to set temp directory permissions: %w", err)
	}

	return tempDir, nil
}

// CloneRepository clones a repository with security controls and validation
func (gh *GitHandler) CloneRepository(ctx context.Context, repoURL string) (*GitCloneResult, error) {
	startTime := time.Now()
	
	// Log the clone attempt
	gh.AuditLogger.WithFields(map[string]interface{}{
		"operation": "git_clone_start",
		"repo_url":  sanitizeURL(repoURL),
		"timestamp": startTime.Unix(),
	}).Info("Starting repository clone")

	result := &GitCloneResult{}

	// Validate repository URL length and format
	if len(repoURL) > 2048 {
		err := fmt.Errorf("repository URL exceeds maximum length of 2048 characters")
		gh.logCloneFailure(repoURL, startTime, err)
		result.Error = err
		return result, err
	}

	// Create clone-specific directory within temp directory
	cloneDir := filepath.Join(gh.TempDir, fmt.Sprintf("clone-%d", startTime.UnixNano()))
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		err = fmt.Errorf("failed to create clone directory: %w", err)
		gh.logCloneFailure(repoURL, startTime, err)
		result.Error = err
		return result, err
	}

	// Create context with timeout
	cloneCtx, cancel := context.WithTimeout(ctx, gh.CloneTimeout)
	defer cancel()

	// Perform the clone with progress tracking
	cloneResult, err := gh.performClone(cloneCtx, repoURL, cloneDir)
	if err != nil {
		gh.logCloneFailure(repoURL, startTime, err)
		result.Error = err
		return result, err
	}

	result.LocalPath = cloneResult.LocalPath
	result.RepoSize = cloneResult.RepoSize
	result.CloneDuration = time.Since(startTime)
	result.Success = true

	// Log successful clone
	gh.AuditLogger.WithFields(map[string]interface{}{
		"operation":      "git_clone_success",
		"repo_url":       sanitizeURL(repoURL),
		"local_path":     cloneResult.LocalPath,
		"repo_size":      cloneResult.RepoSize,
		"clone_duration": result.CloneDuration.Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Info("Repository clone completed successfully")

	return result, nil
}

// performClone executes the actual Git clone operation
func (gh *GitHandler) performClone(ctx context.Context, repoURL, cloneDir string) (*GitCloneResult, error) {
	// Use git clone with specific options for security
	cmd := exec.CommandContext(ctx, "git", "clone", 
		"--depth=1",           // Shallow clone to reduce size
		"--single-branch",     // Only clone the default branch
		"--no-hardlinks",      // Prevent hardlink issues
		repoURL, 
		cloneDir)

	// Set environment variables to prevent credential prompting
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=echo",
	)

	// Execute the clone command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %w, output: %s", err, string(output))
	}

	// Validate repository size after clone
	repoSize, err := calculateDirectorySize(cloneDir)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate repository size: %w", err)
	}

	if repoSize > gh.MaxRepoSize {
		// Clean up oversized repository immediately
		os.RemoveAll(cloneDir)
		return nil, fmt.Errorf("repository size (%d bytes) exceeds maximum allowed size (%d bytes)", repoSize, gh.MaxRepoSize)
	}

	return &GitCloneResult{
		LocalPath: cloneDir,
		RepoSize:  repoSize,
	}, nil
}

// calculateDirectorySize calculates the total size of a directory
func calculateDirectorySize(dirPath string) (int64, error) {
	var totalSize int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize, err
}

// logCloneFailure logs clone failure with security-conscious information
func (gh *GitHandler) logCloneFailure(repoURL string, startTime time.Time, err error) {
	gh.AuditLogger.WithFields(map[string]interface{}{
		"operation":      "git_clone_failure",
		"repo_url":       sanitizeURL(repoURL),
		"error":          err.Error(),
		"failure_duration": time.Since(startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Error("Repository clone failed")
}

// sanitizeURL removes sensitive information from URLs for logging
func sanitizeURL(url string) string {
	// Remove credentials from URLs for secure logging
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) >= 2 {
			protocolPart := strings.Split(parts[0], "://")
			if len(protocolPart) == 2 {
				return protocolPart[0] + "://[REDACTED]@" + parts[1]
			}
		}
	}
	return url
}

// Cleanup removes temporary files and directories
func (gh *GitHandler) Cleanup() error {
	if !gh.tempDirCreated || gh.TempDir == "" {
		return nil
	}

	gh.AuditLogger.WithFields(map[string]interface{}{
		"operation": "git_handler_cleanup",
		"temp_dir":  gh.TempDir,
		"timestamp": time.Now().Unix(),
	}).Info("Cleaning up GitHandler temporary directory")

	if err := os.RemoveAll(gh.TempDir); err != nil {
		gh.AuditLogger.WithFields(map[string]interface{}{
			"operation": "git_handler_cleanup_failure",
			"temp_dir":  gh.TempDir,
			"error":     err.Error(),
			"timestamp": time.Now().Unix(),
		}).Error("Failed to cleanup GitHandler temporary directory")
		return fmt.Errorf("failed to cleanup temp directory: %w", err)
	}

	gh.tempDirCreated = false
	return nil
}

// GetRepositoryInfo extracts basic information from a cloned repository
func (gh *GitHandler) GetRepositoryInfo(repoPath string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Get repository size
	size, err := calculateDirectorySize(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate repository size: %w", err)
	}
	info["size_bytes"] = size

	// Get basic Git information
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		info["is_git_repo"] = true
		
		// Get commit count (if possible)
		cmd := exec.Command("git", "-C", repoPath, "rev-list", "--count", "HEAD")
		if output, err := cmd.Output(); err == nil {
			if count, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
				info["commit_count"] = count
			}
		}
	} else {
		info["is_git_repo"] = false
	}

	return info, nil
}