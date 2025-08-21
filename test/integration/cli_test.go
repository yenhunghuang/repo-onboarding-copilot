package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Build the CLI binary for testing
	err := buildCLI()
	if err != nil {
		panic("Failed to build CLI for testing: " + err.Error())
	}
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	cleanupCLI()
	
	os.Exit(code)
}

func TestCLI_Help(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "help flag",
			args:     []string{"--help"},
			expected: []string{"Repository onboarding analysis tool", "Usage:", "Examples:"},
		},
		{
			name:     "no arguments shows help",
			args:     []string{},
			expected: []string{"Repository onboarding analysis tool", "Usage:"},
		},
		{
			name:     "help command",
			args:     []string{"help"},
			expected: []string{"Repository onboarding analysis tool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(tt.args...)
			
			// Help command should exit with code 0
			assert.NoError(t, err)
			
			for _, expected := range tt.expected {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCLI_Version(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version flag",
			args: []string{"--version"},
		},
		{
			name: "version command",
			args: []string{"version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(tt.args...)
			
			assert.NoError(t, err)
			assert.Contains(t, output, "Repo Onboarding Copilot")
		})
	}
}

func TestCLI_ValidRepositoryURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedMsg []string
	}{
		{
			name:        "valid https github url",
			url:         "https://github.com/owner/repo.git",
			expectedMsg: []string{"Repository URL validated successfully", "Scheme: https, Host: github.com"},
		},
		{
			name:        "valid ssh github url",
			url:         "git@github.com:owner/repo.git",
			expectedMsg: []string{"Repository URL validated successfully", "Scheme: ssh, Host: github.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(tt.url)
			
			assert.NoError(t, err)
			for _, msg := range tt.expectedMsg {
				assert.Contains(t, output, msg)
			}
		})
	}
}

func TestCLI_InvalidRepositoryURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "malicious javascript url",
			url:         "javascript:alert('xss')",
			expectError: true,
		},
		{
			name:        "file url",
			url:         "file:///etc/passwd",
			expectError: true,
		},
		{
			name:        "invalid format",
			url:         "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(tt.url)
			
			if tt.expectError {
				// CLI should exit with non-zero code for invalid URLs
				var exitErr *exec.ExitError
				require.ErrorAs(t, err, &exitErr)
				assert.NotEqual(t, 0, exitErr.ExitCode())
				assert.Contains(t, output, "Invalid repository URL")
			}
		})
	}
}

// Helper functions

func buildCLI() error {
	cmd := exec.Command("go", "build", "-o", "./test-cli", "./cmd")
	cmd.Dir = "../.."
	return cmd.Run()
}

func cleanupCLI() {
	os.Remove("../../test-cli")
}

func runCLI(args ...string) (string, error) {
	cmd := exec.Command("../../test-cli", args...)
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}