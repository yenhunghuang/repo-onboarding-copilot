package metrics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// CoverageAnalyzer provides comprehensive code coverage analysis and testability assessment
type CoverageAnalyzer struct {
	config CoverageConfig
}

// CoverageConfig defines configuration parameters for coverage analysis
type CoverageConfig struct {
	// Complexity thresholds for testability assessment
	LowComplexityThreshold  int `yaml:"low_complexity_threshold" default:"5"`
	HighComplexityThreshold int `yaml:"high_complexity_threshold" default:"15"`

	// Dependency coupling thresholds
	LowCouplingThreshold  int `yaml:"low_coupling_threshold" default:"3"`
	HighCouplingThreshold int `yaml:"high_coupling_threshold" default:"8"`

	// Testability scoring weights
	ComplexityWeight float64 `yaml:"complexity_weight" default:"0.30"`
	CouplingWeight   float64 `yaml:"coupling_weight" default:"0.25"`
	DependencyWeight float64 `yaml:"dependency_weight" default:"0.20"`
	SizeWeight       float64 `yaml:"size_weight" default:"0.15"`
	PatternWeight    float64 `yaml:"pattern_weight" default:"0.10"`

	// Risk assessment parameters
	CriticalRiskThreshold float64 `yaml:"critical_risk_threshold" default:"80.0"`
	HighRiskThreshold     float64 `yaml:"high_risk_threshold" default:"60.0"`
	MediumRiskThreshold   float64 `yaml:"medium_risk_threshold" default:"40.0"`

	// Mock requirement thresholds
	ExternalDependencyThreshold int `yaml:"external_dependency_threshold" default:"2"`
	DatabaseCallThreshold       int `yaml:"database_call_threshold" default:"1"`
	NetworkCallThreshold        int `yaml:"network_call_threshold" default:"1"`
}

// CoverageMetrics contains comprehensive coverage analysis results
type CoverageMetrics struct {
	OverallScore           float64                    `json:"overall_score"`      // 0-100
	EstimatedCoverage      float64                    `json:"estimated_coverage"` // 0-100
	TestabilityScore       float64                    `json:"testability_score"`  // 0-100
	FunctionAnalysis       []FunctionTestability      `json:"function_analysis"`
	FileAnalysis           map[string]FileTestability `json:"file_analysis"`
	UntestedPaths          []UntestedPath             `json:"untested_paths"`
	MockRequirements       []MockRequirement          `json:"mock_requirements"`
	TestingRecommendations []TestingRecommendation    `json:"testing_recommendations"`
	CoverageGaps           []CoverageGap              `json:"coverage_gaps"`
	TestingStrategy        TestingStrategy            `json:"testing_strategy"`
	PriorityMatrix         TestingPriorityMatrix      `json:"priority_matrix"`
	Summary                CoverageSummary            `json:"summary"`
}

// FunctionTestability represents testability analysis for a function
type FunctionTestability struct {
	Name                string                 `json:"name"`
	FilePath            string                 `json:"file_path"`
	StartLine           int                    `json:"start_line"`
	EndLine             int                    `json:"end_line"`
	TestabilityScore    float64                `json:"testability_score"`  // 0-100
	EstimatedCoverage   float64                `json:"estimated_coverage"` // 0-100
	ComplexityFactor    float64                `json:"complexity_factor"`
	CouplingFactor      float64                `json:"coupling_factor"`
	DependencyFactor    float64                `json:"dependency_factor"`
	TestingDifficulty   string                 `json:"testing_difficulty"` // easy, moderate, difficult, very_difficult
	RequiredMocks       []string               `json:"required_mocks"`
	UntestedPaths       []string               `json:"untested_paths"`
	RiskLevel           string                 `json:"risk_level"`     // low, medium, high, critical
	TestingEffort       int                    `json:"testing_effort"` // hours
	RecommendedApproach string                 `json:"recommended_approach"`
	CoverageGaps        []string               `json:"coverage_gaps"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// FileTestability represents testability analysis at the file level
type FileTestability struct {
	FilePath               string   `json:"file_path"`
	OverallScore           float64  `json:"overall_score"`
	TestedFunctions        int      `json:"tested_functions"`
	UntestedFunctions      int      `json:"untested_functions"`
	TestingComplexity      float64  `json:"testing_complexity"`
	MockComplexity         int      `json:"mock_complexity"`
	RecommendedPriority    string   `json:"recommended_priority"` // low, medium, high, critical
	EstimatedEffort        int      `json:"estimated_effort"`     // hours
	CoverageGapCount       int      `json:"coverage_gap_count"`
	TestingRecommendations []string `json:"testing_recommendations"`
}

// UntestedPath represents an identified untested code path
type UntestedPath struct {
	ID              string   `json:"id"`
	FilePath        string   `json:"file_path"`
	FunctionName    string   `json:"function_name"`
	PathType        string   `json:"path_type"` // conditional, loop, exception, async
	StartLine       int      `json:"start_line"`
	EndLine         int      `json:"end_line"`
	Condition       string   `json:"condition"`
	RiskLevel       string   `json:"risk_level"` // low, medium, high, critical
	TestingStrategy string   `json:"testing_strategy"`
	RequiredSetup   []string `json:"required_setup"`
	ExpectedOutcome string   `json:"expected_outcome"`
}

// MockRequirement represents analysis of mocking needs for external dependencies
type MockRequirement struct {
	ID                  string                 `json:"id"`
	FilePath            string                 `json:"file_path"`
	FunctionName        string                 `json:"function_name"`
	DependencyType      string                 `json:"dependency_type"` // database, network, filesystem, external_api
	DependencyName      string                 `json:"dependency_name"`
	MockingComplexity   string                 `json:"mocking_complexity"` // simple, moderate, complex, very_complex
	MockingStrategy     string                 `json:"mocking_strategy"`
	RequiredMockLibrary string                 `json:"required_mock_library"`
	SetupComplexity     int                    `json:"setup_complexity"`     // 1-10
	MaintenanceOverhead string                 `json:"maintenance_overhead"` // low, medium, high
	Alternatives        []string               `json:"alternatives"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// TestingRecommendation provides specific testing guidance
type TestingRecommendation struct {
	ID              string                 `json:"id"`
	Priority        string                 `json:"priority"` // low, medium, high, critical
	Category        string                 `json:"category"` // unit, integration, e2e, performance
	FilePath        string                 `json:"file_path"`
	FunctionName    string                 `json:"function_name"`
	Recommendation  string                 `json:"recommendation"`
	Rationale       string                 `json:"rationale"`
	EstimatedEffort int                    `json:"estimated_effort"` // hours
	ExpectedBenefit string                 `json:"expected_benefit"`
	TestingApproach []string               `json:"testing_approach"`
	RequiredTools   []string               `json:"required_tools"`
	RiskReduction   float64                `json:"risk_reduction"` // 0-100
	Metadata        map[string]interface{} `json:"metadata"`
}

// CoverageGap represents identified gaps in test coverage
type CoverageGap struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"` // function, branch, path, exception
	FilePath        string   `json:"file_path"`
	Location        string   `json:"location"`
	Severity        string   `json:"severity"` // low, medium, high, critical
	Impact          string   `json:"impact"`
	RiskAssessment  string   `json:"risk_assessment"`
	TestingStrategy string   `json:"testing_strategy"`
	EstimatedEffort int      `json:"estimated_effort"`
	Prerequisites   []string `json:"prerequisites"`
}

// TestingStrategy provides overall testing approach recommendations
type TestingStrategy struct {
	OverallApproach      string                `json:"overall_approach"`
	RecommendedFramework string                `json:"recommended_framework"`
	TestingPyramid       TestingPyramid        `json:"testing_pyramid"`
	PriorityOrder        []string              `json:"priority_order"`
	PhaseRecommendations []PhaseRecommendation `json:"phase_recommendations"`
	ResourceRequirements ResourceRequirements  `json:"resource_requirements"`
	TimelineEstimate     TimelineEstimate      `json:"timeline_estimate"`
}

// TestingPyramid defines the recommended distribution of test types
type TestingPyramid struct {
	UnitTestPercentage        float64 `json:"unit_test_percentage"`
	IntegrationTestPercentage float64 `json:"integration_test_percentage"`
	E2ETestPercentage         float64 `json:"e2e_test_percentage"`
	PerformanceTestPercentage float64 `json:"performance_test_percentage"`
}

// PhaseRecommendation provides phase-based implementation guidance
type PhaseRecommendation struct {
	Phase          string   `json:"phase"` // immediate, short_term, long_term
	Description    string   `json:"description"`
	Objectives     []string `json:"objectives"`
	Deliverables   []string `json:"deliverables"`
	EstimatedWeeks int      `json:"estimated_weeks"`
	RequiredSkills []string `json:"required_skills"`
}

