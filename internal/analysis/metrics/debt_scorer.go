package metrics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// DebtScorer analyzes technical debt across JavaScript/TypeScript codebases
type DebtScorer struct {
	config DebtScoringConfig
}

// DebtScoringConfig defines thresholds and weights for technical debt calculation
type DebtScoringConfig struct {
	ComplexityWeight    float64 `yaml:"complexity_weight" json:"complexity_weight"`
	DuplicationWeight   float64 `yaml:"duplication_weight" json:"duplication_weight"`
	CodeSmellWeight     float64 `yaml:"code_smell_weight" json:"code_smell_weight"`
	ArchitectureWeight  float64 `yaml:"architecture_weight" json:"architecture_weight"`
	PerformanceWeight   float64 `yaml:"performance_weight" json:"performance_weight"`
	
	ChangeFrequencyWeight float64 `yaml:"change_frequency_weight" json:"change_frequency_weight"`
	ImpactWeight          float64 `yaml:"impact_weight" json:"impact_weight"`
	RemediationThreshold  float64 `yaml:"remediation_threshold" json:"remediation_threshold"`
	
	TrendAnalysisPeriod   int     `yaml:"trend_analysis_period" json:"trend_analysis_period"` // days
	PriorityCategories    int     `yaml:"priority_categories" json:"priority_categories"`
	MinConfidenceScore    float64 `yaml:"min_confidence_score" json:"min_confidence_score"`
}

// TechnicalDebtMetrics contains comprehensive technical debt analysis
type TechnicalDebtMetrics struct {
	OverallScore        float64             `json:"overall_score"`
	TotalDebtHours      float64             `json:"total_debt_hours"`
	DebtRatio           float64             `json:"debt_ratio"`
	TrendDirection      string              `json:"trend_direction"`
	
	Categories          map[string]DebtCategory `json:"categories"`
	FileDebtScores      map[string]FileDebt     `json:"file_debt_scores"`
	RemediationPlan     []RemediationItem       `json:"remediation_plan"`
	Recommendations     []DebtRecommendation    `json:"recommendations"`
	
	Dashboard           TechnicalDebtDashboard  `json:"dashboard"`
	Summary             DebtSummary             `json:"summary"`
}

// DebtCategory represents a category of technical debt
type DebtCategory struct {
	Name               string              `json:"name"`
	Score              float64             `json:"score"`
	DebtHours          float64             `json:"debt_hours"`
	Items              []TechnicalDebtItem `json:"items"`
	TrendDirection     string              `json:"trend_direction"`
	PriorityLevel      string              `json:"priority_level"`
	RemediationEffort  string              `json:"remediation_effort"`
}

// TechnicalDebtItem represents a specific debt issue
type TechnicalDebtItem struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Category         string                 `json:"category"`
	FilePath         string                 `json:"file_path"`
	FunctionName     string                 `json:"function_name,omitempty"`
	ClassName        string                 `json:"class_name,omitempty"`
	StartLine        int                    `json:"start_line"`
	EndLine          int                    `json:"end_line"`
	
	Description      string                 `json:"description"`
	Severity         string                 `json:"severity"`
	DebtScore        float64                `json:"debt_score"`
	EstimatedHours   float64                `json:"estimated_hours"`
	Priority         string                 `json:"priority"`
	
	ChangeFrequency  float64                `json:"change_frequency"`
	ImpactScore      float64                `json:"impact_score"`
	ConfidenceScore  float64                `json:"confidence_score"`
	
	RemediationSteps []string               `json:"remediation_steps"`
	RelatedIssues    []string               `json:"related_issues"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// FileDebt represents debt metrics for a specific file
type FileDebt struct {
	FilePath         string  `json:"file_path"`
	OverallScore     float64 `json:"overall_score"`
	DebtHours        float64 `json:"debt_hours"`
	ComplexityDebt   float64 `json:"complexity_debt"`
	DuplicationDebt  float64 `json:"duplication_debt"`
	CodeSmellDebt    float64 `json:"code_smell_debt"`
	ArchitectureDebt float64 `json:"architecture_debt"`
	PerformanceDebt  float64 `json:"performance_debt"`
	Priority         string  `json:"priority"`
	RemediationOrder int     `json:"remediation_order"`
}

// RemediationItem represents a prioritized remediation task
type RemediationItem struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Description           string   `json:"description"`
	Category              string   `json:"category"`
	Priority              string   `json:"priority"`
	EstimatedEffort       float64  `json:"estimated_effort"`
	ExpectedROI           float64  `json:"expected_roi"`
	ImpactScore           float64  `json:"impact_score"`
	DependentItems        []string `json:"dependent_items"`
	AffectedFiles         []string `json:"affected_files"`
	RemediationSteps      []string `json:"remediation_steps"`
	SuccessMetrics        []string `json:"success_metrics"`
}

// DebtRecommendation provides strategic debt management advice
type DebtRecommendation struct {
	Type             string   `json:"type"`
	Priority         string   `json:"priority"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Benefits         []string `json:"benefits"`
	Implementation   []string `json:"implementation"`
	EstimatedImpact  float64  `json:"estimated_impact"`
	TimeframeWeeks   int      `json:"timeframe_weeks"`
	ResourcesNeeded  []string `json:"resources_needed"`
}

// TechnicalDebtDashboard provides high-level debt insights
type TechnicalDebtDashboard struct {
	HealthScore           float64                    `json:"health_score"`
	TrendIndicator        string                     `json:"trend_indicator"`
	CriticalIssues        int                        `json:"critical_issues"`
	HighPriorityIssues    int                        `json:"high_priority_issues"`
	TotalDebtHours        float64                    `json:"total_debt_hours"`
	
	CategoryBreakdown     map[string]float64         `json:"category_breakdown"`
	FileRankings          []FileRanking              `json:"file_rankings"`
	QuickWins             []RemediationItem          `json:"quick_wins"`
	LongTermInitiatives   []RemediationItem          `json:"long_term_initiatives"`
	MonthlyTrend          []TrendDataPoint           `json:"monthly_trend"`
}

// FileRanking represents file ranking by debt score
type FileRanking struct {
	FilePath    string  `json:"file_path"`
	DebtScore   float64 `json:"debt_score"`
	DebtHours   float64 `json:"debt_hours"`
	IssueCount  int     `json:"issue_count"`
	Priority    string  `json:"priority"`
}

// TrendDataPoint represents debt trend over time
type TrendDataPoint struct {
	Date       string  `json:"date"`
	DebtScore  float64 `json:"debt_score"`
	DebtHours  float64 `json:"debt_hours"`
	IssueCount int     `json:"issue_count"`
}

// DebtSummary provides executive summary of technical debt
type DebtSummary struct {
	TotalFiles              int     `json:"total_files"`
	FilesWithDebt           int     `json:"files_with_debt"`
	AverageDebtPerFile      float64 `json:"average_debt_per_file"`
	WorstFileDebtScore      float64 `json:"worst_file_debt_score"`
	RecommendedFocus        string  `json:"recommended_focus"`
	EstimatedPaydownWeeks   int     `json:"estimated_paydown_weeks"`
	ROIScore                float64 `json:"roi_score"`
}

