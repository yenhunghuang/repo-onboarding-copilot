// Package validator provides input validation and sanitization for security purposes.
// It implements validation for repository URLs and prevents malicious inputs.
package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/types"
)

// URLValidator validates repository URLs
type URLValidator struct {
	allowedSchemes []string
	maxLength      int
}

// New creates a new URL validator with default settings
func New() *URLValidator {
	return &URLValidator{
		allowedSchemes: []string{"http", "https", "git", "ssh"},
		maxLength:      2048,
	}
}

// ValidateRepositoryURL validates a repository URL for security and format
func (v *URLValidator) ValidateRepositoryURL(rawURL string) (*types.RepositoryURL, error) {
	// Input sanitization
	sanitizedURL := v.sanitizeInput(rawURL)

	// Length check
	if len(sanitizedURL) > v.maxLength {
		return nil, fmt.Errorf("URL exceeds maximum length of %d characters", v.maxLength)
	}

	// Malicious input detection
	if err := v.detectMaliciousInput(sanitizedURL); err != nil {
		return nil, err
	}

	// Handle SSH URLs (git@host:path format)
	if v.isSSHURL(sanitizedURL) {
		return v.parseSSHURL(sanitizedURL)
	}

	// Parse standard URLs
	parsedURL, err := url.Parse(sanitizedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme
	if !v.isAllowedScheme(parsedURL.Scheme) {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Validate host
	if parsedURL.Host == "" {
		return nil, fmt.Errorf("URL must contain a valid host")
	}

	// Validate Git repository path patterns
	if err := v.validateGitPath(parsedURL.Path); err != nil {
		return nil, err
	}

	return &types.RepositoryURL{
		Raw:    sanitizedURL,
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   parsedURL.Path,
	}, nil
}

// sanitizeInput removes potentially dangerous characters and normalizes input
func (v *URLValidator) sanitizeInput(input string) string {
	// Remove leading/trailing whitespace
	cleaned := strings.TrimSpace(input)

	// Remove null bytes and control characters
	cleaned = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(cleaned, "")

	return cleaned
}

// detectMaliciousInput checks for potentially malicious patterns
func (v *URLValidator) detectMaliciousInput(input string) error {
	// Check for common malicious patterns
	maliciousPatterns := []string{
		`javascript:`,
		`data:`,
		`file:`,
		`ftp:`,
		`../`,
		`..\\`,
		`<script`,
		`${`,
		`$(`,
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return fmt.Errorf("potentially malicious input detected")
		}
	}

	return nil
}

// isSSHURL checks if the URL is in SSH format (git@host:path)
func (v *URLValidator) isSSHURL(urlStr string) bool {
	sshPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9._/-]+$`)
	return sshPattern.MatchString(urlStr)
}

// parseSSHURL parses SSH format URLs
func (v *URLValidator) parseSSHURL(urlStr string) (*types.RepositoryURL, error) {
	parts := strings.SplitN(urlStr, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid SSH URL format")
	}

	hostPath := strings.SplitN(parts[1], ":", 2)
	if len(hostPath) != 2 {
		return nil, fmt.Errorf("invalid SSH URL format")
	}

	return &types.RepositoryURL{
		Raw:    urlStr,
		Scheme: "ssh",
		Host:   hostPath[0],
		Path:   hostPath[1],
	}, nil
}

// isAllowedScheme checks if the URL scheme is in the allowed list
func (v *URLValidator) isAllowedScheme(scheme string) bool {
	for _, allowed := range v.allowedSchemes {
		if strings.ToLower(scheme) == allowed {
			return true
		}
	}
	return false
}

// validateGitPath validates that the path looks like a Git repository
func (v *URLValidator) validateGitPath(path string) error {
	if path == "" {
		return fmt.Errorf("repository path cannot be empty")
	}

	// Check for common Git repository patterns
	gitPatterns := []string{
		`.git$`,
		`.git/$`,
		`^/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+`,
	}

	for _, pattern := range gitPatterns {
		matched, _ := regexp.MatchString(pattern, path)
		if matched {
			return nil
		}
	}

	// Allow paths that look like repository paths (owner/repo format)
	repoPattern := regexp.MustCompile(`^/?[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/?`)
	if repoPattern.MatchString(path) {
		return nil
	}

	return fmt.Errorf("path does not appear to be a Git repository")
}
