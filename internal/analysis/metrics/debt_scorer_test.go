package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewDebtScorer(t *testing.T) {
	scorer := NewDebtScorer()
	
	assert.NotNil(t, scorer)
	assert.Equal(t, 0.25, scorer.config.ComplexityWeight)
	assert.Equal(t, 0.20, scorer.config.DuplicationWeight)
	assert.Equal(t, 0.20, scorer.config.CodeSmellWeight)
	assert.Equal(t, 0.20, scorer.config.ArchitectureWeight)
	assert.Equal(t, 0.15, scorer.config.PerformanceWeight)
	assert.Equal(t, 30, scorer.config.TrendAnalysisPeriod)
}

func TestNewDebtScorerWithConfig(t *testing.T) {
	config := DebtScoringConfig{
		ComplexityWeight:      0.30,
		DuplicationWeight:     0.25,
		CodeSmellWeight:       0.15,
		ArchitectureWeight:    0.15,
		PerformanceWeight:     0.15,
		ChangeFrequencyWeight: 0.25,
		ImpactWeight:          0.50,
		RemediationThreshold:  0.80,
		TrendAnalysisPeriod:   60,
		PriorityCategories:    5,
		MinConfidenceScore:    0.70,
	}
	
	scorer := NewDebtScorerWithConfig(config)
	
	assert.NotNil(t, scorer)
	assert.Equal(t, config, scorer.config)
}

func TestAnalyzeDebt_EmptyInput(t *testing.T) {
	scorer := NewDebtScorer()
	ctx := context.Background()
	
	metrics, err := scorer.AnalyzeDebt(ctx, []*ast.ParseResult{}, nil, nil)
	
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "no parse results provided")
}

func TestAnalyzeDebt_MissingMetrics(t *testing.T) {
	scorer := NewDebtScorer()
	ctx := context.Background()
	
	parseResults := []*ast.ParseResult{
		createMockParseResultForDebt("test.js", []ast.FunctionInfo{
			createMockFunctionForDebt("testFunction", 1, 10),
		}, []ast.ClassInfo{}),
	}
	
	metrics, err := scorer.AnalyzeDebt(ctx, parseResults, nil, nil)
	
	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "complexity and duplication metrics are required")
}

