package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewMaintainabilityCalculator(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	assert.NotNil(t, calculator)
	assert.Equal(t, 85.0, calculator.config.GoodThreshold)
	assert.Equal(t, 70.0, calculator.config.FairThreshold)
	assert.Equal(t, 0.23, calculator.config.HalsteadWeight)
	assert.Equal(t, 0.12, calculator.config.ComplexityWeight)
	assert.Equal(t, 0.40, calculator.config.LOCWeight)
	assert.Equal(t, 0.25, calculator.config.CommentWeight)
	assert.Equal(t, 3, calculator.config.MinFunctionLines)
}

func TestNewMaintainabilityCalculatorWithConfig(t *testing.T) {
	config := MaintainabilityConfig{
		GoodThreshold:    90.0,
		FairThreshold:    75.0,
		HalsteadWeight:   0.30,
		ComplexityWeight: 0.15,
		LOCWeight:        0.35,
		CommentWeight:    0.20,
		EnableTrends:     false,
		TrendPeriods:     3,
		ReportTopN:       5,
		MinFunctionLines: 5,
	}

	calculator := NewMaintainabilityCalculatorWithConfig(config)

	assert.NotNil(t, calculator)
	assert.Equal(t, config.GoodThreshold, calculator.config.GoodThreshold)
	assert.Equal(t, config.FairThreshold, calculator.config.FairThreshold)
	assert.Equal(t, config.HalsteadWeight, calculator.config.HalsteadWeight)
	assert.Equal(t, config.MinFunctionLines, calculator.config.MinFunctionLines)
}

func TestAnalyzeMaintainability_EmptyInput(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	metrics, err := calculator.AnalyzeMaintainability(context.Background(), []*ast.ParseResult{}, nil)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 100.0, metrics.OverallIndex) // Perfect score for no code
	assert.Equal(t, "Good", metrics.Classification)
	assert.Equal(t, 0, len(metrics.FunctionMetrics))
	assert.Equal(t, 0, len(metrics.FileMetrics))
	assert.Equal(t, "No code to analyze", metrics.Summary.TopRecommendation)
}

func TestAnalyzeMaintainability_ValidInput(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	parseResults := createMockParseResultsForMaintainability()
	complexityMetrics := createMockComplexityMetricsForMaintainability()

	metrics, err := calculator.AnalyzeMaintainability(context.Background(), parseResults, complexityMetrics)

	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify basic metrics structure
	assert.GreaterOrEqual(t, metrics.OverallIndex, 0.0)
	assert.LessOrEqual(t, metrics.OverallIndex, 100.0)
	assert.NotEmpty(t, metrics.Classification)
	assert.Greater(t, metrics.TotalFunctions, 0)
	assert.Greater(t, metrics.TotalFiles, 0)

	// Verify component breakdown
	assert.GreaterOrEqual(t, metrics.ComponentBreakdown.HalsteadVolume, 0.0)
	assert.GreaterOrEqual(t, metrics.ComponentBreakdown.CyclomaticComplexity, 0.0)
	assert.Greater(t, metrics.ComponentBreakdown.LinesOfCode, 0)
	assert.GreaterOrEqual(t, metrics.ComponentBreakdown.CommentRatio, 0.0)

	// Verify breakdown by level
	totalBreakdown := metrics.IndexByLevel.Good.Count + metrics.IndexByLevel.Fair.Count + metrics.IndexByLevel.Poor.Count
	assert.Equal(t, metrics.TotalFunctions, totalBreakdown)

	// Verify recommendations are generated
	assert.NotNil(t, metrics.ImprovementSuggestions)
	assert.NotNil(t, metrics.Summary)
	assert.NotNil(t, metrics.BenchmarkComparison)
}

func TestCalculateHalsteadMetrics(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	function := ast.FunctionInfo{
		Name:      "testFunction",
		StartLine: 1,
		EndLine:   20,
		Parameters: []ast.ParameterInfo{
			{Name: "param1", Type: "string"},
			{Name: "param2", Type: "number"},
		},
		IsAsync: true,
	}

	result := &ast.ParseResult{
		FilePath: "test.js",
		Language: "javascript",
	}

	halsteadMetrics := calculator.calculateHalsteadMetrics(function, result)

	assert.Greater(t, halsteadMetrics.UniqueOperators, 0)
	assert.Greater(t, halsteadMetrics.UniqueOperands, 0)
	assert.Greater(t, halsteadMetrics.TotalOperators, 0)
	assert.Greater(t, halsteadMetrics.TotalOperands, 0)
	assert.Greater(t, halsteadMetrics.Vocabulary, 0)
	assert.Greater(t, halsteadMetrics.Length, 0)
	assert.Greater(t, halsteadMetrics.Volume, 0.0)
	assert.Greater(t, halsteadMetrics.Difficulty, 0.0)
	assert.Greater(t, halsteadMetrics.Effort, 0.0)
	assert.GreaterOrEqual(t, halsteadMetrics.TimeToUnderstand, 0.0)
	assert.GreaterOrEqual(t, halsteadMetrics.BugsDelivered, 0.0)
}

