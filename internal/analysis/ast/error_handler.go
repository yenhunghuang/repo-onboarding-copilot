package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ErrorHandler manages error handling and graceful degradation for AST parsing
type ErrorHandler struct {
	config ErrorConfig
	stats  ErrorStats
}

// ErrorConfig configures error handling behavior
type ErrorConfig struct {
	MaxErrors          int     `json:"max_errors"`           // Maximum errors before stopping
	ErrorThreshold     float64 `json:"error_threshold"`      // Error rate threshold (0.0-1.0)
	EnableRecovery     bool    `json:"enable_recovery"`      // Enable error recovery attempts
	EnablePartialParse bool    `json:"enable_partial_parse"` // Allow partial parsing results
	LogLevel           string  `json:"log_level"`            // error, warn, info, debug
}

// ErrorStats tracks error statistics during parsing
type ErrorStats struct {
	TotalFiles       int            `json:"total_files"`
	SuccessfulFiles  int            `json:"successful_files"`
	FailedFiles      int            `json:"failed_files"`
	PartialFiles     int            `json:"partial_files"`
	TotalErrors      int            `json:"total_errors"`
	ErrorRate        float64        `json:"error_rate"`
	ErrorTypes       map[string]int `json:"error_types"`
	RecoveryAttempts int            `json:"recovery_attempts"`
	RecoverySuccess  int            `json:"recovery_success"`
}

// ParseError represents detailed parsing error information
type ParseError struct {
	Type        string            `json:"type"` // syntax, timeout, memory, io, unknown
	Message     string            `json:"message"`
	FilePath    string            `json:"file_path"`
	Line        int               `json:"line"`
	Column      int               `json:"column"`
	Context     string            `json:"context"`     // surrounding code context
	Severity    string            `json:"severity"`    // error, warning, info
	Recoverable bool              `json:"recoverable"` // whether recovery was attempted
	Suggestions []string          `json:"suggestions"` // potential fixes
	Metadata    map[string]string `json:"metadata"`
}

