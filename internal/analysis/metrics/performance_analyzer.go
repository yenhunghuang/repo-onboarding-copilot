package metrics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// PerformanceAnalyzer provides comprehensive performance anti-pattern detection and optimization analysis
type PerformanceAnalyzer struct {
	config PerformanceConfig
}

// PerformanceConfig defines configuration parameters for performance analysis
type PerformanceConfig struct {
	// Complexity thresholds for performance assessment
	NestedLoopThreshold    int `yaml:"nested_loop_threshold" default:"2"`
	QueryPatternThreshold  int `yaml:"query_pattern_threshold" default:"3"`
	DOMAccessThreshold     int `yaml:"dom_access_threshold" default:"5"`
	BundleSizeThresholdKB  int `yaml:"bundle_size_threshold_kb" default:"500"`
	ComponentComplexityMax int `yaml:"component_complexity_max" default:"15"`

	// Performance impact weights
	AlgorithmicWeight float64 `yaml:"algorithmic_weight" default:"0.35"`
	MemoryWeight      float64 `yaml:"memory_weight" default:"0.25"`
	NetworkWeight     float64 `yaml:"network_weight" default:"0.20"`
	RenderWeight      float64 `yaml:"render_weight" default:"0.15"`
	BundleWeight      float64 `yaml:"bundle_weight" default:"0.05"`
}

// PerformanceMetrics contains comprehensive performance analysis results
type PerformanceMetrics struct {
	OverallScore              float64                     `json:"overall_score"`
	PerformanceGrade          string                      `json:"performance_grade"`
	AntiPatterns              []AntiPattern               `json:"anti_patterns"`
	Bottlenecks               []PerformanceBottleneck     `json:"bottlenecks"`
	OptimizationOpportunities []OptimizationOpportunity   `json:"optimization_opportunities"`
	BundleAnalysis            *BundleAnalysis             `json:"bundle_analysis"`
	ReactAnalysis             *ReactPerformanceAnalysis   `json:"react_analysis,omitempty"`
	FileAnalysis              []FilePerformanceAnalysis   `json:"file_analysis"`
	Summary                   PerformanceSummary          `json:"summary"`
	Recommendations           []PerformanceRecommendation `json:"recommendations"`
}

// AntiPattern represents a performance anti-pattern
type AntiPattern struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	FilePath    string            `json:"file_path"`
	StartLine   int               `json:"start_line,omitempty"`
	EndLine     int               `json:"end_line,omitempty"`
	Evidence    string            `json:"evidence"`
	Impact      PerformanceImpact `json:"impact"`
}

// PerformanceImpact describes the impact of a performance issue
type PerformanceImpact struct {
	Score         float64  `json:"score"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	AffectedAreas []string `json:"affected_areas"`
}

// PerformanceBottleneck represents a performance bottleneck
type PerformanceBottleneck struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	FilePath    string                 `json:"file_path"`
	StartLine   int                    `json:"start_line,omitempty"`
	EndLine     int                    `json:"end_line,omitempty"`
	Description string                 `json:"description"`
	Impact      BottleneckImpact       `json:"impact"`
	Solution    OptimizationSuggestion `json:"solution"`
}

// BottleneckImpact describes bottleneck impact
type BottleneckImpact struct {
	PerformanceHit  string `json:"performance_hit"`
	ScalabilityRisk string `json:"scalability_risk"`
	MaintenanceCost string `json:"maintenance_cost"`
}

// OptimizationSuggestion provides optimization guidance
type OptimizationSuggestion struct {
	Approach     string  `json:"approach"`
	Explanation  string  `json:"explanation"`
	ExpectedGain float64 `json:"expected_gain"`
	Effort       string  `json:"effort"`
	Priority     string  `json:"priority"`
}

// OptimizationOpportunity represents an opportunity for performance improvement
type OptimizationOpportunity struct {
	Type           string  `json:"type"`
	Priority       string  `json:"priority"`
	Description    string  `json:"description"`
	Impact         string  `json:"impact"`
	Effort         string  `json:"effort"`
	ROI            float64 `json:"roi"`
	Implementation string  `json:"implementation"`
	Evidence       string  `json:"evidence"`
}

// BundleAnalysis contains bundle size analysis
type BundleAnalysis struct {
	EstimatedSizeKB   int                `json:"estimated_size_kb"`
	HeavyDependencies []HeavyDependency  `json:"heavy_dependencies"`
	OptimizationTips  []string           `json:"optimization_tips"`
	TreeShakingIssues []TreeShakingIssue `json:"tree_shaking_issues"`
}

// HeavyDependency represents a heavy library dependency
type HeavyDependency struct {
	Name            string   `json:"name"`
	Source          string   `json:"source"`
	FilePath        string   `json:"file_path"`
	EstimatedSizeKB int      `json:"estimated_size_kb"`
	ImportType      string   `json:"import_type"`
	Usage           string   `json:"usage"`
	Alternatives    []string `json:"alternatives"`
}

// TreeShakingIssue represents an issue preventing tree-shaking
type TreeShakingIssue struct {
	FilePath         string `json:"file_path"`
	ImportSource     string `json:"import_source"`
	Issue            string `json:"issue"`
	Suggestion       string `json:"suggestion"`
	PotentialSavings int    `json:"potential_savings_kb"`
}

// ReactPerformanceAnalysis contains React-specific performance analysis
type ReactPerformanceAnalysis struct {
	ComponentIssues       []ReactComponentIssue  `json:"component_issues"`
	HookIssues            []ReactHookIssue       `json:"hook_issues"`
	RenderOptimizations   []RenderOptimization   `json:"render_optimizations"`
	StateManagementIssues []StateManagementIssue `json:"state_management_issues"`
}

// ReactComponentIssue represents a React component performance issue
type ReactComponentIssue struct {
	ComponentName string `json:"component_name"`
	FilePath      string `json:"file_path"`
	IssueType     string `json:"issue_type"`
	Description   string `json:"description"`
	Severity      string `json:"severity"`
	StartLine     int    `json:"start_line"`
	EndLine       int    `json:"end_line"`
	Suggestion    string `json:"suggestion"`
}

// ReactHookIssue represents a React hook performance issue
type ReactHookIssue struct {
	HookName    string `json:"hook_name"`
	FilePath    string `json:"file_path"`
	IssueType   string `json:"issue_type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion"`
}

// RenderOptimization represents a render optimization opportunity
type RenderOptimization struct {
	ComponentName    string  `json:"component_name"`
	FilePath         string  `json:"file_path"`
	OptimizationType string  `json:"optimization_type"`
	Description      string  `json:"description"`
	ExpectedGain     float64 `json:"expected_gain"`
	Implementation   string  `json:"implementation"`
	Priority         string  `json:"priority"`
}

// StateManagementIssue represents a state management performance issue
type StateManagementIssue struct {
	IssueType   string `json:"issue_type"`
	FilePath    string `json:"file_path"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion"`
}

