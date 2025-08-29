package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewPerformanceAnalyzer(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, 2, analyzer.config.NestedLoopThreshold)
	assert.Equal(t, 3, analyzer.config.QueryPatternThreshold)
	assert.Equal(t, 5, analyzer.config.DOMAccessThreshold)
	assert.Equal(t, 500, analyzer.config.BundleSizeThresholdKB)
	assert.Equal(t, 15, analyzer.config.ComponentComplexityMax)
	assert.Equal(t, 0.35, analyzer.config.AlgorithmicWeight)
	assert.Equal(t, 0.25, analyzer.config.MemoryWeight)
	assert.Equal(t, 0.20, analyzer.config.NetworkWeight)
	assert.Equal(t, 0.15, analyzer.config.RenderWeight)
	assert.Equal(t, 0.05, analyzer.config.BundleWeight)
}

func TestNewPerformanceAnalyzerWithConfig(t *testing.T) {
	config := PerformanceConfig{
		NestedLoopThreshold:    3,
		QueryPatternThreshold:  5,
		DOMAccessThreshold:     7,
		BundleSizeThresholdKB:  300,
		ComponentComplexityMax: 20,
		AlgorithmicWeight:      0.40,
		MemoryWeight:           0.30,
		NetworkWeight:          0.20,
		RenderWeight:           0.10,
		BundleWeight:           0.00,
	}

	analyzer := NewPerformanceAnalyzerWithConfig(config)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config.NestedLoopThreshold, analyzer.config.NestedLoopThreshold)
	assert.Equal(t, config.BundleSizeThresholdKB, analyzer.config.BundleSizeThresholdKB)
	assert.Equal(t, config.AlgorithmicWeight, analyzer.config.AlgorithmicWeight)
}

func TestAnalyzePerformance_EmptyInput(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()

	metrics, err := analyzer.AnalyzePerformance(context.Background(), []*ast.ParseResult{}, nil)

	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, "A", metrics.PerformanceGrade) // Perfect score for no issues
	assert.Equal(t, 100.0, metrics.OverallScore)
	assert.Equal(t, 0, len(metrics.AntiPatterns))
	assert.Equal(t, 0, len(metrics.Bottlenecks))
	assert.Equal(t, 0, len(metrics.OptimizationOpportunities))
}

func TestAnalyzePerformance_ValidInput(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()

	parseResults := createMockParseResultsForPerformance()
	complexityMetrics := createMockComplexityMetricsForPerformance()

	metrics, err := analyzer.AnalyzePerformance(context.Background(), parseResults, complexityMetrics)

	require.NoError(t, err)
	assert.NotNil(t, metrics)

	// Verify basic metrics structure
	assert.GreaterOrEqual(t, metrics.OverallScore, 0.0)
	assert.LessOrEqual(t, metrics.OverallScore, 100.0)
	assert.NotEmpty(t, metrics.PerformanceGrade)

	// Verify summary is populated
	assert.GreaterOrEqual(t, metrics.Summary.TotalAntiPatterns, 0)
	assert.GreaterOrEqual(t, metrics.Summary.CriticalIssues, 0)
	assert.GreaterOrEqual(t, metrics.Summary.HighPriorityIssues, 0)
	assert.GreaterOrEqual(t, metrics.Summary.OptimizationPotential, 0.0)
	assert.NotEmpty(t, metrics.Summary.TopRecommendation)

	// Verify recommendations are generated
	assert.Greater(t, len(metrics.Recommendations), 0)
}

func TestDetectNPlusOneQueriesAST(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{
		AntiPatterns: []AntiPattern{},
	}

	result := &ast.ParseResult{
		FilePath: "test.js",
		Functions: []ast.FunctionInfo{
			{
				Name:      "findUserById",
				StartLine: 10,
				EndLine:   15,
				Parameters: []ast.ParameterInfo{
					{Name: "id", Type: "string"},
				},
			},
		},
	}

	analyzer.detectNPlusOneQueriesAST(result, metrics)

	assert.Greater(t, len(metrics.AntiPatterns), 0)
	antiPattern := metrics.AntiPatterns[0]
	assert.Equal(t, "n_plus_one_query", antiPattern.Type)
	assert.Equal(t, "high", antiPattern.Severity)
	assert.Contains(t, antiPattern.Description, "findUserById")
	assert.Equal(t, "database", antiPattern.Impact.Category)
}

func TestDetectSynchronousLoopsAST(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{
		AntiPatterns: []AntiPattern{},
	}

	result := &ast.ParseResult{
		FilePath: "test.js",
		Functions: []ast.FunctionInfo{
			{
				Name:      "processItems",
				IsAsync:   true,
				StartLine: 20,
				EndLine:   30,
			},
		},
	}

	analyzer.detectSynchronousLoopsAST(result, metrics)

	assert.Greater(t, len(metrics.AntiPatterns), 0)
	antiPattern := metrics.AntiPatterns[0]
	assert.Equal(t, "sync_in_loop", antiPattern.Type)
	assert.Equal(t, "high", antiPattern.Severity)
	assert.Contains(t, antiPattern.Description, "processItems")
}

