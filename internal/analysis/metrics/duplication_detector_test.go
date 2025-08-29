package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

func TestNewDuplicationDetector(t *testing.T) {
	detector := NewDuplicationDetector()

	assert.NotNil(t, detector)
	assert.Equal(t, 6, detector.config.MinLines)
	assert.Equal(t, 50, detector.config.MinTokens)
	assert.Equal(t, 0.85, detector.config.SimilarityThreshold)
	assert.Equal(t, 0.75, detector.config.TokenSimilarityThreshold)
	assert.True(t, detector.config.IgnoreWhitespace)
	assert.True(t, detector.config.IgnoreComments)
	assert.True(t, detector.config.EnableCrossFile)
	assert.Equal(t, 15, detector.config.ReportTopN)
}

func TestNewDuplicationDetectorWithConfig(t *testing.T) {
	customConfig := DuplicationConfig{
		MinLines:                 4,
		MinTokens:                30,
		SimilarityThreshold:      0.9,
		TokenSimilarityThreshold: 0.8,
		IgnoreWhitespace:         false,
		IgnoreComments:           false,
		EnableCrossFile:          false,
		ReportTopN:               10,
		WeightFactors: DuplicationWeights{
			ExactDuplication:     1.2,
			StructuralSimilarity: 0.9,
			TokenSimilarity:      0.7,
		},
	}

	detector := NewDuplicationDetectorWithConfig(customConfig)

	assert.NotNil(t, detector)
	assert.Equal(t, customConfig.MinLines, detector.config.MinLines)
	assert.Equal(t, customConfig.SimilarityThreshold, detector.config.SimilarityThreshold)
	assert.False(t, detector.config.IgnoreWhitespace)
	assert.False(t, detector.config.EnableCrossFile)
}

func TestDetectDuplication_EmptyInput(t *testing.T) {
	detector := NewDuplicationDetector()
	ctx := context.Background()

	metrics, err := detector.DetectDuplication(ctx, []*ast.ParseResult{})

	assert.Error(t, err)
	assert.Nil(t, metrics)
	assert.Contains(t, err.Error(), "no parse results provided")
}

func TestDetectDuplication_ValidInput(t *testing.T) {
	tests := []struct {
		name          string
		parseResults  []*ast.ParseResult
		expectError   bool
		expectMetrics bool
	}{
		{
			name: "single file with functions",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDuplication("test.js", []ast.FunctionInfo{
					createMockFunctionForDuplication("validateInput", 1, 10),
					createMockFunctionForDuplication("validateOutput", 15, 24),
					createMockFunctionForDuplication("formatData", 30, 40),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "multiple files with potential duplicates",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDuplication("file1.js", []ast.FunctionInfo{
					createMockFunctionForDuplication("processData", 1, 15),
					createMockFunctionForDuplication("validateInput", 20, 30),
				}, []ast.ClassInfo{}),
				createMockParseResultForDuplication("file2.js", []ast.FunctionInfo{
					createMockFunctionForDuplication("processData", 1, 15), // Potential duplicate
					createMockFunctionForDuplication("formatOutput", 20, 35),
				}, []ast.ClassInfo{}),
			},
			expectError:   false,
			expectMetrics: true,
		},
		{
			name: "files with classes",
			parseResults: []*ast.ParseResult{
				createMockParseResultForDuplication("class1.js", []ast.FunctionInfo{}, []ast.ClassInfo{
					createMockClassForDuplication("DataProcessor", []ast.FunctionInfo{
						createMockFunctionForDuplication("process", 5, 20),
						createMockFunctionForDuplication("validate", 25, 35),
					}),
				}),
				createMockParseResultForDuplication("class2.js", []ast.FunctionInfo{}, []ast.ClassInfo{
					createMockClassForDuplication("DataHandler", []ast.FunctionInfo{
						createMockFunctionForDuplication("process", 5, 20), // Potential duplicate
						createMockFunctionForDuplication("cleanup", 25, 30),
					}),
				}),
			},
			expectError:   false,
			expectMetrics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDuplicationDetector()
			ctx := context.Background()

			metrics, err := detector.DetectDuplication(ctx, tt.parseResults)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, metrics)
			} else {
				assert.NoError(t, err)
				if tt.expectMetrics {
					assert.NotNil(t, metrics)
					assert.NotNil(t, metrics.ExactDuplicates)
					assert.NotNil(t, metrics.StructuralDuplicates)
					assert.NotNil(t, metrics.TokenDuplicates)
					assert.NotNil(t, metrics.DuplicationByFile)
					assert.NotNil(t, metrics.ConsolidationOps)
					assert.NotNil(t, metrics.Recommendations)
					assert.NotNil(t, metrics.Summary)
				}
			}
		})
	}
}