// FilePerformanceAnalysis contains per-file performance analysis
type FilePerformanceAnalysis struct {
	FilePath        string   `json:"file_path"`
	IssueCount      int      `json:"issue_count"`
	WorstSeverity   string   `json:"worst_severity"`
	Recommendations []string `json:"recommendations"`
}

// PerformanceSummary contains overall performance summary
type PerformanceSummary struct {
	TotalAntiPatterns     int     `json:"total_anti_patterns"`
	CriticalIssues        int     `json:"critical_issues"`
	HighPriorityIssues    int     `json:"high_priority_issues"`
	OptimizationPotential float64 `json:"optimization_potential"`
	TopRecommendation     string  `json:"top_recommendation"`
}

// PerformanceRecommendation represents an actionable performance recommendation
type PerformanceRecommendation struct {
	Priority       string `json:"priority"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Action         string `json:"action"`
	ExpectedImpact string `json:"expected_impact"`
	TimeFrame      string `json:"time_frame"`
}

// NewPerformanceAnalyzer creates a new performance analyzer with default configuration
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return NewPerformanceAnalyzerWithConfig(PerformanceConfig{
		NestedLoopThreshold:    2,
		QueryPatternThreshold:  3,
		DOMAccessThreshold:     5,
		BundleSizeThresholdKB:  500,
		ComponentComplexityMax: 15,
		AlgorithmicWeight:      0.35,
		MemoryWeight:           0.25,
		NetworkWeight:          0.20,
		RenderWeight:           0.15,
		BundleWeight:           0.05,
	})
}

// NewPerformanceAnalyzerWithConfig creates a performance analyzer with custom configuration
func NewPerformanceAnalyzerWithConfig(config PerformanceConfig) *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		config: config,
	}
}

// AnalyzePerformance performs comprehensive performance analysis on parsed results
func (pa *PerformanceAnalyzer) AnalyzePerformance(ctx context.Context, parseResults []*ast.ParseResult, complexityMetrics *ComplexityMetrics) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		AntiPatterns:              []AntiPattern{},
		Bottlenecks:               []PerformanceBottleneck{},
		OptimizationOpportunities: []OptimizationOpportunity{},
		FileAnalysis:              []FilePerformanceAnalysis{},
		Recommendations:           []PerformanceRecommendation{},
	}

	// Detect anti-patterns using AST analysis
	pa.detectAntiPatternsAST(parseResults, metrics)

	// Identify performance bottlenecks
	pa.identifyBottlenecks(parseResults, complexityMetrics, metrics)

	// Analyze bundle size impact
	pa.analyzeBundleSize(parseResults, metrics)

	// Perform React-specific analysis if applicable
	pa.analyzeReactPerformance(parseResults, metrics)

	// Generate optimization opportunities
	pa.generateOptimizationOpportunities(parseResults, complexityMetrics, metrics)

	// Calculate overall performance score
	pa.calculatePerformanceScore(metrics)

	// Generate summary and recommendations
	pa.generateSummaryAndRecommendations(metrics)

	return metrics, nil
}

// detectAntiPatternsAST identifies anti-patterns using AST analysis instead of regex
func (pa *PerformanceAnalyzer) detectAntiPatternsAST(parseResults []*ast.ParseResult, metrics *PerformanceMetrics) {
	for _, result := range parseResults {
		// N+1 Query Pattern Detection
		pa.detectNPlusOneQueriesAST(result, metrics)

		// Synchronous Loop Detection
		pa.detectSynchronousLoopsAST(result, metrics)

		// Memory Leak Patterns
		pa.detectMemoryLeaksAST(result, metrics)

		// Nested Loop Performance Issues
		pa.detectNestedLoopsAST(result, metrics)

		// Large Function Detection
		pa.detectLargeFunctions(result, metrics)

		// Repeated DOM Queries
		pa.detectRepeatedDOMQueriesAST(result, metrics)

		// Inefficient String Operations
		pa.detectStringInefficienciesAST(result, metrics)

		// Blocking Operations
		pa.detectBlockingOperationsAST(result, metrics)
	}
}

// detectNPlusOneQueriesAST identifies N+1 query patterns using AST analysis
func (pa *PerformanceAnalyzer) detectNPlusOneQueriesAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Check if function name suggests query operations
		if pa.containsQueryPattern(function.Name) {
			// Check if function is likely to be called in loops
			if pa.isLikelyInLoop(function) {
				antiPattern := AntiPattern{
					Type:        "n_plus_one_query",
					Description: fmt.Sprintf("Potential N+1 query pattern in function '%s'", function.Name),
					Severity:    "high",
					FilePath:    result.FilePath,
					StartLine:   function.StartLine,
					EndLine:     function.EndLine,
					Evidence:    fmt.Sprintf("Function %s contains query patterns and may be called in loops", function.Name),
					Impact: PerformanceImpact{
						Score:         85,
						Category:      "database",
						Description:   "Database queries in loops cause exponential performance degradation",
						AffectedAreas: []string{"database", "network", "response_time"},
					},
				}
				metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
			}
		}

		// Check for async functions with array parameters (potential sequential operations)
		if function.IsAsync && pa.containsQueryPattern(function.Name) {
			for _, param := range function.Parameters {
				if strings.Contains(strings.ToLower(param.Type), "array") || strings.Contains(strings.ToLower(param.Name), "list") {
					antiPattern := AntiPattern{
						Type:        "sequential_async_queries",
						Description: fmt.Sprintf("Async function '%s' may execute sequential queries on array parameter", function.Name),
						Severity:    "high",
						FilePath:    result.FilePath,
						StartLine:   function.StartLine,
						EndLine:     function.EndLine,
						Evidence:    fmt.Sprintf("Async function %s with array parameter %s", function.Name, param.Name),
						Impact: PerformanceImpact{
							Score:         80,
							Category:      "async",
							Description:   "Sequential async queries should be parallelized with Promise.all",
							AffectedAreas: []string{"parallelization", "response_time"},
						},
					}
					metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
					break // Only report once per function
				}
			}
		}
	}
}

// detectSynchronousLoopsAST identifies synchronous operations in loops using AST analysis
func (pa *PerformanceAnalyzer) detectSynchronousLoopsAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Analyze async functions for potential synchronous loop patterns
		if function.IsAsync {
			// Async functions with iteration-related names may contain sync-in-loop patterns
			if pa.containsArrayIteration(function.Name) {
				antiPattern := AntiPattern{
					Type:        "sync_in_loop",
					Description: fmt.Sprintf("Potential synchronous operations in loop in function '%s'", function.Name),
					Severity:    "high",
					FilePath:    result.FilePath,
					StartLine:   function.StartLine,
					EndLine:     function.EndLine,
					Evidence:    fmt.Sprintf("Async function %s may contain sequential async operations", function.Name),
					Impact: PerformanceImpact{
						Score:         75,
						Category:      "async",
						Description:   "Sequential async operations prevent parallelization",
						AffectedAreas: []string{"response_time", "throughput", "parallelization"},
					},
				}
				metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
			}
		}

		// Check for functions with names suggesting nested iterations
		if pa.hasNestedIterationPattern(function.Name) {
			antiPattern := AntiPattern{
				Type:        "nested_iteration",
				Description: fmt.Sprintf("Potential nested iterations in function '%s'", function.Name),
				Severity:    "medium",
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Function %s may contain nested loops", function.Name),
				Impact: PerformanceImpact{
					Score:         60,
					Category:      "algorithmic",
					Description:   "Nested iterations increase complexity to O(n²)",
					AffectedAreas: []string{"cpu", "time_complexity"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// detectMemoryLeaksAST identifies potential memory leak patterns using AST analysis
func (pa *PerformanceAnalyzer) detectMemoryLeaksAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Check for memory-intensive functions without proper cleanup
		if pa.isMemoryIntensiveFunction(function) {
			antiPattern := AntiPattern{
				Type:        "potential_memory_leak",
				Description: fmt.Sprintf("Function '%s' may have memory leak patterns", function.Name),
				Severity:    "medium",
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Large or complex function %s may not properly clean up resources", function.Name),
				Impact: PerformanceImpact{
					Score:         50,
					Category:      "memory",
					Description:   "Functions with potential memory leaks cause gradual performance degradation",
					AffectedAreas: []string{"memory", "stability"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}

	// Check for event listener patterns in imports
	for _, imp := range result.Imports {
		if strings.Contains(strings.ToLower(imp.Source), "event") ||
			strings.Contains(strings.ToLower(imp.Source), "listener") {
			antiPattern := AntiPattern{
				Type:        "event_listener_risk",
				Description: "Event listener imports detected - ensure proper cleanup",
				Severity:    "low",
				FilePath:    result.FilePath,
				StartLine:   imp.StartLine,
				EndLine:     imp.StartLine,
				Evidence:    fmt.Sprintf("Import: %s", imp.Source),
				Impact: PerformanceImpact{
					Score:         30,
					Category:      "memory",
					Description:   "Event listeners without proper cleanup can cause memory leaks",
					AffectedAreas: []string{"memory", "performance"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// detectNestedLoopsAST identifies nested loop patterns using AST analysis
func (pa *PerformanceAnalyzer) detectNestedLoopsAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Check for functions with nested iteration patterns
		if pa.hasNestedIterationPattern(function.Name) {
			// Estimate nesting depth based on function complexity
			nestingDepth := 2 // Default assumption for nested pattern names
			if function.EndLine-function.StartLine > 100 {
				nestingDepth = 3 // Larger functions likely have deeper nesting
			}

			severity := pa.calculateNestedLoopSeverity(nestingDepth)
			antiPattern := AntiPattern{
				Type:        "nested_loops",
				Description: fmt.Sprintf("Nested loops detected in function '%s' (estimated depth: %d)", function.Name, nestingDepth),
				Severity:    severity,
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Function %s suggests nested iteration pattern", function.Name),
				Impact: PerformanceImpact{
					Score:         pa.getNestedLoopScoreImpact(nestingDepth),
					Category:      "algorithmic",
					Description:   fmt.Sprintf("Nested loops with depth %d result in O(n^%d) complexity", nestingDepth, nestingDepth),
					AffectedAreas: []string{"cpu", "time_complexity", "scalability"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// detectLargeFunctions identifies functions that are too large and may impact performance
func (pa *PerformanceAnalyzer) detectLargeFunctions(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		lineCount := function.EndLine - function.StartLine + 1

		if lineCount > 100 {
			severity := "high"
			if lineCount > 200 {
				severity = "critical"
			}

			antiPattern := AntiPattern{
				Type:        "large_function",
				Description: fmt.Sprintf("Function '%s' is too large (%d lines)", function.Name, lineCount),
				Severity:    severity,
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Function %s spans %d lines", function.Name, lineCount),
				Impact: PerformanceImpact{
					Score:         float64(40 + (lineCount / 10)), // Scale with function size
					Category:      "maintainability",
					Description:   "Large functions are harder to optimize and may contain performance bottlenecks",
					AffectedAreas: []string{"maintainability", "optimization", "testing"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// detectRepeatedDOMQueriesAST identifies potential repeated DOM queries using AST analysis
func (pa *PerformanceAnalyzer) detectRepeatedDOMQueriesAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	domQueryFunctions := map[string]int{}

	for _, function := range result.Functions {
		// Check for DOM query-related function names
		nameLower := strings.ToLower(function.Name)
		if strings.Contains(nameLower, "query") || strings.Contains(nameLower, "select") ||
			strings.Contains(nameLower, "element") || strings.Contains(nameLower, "dom") {
			domQueryFunctions[function.Name]++
		}
	}

	// Report if we find multiple functions that might be doing similar DOM queries
	if len(domQueryFunctions) > pa.config.DOMAccessThreshold {
		antiPattern := AntiPattern{
			Type:        "repeated_dom_queries",
			Description: fmt.Sprintf("Multiple DOM query functions detected (%d functions)", len(domQueryFunctions)),
			Severity:    "medium",
			FilePath:    result.FilePath,
			Evidence:    fmt.Sprintf("Found %d functions with DOM query patterns", len(domQueryFunctions)),
			Impact: PerformanceImpact{
				Score:         45,
				Category:      "dom",
				Description:   "Repeated DOM queries can be cached or optimized",
				AffectedAreas: []string{"rendering", "ui_performance"},
			},
		}
		metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
	}
}

// detectStringInefficienciesAST identifies string operation inefficiencies using AST analysis
func (pa *PerformanceAnalyzer) detectStringInefficienciesAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Check for functions with names suggesting string manipulation in loops
		nameLower := strings.ToLower(function.Name)
		if (strings.Contains(nameLower, "concat") || strings.Contains(nameLower, "append") ||
			strings.Contains(nameLower, "build") || strings.Contains(nameLower, "join")) &&
			pa.containsArrayIteration(function.Name) {

			antiPattern := AntiPattern{
				Type:        "string_concatenation_in_loop",
				Description: fmt.Sprintf("Potential inefficient string operations in function '%s'", function.Name),
				Severity:    "medium",
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Function %s suggests string manipulation in loops", function.Name),
				Impact: PerformanceImpact{
					Score:         40,
					Category:      "algorithmic",
					Description:   "String concatenation in loops creates multiple temporary objects",
					AffectedAreas: []string{"memory", "cpu", "gc_pressure"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// detectBlockingOperationsAST identifies blocking operations using AST analysis
func (pa *PerformanceAnalyzer) detectBlockingOperationsAST(result *ast.ParseResult, metrics *PerformanceMetrics) {
	for _, function := range result.Functions {
		// Check for functions with blocking patterns but are not async
		if pa.hasBlockingPattern(function.Name) && !function.IsAsync {
			antiPattern := AntiPattern{
				Type:        "blocking_operation",
				Description: fmt.Sprintf("Potential blocking operation in non-async function '%s'", function.Name),
				Severity:    "high",
				FilePath:    result.FilePath,
				StartLine:   function.StartLine,
				EndLine:     function.EndLine,
				Evidence:    fmt.Sprintf("Non-async function %s suggests blocking operations", function.Name),
				Impact: PerformanceImpact{
					Score:         70,
					Category:      "blocking",
					Description:   "Blocking operations in main thread cause UI freezes",
					AffectedAreas: []string{"ui_responsiveness", "user_experience"},
				},
			}
			metrics.AntiPatterns = append(metrics.AntiPatterns, antiPattern)
		}
	}
}

// Helper functions for severity calculation
func (pa *PerformanceAnalyzer) calculateNestedLoopSeverity(depth int) string {
	switch {
	case depth >= 4:
		return "critical"
	case depth >= 3:
		return "high"
	case depth >= 2:
		return "medium"
	default:
		return "low"
	}
}

func (pa *PerformanceAnalyzer) getNestedLoopScoreImpact(depth int) float64 {
	// Exponential impact based on nesting depth
	return math.Min(90.0, 30.0*math.Pow(1.5, float64(depth)))
}

// identifyBottlenecks identifies performance bottlenecks using complexity metrics and AST analysis
func (pa *PerformanceAnalyzer) identifyBottlenecks(parseResults []*ast.ParseResult, complexityMetrics *ComplexityMetrics, metrics *PerformanceMetrics) {
	var bottlenecks []PerformanceBottleneck

	// Analyze based on complexity metrics if available
	if complexityMetrics != nil {
		for _, funcMetrics := range complexityMetrics.FunctionMetrics {
			if funcMetrics.WeightedScore > 20.0 { // High complexity threshold
				bottleneck := PerformanceBottleneck{
					Type:        "high_complexity_function",
					Severity:    pa.getComplexityBasedSeverity(funcMetrics.WeightedScore),
					FilePath:    funcMetrics.FilePath,
					StartLine:   funcMetrics.StartLine,
					EndLine:     funcMetrics.EndLine,
					Description: fmt.Sprintf("High complexity function '%s' (score: %.1f)", funcMetrics.Name, funcMetrics.WeightedScore),
					Impact: BottleneckImpact{
						PerformanceHit:  pa.calculateComplexityPerformanceHit(funcMetrics.WeightedScore),
						ScalabilityRisk: pa.getComplexityScalabilityRisk(funcMetrics.WeightedScore),
						MaintenanceCost: "high",
					},
					Solution: OptimizationSuggestion{
						Approach:     "function_decomposition",
						Explanation:  "Break down the function into smaller, more focused functions",
						ExpectedGain: pa.calculateComplexityOptimizationGain(funcMetrics.WeightedScore),
						Effort:       "medium",
						Priority:     pa.getComplexityOptimizationPriority(funcMetrics.WeightedScore),
					},
				}
				bottlenecks = append(bottlenecks, bottleneck)
			}
		}
	}

	// Analyze based on AST patterns
	for _, result := range parseResults {
		// Check for large classes that might be bottlenecks
		for _, class := range result.Classes {
			if len(class.Methods) > 20 || (class.EndLine-class.StartLine) > 500 {
				bottleneck := PerformanceBottleneck{
					Type:        "large_class",
					Severity:    "medium",
					FilePath:    result.FilePath,
					StartLine:   class.StartLine,
					EndLine:     class.EndLine,
					Description: fmt.Sprintf("Large class '%s' with %d methods (%d lines)", class.Name, len(class.Methods), class.EndLine-class.StartLine),
					Impact: BottleneckImpact{
						PerformanceHit:  "medium",
						ScalabilityRisk: "high",
						MaintenanceCost: "high",
					},
					Solution: OptimizationSuggestion{
						Approach:     "class_decomposition",
						Explanation:  "Split the class into smaller, focused classes following Single Responsibility Principle",
						ExpectedGain: 60.0,
						Effort:       "high",
						Priority:     "medium",
					},
				}
				bottlenecks = append(bottlenecks, bottleneck)
			}
		}

		// Check for files with too many imports (possible architecture issues)
		if len(result.Imports) > 30 {
			bottleneck := PerformanceBottleneck{
				Type:        "high_coupling",
				Severity:    "low",
				FilePath:    result.FilePath,
				Description: fmt.Sprintf("File has %d imports, indicating high coupling", len(result.Imports)),
				Impact: BottleneckImpact{
					PerformanceHit:  "low",
					ScalabilityRisk: "medium",
					MaintenanceCost: "medium",
				},
				Solution: OptimizationSuggestion{
					Approach:     "dependency_reduction",
					Explanation:  "Reduce dependencies and improve module boundaries",
					ExpectedGain: 40.0,
					Effort:       "medium",
					Priority:     "low",
				},
			}
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	metrics.Bottlenecks = bottlenecks
}

// Helper functions for bottleneck analysis
func (pa *PerformanceAnalyzer) getComplexityBasedSeverity(score float64) string {
	switch {
	case score > 40.0:
		return "critical"
	case score > 30.0:
		return "high"
	case score > 20.0:
		return "medium"
	default:
		return "low"
	}
}

func (pa *PerformanceAnalyzer) calculateComplexityPerformanceHit(score float64) string {
	switch {
	case score > 40.0:
		return "critical"
	case score > 30.0:
		return "high"
	case score > 20.0:
		return "medium"
	default:
		return "low"
	}
}

func (pa *PerformanceAnalyzer) getComplexityScalabilityRisk(score float64) string {
	switch {
	case score > 35.0:
		return "critical"
	case score > 25.0:
		return "high"
	case score > 20.0:
		return "medium"
	default:
		return "low"
	}
}

func (pa *PerformanceAnalyzer) calculateComplexityOptimizationGain(score float64) float64 {
	return math.Min(80.0, score*2.0) // Cap at 80% gain
}

func (pa *PerformanceAnalyzer) getComplexityOptimizationPriority(score float64) string {
	switch {
	case score > 35.0:
		return "critical"
	case score > 25.0:
		return "high"
	case score > 20.0:
		return "medium"
	default:
		return "low"
	}
}

// analyzeBundleSize analyzes bundle size impact using AST analysis
func (pa *PerformanceAnalyzer) analyzeBundleSize(parseResults []*ast.ParseResult, metrics *PerformanceMetrics) {
	bundleAnalysis := &BundleAnalysis{
		EstimatedSizeKB:   0,
		HeavyDependencies: []HeavyDependency{},
		OptimizationTips:  []string{},
		TreeShakingIssues: []TreeShakingIssue{},
	}

	totalImports := 0
	heavyLibraries := map[string]int{
		"lodash":      70,
		"moment":      67,
		"jquery":      85,
		"rxjs":        45,
		"three":       600,
		"d3":          250,
		"chart.js":    60,
		"bootstrap":   150,
		"material-ui": 340,
		"antd":        2000,
		"react":       45,
		"vue":         35,
		"angular":     130,
	}

	for _, result := range parseResults {
		totalImports += len(result.Imports)

		for _, imp := range result.Imports {
			sourceLower := strings.ToLower(imp.Source)

			// Check for heavy libraries
			for lib, sizeKB := range heavyLibraries {
				if strings.Contains(sourceLower, lib) {
					heavyDep := HeavyDependency{
						Name:            lib,
						Source:          imp.Source,
						FilePath:        result.FilePath,
						EstimatedSizeKB: sizeKB,
						ImportType:      imp.ImportType,
						Usage:           pa.analyzeImportUsage(imp),
						Alternatives:    pa.suggestAlternatives(lib),
					}
					bundleAnalysis.HeavyDependencies = append(bundleAnalysis.HeavyDependencies, heavyDep)
					bundleAnalysis.EstimatedSizeKB += sizeKB
				}
			}

			// Check for tree-shaking issues
			if imp.ImportType == "default" && pa.isTreeShakingLibrary(imp.Source) {
				issue := TreeShakingIssue{
					FilePath:         result.FilePath,
					ImportSource:     imp.Source,
					Issue:            "Default import prevents tree-shaking",
					Suggestion:       "Use named imports to enable tree-shaking",
					PotentialSavings: pa.estimateTreeShakingSavings(imp.Source),
				}
				bundleAnalysis.TreeShakingIssues = append(bundleAnalysis.TreeShakingIssues, issue)
			}
		}
	}

	// Estimate base bundle size from total imports
	bundleAnalysis.EstimatedSizeKB += totalImports * 2 // Average 2KB per import

	// Generate optimization tips
	bundleAnalysis.OptimizationTips = pa.generateBundleOptimizationTips(bundleAnalysis)

	metrics.BundleAnalysis = bundleAnalysis
}

// analyzeImportUsage analyzes how an import is used
func (pa *PerformanceAnalyzer) analyzeImportUsage(imp ast.ImportInfo) string {
	if imp.ImportType == "default" {
		return "full_library"
	} else if imp.ImportType == "named" {
		if len(imp.Specifiers) == 1 {
			return "single_function"
		} else if len(imp.Specifiers) <= 3 {
			return "few_functions"
		} else {
			return "many_functions"
		}
	} else if imp.ImportType == "namespace" {
		return "entire_namespace"
	}
	return "unknown"
}

// suggestAlternatives suggests lighter alternatives for heavy libraries
func (pa *PerformanceAnalyzer) suggestAlternatives(library string) []string {
	alternatives := map[string][]string{
		"lodash":      {"ramda (functional)", "native ES6+ methods", "individual lodash functions"},
		"moment":      {"date-fns", "dayjs", "luxon"},
		"jquery":      {"native DOM APIs", "vanilla JS", "micro-libraries"},
		"three":       {"babylon.js (for specific use cases)", "custom WebGL"},
		"d3":          {"chart.js (for charts)", "native SVG", "lighter charting libs"},
		"bootstrap":   {"tailwindcss", "bulma", "custom CSS"},
		"material-ui": {"chakra-ui", "mantine", "custom components"},
		"antd":        {"chakra-ui", "mantine", "headless UI"},
	}

	if alts, exists := alternatives[library]; exists {
		return alts
	}
	return []string{"Consider lighter alternatives", "Implement needed functionality manually"}
}

// isTreeShakingLibrary checks if a library supports tree-shaking
func (pa *PerformanceAnalyzer) isTreeShakingLibrary(source string) bool {
	treeShakingLibs := []string{"lodash", "rxjs", "ramda", "date-fns", "material-ui"}
	sourceLower := strings.ToLower(source)
	for _, lib := range treeShakingLibs {
		if strings.Contains(sourceLower, lib) {
			return true
		}
	}
	return false
}

// estimateTreeShakingSavings estimates potential bundle size savings from proper tree-shaking
func (pa *PerformanceAnalyzer) estimateTreeShakingSavings(source string) int {
	// Rough estimates based on common libraries
	savings := map[string]int{
		"lodash":      60, // Can save ~60KB with proper tree-shaking
		"rxjs":        30,
		"material-ui": 200,
		"date-fns":    15,
	}

	sourceLower := strings.ToLower(source)
	for lib, saving := range savings {
		if strings.Contains(sourceLower, lib) {
			return saving
		}
	}
	return 10 // Default savings estimate
}

// generateBundleOptimizationTips generates tips for bundle optimization
func (pa *PerformanceAnalyzer) generateBundleOptimizationTips(analysis *BundleAnalysis) []string {
	var tips []string

	if analysis.EstimatedSizeKB > pa.config.BundleSizeThresholdKB {
		tips = append(tips, fmt.Sprintf("Bundle size (%dKB) exceeds threshold (%dKB)",
			analysis.EstimatedSizeKB, pa.config.BundleSizeThresholdKB))
	}

	if len(analysis.HeavyDependencies) > 0 {
		tips = append(tips, "Consider replacing heavy dependencies with lighter alternatives")
	}

	if len(analysis.TreeShakingIssues) > 0 {
		tips = append(tips, "Enable tree-shaking by using named imports instead of default imports")
	}

	// Add general tips
	tips = append(tips,
		"Use dynamic imports for code splitting",
		"Implement lazy loading for non-critical components",
		"Enable gzip compression on your server",
		"Consider using a CDN for heavy libraries",
		"Regularly audit and remove unused dependencies")

	return tips
}

// analyzeReactPerformance performs React-specific performance analysis
func (pa *PerformanceAnalyzer) analyzeReactPerformance(parseResults []*ast.ParseResult, metrics *PerformanceMetrics) {
	reactAnalysis := &ReactPerformanceAnalysis{
		ComponentIssues:       []ReactComponentIssue{},
		HookIssues:            []ReactHookIssue{},
		RenderOptimizations:   []RenderOptimization{},
		StateManagementIssues: []StateManagementIssue{},
	}

	isReactProject := false
	for _, result := range parseResults {
		for _, imp := range result.Imports {
			if strings.Contains(strings.ToLower(imp.Source), "react") {
				isReactProject = true
				break
			}
		}
		if isReactProject {
			break
		}
	}

	if !isReactProject {
		return // Skip React analysis for non-React projects
	}

	for _, result := range parseResults {
		// Analyze React components
		for _, function := range result.Functions {
			if pa.containsReactPattern(function.Name) {
				pa.analyzeReactComponent(function, result, reactAnalysis)
			}
		}

		for _, class := range result.Classes {
			if pa.isReactClassComponent(class) {
				pa.analyzeReactClassComponent(class, result, reactAnalysis)
			}
		}
	}

	if len(reactAnalysis.ComponentIssues) > 0 || len(reactAnalysis.HookIssues) > 0 {
		metrics.ReactAnalysis = reactAnalysis
	}
}

// analyzeReactComponent analyzes a React functional component
func (pa *PerformanceAnalyzer) analyzeReactComponent(function ast.FunctionInfo, result *ast.ParseResult, analysis *ReactPerformanceAnalysis) {
	// Check for large components
	componentSize := function.EndLine - function.StartLine + 1
	if componentSize > pa.config.ComponentComplexityMax*5 { // 75+ lines for default config
		issue := ReactComponentIssue{
			ComponentName: function.Name,
			FilePath:      result.FilePath,
			IssueType:     "large_component",
			Description:   fmt.Sprintf("Component is too large (%d lines)", componentSize),
			Severity:      "medium",
			StartLine:     function.StartLine,
			EndLine:       function.EndLine,
			Suggestion:    "Break down into smaller, focused components",
		}
		analysis.ComponentIssues = append(analysis.ComponentIssues, issue)
	}

	// Check for too many props (parameters)
	if len(function.Parameters) > 7 {
		issue := ReactComponentIssue{
			ComponentName: function.Name,
			FilePath:      result.FilePath,
			IssueType:     "too_many_props",
			Description:   fmt.Sprintf("Component has %d props", len(function.Parameters)),
			Severity:      "medium",
			StartLine:     function.StartLine,
			EndLine:       function.EndLine,
			Suggestion:    "Consider using props object or context for multiple related props",
		}
		analysis.ComponentIssues = append(analysis.ComponentIssues, issue)
	}

	// Suggest render optimization
	optimization := RenderOptimization{
		ComponentName:    function.Name,
		FilePath:         result.FilePath,
		OptimizationType: "memoization",
		Description:      "Consider using React.memo for performance optimization",
		ExpectedGain:     30.0,
		Implementation:   "Wrap component with React.memo if props don't change frequently",
		Priority:         pa.getOptimizationPriority(componentSize, len(function.Parameters)),
	}
	analysis.RenderOptimizations = append(analysis.RenderOptimizations, optimization)
}

// analyzeReactClassComponent analyzes a React class component
func (pa *PerformanceAnalyzer) analyzeReactClassComponent(class ast.ClassInfo, result *ast.ParseResult, analysis *ReactPerformanceAnalysis) {
	// Check for large class components
	componentSize := class.EndLine - class.StartLine + 1
	if componentSize > pa.config.ComponentComplexityMax*10 { // 150+ lines for default config
		issue := ReactComponentIssue{
			ComponentName: class.Name,
			FilePath:      result.FilePath,
			IssueType:     "large_class_component",
			Description:   fmt.Sprintf("Class component is very large (%d lines)", componentSize),
			Severity:      "high",
			StartLine:     class.StartLine,
			EndLine:       class.EndLine,
			Suggestion:    "Consider converting to functional component with hooks or breaking down into smaller components",
		}
		analysis.ComponentIssues = append(analysis.ComponentIssues, issue)
	}

	// Check for too many methods
	if len(class.Methods) > 15 {
		issue := ReactComponentIssue{
			ComponentName: class.Name,
			FilePath:      result.FilePath,
			IssueType:     "complex_class_component",
			Description:   fmt.Sprintf("Class component has %d methods", len(class.Methods)),
			Severity:      "medium",
			StartLine:     class.StartLine,
			EndLine:       class.EndLine,
			Suggestion:    "Consider extracting some methods to custom hooks or utility functions",
		}
		analysis.ComponentIssues = append(analysis.ComponentIssues, issue)
	}

	// Suggest modernization
	optimization := RenderOptimization{
		ComponentName:    class.Name,
		FilePath:         result.FilePath,
		OptimizationType: "modernization",
		Description:      "Consider converting class component to functional component with hooks",
		ExpectedGain:     25.0,
		Implementation:   "Use useState, useEffect, and other hooks to replace class lifecycle methods",
		Priority:         "medium",
	}
	analysis.RenderOptimizations = append(analysis.RenderOptimizations, optimization)
}

// isReactClassComponent checks if a class extends React.Component
func (pa *PerformanceAnalyzer) isReactClassComponent(class ast.ClassInfo) bool {
	extends := strings.ToLower(class.Extends)
	return strings.Contains(extends, "component") || strings.Contains(extends, "react")
}

// getOptimizationPriority determines optimization priority based on component characteristics
func (pa *PerformanceAnalyzer) getOptimizationPriority(size int, paramCount int) string {
	if size > 100 || paramCount > 10 {
		return "high"
	} else if size > 50 || paramCount > 5 {
		return "medium"
	}
	return "low"
}

// generateOptimizationOpportunities generates optimization opportunities based on analysis
func (pa *PerformanceAnalyzer) generateOptimizationOpportunities(parseResults []*ast.ParseResult, complexityMetrics *ComplexityMetrics, metrics *PerformanceMetrics) {
	var opportunities []OptimizationOpportunity

	// Generate opportunities from anti-patterns
	for _, antiPattern := range metrics.AntiPatterns {
		opportunity := pa.convertAntiPatternToOpportunity(antiPattern)
		opportunities = append(opportunities, opportunity)
	}

	// Generate opportunities from bottlenecks
	for _, bottleneck := range metrics.Bottlenecks {
		opportunity := OptimizationOpportunity{
			Type:           bottleneck.Type + "_optimization",
			Priority:       bottleneck.Solution.Priority,
			Description:    bottleneck.Description,
			Impact:         bottleneck.Solution.Explanation,
			Effort:         bottleneck.Solution.Effort,
			ROI:            bottleneck.Solution.ExpectedGain,
			Implementation: bottleneck.Solution.Approach,
			Evidence:       fmt.Sprintf("Detected in %s (lines %d-%d)", bottleneck.FilePath, bottleneck.StartLine, bottleneck.EndLine),
		}
		opportunities = append(opportunities, opportunity)
	}

	// Generate opportunities from imports analysis
	for _, result := range parseResults {
		importOpportunities := pa.analyzeImportPerformanceImpact(result.Imports)
		opportunities = append(opportunities, importOpportunities...)
	}

	// Generate opportunities based on function analysis
	for _, result := range parseResults {
		for _, function := range result.Functions {
			if opportunity := pa.analyzeFunctionForOptimization(function, result.FilePath); opportunity != nil {
				opportunities = append(opportunities, *opportunity)
			}
		}
	}

	// Sort opportunities by ROI and priority
	sort.Slice(opportunities, func(i, j int) bool {
		if opportunities[i].Priority != opportunities[j].Priority {
			return pa.getPriorityScore(opportunities[i].Priority) > pa.getPriorityScore(opportunities[j].Priority)
		}
		return opportunities[i].ROI > opportunities[j].ROI
	})

	metrics.OptimizationOpportunities = opportunities
}

// convertAntiPatternToOpportunity converts an anti-pattern to an optimization opportunity
func (pa *PerformanceAnalyzer) convertAntiPatternToOpportunity(antiPattern AntiPattern) OptimizationOpportunity {
	return OptimizationOpportunity{
		Type:           antiPattern.Type + "_fix",
		Priority:       antiPattern.Severity,
		Description:    antiPattern.Description,
		Impact:         antiPattern.Impact.Description,
		Effort:         pa.getEffortFromSeverity(antiPattern.Severity),
		ROI:            antiPattern.Impact.Score,
		Implementation: pa.getImplementationFromAntiPattern(antiPattern),
		Evidence:       antiPattern.Evidence,
	}
}

// analyzeFunctionForOptimization analyzes a function for optimization opportunities
func (pa *PerformanceAnalyzer) analyzeFunctionForOptimization(function ast.FunctionInfo, filePath string) *OptimizationOpportunity {
	// Check for optimization opportunities in async functions
	if function.IsAsync && len(function.Parameters) > 0 {
		// Async function with parameters might benefit from caching
		return &OptimizationOpportunity{
			Type:           "async_caching",
			Priority:       "medium",
			Description:    fmt.Sprintf("Async function '%s' might benefit from result caching", function.Name),
			Impact:         "Reduce redundant async operations",
			Effort:         "low",
			ROI:            40.0,
			Implementation: "Implement memoization or caching mechanism for frequently called async operations",
			Evidence:       fmt.Sprintf("Function %s in %s (lines %d-%d)", function.Name, filePath, function.StartLine, function.EndLine),
		}
	}

	// Check for functions with many parameters that might need refactoring
	if len(function.Parameters) > 8 {
		return &OptimizationOpportunity{
			Type:           "parameter_object",
			Priority:       "low",
			Description:    fmt.Sprintf("Function '%s' has too many parameters (%d)", function.Name, len(function.Parameters)),
			Impact:         "Improve readability and reduce parameter passing overhead",
			Effort:         "medium",
			ROI:            30.0,
			Implementation: "Replace multiple parameters with a parameter object or configuration object",
			Evidence:       fmt.Sprintf("Function %s has %d parameters", function.Name, len(function.Parameters)),
		}
	}

	return nil
}

// getPriorityScore converts priority string to numeric score for sorting
func (pa *PerformanceAnalyzer) getPriorityScore(priority string) int {
	switch priority {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// getEffortFromSeverity maps severity to effort level
func (pa *PerformanceAnalyzer) getEffortFromSeverity(severity string) string {
	switch severity {
	case "critical":
		return "high"
	case "high":
		return "medium"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

// getImplementationFromAntiPattern provides implementation guidance based on anti-pattern type
func (pa *PerformanceAnalyzer) getImplementationFromAntiPattern(antiPattern AntiPattern) string {
	implementations := map[string]string{
		"n_plus_one_query":             "Implement bulk queries or use data loaders to fetch related data in batches",
		"sequential_async_queries":     "Use Promise.all() to parallelize independent async operations",
		"sync_in_loop":                 "Replace sequential processing with Promise.all() for parallel execution",
		"nested_iteration":             "Consider using hash maps or more efficient algorithms to reduce complexity",
		"potential_memory_leak":        "Implement proper cleanup in useEffect cleanup functions or componentWillUnmount",
		"event_listener_risk":          "Ensure addEventListener is paired with removeEventListener",
		"nested_loops":                 "Optimize algorithm or use memoization to reduce computational complexity",
		"large_function":               "Break down function into smaller, focused functions",
		"repeated_dom_queries":         "Cache DOM query results or use refs in React components",
		"string_concatenation_in_loop": "Use array.join() or template literals instead of string concatenation",
		"blocking_operation":           "Convert to async operation or use web workers for heavy computations",
	}

	if impl, exists := implementations[antiPattern.Type]; exists {
		return impl
	}
	return "Implement best practices to resolve the identified performance issue"
}

// calculatePerformanceScore calculates overall performance score
func (pa *PerformanceAnalyzer) calculatePerformanceScore(metrics *PerformanceMetrics) {
	baseScore := 100.0

	// Deduct points for anti-patterns
	for _, antiPattern := range metrics.AntiPatterns {
		penalty := pa.getAntiPatternPenalty(antiPattern.Severity)
		baseScore -= penalty
	}

	// Deduct points for bottlenecks
	for _, bottleneck := range metrics.Bottlenecks {
		penalty := pa.getBottleneckPenalty(bottleneck.Severity)
		baseScore -= penalty
	}

	// Deduct points for bundle size if analysis is available
	if metrics.BundleAnalysis != nil {
		if metrics.BundleAnalysis.EstimatedSizeKB > pa.config.BundleSizeThresholdKB {
			excess := float64(metrics.BundleAnalysis.EstimatedSizeKB - pa.config.BundleSizeThresholdKB)
			penalty := math.Min(20.0, excess/50.0) // Max 20 points penalty for bundle size
			baseScore -= penalty
		}
	}

	// Ensure score is within bounds
	if baseScore < 0 {
		baseScore = 0
	}
	if baseScore > 100 {
		baseScore = 100
	}

	metrics.OverallScore = baseScore
	metrics.PerformanceGrade = pa.getPerformanceGrade(baseScore)
}

// getAntiPatternPenalty returns penalty score for anti-pattern severity
func (pa *PerformanceAnalyzer) getAntiPatternPenalty(severity string) float64 {
	penalties := map[string]float64{
		"critical": 15.0,
		"high":     10.0,
		"medium":   5.0,
		"low":      2.0,
	}

	if penalty, exists := penalties[severity]; exists {
		return penalty
	}
	return 5.0 // Default penalty
}

// getBottleneckPenalty returns penalty score for bottleneck severity
func (pa *PerformanceAnalyzer) getBottleneckPenalty(severity string) float64 {
	penalties := map[string]float64{
		"critical": 12.0,
		"high":     8.0,
		"medium":   4.0,
		"low":      2.0,
	}

	if penalty, exists := penalties[severity]; exists {
		return penalty
	}
	return 4.0 // Default penalty
}

// getPerformanceGrade converts numeric score to letter grade
func (pa *PerformanceAnalyzer) getPerformanceGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// generateSummaryAndRecommendations generates summary and recommendations
func (pa *PerformanceAnalyzer) generateSummaryAndRecommendations(metrics *PerformanceMetrics) {
	// Generate file analysis summary
	fileAnalysisMap := make(map[string]*FilePerformanceAnalysis)

	// Aggregate issues by file
	for _, antiPattern := range metrics.AntiPatterns {
		if analysis, exists := fileAnalysisMap[antiPattern.FilePath]; exists {
			analysis.IssueCount++
			analysis.WorstSeverity = pa.getWorseSeverity(analysis.WorstSeverity, antiPattern.Severity)
		} else {
			fileAnalysisMap[antiPattern.FilePath] = &FilePerformanceAnalysis{
				FilePath:        antiPattern.FilePath,
				IssueCount:      1,
				WorstSeverity:   antiPattern.Severity,
				Recommendations: []string{},
			}
		}
	}

	for _, bottleneck := range metrics.Bottlenecks {
		if analysis, exists := fileAnalysisMap[bottleneck.FilePath]; exists {
			analysis.IssueCount++
			analysis.WorstSeverity = pa.getWorseSeverity(analysis.WorstSeverity, bottleneck.Severity)
		} else {
			fileAnalysisMap[bottleneck.FilePath] = &FilePerformanceAnalysis{
				FilePath:        bottleneck.FilePath,
				IssueCount:      1,
				WorstSeverity:   bottleneck.Severity,
				Recommendations: []string{},
			}
		}
	}

	// Convert map to slice and add recommendations
	for filePath, analysis := range fileAnalysisMap {
		analysis.Recommendations = pa.generateFileRecommendations(filePath, analysis.IssueCount, analysis.WorstSeverity)
		metrics.FileAnalysis = append(metrics.FileAnalysis, *analysis)
	}

	// Sort file analysis by issue count
	sort.Slice(metrics.FileAnalysis, func(i, j int) bool {
		if metrics.FileAnalysis[i].IssueCount != metrics.FileAnalysis[j].IssueCount {
			return metrics.FileAnalysis[i].IssueCount > metrics.FileAnalysis[j].IssueCount
		}
		return pa.getPriorityScore(metrics.FileAnalysis[i].WorstSeverity) > pa.getPriorityScore(metrics.FileAnalysis[j].WorstSeverity)
	})

	// Generate summary
	metrics.Summary = PerformanceSummary{
		TotalAntiPatterns:     len(metrics.AntiPatterns),
		CriticalIssues:        pa.countIssuesBySeverity(metrics.AntiPatterns, "critical") + pa.countBottlenecksBySeverity(metrics.Bottlenecks, "critical"),
		HighPriorityIssues:    pa.countIssuesBySeverity(metrics.AntiPatterns, "high") + pa.countBottlenecksBySeverity(metrics.Bottlenecks, "high"),
		OptimizationPotential: pa.calculateOptimizationPotential(metrics),
		TopRecommendation:     pa.getTopRecommendation(metrics),
	}

	// Generate performance recommendations
	metrics.Recommendations = pa.generatePerformanceRecommendations(metrics)
}

// getWorseSeverity returns the worse of two severities
func (pa *PerformanceAnalyzer) getWorseSeverity(current, new string) string {
	if pa.getPriorityScore(new) > pa.getPriorityScore(current) {
		return new
	}
	return current
}

// generateFileRecommendations generates recommendations for a specific file
func (pa *PerformanceAnalyzer) generateFileRecommendations(filePath string, issueCount int, worstSeverity string) []string {
	var recommendations []string

	if issueCount > 5 {
		recommendations = append(recommendations, "Consider refactoring this file - high number of performance issues detected")
	}

	if worstSeverity == "critical" {
		recommendations = append(recommendations, "Address critical performance issues immediately")
	} else if worstSeverity == "high" {
		recommendations = append(recommendations, "Prioritize fixing high-severity performance issues")
	}

	recommendations = append(recommendations, "Run performance profiling to identify runtime bottlenecks")
	recommendations = append(recommendations, "Consider code review with focus on performance best practices")

	return recommendations
}

// countIssuesBySeverity counts anti-patterns by severity
func (pa *PerformanceAnalyzer) countIssuesBySeverity(antiPatterns []AntiPattern, severity string) int {
	count := 0
	for _, pattern := range antiPatterns {
		if pattern.Severity == severity {
			count++
		}
	}
	return count
}

// countBottlenecksBySeverity counts bottlenecks by severity
func (pa *PerformanceAnalyzer) countBottlenecksBySeverity(bottlenecks []PerformanceBottleneck, severity string) int {
	count := 0
	for _, bottleneck := range bottlenecks {
		if bottleneck.Severity == severity {
			count++
		}
	}
	return count
}

// calculateOptimizationPotential calculates the total optimization potential
func (pa *PerformanceAnalyzer) calculateOptimizationPotential(metrics *PerformanceMetrics) float64 {
	totalGain := 0.0
	for _, opportunity := range metrics.OptimizationOpportunities {
		totalGain += opportunity.ROI
	}

	// Cap at 100% and take average
	if len(metrics.OptimizationOpportunities) > 0 {
		averageGain := totalGain / float64(len(metrics.OptimizationOpportunities))
		return math.Min(100.0, averageGain)
	}

	return 0.0
}

// getTopRecommendation gets the most important recommendation
func (pa *PerformanceAnalyzer) getTopRecommendation(metrics *PerformanceMetrics) string {
	if len(metrics.OptimizationOpportunities) == 0 {
		return "No specific performance issues detected"
	}

	topOpportunity := metrics.OptimizationOpportunities[0] // Already sorted by priority and ROI
	return fmt.Sprintf("Priority: %s - %s", topOpportunity.Priority, topOpportunity.Description)
}

// generatePerformanceRecommendations generates actionable performance recommendations
func (pa *PerformanceAnalyzer) generatePerformanceRecommendations(metrics *PerformanceMetrics) []PerformanceRecommendation {
	var recommendations []PerformanceRecommendation

	// Critical issues first
	criticalCount := pa.countIssuesBySeverity(metrics.AntiPatterns, "critical") + pa.countBottlenecksBySeverity(metrics.Bottlenecks, "critical")
	if criticalCount > 0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Priority:       "critical",
			Category:       "immediate_action",
			Title:          "Address Critical Performance Issues",
			Description:    fmt.Sprintf("Found %d critical performance issues that need immediate attention", criticalCount),
			Action:         "Review and fix all critical anti-patterns and bottlenecks",
			ExpectedImpact: "High",
			TimeFrame:      "Immediate (within 1-2 days)",
		})
	}

	// Bundle optimization
	if metrics.BundleAnalysis != nil && metrics.BundleAnalysis.EstimatedSizeKB > pa.config.BundleSizeThresholdKB {
		recommendations = append(recommendations, PerformanceRecommendation{
			Priority:       "high",
			Category:       "bundle_optimization",
			Title:          "Optimize Bundle Size",
			Description:    fmt.Sprintf("Bundle size (%dKB) exceeds recommended threshold (%dKB)", metrics.BundleAnalysis.EstimatedSizeKB, pa.config.BundleSizeThresholdKB),
			Action:         "Implement tree-shaking, code splitting, and consider lighter alternatives for heavy dependencies",
			ExpectedImpact: "Medium to High",
			TimeFrame:      "1-2 weeks",
		})
	}

	// React optimization
	if metrics.ReactAnalysis != nil && len(metrics.ReactAnalysis.ComponentIssues) > 0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Priority:       "medium",
			Category:       "react_optimization",
			Title:          "Optimize React Components",
			Description:    fmt.Sprintf("Found %d React component issues", len(metrics.ReactAnalysis.ComponentIssues)),
			Action:         "Implement React.memo, useMemo, useCallback for expensive operations, and break down large components",
			ExpectedImpact: "Medium",
			TimeFrame:      "2-3 weeks",
		})
	}

	// General optimization
	if metrics.Summary.OptimizationPotential > 50.0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Priority:       "medium",
			Category:       "general_optimization",
			Title:          "Implement Performance Best Practices",
			Description:    fmt.Sprintf("%.1f%% optimization potential identified", metrics.Summary.OptimizationPotential),
			Action:         "Follow the optimization opportunities listed in the analysis",
			ExpectedImpact: "Medium",
			TimeFrame:      "3-4 weeks",
		})
	}

	// Monitoring recommendation
	recommendations = append(recommendations, PerformanceRecommendation{
		Priority:       "low",
		Category:       "monitoring",
		Title:          "Implement Performance Monitoring",
		Description:    "Set up continuous performance monitoring to catch regressions early",
		Action:         "Integrate performance monitoring tools and establish performance budgets",
		ExpectedImpact: "Long-term",
		TimeFrame:      "Ongoing",
	})

	return recommendations
}