func TestGetCyclomaticComplexity(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	function := ast.FunctionInfo{
		Name:      "testFunction",
		StartLine: 10,
		EndLine:   25,
		Parameters: []ast.ParameterInfo{
			{Name: "param1", Type: "string"},
		},
	}

	// Test with existing complexity metrics
	complexityMetrics := &ComplexityMetrics{
		FunctionMetrics: []FunctionComplexity{
			{
				Name:            "testFunction",
				StartLine:       10,
				CyclomaticValue: 5,
			},
		},
	}

	complexity := calculator.getCyclomaticComplexity(function, complexityMetrics)
	assert.Equal(t, 5, complexity)

	// Test with nil complexity metrics (fallback)
	complexityFallback := calculator.getCyclomaticComplexity(function, nil)
	assert.GreaterOrEqual(t, complexityFallback, 1)
	assert.LessOrEqual(t, complexityFallback, 10)
}

func TestCalculateCommentRatio(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	result := &ast.ParseResult{
		FilePath: "test.js",
		Language: "javascript",
	}

	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected float64 // approximate expected range
	}{
		{
			name: "small_function",
			function: ast.FunctionInfo{
				Name:       "smallFunc",
				StartLine:  1,
				EndLine:    5,
				Parameters: []ast.ParameterInfo{{Name: "param", Type: "string"}},
				IsExported: false,
			},
			expected: 0.05, // Around 5%
		},
		{
			name: "large_exported_function",
			function: ast.FunctionInfo{
				Name:      "largeFunc",
				StartLine: 1,
				EndLine:   60,
				Parameters: []ast.ParameterInfo{
					{Name: "param1", Type: "string"},
					{Name: "param2", Type: "object"},
					{Name: "param3", Type: "array"},
					{Name: "param4", Type: "boolean"},
					{Name: "param5", Type: "number"},
					{Name: "param6", Type: "function"},
				},
				IsExported: true,
			},
			expected: 0.20, // Around 20%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := calculator.calculateCommentRatio(tt.function, result)
			assert.GreaterOrEqual(t, ratio, 0.0)
			assert.LessOrEqual(t, ratio, 0.25)
			assert.InDelta(t, tt.expected, ratio, 0.10) // Within 10% of expected
		})
	}
}

func TestCalculateMaintainabilityIndex(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		name                 string
		halsteadVolume       float64
		cyclomaticComplexity float64
		linesOfCode          float64
		commentRatio         float64
		expectedRange        [2]float64 // [min, max]
	}{
		{
			name:                 "simple_function",
			halsteadVolume:       100.0,
			cyclomaticComplexity: 2.0,
			linesOfCode:          10.0,
			commentRatio:         0.15,
			expectedRange:        [2]float64{80.0, 100.0}, // Actually gets high score due to small values
		},
		{
			name:                 "complex_function",
			halsteadVolume:       2000.0,
			cyclomaticComplexity: 15.0,
			linesOfCode:          100.0,
			commentRatio:         0.05,
			expectedRange:        [2]float64{50.0, 70.0}, // Fair maintainability
		},
		{
			name:                 "well_documented_function",
			halsteadVolume:       500.0,
			cyclomaticComplexity: 5.0,
			linesOfCode:          30.0,
			commentRatio:         0.25,
			expectedRange:        [2]float64{85.0, 100.0}, // Good maintainability due to good documentation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index := calculator.calculateMaintainabilityIndex(
				tt.halsteadVolume,
				tt.cyclomaticComplexity,
				tt.linesOfCode,
				tt.commentRatio,
			)

			assert.GreaterOrEqual(t, index, 0.0)
			assert.LessOrEqual(t, index, 100.0)
			assert.GreaterOrEqual(t, index, tt.expectedRange[0])
			assert.LessOrEqual(t, index, tt.expectedRange[1])
		})
	}
}

