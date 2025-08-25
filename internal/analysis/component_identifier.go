package analysis

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

// ComponentType represents the type of a component in the application
type ComponentType string

const (
	ReactComponent ComponentType = "react_component"
	Service        ComponentType = "service"
	Utility        ComponentType = "utility"
	Configuration  ComponentType = "config"
	Middleware     ComponentType = "middleware"
)

// Component represents a single component in the application architecture
type Component struct {
	Name         string            `json:"name"`
	Type         ComponentType     `json:"type"`
	FilePath     string            `json:"file_path"`
	Exports      []string          `json:"exports"`
	Dependencies []string          `json:"dependencies"`
	Metadata     map[string]any    `json:"metadata"`
}

// ComponentIdentifier service for identifying and categorizing application components
type ComponentIdentifier struct {
	components []Component
	cache      map[string]Component
}

// NewComponentIdentifier creates a new component identifier instance
func NewComponentIdentifier() *ComponentIdentifier {
	return &ComponentIdentifier{
		components: make([]Component, 0),
		cache:      make(map[string]Component),
	}
}

// IdentifyComponent analyzes a file and determines its component type and properties
func (ci *ComponentIdentifier) IdentifyComponent(filePath string, content string) (*Component, error) {
	// Check cache first for performance
	if cached, exists := ci.cache[filePath]; exists {
		return &cached, nil
	}

	component := Component{
		Name:         extractComponentName(filePath),
		FilePath:     filePath,
		Exports:      make([]string, 0),
		Dependencies: make([]string, 0),
		Metadata:     make(map[string]any),
	}

	// Identify component type based on file content and patterns
	componentType, metadata := ci.analyzeComponentType(filePath, content)
	component.Type = componentType
	component.Metadata = metadata

	// Extract exports and dependencies
	component.Exports = ci.extractExports(content)
	component.Dependencies = ci.extractDependencies(content)

	// Cache the result
	ci.cache[filePath] = component
	
	return &component, nil
}

// analyzeComponentType determines the type of component based on file content analysis
func (ci *ComponentIdentifier) analyzeComponentType(filePath string, content string) (ComponentType, map[string]any) {
	metadata := make(map[string]any)

	// Check for React components first (JSX usage, React imports)
	if ci.isReactComponent(content, metadata) {
		return ReactComponent, metadata
	}

	// Check for middleware patterns early (before config)
	if ci.isMiddleware(content, metadata) {
		return Middleware, metadata
	}

	// Check for configuration files
	if ci.isConfigurationModule(filePath, content, metadata) {
		return Configuration, metadata
	}

	// Check for service patterns (business logic, external integrations)
	if ci.isService(content, metadata) {
		return Service, metadata
	}

	// Default to utility if it appears to be a pure function module
	if ci.isUtility(content, metadata) {
		return Utility, metadata
	}

	// Default fallback
	metadata["detection_confidence"] = "low"
	return Utility, metadata
}

