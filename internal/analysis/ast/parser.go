package ast

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Parser handles AST parsing for JavaScript and TypeScript files
type Parser struct {
	jsParser     *sitter.Parser
	tsParser     *sitter.Parser
	tsxParser    *sitter.Parser
	errorHandler *ErrorHandler
	mu           sync.RWMutex
}

// ParseResult contains the structured AST analysis results
type ParseResult struct {
	FilePath   string                 `json:"file_path"`
	Language   string                 `json:"language"`
	Functions  []FunctionInfo         `json:"functions"`
	Classes    []ClassInfo            `json:"classes"`
	Interfaces []InterfaceInfo        `json:"interfaces"`
	Variables  []VariableInfo         `json:"variables"`
	Imports    []ImportInfo           `json:"imports"`
	Exports    []ExportInfo           `json:"exports"`
	Errors     []ParseError           `json:"errors"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// FunctionInfo represents a parsed function
type FunctionInfo struct {
	Name       string            `json:"name"`
	Parameters []ParameterInfo   `json:"parameters"`
	ReturnType string            `json:"return_type"`
	IsAsync    bool              `json:"is_async"`
	IsExported bool              `json:"is_exported"`
	StartLine  int               `json:"start_line"`
	EndLine    int               `json:"end_line"`
	Metadata   map[string]string `json:"metadata"`
}

// ParameterInfo represents function parameters
type ParameterInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	DefaultValue string `json:"default_value"`
	IsOptional   bool   `json:"is_optional"`
}

// ClassInfo represents a parsed class
type ClassInfo struct {
	Name       string            `json:"name"`
	Extends    string            `json:"extends"`
	Implements []string          `json:"implements"`
	Methods    []FunctionInfo    `json:"methods"`
	Properties []PropertyInfo    `json:"properties"`
	IsExported bool              `json:"is_exported"`
	StartLine  int               `json:"start_line"`
	EndLine    int               `json:"end_line"`
	Metadata   map[string]string `json:"metadata"`
}

// PropertyInfo represents class properties
type PropertyInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsStatic   bool   `json:"is_static"`
	IsPrivate  bool   `json:"is_private"`
	IsReadonly bool   `json:"is_readonly"`
}

// InterfaceInfo represents TypeScript interfaces
type InterfaceInfo struct {
	Name       string            `json:"name"`
	Extends    []string          `json:"extends"`
	Properties []PropertyInfo    `json:"properties"`
	Methods    []FunctionInfo    `json:"methods"`
	IsExported bool              `json:"is_exported"`
	StartLine  int               `json:"start_line"`
	EndLine    int               `json:"end_line"`
	Metadata   map[string]string `json:"metadata"`
}

// VariableInfo represents variable declarations
type VariableInfo struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Kind       string            `json:"kind"` // var, let, const
	IsExported bool              `json:"is_exported"`
	StartLine  int               `json:"start_line"`
	Metadata   map[string]string `json:"metadata"`
}

// ImportInfo represents import statements
type ImportInfo struct {
	Source     string   `json:"source"`
	ImportType string   `json:"import_type"` // default, named, namespace, side-effect
	Specifiers []string `json:"specifiers"`
	LocalName  string   `json:"local_name"`
	StartLine  int      `json:"start_line"`
	IsExternal bool     `json:"is_external"`
}

// ExportInfo represents export statements
type ExportInfo struct {
	Name       string   `json:"name"`
	ExportType string   `json:"export_type"` // default, named, all
	Source     string   `json:"source"`      // re-export source
	Specifiers []string `json:"specifiers"`
	StartLine  int      `json:"start_line"`
}

// Note: ParseError is now defined in error_handler.go

// NewParser creates a new AST parser instance
func NewParser() (*Parser, error) {
	return NewParserWithConfig(ErrorConfig{
		MaxErrors:          100,
		ErrorThreshold:     0.5,
		EnableRecovery:     true,
		EnablePartialParse: true,
		LogLevel:           "error",
	})
}

// NewParserWithConfig creates a new AST parser with custom error handling configuration
func NewParserWithConfig(errorConfig ErrorConfig) (*Parser, error) {
	p := &Parser{
		errorHandler: NewErrorHandler(errorConfig),
	}

	// Initialize JavaScript parser
	p.jsParser = sitter.NewParser()
	p.jsParser.SetLanguage(javascript.GetLanguage())

	// Initialize TypeScript parser
	p.tsParser = sitter.NewParser()
	p.tsParser.SetLanguage(typescript.GetLanguage())

	// Initialize TSX parser
	p.tsxParser = sitter.NewParser()
	p.tsxParser.SetLanguage(tsx.GetLanguage())

	return p, nil
}

// ParseFile parses a single JavaScript or TypeScript file with error handling
func (p *Parser) ParseFile(ctx context.Context, filePath string, content []byte) (*ParseResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Update error handler statistics
	p.errorHandler.stats.TotalFiles++

	// Determine language and parser
	language, parser := p.getParserForFile(filePath)
	if parser == nil {
		err := fmt.Errorf("unsupported file type: %s", filePath)
		p.errorHandler.HandleParseError(err, filePath, content)
		p.errorHandler.stats.FailedFiles++

		// Return error for unsupported files
		return nil, err
	}

	// Initialize result structure
	result := &ParseResult{
		FilePath:   filePath,
		Language:   language,
		Functions:  []FunctionInfo{},
		Classes:    []ClassInfo{},
		Interfaces: []InterfaceInfo{},
		Variables:  []VariableInfo{},
		Imports:    []ImportInfo{},
		Exports:    []ExportInfo{},
		Errors:     []ParseError{},
		Metadata:   make(map[string]interface{}),
	}

	// Parse the content with error handling
	tree, err := parser.ParseCtx(ctx, nil, content)
	if err != nil {
		// Handle parse error gracefully
		parseError := p.errorHandler.HandleParseError(err, filePath, content)
		result.Errors = append(result.Errors, *parseError)

		// If partial parsing is enabled, try to continue with limited extraction
		if p.errorHandler.config.EnablePartialParse {
			p.errorHandler.stats.PartialFiles++
			result.Metadata["parse_status"] = "partial"
			result.Metadata["parse_error"] = err.Error()

			// Try basic pattern matching for partial results
			p.extractPartialInfo(content, result)
			return result, nil
		}

		p.errorHandler.stats.FailedFiles++
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}
	defer tree.Close()

	// Check for syntax errors in the parsed tree
	if tree.RootNode().HasError() {
		// Tree has syntax errors but parsing succeeded partially
		syntaxError := &ParseError{
			Type:     "syntax",
			Message:  "Syntax errors detected in parsed tree",
			FilePath: filePath,
			Severity: "warning",
			Context:  "Tree contains error nodes",
		}
		result.Errors = append(result.Errors, *syntaxError)
		p.errorHandler.stats.PartialFiles++
		result.Metadata["parse_status"] = "partial_with_errors"
	} else {
		result.Metadata["parse_status"] = "success"
	}

	// Extract AST information with error handling
	if err := p.extractASTInfo(tree.RootNode(), content, result); err != nil {
		parseError := p.errorHandler.HandleParseError(err, filePath, content)
		result.Errors = append(result.Errors, *parseError)

		// Continue with partial results if enabled
		if p.errorHandler.config.EnablePartialParse {
			p.errorHandler.stats.PartialFiles++
			result.Metadata["extraction_status"] = "partial"
			return result, nil
		}

		p.errorHandler.stats.FailedFiles++
		return nil, fmt.Errorf("failed to extract AST info: %w", err)
	}

	// Mark as successful
	p.errorHandler.stats.SuccessfulFiles++
	result.Metadata["extraction_status"] = "success"

	return result, nil
}

// getParserForFile determines the appropriate parser based on file extension
func (p *Parser) getParserForFile(filePath string) (string, *sitter.Parser) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".js", ".jsx":
		return "javascript", p.jsParser
	case ".ts":
		return "typescript", p.tsParser
	case ".tsx":
		return "tsx", p.tsxParser
	default:
		return "", nil
	}
}

// IsSupported checks if a file type is supported for parsing
func (p *Parser) IsSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx"
}

// Close releases parser resources
func (p *Parser) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.jsParser != nil {
		p.jsParser.Close()
	}
	if p.tsParser != nil {
		p.tsParser.Close()
	}
	if p.tsxParser != nil {
		p.tsxParser.Close()
	}

	return nil
}