// ResourceRequirements defines testing resource needs
type ResourceRequirements struct {
	DeveloperHours      int      `json:"developer_hours"`
	QAHours             int      `json:"qa_hours"`
	InfrastructureNeed  []string `json:"infrastructure_need"`
	ToolingRequirements []string `json:"tooling_requirements"`
	SkillGaps           []string `json:"skill_gaps"`
}

// TimelineEstimate provides implementation timeline
type TimelineEstimate struct {
	ImmediatePhaseWeeks int `json:"immediate_phase_weeks"`
	ShortTermPhaseWeeks int `json:"short_term_phase_weeks"`
	LongTermPhaseWeeks  int `json:"long_term_phase_weeks"`
	TotalWeeks          int `json:"total_weeks"`
}

// TestingPriorityMatrix provides prioritized testing recommendations
type TestingPriorityMatrix struct {
	CriticalPriority []TestingItem `json:"critical_priority"`
	HighPriority     []TestingItem `json:"high_priority"`
	MediumPriority   []TestingItem `json:"medium_priority"`
	LowPriority      []TestingItem `json:"low_priority"`
}

// TestingItem represents a prioritized testing task
type TestingItem struct {
	FilePath       string  `json:"file_path"`
	FunctionName   string  `json:"function_name"`
	TestingType    string  `json:"testing_type"` // unit, integration, e2e
	RiskScore      float64 `json:"risk_score"`
	EffortEstimate int     `json:"effort_estimate"`
	ImpactScore    float64 `json:"impact_score"`
	ROI            float64 `json:"roi"` // impact/effort ratio
}

// CoverageSummary provides high-level coverage analysis summary
type CoverageSummary struct {
	TotalFunctions        int     `json:"total_functions"`
	TestedFunctions       int     `json:"tested_functions"`
	UntestedFunctions     int     `json:"untested_functions"`
	CoveragePercentage    float64 `json:"coverage_percentage"`
	TestabilityScore      float64 `json:"testability_score"`
	HighRiskFunctions     int     `json:"high_risk_functions"`
	MockingComplexity     string  `json:"mocking_complexity"` // low, medium, high, very_high
	EstimatedTestingWeeks int     `json:"estimated_testing_weeks"`
	RecommendedFocus      string  `json:"recommended_focus"`
	QualityGate           string  `json:"quality_gate"` // pass, warning, fail
}

// NewCoverageAnalyzer creates a new coverage analyzer with default configuration
func NewCoverageAnalyzer() *CoverageAnalyzer {
	return &CoverageAnalyzer{
		config: CoverageConfig{
			LowComplexityThreshold:      5,
			HighComplexityThreshold:     15,
			LowCouplingThreshold:        3,
			HighCouplingThreshold:       8,
			ComplexityWeight:            0.30,
			CouplingWeight:              0.25,
			DependencyWeight:            0.20,
			SizeWeight:                  0.15,
			PatternWeight:               0.10,
			CriticalRiskThreshold:       80.0,
			HighRiskThreshold:           60.0,
			MediumRiskThreshold:         40.0,
			ExternalDependencyThreshold: 2,
			DatabaseCallThreshold:       1,
			NetworkCallThreshold:        1,
		},
	}
}

// NewCoverageAnalyzerWithConfig creates a new coverage analyzer with custom configuration
func NewCoverageAnalyzerWithConfig(config CoverageConfig) *CoverageAnalyzer {
	return &CoverageAnalyzer{
		config: config,
	}
}

// AnalyzeCoverage performs comprehensive coverage analysis on parsed results
func (ca *CoverageAnalyzer) AnalyzeCoverage(ctx context.Context, parseResults []*ast.ParseResult, complexityMetrics *ComplexityMetrics) (*CoverageMetrics, error) {
	if len(parseResults) == 0 {
		return &CoverageMetrics{
			Summary: CoverageSummary{
				QualityGate: "pass",
			},
		}, nil
	}

	metrics := &CoverageMetrics{
		FunctionAnalysis:       []FunctionTestability{},
		FileAnalysis:           make(map[string]FileTestability),
		UntestedPaths:          []UntestedPath{},
		MockRequirements:       []MockRequirement{},
		TestingRecommendations: []TestingRecommendation{},
		CoverageGaps:           []CoverageGap{},
	}

	// Analyze each file for testability
	for _, parseResult := range parseResults {
		if err := ca.analyzeFileTestability(parseResult, complexityMetrics, metrics); err != nil {
			return nil, fmt.Errorf("failed to analyze file %s: %w", parseResult.FilePath, err)
		}
	}

	// Identify untested paths through AST analysis
	ca.identifyUntestedPaths(parseResults, metrics)

	// Analyze mock requirements for external dependencies
	ca.analyzeMockRequirements(parseResults, metrics)

	// Generate testing recommendations based on risk and complexity
	ca.generateTestingRecommendations(metrics)

	// Create testing priority matrix
	ca.createTestingPriorityMatrix(metrics)

	// Generate testing strategy
	ca.generateTestingStrategy(metrics)

	// Calculate overall metrics and summary
	ca.calculateOverallMetrics(metrics)

	return metrics, nil
}

// analyzeFileTestability analyzes testability for a single file
func (ca *CoverageAnalyzer) analyzeFileTestability(parseResult *ast.ParseResult, complexityMetrics *ComplexityMetrics, metrics *CoverageMetrics) error {
	fileTestability := FileTestability{
		FilePath:               parseResult.FilePath,
		TestingRecommendations: []string{},
	}

	totalScore := 0.0
	functionCount := 0
	totalEffort := 0
	untestedCount := 0
	mockComplexity := 0

	// Analyze each function in the file
	for _, function := range parseResult.Functions {
		functionTestability := ca.analyzeFunctionTestability(function, parseResult, complexityMetrics)
		metrics.FunctionAnalysis = append(metrics.FunctionAnalysis, functionTestability)

		totalScore += functionTestability.TestabilityScore
		functionCount++
		totalEffort += functionTestability.TestingEffort

		if functionTestability.TestabilityScore < 60.0 {
			untestedCount++
		}

		mockComplexity += len(functionTestability.RequiredMocks)
	}

	// Calculate file-level metrics
	if functionCount > 0 {
		fileTestability.OverallScore = totalScore / float64(functionCount)
		fileTestability.TestedFunctions = functionCount - untestedCount
		fileTestability.UntestedFunctions = untestedCount
		fileTestability.TestingComplexity = ca.calculateFileTestingComplexity(parseResult)
		fileTestability.MockComplexity = mockComplexity
		fileTestability.EstimatedEffort = totalEffort
		fileTestability.RecommendedPriority = ca.determineFilePriority(fileTestability)
		fileTestability.CoverageGapCount = ca.countCoverageGaps(parseResult)

		// Generate file-level recommendations
		fileTestability.TestingRecommendations = ca.generateFileTestingRecommendations(fileTestability)
	}

	metrics.FileAnalysis[parseResult.FilePath] = fileTestability
	return nil
}

// analyzeFunctionTestability performs detailed testability analysis for a function
func (ca *CoverageAnalyzer) analyzeFunctionTestability(function ast.FunctionInfo, parseResult *ast.ParseResult, complexityMetrics *ComplexityMetrics) FunctionTestability {
	testability := FunctionTestability{
		Name:          function.Name,
		FilePath:      parseResult.FilePath,
		StartLine:     function.StartLine,
		EndLine:       function.EndLine,
		RequiredMocks: []string{},
		UntestedPaths: []string{},
		CoverageGaps:  []string{},
		Metadata:      make(map[string]interface{}),
	}

	// Calculate complexity factor
	testability.ComplexityFactor = ca.calculateComplexityFactor(function, parseResult, complexityMetrics)

	// Calculate coupling factor based on dependencies
	testability.CouplingFactor = ca.calculateCouplingFactor(function, parseResult)

	// Calculate dependency factor
	testability.DependencyFactor = ca.calculateDependencyFactor(function, parseResult)

	// Calculate overall testability score
	testability.TestabilityScore = ca.calculateTestabilityScore(testability)

	// Estimate coverage potential
	testability.EstimatedCoverage = ca.estimateCoveragePotential(testability)

	// Determine testing difficulty
	testability.TestingDifficulty = ca.determineTestingDifficulty(testability)

	// Identify required mocks
	testability.RequiredMocks = ca.identifyRequiredMocks(function, parseResult)

	// Identify untested paths
	testability.UntestedPaths = ca.identifyFunctionUntestedPaths(function)

	// Assess risk level
	testability.RiskLevel = ca.assessFunctionRiskLevel(testability)

	// Estimate testing effort
	testability.TestingEffort = ca.estimateTestingEffort(testability)

	// Recommend testing approach
	testability.RecommendedApproach = ca.recommendTestingApproach(testability)

	// Identify coverage gaps
	testability.CoverageGaps = ca.identifyFunctionCoverageGaps(function, parseResult)

	// Add metadata
	testability.Metadata["line_count"] = function.EndLine - function.StartLine + 1
	testability.Metadata["parameter_count"] = len(function.Parameters)
	testability.Metadata["is_async"] = function.IsAsync
	testability.Metadata["is_exported"] = function.IsExported

	return testability
}