func TestClassifyMaintainability(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		index    float64
		expected string
	}{
		{95.0, "Good"},
		{85.0, "Good"},
		{80.0, "Fair"},
		{70.0, "Fair"},
		{65.0, "Poor"},
		{30.0, "Poor"},
		{0.0, "Poor"},
		{100.0, "Good"},
	}

	for _, tt := range tests {
		classification := calculator.classifyMaintainability(tt.index)
		assert.Equal(t, tt.expected, classification, "Index %.1f should be classified as %s", tt.index, tt.expected)
	}
}

func TestShouldAnalyzeFunction(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected bool
	}{
		{
			name: "valid_function",
			function: ast.FunctionInfo{
				Name:      "validFunction",
				StartLine: 1,
				EndLine:   10,
			},
			expected: true,
		},
		{
			name: "too_small_function",
			function: ast.FunctionInfo{
				Name:      "smallFunction",
				StartLine: 1,
				EndLine:   2,
			},
			expected: false,
		},
		{
			name: "unnamed_function",
			function: ast.FunctionInfo{
				Name:      "",
				StartLine: 1,
				EndLine:   10,
			},
			expected: false,
		},
		{
			name: "exactly_min_size",
			function: ast.FunctionInfo{
				Name:      "minSizeFunction",
				StartLine: 1,
				EndLine:   3,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.shouldAnalyzeFunction(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMaintainabilityBreakdown(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	functions := []FunctionMaintainability{
		{Name: "goodFunc1", MaintainabilityIndex: 90.0, Classification: "Good"},
		{Name: "goodFunc2", MaintainabilityIndex: 88.0, Classification: "Good"},
		{Name: "fairFunc1", MaintainabilityIndex: 75.0, Classification: "Fair"},
		{Name: "fairFunc2", MaintainabilityIndex: 72.0, Classification: "Fair"},
		{Name: "poorFunc1", MaintainabilityIndex: 65.0, Classification: "Poor"},
	}

	breakdown := calculator.generateMaintainabilityBreakdown(functions)

	// Check counts
	assert.Equal(t, 2, breakdown.Good.Count)
	assert.Equal(t, 2, breakdown.Fair.Count)
	assert.Equal(t, 1, breakdown.Poor.Count)

	// Check percentages
	assert.InDelta(t, 40.0, breakdown.Good.Percentage, 0.1) // 2/5 = 40%
	assert.InDelta(t, 40.0, breakdown.Fair.Percentage, 0.1) // 2/5 = 40%
	assert.InDelta(t, 20.0, breakdown.Poor.Percentage, 0.1) // 1/5 = 20%

	// Check averages
	assert.InDelta(t, 89.0, breakdown.Good.AverageIndex, 0.1) // (90+88)/2 = 89
	assert.InDelta(t, 73.5, breakdown.Fair.AverageIndex, 0.1) // (75+72)/2 = 73.5
	assert.InDelta(t, 65.0, breakdown.Poor.AverageIndex, 0.1) // 65/1 = 65

	// Check function names
	assert.Contains(t, breakdown.Good.Functions, "goodFunc1")
	assert.Contains(t, breakdown.Good.Functions, "goodFunc2")
	assert.Contains(t, breakdown.Fair.Functions, "fairFunc1")
	assert.Contains(t, breakdown.Fair.Functions, "fairFunc2")
	assert.Contains(t, breakdown.Poor.Functions, "poorFunc1")
}

func TestIdentifyImprovementFactors(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		name            string
		components      MaintainabilityComponents
		index           float64
		expectedFactors []string
	}{
		{
			name: "high_complexity_function",
			components: MaintainabilityComponents{
				CyclomaticComplexity: 15.0,
				LinesOfCode:          30,
				CommentRatio:         0.12,
				HalsteadVolume:       800.0,
			},
			index:           65.0,
			expectedFactors: []string{"High cyclomatic complexity", "Multiple maintainability issues"},
		},
		{
			name: "large_undocumented_function",
			components: MaintainabilityComponents{
				CyclomaticComplexity: 5.0,
				LinesOfCode:          80,
				CommentRatio:         0.02,
				HalsteadVolume:       600.0,
			},
			index:           68.0,
			expectedFactors: []string{"Function too large", "Insufficient documentation", "Multiple maintainability issues"},
		},
		{
			name: "complex_algorithmic_function",
			components: MaintainabilityComponents{
				CyclomaticComplexity: 12.0,
				LinesOfCode:          40,
				CommentRatio:         0.08,
				HalsteadVolume:       1200.0,
			},
			index:           60.0,
			expectedFactors: []string{"High cyclomatic complexity", "Insufficient documentation", "Complex algorithmic logic", "Multiple maintainability issues"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factors := calculator.identifyImprovementFactors(tt.components, tt.index)

			for _, expectedFactor := range tt.expectedFactors {
				assert.Contains(t, factors, expectedFactor)
			}
		})
	}
}

func TestGenerateRecommendedActions(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		name            string
		components      MaintainabilityComponents
		function        ast.FunctionInfo
		expectedActions []string
	}{
		{
			name: "complex_large_function",
			components: MaintainabilityComponents{
				CyclomaticComplexity: 12.0,
				LinesOfCode:          60,
				CommentRatio:         0.05,
				HalsteadVolume:       800.0,
			},
			function: ast.FunctionInfo{
				Name: "complexFunc",
				Parameters: []ast.ParameterInfo{
					{Name: "param1"}, {Name: "param2"}, {Name: "param3"},
				},
				IsAsync: false,
			},
			expectedActions: []string{
				"Break down complex conditional logic",
				"Extract nested logic into separate functions",
				"Split into smaller, focused functions",
				"Extract reusable logic into utility functions",
				"Add function documentation and inline comments",
				"Document complex business logic",
			},
		},
		{
			name: "async_function_with_many_params",
			components: MaintainabilityComponents{
				CyclomaticComplexity: 8.0,
				LinesOfCode:          25,
				CommentRatio:         0.15,
				HalsteadVolume:       1100.0,
			},
			function: ast.FunctionInfo{
				Name: "asyncFunc",
				Parameters: []ast.ParameterInfo{
					{Name: "param1"}, {Name: "param2"}, {Name: "param3"},
					{Name: "param4"}, {Name: "param5"}, {Name: "param6"},
				},
				IsAsync: true,
			},
			expectedActions: []string{
				"Simplify algorithmic complexity",
				"Use more descriptive variable names",
				"Reduce parameter count using objects or configuration",
				"Simplify async/await patterns",
				"Consider breaking down async operations",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := calculator.generateRecommendedActions(tt.components, tt.function)

			for _, expectedAction := range tt.expectedActions {
				assert.Contains(t, actions, expectedAction)
			}
		})
	}
}

func TestGenerateImprovementSuggestions(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	functions := []FunctionMaintainability{
		{
			Name:                 "poorFunc",
			FilePath:             "test1.js",
			MaintainabilityIndex: 60.0,
			Classification:       "Poor",
			RecommendedActions:   []string{"Reduce complexity", "Add documentation"},
			Components: MaintainabilityComponents{
				LinesOfCode:  80,
				CommentRatio: 0.02,
			},
		},
		{
			Name:                 "largeFunc",
			FilePath:             "test2.js",
			MaintainabilityIndex: 75.0,
			Classification:       "Fair",
			Components: MaintainabilityComponents{
				LinesOfCode:  60,
				CommentRatio: 0.10,
			},
		},
	}

	fileMetrics := map[string]FileMaintainability{
		"test1.js": {
			FilePath:   "test1.js",
			Components: MaintainabilityComponents{CommentRatio: 0.03},
		},
		"test2.js": {
			FilePath:   "test2.js",
			Components: MaintainabilityComponents{CommentRatio: 0.02},
		},
	}

	improvements := calculator.generateImprovementSuggestions(functions, fileMetrics)

	assert.Greater(t, len(improvements), 0)

	// Check for critical improvement (poor function)
	foundCritical := false
	for _, improvement := range improvements {
		if improvement.Priority == "critical" {
			foundCritical = true
			assert.Contains(t, improvement.Description, "poorFunc")
			assert.Equal(t, "complexity_reduction", improvement.Type)
			assert.Greater(t, improvement.ImpactEstimate, 0.0)
			break
		}
	}
	assert.True(t, foundCritical, "Should find critical improvement for poor function")

	// Check for documentation improvement
	foundDocumentation := false
	for _, improvement := range improvements {
		if improvement.Type == "documentation" {
			foundDocumentation = true
			assert.Equal(t, "medium", improvement.Priority)
			assert.Greater(t, len(improvement.AffectedFiles), 0)
			break
		}
	}
	assert.True(t, foundDocumentation, "Should find documentation improvement")

	// Check for function decomposition
	foundDecomposition := false
	for _, improvement := range improvements {
		if improvement.Type == "function_decomposition" {
			foundDecomposition = true
			assert.Equal(t, "high", improvement.Priority)
			assert.Greater(t, len(improvement.AffectedFunctions), 0)
			break
		}
	}
	assert.True(t, foundDecomposition, "Should find function decomposition improvement")
}

func TestCreateMaintainabilitySummary(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	metrics := &MaintainabilityMetrics{
		ImprovementSuggestions: []MaintainabilityImprovement{
			{Priority: "critical", ImpactEstimate: 10.0},
			{Priority: "critical", ImpactEstimate: 8.0},
			{Priority: "high", ImpactEstimate: 6.0},
			{Priority: "medium", ImpactEstimate: 4.0},
		},
	}

	// Set a top recommendation
	metrics.ImprovementSuggestions[0].Description = "Fix critical function"

	summary := calculator.createMaintainabilitySummary(metrics)

	assert.Equal(t, 4, summary.TotalIssues)
	assert.Equal(t, 2, summary.CriticalIssues)
	assert.Equal(t, 1, summary.HighPriorityIssues)
	assert.InDelta(t, 28.0, summary.ImprovementPotential, 0.1) // 10+8+6+4 = 28
	assert.Equal(t, "Fix critical function", summary.TopRecommendation)
	assert.Equal(t, 22, summary.EstimatedEffortHours)            // 2*8 + 1*4 + 1*2 = 22 (corrected calculation)
	assert.InDelta(t, 25.0, summary.PredictedIndexIncrease, 0.1) // Capped at 25
}

func TestGenerateBenchmarkComparison(t *testing.T) {
	calculator := NewMaintainabilityCalculator()

	tests := []struct {
		overallIndex    float64
		expectedRanking string
	}{
		{90.0, "top_10_percent"},
		{80.0, "above_average"},
		{70.0, "average"},
		{60.0, "below_average"},
	}

	for _, tt := range tests {
		benchmark := calculator.generateBenchmarkComparison(tt.overallIndex)

		assert.Equal(t, tt.expectedRanking, benchmark.ProjectRanking)
		assert.Greater(t, benchmark.IndustryAverage, 0.0)
		assert.Greater(t, len(benchmark.SimilarProjects), 0)
		assert.NotEmpty(t, benchmark.BenchmarkSource)

		expectedGap := tt.overallIndex - benchmark.IndustryAverage
		assert.InDelta(t, expectedGap, benchmark.CompetitiveGap, 0.1)
	}
}

// Helper functions for creating mock data

func createMockParseResultsForMaintainability() []*ast.ParseResult {
	return []*ast.ParseResult{
		{
			FilePath: "test1.js",
			Language: "javascript",
			Functions: []ast.FunctionInfo{
				{
					Name:      "simpleFunction",
					StartLine: 1,
					EndLine:   15,
					Parameters: []ast.ParameterInfo{
						{Name: "param1", Type: "string"},
					},
					IsAsync:    false,
					IsExported: true,
				},
				{
					Name:      "complexFunction",
					StartLine: 20,
					EndLine:   80,
					Parameters: []ast.ParameterInfo{
						{Name: "param1", Type: "object"},
						{Name: "param2", Type: "array"},
						{Name: "param3", Type: "function"},
						{Name: "param4", Type: "boolean"},
						{Name: "param5", Type: "number"},
						{Name: "param6", Type: "string"},
					},
					IsAsync:    true,
					IsExported: false,
				},
			},
		},
		{
			FilePath: "test2.ts",
			Language: "typescript",
			Functions: []ast.FunctionInfo{
				{
					Name:      "mediumFunction",
					StartLine: 1,
					EndLine:   35,
					Parameters: []ast.ParameterInfo{
						{Name: "data", Type: "object"},
						{Name: "options", Type: "object"},
					},
					IsAsync:    true,
					IsExported: true,
				},
			},
		},
	}
}

func createMockComplexityMetricsForMaintainability() *ComplexityMetrics {
	return &ComplexityMetrics{
		OverallScore:      75.0,
		AverageComplexity: 8.5,
		TotalFunctions:    3,
		FunctionMetrics: []FunctionComplexity{
			{
				Name:            "simpleFunction",
				FilePath:        "test1.js",
				StartLine:       1,
				EndLine:         15,
				CyclomaticValue: 3,
				WeightedScore:   5.0,
				SeverityLevel:   "low",
			},
			{
				Name:            "complexFunction",
				FilePath:        "test1.js",
				StartLine:       20,
				EndLine:         80,
				CyclomaticValue: 15,
				WeightedScore:   20.0,
				SeverityLevel:   "high",
			},
			{
				Name:            "mediumFunction",
				FilePath:        "test2.ts",
				StartLine:       1,
				EndLine:         35,
				CyclomaticValue: 7,
				WeightedScore:   10.0,
				SeverityLevel:   "medium",
			},
		},
		ClassMetrics: []ClassComplexity{},
	}
}
