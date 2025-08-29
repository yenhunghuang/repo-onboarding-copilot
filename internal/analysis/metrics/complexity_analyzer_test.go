package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewComplexityAnalyzer(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	assert.NotNil(t, analyzer)
	assert.Equal(t, 10, analyzer.config.LowThreshold)
	assert.Equal(t, 15, analyzer.config.MediumThreshold)
	assert.Equal(t, 20, analyzer.config.HighThreshold)
	assert.Equal(t, 4, analyzer.config.MaxNestingDepth)
	assert.Equal(t, 20, analyzer.config.ReportTopN)
	assert.True(t, analyzer.config.EnableTrends)
}

func TestNewComplexityAnalyzerWithConfig(t *testing.T) {
	customConfig := ComplexityConfig{
		LowThreshold:    5,
		MediumThreshold: 10,
		HighThreshold:   15,
		MaxNestingDepth: 3,
		ReportTopN:      10,
		EnableTrends:    false,
		WeightFactors: Weights{
			Cyclomatic:   0.5,
			Cognitive:    0.3,
			NestingDepth: 0.2,
		},
	}

	analyzer := NewComplexityAnalyzerWithConfig(customConfig)

	assert.NotNil(t, analyzer)
	assert.Equal(t, customConfig.LowThreshold, analyzer.config.LowThreshold)
	assert.Equal(t, customConfig.MediumThreshold, analyzer.config.MediumThreshold)
	assert.Equal(t, customConfig.HighThreshold, analyzer.config.HighThreshold)
	assert.Equal(t, customConfig.MaxNestingDepth, analyzer.config.MaxNestingDepth)
	assert.False(t, analyzer.config.EnableTrends)
}

func TestAnalyzeComplexity_EmptyInput(t *testing.T) {
	analyzer := NewComplexityAnalyzer()
	ctx := context.Background()

	metrics, err := analyzer.AnalyzeComplexity(ctx, []*ast.ParseResult{})

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "no parse results provided")
}

