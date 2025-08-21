package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

func TestNewGitHandler(t *testing.T) {
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
			gh, err := NewGitHandler(tt.auditLogger)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, gh)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gh)
				assert.NotEmpty(t, gh.TempDir)
				assert.Equal(t, 30*time.Minute, gh.CloneTimeout)
				assert.Equal(t, int64(10*1024*1024*1024), gh.MaxRepoSize)
				assert.True(t, gh.tempDirCreated)
				
				// Verify temp directory exists and has correct permissions
				stat, err := os.Stat(gh.TempDir)
				assert.NoError(t, err)
				assert.True(t, stat.IsDir())
				assert.Equal(t, os.FileMode(0700), stat.Mode().Perm())
				
				// Cleanup
				err = gh.Cleanup()
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateSecureTempDir(t *testing.T) {
	tempDir, err := createSecureTempDir()
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Verify directory exists
	stat, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, stat.IsDir())

	// Verify permissions are restrictive (700)
	assert.Equal(t, os.FileMode(0700), stat.Mode().Perm())

	// Verify directory name pattern
	assert.Contains(t, tempDir, "repo-sandbox-")
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "https with credentials",
			input:    "https://user:pass@github.com/owner/repo.git",
			expected: "https://[REDACTED]@github.com/owner/repo.git",
		},
		{
			name:     "http with credentials",
			input:    "http://user:token@gitlab.com/owner/repo.git",
			expected: "http://[REDACTED]@gitlab.com/owner/repo.git",
		},
		{
			name:     "no credentials",
			input:    "https://github.com/owner/repo.git",
			expected: "https://github.com/owner/repo.git",
		},
		{
			name:     "ssh url",
			input:    "git@github.com:owner/repo.git",
			expected: "git@github.com:owner/repo.git",
		},
		{
			name:     "malformed url",
			input:    "not-a-url",
			expected: "not-a-url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDirectorySize(t *testing.T) {
	// Create test directory structure
	tempDir, err := os.MkdirTemp("", "test-size-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files with known sizes
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "subdir", "file2.txt")
	
	require.NoError(t, os.MkdirAll(filepath.Dir(file2), 0755))
	require.NoError(t, os.WriteFile(file1, []byte("hello"), 0644)) // 5 bytes
	require.NoError(t, os.WriteFile(file2, []byte("world!"), 0644)) // 6 bytes

	size, err := calculateDirectorySize(tempDir)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), size) // 5 + 6 bytes
}

func TestCalculateDirectorySizeNonExistent(t *testing.T) {
	size, err := calculateDirectorySize("/non/existent/directory")
	assert.Error(t, err)
	assert.Equal(t, int64(0), size)
}

func TestGitHandlerCleanup(t *testing.T) {
	auditLogger := logger.New()
	gh, err := NewGitHandler(auditLogger)
	require.NoError(t, err)

	tempDir := gh.TempDir
	
	// Verify directory exists
	_, err = os.Stat(tempDir)
	assert.NoError(t, err)

	// Cleanup
	err = gh.Cleanup()
	assert.NoError(t, err)

	// Verify directory is removed
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err))

	// Verify cleanup flag is updated
	assert.False(t, gh.tempDirCreated)

	// Second cleanup should be no-op
	err = gh.Cleanup()
	assert.NoError(t, err)
}

func TestCloneRepositoryValidation(t *testing.T) {
	auditLogger := logger.New()
	gh, err := NewGitHandler(auditLogger)
	require.NoError(t, err)
	defer gh.Cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		repoURL string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "url too long",
			repoURL: string(make([]byte, 2049)), // Exceed 2048 char limit
			wantErr: true,
			errMsg:  "repository URL exceeds maximum length",
		},
		{
			name:    "empty url",
			repoURL: "",
			wantErr: true, // Will fail at git level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gh.CloneRepository(ctx, tt.repoURL)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.NotNil(t, result)
				assert.False(t, result.Success)
			}
		})
	}
}

func TestCloneRepositoryTimeout(t *testing.T) {
	auditLogger := logger.New()
	gh, err := NewGitHandler(auditLogger)
	require.NoError(t, err)
	defer gh.Cleanup()

	// Set very short timeout for testing
	gh.CloneTimeout = 1 * time.Nanosecond

	ctx := context.Background()
	result, err := gh.CloneRepository(ctx, "https://github.com/octocat/Hello-World.git")
	
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}

func TestGetRepositoryInfo(t *testing.T) {
	auditLogger := logger.New()
	gh, err := NewGitHandler(auditLogger)
	require.NoError(t, err)
	defer gh.Cleanup()

	// Create test directory structure
	testDir := filepath.Join(gh.TempDir, "test-repo")
	require.NoError(t, os.MkdirAll(testDir, 0755))
	
	// Create a test file
	testFile := filepath.Join(testDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	// Test non-git directory
	info, err := gh.GetRepositoryInfo(testDir)
	assert.NoError(t, err)
	assert.Equal(t, false, info["is_git_repo"])
	assert.Equal(t, int64(12), info["size_bytes"]) // "test content" = 12 bytes

	// Test non-existent directory
	info, err = gh.GetRepositoryInfo("/non/existent")
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestGitHandlerDefaults(t *testing.T) {
	auditLogger := logger.New()
	gh, err := NewGitHandler(auditLogger)
	require.NoError(t, err)
	defer gh.Cleanup()

	// Verify default values
	assert.Equal(t, 30*time.Minute, gh.CloneTimeout)
	assert.Equal(t, int64(10*1024*1024*1024), gh.MaxRepoSize) // 10GB
	assert.NotEmpty(t, gh.TempDir)
	assert.NotNil(t, gh.AuditLogger)
	assert.True(t, gh.tempDirCreated)
}