// isReactComponent detects React functional and class components
func (ci *ComponentIdentifier) isReactComponent(content string, metadata map[string]any) bool {
	// Check for React imports
	hasReactImport := strings.Contains(content, "import React") || 
					  strings.Contains(content, "from 'react'") ||
					  strings.Contains(content, "from \"react\"")

	// More comprehensive JSX patterns
	hasJSX := (strings.Contains(content, "return <") || 
			   strings.Contains(content, "? <") ||
			   (strings.Contains(content, "return (") && 
			   (strings.Contains(content, "<div") || 
			    strings.Contains(content, "<span") ||
			    strings.Contains(content, "<button") ||
			    strings.Contains(content, "<Component") ||
			    strings.Contains(content, "{children}") ||
			    strings.Contains(content, "<>"))))

	// Check for functional component patterns (arrow functions returning JSX)
	hasFunctionalComponent := (strings.Contains(content, "const ") || strings.Contains(content, "function ")) && 
							  (strings.Contains(content, "=> {") || strings.Contains(content, "() {")) &&
							  (hasJSX || strings.Contains(content, "return <"))

	// Check for class component patterns  
	hasClassComponent := strings.Contains(content, "class ") &&
						 strings.Contains(content, "extends ") &&
						 (strings.Contains(content, "React.Component") || strings.Contains(content, "Component"))

	// Check for hooks usage (more comprehensive)
	hasHooks := strings.Contains(content, "useState") ||
				strings.Contains(content, "useEffect") ||
				strings.Contains(content, "useContext") ||
				strings.Contains(content, "useReducer") ||
				strings.Contains(content, "useMemo") ||
				strings.Contains(content, "useCallback") ||
				strings.Contains(content, "useRef")

	// More flexible React component detection
	isReactComp := hasReactImport && (hasJSX || hasFunctionalComponent || hasClassComponent || hasHooks)
	
	// Also check for custom hooks (functions starting with 'use' and using hooks)
	isCustomHook := hasHooks && strings.Contains(content, "const use") && !hasJSX

	if isReactComp || isCustomHook {
		metadata["has_jsx"] = hasJSX
		metadata["is_functional"] = hasFunctionalComponent && !hasClassComponent
		metadata["is_class"] = hasClassComponent
		metadata["uses_hooks"] = hasHooks
		metadata["detection_confidence"] = "high"
		
		// Check for HOC patterns
		if strings.Contains(content, "withAuth") || 
		   strings.Contains(content, "withRouter") ||
		   strings.Contains(content, "withStyles") {
			metadata["is_hoc"] = true
		}
		
		// Check for custom hooks
		if isCustomHook || (hasHooks && !hasJSX && strings.Contains(content, "use")) {
			metadata["is_custom_hook"] = true
		}
	}

	return isReactComp || isCustomHook
}

// isService detects service modules (business logic, external API integrations)
func (ci *ComponentIdentifier) isService(content string, metadata map[string]any) bool {
	// Look for service patterns
	hasServiceKeywords := strings.Contains(content, "Service") ||
						  strings.Contains(content, "Client") ||
						  strings.Contains(content, "Repository")

	// Check for HTTP client usage
	hasHTTPClient := strings.Contains(content, "axios") ||
					 strings.Contains(content, "fetch") ||
					 strings.Contains(content, "http.") ||
					 strings.Contains(content, "request")

	// Check for database operations - be more specific
	hasDBOperations := (strings.Contains(content, "SELECT") ||
					   strings.Contains(content, "INSERT") ||
					   strings.Contains(content, "UPDATE") ||
					   strings.Contains(content, "DELETE")) &&
					   (strings.Contains(content, "query") ||
					   strings.Contains(content, "findOne") ||
					   strings.Contains(content, "create"))

	// Check for async patterns (common in services)
	hasAsyncPatterns := strings.Contains(content, "async ") ||
						strings.Contains(content, "await ") ||
						strings.Contains(content, ".then(") ||
						strings.Contains(content, "Promise")

	// Don't classify as service if it's mostly constants/configuration
	hasConstantPatterns := strings.Contains(content, "const API_") ||
						   strings.Contains(content, "const HTTP_") ||
						   strings.Contains(content, "ENDPOINTS") ||
						   strings.Contains(content, "STATUS")

	isServiceType := (hasServiceKeywords || hasHTTPClient || hasDBOperations) && !hasConstantPatterns

	if isServiceType {
		metadata["has_http_client"] = hasHTTPClient
		metadata["has_db_operations"] = hasDBOperations
		metadata["has_async_patterns"] = hasAsyncPatterns
		metadata["detection_confidence"] = "high"
	}

	return isServiceType
}

