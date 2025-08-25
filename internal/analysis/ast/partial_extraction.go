package ast

import (
	"regexp"
	"strings"
)

// extractPartialInfo attempts to extract basic information using regex patterns
// when tree-sitter parsing fails completely
func (p *Parser) extractPartialInfo(content []byte, result *ParseResult) {
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Extract imports using regex patterns
	p.extractPartialImports(contentStr, lines, result)

	// Extract exports using regex patterns
	p.extractPartialExports(contentStr, lines, result)

	// Extract basic function declarations
	p.extractPartialFunctions(contentStr, lines, result)

	// Extract basic class declarations
	p.extractPartialClasses(contentStr, lines, result)

	// Extract basic variable declarations
	p.extractPartialVariables(contentStr, lines, result)

	// Add metadata about partial extraction
	result.Metadata["partial_extraction"] = "true"
	result.Metadata["extraction_method"] = "regex_patterns"
}

// extractPartialImports extracts import statements using regex
func (p *Parser) extractPartialImports(content string, lines []string, result *ParseResult) {
	// Common import patterns
	patterns := []*regexp.Regexp{
		// import { named } from 'module'
		regexp.MustCompile(`import\s*\{\s*([^}]+)\s*\}\s*from\s*['"]([^'"]+)['"]`),
		// import defaultName from 'module'
		regexp.MustCompile(`import\s+(\w+)\s+from\s*['"]([^'"]+)['"]`),
		// import * as name from 'module'
		regexp.MustCompile(`import\s*\*\s*as\s+(\w+)\s+from\s*['"]([^'"]+)['"]`),
		// import 'module' (side-effect)
		regexp.MustCompile(`import\s*['"]([^'"]+)['"]`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 0 {
				importInfo := ImportInfo{
					StartLine: lineNum + 1,
				}

				switch {
				case strings.Contains(pattern.String(), `\{\s*([^}]+)\s*\}`): // Named imports
					importInfo.ImportType = "named"
					importInfo.Source = matches[2]
					// Parse named imports
					namedImports := strings.Split(matches[1], ",")
					for _, imp := range namedImports {
						imp = strings.TrimSpace(imp)
						if imp != "" {
							importInfo.Specifiers = append(importInfo.Specifiers, imp)
						}
					}

				case strings.Contains(pattern.String(), `\*\s*as`): // Namespace imports
					importInfo.ImportType = "namespace"
					importInfo.LocalName = matches[1]
					importInfo.Source = matches[2]
					importInfo.Specifiers = []string{matches[1]}

				case len(matches) == 3: // Default imports
					importInfo.ImportType = "default"
					importInfo.LocalName = matches[1]
					importInfo.Source = matches[2]
					importInfo.Specifiers = []string{matches[1]}

				case len(matches) == 2: // Side-effect imports
					importInfo.ImportType = "side-effect"
					importInfo.Source = matches[1]
				}

				// Determine if external
				importInfo.IsExternal = p.isExternalImport(importInfo.Source)

				result.Imports = append(result.Imports, importInfo)
				break // Only match first pattern per line
			}
		}
	}
}

// extractPartialExports extracts export statements using regex
func (p *Parser) extractPartialExports(content string, lines []string, result *ParseResult) {
	patterns := []*regexp.Regexp{
		// export { named }
		regexp.MustCompile(`export\s*\{\s*([^}]+)\s*\}`),
		// export default name
		regexp.MustCompile(`export\s+default\s+(\w+)`),
		// export function name()
		regexp.MustCompile(`export\s+function\s+(\w+)`),
		// export class Name
		regexp.MustCompile(`export\s+class\s+(\w+)`),
		// export const/let/var name
		regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 0 {
				exportInfo := ExportInfo{
					StartLine: lineNum + 1,
				}

				switch {
				case strings.Contains(pattern.String(), `\{\s*([^}]+)\s*\}`): // Named exports
					exportInfo.ExportType = "named"
					// Parse named exports
					namedExports := strings.Split(matches[1], ",")
					for _, exp := range namedExports {
						exp = strings.TrimSpace(exp)
						if exp != "" {
							exportInfo.Specifiers = append(exportInfo.Specifiers, exp)
						}
					}

				case strings.Contains(pattern.String(), "default"): // Default exports
					exportInfo.ExportType = "default"
					exportInfo.Name = matches[1]

				default: // Function, class, or variable exports
					exportInfo.ExportType = "named"
					exportInfo.Name = matches[1]
					exportInfo.Specifiers = []string{matches[1]}
				}

				result.Exports = append(result.Exports, exportInfo)
				break // Only match first pattern per line
			}
		}
	}
}

