package metrics

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// ComplexityAnalyzer performs cyclomatic complexity analysis on JavaScript/TypeScript code
type ComplexityAnalyzer struct {
	config ComplexityConfig
}

// ComplexityConfig defines thresholds and settings for complexity analysis
type ComplexityConfig struct {
	LowThreshold    int     `yaml:"low_threshold" json:"low_threshold"`
	MediumThreshold int     `yaml:"medium_threshold" json:"medium_threshold"`
	HighThreshold   int     `yaml:"high_threshold" json:"high_threshold"`
	MaxNestingDepth int     `yaml:"max_nesting_depth" json:"max_nesting_depth"`
	ReportTopN      int     `yaml:"report_top_n" json:"report_top_n"`
	TrendPeriods    int     `yaml:"trend_periods" json:"trend_periods"`
	EnableTrends    bool    `yaml:"enable_trends" json:"enable_trends"`
	WeightFactors   Weights `yaml:"weight_factors" json:"weight_factors"`
}

// Weights for different complexity factors
type Weights struct {
	Cyclomatic   float64 `yaml:"cyclomatic" json:"cyclomatic"`
	Cognitive    float64 `yaml:"cognitive" json:"cognitive"`
	NestingDepth float64 `yaml:"nesting_depth" json:"nesting_depth"`
	NestedLoops  float64 `yaml:"nested_loops" json:"nested_loops"`
	Conditionals float64 `yaml:"conditionals" json:"conditionals"`
}

// ComplexityMetrics contains comprehensive complexity analysis results
type ComplexityMetrics struct {
	OverallScore      float64                    `json:"overall_score"`
	AverageComplexity float64                    `json:"average_complexity"`
	MaxComplexity     int                        `json:"max_complexity"`
	TotalFunctions    int                        `json:"total_functions"`
	ComplexityByLevel ComplexityBreakdown        `json:"complexity_by_level"`
	FunctionMetrics   []FunctionComplexity       `json:"function_metrics"`
	ClassMetrics      []ClassComplexity          `json:"class_metrics"`
	FileMetrics       map[string]FileComplexity  `json:"file_metrics"`
	TrendAnalysis     *ComplexityTrend           `json:"trend_analysis,omitempty"`
	Recommendations   []ComplexityRecommendation `json:"recommendations"`
	Summary           ComplexitySummary          `json:"summary"`
}

// ComplexityBreakdown categorizes functions by complexity level
type ComplexityBreakdown struct {
	Low    ComplexityLevel `json:"low"`
	Medium ComplexityLevel `json:"medium"`
	High   ComplexityLevel `json:"high"`
	Severe ComplexityLevel `json:"severe"`
}

// ComplexityLevel contains metrics for a specific complexity range
type ComplexityLevel struct {
	Count      int      `json:"count"`
	Percentage float64  `json:"percentage"`
	Functions  []string `json:"functions"`
}