// NewDebtScorer creates a new technical debt scorer with default configuration
func NewDebtScorer() *DebtScorer {
	return &DebtScorer{
		config: DebtScoringConfig{
			ComplexityWeight:      0.25,
			DuplicationWeight:     0.20,
			CodeSmellWeight:       0.20,
			ArchitectureWeight:    0.20,
			PerformanceWeight:     0.15,
			
			ChangeFrequencyWeight: 0.30,
			ImpactWeight:          0.40,
			RemediationThreshold:  0.70,
			
			TrendAnalysisPeriod:   30,
			PriorityCategories:    4,
			MinConfidenceScore:    0.60,
		},
	}
}

// NewDebtScorerWithConfig creates a debt scorer with custom configuration
func NewDebtScorerWithConfig(config DebtScoringConfig) *DebtScorer {
	return &DebtScorer{
		config: config,
	}
}

// AnalyzeDebt performs comprehensive technical debt analysis
func (ds *DebtScorer) AnalyzeDebt(ctx context.Context, parseResults []*ast.ParseResult, complexityMetrics *ComplexityMetrics, duplicationMetrics *DuplicationMetrics) (*TechnicalDebtMetrics, error) {
	if len(parseResults) == 0 {
		return nil, fmt.Errorf("no parse results provided for debt analysis")
	}
	
	if complexityMetrics == nil || duplicationMetrics == nil {
		return nil, fmt.Errorf("complexity and duplication metrics are required for debt analysis")
	}
	
	metrics := &TechnicalDebtMetrics{
		Categories:     make(map[string]DebtCategory),
		FileDebtScores: make(map[string]FileDebt),
		RemediationPlan: []RemediationItem{},
		Recommendations: []DebtRecommendation{},
	}
	
	// Analyze different debt categories
	codeSmellItems, err := ds.analyzeCodeSmells(parseResults)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze code smells: %w", err)
	}
	
	architectureItems, err := ds.analyzeArchitectureViolations(parseResults)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze architecture violations: %w", err)
	}
	
	performanceItems, err := ds.analyzePerformanceIssues(parseResults)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze performance issues: %w", err)
	}
	
	// Combine all debt items
	allDebtItems := []TechnicalDebtItem{}
	allDebtItems = append(allDebtItems, codeSmellItems...)
	allDebtItems = append(allDebtItems, architectureItems...)
	allDebtItems = append(allDebtItems, performanceItems...)
	
	// Add complexity and duplication items
	complexityItems := ds.convertComplexityToDebt(complexityMetrics)
	duplicationItems := ds.convertDuplicationToDebt(duplicationMetrics)
	allDebtItems = append(allDebtItems, complexityItems...)
	allDebtItems = append(allDebtItems, duplicationItems...)
	
	// Calculate debt scores and prioritization
	ds.calculateDebtScores(allDebtItems)
	ds.calculatePriorities(allDebtItems)
	
	// Organize by categories
	metrics.Categories = ds.organizeByCategories(allDebtItems)
	
	// Calculate file-level debt scores
	metrics.FileDebtScores = ds.calculateFileDebtScores(parseResults, allDebtItems)
	
	// Generate remediation plan
	metrics.RemediationPlan = ds.generateRemediationPlan(allDebtItems)
	
	// Generate recommendations
	metrics.Recommendations = ds.generateRecommendations(allDebtItems, metrics.Categories)
	
	// Calculate overall metrics
	metrics.OverallScore = ds.calculateOverallScore(allDebtItems)
	metrics.TotalDebtHours = ds.calculateTotalDebtHours(allDebtItems)
	metrics.DebtRatio = ds.calculateDebtRatio(parseResults, allDebtItems)
	metrics.TrendDirection = ds.calculateTrendDirection(allDebtItems)
	
	// Generate dashboard
	metrics.Dashboard = ds.generateDashboard(allDebtItems, metrics.FileDebtScores)
	
	// Generate summary
	metrics.Summary = ds.generateSummary(parseResults, allDebtItems, metrics.FileDebtScores)
	
	return metrics, nil
}

// analyzeCodeSmells identifies code smell patterns
func (ds *DebtScorer) analyzeCodeSmells(parseResults []*ast.ParseResult) ([]TechnicalDebtItem, error) {
	items := []TechnicalDebtItem{}
	itemID := 0
	
	for _, parseResult := range parseResults {
		// Analyze functions for code smells
		for _, function := range parseResult.Functions {
			// Long Method smell
			if ds.isLongMethod(function) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("code_smell_%d", itemID),
					Type:            "long_method",
					Category:        "Code Smells",
					FilePath:        parseResult.FilePath,
					FunctionName:    function.Name,
					StartLine:       function.StartLine,
					EndLine:         function.EndLine,
					Description:     fmt.Sprintf("Method '%s' is too long (%d lines) and should be broken down", function.Name, function.EndLine-function.StartLine+1),
					Severity:        ds.determineLongMethodSeverity(function),
					EstimatedHours:  ds.estimateLongMethodRemediationEffort(function),
					RemediationSteps: []string{
						"Identify distinct responsibilities within the method",
						"Extract logical blocks into separate methods",
						"Ensure extracted methods have clear single purposes",
						"Update tests to cover new method structure",
						"Verify functionality remains unchanged",
					},
					Metadata: map[string]interface{}{
						"line_count":    function.EndLine - function.StartLine + 1,
						"is_async":      function.IsAsync,
						"param_count":   len(function.Parameters),
						"is_exported":   function.IsExported,
					},
				}
				items = append(items, item)
				itemID++
			}
			
			// Too Many Parameters smell
			if ds.hasTooManyParameters(function) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("code_smell_%d", itemID),
					Type:            "too_many_parameters",
					Category:        "Code Smells",
					FilePath:        parseResult.FilePath,
					FunctionName:    function.Name,
					StartLine:       function.StartLine,
					EndLine:         function.EndLine,
					Description:     fmt.Sprintf("Method '%s' has too many parameters (%d) making it hard to use", function.Name, len(function.Parameters)),
					Severity:        ds.determineTooManyParametersSeverity(function),
					EstimatedHours:  ds.estimateParameterRefactoringEffort(function),
					RemediationSteps: []string{
						"Group related parameters into configuration objects",
						"Use parameter objects or options pattern",
						"Consider builder pattern for complex configurations",
						"Update all callers to use new signature",
						"Add parameter validation in new structure",
					},
					Metadata: map[string]interface{}{
						"parameter_count": len(function.Parameters),
						"function_length": function.EndLine - function.StartLine + 1,
					},
				}
				items = append(items, item)
				itemID++
			}
		}
		
		// Analyze classes for code smells
		for _, class := range parseResult.Classes {
			// Large Class smell
			if ds.isLargeClass(class) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("code_smell_%d", itemID),
					Type:            "large_class",
					Category:        "Code Smells",
					FilePath:        parseResult.FilePath,
					ClassName:       class.Name,
					StartLine:       class.StartLine,
					EndLine:         class.EndLine,
					Description:     fmt.Sprintf("Class '%s' is too large with %d methods and should be split", class.Name, len(class.Methods)),
					Severity:        ds.determineLargeClassSeverity(class),
					EstimatedHours:  ds.estimateLargeClassRemediationEffort(class),
					RemediationSteps: []string{
						"Identify cohesive groups of methods and properties",
						"Extract related functionality into separate classes",
						"Use composition over inheritance where appropriate",
						"Ensure clear separation of concerns",
						"Update imports and references throughout codebase",
					},
					Metadata: map[string]interface{}{
						"method_count":    len(class.Methods),
						"property_count":  len(class.Properties),
						"line_count":      class.EndLine - class.StartLine + 1,
					},
				}
				items = append(items, item)
				itemID++
			}
			
			// Too Many Methods smell
			if ds.hasTooManyMethods(class) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("code_smell_%d", itemID),
					Type:            "too_many_methods",
					Category:        "Code Smells",
					FilePath:        parseResult.FilePath,
					ClassName:       class.Name,
					StartLine:       class.StartLine,
					EndLine:         class.EndLine,
					Description:     fmt.Sprintf("Class '%s' has too many methods (%d) indicating multiple responsibilities", class.Name, len(class.Methods)),
					Severity:        "medium",
					EstimatedHours:  float64(len(class.Methods)) * 0.5, // 30 minutes per method
					RemediationSteps: []string{
						"Apply Single Responsibility Principle",
						"Extract methods into logical service classes",
						"Use delegation pattern where appropriate",
						"Consider facade pattern for complex interactions",
					},
					Metadata: map[string]interface{}{
						"method_count": len(class.Methods),
					},
				}
				items = append(items, item)
				itemID++
			}
		}
	}
	
	return items, nil
}

