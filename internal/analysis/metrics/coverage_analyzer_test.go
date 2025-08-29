package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewCoverageAnalyzer(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, 5, analyzer.config.LowComplexityThreshold)
	assert.Equal(t, 15, analyzer.config.HighComplexityThreshold)
	assert.Equal(t, 0.30, analyzer.config.ComplexityWeight)
	assert.Equal(t, 0.25, analyzer.config.CouplingWeight)
	assert.Equal(t, 0.20, analyzer.config.DependencyWeight)
	assert.Equal(t, 0.15, analyzer.config.SizeWeight)
	assert.Equal(t, 0.10, analyzer.config.PatternWeight)
}

func TestNewCoverageAnalyzerWithConfig(t *testing.T) {
	config := CoverageConfig{
		LowComplexityThreshold:  3,
		HighComplexityThreshold: 10,
		ComplexityWeight:        0.40,
		CouplingWeight:          0.30,
		DependencyWeight:        0.20,
		SizeWeight:              0.10,
		PatternWeight:           0.00,
	}

	analyzer := NewCoverageAnalyzerWithConfig(config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, 3, analyzer.config.LowComplexityThreshold)
	assert.Equal(t, 10, analyzer.config.HighComplexityThreshold)
	assert.Equal(t, 0.40, analyzer.config.ComplexityWeight)
}

func TestAnalyzeCoverage_EmptyInput(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	metrics, err := analyzer.AnalyzeCoverage(context.Background(), []*ast.ParseResult{}, nil)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, "pass", metrics.Summary.QualityGate)
	assert.Equal(t, 0, len(metrics.FunctionAnalysis))
	assert.Equal(t, 0, len(metrics.UntestedPaths))
	assert.Equal(t, 0, len(metrics.MockRequirements))
}

func TestAnalyzeCoverage_ValidInput(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	parseResults := createMockParseResults()
	complexityMetrics := createMockComplexityMetricsForCoverage()

	metrics, err := analyzer.AnalyzeCoverage(context.Background(), parseResults, complexityMetrics)

	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify function analysis
	assert.Greater(t, len(metrics.FunctionAnalysis), 0)

	// Verify file analysis
	assert.Greater(t, len(metrics.FileAnalysis), 0)

	// Verify overall metrics
	assert.GreaterOrEqual(t, metrics.OverallScore, 0.0)
	assert.LessOrEqual(t, metrics.OverallScore, 100.0)
	assert.GreaterOrEqual(t, metrics.TestabilityScore, 0.0)
	assert.LessOrEqual(t, metrics.TestabilityScore, 100.0)

	// Verify summary
	assert.NotEmpty(t, metrics.Summary.QualityGate)
	assert.GreaterOrEqual(t, metrics.Summary.TotalFunctions, 0)
	assert.GreaterOrEqual(t, metrics.Summary.EstimatedTestingWeeks, 1)
}

func TestCoverageAnalysisIntegration(t *testing.T) {
	analyzer := NewCoverageAnalyzer()

	parseResults := createComplexMockParseResults()
	complexityMetrics := createMockComplexityMetricsForCoverage()

	metrics, err := analyzer.AnalyzeCoverage(context.Background(), parseResults, complexityMetrics)

	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify comprehensive analysis
	assert.Greater(t, len(metrics.FunctionAnalysis), 0, "Should analyze functions")
	assert.Greater(t, len(metrics.UntestedPaths), 0, "Should identify untested paths")
	assert.Greater(t, len(metrics.TestingRecommendations), 0, "Should generate recommendations")

	// Verify priority matrix
	totalItems := len(metrics.PriorityMatrix.CriticalPriority) +
		len(metrics.PriorityMatrix.HighPriority) +
		len(metrics.PriorityMatrix.MediumPriority) +
		len(metrics.PriorityMatrix.LowPriority)
	assert.Equal(t, len(metrics.FunctionAnalysis), totalItems, "All functions should be prioritized")

	// Verify testing strategy
	assert.NotEmpty(t, metrics.TestingStrategy.OverallApproach)
	assert.NotEmpty(t, metrics.TestingStrategy.RecommendedFramework)
	assert.Greater(t, len(metrics.TestingStrategy.PhaseRecommendations), 0)

	// Verify realistic metrics
	assert.GreaterOrEqual(t, metrics.TestingStrategy.TimelineEstimate.TotalWeeks, 3, "Should have realistic timeline")
	assert.Greater(t, metrics.TestingStrategy.ResourceRequirements.DeveloperHours, 0)
}

func TestCoverageConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config CoverageConfig
	}{
		{
			name: "default_config",
			config: CoverageConfig{
				LowComplexityThreshold:  5,
				HighComplexityThreshold: 15,
				ComplexityWeight:        0.30,
				CouplingWeight:          0.25,
				DependencyWeight:        0.20,
				SizeWeight:              0.15,
				PatternWeight:           0.10,
			},
		},
		{
			name: "high_complexity_focus",
			config: CoverageConfig{
				LowComplexityThreshold:  3,
				HighComplexityThreshold: 8,
				ComplexityWeight:        0.50,
				CouplingWeight:          0.20,
				DependencyWeight:        0.15,
				SizeWeight:              0.10,
				PatternWeight:           0.05,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCoverageAnalyzerWithConfig(tt.config)

			// Verify weights sum to 1.0 (approximately)
			totalWeight := tt.config.ComplexityWeight + tt.config.CouplingWeight +
				tt.config.DependencyWeight + tt.config.SizeWeight + tt.config.PatternWeight
			assert.InDelta(t, 1.0, totalWeight, 0.01, "Weights should sum to approximately 1.0")

			// Test with sample data
			parseResults := createMockParseResults()
			complexityMetrics := createMockComplexityMetricsForCoverage()

			metrics, err := analyzer.AnalyzeCoverage(context.Background(), parseResults, complexityMetrics)
			require.NoError(t, err)
			assert.NotNil(t, metrics)
		})
	}
}

// Helper functions for creating mock data

func createMockParseResults() []*ast.ParseResult {
	return []*ast.ParseResult{
		{
			FilePath: "test.js",
			Language: "javascript",
			Functions: []ast.FunctionInfo{
				{
					Name:       "simpleFunction",
					Parameters: []ast.ParameterInfo{{Name: "param1", Type: "string"}},
					ReturnType: "string",
					IsAsync:    false,
					IsExported: true,
					StartLine:  1,
					EndLine:    5,
				},
				{
					Name: "complexFunction",
					Parameters: []ast.ParameterInfo{
						{Name: "param1", Type: "object"},
						{Name: "param2", Type: "string", IsOptional: true},
						{Name: "param3", Type: "number"},
					},
					ReturnType: "Promise<object>",
					IsAsync:    true,
					IsExported: true,
					StartLine:  10,
					EndLine:    30,
				},
			},
			Classes: []ast.ClassInfo{},
			Imports: []ast.ImportInfo{
				{Source: "fs", ImportType: "default"},
				{Source: "http", ImportType: "named"},
			},
		},
	}
}

func createComplexMockParseResults() []*ast.ParseResult {
	return []*ast.ParseResult{
		{
			FilePath: "complex.js",
			Language: "javascript",
			Functions: []ast.FunctionInfo{
				{
					Name: "criticalFunction",
					Parameters: []ast.ParameterInfo{
						{Name: "param1", Type: "object"},
						{Name: "param2", Type: "string"},
						{Name: "param3", Type: "number"},
						{Name: "param4", Type: "boolean", IsOptional: true},
						{Name: "param5", Type: "array", IsOptional: true},
					},
					ReturnType: "Promise<object>",
					IsAsync:    true,
					IsExported: true,
					StartLine:  1,
					EndLine:    50, // Large function
				},
			},
			Imports: []ast.ImportInfo{
				{Source: "database", ImportType: "default"},
				{Source: "http", ImportType: "named"},
				{Source: "redis", ImportType: "default"},
				{Source: "fs", ImportType: "named"},
			},
		},
		{
			FilePath: "simple.js",
			Language: "javascript",
			Functions: []ast.FunctionInfo{
				{
					Name:       "easyFunction",
					Parameters: []ast.ParameterInfo{},
					ReturnType: "string",
					IsAsync:    false,
					IsExported: true,
					StartLine:  1,
					EndLine:    3,
				},
			},
			Imports: []ast.ImportInfo{},
		},
	}
}

func createMockComplexityMetricsForCoverage() *ComplexityMetrics {
	return &ComplexityMetrics{
		OverallScore:      25.0,
		AverageComplexity: 12.5,
		TotalFunctions:    3,
		FunctionMetrics: []FunctionComplexity{
			{
				Name:            "simpleFunction",
				FilePath:        "test.js",
				StartLine:       1,
				EndLine:         5,
				WeightedScore:   8.0,
				CyclomaticValue: 3,
				CognitiveValue:  2,
				SeverityLevel:   "low",
			},
			{
				Name:            "complexFunction",
				FilePath:        "test.js",
				StartLine:       10,
				EndLine:         30,
				WeightedScore:   25.0,
				CyclomaticValue: 15,
				CognitiveValue:  20,
				SeverityLevel:   "high",
			},
			{
				Name:            "criticalFunction",
				FilePath:        "complex.js",
				StartLine:       1,
				EndLine:         50,
				WeightedScore:   45.0,
				CyclomaticValue: 25,
				CognitiveValue:  30,
				SeverityLevel:   "severe",
			},
		},
		ClassMetrics: []ClassComplexity{},
	}
}