// calculateComplexityFactor determines the complexity impact on testability
func (ca *CoverageAnalyzer) calculateComplexityFactor(function ast.FunctionInfo, parseResult *ast.ParseResult, complexityMetrics *ComplexityMetrics) float64 {
	if complexityMetrics == nil {
		return 50.0 // Default moderate complexity
	}

	// Find the function in complexity metrics
	for _, funcMetric := range complexityMetrics.FunctionMetrics {
		if funcMetric.Name == function.Name && funcMetric.FilePath == parseResult.FilePath {
			// Convert weighted score to testability impact (inverse relationship)
			if funcMetric.WeightedScore <= float64(ca.config.LowComplexityThreshold) {
				return 20.0 // Low impact on testability
			} else if funcMetric.WeightedScore <= float64(ca.config.HighComplexityThreshold) {
				return 50.0 // Moderate impact
			} else {
				return 80.0 // High impact on testability
			}
		}
	}

	// Fallback: estimate based on function size and parameters
	lineCount := function.EndLine - function.StartLine
	paramCount := len(function.Parameters)

	complexityScore := float64(lineCount)*0.5 + float64(paramCount)*2.0

	if complexityScore <= 10 {
		return 20.0
	} else if complexityScore <= 30 {
		return 50.0
	} else {
		return 80.0
	}
}

// calculateCouplingFactor assesses coupling impact on testability
func (ca *CoverageAnalyzer) calculateCouplingFactor(function ast.FunctionInfo, parseResult *ast.ParseResult) float64 {
	// Count external dependencies and imports usage
	externalCalls := 0
	internalDependencies := 0

	// Analyze imports for external dependencies
	for _, imp := range parseResult.Imports {
		if ca.isExternalDependency(imp.Source) {
			externalCalls++
		} else {
			internalDependencies++
		}
	}

	// Calculate coupling based on dependencies
	totalCoupling := externalCalls*2 + internalDependencies

	if totalCoupling <= ca.config.LowCouplingThreshold {
		return 20.0 // Low coupling
	} else if totalCoupling <= ca.config.HighCouplingThreshold {
		return 50.0 // Moderate coupling
	} else {
		return 80.0 // High coupling
	}
}

// calculateDependencyFactor assesses external dependency impact
func (ca *CoverageAnalyzer) calculateDependencyFactor(function ast.FunctionInfo, parseResult *ast.ParseResult) float64 {
	// Count different types of dependencies
	databaseDeps := 0
	networkDeps := 0
	fileDeps := 0
	externalAPIDeps := 0

	for _, imp := range parseResult.Imports {
		source := strings.ToLower(imp.Source)
		switch {
		case strings.Contains(source, "database") || strings.Contains(source, "sql") || strings.Contains(source, "mongodb"):
			databaseDeps++
		case strings.Contains(source, "http") || strings.Contains(source, "fetch") || strings.Contains(source, "axios"):
			networkDeps++
		case strings.Contains(source, "fs") || strings.Contains(source, "file") || strings.Contains(source, "path"):
			fileDeps++
		case strings.Contains(source, "api") || strings.Contains(source, "client"):
			externalAPIDeps++
		}
	}

	// Calculate dependency impact
	dependencyScore := float64(databaseDeps*3 + networkDeps*2 + fileDeps*1 + externalAPIDeps*2)

	if dependencyScore <= 2 {
		return 20.0 // Low dependency impact
	} else if dependencyScore <= 6 {
		return 50.0 // Moderate dependency impact
	} else {
		return 80.0 // High dependency impact
	}
}

// calculateTestabilityScore computes the overall testability score
func (ca *CoverageAnalyzer) calculateTestabilityScore(testability FunctionTestability) float64 {
	// Inverse scoring - lower factors mean higher testability
	complexityImpact := (100.0 - testability.ComplexityFactor) * ca.config.ComplexityWeight
	couplingImpact := (100.0 - testability.CouplingFactor) * ca.config.CouplingWeight
	dependencyImpact := (100.0 - testability.DependencyFactor) * ca.config.DependencyWeight

	// Size factor based on line count
	lineCount := testability.EndLine - testability.StartLine + 1
	sizeFactor := math.Min(100.0, float64(lineCount)*2.0) // Larger functions are harder to test
	sizeImpact := (100.0 - sizeFactor) * ca.config.SizeWeight

	// Pattern factor (async functions, exported functions are easier to test)
	patternFactor := 50.0 // Default
	if isAsync, ok := testability.Metadata["is_async"].(bool); ok && isAsync {
		patternFactor += 10.0 // Async functions can be complex to test
	}
	if isExported, ok := testability.Metadata["is_exported"].(bool); ok && isExported {
		patternFactor -= 20.0 // Exported functions are easier to test
	}
	patternImpact := (100.0 - patternFactor) * ca.config.PatternWeight

	score := complexityImpact + couplingImpact + dependencyImpact + sizeImpact + patternImpact

	// Ensure score is within bounds
	return math.Max(0.0, math.Min(100.0, score))
}

// estimateCoveragePotential estimates how much coverage this function can achieve
func (ca *CoverageAnalyzer) estimateCoveragePotential(testability FunctionTestability) float64 {
	baseScore := testability.TestabilityScore

	// Adjust based on complexity patterns
	if len(testability.UntestedPaths) > 0 {
		pathPenalty := float64(len(testability.UntestedPaths)) * 5.0
		baseScore -= pathPenalty
	}

	// Mock requirements reduce coverage potential
	if len(testability.RequiredMocks) > 0 {
		mockPenalty := float64(len(testability.RequiredMocks)) * 8.0
		baseScore -= mockPenalty
	}

	return math.Max(0.0, math.Min(100.0, baseScore))
}

// determineTestingDifficulty classifies the testing difficulty level
func (ca *CoverageAnalyzer) determineTestingDifficulty(testability FunctionTestability) string {
	score := testability.TestabilityScore

	if score >= 80.0 {
		return "easy"
	} else if score >= 60.0 {
		return "moderate"
	} else if score >= 40.0 {
		return "difficult"
	} else {
		return "very_difficult"
	}
}

// identifyRequiredMocks identifies mocking requirements for the function
func (ca *CoverageAnalyzer) identifyRequiredMocks(function ast.FunctionInfo, parseResult *ast.ParseResult) []string {
	mocks := []string{}

	for _, imp := range parseResult.Imports {
		source := strings.ToLower(imp.Source)
		switch {
		case strings.Contains(source, "database") || strings.Contains(source, "sql"):
			mocks = append(mocks, "database_connection")
		case strings.Contains(source, "http") || strings.Contains(source, "fetch"):
			mocks = append(mocks, "http_client")
		case strings.Contains(source, "fs") || strings.Contains(source, "file"):
			mocks = append(mocks, "filesystem")
		case strings.Contains(source, "redis") || strings.Contains(source, "cache"):
			mocks = append(mocks, "cache_client")
		case strings.Contains(source, "email") || strings.Contains(source, "smtp"):
			mocks = append(mocks, "email_service")
		case strings.Contains(source, "auth") || strings.Contains(source, "jwt"):
			mocks = append(mocks, "auth_service")
		}
	}

	return mocks
}

// identifyFunctionUntestedPaths identifies potentially untested paths in the function
func (ca *CoverageAnalyzer) identifyFunctionUntestedPaths(function ast.FunctionInfo) []string {
	paths := []string{}

	// Estimate paths based on function characteristics
	if function.IsAsync {
		paths = append(paths, "async_error_handling", "promise_rejection")
	}

	// Estimate based on parameter count (more parameters = more edge cases)
	if len(function.Parameters) > 3 {
		paths = append(paths, "parameter_validation", "edge_case_parameters")
	}

	// Check for optional parameters
	for _, param := range function.Parameters {
		if param.IsOptional {
			paths = append(paths, fmt.Sprintf("optional_param_%s", param.Name))
		}
	}

	return paths
}

// assessFunctionRiskLevel determines the risk level for testing this function
func (ca *CoverageAnalyzer) assessFunctionRiskLevel(testability FunctionTestability) string {
	score := testability.TestabilityScore

	if score < ca.config.MediumRiskThreshold {
		return "critical"
	} else if score < ca.config.HighRiskThreshold {
		return "high"
	} else if score < ca.config.CriticalRiskThreshold {
		return "medium"
	} else {
		return "low"
	}
}