// analyzeArchitectureViolations identifies architectural debt patterns
func (ds *DebtScorer) analyzeArchitectureViolations(parseResults []*ast.ParseResult) ([]TechnicalDebtItem, error) {
	items := []TechnicalDebtItem{}
	itemID := 1000 // Start with higher ID to avoid conflicts
	
	for _, parseResult := range parseResults {
		// Analyze circular dependencies
		if ds.hasCircularDependencies(parseResult) {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("arch_violation_%d", itemID),
				Type:            "circular_dependency",
				Category:        "Architecture Violations",
				FilePath:        parseResult.FilePath,
				StartLine:       1,
				EndLine:         ds.estimateFileLineCount(parseResult),
				Description:     fmt.Sprintf("File '%s' has circular dependency patterns that violate clean architecture", parseResult.FilePath),
				Severity:        "high",
				EstimatedHours:  4.0,
				RemediationSteps: []string{
					"Map current dependency graph",
					"Identify circular dependency chains",
					"Extract shared dependencies into separate modules",
					"Apply dependency inversion principle",
					"Refactor import statements",
					"Validate architecture with dependency analysis tools",
				},
				Metadata: map[string]interface{}{
					"import_count": len(parseResult.Imports),
					"export_count": len(parseResult.Exports),
				},
			}
			items = append(items, item)
			itemID++
		}
		
		// Analyze God Object pattern
		if ds.hasGodObjectPattern(parseResult) {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("arch_violation_%d", itemID),
				Type:            "god_object",
				Category:        "Architecture Violations",
				FilePath:        parseResult.FilePath,
				StartLine:       1,
				EndLine:         ds.estimateFileLineCount(parseResult),
				Description:     fmt.Sprintf("File '%s' contains god object with excessive responsibilities", parseResult.FilePath),
				Severity:        "high",
				EstimatedHours:  6.0,
				RemediationSteps: []string{
					"Identify distinct responsibilities",
					"Extract functionality into focused services",
					"Apply single responsibility principle",
					"Create proper abstraction layers",
					"Implement dependency injection",
				},
				Metadata: map[string]interface{}{
					"function_count": len(parseResult.Functions),
					"class_count":    len(parseResult.Classes),
				},
			}
			items = append(items, item)
			itemID++
		}
		
		// Analyze tight coupling
		if ds.hasTightCoupling(parseResult) {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("arch_violation_%d", itemID),
				Type:            "tight_coupling",
				Category:        "Architecture Violations",
				FilePath:        parseResult.FilePath,
				StartLine:       1,
				EndLine:         ds.estimateFileLineCount(parseResult),
				Description:     fmt.Sprintf("File '%s' exhibits tight coupling with other modules", parseResult.FilePath),
				Severity:        "medium",
				EstimatedHours:  3.0,
				RemediationSteps: []string{
					"Identify coupling points",
					"Introduce abstractions and interfaces",
					"Apply dependency inversion",
					"Use event-driven patterns where appropriate",
					"Implement proper module boundaries",
				},
				Metadata: map[string]interface{}{
					"coupling_score": ds.calculateCouplingScore(parseResult),
				},
			}
			items = append(items, item)
			itemID++
		}
		
		// Analyze layering violations
		if ds.hasLayeringViolations(parseResult) {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("arch_violation_%d", itemID),
				Type:            "layering_violation",
				Category:        "Architecture Violations",
				FilePath:        parseResult.FilePath,
				StartLine:       1,
				EndLine:         ds.estimateFileLineCount(parseResult),
				Description:     fmt.Sprintf("File '%s' violates architectural layering principles", parseResult.FilePath),
				Severity:        "medium",
				EstimatedHours:  2.5,
				RemediationSteps: []string{
					"Define clear architectural layers",
					"Ensure unidirectional dependencies",
					"Move misplaced functionality to appropriate layers",
					"Create proper interfaces between layers",
				},
				Metadata: map[string]interface{}{
					"layer_violations": ds.countLayeringViolations(parseResult),
				},
			}
			items = append(items, item)
			itemID++
		}
	}
	
	return items, nil
}

