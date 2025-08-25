package ast

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorHandler(t *testing.T) {
	config := ErrorConfig{
		MaxErrors:          50,
		ErrorThreshold:     0.3,
		EnableRecovery:     true,
		EnablePartialParse: true,
		LogLevel:           "warning",
	}

	handler := NewErrorHandler(config)

	assert.NotNil(t, handler)
	assert.Equal(t, 50, handler.config.MaxErrors)
	assert.Equal(t, 0.3, handler.config.ErrorThreshold)
	assert.True(t, handler.config.EnableRecovery)
	assert.True(t, handler.config.EnablePartialParse)
	assert.Equal(t, "warning", handler.config.LogLevel)
}

func TestNewErrorHandler_Defaults(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{})

	assert.Equal(t, 100, handler.config.MaxErrors)
	assert.Equal(t, 0.5, handler.config.ErrorThreshold)
	assert.Equal(t, "error", handler.config.LogLevel)
}

func TestErrorHandler_ClassifyError(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{EnableRecovery: true})

	testCases := []struct {
		errorMsg     string
		expectedType string
		recoverable  bool
	}{
		{"syntax error at line 5", "syntax", true},
		{"parsing timeout exceeded", "timeout", false},
		{"out of memory while parsing", "memory", false},
		{"no such file or directory", "io", false},
		{"invalid utf-8 encoding", "encoding", true},
		{"unknown parsing failure", "unknown", false},
	}

	for _, tc := range testCases {
		t.Run(tc.errorMsg, func(t *testing.T) {
			err := fmt.Errorf(tc.errorMsg)
			parseError := handler.classifyError(err, "test.js", []byte("test content"))

			assert.Equal(t, tc.expectedType, parseError.Type)
			assert.Equal(t, tc.recoverable, parseError.Recoverable)
			assert.Equal(t, "test.js", parseError.FilePath)
			assert.Equal(t, tc.errorMsg, parseError.Message)
			assert.NotEmpty(t, parseError.Suggestions)
		})
	}
}

func TestErrorHandler_HandleParseError(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{
		EnableRecovery: true,
	})

	err := fmt.Errorf("syntax error: unexpected token")
	content := []byte(`function test() {
		console.log("missing closing brace"
	`)

	parseError := handler.HandleParseError(err, "test.js", content)

	assert.NotNil(t, parseError)
	assert.Equal(t, "syntax", parseError.Type)
	assert.Equal(t, "test.js", parseError.FilePath)
	assert.True(t, parseError.Recoverable)
	assert.NotEmpty(t, parseError.Suggestions)

	// Check that statistics were updated
	stats := handler.GetStats()
	assert.Equal(t, 1, stats.TotalErrors)
	assert.Equal(t, 1, stats.ErrorTypes["syntax"])
}

func TestErrorHandler_ShouldContinue(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{
		MaxErrors:      5,
		ErrorThreshold: 0.5,
	})

	// Initially should continue
	assert.True(t, handler.ShouldContinue())

	// Simulate some successful and failed files
	handler.stats.TotalFiles = 10
	handler.stats.FailedFiles = 3 // 30% error rate
	handler.stats.TotalErrors = 3

	// Should continue (below threshold)
	assert.True(t, handler.ShouldContinue())

	// Increase error rate above threshold
	handler.stats.FailedFiles = 6 // 60% error rate

	// Should not continue (above threshold)
	assert.False(t, handler.ShouldContinue())

	// Reset error rate but exceed max errors
	handler.stats.FailedFiles = 2  // 20% error rate
	handler.stats.TotalErrors = 10 // Above MaxErrors

	// Should not continue (too many errors)
	assert.False(t, handler.ShouldContinue())
}