// estimateTestingEffort calculates estimated effort in hours
func (ca *CoverageAnalyzer) estimateTestingEffort(testability FunctionTestability) int {
	baseEffort := 2 // Base 2 hours per function

	// Adjust based on testability score
	if testability.TestabilityScore < 40 {
		baseEffort += 6 // Very difficult functions
	} else if testability.TestabilityScore < 60 {
		baseEffort += 4 // Difficult functions
	} else if testability.TestabilityScore < 80 {
		baseEffort += 2 // Moderate functions
	}

	// Add effort for mocks
	baseEffort += len(testability.RequiredMocks) * 1

	// Add effort for untested paths
	baseEffort += len(testability.UntestedPaths) * 1

	// Add effort for coverage gaps
	baseEffort += len(testability.CoverageGaps) * 1

	return baseEffort
}

// recommendTestingApproach suggests the best testing approach
func (ca *CoverageAnalyzer) recommendTestingApproach(testability FunctionTestability) string {
	if len(testability.RequiredMocks) > 3 {
		return "integration_testing_with_test_doubles"
	} else if testability.TestabilityScore < 40 {
		return "refactor_then_unit_test"
	} else if len(testability.RequiredMocks) > 0 {
		return "unit_testing_with_mocks"
	} else {
		return "pure_unit_testing"
	}
}

// identifyFunctionCoverageGaps identifies specific coverage gaps
func (ca *CoverageAnalyzer) identifyFunctionCoverageGaps(function ast.FunctionInfo, parseResult *ast.ParseResult) []string {
	gaps := []string{}

	// Error handling gaps
	if function.IsAsync {
		gaps = append(gaps, "async_error_scenarios")
	}

	// Parameter validation gaps
	if len(function.Parameters) > 0 {
		gaps = append(gaps, "parameter_validation")
		gaps = append(gaps, "null_parameter_handling")
	}

	// Return type gaps
	if function.ReturnType != "" && function.ReturnType != "void" {
		gaps = append(gaps, "return_value_validation")
	}

	return gaps
}

// isExternalDependency checks if an import is an external dependency
func (ca *CoverageAnalyzer) isExternalDependency(source string) bool {
	// Check for common external dependency patterns
	external := []string{
		"http", "https", "fetch", "axios", "request",
		"database", "db", "sql", "mongodb", "postgres", "mysql",
		"redis", "cache", "memcached",
		"email", "smtp", "mail",
		"auth", "jwt", "oauth",
		"fs", "file", "path",
		"crypto", "bcrypt", "hash",
	}

	lowerSource := strings.ToLower(source)
	for _, pattern := range external {
		if strings.Contains(lowerSource, pattern) {
			return true
		}
	}

	// Check if it's a relative import
	return !strings.HasPrefix(source, ".") && !strings.HasPrefix(source, "/")
}

// calculateFileTestingComplexity calculates testing complexity at file level
func (ca *CoverageAnalyzer) calculateFileTestingComplexity(parseResult *ast.ParseResult) float64 {
	totalComplexity := 0.0

	// Factor in number of functions
	functionCount := len(parseResult.Functions)
	totalComplexity += float64(functionCount) * 10.0

	// Factor in number of classes
	classCount := len(parseResult.Classes)
	totalComplexity += float64(classCount) * 15.0

	// Factor in number of imports (external dependencies)
	externalImports := 0
	for _, imp := range parseResult.Imports {
		if ca.isExternalDependency(imp.Source) {
			externalImports++
		}
	}
	totalComplexity += float64(externalImports) * 5.0

	return math.Min(100.0, totalComplexity)
}

// determineFilePriority determines testing priority for a file
func (ca *CoverageAnalyzer) determineFilePriority(fileTestability FileTestability) string {
	score := fileTestability.OverallScore
	mockComplexity := fileTestability.MockComplexity
	untestedFunctions := fileTestability.UntestedFunctions

	// High priority if low testability score or many untested functions
	if score < 40.0 || untestedFunctions > 3 || mockComplexity > 5 {
		return "critical"
	} else if score < 60.0 || untestedFunctions > 1 || mockComplexity > 2 {
		return "high"
	} else if score < 80.0 || untestedFunctions > 0 {
		return "medium"
	} else {
		return "low"
	}
}

// countCoverageGaps counts potential coverage gaps in a file
func (ca *CoverageAnalyzer) countCoverageGaps(parseResult *ast.ParseResult) int {
	gaps := 0

	// Count functions with complex patterns
	for _, function := range parseResult.Functions {
		if function.IsAsync {
			gaps++ // Async functions have error handling gaps
		}
		if len(function.Parameters) > 3 {
			gaps++ // Complex parameter handling
		}
		if len(function.Parameters) > 0 {
			for _, param := range function.Parameters {
				if param.IsOptional {
					gaps++ // Optional parameter handling
				}
			}
		}
	}

	// Count classes with complex inheritance
	for _, class := range parseResult.Classes {
		if class.Extends != "" {
			gaps++ // Inheritance testing complexity
		}
		if len(class.Implements) > 0 {
			gaps += len(class.Implements) // Interface implementation gaps
		}
	}

	return gaps
}

// generateFileTestingRecommendations generates file-level testing recommendations
func (ca *CoverageAnalyzer) generateFileTestingRecommendations(fileTestability FileTestability) []string {
	recommendations := []string{}

	if fileTestability.OverallScore < 40.0 {
		recommendations = append(recommendations, "Consider refactoring before adding tests")
		recommendations = append(recommendations, "Focus on reducing complexity and dependencies")
	}

	if fileTestability.MockComplexity > 5 {
		recommendations = append(recommendations, "High mocking complexity - consider dependency injection")
		recommendations = append(recommendations, "Evaluate integration testing approach")
	}

	if fileTestability.UntestedFunctions > 2 {
		recommendations = append(recommendations, "Prioritize testing high-risk functions first")
		recommendations = append(recommendations, "Consider test-driven development for new features")
	}

	if fileTestability.CoverageGapCount > 5 {
		recommendations = append(recommendations, "Multiple coverage gaps identified")
		recommendations = append(recommendations, "Implement comprehensive error handling tests")
	}

	switch fileTestability.RecommendedPriority {
	case "critical":
		recommendations = append(recommendations, "Critical priority - immediate testing required")
	case "high":
		recommendations = append(recommendations, "High priority - schedule testing within current sprint")
	case "medium":
		recommendations = append(recommendations, "Medium priority - plan testing for next release")
	case "low":
		recommendations = append(recommendations, "Low priority - maintain existing test coverage")
	}

	return recommendations
}

// identifyUntestedPaths identifies untested code paths through AST analysis
func (ca *CoverageAnalyzer) identifyUntestedPaths(parseResults []*ast.ParseResult, metrics *CoverageMetrics) {
	pathID := 1000

	for _, parseResult := range parseResults {
		for _, function := range parseResult.Functions {
			// Analyze conditional paths
			ca.analyzeConditionalPaths(function, parseResult, metrics, &pathID)

			// Analyze loop paths
			ca.analyzeLoopPaths(function, parseResult, metrics, &pathID)

			// Analyze exception paths
			ca.analyzeExceptionPaths(function, parseResult, metrics, &pathID)

			// Analyze async paths
			ca.analyzeAsyncPaths(function, parseResult, metrics, &pathID)
		}
	}
}

// analyzeConditionalPaths identifies untested conditional branches
func (ca *CoverageAnalyzer) analyzeConditionalPaths(function ast.FunctionInfo, parseResult *ast.ParseResult, metrics *CoverageMetrics, pathID *int) {
	// Estimate conditional paths based on function complexity
	lineCount := function.EndLine - function.StartLine
	paramCount := len(function.Parameters)

	// More parameters and larger functions likely have more conditionals
	estimatedConditionals := paramCount + lineCount/10

	for i := 0; i < estimatedConditionals; i++ {
		path := UntestedPath{
			ID:              fmt.Sprintf("conditional_%d", *pathID),
			FilePath:        parseResult.FilePath,
			FunctionName:    function.Name,
			PathType:        "conditional",
			StartLine:       function.StartLine + i*2,
			EndLine:         function.StartLine + i*2 + 1,
			Condition:       fmt.Sprintf("conditional_branch_%d", i+1),
			RiskLevel:       ca.assessPathRiskLevel("conditional", function),
			TestingStrategy: "branch_coverage_testing",
			RequiredSetup:   []string{"setup_test_data", "mock_dependencies"},
			ExpectedOutcome: "verify_branch_execution",
		}
		metrics.UntestedPaths = append(metrics.UntestedPaths, path)
		*pathID++
	}
}