func TestDetectMemoryLeaksAST(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{
		AntiPatterns: []AntiPattern{},
	}

	result := &ast.ParseResult{
		FilePath: "test.js",
		Functions: []ast.FunctionInfo{
			{
				Name:      "cacheBuilder",
				StartLine: 1,
				EndLine:   60, // Large function
				Parameters: []ast.ParameterInfo{
					{Name: "data1", Type: "object"},
					{Name: "data2", Type: "array"},
					{Name: "data3", Type: "string"},
					{Name: "data4", Type: "number"},
					{Name: "data5", Type: "boolean"},
					{Name: "data6", Type: "object"}, // 6+ parameters
				},
			},
		},
		Imports: []ast.ImportInfo{
			{Source: "event-emitter", StartLine: 1},
		},
	}

	analyzer.detectMemoryLeaksAST(result, metrics)

	assert.Greater(t, len(metrics.AntiPatterns), 0)
	// Should detect both memory-intensive function and event listener risk
	assert.GreaterOrEqual(t, len(metrics.AntiPatterns), 2)

	// Check for memory leak patterns
	foundMemoryLeak := false
	foundEventRisk := false
	for _, antiPattern := range metrics.AntiPatterns {
		if antiPattern.Type == "potential_memory_leak" {
			foundMemoryLeak = true
		}
		if antiPattern.Type == "event_listener_risk" {
			foundEventRisk = true
		}
	}
	assert.True(t, foundMemoryLeak, "Should detect potential memory leak")
	assert.True(t, foundEventRisk, "Should detect event listener risk")
}

func TestDetectLargeFunctions(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{
		AntiPatterns: []AntiPattern{},
	}

	result := &ast.ParseResult{
		FilePath: "test.js",
		Functions: []ast.FunctionInfo{
			{
				Name:      "hugeFunction",
				StartLine: 1,
				EndLine:   150, // 150 lines - should trigger large function detection
			},
		},
	}

	analyzer.detectLargeFunctions(result, metrics)

	assert.Greater(t, len(metrics.AntiPatterns), 0)
	antiPattern := metrics.AntiPatterns[0]
	assert.Equal(t, "large_function", antiPattern.Type)
	assert.Equal(t, "high", antiPattern.Severity)
	assert.Contains(t, antiPattern.Description, "hugeFunction")
	assert.Contains(t, antiPattern.Description, "150 lines")
}

func TestAnalyzeBundleSize(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{}

	parseResults := []*ast.ParseResult{
		{
			FilePath: "app.js",
			Imports: []ast.ImportInfo{
				{Source: "lodash", ImportType: "default"},
				{Source: "moment", ImportType: "default"},
				{Source: "react", ImportType: "named", Specifiers: []string{"useState", "useEffect"}},
			},
		},
	}

	analyzer.analyzeBundleSize(parseResults, metrics)

	require.NotNil(t, metrics.BundleAnalysis)
	assert.Greater(t, metrics.BundleAnalysis.EstimatedSizeKB, 0)
	assert.Greater(t, len(metrics.BundleAnalysis.HeavyDependencies), 0)
	assert.Greater(t, len(metrics.BundleAnalysis.OptimizationTips), 0)

	// Check that heavy libraries are detected
	foundLodash := false
	foundMoment := false
	for _, dep := range metrics.BundleAnalysis.HeavyDependencies {
		if dep.Name == "lodash" {
			foundLodash = true
		}
		if dep.Name == "moment" {
			foundMoment = true
		}
	}
	assert.True(t, foundLodash, "Should detect lodash as heavy dependency")
	assert.True(t, foundMoment, "Should detect moment as heavy dependency")
}

func TestAnalyzeReactPerformance(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{}

	parseResults := []*ast.ParseResult{
		{
			FilePath: "Component.jsx",
			Imports: []ast.ImportInfo{
				{Source: "react", ImportType: "named"},
			},
			Functions: []ast.FunctionInfo{
				{
					Name:      "LargeComponent",
					StartLine: 1,
					EndLine:   100, // Large component
					Parameters: []ast.ParameterInfo{
						{Name: "prop1", Type: "string"},
						{Name: "prop2", Type: "object"},
						{Name: "prop3", Type: "function"},
						{Name: "prop4", Type: "array"},
						{Name: "prop5", Type: "boolean"},
						{Name: "prop6", Type: "number"},
						{Name: "prop7", Type: "string"},
						{Name: "prop8", Type: "object"}, // 8+ props
					},
				},
			},
			Classes: []ast.ClassInfo{
				{
					Name:      "ClassComponent",
					Extends:   "React.Component",
					StartLine: 200,
					EndLine:   350,                          // Large class component
					Methods:   make([]ast.FunctionInfo, 20), // Many methods
				},
			},
		},
	}

	analyzer.analyzeReactPerformance(parseResults, metrics)

	require.NotNil(t, metrics.ReactAnalysis)
	assert.Greater(t, len(metrics.ReactAnalysis.ComponentIssues), 0)
	assert.Greater(t, len(metrics.ReactAnalysis.RenderOptimizations), 0)

	// Check for specific issues
	foundLargeComponent := false
	foundTooManyProps := false
	foundLargeClassComponent := false

	for _, issue := range metrics.ReactAnalysis.ComponentIssues {
		if issue.IssueType == "large_component" {
			foundLargeComponent = true
		}
		if issue.IssueType == "too_many_props" {
			foundTooManyProps = true
		}
		if issue.IssueType == "large_class_component" {
			foundLargeClassComponent = true
		}
	}

	assert.True(t, foundLargeComponent, "Should detect large component")
	assert.True(t, foundTooManyProps, "Should detect too many props")
	assert.True(t, foundLargeClassComponent, "Should detect large class component")
}