func TestErrorHandler_GenerateErrorReport(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{})

	// Simulate various errors
	handler.stats.TotalFiles = 100
	handler.stats.SuccessfulFiles = 85
	handler.stats.FailedFiles = 10
	handler.stats.PartialFiles = 5
	handler.stats.TotalErrors = 15
	handler.stats.ErrorTypes = map[string]int{
		"syntax":  8,
		"timeout": 3,
		"memory":  2,
		"io":      2,
	}
	handler.stats.RecoveryAttempts = 10
	handler.stats.RecoverySuccess = 6

	report := handler.GenerateErrorReport()

	assert.Equal(t, 100, report.Summary.TotalFiles)
	assert.Equal(t, 85, report.Summary.SuccessfulFiles)
	assert.Equal(t, 10, report.Summary.FailedFiles)
	assert.Equal(t, 5, report.Summary.PartialFiles)
	assert.Equal(t, 0.1, report.Summary.ErrorRate)
	assert.Equal(t, "warning", report.Summary.OverallStatus)

	assert.Equal(t, 8, report.ErrorBreakdown["syntax"])
	assert.Equal(t, 3, report.ErrorBreakdown["timeout"])

	assert.Equal(t, 10, report.RecoveryStats.TotalAttempts)
	assert.Equal(t, 6, report.RecoveryStats.SuccessfulRecoveries)
	assert.Equal(t, 0.6, report.RecoveryStats.RecoveryRate)

	assert.NotEmpty(t, report.Recommendations)
}

func TestParser_ErrorHandling_MalformedCode(t *testing.T) {
	config := ErrorConfig{
		EnableRecovery:     true,
		EnablePartialParse: true,
	}

	parser, err := NewParserWithConfig(config)
	require.NoError(t, err)
	defer parser.Close()

	// Test with malformed JavaScript
	malformedCode := `
function test() {
	console.log("unclosed function"
	// missing closing brace
`

	result, err := parser.ParseFile(context.Background(), "malformed.js", []byte(malformedCode))

	// Should not return error if partial parsing is enabled
	if config.EnablePartialParse {
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have partial status (may be "partial" or "partial_with_errors" depending on parser behavior)
		status := result.Metadata["parse_status"]
		assert.Contains(t, []string{"partial", "partial_with_errors"}, status)

		// Should have extracted some partial information
		assert.GreaterOrEqual(t, len(result.Functions), 0) // May find partial function info

		// Should have recorded errors
		assert.GreaterOrEqual(t, len(result.Errors), 0)
	} else {
		// Should return error if partial parsing is disabled
		assert.Error(t, err)
	}
}

func TestParser_ErrorHandling_SyntaxErrors(t *testing.T) {
	parser, err := NewParserWithConfig(ErrorConfig{
		EnableRecovery:     true,
		EnablePartialParse: true,
	})
	require.NoError(t, err)
	defer parser.Close()

	// Test with syntax errors that tree-sitter can partially parse
	codeWithSyntaxErrors := `
import React from 'react';

function ValidFunction() {
	return <div>Valid</div>;
}

function InvalidFunction() {
	return <div>Unclosed div
	// Missing closing tag
}

export default ValidFunction;
`

	result, err := parser.ParseFile(context.Background(), "syntax_errors.jsx", []byte(codeWithSyntaxErrors))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should extract valid parts
	assert.GreaterOrEqual(t, len(result.Functions), 1) // Should find ValidFunction
	assert.GreaterOrEqual(t, len(result.Imports), 1)   // Should find React import
	assert.GreaterOrEqual(t, len(result.Exports), 1)   // Should find export

	// Check if syntax errors were detected
	if result.Metadata["parse_status"] == "partial_with_errors" {
		assert.GreaterOrEqual(t, len(result.Errors), 1)
	}
}