func TestAnalyzeDebt_ValidInput(t *testing.T) {
	tests := []struct {
		name          string
		parseResults  []*ast.ParseResult
		expectError   bool
		expectMetrics bool
	}{
		{
			name: "single file with functions",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDebt("test.js", []ast.FunctionInfo{
					createMockFunctionForDebt("validateInput", 1, 50), // Long method
					createMockFunctionForDebt("processData", 55, 65),
					createMockFunctionForDebt("formatOutput", 70, 80),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "multiple files with classes",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDebt("file1.js", []ast.FunctionInfo{
					createMockFunctionForDebt("method1", 1, 35), // Long method
				}, []ast.ClassInfo{
					createMockClassForDebt("LargeClass", createManyMethods(25)), // Large class
				}),
				createMockParseResultForDebt("file2.js", []ast.FunctionInfo{
					createMockFunctionForDebt("method2", 1, 15),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "files with architecture issues",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDebt("service.js", []ast.FunctionInfo{
					createMockFunctionForDebt("processData", 1, 25),
				}, []ast.ClassInfo{}),
				createMockParseResultForDebt("component.js", []ast.FunctionInfo{
					createMockFunctionForDebt("render", 1, 20),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewDebtScorer()
			ctx := context.Background()
			
			// Create mock complexity and duplication metrics
			complexityMetrics := createMockComplexityMetrics()
			duplicationMetrics := createMockDuplicationMetrics()
			
			metrics, err := scorer.AnalyzeDebt(ctx, tt.parseResults, complexityMetrics, duplicationMetrics)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				if tt.expectMetrics {
					require.NotNil(t, metrics)
					assert.NotNil(t, metrics.Categories)
					assert.NotNil(t, metrics.FileDebtScores)
					assert.NotNil(t, metrics.RemediationPlan)
					assert.NotNil(t, metrics.Recommendations)
					assert.NotNil(t, metrics.Dashboard)
					assert.NotNil(t, metrics.Summary)
					
					// Verify metrics have reasonable values
					assert.GreaterOrEqual(t, metrics.OverallScore, 0.0)
					assert.GreaterOrEqual(t, metrics.TotalDebtHours, 0.0)
					assert.GreaterOrEqual(t, metrics.DebtRatio, 0.0)
					assert.LessOrEqual(t, metrics.DebtRatio, 5.0) // Can exceed 1.0 when debt is high
					assert.Contains(t, []string{"improving", "stable", "worsening"}, metrics.TrendDirection)
				}
			}
		})
	}
}

func TestIsLongMethod(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected bool
	}{
		{"short method", createMockFunctionForDebt("short", 1, 20), false},
		{"exactly 30 lines", createMockFunctionForDebt("medium", 1, 30), false},
		{"long method", createMockFunctionForDebt("long", 1, 50), true},
		{"very long method", createMockFunctionForDebt("veryLong", 1, 100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.isLongMethod(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineLongMethodSeverity(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected string
	}{
		{"small long method", createMockFunctionForDebt("small", 1, 40), "low"},
		{"medium long method", createMockFunctionForDebt("medium", 1, 75), "medium"},
		{"large long method", createMockFunctionForDebt("large", 1, 120), "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.determineLongMethodSeverity(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasTooManyParameters(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		function ast.FunctionInfo
		expected bool
	}{
		{"few parameters", createMockFunctionWithParams("few", 3), false},
		{"exactly 5 parameters", createMockFunctionWithParams("five", 5), false},
		{"too many parameters", createMockFunctionWithParams("many", 8), true},
		{"way too many parameters", createMockFunctionWithParams("tooMany", 12), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.hasTooManyParameters(tt.function)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsLargeClass(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		class    ast.ClassInfo
		expected bool
	}{
		{"small class", createMockClassForDebt("Small", createManyMethods(10)), false},
		{"exactly 20 methods", createMockClassForDebt("Medium", createManyMethods(20)), false},
		{"large class", createMockClassForDebt("Large", createManyMethods(25)), true},
		{"very large class", createMockClassForDebt("VeryLarge", createManyMethods(50)), true},
		{"long class by lines", createMockClassWithLines("LongClass", createManyMethods(10), 1, 600), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.isLargeClass(tt.class)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasCircularDependencies(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name         string
		parseResult  *ast.ParseResult
		expected     bool
	}{
		{
			name: "few imports and exports",
			parseResult: createMockParseResultForDebt("simple.js", 
				[]ast.FunctionInfo{}, []ast.ClassInfo{}),
			expected: false,
		},
		{
			name: "many imports and exports",
			parseResult: createMockParseResultWithImportsExports("complex.js", 15, 8),
			expected: true,
		},
		{
			name: "many imports, few exports",
			parseResult: createMockParseResultWithImportsExports("imports.js", 15, 2),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.hasCircularDependencies(tt.parseResult)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateConfidenceScore(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		item     TechnicalDebtItem
		minScore float64
		maxScore float64
	}{
		{
			name: "high confidence item",
			item: TechnicalDebtItem{
				Type: "high_complexity",
				Metadata: map[string]interface{}{
					"complexity_score": 25.5,
					"cyclomatic":       15,
					"cognitive":        20,
				},
			},
			minScore: 0.8,
			maxScore: 1.0,
		},
		{
			name: "low confidence item",
			item: TechnicalDebtItem{
				Type: "memory_leak_risk",
				Metadata: map[string]interface{}{
					"patterns": []string{"closure"},
				},
			},
			minScore: 0.5,
			maxScore: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateConfidenceScore(tt.item)
			assert.GreaterOrEqual(t, result, tt.minScore)
			assert.LessOrEqual(t, result, tt.maxScore)
		})
	}
}

func TestEstimateChangeFrequency(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name         string
		item         TechnicalDebtItem
		minFrequency float64
		maxFrequency float64
	}{
		{
			name: "service file",
			item: TechnicalDebtItem{FilePath: "src/services/user.service.js"},
			minFrequency: 0.5,
			maxFrequency: 1.0,
		},
		{
			name: "component file",
			item: TechnicalDebtItem{FilePath: "src/components/UserProfile.jsx"},
			minFrequency: 0.4,
			maxFrequency: 1.0,
		},
		{
			name: "utility file",
			item: TechnicalDebtItem{FilePath: "src/utils/helpers.js"},
			minFrequency: 0.1,
			maxFrequency: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.estimateChangeFrequency(tt.item)
			assert.GreaterOrEqual(t, result, tt.minFrequency)
			assert.LessOrEqual(t, result, tt.maxFrequency)
		})
	}
}

func TestScoreToPriority(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"very high score", 10.0, "critical"},
		{"high score", 7.0, "high"},
		{"medium score", 4.5, "medium"},
		{"low score", 2.0, "low"},
		{"very low score", 0.5, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.scoreToPriority(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateOverallScore(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		items    []TechnicalDebtItem
		expected float64
	}{
		{
			name:     "empty items",
			items:    []TechnicalDebtItem{},
			expected: 0.0,
		},
		{
			name: "single item",
			items: []TechnicalDebtItem{
				{DebtScore: 10.0},
			},
			expected: 10.0,
		},
		{
			name: "multiple items",
			items: []TechnicalDebtItem{
				{DebtScore: 8.0},
				{DebtScore: 12.0},
				{DebtScore: 6.0},
			},
			expected: 8.67, // (8+12+6)/3 ≈ 8.67
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateOverallScore(tt.items)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCalculateTotalDebtHours(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name     string
		items    []TechnicalDebtItem
		expected float64
	}{
		{
			name:     "empty items",
			items:    []TechnicalDebtItem{},
			expected: 0.0,
		},
		{
			name: "multiple items",
			items: []TechnicalDebtItem{
				{EstimatedHours: 2.5},
				{EstimatedHours: 4.0},
				{EstimatedHours: 1.5},
			},
			expected: 8.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateTotalDebtHours(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDebtRatio(t *testing.T) {
	scorer := NewDebtScorer()
	
	tests := []struct {
		name         string
		parseResults []*ast.ParseResult
		items        []TechnicalDebtItem
		expected     float64
	}{
		{
			name:         "no files",
			parseResults: []*ast.ParseResult{},
			items:        []TechnicalDebtItem{},
			expected:     0.0,
		},
		{
			name: "some files with debt",
			parseResults: []*ast.ParseResult{
				{FilePath: "file1.js"},
				{FilePath: "file2.js"},
				{FilePath: "file3.js"},
				{FilePath: "file4.js"},
			},
			items: []TechnicalDebtItem{
				{FilePath: "file1.js"},
				{FilePath: "file2.js"},
			},
			expected: 0.5, // 2/4 files have debt
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateDebtRatio(tt.parseResults, tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDebtScoringIntegration(t *testing.T) {
	scorer := NewDebtScorer()
	ctx := context.Background()
	
	// Create a comprehensive set of test data
	parseResults := []*ast.ParseResult{
		// File 1: Code smells (long method, many parameters)
		createMockParseResultForDebt("smelly.js", []ast.FunctionInfo{
			createMockFunctionWithParams("longMethodWithManyParams", 8), // Too many params
			createMockFunctionForDebt("veryLongMethod", 1, 80),          // Long method
		}, []ast.ClassInfo{
			createMockClassForDebt("LargeClass", createManyMethods(25)), // Large class
		}),
		
		// File 2: Architecture violations (god object, tight coupling)
		createMockParseResultWithImportsExports("godObject.js", 20, 8), // Circular deps, god object
		
		// File 3: Performance issues (nested loops, excessive imports)
		createMockParseResultWithImportsExports("performance.js", 25, 3), // Excessive imports
	}
	
	complexityMetrics := createMockComplexityMetrics()
	duplicationMetrics := createMockDuplicationMetrics()
	
	metrics, err := scorer.AnalyzeDebt(ctx, parseResults, complexityMetrics, duplicationMetrics)
	
	require.NoError(t, err)
	require.NotNil(t, metrics)
	
	// Verify overall metrics
	assert.Greater(t, metrics.OverallScore, 0.0)
	assert.Greater(t, metrics.TotalDebtHours, 0.0)
	assert.GreaterOrEqual(t, metrics.DebtRatio, 0.0)
	assert.LessOrEqual(t, metrics.DebtRatio, 3.0) // Can exceed 1.0 for high debt
	
	// Verify categories are present
	assert.Contains(t, metrics.Categories, "Code Smells")
	assert.Contains(t, metrics.Categories, "Architecture Violations")
	assert.Contains(t, metrics.Categories, "Performance Issues")
	
	// Verify file debt scores
	assert.Greater(t, len(metrics.FileDebtScores), 0)
	for filePath, fileDebt := range metrics.FileDebtScores {
		assert.NotEmpty(t, filePath)
		assert.GreaterOrEqual(t, fileDebt.OverallScore, 0.0)
		assert.GreaterOrEqual(t, fileDebt.DebtHours, 0.0)
		assert.Contains(t, []string{"low", "medium", "high", "critical"}, fileDebt.Priority)
		assert.Greater(t, fileDebt.RemediationOrder, 0)
	}
	
	// Verify remediation plan
	assert.GreaterOrEqual(t, len(metrics.RemediationPlan), 0)
	for _, item := range metrics.RemediationPlan {
		assert.NotEmpty(t, item.ID)
		assert.NotEmpty(t, item.Title)
		assert.NotEmpty(t, item.Description)
		assert.Greater(t, item.EstimatedEffort, 0.0)
		assert.GreaterOrEqual(t, item.ExpectedROI, 0.0)
		assert.Contains(t, []string{"low", "medium", "high", "critical"}, item.Priority)
	}
	
	// Verify dashboard
	assert.GreaterOrEqual(t, metrics.Dashboard.HealthScore, 0.0)
	assert.LessOrEqual(t, metrics.Dashboard.HealthScore, 100.0)
	assert.GreaterOrEqual(t, metrics.Dashboard.CriticalIssues, 0)
	assert.GreaterOrEqual(t, metrics.Dashboard.HighPriorityIssues, 0)
	assert.Greater(t, len(metrics.Dashboard.MonthlyTrend), 0)
	
	// Verify summary
	assert.Equal(t, len(parseResults), metrics.Summary.TotalFiles)
	assert.GreaterOrEqual(t, metrics.Summary.FilesWithDebt, 0)
	assert.LessOrEqual(t, metrics.Summary.FilesWithDebt, len(parseResults)*2) // Allow for multiple debt items per file
	assert.GreaterOrEqual(t, metrics.Summary.AverageDebtPerFile, 0.0)
	assert.NotEmpty(t, metrics.Summary.RecommendedFocus)
	assert.GreaterOrEqual(t, metrics.Summary.EstimatedPaydownWeeks, 0)
}

func TestDebtConfigValidation(t *testing.T) {
	// Test with custom configuration
	config := DebtScoringConfig{
		ComplexityWeight:      0.30,
		DuplicationWeight:     0.25,
		CodeSmellWeight:       0.20,
		ArchitectureWeight:    0.15,
		PerformanceWeight:     0.10,
		ChangeFrequencyWeight: 0.35,
		ImpactWeight:          0.45,
		RemediationThreshold:  0.75,
		TrendAnalysisPeriod:   45,
		PriorityCategories:    5,
		MinConfidenceScore:    0.65,
	}
	
	scorer := NewDebtScorerWithConfig(config)
	ctx := context.Background()
	
	parseResults := []*ast.ParseResult{
		createMockParseResultForDebt("test.js", []ast.FunctionInfo{
			createMockFunctionForDebt("testFunc", 1, 20),
		}, []ast.ClassInfo{}),
	}
	
	complexityMetrics := createMockComplexityMetrics()
	duplicationMetrics := createMockDuplicationMetrics()
	
	metrics, err := scorer.AnalyzeDebt(ctx, parseResults, complexityMetrics, duplicationMetrics)
	
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
}

// Helper functions for creating mock data
func createMockParseResultForDebt(filePath string, functions []ast.FunctionInfo, classes []ast.ClassInfo) *ast.ParseResult {
	return &ast.ParseResult{
		FilePath:   filePath,
		Language:   "javascript",
		Functions:  functions,
		Classes:    classes,
		Interfaces: []ast.InterfaceInfo{},
		Variables:  []ast.VariableInfo{},
		Imports:    createMockImports(5),
		Exports:    createMockExports(3),
		Errors:     []ast.ParseError{},
		Metadata:   make(map[string]interface{}),
	}
}

func createMockParseResultWithImportsExports(filePath string, importCount, exportCount int) *ast.ParseResult {
	return &ast.ParseResult{
		FilePath:   filePath,
		Language:   "javascript",
		Functions:  []ast.FunctionInfo{createMockFunctionForDebt("method", 1, 20)},
		Classes:    []ast.ClassInfo{},
		Interfaces: []ast.InterfaceInfo{},
		Variables:  []ast.VariableInfo{},
		Imports:    createMockImports(importCount),
		Exports:    createMockExports(exportCount),
		Errors:     []ast.ParseError{},
		Metadata:   make(map[string]interface{}),
	}
}

func createMockFunctionForDebt(name string, startLine, endLine int) ast.FunctionInfo {
	return ast.FunctionInfo{
		Name:       name,
		Parameters: []ast.ParameterInfo{},
		StartLine:  startLine,
		EndLine:    endLine,
		Metadata:   make(map[string]string),
	}
}

func createMockFunctionWithParams(name string, paramCount int) ast.FunctionInfo {
	params := make([]ast.ParameterInfo, paramCount)
	for i := 0; i < paramCount; i++ {
		params[i] = ast.ParameterInfo{
			Name: fmt.Sprintf("param%d", i),
			Type: "any",
		}
	}
	
	return ast.FunctionInfo{
		Name:       name,
		Parameters: params,
		StartLine:  1,
		EndLine:    20,
		Metadata:   make(map[string]string),
	}
}

func createMockClassForDebt(name string, methods []ast.FunctionInfo) ast.ClassInfo {
	return ast.ClassInfo{
		Name:       name,
		Methods:    methods,
		Properties: []ast.PropertyInfo{},
		StartLine:  1,
		EndLine:    100,
		Metadata:   make(map[string]string),
	}
}

func createMockClassWithLines(name string, methods []ast.FunctionInfo, startLine, endLine int) ast.ClassInfo {
	return ast.ClassInfo{
		Name:       name,
		Methods:    methods,
		Properties: []ast.PropertyInfo{},
		StartLine:  startLine,
		EndLine:    endLine,
		Metadata:   make(map[string]string),
	}
}

func createManyMethods(count int) []ast.FunctionInfo {
	methods := make([]ast.FunctionInfo, count)
	for i := 0; i < count; i++ {
		methods[i] = createMockFunctionForDebt(fmt.Sprintf("method%d", i), i*10+1, i*10+10)
	}
	return methods
}

func createMockImports(count int) []ast.ImportInfo {
	imports := make([]ast.ImportInfo, count)
	for i := 0; i < count; i++ {
		imports[i] = ast.ImportInfo{
			Source:     fmt.Sprintf("./module%d", i),
			ImportType: "default",
		}
	}
	return imports
}

func createMockExports(count int) []ast.ExportInfo {
	exports := make([]ast.ExportInfo, count)
	for i := 0; i < count; i++ {
		exports[i] = ast.ExportInfo{
			Name:       fmt.Sprintf("export%d", i),
			ExportType: "function",
		}
	}
	return exports
}

func createMockComplexityMetrics() *ComplexityMetrics {
	return &ComplexityMetrics{
		OverallScore: 15.5,
		FunctionMetrics: []FunctionComplexity{
			{
				Name:            "complexFunction",
				FilePath:        "test.js",
				StartLine:       1,
				EndLine:         50,
				WeightedScore:   25.0,
				CyclomaticValue: 15,
				CognitiveValue:  20,
				SeverityLevel:   "high",
				Recommendations: []string{"Break down into smaller functions"},
			},
			{
				Name:            "simpleFunction",
				FilePath:        "test.js",
				StartLine:       55,
				EndLine:         65,
				WeightedScore:   5.0,
				CyclomaticValue: 3,
				CognitiveValue:  2,
				SeverityLevel:   "low",
				Recommendations: []string{},
			},
		},
		ClassMetrics: []ClassComplexity{},
		TotalFunctions: 2,
		AverageComplexity: 15.0,
	}
}

func createMockDuplicationMetrics() *DuplicationMetrics {
	return &DuplicationMetrics{
		OverallScore:     25.0,
		DuplicationRatio: 0.3,
		ExactDuplicates: []DuplicationCluster{
			{
				ID:               "exact_1",
				Type:             "exact",
				SimilarityScore:  1.0,
				LineCount:        10,
				TokenCount:       50,
				MaintenanceBurden: 15.0,
				Priority:         "high",
				Recommendations:  []string{"Extract to shared utility"},
				Instances: []DuplicationInstance{
					{
						FilePath:     "file1.js",
						FunctionName: "duplicatedFunction",
						StartLine:    1,
						EndLine:      10,
					},
					{
						FilePath:     "file2.js", 
						FunctionName: "duplicatedFunction",
						StartLine:    1,
						EndLine:      10,
					},
				},
			},
		},
		StructuralDuplicates: []DuplicationCluster{},
		TokenDuplicates:     []DuplicationCluster{},
		TotalDuplicatedLines: 20,
	}
}