// analyzePerformanceIssues identifies performance-related debt
func (ds *DebtScorer) analyzePerformanceIssues(parseResults []*ast.ParseResult) ([]TechnicalDebtItem, error) {
	items := []TechnicalDebtItem{}
	itemID := 2000 // Start with higher ID to avoid conflicts
	
	for _, parseResult := range parseResults {
		// Analyze functions for performance issues
		for _, function := range parseResult.Functions {
			// Nested loops detection
			if ds.hasNestedLoops(function) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("performance_%d", itemID),
					Type:            "nested_loops",
					Category:        "Performance Issues",
					FilePath:        parseResult.FilePath,
					FunctionName:    function.Name,
					StartLine:       function.StartLine,
					EndLine:         function.EndLine,
					Description:     fmt.Sprintf("Function '%s' contains nested loops that may cause performance issues", function.Name),
					Severity:        ds.determineNestedLoopSeverity(function),
					EstimatedHours:  2.0,
					RemediationSteps: []string{
						"Analyze loop complexity and data access patterns",
						"Consider using more efficient algorithms",
						"Implement caching for repeated calculations",
						"Use appropriate data structures (Maps, Sets)",
						"Consider async processing for large datasets",
					},
					Metadata: map[string]interface{}{
						"estimated_complexity": ds.estimateAlgorithmicComplexity(function),
					},
				}
				items = append(items, item)
				itemID++
			}
			
			// Synchronous operations in async context
			if ds.hasSyncInAsyncAntiPattern(function) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("performance_%d", itemID),
					Type:            "sync_in_async",
					Category:        "Performance Issues",
					FilePath:        parseResult.FilePath,
					FunctionName:    function.Name,
					StartLine:       function.StartLine,
					EndLine:         function.EndLine,
					Description:     fmt.Sprintf("Async function '%s' contains blocking synchronous operations", function.Name),
					Severity:        "high",
					EstimatedHours:  1.5,
					RemediationSteps: []string{
						"Identify blocking synchronous operations",
						"Replace with async equivalents",
						"Implement proper error handling for async operations",
						"Use Promise.all() for parallel operations",
						"Add proper await statements",
					},
					Metadata: map[string]interface{}{
						"is_async": function.IsAsync,
					},
				}
				items = append(items, item)
				itemID++
			}
			
			// Memory leaks potential
			if ds.hasMemoryLeakPotential(function) {
				item := TechnicalDebtItem{
					ID:              fmt.Sprintf("performance_%d", itemID),
					Type:            "memory_leak_risk",
					Category:        "Performance Issues",
					FilePath:        parseResult.FilePath,
					FunctionName:    function.Name,
					StartLine:       function.StartLine,
					EndLine:         function.EndLine,
					Description:     fmt.Sprintf("Function '%s' has patterns that may lead to memory leaks", function.Name),
					Severity:        "medium",
					EstimatedHours:  1.0,
					RemediationSteps: []string{
						"Review event listener registration and cleanup",
						"Ensure proper cleanup of timers and intervals",
						"Check for closure-related memory retention",
						"Implement proper resource disposal patterns",
						"Add memory usage monitoring",
					},
					Metadata: map[string]interface{}{
						"risk_patterns": ds.identifyMemoryLeakPatterns(function),
					},
				}
				items = append(items, item)
				itemID++
			}
		}
		
		// File-level performance issues
		if ds.hasExcessiveImports(parseResult) {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("performance_%d", itemID),
				Type:            "excessive_imports",
				Category:        "Performance Issues",
				FilePath:        parseResult.FilePath,
				StartLine:       1,
				EndLine:         10, // Imports are typically at the top
				Description:     fmt.Sprintf("File '%s' has excessive imports affecting bundle size", parseResult.FilePath),
				Severity:        "low",
				EstimatedHours:  0.5,
				RemediationSteps: []string{
					"Audit import statements for unused imports",
					"Use tree shaking compatible import patterns",
					"Implement lazy loading for heavy dependencies",
					"Consider code splitting for large modules",
				},
				Metadata: map[string]interface{}{
					"import_count": len(parseResult.Imports),
				},
			}
			items = append(items, item)
			itemID++
		}
	}
	
	return items, nil
}

// Helper functions for debt analysis
func (ds *DebtScorer) isLongMethod(function ast.FunctionInfo) bool {
	lineCount := function.EndLine - function.StartLine + 1
	return lineCount > 30 // Threshold for long method
}

func (ds *DebtScorer) determineLongMethodSeverity(function ast.FunctionInfo) string {
	lineCount := function.EndLine - function.StartLine + 1
	if lineCount > 100 {
		return "high"
	} else if lineCount > 50 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) estimateLongMethodRemediationEffort(function ast.FunctionInfo) float64 {
	lineCount := function.EndLine - function.StartLine + 1
	baseHours := 2.0
	return baseHours + (float64(lineCount)/50.0)*1.0 // Add 1 hour per 50 lines
}

func (ds *DebtScorer) hasTooManyParameters(function ast.FunctionInfo) bool {
	return len(function.Parameters) > 5 // Threshold for too many parameters
}

func (ds *DebtScorer) determineTooManyParametersSeverity(function ast.FunctionInfo) string {
	paramCount := len(function.Parameters)
	if paramCount > 10 {
		return "high"
	} else if paramCount > 7 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) estimateParameterRefactoringEffort(function ast.FunctionInfo) float64 {
	paramCount := len(function.Parameters)
	return float64(paramCount) * 0.25 // 15 minutes per parameter
}

func (ds *DebtScorer) isLargeClass(class ast.ClassInfo) bool {
	return len(class.Methods) > 20 || (class.EndLine-class.StartLine+1) > 500
}

func (ds *DebtScorer) determineLargeClassSeverity(class ast.ClassInfo) string {
	methodCount := len(class.Methods)
	lineCount := class.EndLine - class.StartLine + 1
	
	if methodCount > 50 || lineCount > 1000 {
		return "high"
	} else if methodCount > 30 || lineCount > 750 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) estimateLargeClassRemediationEffort(class ast.ClassInfo) float64 {
	methodCount := len(class.Methods)
	baseHours := 4.0
	return baseHours + (float64(methodCount)/10.0)*2.0 // Add 2 hours per 10 methods
}

func (ds *DebtScorer) hasTooManyMethods(class ast.ClassInfo) bool {
	return len(class.Methods) > 15
}

func (ds *DebtScorer) hasCircularDependencies(parseResult *ast.ParseResult) bool {
	// Simplified heuristic: Check if file imports and exports suggest circular patterns
	return len(parseResult.Imports) > 10 && len(parseResult.Exports) > 5
}

func (ds *DebtScorer) hasGodObjectPattern(parseResult *ast.ParseResult) bool {
	// Check for excessive functionality in a single file
	totalComplexity := len(parseResult.Functions) + len(parseResult.Classes)*3
	return totalComplexity > 50
}

func (ds *DebtScorer) hasTightCoupling(parseResult *ast.ParseResult) bool {
	// Heuristic based on import/export ratios
	return len(parseResult.Imports) > 15
}

func (ds *DebtScorer) calculateCouplingScore(parseResult *ast.ParseResult) float64 {
	importCount := float64(len(parseResult.Imports))
	exportCount := float64(len(parseResult.Exports))
	functionCount := float64(len(parseResult.Functions))
	
	if functionCount == 0 {
		return 0.0
	}
	
	return (importCount + exportCount) / functionCount
}

func (ds *DebtScorer) hasLayeringViolations(parseResult *ast.ParseResult) bool {
	// Simplified check based on file path and imports
	fileName := strings.ToLower(parseResult.FilePath)
	
	// Check if a component/UI file imports business logic
	if strings.Contains(fileName, "component") || strings.Contains(fileName, "ui") {
		for _, imp := range parseResult.Imports {
			if strings.Contains(strings.ToLower(imp.Source), "service") || 
			   strings.Contains(strings.ToLower(imp.Source), "repository") {
				return true
			}
		}
	}
	
	return false
}

func (ds *DebtScorer) countLayeringViolations(parseResult *ast.ParseResult) int {
	violations := 0
	fileName := strings.ToLower(parseResult.FilePath)
	
	for _, imp := range parseResult.Imports {
		impPath := strings.ToLower(imp.Source)
		
		// UI importing business logic
		if (strings.Contains(fileName, "component") || strings.Contains(fileName, "ui")) &&
		   (strings.Contains(impPath, "service") || strings.Contains(impPath, "repository")) {
			violations++
		}
		
		// Business logic importing UI
		if (strings.Contains(fileName, "service") || strings.Contains(fileName, "repository")) &&
		   (strings.Contains(impPath, "component") || strings.Contains(impPath, "ui")) {
			violations++
		}
	}
	
	return violations
}