// analyzeLoopPaths identifies untested loop scenarios
func (ca *CoverageAnalyzer) analyzeLoopPaths(function ast.FunctionInfo, parseResult *ast.ParseResult, metrics *CoverageMetrics, pathID *int) {
	// Estimate loop paths based on function characteristics
	if function.EndLine-function.StartLine > 20 { // Larger functions likely have loops
		path := UntestedPath{
			ID:              fmt.Sprintf("loop_%d", *pathID),
			FilePath:        parseResult.FilePath,
			FunctionName:    function.Name,
			PathType:        "loop",
			StartLine:       function.StartLine + 5,
			EndLine:         function.EndLine - 5,
			Condition:       "loop_iterations",
			RiskLevel:       ca.assessPathRiskLevel("loop", function),
			TestingStrategy: "boundary_value_testing",
			RequiredSetup:   []string{"setup_loop_data", "test_empty_collection", "test_single_item", "test_multiple_items"},
			ExpectedOutcome: "verify_loop_behavior",
		}
		metrics.UntestedPaths = append(metrics.UntestedPaths, path)
		*pathID++
	}
}

// analyzeExceptionPaths identifies untested exception handling
func (ca *CoverageAnalyzer) analyzeExceptionPaths(function ast.FunctionInfo, parseResult *ast.ParseResult, metrics *CoverageMetrics, pathID *int) {
	// Functions with parameters or external dependencies likely need exception handling
	if len(function.Parameters) > 0 || len(parseResult.Imports) > 2 {
		path := UntestedPath{
			ID:              fmt.Sprintf("exception_%d", *pathID),
			FilePath:        parseResult.FilePath,
			FunctionName:    function.Name,
			PathType:        "exception",
			StartLine:       function.StartLine,
			EndLine:         function.EndLine,
			Condition:       "error_scenarios",
			RiskLevel:       ca.assessPathRiskLevel("exception", function),
			TestingStrategy: "error_injection_testing",
			RequiredSetup:   []string{"mock_error_conditions", "setup_invalid_inputs"},
			ExpectedOutcome: "verify_error_handling",
		}
		metrics.UntestedPaths = append(metrics.UntestedPaths, path)
		*pathID++
	}
}

// analyzeAsyncPaths identifies untested async execution paths
func (ca *CoverageAnalyzer) analyzeAsyncPaths(function ast.FunctionInfo, parseResult *ast.ParseResult, metrics *CoverageMetrics, pathID *int) {
	if function.IsAsync {
		// Success path
		successPath := UntestedPath{
			ID:              fmt.Sprintf("async_success_%d", *pathID),
			FilePath:        parseResult.FilePath,
			FunctionName:    function.Name,
			PathType:        "async",
			StartLine:       function.StartLine,
			EndLine:         function.EndLine,
			Condition:       "async_success_scenario",
			RiskLevel:       ca.assessPathRiskLevel("async", function),
			TestingStrategy: "async_testing_with_awaits",
			RequiredSetup:   []string{"setup_async_mocks", "configure_promises"},
			ExpectedOutcome: "verify_successful_resolution",
		}
		metrics.UntestedPaths = append(metrics.UntestedPaths, successPath)
		*pathID++

		// Error path
		errorPath := UntestedPath{
			ID:              fmt.Sprintf("async_error_%d", *pathID),
			FilePath:        parseResult.FilePath,
			FunctionName:    function.Name,
			PathType:        "async",
			StartLine:       function.StartLine,
			EndLine:         function.EndLine,
			Condition:       "async_error_scenario",
			RiskLevel:       "high",
			TestingStrategy: "async_error_testing",
			RequiredSetup:   []string{"setup_promise_rejection", "mock_async_errors"},
			ExpectedOutcome: "verify_error_propagation",
		}
		metrics.UntestedPaths = append(metrics.UntestedPaths, errorPath)
		*pathID++
	}
}

// assessPathRiskLevel determines risk level for a specific path type
func (ca *CoverageAnalyzer) assessPathRiskLevel(pathType string, function ast.FunctionInfo) string {
	switch pathType {
	case "conditional":
		if len(function.Parameters) > 3 {
			return "high"
		}
		return "medium"
	case "loop":
		return "high" // Loops are often high risk
	case "exception":
		return "critical" // Exception handling is critical
	case "async":
		if len(function.Parameters) > 0 {
			return "high"
		}
		return "medium"
	default:
		return "medium"
	}
}

// analyzeMockRequirements analyzes mock requirements for external dependencies
func (ca *CoverageAnalyzer) analyzeMockRequirements(parseResults []*ast.ParseResult, metrics *CoverageMetrics) {
	mockID := 2000

	for _, parseResult := range parseResults {
		ca.analyzeFileMockRequirements(parseResult, metrics, &mockID)
	}
}

// analyzeFileMockRequirements analyzes mock requirements for a single file
func (ca *CoverageAnalyzer) analyzeFileMockRequirements(parseResult *ast.ParseResult, metrics *CoverageMetrics, mockID *int) {
	// Analyze imports for mock requirements
	for _, imp := range parseResult.Imports {
		mockReq := ca.createMockRequirement(imp, parseResult, mockID)
		if mockReq != nil {
			metrics.MockRequirements = append(metrics.MockRequirements, *mockReq)
		}
	}

	// Analyze functions for specific mock needs
	for _, function := range parseResult.Functions {
		ca.analyzeFunctionMockRequirements(function, parseResult, metrics, mockID)
	}
}

// createMockRequirement creates a mock requirement based on import analysis
func (ca *CoverageAnalyzer) createMockRequirement(imp ast.ImportInfo, parseResult *ast.ParseResult, mockID *int) *MockRequirement {
	source := strings.ToLower(imp.Source)

	var depType, depName, mockStrategy, mockLibrary string
	var setupComplexity int
	var maintenanceOverhead string

	switch {
	case strings.Contains(source, "database") || strings.Contains(source, "sql") || strings.Contains(source, "mongodb"):
		depType = "database"
		depName = imp.Source
		mockStrategy = "in_memory_database_or_repository_pattern"
		mockLibrary = "test_database_library"
		setupComplexity = 8
		maintenanceOverhead = "high"

	case strings.Contains(source, "http") || strings.Contains(source, "fetch") || strings.Contains(source, "axios"):
		depType = "network"
		depName = imp.Source
		mockStrategy = "http_mock_server_or_stub"
		mockLibrary = "nock_or_msw"
		setupComplexity = 6
		maintenanceOverhead = "medium"

	case strings.Contains(source, "fs") || strings.Contains(source, "file") || strings.Contains(source, "path"):
		depType = "filesystem"
		depName = imp.Source
		mockStrategy = "virtual_filesystem_or_temp_files"
		mockLibrary = "mock_fs"
		setupComplexity = 4
		maintenanceOverhead = "low"

	case strings.Contains(source, "redis") || strings.Contains(source, "cache"):
		depType = "cache"
		depName = imp.Source
		mockStrategy = "in_memory_cache_mock"
		mockLibrary = "redis_memory_server"
		setupComplexity = 5
		maintenanceOverhead = "medium"

	default:
		if ca.isExternalDependency(imp.Source) {
			depType = "external_api"
			depName = imp.Source
			mockStrategy = "api_mock_or_stub"
			mockLibrary = "generic_mock_library"
			setupComplexity = 5
			maintenanceOverhead = "medium"
		} else {
			return nil // Internal dependency, no mock needed
		}
	}

	mockComplexity := "moderate"
	if setupComplexity >= 8 {
		mockComplexity = "very_complex"
	} else if setupComplexity >= 6 {
		mockComplexity = "complex"
	} else if setupComplexity <= 3 {
		mockComplexity = "simple"
	}

	*mockID++
	return &MockRequirement{
		ID:                  fmt.Sprintf("mock_%d", *mockID),
		FilePath:            parseResult.FilePath,
		FunctionName:        "file_level",
		DependencyType:      depType,
		DependencyName:      depName,
		MockingComplexity:   mockComplexity,
		MockingStrategy:     mockStrategy,
		RequiredMockLibrary: mockLibrary,
		SetupComplexity:     setupComplexity,
		MaintenanceOverhead: maintenanceOverhead,
		Alternatives:        ca.generateMockAlternatives(depType),
		Metadata: map[string]interface{}{
			"import_source": imp.Source,
			"import_type":   imp.ImportType,
		},
	}
}