// FunctionComplexity contains detailed complexity metrics for a single function
type FunctionComplexity struct {
	Name              string                 `json:"name"`
	FilePath          string                 `json:"file_path"`
	StartLine         int                    `json:"start_line"`
	EndLine           int                    `json:"end_line"`
	CyclomaticValue   int                    `json:"cyclomatic_value"`
	CognitiveValue    int                    `json:"cognitive_value"`
	NestingDepth      int                    `json:"nesting_depth"`
	SeverityLevel     string                 `json:"severity_level"`
	WeightedScore     float64                `json:"weighted_score"`
	ComplexityFactors ComplexityFactors      `json:"complexity_factors"`
	Recommendations   []string               `json:"recommendations"`
	RefactoringRisk   string                 `json:"refactoring_risk"`
	TestingDifficulty string                 `json:"testing_difficulty"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ComplexityFactors details what contributes to complexity
type ComplexityFactors struct {
	DecisionPoints   int      `json:"decision_points"`
	NestedLoops      int      `json:"nested_loops"`
	ConditionalDepth int      `json:"conditional_depth"`
	TryCatchBlocks   int      `json:"try_catch_blocks"`
	SwitchStatements int      `json:"switch_statements"`
	LogicalOperators int      `json:"logical_operators"`
	AntiPatterns     []string `json:"anti_patterns"`
}

// ClassComplexity aggregates complexity metrics for class methods
type ClassComplexity struct {
	Name            string               `json:"name"`
	FilePath        string               `json:"file_path"`
	TotalComplexity int                  `json:"total_complexity"`
	AverageMethod   float64              `json:"average_method_complexity"`
	MaxMethod       int                  `json:"max_method_complexity"`
	MethodCount     int                  `json:"method_count"`
	Methods         []FunctionComplexity `json:"methods"`
	OverallRisk     string               `json:"overall_risk"`
}

// FileComplexity aggregates complexity metrics at the file level
type FileComplexity struct {
	FilePath            string  `json:"file_path"`
	TotalComplexity     int     `json:"total_complexity"`
	AverageComplexity   float64 `json:"average_complexity"`
	FunctionCount       int     `json:"function_count"`
	ClassCount          int     `json:"class_count"`
	MaxComplexity       int     `json:"max_complexity"`
	ComplexityDensity   float64 `json:"complexity_density"`
	MaintainabilityRisk string  `json:"maintainability_risk"`
}

// ComplexityTrend tracks complexity changes over time
type ComplexityTrend struct {
	Direction      string    `json:"direction"` // increasing, decreasing, stable
	ChangeRate     float64   `json:"change_rate"`
	HistoricalData []float64 `json:"historical_data"`
	Prediction     float64   `json:"prediction"`
	HealthScore    float64   `json:"health_score"`
}

// ComplexityRecommendation provides actionable improvement suggestions
type ComplexityRecommendation struct {
	Priority       string   `json:"priority"` // critical, high, medium, low
	Category       string   `json:"category"` // refactoring, testing, architecture
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Impact         string   `json:"impact"` // high, medium, low
	Effort         string   `json:"effort"` // high, medium, low
	Functions      []string `json:"functions"`
	Techniques     []string `json:"techniques"`
	EstimatedHours int      `json:"estimated_hours"`
}

// ComplexitySummary provides executive-level overview
type ComplexitySummary struct {
	HealthScore       float64 `json:"health_score"`       // 0-100
	RiskLevel         string  `json:"risk_level"`         // low, medium, high, critical
	TechnicalDebt     string  `json:"technical_debt"`     // low, medium, high, critical
	RefactoringNeeded int     `json:"refactoring_needed"` // number of functions
	TestingGaps       int     `json:"testing_gaps"`       // hard to test functions
	MaintenanceRisk   string  `json:"maintenance_risk"`   // low, medium, high, critical
}

// NewComplexityAnalyzer creates a new complexity analyzer with default configuration
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		config: ComplexityConfig{
			LowThreshold:    10,
			MediumThreshold: 15,
			HighThreshold:   20,
			MaxNestingDepth: 4,
			ReportTopN:      20,
			TrendPeriods:    5,
			EnableTrends:    true,
			WeightFactors: Weights{
				Cyclomatic:   0.4,
				Cognitive:    0.3,
				NestingDepth: 0.2,
				NestedLoops:  0.05,
				Conditionals: 0.05,
			},
		},
	}
}

// NewComplexityAnalyzerWithConfig creates analyzer with custom configuration
func NewComplexityAnalyzerWithConfig(config ComplexityConfig) *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		config: config,
	}
}

// AnalyzeComplexity performs comprehensive complexity analysis on parsed AST results
func (ca *ComplexityAnalyzer) AnalyzeComplexity(ctx context.Context, parseResults []*ast.ParseResult) (*ComplexityMetrics, error) {
	if len(parseResults) == 0 {
		return nil, fmt.Errorf("no parse results provided for complexity analysis")
	}

	metrics := &ComplexityMetrics{
		FunctionMetrics: []FunctionComplexity{},
		ClassMetrics:    []ClassComplexity{},
		FileMetrics:     make(map[string]FileComplexity),
		ComplexityByLevel: ComplexityBreakdown{
			Low:    ComplexityLevel{Functions: []string{}},
			Medium: ComplexityLevel{Functions: []string{}},
			High:   ComplexityLevel{Functions: []string{}},
			Severe: ComplexityLevel{Functions: []string{}},
		},
		Recommendations: []ComplexityRecommendation{},
	}

	// Analyze each parsed file
	for _, parseResult := range parseResults {
		if err := ca.analyzeFile(ctx, parseResult, metrics); err != nil {
			return nil, fmt.Errorf("failed to analyze file %s: %w", parseResult.FilePath, err)
		}
	}

	// Calculate aggregate metrics
	ca.calculateAggregateMetrics(metrics)

	// Generate recommendations
	ca.generateRecommendations(metrics)

	// Perform trend analysis if enabled
	if ca.config.EnableTrends {
		ca.performTrendAnalysis(metrics)
	}

	// Generate summary
	ca.generateSummary(metrics)

	return metrics, nil
}

// analyzeFile analyzes complexity for a single file
func (ca *ComplexityAnalyzer) analyzeFile(ctx context.Context, parseResult *ast.ParseResult, metrics *ComplexityMetrics) error {
	fileMetric := FileComplexity{
		FilePath: parseResult.FilePath,
	}

	// Analyze standalone functions
	for _, function := range parseResult.Functions {
		complexity, err := ca.analyzeFunctionComplexity(function, parseResult)
		if err != nil {
			return fmt.Errorf("failed to analyze function %s: %w", function.Name, err)
		}

		metrics.FunctionMetrics = append(metrics.FunctionMetrics, *complexity)
		ca.categorizeFunction(complexity, &metrics.ComplexityByLevel)

		fileMetric.TotalComplexity += complexity.CyclomaticValue
		fileMetric.FunctionCount++

		if complexity.CyclomaticValue > fileMetric.MaxComplexity {
			fileMetric.MaxComplexity = complexity.CyclomaticValue
		}
	}

	// Analyze class methods
	for _, class := range parseResult.Classes {
		classMetric, err := ca.analyzeClassComplexity(class, parseResult)
		if err != nil {
			return fmt.Errorf("failed to analyze class %s: %w", class.Name, err)
		}

		metrics.ClassMetrics = append(metrics.ClassMetrics, *classMetric)

		// Add class methods to overall function metrics
		for _, method := range classMetric.Methods {
			metrics.FunctionMetrics = append(metrics.FunctionMetrics, method)
			ca.categorizeFunction(&method, &metrics.ComplexityByLevel)
		}

		fileMetric.TotalComplexity += classMetric.TotalComplexity
		fileMetric.ClassCount++
		fileMetric.FunctionCount += classMetric.MethodCount

		if classMetric.MaxMethod > fileMetric.MaxComplexity {
			fileMetric.MaxComplexity = classMetric.MaxMethod
		}
	}

	// Calculate file-level metrics
	if fileMetric.FunctionCount > 0 {
		fileMetric.AverageComplexity = float64(fileMetric.TotalComplexity) / float64(fileMetric.FunctionCount)
		fileMetric.ComplexityDensity = float64(fileMetric.TotalComplexity) / float64(len(parseResult.Functions)+len(parseResult.Classes))
		fileMetric.MaintainabilityRisk = ca.assessMaintainabilityRisk(fileMetric.AverageComplexity, fileMetric.MaxComplexity)
	}

	metrics.FileMetrics[parseResult.FilePath] = fileMetric
	return nil
}

// analyzeFunctionComplexity calculates detailed complexity metrics for a function
func (ca *ComplexityAnalyzer) analyzeFunctionComplexity(function ast.FunctionInfo, parseResult *ast.ParseResult) (*FunctionComplexity, error) {
	complexity := &FunctionComplexity{
		Name:      function.Name,
		FilePath:  parseResult.FilePath,
		StartLine: function.StartLine,
		EndLine:   function.EndLine,
		ComplexityFactors: ComplexityFactors{
			AntiPatterns: []string{},
		},
		Recommendations: []string{},
		Metadata:        make(map[string]interface{}),
	}

	// For this implementation, we'll calculate based on available metadata
	// In a real implementation, we'd need to parse the function body AST

	// Basic cyclomatic complexity calculation (simplified)
	complexity.CyclomaticValue = ca.calculateBasicComplexity(function)
	complexity.CognitiveValue = ca.calculateCognitiveComplexity(function)
	complexity.NestingDepth = ca.calculateNestingDepth(function)

	// Calculate weighted score
	complexity.WeightedScore = ca.calculateWeightedScore(complexity)

	// Determine severity level
	complexity.SeverityLevel = ca.determineSeverityLevel(complexity.CyclomaticValue)

	// Assess risks
	complexity.RefactoringRisk = ca.assessRefactoringRisk(complexity)
	complexity.TestingDifficulty = ca.assessTestingDifficulty(complexity)

	// Generate recommendations
	complexity.Recommendations = ca.generateFunctionRecommendations(complexity)

	// Detect anti-patterns
	complexity.ComplexityFactors.AntiPatterns = ca.detectAntiPatterns(function, complexity)

	return complexity, nil
}

// analyzeClassComplexity calculates complexity metrics for a class
func (ca *ComplexityAnalyzer) analyzeClassComplexity(class ast.ClassInfo, parseResult *ast.ParseResult) (*ClassComplexity, error) {
	classMetric := &ClassComplexity{
		Name:     class.Name,
		FilePath: parseResult.FilePath,
		Methods:  []FunctionComplexity{},
	}

	totalComplexity := 0
	maxMethodComplexity := 0

	for _, method := range class.Methods {
		methodComplexity, err := ca.analyzeFunctionComplexity(method, parseResult)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze method %s: %w", method.Name, err)
		}

		classMetric.Methods = append(classMetric.Methods, *methodComplexity)
		totalComplexity += methodComplexity.CyclomaticValue

		if methodComplexity.CyclomaticValue > maxMethodComplexity {
			maxMethodComplexity = methodComplexity.CyclomaticValue
		}
	}

	classMetric.TotalComplexity = totalComplexity
	classMetric.MethodCount = len(class.Methods)
	classMetric.MaxMethod = maxMethodComplexity

	if classMetric.MethodCount > 0 {
		classMetric.AverageMethod = float64(totalComplexity) / float64(classMetric.MethodCount)
	}

	classMetric.OverallRisk = ca.assessClassRisk(classMetric)

	return classMetric, nil
}

// calculateBasicComplexity performs basic cyclomatic complexity calculation
func (ca *ComplexityAnalyzer) calculateBasicComplexity(function ast.FunctionInfo) int {
	// Start with base complexity of 1
	complexity := 1

	// Add complexity for parameters (indication of decision points)
	complexity += len(function.Parameters) / 3 // Rough heuristic

	// Add complexity for async functions (additional error paths)
	if function.IsAsync {
		complexity += 1
	}

	// Additional complexity based on function size (lines)
	lines := function.EndLine - function.StartLine + 1
	if lines > 50 {
		complexity += 2
	} else if lines > 20 {
		complexity += 1
	}

	return complexity
}

// calculateCognitiveComplexity calculates cognitive complexity
func (ca *ComplexityAnalyzer) calculateCognitiveComplexity(function ast.FunctionInfo) int {
	// Simplified cognitive complexity calculation
	cognitive := ca.calculateBasicComplexity(function)

	// Add penalties for nested structures (estimated)
	if function.EndLine-function.StartLine > 30 {
		cognitive += 2 // Likely has nested structures
	}

	return cognitive
}

// calculateNestingDepth estimates maximum nesting depth
func (ca *ComplexityAnalyzer) calculateNestingDepth(function ast.FunctionInfo) int {
	// Simplified nesting depth calculation based on function size
	lines := function.EndLine - function.StartLine + 1

	if lines > 100 {
		return 5 // Likely deeply nested
	} else if lines > 50 {
		return 4
	} else if lines > 30 {
		return 3
	} else if lines > 15 {
		return 2
	}

	return 1
}

// calculateWeightedScore calculates composite complexity score
func (ca *ComplexityAnalyzer) calculateWeightedScore(complexity *FunctionComplexity) float64 {
	weights := ca.config.WeightFactors

	score := float64(complexity.CyclomaticValue)*weights.Cyclomatic +
		float64(complexity.CognitiveValue)*weights.Cognitive +
		float64(complexity.NestingDepth)*weights.NestingDepth +
		float64(complexity.ComplexityFactors.NestedLoops)*weights.NestedLoops +
		float64(complexity.ComplexityFactors.DecisionPoints)*weights.Conditionals

	return math.Round(score*100) / 100
}

// determineSeverityLevel categorizes complexity level
func (ca *ComplexityAnalyzer) determineSeverityLevel(complexity int) string {
	if complexity >= ca.config.HighThreshold {
		return "severe"
	} else if complexity >= ca.config.MediumThreshold {
		return "high"
	} else if complexity >= ca.config.LowThreshold {
		return "medium"
	}
	return "low"
}

// assessRefactoringRisk evaluates refactoring difficulty
func (ca *ComplexityAnalyzer) assessRefactoringRisk(complexity *FunctionComplexity) string {
	if complexity.CyclomaticValue > 25 || complexity.NestingDepth > 4 {
		return "critical"
	} else if complexity.CyclomaticValue > 15 || complexity.NestingDepth > 3 {
		return "high"
	} else if complexity.CyclomaticValue > 10 {
		return "medium"
	}
	return "low"
}

// assessTestingDifficulty evaluates testing complexity
func (ca *ComplexityAnalyzer) assessTestingDifficulty(complexity *FunctionComplexity) string {
	score := complexity.CyclomaticValue + complexity.NestingDepth*2

	if score > 30 {
		return "very_difficult"
	} else if score > 20 {
		return "difficult"
	} else if score > 10 {
		return "moderate"
	}
	return "easy"
}

// generateFunctionRecommendations creates specific recommendations for a function
func (ca *ComplexityAnalyzer) generateFunctionRecommendations(complexity *FunctionComplexity) []string {
	recommendations := []string{}

	if complexity.CyclomaticValue > ca.config.HighThreshold {
		recommendations = append(recommendations, "Consider breaking this function into smaller, more focused functions")
		recommendations = append(recommendations, "Extract complex conditional logic into separate helper functions")
	}

	if complexity.NestingDepth > ca.config.MaxNestingDepth {
		recommendations = append(recommendations, "Reduce nesting depth using early returns or guard clauses")
		recommendations = append(recommendations, "Consider using the strategy pattern for complex conditional logic")
	}

	if complexity.TestingDifficulty == "very_difficult" || complexity.TestingDifficulty == "difficult" {
		recommendations = append(recommendations, "Add unit tests with comprehensive edge case coverage")
		recommendations = append(recommendations, "Consider dependency injection to improve testability")
	}

	return recommendations
}

// detectAntiPatterns identifies common complexity anti-patterns
func (ca *ComplexityAnalyzer) detectAntiPatterns(function ast.FunctionInfo, complexity *FunctionComplexity) []string {
	patterns := []string{}

	// Large function anti-pattern
	if function.EndLine-function.StartLine > 100 {
		patterns = append(patterns, "large_function")
	}

	// High complexity anti-pattern
	if complexity.CyclomaticValue > 20 {
		patterns = append(patterns, "high_complexity")
	}

	// Deep nesting anti-pattern
	if complexity.NestingDepth > 4 {
		patterns = append(patterns, "deep_nesting")
	}

	// Too many parameters
	if len(function.Parameters) > 7 {
		patterns = append(patterns, "too_many_parameters")
	}

	return patterns
}

// categorizeFunction places function into appropriate complexity category
func (ca *ComplexityAnalyzer) categorizeFunction(complexity *FunctionComplexity, breakdown *ComplexityBreakdown) {
	functionName := fmt.Sprintf("%s (%s:%d)", complexity.Name, complexity.FilePath, complexity.StartLine)

	switch complexity.SeverityLevel {
	case "low":
		breakdown.Low.Count++
		breakdown.Low.Functions = append(breakdown.Low.Functions, functionName)
	case "medium":
		breakdown.Medium.Count++
		breakdown.Medium.Functions = append(breakdown.Medium.Functions, functionName)
	case "high":
		breakdown.High.Count++
		breakdown.High.Functions = append(breakdown.High.Functions, functionName)
	case "severe":
		breakdown.Severe.Count++
		breakdown.Severe.Functions = append(breakdown.Severe.Functions, functionName)
	}
}

// calculateAggregateMetrics computes overall complexity statistics
func (ca *ComplexityAnalyzer) calculateAggregateMetrics(metrics *ComplexityMetrics) {
	if len(metrics.FunctionMetrics) == 0 {
		return
	}

	totalComplexity := 0
	maxComplexity := 0

	for _, function := range metrics.FunctionMetrics {
		totalComplexity += function.CyclomaticValue
		if function.CyclomaticValue > maxComplexity {
			maxComplexity = function.CyclomaticValue
		}
	}

	metrics.TotalFunctions = len(metrics.FunctionMetrics)
	metrics.AverageComplexity = float64(totalComplexity) / float64(metrics.TotalFunctions)
	metrics.MaxComplexity = maxComplexity

	// Calculate percentages for complexity breakdown
	total := float64(metrics.TotalFunctions)
	metrics.ComplexityByLevel.Low.Percentage = float64(metrics.ComplexityByLevel.Low.Count) / total * 100
	metrics.ComplexityByLevel.Medium.Percentage = float64(metrics.ComplexityByLevel.Medium.Count) / total * 100
	metrics.ComplexityByLevel.High.Percentage = float64(metrics.ComplexityByLevel.High.Count) / total * 100
	metrics.ComplexityByLevel.Severe.Percentage = float64(metrics.ComplexityByLevel.Severe.Count) / total * 100

	// Calculate overall score (inverse of average complexity, normalized)
	metrics.OverallScore = math.Max(0, 100-(metrics.AverageComplexity*5))
}

// generateRecommendations creates prioritized improvement recommendations
func (ca *ComplexityAnalyzer) generateRecommendations(metrics *ComplexityMetrics) {
	// Sort functions by complexity for targeted recommendations
	functions := make([]FunctionComplexity, len(metrics.FunctionMetrics))
	copy(functions, metrics.FunctionMetrics)

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].WeightedScore > functions[j].WeightedScore
	})

	// Generate high-priority recommendations for most complex functions
	topN := ca.config.ReportTopN
	if len(functions) < topN {
		topN = len(functions)
	}

	if topN > 0 {
		criticalFunctions := []string{}
		for i := 0; i < topN && i < len(functions); i++ {
			if functions[i].SeverityLevel == "severe" || functions[i].SeverityLevel == "high" {
				criticalFunctions = append(criticalFunctions, functions[i].Name)
			}
		}

		if len(criticalFunctions) > 0 {
			metrics.Recommendations = append(metrics.Recommendations, ComplexityRecommendation{
				Priority:       "critical",
				Category:       "refactoring",
				Title:          "Refactor High-Complexity Functions",
				Description:    fmt.Sprintf("Refactor %d functions with severe/high complexity", len(criticalFunctions)),
				Impact:         "high",
				Effort:         "medium",
				Functions:      criticalFunctions,
				Techniques:     []string{"extract_method", "reduce_nesting", "simplify_conditionals"},
				EstimatedHours: len(criticalFunctions) * 4,
			})
		}
	}

	// Add testing recommendations
	if metrics.ComplexityByLevel.High.Count+metrics.ComplexityByLevel.Severe.Count > 0 {
		metrics.Recommendations = append(metrics.Recommendations, ComplexityRecommendation{
			Priority:       "high",
			Category:       "testing",
			Title:          "Improve Test Coverage for Complex Functions",
			Description:    "Add comprehensive tests for functions with high/severe complexity",
			Impact:         "high",
			Effort:         "medium",
			Functions:      append(metrics.ComplexityByLevel.High.Functions, metrics.ComplexityByLevel.Severe.Functions...),
			Techniques:     []string{"unit_testing", "edge_case_testing", "integration_testing"},
			EstimatedHours: (metrics.ComplexityByLevel.High.Count + metrics.ComplexityByLevel.Severe.Count) * 2,
		})
	}
}

// performTrendAnalysis analyzes complexity trends over time
func (ca *ComplexityAnalyzer) performTrendAnalysis(metrics *ComplexityMetrics) {
	// Simplified trend analysis - in practice, this would use historical data
	trend := &ComplexityTrend{
		Direction:      "stable",
		ChangeRate:     0.0,
		HistoricalData: []float64{metrics.AverageComplexity},
		Prediction:     metrics.AverageComplexity,
		HealthScore:    metrics.OverallScore,
	}

	// Determine health score based on complexity distribution
	if metrics.ComplexityByLevel.Severe.Percentage > 10 {
		trend.HealthScore = 30
		trend.Direction = "concerning"
	} else if metrics.ComplexityByLevel.High.Percentage > 20 {
		trend.HealthScore = 60
		trend.Direction = "needs_attention"
	} else if metrics.ComplexityByLevel.Low.Percentage > 70 {
		trend.HealthScore = 90
		trend.Direction = "healthy"
	}

	metrics.TrendAnalysis = trend
}

// generateSummary creates executive-level summary
func (ca *ComplexityAnalyzer) generateSummary(metrics *ComplexityMetrics) {
	summary := ComplexitySummary{
		HealthScore:       metrics.OverallScore,
		RefactoringNeeded: metrics.ComplexityByLevel.High.Count + metrics.ComplexityByLevel.Severe.Count,
	}

	// Determine risk levels
	if metrics.ComplexityByLevel.Severe.Percentage > 5 {
		summary.RiskLevel = "critical"
		summary.TechnicalDebt = "critical"
		summary.MaintenanceRisk = "critical"
	} else if metrics.ComplexityByLevel.High.Percentage > 15 {
		summary.RiskLevel = "high"
		summary.TechnicalDebt = "high"
		summary.MaintenanceRisk = "high"
	} else if metrics.ComplexityByLevel.Medium.Percentage > 40 {
		summary.RiskLevel = "medium"
		summary.TechnicalDebt = "medium"
		summary.MaintenanceRisk = "medium"
	} else {
		summary.RiskLevel = "low"
		summary.TechnicalDebt = "low"
		summary.MaintenanceRisk = "low"
	}

	// Count functions that are difficult to test
	testingGaps := 0
	for _, function := range metrics.FunctionMetrics {
		if function.TestingDifficulty == "very_difficult" || function.TestingDifficulty == "difficult" {
			testingGaps++
		}
	}
	summary.TestingGaps = testingGaps

	metrics.Summary = summary
}

// assessMaintainabilityRisk evaluates file-level maintenance difficulty
func (ca *ComplexityAnalyzer) assessMaintainabilityRisk(avgComplexity float64, maxComplexity int) string {
	if avgComplexity > 15 || maxComplexity > 25 {
		return "critical"
	} else if avgComplexity > 10 || maxComplexity > 20 {
		return "high"
	} else if avgComplexity > 7 || maxComplexity > 15 {
		return "medium"
	}
	return "low"
}

// assessClassRisk evaluates overall class complexity risk
func (ca *ComplexityAnalyzer) assessClassRisk(classMetric *ClassComplexity) string {
	if classMetric.AverageMethod > 15 || classMetric.MaxMethod > 25 {
		return "critical"
	} else if classMetric.AverageMethod > 10 || classMetric.MaxMethod > 20 {
		return "high"
	} else if classMetric.AverageMethod > 7 || classMetric.MaxMethod > 15 {
		return "medium"
	}
	return "low"
}