func (ds *DebtScorer) hasNestedLoops(function ast.FunctionInfo) bool {
	// Simplified heuristic based on function complexity
	lineCount := function.EndLine - function.StartLine + 1
	return lineCount > 20 && len(function.Parameters) > 2 // Rough approximation
}

func (ds *DebtScorer) determineNestedLoopSeverity(function ast.FunctionInfo) string {
	lineCount := function.EndLine - function.StartLine + 1
	if lineCount > 50 {
		return "high"
	} else if lineCount > 25 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) estimateAlgorithmicComplexity(function ast.FunctionInfo) string {
	lineCount := function.EndLine - function.StartLine + 1
	paramCount := len(function.Parameters)
	
	if lineCount > 50 && paramCount > 3 {
		return "O(n²) or worse"
	} else if lineCount > 25 {
		return "O(n log n)"
	}
	return "O(n)"
}

func (ds *DebtScorer) hasSyncInAsyncAntiPattern(function ast.FunctionInfo) bool {
	// Check if async function has potential blocking patterns
	return function.IsAsync && (function.EndLine-function.StartLine+1) > 15
}

func (ds *DebtScorer) hasMemoryLeakPotential(function ast.FunctionInfo) bool {
	// Heuristic: Long functions with parameters might have closure issues
	lineCount := function.EndLine - function.StartLine + 1
	return lineCount > 30 && len(function.Parameters) > 3
}

func (ds *DebtScorer) identifyMemoryLeakPatterns(function ast.FunctionInfo) []string {
	patterns := []string{}
	
	if function.IsAsync {
		patterns = append(patterns, "async_callback_retention")
	}
	
	if len(function.Parameters) > 5 {
		patterns = append(patterns, "parameter_closure_retention")
	}
	
	lineCount := function.EndLine - function.StartLine + 1
	if lineCount > 50 {
		patterns = append(patterns, "large_function_scope")
	}
	
	return patterns
}

func (ds *DebtScorer) hasExcessiveImports(parseResult *ast.ParseResult) bool {
	return len(parseResult.Imports) > 20
}

func (ds *DebtScorer) estimateFileLineCount(parseResult *ast.ParseResult) int {
	maxLine := 0
	
	for _, function := range parseResult.Functions {
		if function.EndLine > maxLine {
			maxLine = function.EndLine
		}
	}
	
	for _, class := range parseResult.Classes {
		if class.EndLine > maxLine {
			maxLine = class.EndLine
		}
	}
	
	if maxLine == 0 {
		maxLine = 100 // Default estimate
	}
	
	return maxLine
}

// convertComplexityToDebt converts complexity metrics to debt items
func (ds *DebtScorer) convertComplexityToDebt(complexityMetrics *ComplexityMetrics) []TechnicalDebtItem {
	items := []TechnicalDebtItem{}
	itemID := 3000
	
	for _, functionMetric := range complexityMetrics.FunctionMetrics {
		if functionMetric.SeverityLevel == "high" || functionMetric.SeverityLevel == "severe" {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("complexity_debt_%d", itemID),
				Type:            "high_complexity",
				Category:        "Complexity Debt",
				FilePath:        functionMetric.FilePath,
				FunctionName:    functionMetric.Name,
				StartLine:       functionMetric.StartLine,
				EndLine:         functionMetric.EndLine,
				Description:     fmt.Sprintf("Function '%s' has high complexity score of %.1f", functionMetric.Name, functionMetric.WeightedScore),
				Severity:        functionMetric.SeverityLevel,
				EstimatedHours:  ds.estimateComplexityRemediationEffort(functionMetric.WeightedScore),
				RemediationSteps: functionMetric.Recommendations,
				Metadata: map[string]interface{}{
					"complexity_score": functionMetric.WeightedScore,
					"cyclomatic":       functionMetric.CyclomaticValue,
					"cognitive":        functionMetric.CognitiveValue,
				},
			}
			items = append(items, item)
			itemID++
		}
	}
	
	return items
}

// convertDuplicationToDebt converts duplication metrics to debt items  
func (ds *DebtScorer) convertDuplicationToDebt(duplicationMetrics *DuplicationMetrics) []TechnicalDebtItem {
	items := []TechnicalDebtItem{}
	itemID := 4000
	
	// Convert exact duplicates
	for _, cluster := range duplicationMetrics.ExactDuplicates {
		if cluster.Priority == "high" || cluster.Priority == "critical" {
			item := TechnicalDebtItem{
				ID:              fmt.Sprintf("duplication_debt_%d", itemID),
				Type:            "exact_duplication",
				Category:        "Duplication Debt",
				FilePath:        cluster.Instances[0].FilePath,
				FunctionName:    cluster.Instances[0].FunctionName,
				StartLine:       cluster.Instances[0].StartLine,
				EndLine:         cluster.Instances[0].EndLine,
				Description:     fmt.Sprintf("Exact code duplication with %d instances across files", len(cluster.Instances)),
				Severity:        ds.mapPriorityToSeverity(cluster.Priority),
				EstimatedHours:  cluster.MaintenanceBurden * 0.1, // Convert burden to hours
				RemediationSteps: cluster.Recommendations,
				Metadata: map[string]interface{}{
					"instance_count":     len(cluster.Instances),
					"similarity_score":   cluster.SimilarityScore,
					"maintenance_burden": cluster.MaintenanceBurden,
				},
			}
			items = append(items, item)
			itemID++
		}
	}
	
	return items
}

// calculateDebtScores calculates debt scores for all items
func (ds *DebtScorer) calculateDebtScores(items []TechnicalDebtItem) {
	for i := range items {
		item := &items[i]
		
		// Base score from severity
		baseScore := ds.severityToScore(item.Severity)
		
		// Adjust for estimated hours
		effortMultiplier := math.Min(item.EstimatedHours/10.0, 2.0) // Cap at 2x multiplier
		
		// Calculate final debt score
		item.DebtScore = baseScore * (1.0 + effortMultiplier)
		
		// Set confidence score based on available metadata
		item.ConfidenceScore = ds.calculateConfidenceScore(*item)
	}
}

// calculatePriorities determines priority for debt items
func (ds *DebtScorer) calculatePriorities(items []TechnicalDebtItem) {
	for i := range items {
		item := &items[i]
		
		// Calculate impact score based on debt score and confidence
		item.ImpactScore = item.DebtScore * item.ConfidenceScore
		
		// Simulate change frequency (in real implementation, this would come from VCS)
		item.ChangeFrequency = ds.estimateChangeFrequency(*item)
		
		// Calculate priority based on impact and change frequency
		priorityScore := (item.ImpactScore * ds.config.ImpactWeight) + 
						(item.ChangeFrequency * ds.config.ChangeFrequencyWeight)
		
		item.Priority = ds.scoreToPriority(priorityScore)
	}
}