// analyzeFunctionMockRequirements analyzes specific function mock requirements
func (ca *CoverageAnalyzer) analyzeFunctionMockRequirements(function ast.FunctionInfo, parseResult *ast.ParseResult, metrics *CoverageMetrics, mockID *int) {
	// Functions with many parameters may need parameter mocking
	if len(function.Parameters) > 4 {
		*mockID++
		mockReq := MockRequirement{
			ID:                  fmt.Sprintf("param_mock_%d", *mockID),
			FilePath:            parseResult.FilePath,
			FunctionName:        function.Name,
			DependencyType:      "parameters",
			DependencyName:      "function_parameters",
			MockingComplexity:   "moderate",
			MockingStrategy:     "parameter_object_mocking",
			RequiredMockLibrary: "jest_or_sinon",
			SetupComplexity:     3,
			MaintenanceOverhead: "low",
			Alternatives:        []string{"builder_pattern", "factory_functions", "default_parameters"},
			Metadata: map[string]interface{}{
				"parameter_count": len(function.Parameters),
				"function_name":   function.Name,
			},
		}
		metrics.MockRequirements = append(metrics.MockRequirements, mockReq)
	}
}

// generateMockAlternatives generates alternative mocking strategies
func (ca *CoverageAnalyzer) generateMockAlternatives(depType string) []string {
	switch depType {
	case "database":
		return []string{"in_memory_database", "docker_test_container", "database_transaction_rollback"}
	case "network":
		return []string{"http_mock_server", "contract_testing", "recorded_responses"}
	case "filesystem":
		return []string{"virtual_filesystem", "temporary_directories", "dependency_injection"}
	case "cache":
		return []string{"map_based_cache", "mock_implementation", "test_cache_instance"}
	case "external_api":
		return []string{"wiremock", "contract_testing", "recorded_interactions"}
	default:
		return []string{"stub", "spy", "fake_implementation"}
	}
}

// Final methods for CoverageAnalyzer - these complete the implementation

// generateTestingRecommendations generates testing recommendations based on risk and complexity
func (ca *CoverageAnalyzer) generateTestingRecommendations(metrics *CoverageMetrics) {
	recID := 3000

	// Generate recommendations for high-risk functions
	for _, funcTestability := range metrics.FunctionAnalysis {
		if funcTestability.RiskLevel == "critical" || funcTestability.RiskLevel == "high" {
			recommendation := TestingRecommendation{
				ID:              fmt.Sprintf("rec_%d", recID),
				Priority:        funcTestability.RiskLevel,
				Category:        ca.recommendTestCategory(funcTestability),
				FilePath:        funcTestability.FilePath,
				FunctionName:    funcTestability.Name,
				Recommendation:  ca.generateFunctionRecommendation(funcTestability),
				Rationale:       ca.generateRecommendationRationale(funcTestability),
				EstimatedEffort: funcTestability.TestingEffort,
				ExpectedBenefit: ca.calculateExpectedBenefit(funcTestability),
				TestingApproach: ca.generateTestingApproach(funcTestability),
				RequiredTools:   ca.identifyRequiredTools(funcTestability),
				RiskReduction:   ca.calculateRiskReduction(funcTestability),
				Metadata: map[string]interface{}{
					"testability_score": funcTestability.TestabilityScore,
					"complexity_factor": funcTestability.ComplexityFactor,
					"mock_count":        len(funcTestability.RequiredMocks),
				},
			}
			metrics.TestingRecommendations = append(metrics.TestingRecommendations, recommendation)
			recID++
		}
	}

	// Generate coverage gap recommendations
	ca.generateCoverageGapRecommendations(metrics, &recID)
}

// recommendTestCategory recommends the appropriate test category
func (ca *CoverageAnalyzer) recommendTestCategory(funcTestability FunctionTestability) string {
	mockCount := len(funcTestability.RequiredMocks)

	if mockCount == 0 {
		return "unit"
	} else if mockCount <= 2 {
		return "unit"
	} else if mockCount <= 4 {
		return "integration"
	} else {
		return "e2e"
	}
}

// generateFunctionRecommendation generates specific testing recommendation
func (ca *CoverageAnalyzer) generateFunctionRecommendation(funcTestability FunctionTestability) string {
	if funcTestability.TestabilityScore < 40 {
		return fmt.Sprintf("Refactor function '%s' to reduce complexity before testing", funcTestability.Name)
	} else if len(funcTestability.RequiredMocks) > 3 {
		return fmt.Sprintf("Implement integration tests for '%s' with comprehensive mocking", funcTestability.Name)
	} else if funcTestability.RiskLevel == "critical" {
		return fmt.Sprintf("Prioritize comprehensive testing for critical function '%s'", funcTestability.Name)
	} else {
		return fmt.Sprintf("Add unit tests for '%s' with %s approach", funcTestability.Name, funcTestability.RecommendedApproach)
	}
}

// generateRecommendationRationale provides the reasoning behind recommendations
func (ca *CoverageAnalyzer) generateRecommendationRationale(funcTestability FunctionTestability) string {
	reasons := []string{}

	if funcTestability.TestabilityScore < 60 {
		reasons = append(reasons, "low testability score")
	}
	if funcTestability.ComplexityFactor > 60 {
		reasons = append(reasons, "high complexity")
	}
	if len(funcTestability.RequiredMocks) > 2 {
		reasons = append(reasons, "multiple external dependencies")
	}
	if len(funcTestability.UntestedPaths) > 0 {
		reasons = append(reasons, "untested execution paths")
	}

	if len(reasons) == 0 {
		return "Standard testing practices recommended"
	}

	return fmt.Sprintf("Recommended due to: %v", reasons)
}

// calculateExpectedBenefit calculates the expected benefit of testing
func (ca *CoverageAnalyzer) calculateExpectedBenefit(funcTestability FunctionTestability) string {
	score := funcTestability.TestabilityScore
	riskLevel := funcTestability.RiskLevel

	if riskLevel == "critical" && score < 40 {
		return "Very High - Critical function with testing challenges"
	} else if riskLevel == "high" || score < 50 {
		return "High - Significant risk reduction expected"
	} else if riskLevel == "medium" || score < 70 {
		return "Medium - Moderate improvement in reliability"
	} else {
		return "Low - Incremental quality improvement"
	}
}

// generateTestingApproach provides detailed testing approach
func (ca *CoverageAnalyzer) generateTestingApproach(funcTestability FunctionTestability) []string {
	approaches := []string{}

	// Base approach
	approaches = append(approaches, funcTestability.RecommendedApproach)

	// Add specific patterns based on function characteristics
	if isAsync, ok := funcTestability.Metadata["is_async"].(bool); ok && isAsync {
		approaches = append(approaches, "async_await_testing")
	}

	if len(funcTestability.RequiredMocks) > 0 {
		approaches = append(approaches, "dependency_mocking")
	}

	if len(funcTestability.UntestedPaths) > 0 {
		approaches = append(approaches, "branch_coverage_testing")
		approaches = append(approaches, "edge_case_testing")
	}

	if funcTestability.ComplexityFactor > 70 {
		approaches = append(approaches, "test_driven_refactoring")
	}

	return approaches
}

// identifyRequiredTools identifies testing tools needed
func (ca *CoverageAnalyzer) identifyRequiredTools(funcTestability FunctionTestability) []string {
	tools := []string{"jest", "testing_framework"}

	// Add tools based on mock requirements
	for _, mock := range funcTestability.RequiredMocks {
		switch mock {
		case "database_connection":
			tools = append(tools, "database_testing_library")
		case "http_client":
			tools = append(tools, "nock", "msw")
		case "filesystem":
			tools = append(tools, "mock_fs")
		case "cache_client":
			tools = append(tools, "redis_memory_server")
		}
	}

	// Add tools based on function characteristics
	if isAsync, ok := funcTestability.Metadata["is_async"].(bool); ok && isAsync {
		tools = append(tools, "async_testing_utilities")
	}

	if funcTestability.ComplexityFactor > 70 {
		tools = append(tools, "coverage_analysis_tool")
	}

	return tools
}

// calculateRiskReduction calculates expected risk reduction percentage
func (ca *CoverageAnalyzer) calculateRiskReduction(funcTestability FunctionTestability) float64 {
	baseReduction := 60.0 // Base 60% risk reduction from testing

	// Adjust based on testability score
	if funcTestability.TestabilityScore > 80 {
		baseReduction += 20.0 // Easy to test functions get more benefit
	} else if funcTestability.TestabilityScore < 40 {
		baseReduction -= 20.0 // Hard to test functions get less benefit initially
	}

	// Adjust based on coverage potential
	coverageFactor := funcTestability.EstimatedCoverage / 100.0
	baseReduction *= coverageFactor

	return math.Min(90.0, math.Max(20.0, baseReduction))
}

