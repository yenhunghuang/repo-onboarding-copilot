// Package types defines shared type definitions used across the application.
// It provides common data structures and interfaces for clean architecture.
package types

// RepositoryURL represents a validated repository URL
type RepositoryURL struct {
	Raw    string `json:"raw"`
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Path   string `json:"path"`
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return ve.Message
}