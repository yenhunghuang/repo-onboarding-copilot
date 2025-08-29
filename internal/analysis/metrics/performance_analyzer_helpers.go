package metrics

import (
	"strings"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// Helper functions for AST-based performance analysis

// containsQueryPattern checks if a function name suggests database query operations
func (pa *PerformanceAnalyzer) containsQueryPattern(functionName string) bool {
	queryKeywords := []string{
		"find", "query", "select", "get", "fetch", "load", "retrieve", "search",
		"findOne", "findMany", "findById", "findBy", "getById", "getBy",
		"loadUser", "loadData", "fetchUser", "fetchData", "searchUser",
	}
	
	nameLower := strings.ToLower(functionName)
	for _, keyword := range queryKeywords {
		if strings.Contains(nameLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// containsArrayIteration checks if a function name suggests array iteration
func (pa *PerformanceAnalyzer) containsArrayIteration(functionName string) bool {
	iterationKeywords := []string{
		"process", "handle", "iterate", "loop", "forEach", "map", "filter",
		"transform", "convert", "parse", "validate", "check", "update", "modify",
		"processAll", "handleAll", "updateAll", "validateAll", "checkAll",
	}
	
	nameLower := strings.ToLower(functionName)
	for _, keyword := range iterationKeywords {
		if strings.Contains(nameLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// isLikelyInLoop checks if a function is likely to be called within a loop
func (pa *PerformanceAnalyzer) isLikelyInLoop(function ast.FunctionInfo) bool {
	// Functions with single parameters are more likely to be called in loops
	if len(function.Parameters) == 1 {
		param := function.Parameters[0]
		// Check for id-like parameters
		paramLower := strings.ToLower(param.Name)
		if strings.Contains(paramLower, "id") || strings.Contains(paramLower, "key") || 
		   strings.Contains(paramLower, "item") || strings.Contains(paramLower, "element") {
			return true
		}
	}
	
	// Functions with "ById" pattern are typically called in loops
	if strings.Contains(strings.ToLower(function.Name), "byid") {
		return true
	}
	
	return false
}

// hasNestedIterationPattern checks if function name suggests nested iterations
func (pa *PerformanceAnalyzer) hasNestedIterationPattern(functionName string) bool {
	nestedPatterns := []string{
		"nested", "double", "deep", "multi", "cross", "matrix", "grid",
		"compareAll", "matchAll", "crossRef", "intersect", "cartesian",
	}
	
	nameLower := strings.ToLower(functionName)
	for _, pattern := range nestedPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// containsReactPattern checks if function suggests React component patterns
func (pa *PerformanceAnalyzer) containsReactPattern(functionName string) bool {
	reactPatterns := []string{
		"render", "component", "hook", "use", "with", "hoc", "jsx",
		"useState", "useEffect", "useMemo", "useCallback", "createComponent",
	}
	
	nameLower := strings.ToLower(functionName)
	for _, pattern := range reactPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// isMemoryIntensiveFunction checks if function might be memory intensive
func (pa *PerformanceAnalyzer) isMemoryIntensiveFunction(function ast.FunctionInfo) bool {
	// Large functions are more likely to be memory intensive
	if function.EndLine - function.StartLine > 50 {
		return true
	}
	
	// Functions with many parameters might allocate more memory
	if len(function.Parameters) > 5 {
		return true
	}
	
	memoryIntensivePatterns := []string{
		"cache", "buffer", "store", "accumulate", "collect", "aggregate",
		"build", "create", "generate", "transform", "convert", "parse",
	}
	
	nameLower := strings.ToLower(function.Name)
	for _, pattern := range memoryIntensivePatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	
	return false
}

// hasBlockingPattern checks if function name suggests blocking operations
func (pa *PerformanceAnalyzer) hasBlockingPattern(functionName string) bool {
	blockingPatterns := []string{
		"sync", "block", "wait", "sleep", "lock", "mutex", "semaphore",
		"synchronous", "sequential", "blocking", "freeze",
	}
	
	nameLower := strings.ToLower(functionName)
	for _, pattern := range blockingPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	return false
}

// analyzeImportPerformanceImpact analyzes imports for performance impact
func (pa *PerformanceAnalyzer) analyzeImportPerformanceImpact(imports []ast.ImportInfo) []OptimizationOpportunity {
	var opportunities []OptimizationOpportunity
	
	heavyLibraries := map[string]struct{}{
		"lodash":    {},
		"moment":    {},
		"jquery":    {},
		"rxjs":      {},
		"three":     {},
		"d3":        {},
		"chartjs":   {},
		"bootstrap": {},
	}
	
	for _, imp := range imports {
		source := strings.ToLower(imp.Source)
		for heavyLib := range heavyLibraries {
			if strings.Contains(source, heavyLib) {
				opportunities = append(opportunities, OptimizationOpportunity{
					Type:        "bundle_optimization",
					Priority:    "medium",
					Description: "Heavy library detected",
					Impact:      "Bundle size reduction",
					Effort:      "medium",
					ROI:         75.0,
					Implementation: "Consider lighter alternatives or tree shaking",
					Evidence:    "Import: " + imp.Source,
				})
			}
		}
	}
	
	return opportunities
}

// calculateComplexityBasedScore calculates performance score based on function complexity
func (pa *PerformanceAnalyzer) calculateComplexityBasedScore(functions []ast.FunctionInfo) float64 {
	if len(functions) == 0 {
		return 100.0 // Perfect score for no functions
	}
	
	totalPenalty := 0.0
	for _, function := range functions {
		// Penalty based on function size
		lineCount := function.EndLine - function.StartLine + 1
		if lineCount > 100 {
			totalPenalty += 20.0
		} else if lineCount > 50 {
			totalPenalty += 10.0
		} else if lineCount > 20 {
			totalPenalty += 5.0
		}
		
		// Penalty for too many parameters
		if len(function.Parameters) > 7 {
			totalPenalty += 15.0
		} else if len(function.Parameters) > 4 {
			totalPenalty += 8.0
		}
		
		// Penalty for async functions (potential for inefficient async handling)
		if function.IsAsync {
			totalPenalty += 3.0
		}
	}
	
	// Average penalty per function, capped at 50 points
	averagePenalty := totalPenalty / float64(len(functions))
	if averagePenalty > 50.0 {
		averagePenalty = 50.0
	}
	
	return 100.0 - averagePenalty
}