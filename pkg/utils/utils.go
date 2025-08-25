// Package utils provides utility functions used across the application.
// It contains helper functions for common operations and data manipulation.
package utils

import (
	"fmt"
	"strings"
)

// StringInSlice checks if a string exists in a slice of strings
func StringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// TrimWhitespace removes leading and trailing whitespace from strings
func TrimWhitespace(s string) string {
	return strings.TrimSpace(s)
}

// FormatError wraps an error with additional context
func FormatError(operation string, err error) error {
	return fmt.Errorf("%s: %w", operation, err)
}