// generateCoverageGapRecommendations generates recommendations for coverage gaps
func (ca *CoverageAnalyzer) generateCoverageGapRecommendations(metrics *CoverageMetrics, recID *int) {
	gapID := 4000

	// Analyze untested paths to generate gap recommendations
	pathTypes := make(map[string][]UntestedPath)
	for _, path := range metrics.UntestedPaths {
		pathTypes[path.PathType] = append(pathTypes[path.PathType], path)
	}

	// Generate recommendations for each path type
	for pathType, paths := range pathTypes {
		if len(paths) > 2 { // Only recommend if multiple instances
			gap := CoverageGap{
				ID:              fmt.Sprintf("gap_%d", gapID),
				Type:            pathType,
				FilePath:        "multiple_files",
				Location:        fmt.Sprintf("%d_untested_%s_paths", len(paths), pathType),
				Severity:        ca.assessGapSeverity(pathType, len(paths)),
				Impact:          ca.calculateGapImpact(pathType, len(paths)),
				RiskAssessment:  ca.assessGapRisk(pathType, len(paths)),
				TestingStrategy: ca.recommendGapStrategy(pathType),
				EstimatedEffort: len(paths) * 2,
				Prerequisites:   ca.getGapPrerequisites(pathType),
			}
			metrics.CoverageGaps = append(metrics.CoverageGaps, gap)
			gapID++

			// Generate corresponding recommendation
			recommendation := TestingRecommendation{
				ID:              fmt.Sprintf("rec_%d", *recID),
				Priority:        gap.Severity,
				Category:        "coverage",
				FilePath:        "multiple_files",
				FunctionName:    "multiple_functions",
				Recommendation:  fmt.Sprintf("Address %s coverage gaps across %d functions", pathType, len(paths)),
				Rationale:       fmt.Sprintf("Identified %d untested %s paths", len(paths), pathType),
				EstimatedEffort: gap.EstimatedEffort,
				ExpectedBenefit: "Improved code coverage and reliability",
				TestingApproach: []string{gap.TestingStrategy},
				RequiredTools:   ca.getStrategyTools(pathType),
				RiskReduction:   float64(len(paths)) * 5.0, // 5% per path
				Metadata: map[string]interface{}{
					"gap_type":  pathType,
					"gap_count": len(paths),
					"gap_files": ca.extractUniqueFiles(paths),
				},
			}
			metrics.TestingRecommendations = append(metrics.TestingRecommendations, recommendation)
			*recID++
		}
	}
}

// assessGapSeverity determines severity of coverage gaps
func (ca *CoverageAnalyzer) assessGapSeverity(pathType string, count int) string {
	switch pathType {
	case "exception":
		if count > 5 {
			return "critical"
		} else if count > 2 {
			return "high"
		}
		return "medium"
	case "async":
		if count > 3 {
			return "high"
		}
		return "medium"
	case "conditional":
		if count > 10 {
			return "high"
		} else if count > 5 {
			return "medium"
		}
		return "low"
	case "loop":
		if count > 5 {
			return "high"
		}
		return "medium"
	default:
		return "low"
	}
}

// calculateGapImpact calculates the impact of coverage gaps
func (ca *CoverageAnalyzer) calculateGapImpact(pathType string, count int) string {
	impact := "Low impact on overall code quality"

	switch pathType {
	case "exception":
		impact = "High impact - unhandled errors can cause system failures"
	case "async":
		impact = "Medium-High impact - async errors can be difficult to debug"
	case "loop":
		impact = "Medium impact - loop edge cases can cause performance issues"
	case "conditional":
		impact = "Medium impact - untested branches can hide logical errors"
	}

	if count > 5 {
		impact = "High " + impact
	}

	return impact
}

// assessGapRisk assesses the risk level of coverage gaps
func (ca *CoverageAnalyzer) assessGapRisk(pathType string, count int) string {
	risk := fmt.Sprintf("%d untested %s paths represent ", count, pathType)

	switch pathType {
	case "exception":
		risk += "critical reliability risk"
	case "async":
		risk += "high operational risk"
	case "loop":
		risk += "medium performance risk"
	case "conditional":
		risk += "medium logical risk"
	default:
		risk += "low to medium risk"
	}

	return risk
}

// recommendGapStrategy recommends strategy for addressing gaps
func (ca *CoverageAnalyzer) recommendGapStrategy(pathType string) string {
	switch pathType {
	case "exception":
		return "comprehensive_error_testing"
	case "async":
		return "async_pattern_testing"
	case "loop":
		return "boundary_value_analysis"
	case "conditional":
		return "branch_coverage_analysis"
	default:
		return "systematic_testing_approach"
	}
}

// getGapPrerequisites gets prerequisites for addressing gaps
func (ca *CoverageAnalyzer) getGapPrerequisites(pathType string) []string {
	switch pathType {
	case "exception":
		return []string{"error_mocking_library", "exception_testing_framework"}
	case "async":
		return []string{"async_testing_utilities", "promise_testing_library"}
	case "loop":
		return []string{"data_generation_library", "boundary_testing_tools"}
	case "conditional":
		return []string{"code_coverage_tool", "branch_analysis_tool"}
	default:
		return []string{"testing_framework", "coverage_analysis_tool"}
	}
}

// getStrategyTools gets tools needed for testing strategies
func (ca *CoverageAnalyzer) getStrategyTools(pathType string) []string {
	switch pathType {
	case "exception":
		return []string{"jest", "error_testing_library", "mock_error_conditions"}
	case "async":
		return []string{"jest", "async_testing_utilities", "promise_testing"}
	case "loop":
		return []string{"jest", "property_based_testing", "data_generators"}
	case "conditional":
		return []string{"jest", "coverage_reporting", "branch_coverage_tool"}
	default:
		return []string{"jest", "testing_utilities"}
	}
}

// extractUniqueFiles extracts unique file paths from untested paths
func (ca *CoverageAnalyzer) extractUniqueFiles(paths []UntestedPath) []string {
	fileSet := make(map[string]bool)
	var files []string

	for _, path := range paths {
		if !fileSet[path.FilePath] {
			fileSet[path.FilePath] = true
			files = append(files, path.FilePath)
		}
	}

	return files
}

// createTestingPriorityMatrix creates prioritized testing recommendations
func (ca *CoverageAnalyzer) createTestingPriorityMatrix(metrics *CoverageMetrics) {
	var testingItems []TestingItem

	// Create testing items from function analysis
	for _, funcTestability := range metrics.FunctionAnalysis {
		item := TestingItem{
			FilePath:       funcTestability.FilePath,
			FunctionName:   funcTestability.Name,
			TestingType:    ca.recommendTestCategory(funcTestability),
			RiskScore:      100.0 - funcTestability.TestabilityScore, // Inverse of testability
			EffortEstimate: funcTestability.TestingEffort,
			ImpactScore:    ca.calculateImpactScore(funcTestability),
			ROI:            0, // Will be calculated below
		}

		// Calculate ROI (Impact / Effort)
		if item.EffortEstimate > 0 {
			item.ROI = item.ImpactScore / float64(item.EffortEstimate)
		}

		testingItems = append(testingItems, item)
	}

	// Sort by ROI descending
	sort.Slice(testingItems, func(i, j int) bool {
		return testingItems[i].ROI > testingItems[j].ROI
	})

	// Distribute into priority buckets
	metrics.PriorityMatrix = TestingPriorityMatrix{
		CriticalPriority: []TestingItem{},
		HighPriority:     []TestingItem{},
		MediumPriority:   []TestingItem{},
		LowPriority:      []TestingItem{},
	}

	for _, item := range testingItems {
		switch {
		case item.RiskScore >= 80 || item.ROI >= 8.0:
			metrics.PriorityMatrix.CriticalPriority = append(metrics.PriorityMatrix.CriticalPriority, item)
		case item.RiskScore >= 60 || item.ROI >= 4.0:
			metrics.PriorityMatrix.HighPriority = append(metrics.PriorityMatrix.HighPriority, item)
		case item.RiskScore >= 40 || item.ROI >= 2.0:
			metrics.PriorityMatrix.MediumPriority = append(metrics.PriorityMatrix.MediumPriority, item)
		default:
			metrics.PriorityMatrix.LowPriority = append(metrics.PriorityMatrix.LowPriority, item)
		}
	}
}

// calculateImpactScore calculates the impact score for a function
func (ca *CoverageAnalyzer) calculateImpactScore(funcTestability FunctionTestability) float64 {
	baseImpact := 50.0 // Base impact score

	// Higher complexity = higher impact if tested
	baseImpact += funcTestability.ComplexityFactor * 0.3

	// More dependencies = higher impact if tested well
	baseImpact += funcTestability.DependencyFactor * 0.2

	// Critical functions have higher impact
	switch funcTestability.RiskLevel {
	case "critical":
		baseImpact += 30.0
	case "high":
		baseImpact += 20.0
	case "medium":
		baseImpact += 10.0
	}

	// Exported functions have higher impact (public API)
	if isExported, ok := funcTestability.Metadata["is_exported"].(bool); ok && isExported {
		baseImpact += 15.0
	}

	return math.Min(100.0, baseImpact)
}

