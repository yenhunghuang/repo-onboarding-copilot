package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLValidator_ValidateRepositoryURL(t *testing.T) {
	validator := New()

	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedURL struct {
			scheme string
			host   string
			path   string
		}
	}{
		// Valid HTTPS URLs
		{
			name:        "valid https github url",
			input:       "https://github.com/owner/repo.git",
			expectError: false,
			expectedURL: struct {
				scheme string
				host   string
				path   string
			}{
				scheme: "https",
				host:   "github.com",
				path:   "/owner/repo.git",
			},
		},
		{
			name:        "valid https gitlab url",
			input:       "https://gitlab.com/owner/repo",
			expectError: false,
			expectedURL: struct {
				scheme string
				host   string
				path   string
			}{
				scheme: "https",
				host:   "gitlab.com",
				path:   "/owner/repo",
			},
		},
		// Valid SSH URLs
		{
			name:        "valid ssh github url",
			input:       "git@github.com:owner/repo.git",
			expectError: false,
			expectedURL: struct {
				scheme string
				host   string
				path   string
			}{
				scheme: "ssh",
				host:   "github.com",
				path:   "owner/repo.git",
			},
		},
		{
			name:        "valid ssh gitlab url",
			input:       "git@gitlab.com:owner/repo",
			expectError: false,
			expectedURL: struct {
				scheme string
				host   string
				path   string
			}{
				scheme: "ssh",
				host:   "gitlab.com",
				path:   "owner/repo",
			},
		},
		// Valid Git URLs
		{
			name:        "valid git url",
			input:       "git://git.example.com/owner/repo.git",
			expectError: false,
			expectedURL: struct {
				scheme string
				host   string
				path   string
			}{
				scheme: "git",
				host:   "git.example.com",
				path:   "/owner/repo.git",
			},
		},
		// Invalid URLs - malicious inputs
		{
			name:        "javascript injection",
			input:       "javascript:alert('xss')",
			expectError: true,
		},
		{
			name:        "data url",
			input:       "data:text/plain,malicious",
			expectError: true,
		},
		{
			name:        "file url",
			input:       "file:///etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal",
			input:       "https://github.com/../../../etc/passwd",
			expectError: true,
		},
		// Invalid URLs - format issues
		{
			name:        "empty url",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   \t  \n  ",
			expectError: true,
		},
		{
			name:        "unsupported scheme",
			input:       "ftp://example.com/repo",
			expectError: true,
		},
		{
			name:        "no host",
			input:       "https:///repo",
			expectError: true,
		},
		{
			name:        "invalid ssh format",
			input:       "git@github.com/owner/repo",
			expectError: true,
		},
		// Length limits
		{
			name:        "url too long",
			input:       "https://github.com/" + strings.Repeat("a", 3000),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateRepositoryURL(tt.input)

			if tt.expectError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err, "Unexpected error for input: %s", tt.input)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedURL.scheme, result.Scheme)
				assert.Equal(t, tt.expectedURL.host, result.Host)
				assert.Equal(t, tt.expectedURL.path, result.Path)
			}
		})
	}
}