// isUtility detects utility modules (pure functions, helpers)
func (ci *ComponentIdentifier) isUtility(content string, metadata map[string]any) bool {
	// Look for utility patterns - pure functions, no side effects
	hasUtilKeywords := strings.Contains(content, "util") ||
					   strings.Contains(content, "helper") ||
					   strings.Contains(content, "tool") ||
					   strings.Contains(content, "format") ||
					   strings.Contains(content, "validate")

	// Check for pure function patterns
	hasExportedFunctions := strings.Contains(content, "export function") ||
							strings.Contains(content, "export const") ||
							strings.Contains(content, "module.exports")

	// Avoid classification as utility if it has side effects
	hasSideEffects := strings.Contains(content, "console.") ||
					  strings.Contains(content, "document.") ||
					  strings.Contains(content, "window.") ||
					  strings.Contains(content, "localStorage") ||
					  strings.Contains(content, "fetch") ||
					  strings.Contains(content, "axios")

	isUtilityType := (hasUtilKeywords || hasExportedFunctions) && !hasSideEffects

	if isUtilityType {
		metadata["has_pure_functions"] = hasExportedFunctions && !hasSideEffects
		metadata["detection_confidence"] = "medium"
	}

	return isUtilityType
}

// isConfigurationModule detects configuration files
func (ci *ComponentIdentifier) isConfigurationModule(filePath string, content string, metadata map[string]any) bool {
	// Check file patterns
	fileName := filepath.Base(filePath)
	hasConfigName := strings.Contains(fileName, "config") ||
					 strings.Contains(fileName, "setting") ||
					 strings.Contains(fileName, "constant") ||
					 strings.Contains(fileName, "env")

	// Check for configuration content patterns
	hasConfigContent := strings.Contains(content, "export const config") ||
						strings.Contains(content, "module.exports") ||
						strings.Contains(content, "process.env") ||
						strings.Contains(content, "NODE_ENV") ||
						strings.Contains(content, "API_URL") ||
						strings.Contains(content, "DATABASE_URL")

	// Check for constant patterns - important for constants files
	hasConstantPatterns := strings.Contains(content, "const API_") ||
						   strings.Contains(content, "const HTTP_") ||
						   strings.Contains(content, "ENDPOINTS") ||
						   strings.Contains(content, "STATUS") ||
						   strings.Contains(content, "VALIDATION_RULES") ||
						   (strings.Contains(content, "export const") && 
						    strings.Contains(content, ": {"))

	// Check file extensions commonly used for config
	hasConfigExt := strings.HasSuffix(fileName, ".config.js") ||
					strings.HasSuffix(fileName, ".config.ts") ||
					strings.HasSuffix(fileName, ".env") ||
					strings.HasSuffix(fileName, ".json")

	isConfigType := hasConfigName || hasConfigContent || hasConfigExt || hasConstantPatterns

	if isConfigType {
		metadata["has_env_vars"] = strings.Contains(content, "process.env")
		metadata["is_json"] = strings.HasSuffix(fileName, ".json")
		metadata["detection_confidence"] = "high"
	}

	return isConfigType
}

// isMiddleware detects middleware patterns
func (ci *ComponentIdentifier) isMiddleware(content string, metadata map[string]any) bool {
	// Check for middleware keywords and patterns
	hasMiddlewareKeywords := strings.Contains(content, "middleware") ||
							 strings.Contains(content, "Middleware") ||
							 strings.Contains(content, "next()")

	// Check for Express.js middleware patterns - more specific
	hasExpressPattern := (strings.Contains(content, "(req, res, next)") ||
						 (strings.Contains(content, "req") &&
						  strings.Contains(content, "res") &&
						  strings.Contains(content, "next") &&
						  (strings.Contains(content, "req.headers") ||
						   strings.Contains(content, "res.send") ||
						   strings.Contains(content, "res.status"))))

	// Check for authentication middleware patterns
	hasAuthPattern := strings.Contains(content, "auth") &&
					  (strings.Contains(content, "token") ||
					   strings.Contains(content, "jwt") ||
					   strings.Contains(content, "bearer")) &&
					  strings.Contains(content, "req.headers")

	isMiddlewareType := hasMiddlewareKeywords || hasExpressPattern || hasAuthPattern

	if isMiddlewareType {
		metadata["is_express_middleware"] = hasExpressPattern
		metadata["is_auth_middleware"] = hasAuthPattern
		metadata["detection_confidence"] = "high"
	}

	return isMiddlewareType
}