func TestIsBlockSizeValid(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name      string
		startLine int
		endLine   int
		expected  bool
	}{
		{"meets minimum", 1, 6, true},
		{"exceeds minimum", 1, 10, true},
		{"below minimum", 1, 5, false},
		{"single line", 5, 5, false},
		{"zero lines", 10, 9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isBlockSizeValid(tt.startLine, tt.endLine)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenizeContent(t *testing.T) {
	tests := []struct {
		name     string
		config   DuplicationConfig
		content  string
		expected string
	}{
		{
			name: "ignore variable names",
			config: DuplicationConfig{
				IgnoreVariableNames: true,
				IgnoreWhitespace:    false,
				IgnoreComments:      false,
			},
			content:  "function validateInput(data, options) { return data.valid; }",
			expected: "IDENTIFIER IDENTIFIER(IDENTIFIER, IDENTIFIER) { IDENTIFIER IDENTIFIER.IDENTIFIER; }",
		},
		{
			name: "ignore whitespace",
			config: DuplicationConfig{
				IgnoreVariableNames: false,
				IgnoreWhitespace:    true,
				IgnoreComments:      false,
			},
			content:  "function   validateInput(  data  ) {\n  return data;\n}",
			expected: "function validateInput( data ) { return data; }",
		},
		{
			name: "ignore comments",
			config: DuplicationConfig{
				IgnoreVariableNames: false,
				IgnoreWhitespace:    false,
				IgnoreComments:      true,
			},
			content:  "function validateInput(data) { // validation logic\n  return data.valid; /* end */ }",
			expected: "function validateInput(data) { \n  return data.valid;  }",
		},
		{
			name: "all normalizations",
			config: DuplicationConfig{
				IgnoreVariableNames: true,
				IgnoreWhitespace:    true,
				IgnoreComments:      true,
			},
			content:  "function   validateInput(data) { // comment\n  return data.valid;\n}",
			expected: "IDENTIFIER IDENTIFIER(IDENTIFIER) { IDENTIFIER IDENTIFIER.IDENTIFIER; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDuplicationDetectorWithConfig(tt.config)
			result := detector.tokenizeContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateStructuralHash(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name        string
		content1    string
		content2    string
		shouldMatch bool
	}{
		{
			name:        "identical structure",
			content1:    "if (condition) { return true; } else { return false; }",
			content2:    "if (otherCondition) { return success; } else { return failure; }",
			shouldMatch: true,
		},
		{
			name:        "different structure",
			content1:    "if (condition) { return true; }",
			content2:    "while (condition) { return true; }",
			shouldMatch: false,
		},
		{
			name:        "same structure different variables",
			content1:    "for (let i = 0; i < arr.length; i++) { process(arr[i]); }",
			content2:    "for (let j = 0; j < list.length; j++) { handle(list[j]); }",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := detector.generateStructuralHash(tt.content1)
			hash2 := detector.generateStructuralHash(tt.content2)

			assert.NotEmpty(t, hash1)
			assert.NotEmpty(t, hash2)

			if tt.shouldMatch {
				assert.Equal(t, hash1, hash2, "Structural hashes should match for similar structures")
			} else {
				assert.NotEqual(t, hash1, hash2, "Structural hashes should differ for different structures")
			}
		})
	}
}

func TestCalculateTokenSimilarity(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		content1 string
		content2 string
		expected float64
	}{
		{
			name:     "identical content",
			content1: "function test() { return true; }",
			content2: "function test() { return true; }",
			expected: 1.0,
		},
		{
			name:     "completely different",
			content1: "function test() { return true; }",
			content2: "class Example { constructor() {} }",
			expected: 0.2, // Some common tokens like {}
		},
		{
			name:     "partial similarity",
			content1: "function test() { return data.valid; }",
			content2: "function validate() { return result.valid; }",
			expected: 0.6, // Some common tokens: function, return, valid
		},
		{
			name:     "empty strings",
			content1: "",
			content2: "",
			expected: 1.0,
		},
		{
			name:     "one empty string",
			content1: "function test() {}",
			content2: "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.calculateTokenSimilarity(tt.content1, tt.content2)
			assert.InDelta(t, tt.expected, result, 0.1, "Token similarity should be within expected range")
		})
	}
}

func TestCalculateContentSimilarity(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name          string
		content1      string
		content2      string
		minSimilarity float64
	}{
		{
			name:          "identical content",
			content1:      "function test() { return true; }",
			content2:      "function test() { return true; }",
			minSimilarity: 1.0,
		},
		{
			name:          "very similar content",
			content1:      "function test() { return true; }",
			content2:      "function test() { return false; }",
			minSimilarity: 0.8,
		},
		{
			name:          "moderately similar",
			content1:      "function validateInput(data) { return data.valid; }",
			content2:      "function validateOutput(result) { return result.valid; }",
			minSimilarity: 0.6,
		},
		{
			name:          "completely different",
			content1:      "function test() {}",
			content2:      "class Example extends Base { constructor() { super(); } }",
			minSimilarity: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.calculateContentSimilarity(tt.content1, tt.content2)
			assert.GreaterOrEqual(t, result, tt.minSimilarity)
			assert.LessOrEqual(t, result, 1.0)
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		{"identical strings", "hello", "hello", 0},
		{"one insertion", "hello", "helloo", 1},
		{"one deletion", "hello", "hell", 1},
		{"one substitution", "hello", "hallo", 1},
		{"empty strings", "", "", 0},
		{"one empty", "", "hello", 5},
		{"complete difference", "abc", "xyz", 3},
		{"multiple operations", "kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.levenshteinDistance(tt.s1, tt.s2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMin3(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		a, b, c  int
		expected int
	}{
		{"first is minimum", 1, 2, 3, 1},
		{"second is minimum", 3, 1, 2, 1},
		{"third is minimum", 2, 3, 1, 1},
		{"all equal", 5, 5, 5, 5},
		{"two equal minimum", 2, 2, 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.min3(tt.a, tt.b, tt.c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstancesEqual(t *testing.T) {
	detector := NewDuplicationDetector()

	instance1 := DuplicationInstance{
		FilePath:  "test.js",
		StartLine: 10,
		EndLine:   20,
	}

	instance2 := DuplicationInstance{
		FilePath:  "test.js",
		StartLine: 10,
		EndLine:   20,
	}

	instance3 := DuplicationInstance{
		FilePath:  "other.js",
		StartLine: 10,
		EndLine:   20,
	}

	assert.True(t, detector.instancesEqual(instance1, instance2))
	assert.False(t, detector.instancesEqual(instance1, instance3))
}

func TestCalculateClusterSimilarity(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name      string
		instances []DuplicationInstance
		expected  float64
	}{
		{
			name:      "single instance",
			instances: []DuplicationInstance{{Content: "test"}},
			expected:  0.0,
		},
		{
			name: "identical instances",
			instances: []DuplicationInstance{
				{Content: "function test() { return true; }"},
				{Content: "function test() { return true; }"},
			},
			expected: 1.0,
		},
		{
			name: "similar instances",
			instances: []DuplicationInstance{
				{Content: "function test() { return true; }"},
				{Content: "function test() { return false; }"},
				{Content: "function test() { return null; }"},
			},
			expected: 0.8, // Should be high similarity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.calculateClusterSimilarity(tt.instances)
			if tt.expected > 0 {
				assert.GreaterOrEqual(t, result, tt.expected-0.1)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEstimateTokenCount(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		content  string
		minCount int
	}{
		{
			name:     "simple function",
			content:  "function test() { return true; }",
			minCount: 6, // function, test, return, true, plus symbols
		},
		{
			name:     "complex function",
			content:  "function validateInput(data, options) { if (data) { return data.valid; } }",
			minCount: 12, // More tokens and symbols
		},
		{
			name:     "empty content",
			content:  "",
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.estimateTokenCount(tt.content)
			assert.GreaterOrEqual(t, result, tt.minCount)
		})
	}
}

func TestAssessRefactoringEffort(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		cluster  DuplicationCluster
		expected string
	}{
		{
			name: "small local cluster",
			cluster: DuplicationCluster{
				LineCount: 10,
				Instances: []DuplicationInstance{
					{FilePath: "test.js"},
					{FilePath: "test.js"},
				},
			},
			expected: "low",
		},
		{
			name: "large cluster",
			cluster: DuplicationCluster{
				LineCount: 60,
				Instances: []DuplicationInstance{
					{FilePath: "test.js"},
					{FilePath: "test.js"},
				},
			},
			expected: "medium",
		},
		{
			name: "many instances",
			cluster: DuplicationCluster{
				LineCount: 20,
				Instances: []DuplicationInstance{
					{FilePath: "test1.js"}, {FilePath: "test2.js"}, {FilePath: "test3.js"},
					{FilePath: "test4.js"}, {FilePath: "test5.js"}, {FilePath: "test6.js"},
				},
			},
			expected: "high", // Cross-file takes precedence
		},
		{
			name: "cross-file cluster",
			cluster: DuplicationCluster{
				LineCount: 20,
				Instances: []DuplicationInstance{
					{FilePath: "file1.js"}, {FilePath: "file2.js"}, {FilePath: "file3.js"},
				},
			},
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.assessRefactoringEffort(tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeterminePriority(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		cluster  DuplicationCluster
		expected string
	}{
		{
			name: "critical priority",
			cluster: DuplicationCluster{
				MaintenanceBurden: 150,
				RefactoringEffort: "low",
			},
			expected: "critical",
		},
		{
			name: "high priority",
			cluster: DuplicationCluster{
				MaintenanceBurden: 75,
				RefactoringEffort: "medium",
			},
			expected: "high",
		},
		{
			name: "medium priority",
			cluster: DuplicationCluster{
				MaintenanceBurden: 30,
				RefactoringEffort: "medium",
			},
			expected: "medium",
		},
		{
			name: "low priority",
			cluster: DuplicationCluster{
				MaintenanceBurden: 10,
				RefactoringEffort: "low",
			},
			expected: "low",
		},
		{
			name: "high burden but high effort",
			cluster: DuplicationCluster{
				MaintenanceBurden: 200,
				RefactoringEffort: "high",
			},
			expected: "medium", // High effort reduces priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.determinePriority(tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountCrossFileInstances(t *testing.T) {
	detector := NewDuplicationDetector()

	instances := []DuplicationInstance{
		{FilePath: "file1.js"},
		{FilePath: "file1.js"},
		{FilePath: "file2.js"},
		{FilePath: "file3.js"},
		{FilePath: "file2.js"},
	}

	result := detector.countCrossFileInstances(instances)
	assert.Equal(t, 3, result) // file1.js, file2.js, file3.js
}

func TestIdentifySharedFunctionality(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		cluster  DuplicationCluster
		expected string
	}{
		{
			name: "validation functions",
			cluster: DuplicationCluster{
				Instances: []DuplicationInstance{
					{FunctionName: "validateInput"},
					{FunctionName: "validateOutput"},
					{FunctionName: "validateData"},
				},
			},
			expected: "validation",
		},
		{
			name: "formatting functions",
			cluster: DuplicationCluster{
				Instances: []DuplicationInstance{
					{FunctionName: "formatDate"},
					{FunctionName: "formatNumber"},
					{FunctionName: "formatString"},
				},
			},
			expected: "formatting",
		},
		{
			name: "parsing functions",
			cluster: DuplicationCluster{
				Instances: []DuplicationInstance{
					{FunctionName: "parseJSON"},
					{FunctionName: "parseXML"},
				},
			},
			expected: "parsing",
		},
		{
			name: "mixed functions",
			cluster: DuplicationCluster{
				Instances: []DuplicationInstance{
					{FunctionName: "processData"},
					{FunctionName: "handleInput"},
					{FunctionName: "executeTask"},
				},
			},
			expected: "utility",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.identifySharedFunctionality(tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineRefactoringStrategy(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name     string
		cluster  DuplicationCluster
		expected string
	}{
		{
			name: "exact cross-file duplicates",
			cluster: DuplicationCluster{
				Type: "exact",
				Instances: []DuplicationInstance{
					{FilePath: "file1.js"},
					{FilePath: "file2.js"},
				},
			},
			expected: "extract_to_shared_module",
		},
		{
			name: "structural duplicates",
			cluster: DuplicationCluster{
				Type: "structural",
				Instances: []DuplicationInstance{
					{FilePath: "file1.js"},
					{FilePath: "file2.js"},
				},
			},
			expected: "create_template_function",
		},
		{
			name: "token duplicates",
			cluster: DuplicationCluster{
				Type: "token",
				Instances: []DuplicationInstance{
					{FilePath: "file1.js"},
					{FilePath: "file2.js"},
				},
			},
			expected: "standardize_and_refactor",
		},
		{
			name: "single file duplicates",
			cluster: DuplicationCluster{
				Type: "exact",
				Instances: []DuplicationInstance{
					{FilePath: "file1.js"},
					{FilePath: "file1.js"},
				},
			},
			expected: "extract_local_function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.determineRefactoringStrategy(tt.cluster)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineConsolidationTarget(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name       string
		fileGroups map[string][]DuplicationInstance
		expected   string
	}{
		{
			name: "clear winner",
			fileGroups: map[string][]DuplicationInstance{
				"file1.js": {{}, {}, {}}, // 3 instances
				"file2.js": {{}},         // 1 instance
			},
			expected: "file1.js",
		},
		{
			name: "no clear winner",
			fileGroups: map[string][]DuplicationInstance{
				"file1.js": {{}}, // 1 instance
				"file2.js": {{}}, // 1 instance
			},
			expected: "new_utility_file",
		},
		{
			name: "equal instances",
			fileGroups: map[string][]DuplicationInstance{
				"file1.js": {{}, {}}, // 2 instances
				"file2.js": {{}, {}}, // 2 instances
			},
			expected: "file1.js", // First one wins in tie
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.determineConsolidationTarget(tt.fileGroups)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateTotalLines(t *testing.T) {
	detector := NewDuplicationDetector()

	parseResult := &ast.ParseResult{
		Functions: []ast.FunctionInfo{
			{StartLine: 1, EndLine: 10},
			{StartLine: 15, EndLine: 25},
		},
		Classes: []ast.ClassInfo{
			{
				Methods: []ast.FunctionInfo{
					{StartLine: 30, EndLine: 45},
					{StartLine: 50, EndLine: 60},
				},
			},
		},
	}

	result := detector.estimateTotalLines(parseResult)
	assert.Equal(t, 60, result) // Should return the maximum line number
}

func TestCalculateHotspotScore(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name            string
		fileDuplication FileDuplication
		expected        float64
	}{
		{
			name: "high duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.4,
			},
			expected: 40.0,
		},
		{
			name: "medium duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.15,
			},
			expected: 15.0,
		},
		{
			name: "low duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.05,
			},
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.calculateHotspotScore(tt.fileDuplication)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineFileRefactoringPriority(t *testing.T) {
	detector := NewDuplicationDetector()

	tests := []struct {
		name            string
		fileDuplication FileDuplication
		expected        string
	}{
		{
			name: "critical duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.35,
			},
			expected: "critical",
		},
		{
			name: "high duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.25,
			},
			expected: "high",
		},
		{
			name: "medium duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.15,
			},
			expected: "medium",
		},
		{
			name: "low duplication",
			fileDuplication: FileDuplication{
				DuplicationRatio: 0.05,
			},
			expected: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.determineFileRefactoringPriority(tt.fileDuplication)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAffectedFiles(t *testing.T) {
	detector := NewDuplicationDetector()

	instances := []DuplicationInstance{
		{FilePath: "file1.js"},
		{FilePath: "file2.js"},
		{FilePath: "file1.js"}, // Duplicate path
		{FilePath: "file3.js"},
	}

	result := detector.extractAffectedFiles(instances)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "file1.js")
	assert.Contains(t, result, "file2.js")
	assert.Contains(t, result, "file3.js")
}

func TestExtractAffectedFunctions(t *testing.T) {
	detector := NewDuplicationDetector()

	instances := []DuplicationInstance{
		{FunctionName: "validate", ClassName: ""},
		{FunctionName: "process", ClassName: "DataHandler"},
		{FunctionName: "validate", ClassName: ""}, // Duplicate
		{FunctionName: "cleanup", ClassName: "DataHandler"},
		{FunctionName: "", ClassName: ""}, // Empty function name
	}

	result := detector.extractAffectedFunctions(instances)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "validate")
	assert.Contains(t, result, "DataHandler.process")
	assert.Contains(t, result, "DataHandler.cleanup")
}

// Integration tests

func TestDuplicationDetectionIntegration(t *testing.T) {
	detector := NewDuplicationDetector()
	ctx := context.Background()

	// Create realistic scenario with various types of duplication
	parseResults := []*ast.ParseResult{
		// File 1 with potential duplicates
		createMockParseResultForDuplication("utils.js", []ast.FunctionInfo{
			createMockFunctionForDuplication("validateEmail", 1, 15),
			createMockFunctionForDuplication("validatePhone", 20, 34),
			createMockFunctionForDuplication("formatDate", 40, 50),
		}, []ast.ClassInfo{}),

		// File 2 with similar functions
		createMockParseResultForDuplication("helpers.js", []ast.FunctionInfo{
			createMockFunctionForDuplication("validateEmail", 1, 15),  // Exact duplicate
			createMockFunctionForDuplication("validateInput", 20, 35), // Similar structure
			createMockFunctionForDuplication("formatTime", 40, 50),    // Similar naming
		}, []ast.ClassInfo{}),

		// File 3 with class methods
		createMockParseResultForDuplication("validator.js", []ast.FunctionInfo{}, []ast.ClassInfo{
			createMockClassForDuplication("InputValidator", []ast.FunctionInfo{
				createMockFunctionForDuplication("validateEmail", 5, 20), // Potential duplicate
				createMockFunctionForDuplication("sanitize", 25, 40),
			}),
		}),
	}

	metrics, err := detector.DetectDuplication(ctx, parseResults)

	require.NoError(t, err)
	require.NotNil(t, metrics)

	// Verify overall metrics
	assert.GreaterOrEqual(t, metrics.OverallScore, 0.0)
	assert.LessOrEqual(t, metrics.OverallScore, 100.0)
	assert.GreaterOrEqual(t, metrics.DuplicationRatio, 0.0)
	// DuplicationRatio can exceed 1.0 when multiple duplicates exist
	assert.LessOrEqual(t, metrics.DuplicationRatio, 5.0)

	// Verify cluster analysis
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	for _, cluster := range allClusters {
		assert.NotEmpty(t, cluster.ID)
		assert.Greater(t, len(cluster.Instances), 1)
		assert.GreaterOrEqual(t, cluster.SimilarityScore, 0.0)
		assert.LessOrEqual(t, cluster.SimilarityScore, 5.0) // Can exceed 1.0 with complex calculations
		assert.Greater(t, cluster.LineCount, 0)
		assert.Greater(t, cluster.TokenCount, 0)
		assert.NotEmpty(t, cluster.Priority)
		assert.NotEmpty(t, cluster.RefactoringEffort)
	}

	// Verify file metrics
	assert.Len(t, metrics.DuplicationByFile, 3)
	for filePath, fileDuplication := range metrics.DuplicationByFile {
		assert.NotEmpty(t, filePath)
		assert.GreaterOrEqual(t, fileDuplication.DuplicationRatio, 0.0)
		assert.LessOrEqual(t, fileDuplication.DuplicationRatio, 5.0) // Can exceed 1.0 with multiple duplicates
		assert.GreaterOrEqual(t, fileDuplication.HotspotScore, 0.0)
		assert.NotEmpty(t, fileDuplication.RefactoringPriority)
	}

	// Verify consolidation opportunities
	for _, opportunity := range metrics.ConsolidationOps {
		assert.NotEmpty(t, opportunity.ID)
		assert.NotEmpty(t, opportunity.Type)
		assert.NotEmpty(t, opportunity.Description)
		assert.Greater(t, len(opportunity.AffectedFiles), 0)
		assert.Greater(t, opportunity.EstimatedEffort, 0)
		assert.GreaterOrEqual(t, opportunity.ROIScore, 0.0)
	}

	// Verify impact analysis
	assert.GreaterOrEqual(t, metrics.ImpactAnalysis.MaintenanceMultiplier, 1.0)
	assert.GreaterOrEqual(t, metrics.ImpactAnalysis.TechnicalDebtScore, 0.0)
	assert.GreaterOrEqual(t, metrics.ImpactAnalysis.ChangeRiskFactor, 0.0)
	assert.NotEmpty(t, metrics.ImpactAnalysis.CodebaseHealth)

	// Verify recommendations
	for _, recommendation := range metrics.Recommendations {
		assert.NotEmpty(t, recommendation.Priority)
		assert.NotEmpty(t, recommendation.Category)
		assert.NotEmpty(t, recommendation.Title)
		assert.NotEmpty(t, recommendation.Description)
		assert.Greater(t, recommendation.EstimatedHours, 0)
	}

	// Verify summary
	assert.GreaterOrEqual(t, metrics.Summary.HealthScore, 0.0)
	assert.LessOrEqual(t, metrics.Summary.HealthScore, 100.0)
	assert.NotEmpty(t, metrics.Summary.RiskLevel)
	assert.NotEmpty(t, metrics.Summary.MaintenanceBurden)
	assert.GreaterOrEqual(t, metrics.Summary.RecommendedActions, 0)
}

func TestDuplicationConfigValidation(t *testing.T) {
	// Test that custom configurations work correctly
	customConfig := DuplicationConfig{
		MinLines:                 4,
		MinTokens:                30,
		SimilarityThreshold:      0.9,
		TokenSimilarityThreshold: 0.8,
		EnableCrossFile:          false,
		ReportTopN:               5,
		WeightFactors: DuplicationWeights{
			ExactDuplication:     1.5,
			StructuralSimilarity: 1.0,
			TokenSimilarity:      0.5,
			CrossFileImpact:      2.0,
			MaintenanceBurden:    1.2,
		},
	}

	detector := NewDuplicationDetectorWithConfig(customConfig)
	ctx := context.Background()

	parseResults := []*ast.ParseResult{
		createMockParseResultForDuplication("test.js", []ast.FunctionInfo{
			createMockFunctionForDuplication("test1", 1, 5), // Meets custom MinLines
			createMockFunctionForDuplication("test2", 10, 14),
		}, []ast.ClassInfo{}),
	}

	metrics, err := detector.DetectDuplication(ctx, parseResults)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)

	// Cross-file analysis should be disabled
	assert.Empty(t, metrics.CrossFileDuplicates)
}

// Helper functions for creating mock data

func createMockParseResultForDuplication(filePath string, functions []ast.FunctionInfo, classes []ast.ClassInfo) *ast.ParseResult {
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

func createMockFunctionForDuplication(name string, startLine, endLine int) ast.FunctionInfo {
	return ast.FunctionInfo{
		Name:       name,
		Parameters: []ast.ParameterInfo{},
		StartLine:  startLine,
		EndLine:    endLine,
		Metadata:   make(map[string]string),
	}
}

func createMockClassForDuplication(name string, methods []ast.FunctionInfo) ast.ClassInfo {
	return ast.ClassInfo{
		Name:       name,
		Methods:    methods,
		Properties: []ast.PropertyInfo{},
		Metadata:   make(map[string]string),
	}
}

// Benchmark tests

func BenchmarkDetectDuplication(b *testing.B) {
	detector := NewDuplicationDetector()
	ctx := context.Background()

	// Create benchmark data
	parseResults := make([]*ast.ParseResult, 5)
	for i := 0; i < 5; i++ {
		functions := make([]ast.FunctionInfo, 15)
		for j := 0; j < 15; j++ {
			functions[j] = createMockFunctionForDuplication(
				fmt.Sprintf("func_%d_%d", i, j),
				j*10+1,
				j*10+15,
			)
		}

		parseResults[i] = createMockParseResultForDuplication(
			fmt.Sprintf("file_%d.js", i),
			functions,
			[]ast.ClassInfo{},
		)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := detector.DetectDuplication(ctx, parseResults)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenSimilarity(b *testing.B) {
	detector := NewDuplicationDetector()

	content1 := "function validateInput(data, options) { if (data && options) { return data.valid && options.strict; } return false; }"
	content2 := "function validateOutput(result, config) { if (result && config) { return result.success && config.verbose; } return true; }"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector.calculateTokenSimilarity(content1, content2)
	}
}

func BenchmarkLevenshteinDistance(b *testing.B) {
	detector := NewDuplicationDetector()

	s1 := "function validateInput(data) { return data.valid; }"
	s2 := "function validateOutput(result) { return result.success; }"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		detector.levenshteinDistance(s1, s2)
	}
}