// organizeByCategories groups debt items by category
func (ds *DebtScorer) organizeByCategories(items []TechnicalDebtItem) map[string]DebtCategory {
	categories := make(map[string]DebtCategory)
	
	// Group items by category
	categoryItems := make(map[string][]TechnicalDebtItem)
	for _, item := range items {
		categoryItems[item.Category] = append(categoryItems[item.Category], item)
	}
	
	// Create category summaries
	for categoryName, catItems := range categoryItems {
		category := DebtCategory{
			Name:  categoryName,
			Items: catItems,
		}
		
		// Calculate category metrics
		totalScore := 0.0
		totalHours := 0.0
		for _, item := range catItems {
			totalScore += item.DebtScore
			totalHours += item.EstimatedHours
		}
		
		category.Score = totalScore
		category.DebtHours = totalHours
		category.TrendDirection = ds.calculateCategoryTrend(catItems)
		category.PriorityLevel = ds.calculateCategoryPriority(catItems)
		category.RemediationEffort = ds.calculateCategoryEffort(totalHours)
		
		categories[categoryName] = category
	}
	
	return categories
}

// calculateFileDebtScores calculates debt scores per file
func (ds *DebtScorer) calculateFileDebtScores(parseResults []*ast.ParseResult, items []TechnicalDebtItem) map[string]FileDebt {
	fileDebts := make(map[string]FileDebt)
	
	// Initialize file debt records
	for _, parseResult := range parseResults {
		fileDebts[parseResult.FilePath] = FileDebt{
			FilePath: parseResult.FilePath,
		}
	}
	
	// Aggregate debt by file
	for _, item := range items {
		debt, exists := fileDebts[item.FilePath]
		if !exists {
			debt = FileDebt{FilePath: item.FilePath}
		}
		
		debt.DebtHours += item.EstimatedHours
		debt.OverallScore += item.DebtScore
		
		// Categorize debt by type
		switch item.Category {
		case "Complexity Debt":
			debt.ComplexityDebt += item.DebtScore
		case "Code Smells":
			debt.CodeSmellDebt += item.DebtScore
		case "Architecture Violations":
			debt.ArchitectureDebt += item.DebtScore
		case "Performance Issues":
			debt.PerformanceDebt += item.DebtScore
		case "Duplication Debt":
			debt.DuplicationDebt += item.DebtScore
		}
		
		fileDebts[item.FilePath] = debt
	}
	
	// Calculate priorities and remediation order
	fileList := make([]FileDebt, 0, len(fileDebts))
	for _, debt := range fileDebts {
		debt.Priority = ds.scoreToFilePriority(debt.OverallScore)
		fileList = append(fileList, debt)
	}
	
	// Sort by overall score for remediation order
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].OverallScore > fileList[j].OverallScore
	})
	
	// Set remediation order
	for i, debt := range fileList {
		debt.RemediationOrder = i + 1
		fileDebts[debt.FilePath] = debt
	}
	
	return fileDebts
}

// generateRemediationPlan creates prioritized remediation plan
func (ds *DebtScorer) generateRemediationPlan(items []TechnicalDebtItem) []RemediationItem {
	plan := []RemediationItem{}
	
	// Sort items by priority and impact
	sortedItems := make([]TechnicalDebtItem, len(items))
	copy(sortedItems, items)
	
	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].ImpactScore > sortedItems[j].ImpactScore
	})
	
	// Group related items for batch remediation
	itemGroups := ds.groupRelatedItems(sortedItems)
	
	for i, group := range itemGroups {
		if len(group) == 0 {
			continue
		}
		
		// Calculate combined metrics for the group
		totalEffort := 0.0
		totalROI := 0.0
		affectedFiles := make(map[string]bool)
		
		for _, item := range group {
			totalEffort += item.EstimatedHours
			totalROI += ds.calculateItemROI(item)
			affectedFiles[item.FilePath] = true
		}
		
		files := make([]string, 0, len(affectedFiles))
		for file := range affectedFiles {
			files = append(files, file)
		}
		
		remediationItem := RemediationItem{
			ID:              fmt.Sprintf("remediation_%d", i),
			Title:           ds.generateRemediationTitle(group),
			Description:     ds.generateRemediationDescription(group),
			Category:        group[0].Category,
			Priority:        ds.calculateGroupPriority(group),
			EstimatedEffort: totalEffort,
			ExpectedROI:     totalROI / float64(len(group)),
			ImpactScore:     ds.calculateGroupImpactScore(group),
			AffectedFiles:   files,
			RemediationSteps: ds.mergeRemediationSteps(group),
			SuccessMetrics:  ds.generateSuccessMetrics(group),
		}
		
		plan = append(plan, remediationItem)
	}
	
	return plan
}

// generateRecommendations creates strategic recommendations
func (ds *DebtScorer) generateRecommendations(items []TechnicalDebtItem, categories map[string]DebtCategory) []DebtRecommendation {
	recommendations := []DebtRecommendation{}
	
	// Analyze category patterns for strategic recommendations
	for categoryName, category := range categories {
		if category.Score > 50.0 { // Significant debt in this category
			recommendation := ds.generateCategoryRecommendation(categoryName, category)
			recommendations = append(recommendations, recommendation)
		}
	}
	
	// Add general recommendations based on overall patterns
	generalRecs := ds.generateGeneralRecommendations(items, categories)
	recommendations = append(recommendations, generalRecs...)
	
	return recommendations
}

// Helper functions for debt scoring
func (ds *DebtScorer) estimateComplexityRemediationEffort(complexityScore float64) float64 {
	baseHours := 1.0
	return baseHours + (complexityScore/10.0)*2.0 // 2 hours per 10 points of complexity
}