// generateTestingStrategy generates overall testing strategy
func (ca *CoverageAnalyzer) generateTestingStrategy(metrics *CoverageMetrics) {
	totalFunctions := len(metrics.FunctionAnalysis)
	criticalFunctions := len(metrics.PriorityMatrix.CriticalPriority)
	highPriorityFunctions := len(metrics.PriorityMatrix.HighPriority)

	// Determine overall approach
	var overallApproach string
	if criticalFunctions > totalFunctions/4 {
		overallApproach = "risk_driven_testing_with_immediate_critical_focus"
	} else if highPriorityFunctions > totalFunctions/3 {
		overallApproach = "balanced_testing_with_priority_focus"
	} else {
		overallApproach = "comprehensive_coverage_driven_testing"
	}

	// Recommend testing framework
	recommendedFramework := "jest_with_comprehensive_mocking"

	// Calculate testing pyramid distribution
	pyramid := TestingPyramid{
		UnitTestPercentage:        70.0,
		IntegrationTestPercentage: 20.0,
		E2ETestPercentage:         8.0,
		PerformanceTestPercentage: 2.0,
	}

	// Adjust pyramid based on mock complexity
	totalMocks := len(metrics.MockRequirements)
	if totalMocks > totalFunctions {
		// High mock complexity - increase integration testing
		pyramid.UnitTestPercentage = 60.0
		pyramid.IntegrationTestPercentage = 30.0
		pyramid.E2ETestPercentage = 8.0
		pyramid.PerformanceTestPercentage = 2.0
	}

	// Generate priority order
	priorityOrder := []string{}
	if len(metrics.PriorityMatrix.CriticalPriority) > 0 {
		priorityOrder = append(priorityOrder, "critical_functions")
	}
	if len(metrics.PriorityMatrix.HighPriority) > 0 {
		priorityOrder = append(priorityOrder, "high_priority_functions")
	}
	priorityOrder = append(priorityOrder, "medium_priority_functions", "low_priority_functions")

	// Generate phase recommendations
	phases := []PhaseRecommendation{
		{
			Phase:          "immediate",
			Description:    "Address critical testing gaps and high-risk functions",
			Objectives:     []string{"test_critical_functions", "setup_testing_infrastructure", "address_security_gaps"},
			Deliverables:   []string{"critical_function_tests", "testing_framework_setup", "initial_coverage_report"},
			EstimatedWeeks: ca.calculatePhaseWeeks(len(metrics.PriorityMatrix.CriticalPriority)),
			RequiredSkills: []string{"unit_testing", "mocking", "test_framework_setup"},
		},
		{
			Phase:          "short_term",
			Description:    "Expand testing coverage to high and medium priority functions",
			Objectives:     []string{"increase_coverage", "establish_testing_patterns", "automate_testing"},
			Deliverables:   []string{"comprehensive_test_suite", "ci_cd_integration", "coverage_reporting"},
			EstimatedWeeks: ca.calculatePhaseWeeks(len(metrics.PriorityMatrix.HighPriority) + len(metrics.PriorityMatrix.MediumPriority)),
			RequiredSkills: []string{"integration_testing", "test_automation", "coverage_analysis"},
		},
		{
			Phase:          "long_term",
			Description:    "Achieve comprehensive coverage and establish testing culture",
			Objectives:     []string{"complete_coverage", "performance_testing", "maintenance_optimization"},
			Deliverables:   []string{"full_test_coverage", "performance_test_suite", "testing_documentation"},
			EstimatedWeeks: ca.calculatePhaseWeeks(len(metrics.PriorityMatrix.LowPriority)),
			RequiredSkills: []string{"performance_testing", "test_maintenance", "testing_strategy"},
		},
	}

	// Calculate resource requirements
	resources := ResourceRequirements{
		DeveloperHours:      ca.calculateTotalDeveloperHours(metrics),
		QAHours:             ca.calculateTotalDeveloperHours(metrics) / 2, // 50% of dev hours
		InfrastructureNeed:  []string{"ci_cd_pipeline", "test_databases", "coverage_reporting_tools"},
		ToolingRequirements: []string{"testing_framework", "mocking_libraries", "coverage_analysis_tools"},
		SkillGaps:           []string{"advanced_mocking", "integration_testing", "performance_testing"},
	}

	// Calculate timeline
	timeline := TimelineEstimate{
		ImmediatePhaseWeeks: phases[0].EstimatedWeeks,
		ShortTermPhaseWeeks: phases[1].EstimatedWeeks,
		LongTermPhaseWeeks:  phases[2].EstimatedWeeks,
		TotalWeeks:          phases[0].EstimatedWeeks + phases[1].EstimatedWeeks + phases[2].EstimatedWeeks,
	}

	metrics.TestingStrategy = TestingStrategy{
		OverallApproach:      overallApproach,
		RecommendedFramework: recommendedFramework,
		TestingPyramid:       pyramid,
		PriorityOrder:        priorityOrder,
		PhaseRecommendations: phases,
		ResourceRequirements: resources,
		TimelineEstimate:     timeline,
	}
}

// calculatePhaseWeeks calculates estimated weeks for a testing phase
func (ca *CoverageAnalyzer) calculatePhaseWeeks(functionCount int) int {
	if functionCount == 0 {
		return 1
	}
	// Estimate 1 week per 5 functions, minimum 2 weeks
	weeks := int(math.Ceil(float64(functionCount) / 5.0))
	return int(math.Max(2, float64(weeks)))
}

// calculateTotalDeveloperHours calculates total developer hours needed
func (ca *CoverageAnalyzer) calculateTotalDeveloperHours(metrics *CoverageMetrics) int {
	totalHours := 0

	for _, funcTestability := range metrics.FunctionAnalysis {
		totalHours += funcTestability.TestingEffort
	}

	// Add overhead for setup, infrastructure, and maintenance
	overhead := int(float64(totalHours) * 0.3) // 30% overhead

	return totalHours + overhead
}

// calculateOverallMetrics calculates final metrics and summary
func (ca *CoverageAnalyzer) calculateOverallMetrics(metrics *CoverageMetrics) {
	if len(metrics.FunctionAnalysis) == 0 {
		metrics.Summary = CoverageSummary{
			QualityGate: "pass",
		}
		return
	}

	totalFunctions := len(metrics.FunctionAnalysis)
	totalScore := 0.0
	totalCoverage := 0.0
	testedFunctions := 0
	highRiskFunctions := 0

	// Calculate aggregate metrics
	for _, funcTestability := range metrics.FunctionAnalysis {
		totalScore += funcTestability.TestabilityScore
		totalCoverage += funcTestability.EstimatedCoverage

		if funcTestability.TestabilityScore >= 60.0 {
			testedFunctions++
		}

		if funcTestability.RiskLevel == "high" || funcTestability.RiskLevel == "critical" {
			highRiskFunctions++
		}
	}

	// Calculate overall metrics
	metrics.OverallScore = totalScore / float64(totalFunctions)
	metrics.EstimatedCoverage = totalCoverage / float64(totalFunctions)
	metrics.TestabilityScore = metrics.OverallScore

	// Determine mocking complexity
	mockingComplexity := "low"
	totalMocks := len(metrics.MockRequirements)
	if totalMocks > totalFunctions*2 {
		mockingComplexity = "very_high"
	} else if totalMocks > totalFunctions {
		mockingComplexity = "high"
	} else if totalMocks > totalFunctions/2 {
		mockingComplexity = "medium"
	}

	// Estimate testing weeks
	totalEffort := ca.calculateTotalDeveloperHours(metrics)
	estimatedWeeks := int(math.Ceil(float64(totalEffort) / 40.0)) // 40 hours per week

	// Determine recommended focus
	recommendedFocus := "balanced_coverage"
	if highRiskFunctions > totalFunctions/3 {
		recommendedFocus = "high_risk_functions"
	} else if len(metrics.MockRequirements) > totalFunctions {
		recommendedFocus = "integration_testing"
	} else if metrics.OverallScore < 60 {
		recommendedFocus = "refactoring_for_testability"
	}

	// Determine quality gate
	qualityGate := "pass"
	if metrics.OverallScore < 40 || highRiskFunctions > totalFunctions/2 {
		qualityGate = "fail"
	} else if metrics.OverallScore < 60 || highRiskFunctions > totalFunctions/4 {
		qualityGate = "warning"
	}

	// Create summary
	metrics.Summary = CoverageSummary{
		TotalFunctions:        totalFunctions,
		TestedFunctions:       testedFunctions,
		UntestedFunctions:     totalFunctions - testedFunctions,
		CoveragePercentage:    metrics.EstimatedCoverage,
		TestabilityScore:      metrics.TestabilityScore,
		HighRiskFunctions:     highRiskFunctions,
		MockingComplexity:     mockingComplexity,
		EstimatedTestingWeeks: estimatedWeeks,
		RecommendedFocus:      recommendedFocus,
		QualityGate:           qualityGate,
	}
}
