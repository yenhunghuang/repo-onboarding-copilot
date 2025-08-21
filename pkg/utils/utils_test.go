package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		slice    []string
		expected bool
	}{
		{
			name:     "string exists in slice",
			str:      "test",
			slice:    []string{"one", "test", "three"},
			expected: true,
		},
		{
			name:     "string does not exist in slice",
			str:      "missing",
			slice:    []string{"one", "two", "three"},
			expected: false,
		},
		{
			name:     "empty slice",
			str:      "test",
			slice:    []string{},
			expected: false,
		},
		{
			name:     "empty string in slice",
			str:      "",
			slice:    []string{"", "test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringInSlice(tt.str, tt.slice)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trim whitespace",
			input:    "  https://github.com/owner/repo  ",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "no change needed",
			input:    "https://github.com/owner/repo",
			expected: "https://github.com/owner/repo",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \t  \n  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatError(t *testing.T) {
	originalErr := errors.New("original error")
	formattedErr := FormatError("operation failed", originalErr)
	
	assert.Error(t, formattedErr)
	assert.Contains(t, formattedErr.Error(), "operation failed")
	assert.Contains(t, formattedErr.Error(), "original error")
	
	// Test that it wraps the original error
	assert.ErrorIs(t, formattedErr, originalErr)
}