// extractPartialFunctions extracts function declarations using regex
func (p *Parser) extractPartialFunctions(content string, lines []string, result *ParseResult) {
	patterns := []*regexp.Regexp{
		// function name(params)
		regexp.MustCompile(`(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)`),
		// const name = function(params)
		regexp.MustCompile(`(?:export\s+)?const\s+(\w+)\s*=\s*(?:async\s+)?function\s*\(([^)]*)\)`),
		// const name = (params) =>
		regexp.MustCompile(`(?:export\s+)?const\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)\s*=>`),
		// const name = async (params) =>
		regexp.MustCompile(`(?:export\s+)?const\s+(\w+)\s*=\s*async\s*\(([^)]*)\)\s*=>`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 2 {
				functionInfo := FunctionInfo{
					Name:       matches[1],
					Parameters: []ParameterInfo{},
					StartLine:  lineNum + 1,
					EndLine:    lineNum + 1, // Approximate
					IsAsync:    strings.Contains(line, "async"),
					IsExported: strings.Contains(line, "export"),
					Metadata:   make(map[string]string),
				}

				// Parse parameters if present
				if len(matches) > 2 && matches[2] != "" {
					params := strings.Split(matches[2], ",")
					for _, param := range params {
						param = strings.TrimSpace(param)
						if param != "" {
							// Simple parameter parsing
							paramName := param
							paramType := ""

							// Check for TypeScript type annotation
							if colonIndex := strings.Index(param, ":"); colonIndex != -1 {
								paramName = strings.TrimSpace(param[:colonIndex])
								paramType = strings.TrimSpace(param[colonIndex+1:])
							}

							functionInfo.Parameters = append(functionInfo.Parameters, ParameterInfo{
								Name: paramName,
								Type: paramType,
							})
						}
					}
				}

				functionInfo.Metadata["extraction_method"] = "regex"
				functionInfo.Metadata["pattern_matched"] = pattern.String()

				result.Functions = append(result.Functions, functionInfo)
				break // Only match first pattern per line
			}
		}
	}
}

// extractPartialClasses extracts class declarations using regex
func (p *Parser) extractPartialClasses(content string, lines []string, result *ParseResult) {
	patterns := []*regexp.Regexp{
		// class Name extends Parent
		regexp.MustCompile(`(?:export\s+)?class\s+(\w+)(?:\s+extends\s+(\w+))?`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 2 {
				classInfo := ClassInfo{
					Name:       matches[1],
					Methods:    []FunctionInfo{},
					Properties: []PropertyInfo{},
					StartLine:  lineNum + 1,
					EndLine:    lineNum + 1, // Approximate
					IsExported: strings.Contains(line, "export"),
					Metadata:   make(map[string]string),
				}

				// Check for extends clause
				if len(matches) > 2 && matches[2] != "" {
					classInfo.Extends = matches[2]
				}

				classInfo.Metadata["extraction_method"] = "regex"

				result.Classes = append(result.Classes, classInfo)
				break // Only match first pattern per line
			}
		}
	}
}

// extractPartialVariables extracts variable declarations using regex
func (p *Parser) extractPartialVariables(content string, lines []string, result *ParseResult) {
	patterns := []*regexp.Regexp{
		// const/let/var name: type = value
		regexp.MustCompile(`(?:export\s+)?(const|let|var)\s+(\w+)(?:\s*:\s*([^=]+))?\s*=`),
		// const/let/var name: type
		regexp.MustCompile(`(?:export\s+)?(const|let|var)\s+(\w+)\s*:\s*([^;,\n]+)`),
	}

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		for _, pattern := range patterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				variableInfo := VariableInfo{
					Name:       matches[2],
					Kind:       matches[1],
					StartLine:  lineNum + 1,
					IsExported: strings.Contains(line, "export"),
					Metadata:   make(map[string]string),
				}

				// Extract type if present
				if len(matches) > 3 && matches[3] != "" {
					variableInfo.Type = strings.TrimSpace(matches[3])
				}

				variableInfo.Metadata["extraction_method"] = "regex"

				result.Variables = append(result.Variables, variableInfo)
				break // Only match first pattern per line
			}
		}
	}
}

// GetErrorHandler returns the parser's error handler for external access
func (p *Parser) GetErrorHandler() *ErrorHandler {
	return p.errorHandler
}

// ShouldContinueParsing checks if parsing should continue based on error rate
func (p *Parser) ShouldContinueParsing() bool {
	if p.errorHandler == nil {
		return true
	}
	return p.errorHandler.ShouldContinue()
}