func (ds *DebtScorer) mapPriorityToSeverity(priority string) string {
	switch priority {
	case "critical":
		return "high"
	case "high":
		return "high" 
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func (ds *DebtScorer) severityToScore(severity string) float64 {
	switch severity {
	case "high":
		return 10.0
	case "medium":
		return 6.0
	case "low":
		return 3.0
	default:
		return 1.0
	}
}

func (ds *DebtScorer) calculateConfidenceScore(item TechnicalDebtItem) float64 {
	confidence := 0.7 // Base confidence
	
	// Increase confidence based on available metadata
	if len(item.Metadata) > 2 {
		confidence += 0.1
	}
	
	// Adjust based on item type
	switch item.Type {
	case "high_complexity", "exact_duplication":
		confidence += 0.2 // More reliable detection
	case "memory_leak_risk", "circular_dependency":
		confidence -= 0.1 // More heuristic-based
	}
	
	return math.Min(confidence, 1.0)
}

func (ds *DebtScorer) estimateChangeFrequency(item TechnicalDebtItem) float64 {
	// Simulate change frequency based on file type and function characteristics
	frequency := 0.3 // Base frequency
	
	// Files with more functionality tend to change more
	if strings.Contains(strings.ToLower(item.FilePath), "service") ||
	   strings.Contains(strings.ToLower(item.FilePath), "controller") {
		frequency += 0.3
	}
	
	// Component files change frequently
	if strings.Contains(strings.ToLower(item.FilePath), "component") {
		frequency += 0.2
	}
	
	// Utility files change less
	if strings.Contains(strings.ToLower(item.FilePath), "util") ||
	   strings.Contains(strings.ToLower(item.FilePath), "helper") {
		frequency -= 0.1
	}
	
	return math.Max(0.1, math.Min(frequency, 1.0))
}

func (ds *DebtScorer) scoreToPriority(score float64) string {
	if score > 8.0 {
		return "critical"
	} else if score > 6.0 {
		return "high"
	} else if score > 3.0 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) calculateCategoryTrend(items []TechnicalDebtItem) string {
	// Simplified trend calculation based on item severity distribution
	highSeverityCount := 0
	for _, item := range items {
		if item.Severity == "high" {
			highSeverityCount++
		}
	}
	
	if float64(highSeverityCount)/float64(len(items)) > 0.5 {
		return "worsening"
	} else if float64(highSeverityCount)/float64(len(items)) < 0.2 {
		return "improving"
	}
	return "stable"
}

func (ds *DebtScorer) calculateCategoryPriority(items []TechnicalDebtItem) string {
	criticalCount := 0
	highCount := 0
	
	for _, item := range items {
		switch item.Priority {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}
	
	if criticalCount > 0 {
		return "critical"
	} else if highCount > len(items)/2 {
		return "high"
	} else if highCount > 0 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) calculateCategoryEffort(totalHours float64) string {
	if totalHours > 40 {
		return "high"
	} else if totalHours > 20 {
		return "medium"
	} else if totalHours > 5 {
		return "low"
	}
	return "minimal"
}

func (ds *DebtScorer) scoreToFilePriority(score float64) string {
	if score > 50.0 {
		return "critical"
	} else if score > 30.0 {
		return "high"
	} else if score > 15.0 {
		return "medium"
	}
	return "low"
}

// Remaining essential functions for debt analysis
func (ds *DebtScorer) calculateOverallScore(items []TechnicalDebtItem) float64 {
	if len(items) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, item := range items {
		totalScore += item.DebtScore
	}
	
	return totalScore / float64(len(items))
}

func (ds *DebtScorer) calculateTotalDebtHours(items []TechnicalDebtItem) float64 {
	totalHours := 0.0
	for _, item := range items {
		totalHours += item.EstimatedHours
	}
	return totalHours
}

func (ds *DebtScorer) calculateDebtRatio(parseResults []*ast.ParseResult, items []TechnicalDebtItem) float64 {
	if len(parseResults) == 0 {
		return 0.0
	}
	
	filesWithDebt := make(map[string]bool)
	for _, item := range items {
		filesWithDebt[item.FilePath] = true
	}
	
	return float64(len(filesWithDebt)) / float64(len(parseResults))
}

func (ds *DebtScorer) calculateTrendDirection(items []TechnicalDebtItem) string {
	if len(items) == 0 {
		return "stable"
	}
	
	// Simplified trend based on severity distribution
	highSeverityCount := 0
	for _, item := range items {
		if item.Severity == "high" {
			highSeverityCount++
		}
	}
	
	ratio := float64(highSeverityCount) / float64(len(items))
	if ratio > 0.4 {
		return "worsening"
	} else if ratio < 0.1 {
		return "improving" 
	}
	return "stable"
}

func (ds *DebtScorer) generateDashboard(items []TechnicalDebtItem, fileDebts map[string]FileDebt) TechnicalDebtDashboard {
	dashboard := TechnicalDebtDashboard{
		CategoryBreakdown:   make(map[string]float64),
		FileRankings:        []FileRanking{},
		QuickWins:          []RemediationItem{},
		LongTermInitiatives: []RemediationItem{},
		MonthlyTrend:       ds.generateMockTrendData(),
	}
	
	// Calculate health score
	dashboard.HealthScore = 100.0 - math.Min(ds.calculateOverallScore(items)*5, 100.0)
	
	// Count issues by priority
	for _, item := range items {
		switch item.Priority {
		case "critical":
			dashboard.CriticalIssues++
		case "high":
			dashboard.HighPriorityIssues++
		}
		
		dashboard.CategoryBreakdown[item.Category] += item.DebtScore
	}
	
	dashboard.TotalDebtHours = ds.calculateTotalDebtHours(items)
	dashboard.TrendIndicator = ds.calculateTrendDirection(items)
	
	// Generate file rankings
	fileList := make([]FileRanking, 0, len(fileDebts))
	for _, debt := range fileDebts {
		if debt.OverallScore > 0 {
			ranking := FileRanking{
				FilePath:   debt.FilePath,
				DebtScore:  debt.OverallScore,
				DebtHours:  debt.DebtHours,
				IssueCount: ds.countFileIssues(debt.FilePath, items),
				Priority:   debt.Priority,
			}
			fileList = append(fileList, ranking)
		}
	}
	
	// Sort by debt score
	sort.Slice(fileList, func(i, j int) bool {
		return fileList[i].DebtScore > fileList[j].DebtScore
	})
	
	// Take top 10
	if len(fileList) > 10 {
		dashboard.FileRankings = fileList[:10]
	} else {
		dashboard.FileRankings = fileList
	}
	
	// Generate quick wins and long-term initiatives
	dashboard.QuickWins, dashboard.LongTermInitiatives = ds.categorizeRemediationEfforts(items)
	
	return dashboard
}

func (ds *DebtScorer) generateSummary(parseResults []*ast.ParseResult, items []TechnicalDebtItem, fileDebts map[string]FileDebt) DebtSummary {
	summary := DebtSummary{
		TotalFiles:    len(parseResults),
		FilesWithDebt: len(fileDebts),
	}
	
	if summary.TotalFiles > 0 {
		totalDebtScore := 0.0
		maxDebtScore := 0.0
		
		for _, debt := range fileDebts {
			totalDebtScore += debt.OverallScore
			if debt.OverallScore > maxDebtScore {
				maxDebtScore = debt.OverallScore
			}
		}
		
		if summary.FilesWithDebt > 0 {
			summary.AverageDebtPerFile = totalDebtScore / float64(summary.FilesWithDebt)
		}
		summary.WorstFileDebtScore = maxDebtScore
	}
	
	// Determine recommended focus
	summary.RecommendedFocus = ds.determineRecommendedFocus(items)
	
	// Estimate paydown timeline
	totalEffort := ds.calculateTotalDebtHours(items)
	summary.EstimatedPaydownWeeks = int(math.Ceil(totalEffort / 40.0)) // 40 hours per week
	
	// Calculate ROI
	summary.ROIScore = ds.calculateOverallROI(items)
	
	return summary
}

// Helper functions for dashboard and summary generation
func (ds *DebtScorer) countFileIssues(filePath string, items []TechnicalDebtItem) int {
	count := 0
	for _, item := range items {
		if item.FilePath == filePath {
			count++
		}
	}
	return count
}

func (ds *DebtScorer) generateMockTrendData() []TrendDataPoint {
	// Generate mock trend data for the last 6 months
	trends := []TrendDataPoint{}
	currentTime := time.Now()
	
	for i := 5; i >= 0; i-- {
		date := currentTime.AddDate(0, -i, 0)
		trend := TrendDataPoint{
			Date:       date.Format("2006-01"),
			DebtScore:  45.0 + float64(i)*2.5, // Simulate worsening trend
			DebtHours:  120.0 + float64(i)*15.0,
			IssueCount: 25 + i*3,
		}
		trends = append(trends, trend)
	}
	
	return trends
}

func (ds *DebtScorer) categorizeRemediationEfforts(items []TechnicalDebtItem) ([]RemediationItem, []RemediationItem) {
	quickWins := []RemediationItem{}
	longTerm := []RemediationItem{}
	
	// Group items by effort level
	for _, item := range items {
		if item.EstimatedHours <= 2.0 && item.ImpactScore > 5.0 { // High impact, low effort
			quickWin := RemediationItem{
				ID:              fmt.Sprintf("quickwin_%s", item.ID),
				Title:           fmt.Sprintf("Fix %s", item.Type),
				Description:     item.Description,
				Category:        item.Category,
				Priority:        item.Priority,
				EstimatedEffort: item.EstimatedHours,
				ExpectedROI:     ds.calculateItemROI(item),
				ImpactScore:     item.ImpactScore,
				AffectedFiles:   []string{item.FilePath},
				RemediationSteps: item.RemediationSteps,
			}
			quickWins = append(quickWins, quickWin)
		} else if item.EstimatedHours > 8.0 { // High effort items
			longTermItem := RemediationItem{
				ID:              fmt.Sprintf("longterm_%s", item.ID),
				Title:           fmt.Sprintf("Resolve %s", item.Type),
				Description:     item.Description,
				Category:        item.Category,
				Priority:        item.Priority,
				EstimatedEffort: item.EstimatedHours,
				ExpectedROI:     ds.calculateItemROI(item),
				ImpactScore:     item.ImpactScore,
				AffectedFiles:   []string{item.FilePath},
				RemediationSteps: item.RemediationSteps,
			}
			longTerm = append(longTerm, longTermItem)
		}
	}
	
	return quickWins, longTerm
}

func (ds *DebtScorer) determineRecommendedFocus(items []TechnicalDebtItem) string {
	categoryScores := make(map[string]float64)
	
	for _, item := range items {
		categoryScores[item.Category] += item.DebtScore
	}
	
	maxScore := 0.0
	recommendedCategory := "Code Quality"
	
	for category, score := range categoryScores {
		if score > maxScore {
			maxScore = score
			recommendedCategory = category
		}
	}
	
	return recommendedCategory
}

func (ds *DebtScorer) calculateOverallROI(items []TechnicalDebtItem) float64 {
	if len(items) == 0 {
		return 0.0
	}
	
	totalROI := 0.0
	for _, item := range items {
		totalROI += ds.calculateItemROI(item)
	}
	
	return totalROI / float64(len(items))
}

func (ds *DebtScorer) calculateItemROI(item TechnicalDebtItem) float64 {
	if item.EstimatedHours == 0 {
		return 0.0
	}
	
	// ROI = (Impact Score * Change Frequency) / Estimated Hours
	return (item.ImpactScore * item.ChangeFrequency) / item.EstimatedHours
}

// Placeholder implementations for complex functions
func (ds *DebtScorer) groupRelatedItems(items []TechnicalDebtItem) [][]TechnicalDebtItem {
	// Simple grouping by category for now
	groups := make(map[string][]TechnicalDebtItem)
	
	for _, item := range items {
		groups[item.Category] = append(groups[item.Category], item)
	}
	
	result := [][]TechnicalDebtItem{}
	for _, group := range groups {
		if len(group) > 0 {
			result = append(result, group)
		}
	}
	
	return result
}

func (ds *DebtScorer) generateRemediationTitle(group []TechnicalDebtItem) string {
	if len(group) == 0 {
		return "Remediation Task"
	}
	
	category := group[0].Category
	return fmt.Sprintf("Address %s Issues", category)
}

func (ds *DebtScorer) generateRemediationDescription(group []TechnicalDebtItem) string {
	if len(group) == 0 {
		return "No description available"
	}
	
	return fmt.Sprintf("Resolve %d %s issues affecting code quality and maintainability", len(group), group[0].Category)
}

func (ds *DebtScorer) calculateGroupPriority(group []TechnicalDebtItem) string {
	criticalCount := 0
	highCount := 0
	
	for _, item := range group {
		switch item.Priority {
		case "critical":
			criticalCount++
		case "high":
			highCount++
		}
	}
	
	if criticalCount > 0 {
		return "critical"
	} else if highCount > len(group)/2 {
		return "high"
	} else if highCount > 0 {
		return "medium"
	}
	return "low"
}

func (ds *DebtScorer) calculateGroupImpactScore(group []TechnicalDebtItem) float64 {
	totalImpact := 0.0
	for _, item := range group {
		totalImpact += item.ImpactScore
	}
	return totalImpact
}

func (ds *DebtScorer) mergeRemediationSteps(group []TechnicalDebtItem) []string {
	stepSet := make(map[string]bool)
	steps := []string{}
	
	for _, item := range group {
		for _, step := range item.RemediationSteps {
			if !stepSet[step] {
				stepSet[step] = true
				steps = append(steps, step)
			}
		}
	}
	
	return steps
}

func (ds *DebtScorer) generateSuccessMetrics(group []TechnicalDebtItem) []string {
	return []string{
		"Reduced complexity scores",
		"Improved maintainability index",
		"Decreased technical debt ratio",
		"Enhanced code quality metrics",
		"Reduced remediation effort estimates",
	}
}

func (ds *DebtScorer) generateCategoryRecommendation(categoryName string, category DebtCategory) DebtRecommendation {
	recommendation := DebtRecommendation{
		Type:        "strategic",
		Priority:    category.PriorityLevel,
		Title:       fmt.Sprintf("Address %s Issues", categoryName),
		Description: fmt.Sprintf("Focus on resolving %s issues with %.1f total debt score", categoryName, category.Score),
		Benefits: []string{
			"Improved code maintainability",
			"Reduced future development effort", 
			"Enhanced team productivity",
			"Better software quality",
		},
		Implementation: []string{
			"Prioritize high-impact items",
			"Establish coding standards",
			"Implement automated checks",
			"Provide team training",
		},
		EstimatedImpact:  category.Score * 0.8, // 80% debt reduction
		TimeframeWeeks:   int(math.Ceil(category.DebtHours / 40.0)),
		ResourcesNeeded:  []string{"Senior Developer", "Code Review Process"},
	}
	
	return recommendation
}

func (ds *DebtScorer) generateGeneralRecommendations(items []TechnicalDebtItem, categories map[string]DebtCategory) []DebtRecommendation {
	recommendations := []DebtRecommendation{}
	
	// Add process improvement recommendation
	processRec := DebtRecommendation{
		Type:        "process",
		Priority:    "medium",
		Title:       "Establish Technical Debt Management Process",
		Description: "Implement systematic approach to prevent and manage technical debt",
		Benefits: []string{
			"Proactive debt prevention",
			"Consistent quality standards",
			"Better resource planning",
		},
		Implementation: []string{
			"Define debt tracking process",
			"Establish quality gates",
			"Regular debt review sessions",
			"Team training on best practices",
		},
		EstimatedImpact:  25.0,
		TimeframeWeeks:   8,
		ResourcesNeeded:  []string{"Technical Lead", "Development Team"},
	}
	
	recommendations = append(recommendations, processRec)
	
	return recommendations
}