// extractExports extracts exported functions/variables from the content
func (ci *ComponentIdentifier) extractExports(content string) []string {
	exports := make([]string, 0)

	// Simple regex-based extraction for common export patterns
	// This is a simplified version - in production, you'd use AST parsing
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Export function patterns
		if strings.HasPrefix(line, "export function ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				funcName := strings.Split(parts[2], "(")[0]
				exports = append(exports, funcName)
			}
		}
		
		// Export const patterns
		if strings.HasPrefix(line, "export const ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				constName := strings.Split(parts[2], " ")[0]
				constName = strings.Split(constName, "=")[0]
				exports = append(exports, constName)
			}
		}
		
		// Export default patterns
		if strings.HasPrefix(line, "export default ") {
			exports = append(exports, "default")
		}
		
		// Module.exports patterns (CommonJS)
		if strings.Contains(line, "module.exports") {
			exports = append(exports, "module")
		}
	}

	return exports
}

// extractDependencies extracts import dependencies from the content
func (ci *ComponentIdentifier) extractDependencies(content string) []string {
	dependencies := make([]string, 0)

	// Simple regex-based extraction for common import patterns
	lines := strings.Split(content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// ES6 import patterns
		if strings.HasPrefix(line, "import ") {
			// Extract module name from import statement
			if strings.Contains(line, "from ") {
				parts := strings.Split(line, "from ")
				if len(parts) >= 2 {
					module := strings.TrimSpace(parts[1])
					// Remove quotes and semicolon
					module = strings.Trim(module, " \t\n\r")
					module = strings.TrimSuffix(module, ";")
					module = strings.Trim(module, "'\"")
					if module != "" {
						dependencies = append(dependencies, module)
					}
				}
			} else {
				// Handle import 'module'; patterns (like CSS imports)
				parts := strings.Split(line, "import ")
				if len(parts) >= 2 {
					module := strings.TrimSpace(parts[1])
					module = strings.TrimSuffix(module, ";")
					module = strings.Trim(module, "'\"")
					if module != "" {
						dependencies = append(dependencies, module)
					}
				}
			}
		}
		
		// CommonJS require patterns
		if strings.Contains(line, "require(") {
			start := strings.Index(line, "require(")
			if start != -1 {
				remaining := line[start+8:] // Skip "require("
				end := strings.Index(remaining, ")")
				if end != -1 {
					module := remaining[:end]
					module = strings.Trim(module, " \t\n\r'\"")
					if module != "" {
						dependencies = append(dependencies, module)
					}
				}
			}
		}
	}

	return dependencies
}

// extractComponentName extracts a meaningful component name from the file path
func extractComponentName(filePath string) string {
	fileName := filepath.Base(filePath)
	
	// Remove file extension
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	
	// Remove common suffixes
	name = strings.TrimSuffix(name, ".component")
	name = strings.TrimSuffix(name, ".service")
	name = strings.TrimSuffix(name, ".util")
	name = strings.TrimSuffix(name, ".config")
	name = strings.TrimSuffix(name, ".middleware")
	
	// Handle index files by using parent directory name
	if name == "index" {
		parentDir := filepath.Base(filepath.Dir(filePath))
		if parentDir != "." && parentDir != "/" {
			name = parentDir
		}
	}

	return name
}

// GetComponents returns all identified components
func (ci *ComponentIdentifier) GetComponents() []Component {
	return ci.components
}

// AddComponent adds a component to the collection
func (ci *ComponentIdentifier) AddComponent(component Component) {
	ci.components = append(ci.components, component)
	ci.cache[component.FilePath] = component
}

// GetComponentStats returns statistics about identified components
func (ci *ComponentIdentifier) GetComponentStats() map[ComponentType]int {
	stats := make(map[ComponentType]int)
	
	for _, component := range ci.components {
		stats[component.Type]++
	}
	
	return stats
}

// ExportToJSON exports all components to JSON format
func (ci *ComponentIdentifier) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(ci.components, "", "  ")
}