func TestAnalyzeComplexity_ValidInput(t *testing.T) {
	tests := []struct {
		name          string
		parseResults  []*ast.ParseResult
		expectError   bool
		expectMetrics bool
	}{
		{
			name: "single file with functions",
			parseResults: []*ast.ParseResult{
				createMockParseResult("test.js", []ast.FunctionInfo{
					createMockFunction("simpleFunction", 1, 5, []ast.ParameterInfo{}),
					createMockFunction("complexFunction", 10, 50, []ast.ParameterInfo{
						{Name: "param1", Type: "string"},
						{Name: "param2", Type: "number"},
						{Name: "param3", Type: "boolean"},
					}),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "single file with classes",
			parseResults: []*ast.ParseResult{
				createMockParseResult("test.js", []ast.FunctionInfo{}, []ast.ClassInfo{
					createMockClass("TestClass", []ast.FunctionInfo{
						createMockFunction("method1", 1, 10, []ast.ParameterInfo{}),
						createMockFunction("method2", 15, 30, []ast.ParameterInfo{}),
					}),
				}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "multiple files",
			parseResults: []*ast.ParseResult{
				createMockParseResult("file1.js", []ast.FunctionInfo{
					createMockFunction("func1", 1, 10, []ast.ParameterInfo{}),
				}, []ast.ClassInfo{}),
				createMockParseResult("file2.js", []ast.FunctionInfo{
					createMockFunction("func2", 1, 20, []ast.ParameterInfo{}),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewComplexityAnalyzer()
			ctx := context.Background()

			metrics, err := analyzer.AnalyzeComplexity(ctx, tt.parseResults)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				if tt.expectMetrics {
					assert.NotNil(t, metrics)
					assert.NotNil(t, metrics.FunctionMetrics)
					assert.NotNil(t, metrics.FileMetrics)
					assert.NotNil(t, metrics.Recommendations)
				}
			}
		})
	}
}

func TestCalculateBasicComplexity(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected int
	}{
		{
			name:     "simple function",
			function: createMockFunction("simple", 1, 5, []ast.ParameterInfo{}),
			expected: 1, // Base complexity
		},
		{
			name: "function with parameters",
			function: createMockFunction("withParams", 1, 10, []ast.ParameterInfo{
				{Name: "param1"}, {Name: "param2"}, {Name: "param3"},
			}),
			expected: 2, // Base + parameters/3
		},
		{
			name: "async function",
			function: ast.FunctionInfo{
				Name:       "asyncFunc",
				StartLine:  1,
				EndLine:    10,
				Parameters: []ast.ParameterInfo{},
				IsAsync:    true,
			},
			expected: 2, // Base + async
		},
		{
			name:     "large function",
			function: createMockFunction("large", 1, 60, []ast.ParameterInfo{}),
			expected: 3, // Base + size penalty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateBasicComplexity(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateCognitiveComplexity(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name     string
		function ast.FunctionInfo
		minimum  int // Cognitive should be at least this value
	}{
		{
			name:     "simple function",
			function: createMockFunction("simple", 1, 5, []ast.ParameterInfo{}),
			minimum:  1,
		},
		{
			name:     "complex function",
			function: createMockFunction("complex", 1, 50, []ast.ParameterInfo{}),
			minimum:  3, // Should have nesting penalty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateCognitiveComplexity(tt.function)
			assert.GreaterOrEqual(t, result, tt.minimum)
		})
	}
}

func TestCalculateNestingDepth(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected int
	}{
		{
			name:     "small function",
			function: createMockFunction("small", 1, 10, []ast.ParameterInfo{}),
			expected: 1,
		},
		{
			name:     "medium function",
			function: createMockFunction("medium", 1, 25, []ast.ParameterInfo{}),
			expected: 2,
		},
		{
			name:     "large function",
			function: createMockFunction("large", 1, 60, []ast.ParameterInfo{}),
			expected: 4,
		},
		{
			name:     "very large function",
			function: createMockFunction("veryLarge", 1, 120, []ast.ParameterInfo{}),
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateNestingDepth(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineSeverityLevel(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name       string
		complexity int
		expected   string
	}{
		{"low complexity", 5, "low"},
		{"medium complexity", 12, "medium"},
		{"high complexity", 18, "high"},
		{"severe complexity", 25, "severe"},
		{"at low threshold", 10, "medium"},
		{"at medium threshold", 15, "high"},
		{"at high threshold", 20, "severe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.determineSeverityLevel(tt.complexity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAssessRefactoringRisk(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name       string
		complexity *FunctionComplexity
		expected   string
	}{
		{
			name: "low risk",
			complexity: &FunctionComplexity{
				CyclomaticValue: 5,
				NestingDepth:    2,
			},
			expected: "low",
		},
		{
			name: "medium risk",
			complexity: &FunctionComplexity{
				CyclomaticValue: 12,
				NestingDepth:    3,
			},
			expected: "medium",
		},
		{
			name: "high risk",
			complexity: &FunctionComplexity{
				CyclomaticValue: 18,
				NestingDepth:    4,
			},
			expected: "high",
		},
		{
			name: "critical risk",
			complexity: &FunctionComplexity{
				CyclomaticValue: 30,
				NestingDepth:    5,
			},
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.assessRefactoringRisk(tt.complexity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAssessTestingDifficulty(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name       string
		complexity *FunctionComplexity
		expected   string
	}{
		{
			name: "easy testing",
			complexity: &FunctionComplexity{
				CyclomaticValue: 3,
				NestingDepth:    1,
			},
			expected: "easy",
		},
		{
			name: "moderate testing",
			complexity: &FunctionComplexity{
				CyclomaticValue: 8,
				NestingDepth:    2,
			},
			expected: "moderate",
		},
		{
			name: "difficult testing",
			complexity: &FunctionComplexity{
				CyclomaticValue: 15,
				NestingDepth:    3,
			},
			expected: "difficult",
		},
		{
			name: "very difficult testing",
			complexity: &FunctionComplexity{
				CyclomaticValue: 25,
				NestingDepth:    4,
			},
			expected: "very_difficult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.assessTestingDifficulty(tt.complexity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectAntiPatterns(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name       string
		function   ast.FunctionInfo
		complexity *FunctionComplexity
		expected   []string
	}{
		{
			name:       "simple function - no patterns",
			function:   createMockFunction("simple", 1, 10, []ast.ParameterInfo{}),
			complexity: &FunctionComplexity{CyclomaticValue: 5, NestingDepth: 2},
			expected:   []string{},
		},
		{
			name:       "large function",
			function:   createMockFunction("large", 1, 120, []ast.ParameterInfo{}),
			complexity: &FunctionComplexity{CyclomaticValue: 5, NestingDepth: 2},
			expected:   []string{"large_function"},
		},
		{
			name:       "high complexity",
			function:   createMockFunction("complex", 1, 20, []ast.ParameterInfo{}),
			complexity: &FunctionComplexity{CyclomaticValue: 25, NestingDepth: 2},
			expected:   []string{"high_complexity"},
		},
		{
			name:       "deep nesting",
			function:   createMockFunction("nested", 1, 30, []ast.ParameterInfo{}),
			complexity: &FunctionComplexity{CyclomaticValue: 10, NestingDepth: 5},
			expected:   []string{"deep_nesting"},
		},
		{
			name: "too many parameters",
			function: createMockFunction("params", 1, 20, []ast.ParameterInfo{
				{Name: "p1"}, {Name: "p2"}, {Name: "p3"}, {Name: "p4"},
				{Name: "p5"}, {Name: "p6"}, {Name: "p7"}, {Name: "p8"},
			}),
			complexity: &FunctionComplexity{CyclomaticValue: 8, NestingDepth: 2},
			expected:   []string{"too_many_parameters"},
		},
		{
			name: "multiple patterns",
			function: createMockFunction("problematic", 1, 150, []ast.ParameterInfo{
				{Name: "p1"}, {Name: "p2"}, {Name: "p3"}, {Name: "p4"},
				{Name: "p5"}, {Name: "p6"}, {Name: "p7"}, {Name: "p8"},
			}),
			complexity: &FunctionComplexity{CyclomaticValue: 25, NestingDepth: 5},
			expected:   []string{"large_function", "high_complexity", "deep_nesting", "too_many_parameters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.detectAntiPatterns(tt.function, tt.complexity)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestGenerateFunctionRecommendations(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name           string
		complexity     *FunctionComplexity
		expectedMinLen int
	}{
		{
			name: "high complexity function",
			complexity: &FunctionComplexity{
				CyclomaticValue:   25,
				NestingDepth:      3,
				TestingDifficulty: "moderate",
			},
			expectedMinLen: 2, // Should get refactoring recommendations
		},
		{
			name: "deeply nested function",
			complexity: &FunctionComplexity{
				CyclomaticValue:   12,
				NestingDepth:      5,
				TestingDifficulty: "moderate",
			},
			expectedMinLen: 2, // Should get nesting recommendations
		},
		{
			name: "difficult to test function",
			complexity: &FunctionComplexity{
				CyclomaticValue:   15,
				NestingDepth:      3,
				TestingDifficulty: "very_difficult",
			},
			expectedMinLen: 2, // Should get testing recommendations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.generateFunctionRecommendations(tt.complexity)
			assert.GreaterOrEqual(t, len(result), tt.expectedMinLen)
		})
	}
}

func TestAssessMaintainabilityRisk(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name          string
		avgComplexity float64
		maxComplexity int
		expected      string
	}{
		{"low risk", 5.0, 10, "low"},
		{"medium risk", 8.0, 18, "medium"},
		{"high risk", 12.0, 22, "high"},
		{"critical risk", 20.0, 30, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.assessMaintainabilityRisk(tt.avgComplexity, tt.maxComplexity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAssessClassRisk(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name        string
		classMetric *ClassComplexity
		expected    string
	}{
		{
			name: "low risk class",
			classMetric: &ClassComplexity{
				AverageMethod: 5.0,
				MaxMethod:     8,
			},
			expected: "low",
		},
		{
			name: "medium risk class",
			classMetric: &ClassComplexity{
				AverageMethod: 8.0,
				MaxMethod:     18,
			},
			expected: "medium",
		},
		{
			name: "high risk class",
			classMetric: &ClassComplexity{
				AverageMethod: 12.0,
				MaxMethod:     22,
			},
			expected: "high",
		},
		{
			name: "critical risk class",
			classMetric: &ClassComplexity{
				AverageMethod: 18.0,
				MaxMethod:     30,
			},
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.assessClassRisk(tt.classMetric)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateWeightedScore(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	complexity := &FunctionComplexity{
		CyclomaticValue: 10,
		CognitiveValue:  8,
		NestingDepth:    3,
		ComplexityFactors: ComplexityFactors{
			NestedLoops:    2,
			DecisionPoints: 5,
		},
	}

	score := analyzer.calculateWeightedScore(complexity)

	// Should be a weighted combination of all factors
	assert.Greater(t, score, 0.0)
	assert.Less(t, score, 100.0) // Reasonable upper bound
}

// Helper functions for creating mock data

func createMockParseResult(filePath string, functions []ast.FunctionInfo, classes []ast.ClassInfo) *ast.ParseResult {
	return &ast.ParseResult{
		FilePath:   filePath,
		Language:   "javascript",
		Functions:  functions,
		Classes:    classes,
		Interfaces: []ast.InterfaceInfo{},
		Variables:  []ast.VariableInfo{},
		Imports:    []ast.ImportInfo{},
		Exports:    []ast.ExportInfo{},
		Errors:     []ast.ParseError{},
		Metadata:   make(map[string]interface{}),
	}
}

func createMockFunction(name string, startLine, endLine int, params []ast.ParameterInfo) ast.FunctionInfo {
	return ast.FunctionInfo{
		Name:       name,
		Parameters: params,
		StartLine:  startLine,
		EndLine:    endLine,
		Metadata:   make(map[string]string),
	}
}

func createMockClass(name string, methods []ast.FunctionInfo) ast.ClassInfo {
	return ast.ClassInfo{
		Name:       name,
		Methods:    methods,
		Properties: []ast.PropertyInfo{},
		Metadata:   make(map[string]string),
	}
}

// Integration tests

func TestComplexityAnalysisIntegration(t *testing.T) {
	analyzer := NewComplexityAnalyzer()
	ctx := context.Background()

	// Create a realistic scenario with mixed complexity functions
	parseResults := []*ast.ParseResult{
		createMockParseResult("component.js", []ast.FunctionInfo{
			// Simple function
			createMockFunction("getDisplayName", 1, 5, []ast.ParameterInfo{
				{Name: "name", Type: "string"},
			}),
			// Medium complexity function
			createMockFunction("validateInput", 10, 25, []ast.ParameterInfo{
				{Name: "input", Type: "any"},
				{Name: "options", Type: "ValidationOptions"},
			}),
			// High complexity function
			{
				Name:      "processComplexData",
				StartLine: 30,
				EndLine:   80,
				Parameters: []ast.ParameterInfo{
					{Name: "data", Type: "any[]"},
					{Name: "filters", Type: "Filter[]"},
					{Name: "options", Type: "ProcessingOptions"},
					{Name: "callback", Type: "Function"},
				},
				IsAsync:  true,
				Metadata: make(map[string]string),
			},
		}, []ast.ClassInfo{
			// Class with mixed method complexity
			{
				Name: "DataProcessor",
				Methods: []ast.FunctionInfo{
					createMockFunction("constructor", 1, 5, []ast.ParameterInfo{}),
					createMockFunction("process", 10, 50, []ast.ParameterInfo{}),
					createMockFunction("cleanup", 60, 65, []ast.ParameterInfo{}),
				},
				Properties: []ast.PropertyInfo{},
				Metadata:   make(map[string]string),
			},
		}),
	}

	metrics, err := analyzer.AnalyzeComplexity(ctx, parseResults)

	require.NoError(t, err)
	require.NotNil(t, metrics)

	// Verify overall metrics
	assert.Greater(t, metrics.TotalFunctions, 0)
	assert.Greater(t, metrics.AverageComplexity, 0.0)
	assert.Greater(t, metrics.MaxComplexity, 0)
	assert.Greater(t, metrics.OverallScore, 0.0)

	// Verify function metrics
	assert.Greater(t, len(metrics.FunctionMetrics), 0)
	for _, funcMetric := range metrics.FunctionMetrics {
		assert.NotEmpty(t, funcMetric.Name)
		assert.Greater(t, funcMetric.CyclomaticValue, 0)
		assert.Greater(t, funcMetric.CognitiveValue, 0)
		assert.Greater(t, funcMetric.NestingDepth, 0)
		assert.NotEmpty(t, funcMetric.SeverityLevel)
		assert.NotEmpty(t, funcMetric.RefactoringRisk)
		assert.NotEmpty(t, funcMetric.TestingDifficulty)
		assert.Greater(t, funcMetric.WeightedScore, 0.0)
	}

	// Verify class metrics
	assert.Greater(t, len(metrics.ClassMetrics), 0)
	for _, classMetric := range metrics.ClassMetrics {
		assert.NotEmpty(t, classMetric.Name)
		assert.Greater(t, classMetric.TotalComplexity, 0)
		assert.Greater(t, classMetric.MethodCount, 0)
		assert.NotEmpty(t, classMetric.OverallRisk)
	}

	// Verify file metrics
	assert.Greater(t, len(metrics.FileMetrics), 0)

	// Verify complexity breakdown
	breakdown := metrics.ComplexityByLevel
	totalCategorized := breakdown.Low.Count + breakdown.Medium.Count + breakdown.High.Count + breakdown.Severe.Count
	assert.Equal(t, metrics.TotalFunctions, totalCategorized)

	// Verify recommendations are generated
	assert.NotNil(t, metrics.Recommendations)

	// Verify summary
	assert.NotNil(t, metrics.Summary)
	assert.Greater(t, metrics.Summary.HealthScore, 0.0)
	assert.NotEmpty(t, metrics.Summary.RiskLevel)
	assert.NotEmpty(t, metrics.Summary.TechnicalDebt)
	assert.NotEmpty(t, metrics.Summary.MaintenanceRisk)

	// Verify trend analysis (if enabled)
	if analyzer.config.EnableTrends {
		assert.NotNil(t, metrics.TrendAnalysis)
		assert.NotEmpty(t, metrics.TrendAnalysis.Direction)
		assert.Greater(t, metrics.TrendAnalysis.HealthScore, 0.0)
	}
}

func TestComplexityConfigValidation(t *testing.T) {
	// Test that custom configurations work correctly
	customConfig := ComplexityConfig{
		LowThreshold:    8,
		MediumThreshold: 12,
		HighThreshold:   18,
		MaxNestingDepth: 3,
		ReportTopN:      15,
		EnableTrends:    true,
		WeightFactors: Weights{
			Cyclomatic:   0.5,
			Cognitive:    0.25,
			NestingDepth: 0.15,
			NestedLoops:  0.05,
			Conditionals: 0.05,
		},
	}

	analyzer := NewComplexityAnalyzerWithConfig(customConfig)

	// Test that thresholds are respected
	assert.Equal(t, "low", analyzer.determineSeverityLevel(5))
	assert.Equal(t, "medium", analyzer.determineSeverityLevel(10))
	assert.Equal(t, "high", analyzer.determineSeverityLevel(15))
	assert.Equal(t, "severe", analyzer.determineSeverityLevel(20))
}

// Benchmark tests

func BenchmarkAnalyzeComplexity(b *testing.B) {
	analyzer := NewComplexityAnalyzer()
	ctx := context.Background()

	// Create benchmark data
	parseResults := make([]*ast.ParseResult, 10)
	for i := 0; i < 10; i++ {
		functions := make([]ast.FunctionInfo, 20)
		for j := 0; j < 20; j++ {
			functions[j] = createMockFunction(
				fmt.Sprintf("func_%d_%d", i, j),
				j*10+1,
				j*10+15,
				[]ast.ParameterInfo{
					{Name: "param1", Type: "string"},
					{Name: "param2", Type: "number"},
				},
			)
		}

		parseResults[i] = createMockParseResult(
			fmt.Sprintf("file_%d.js", i),
			functions,
			[]ast.ClassInfo{},
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeComplexity(ctx, parseResults)
		if err != nil {
			b.Fatal(err)
		}
	}
}