func TestIdentifyBottlenecks(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()
	metrics := &PerformanceMetrics{
		Bottlenecks: []PerformanceBottleneck{},
	}

	complexityMetrics := &ComplexityMetrics{
		FunctionMetrics: []FunctionComplexity{
			{
				Name:          "complexFunction",
				FilePath:      "test.js",
				StartLine:     1,
				EndLine:       50,
				WeightedScore: 25.0, // High complexity
			},
		},
	}

	parseResults := []*ast.ParseResult{
		{
			FilePath: "test.js",
			Classes: []ast.ClassInfo{
				{
					Name:      "HugeClass",
					StartLine: 1,
					EndLine:   600,                          // Very large class
					Methods:   make([]ast.FunctionInfo, 25), // Many methods
				},
			},
			Imports: make([]ast.ImportInfo, 35), // Many imports
		},
	}

	analyzer.identifyBottlenecks(parseResults, complexityMetrics, metrics)

	assert.Greater(t, len(metrics.Bottlenecks), 0)

	// Check for different types of bottlenecks
	foundComplexityBottleneck := false
	foundLargeClassBottleneck := false
	foundCouplingBottleneck := false

	for _, bottleneck := range metrics.Bottlenecks {
		if bottleneck.Type == "high_complexity_function" {
			foundComplexityBottleneck = true
		}
		if bottleneck.Type == "large_class" {
			foundLargeClassBottleneck = true
		}
		if bottleneck.Type == "high_coupling" {
			foundCouplingBottleneck = true
		}
	}

	assert.True(t, foundComplexityBottleneck, "Should detect complexity bottleneck")
	assert.True(t, foundLargeClassBottleneck, "Should detect large class bottleneck")
	assert.True(t, foundCouplingBottleneck, "Should detect coupling bottleneck")
}

func TestPerformanceGrading(t *testing.T) {
	analyzer := NewPerformanceAnalyzer()

	tests := []struct {
		score    float64
		expected string
	}{
		{95.0, "A"},
		{85.0, "B"},
		{75.0, "C"},
		{65.0, "D"},
		{45.0, "F"},
	}

	for _, tt := range tests {
		grade := analyzer.getPerformanceGrade(tt.score)
		assert.Equal(t, tt.expected, grade, "Score %.1f should get grade %s", tt.score, tt.expected)
	}
}

// Helper functions for creating mock data

func createMockParseResultsForPerformance() []*ast.ParseResult {
	return []*ast.ParseResult{
		{
			FilePath: "test.js",
			Language: "javascript",
			Functions: []ast.FunctionInfo{
				{
					Name:      "processUser",
					IsAsync:   true,
					StartLine: 10,
					EndLine:   25,
					Parameters: []ast.ParameterInfo{
						{Name: "userId", Type: "string"},
						{Name: "options", Type: "object"},
					},
				},
				{
					Name:      "findById",
					StartLine: 30,
					EndLine:   35,
					Parameters: []ast.ParameterInfo{
						{Name: "id", Type: "string"},
					},
				},
			},
			Imports: []ast.ImportInfo{
				{Source: "lodash", ImportType: "default"},
				{Source: "axios", ImportType: "named"},
			},
		},
		{
			FilePath: "component.jsx",
			Language: "tsx",
			Functions: []ast.FunctionInfo{
				{
					Name:      "UserCard",
					StartLine: 5,
					EndLine:   45,
					Parameters: []ast.ParameterInfo{
						{Name: "user", Type: "object"},
						{Name: "onClick", Type: "function"},
					},
				},
			},
			Imports: []ast.ImportInfo{
				{Source: "react", ImportType: "named"},
			},
		},
	}
}

func createMockComplexityMetricsForPerformance() *ComplexityMetrics {
	return &ComplexityMetrics{
		OverallScore:      75.0,
		AverageComplexity: 8.5,
		TotalFunctions:    2,
		FunctionMetrics: []FunctionComplexity{
			{
				Name:            "processUser",
				FilePath:        "test.js",
				StartLine:       10,
				EndLine:         25,
				WeightedScore:   12.0,
				CyclomaticValue: 6,
				CognitiveValue:  8,
				SeverityLevel:   "medium",
			},
			{
				Name:            "findById",
				FilePath:        "test.js",
				StartLine:       30,
				EndLine:         35,
				WeightedScore:   5.0,
				CyclomaticValue: 2,
				CognitiveValue:  3,
				SeverityLevel:   "low",
			},
		},
		ClassMetrics: []ClassComplexity{},
	}
}