// RecoveryStrategy defines how to handle specific error types
type RecoveryStrategy struct {
	ErrorType    string
	MaxAttempts  int
	RecoveryFunc func(*ParseError, []byte) ([]byte, error)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(config ErrorConfig) *ErrorHandler {
	if config.MaxErrors <= 0 {
		config.MaxErrors = 100
	}
	if config.ErrorThreshold <= 0 {
		config.ErrorThreshold = 0.5 // 50% error rate threshold
	}
	if config.LogLevel == "" {
		config.LogLevel = "error"
	}

	return &ErrorHandler{
		config: config,
		stats: ErrorStats{
			ErrorTypes: make(map[string]int),
		},
	}
}

// HandleParseError processes a parsing error and determines recovery strategy
func (eh *ErrorHandler) HandleParseError(err error, filePath string, content []byte) *ParseError {
	parseError := eh.classifyError(err, filePath, content)

	// Update statistics
	eh.updateStats(parseError)

	// Attempt recovery if enabled and error is recoverable
	if eh.config.EnableRecovery && parseError.Recoverable {
		eh.attemptRecovery(parseError, content)
	}

	return parseError
}

// ShouldContinue determines if parsing should continue based on error rate
func (eh *ErrorHandler) ShouldContinue() bool {
	if eh.stats.TotalFiles == 0 {
		return true
	}

	// Check if we've exceeded maximum errors
	if eh.stats.TotalErrors >= eh.config.MaxErrors {
		return false
	}

	// Check error rate threshold
	errorRate := float64(eh.stats.FailedFiles) / float64(eh.stats.TotalFiles)
	return errorRate <= eh.config.ErrorThreshold
}

// GetStats returns current error statistics
func (eh *ErrorHandler) GetStats() ErrorStats {
	if eh.stats.TotalFiles > 0 {
		eh.stats.ErrorRate = float64(eh.stats.FailedFiles) / float64(eh.stats.TotalFiles)
	}
	return eh.stats
}

// GenerateErrorReport creates a detailed error report
func (eh *ErrorHandler) GenerateErrorReport() ErrorReport {
	stats := eh.GetStats()

	report := ErrorReport{
		Summary: ErrorSummary{
			TotalFiles:      stats.TotalFiles,
			SuccessfulFiles: stats.SuccessfulFiles,
			FailedFiles:     stats.FailedFiles,
			PartialFiles:    stats.PartialFiles,
			ErrorRate:       stats.ErrorRate,
			OverallStatus:   eh.getOverallStatus(),
		},
		ErrorBreakdown: stats.ErrorTypes,
		RecoveryStats: RecoveryStats{
			TotalAttempts:        stats.RecoveryAttempts,
			SuccessfulRecoveries: stats.RecoverySuccess,
			RecoveryRate:         eh.calculateRecoveryRate(),
		},
		Recommendations: eh.generateRecommendations(),
	}

	return report
}

// ErrorReport contains comprehensive error analysis
type ErrorReport struct {
	Summary         ErrorSummary   `json:"summary"`
	ErrorBreakdown  map[string]int `json:"error_breakdown"`
	RecoveryStats   RecoveryStats  `json:"recovery_stats"`
	Recommendations []string       `json:"recommendations"`
}

// ErrorSummary provides high-level error statistics
type ErrorSummary struct {
	TotalFiles      int     `json:"total_files"`
	SuccessfulFiles int     `json:"successful_files"`
	FailedFiles     int     `json:"failed_files"`
	PartialFiles    int     `json:"partial_files"`
	ErrorRate       float64 `json:"error_rate"`
	OverallStatus   string  `json:"overall_status"` // success, warning, error, critical
}

// RecoveryStats tracks error recovery performance
type RecoveryStats struct {
	TotalAttempts        int     `json:"total_attempts"`
	SuccessfulRecoveries int     `json:"successful_recoveries"`
	RecoveryRate         float64 `json:"recovery_rate"`
}

// Private methods

func (eh *ErrorHandler) classifyError(err error, filePath string, content []byte) *ParseError {
	parseError := &ParseError{
		FilePath:    filePath,
		Message:     err.Error(),
		Severity:    "error",
		Recoverable: false,
		Suggestions: []string{},
		Metadata:    make(map[string]string),
	}

	errorMsg := strings.ToLower(err.Error())

	// Classify error type based on error message
	switch {
	case strings.Contains(errorMsg, "syntax"):
		parseError.Type = "syntax"
		parseError.Recoverable = true
		parseError.Suggestions = append(parseError.Suggestions, "Check for missing brackets, parentheses, or semicolons")
		parseError = eh.extractSyntaxErrorDetails(parseError, content)

	case strings.Contains(errorMsg, "timeout"):
		parseError.Type = "timeout"
		parseError.Severity = "warning"
		parseError.Suggestions = append(parseError.Suggestions, "Consider reducing file size or increasing timeout")

	case strings.Contains(errorMsg, "memory") || strings.Contains(errorMsg, "out of memory"):
		parseError.Type = "memory"
		parseError.Suggestions = append(parseError.Suggestions, "File may be too large for parsing")

	case strings.Contains(errorMsg, "no such file") || strings.Contains(errorMsg, "permission"):
		parseError.Type = "io"
		parseError.Suggestions = append(parseError.Suggestions, "Check file permissions and existence")

	case strings.Contains(errorMsg, "encoding") || strings.Contains(errorMsg, "utf"):
		parseError.Type = "encoding"
		parseError.Recoverable = true
		parseError.Suggestions = append(parseError.Suggestions, "File may have encoding issues")

	default:
		parseError.Type = "unknown"
		parseError.Suggestions = append(parseError.Suggestions, "Review file for structural issues")
	}

	// Add file context
	parseError.Metadata["file_size"] = fmt.Sprintf("%d", len(content))
	parseError.Metadata["file_extension"] = getFileExtension(filePath)

	return parseError
}

func (eh *ErrorHandler) extractSyntaxErrorDetails(parseError *ParseError, content []byte) *ParseError {
	// Try to extract line/column information from error message
	// This is a simplified implementation - real tree-sitter errors may have different formats
	lines := strings.Split(string(content), "\n")

	// Find potential problematic areas (unclosed brackets, etc.)
	bracketStack := []rune{}
	for lineNum, line := range lines {
		for colNum, char := range line {
			switch char {
			case '(', '[', '{':
				bracketStack = append(bracketStack, char)
			case ')', ']', '}':
				if len(bracketStack) == 0 {
					parseError.Line = lineNum + 1
					parseError.Column = colNum + 1
					parseError.Context = eh.getLineContext(lines, lineNum, 2)
					parseError.Suggestions = append(parseError.Suggestions, "Unexpected closing bracket")
					return parseError
				}
				// Pop matching bracket
				bracketStack = bracketStack[:len(bracketStack)-1]
			}
		}
	}

	// Check for unclosed brackets
	if len(bracketStack) > 0 {
		parseError.Line = len(lines)
		parseError.Column = len(lines[len(lines)-1])
		parseError.Context = eh.getLineContext(lines, len(lines)-1, 2)
		parseError.Suggestions = append(parseError.Suggestions, "Unclosed brackets detected")
	}

	return parseError
}

func (eh *ErrorHandler) getLineContext(lines []string, lineNum, contextLines int) string {
	start := max(0, lineNum-contextLines)
	end := min(len(lines), lineNum+contextLines+1)

	contextBuilder := strings.Builder{}
	for i := start; i < end; i++ {
		prefix := "  "
		if i == lineNum {
			prefix = "→ "
		}
		contextBuilder.WriteString(fmt.Sprintf("%s%d: %s\n", prefix, i+1, lines[i]))
	}

	return contextBuilder.String()
}

func (eh *ErrorHandler) updateStats(parseError *ParseError) {
	eh.stats.TotalErrors++
	eh.stats.ErrorTypes[parseError.Type]++

	// File-level statistics are updated by the caller
}

func (eh *ErrorHandler) attemptRecovery(parseError *ParseError, content []byte) {
	eh.stats.RecoveryAttempts++

	// Implement recovery strategies based on error type
	switch parseError.Type {
	case "syntax":
		if eh.attemptSyntaxRecovery(parseError, content) {
			eh.stats.RecoverySuccess++
			parseError.Metadata["recovery_attempted"] = "true"
			parseError.Metadata["recovery_successful"] = "true"
		}

	case "encoding":
		if eh.attemptEncodingRecovery(parseError, content) {
			eh.stats.RecoverySuccess++
			parseError.Metadata["recovery_attempted"] = "true"
			parseError.Metadata["recovery_successful"] = "true"
		}
	}
}

func (eh *ErrorHandler) attemptSyntaxRecovery(parseError *ParseError, content []byte) bool {
	// Simple syntax recovery attempts
	contentStr := string(content)

	// Try adding missing closing brackets
	openBrackets := strings.Count(contentStr, "{") - strings.Count(contentStr, "}")
	if openBrackets > 0 {
		parseError.Suggestions = append(parseError.Suggestions,
			fmt.Sprintf("Try adding %d closing brace(s)", openBrackets))
		return true
	}

	openParens := strings.Count(contentStr, "(") - strings.Count(contentStr, ")")
	if openParens > 0 {
		parseError.Suggestions = append(parseError.Suggestions,
			fmt.Sprintf("Try adding %d closing parenthesis/es", openParens))
		return true
	}

	return false
}

func (eh *ErrorHandler) attemptEncodingRecovery(parseError *ParseError, content []byte) bool {
	// Try to detect and suggest encoding fixes
	if !isValidUTF8(content) {
		parseError.Suggestions = append(parseError.Suggestions, "File appears to have encoding issues - try converting to UTF-8")
		return true
	}

	return false
}

func (eh *ErrorHandler) getOverallStatus() string {
	stats := eh.stats
	if stats.TotalFiles == 0 {
		return "unknown"
	}

	errorRate := float64(stats.FailedFiles) / float64(stats.TotalFiles)

	switch {
	case errorRate == 0:
		return "success"
	case errorRate <= 0.1:
		return "warning"
	case errorRate <= 0.5:
		return "error"
	default:
		return "critical"
	}
}

func (eh *ErrorHandler) calculateRecoveryRate() float64 {
	if eh.stats.RecoveryAttempts == 0 {
		return 0.0
	}
	return float64(eh.stats.RecoverySuccess) / float64(eh.stats.RecoveryAttempts)
}

func (eh *ErrorHandler) generateRecommendations() []string {
	var recommendations []string
	stats := eh.stats

	if stats.TotalFiles == 0 {
		return recommendations
	}

	errorRate := float64(stats.FailedFiles) / float64(stats.TotalFiles)

	// Generate recommendations based on error patterns
	if errorRate > 0.3 {
		recommendations = append(recommendations, "High error rate detected - consider reviewing file quality")
	}

	if stats.ErrorTypes["syntax"] > stats.TotalErrors/2 {
		recommendations = append(recommendations, "Many syntax errors - ensure files are well-formed before parsing")
	}

	if stats.ErrorTypes["memory"] > 0 {
		recommendations = append(recommendations, "Memory errors detected - consider implementing file size limits")
	}

	if stats.ErrorTypes["timeout"] > 0 {
		recommendations = append(recommendations, "Timeout errors detected - consider increasing timeout or reducing file size")
	}

	if eh.calculateRecoveryRate() < 0.3 && eh.stats.RecoveryAttempts > 0 {
		recommendations = append(recommendations, "Low recovery rate - consider improving error recovery strategies")
	}

	return recommendations
}

// Utility functions

func getFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

func isValidUTF8(content []byte) bool {
	// Simple UTF-8 validation
	return utf8.Valid(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