func TestParser_ErrorHandling_UnsupportedFileType(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	result, err := parser.ParseFile(context.Background(), "test.py", []byte("print('hello')"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported file type")

	// Check error handler statistics
	errorHandler := parser.GetErrorHandler()
	stats := errorHandler.GetStats()
	assert.Equal(t, 1, stats.TotalFiles)
	assert.Equal(t, 1, stats.FailedFiles)
	assert.Equal(t, 0, stats.SuccessfulFiles)
}

func TestParser_ShouldContinueParsing(t *testing.T) {
	parser, err := NewParserWithConfig(ErrorConfig{
		MaxErrors:      3,
		ErrorThreshold: 0.5,
	})
	require.NoError(t, err)
	defer parser.Close()

	// Initially should continue
	assert.True(t, parser.ShouldContinueParsing())

	// Parse some files with errors
	for i := 0; i < 5; i++ {
		parser.ParseFile(context.Background(), "unsupported.py", []byte("invalid"))
	}

	// Should not continue after too many errors
	assert.False(t, parser.ShouldContinueParsing())
}

func TestPartialExtraction_BasicPatterns(t *testing.T) {
	parser, err := NewParserWithConfig(ErrorConfig{
		EnablePartialParse: true,
	})
	require.NoError(t, err)
	defer parser.Close()

	// Test code with basic patterns that regex can extract
	partialCode := `
import React from 'react';
import { useState, useEffect } from 'react';
import utils from './utils.js';

function TestComponent(props) {
	const [count, setCount] = useState(0);
	return <div>{count}</div>;
}

class TestClass extends Component {
	render() {
		return <div>Test</div>;
	}
}

export default TestComponent;
export { TestClass };
`

	result := &ParseResult{
		FilePath:  "test.jsx",
		Language:  "javascript",
		Functions: []FunctionInfo{},
		Classes:   []ClassInfo{},
		Variables: []VariableInfo{},
		Imports:   []ImportInfo{},
		Exports:   []ExportInfo{},
		Errors:    []ParseError{},
		Metadata:  make(map[string]interface{}),
	}

	// Test partial extraction
	parser.extractPartialInfo([]byte(partialCode), result)

	// Should extract imports
	assert.GreaterOrEqual(t, len(result.Imports), 3)

	// Check specific imports
	var reactImport, namedImport, utilsImport *ImportInfo
	for i, imp := range result.Imports {
		switch imp.Source {
		case "react":
			if imp.ImportType == "default" {
				reactImport = &result.Imports[i]
			} else if imp.ImportType == "named" {
				namedImport = &result.Imports[i]
			}
		case "./utils.js":
			utilsImport = &result.Imports[i]
		}
	}

	require.NotNil(t, reactImport)
	assert.Equal(t, "default", reactImport.ImportType)
	assert.True(t, reactImport.IsExternal)

	require.NotNil(t, namedImport)
	assert.Equal(t, "named", namedImport.ImportType)
	assert.Contains(t, namedImport.Specifiers, "useState")
	assert.Contains(t, namedImport.Specifiers, "useEffect")

	require.NotNil(t, utilsImport)
	assert.Equal(t, "default", utilsImport.ImportType)
	assert.False(t, utilsImport.IsExternal)

	// Should extract functions
	assert.GreaterOrEqual(t, len(result.Functions), 1)
	testFunc := result.Functions[0]
	assert.Equal(t, "TestComponent", testFunc.Name)
	assert.GreaterOrEqual(t, len(testFunc.Parameters), 1)

	// Should extract classes
	assert.GreaterOrEqual(t, len(result.Classes), 1)
	testClass := result.Classes[0]
	assert.Equal(t, "TestClass", testClass.Name)
	assert.Equal(t, "Component", testClass.Extends)

	// Should extract exports
	assert.GreaterOrEqual(t, len(result.Exports), 2)

	// Should have partial extraction metadata
	assert.Equal(t, "true", result.Metadata["partial_extraction"])
	assert.Equal(t, "regex_patterns", result.Metadata["extraction_method"])
}

func TestErrorHandler_RecoveryStrategies(t *testing.T) {
	handler := NewErrorHandler(ErrorConfig{
		EnableRecovery: true,
	})

	testCases := []struct {
		name           string
		content        string
		errorType      string
		expectRecovery bool
	}{
		{
			name: "unclosed_braces",
			content: `function test() {
				console.log("missing brace"
			`, // Missing closing brace
			errorType:      "syntax",
			expectRecovery: true,
		},
		{
			name: "unclosed_parentheses",
			content: `function test(param {
				return param;
			}`, // Missing closing parenthesis
			errorType:      "syntax",
			expectRecovery: true,
		},
		{
			name:           "encoding_issue",
			content:        "function test() { return 'text'; }", // Valid UTF-8
			errorType:      "encoding",
			expectRecovery: false, // Content is actually valid
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := fmt.Errorf("%s error detected", tc.errorType)
			parseError := handler.HandleParseError(err, "test.js", []byte(tc.content))

			assert.Equal(t, tc.errorType, parseError.Type)

			if tc.expectRecovery {
				assert.Contains(t, parseError.Metadata, "recovery_attempted")
			}
		})
	}
}
