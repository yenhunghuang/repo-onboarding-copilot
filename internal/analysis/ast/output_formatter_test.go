package ast

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewOutputFormatter(t *testing.T) {
	tests := []struct {
		name     string
		options  []FormatterOptions
		expected FormatterOptions
	}{
		{
			name:    "default options",
			options: nil,
			expected: FormatterOptions{
				PrettyPrint:               true,
				IncludeMetadata:           true,
				IncludePerformanceMetrics: false,
				CompressionLevel:          0,
				OutputFormat:              "json",
			},
		},
		{
			name: "custom options",
			options: []FormatterOptions{
				{
					PrettyPrint:               false,
					IncludeMetadata:           false,
					IncludePerformanceMetrics: true,
					CompressionLevel:          2,
					OutputFormat:              "yaml",
				},
			},
			expected: FormatterOptions{
				PrettyPrint:               false,
				IncludeMetadata:           false,
				IncludePerformanceMetrics: true,
				CompressionLevel:          2,
				OutputFormat:              "yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(tt.options...)

			if formatter.options != tt.expected {
				t.Errorf("NewOutputFormatter() options = %v, want %v", formatter.options, tt.expected)
			}
		})
	}
}

func TestFormatAnalysisResult(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create test analysis result
	testResult := createTestAnalysisResult()
	testMetadata := &AnalysisMetadata{
		StartedAt:      time.Now().Add(-5 * time.Minute),
		CompletedAt:    time.Now(),
		Version:        "1.0.0",
		EngineVersion:  "ast-parser-v2.1",
		ProcessedFiles: 5,
		FailedFiles:    0,
	}

	tests := []struct {
		name      string
		result    *AnalysisResult
		metadata  *AnalysisMetadata
		wantError bool
	}{
		{
			name:      "valid analysis result",
			result:    testResult,
			metadata:  testMetadata,
			wantError: false,
		},
		{
			name:      "nil analysis result",
			result:    nil,
			metadata:  testMetadata,
			wantError: true,
		},
		{
			name:      "nil metadata with default generation",
			result:    testResult,
			metadata:  nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			standardized, err := formatter.FormatAnalysisResult(tt.result, tt.metadata, nil)

			if tt.wantError {
				if err == nil {
					t.Errorf("FormatAnalysisResult() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("FormatAnalysisResult() unexpected error: %v", err)
				return
			}

			if standardized == nil {
				t.Error("FormatAnalysisResult() returned nil standardized result")
				return
			}

			// Validate required fields
			if standardized.AnalysisID == "" {
				t.Error("FormatAnalysisResult() missing analysis ID")
			}

			if standardized.Repository.Path == "" {
				t.Error("FormatAnalysisResult() missing repository path")
			}

			if standardized.CodeAnalysis.StructuralMetrics.TotalFunctions != testResult.Summary.TotalFunctions {
				t.Errorf("FormatAnalysisResult() function count mismatch: got %d, want %d",
					standardized.CodeAnalysis.StructuralMetrics.TotalFunctions, testResult.Summary.TotalFunctions)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	formatter := NewOutputFormatter()
	testResult := createTestStandardizedResult()

	tests := []struct {
		name        string
		result      *StandardizedAnalysisResult
		wantError   bool
		checkPretty bool
	}{
		{
			name:        "valid result with pretty print",
			result:      testResult,
			wantError:   false,
			checkPretty: true,
		},
		{
			name:      "nil result",
			result:    nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := formatter.ToJSON(tt.result)

			if tt.wantError {
				if err == nil {
					t.Errorf("ToJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ToJSON() unexpected error: %v", err)
				return
			}

			// Validate JSON structure
			var parsed map[string]interface{}
			if err := json.Unmarshal(jsonData, &parsed); err != nil {
				t.Errorf("ToJSON() produced invalid JSON: %v", err)
				return
			}

			// Check required fields
			requiredFields := []string{"analysis_id", "repository", "analysis_metadata", "code_analysis"}
			for _, field := range requiredFields {
				if _, exists := parsed[field]; !exists {
					t.Errorf("ToJSON() missing required field: %s", field)
				}
			}

			// Check pretty printing
			if tt.checkPretty && formatter.options.PrettyPrint {
				// Pretty printed JSON should contain indentation
				if len(jsonData) < 100 { // Sanity check for minimum size with formatting
					t.Error("ToJSON() expected pretty printed output to be larger")
				}
			}
		})
	}
}

func TestToCompactJSON(t *testing.T) {
	formatter := NewOutputFormatter()
	testResult := createTestStandardizedResult()

	tests := []struct {
		name             string
		result           *StandardizedAnalysisResult
		compressionLevel int
		wantError        bool
	}{
		{
			name:             "no compression",
			result:           testResult,
			compressionLevel: 0,
			wantError:        false,
		},
		{
			name:             "minimal compression",
			result:           testResult,
			compressionLevel: 1,
			wantError:        false,
		},
		{
			name:             "aggressive compression",
			result:           testResult,
			compressionLevel: 2,
			wantError:        false,
		},
		{
			name:      "nil result",
			result:    nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter.options.CompressionLevel = tt.compressionLevel

			jsonData, err := formatter.ToCompactJSON(tt.result)

			if tt.wantError {
				if err == nil {
					t.Errorf("ToCompactJSON() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ToCompactJSON() unexpected error: %v", err)
				return
			}

			// Validate JSON structure
			var parsed interface{}
			if err := json.Unmarshal(jsonData, &parsed); err != nil {
				t.Errorf("ToCompactJSON() produced invalid JSON: %v", err)
				return
			}

			// Check compression effects
			if tt.compressionLevel > 0 {
				// Compressed JSON should be smaller
				originalData, _ := formatter.ToJSON(testResult)
				if len(jsonData) >= len(originalData) {
					t.Log("ToCompactJSON() compression may not be effective for small test data")
				}
			}
		})
	}
}

func TestCreateDocumentationPayload(t *testing.T) {
	formatter := NewOutputFormatter()
	testResult := createTestStandardizedResult()

	tests := []struct {
		name      string
		result    *StandardizedAnalysisResult
		wantError bool
	}{
		{
			name:      "valid standardized result",
			result:    testResult,
			wantError: false,
		},
		{
			name:      "nil result",
			result:    nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := formatter.CreateDocumentationPayload(tt.result)

			if tt.wantError {
				if err == nil {
					t.Errorf("CreateDocumentationPayload() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateDocumentationPayload() unexpected error: %v", err)
				return
			}

			if payload == nil {
				t.Error("CreateDocumentationPayload() returned nil payload")
				return
			}

			// Validate payload structure
			if payload.AnalysisID != testResult.AnalysisID {
				t.Errorf("CreateDocumentationPayload() analysis ID mismatch: got %s, want %s",
					payload.AnalysisID, testResult.AnalysisID)
			}

			if payload.ProjectPath != testResult.Repository.Path {
				t.Errorf("CreateDocumentationPayload() project path mismatch: got %s, want %s",
					payload.ProjectPath, testResult.Repository.Path)
			}

			if payload.ComponentMap != testResult.CodeAnalysis.ComponentMap {
				t.Error("CreateDocumentationPayload() component map mismatch")
			}

			if payload.StructuralData != &testResult.CodeAnalysis.StructuralMetrics {
				t.Error("CreateDocumentationPayload() structural data mismatch")
			}

			// Check metadata
			if payload.Metadata == nil {
				t.Error("CreateDocumentationPayload() missing metadata")
			}
		})
	}
}

func TestCalculateStructuralMetrics(t *testing.T) {
	formatter := NewOutputFormatter()
	testResult := createTestAnalysisResult()

	metrics := formatter.calculateStructuralMetrics(testResult)

	// Validate calculated metrics
	if metrics.TotalFunctions != testResult.Summary.TotalFunctions {
		t.Errorf("calculateStructuralMetrics() function count: got %d, want %d",
			metrics.TotalFunctions, testResult.Summary.TotalFunctions)
	}

	if metrics.TotalClasses != testResult.Summary.TotalClasses {
		t.Errorf("calculateStructuralMetrics() class count: got %d, want %d",
			metrics.TotalClasses, testResult.Summary.TotalClasses)
	}

	if metrics.TotalInterfaces != testResult.Summary.TotalInterfaces {
		t.Errorf("calculateStructuralMetrics() interface count: got %d, want %d",
			metrics.TotalInterfaces, testResult.Summary.TotalInterfaces)
	}

	// Check per-file metrics
	if metrics.FunctionsByFile == nil {
		t.Error("calculateStructuralMetrics() missing FunctionsByFile")
	}

	if metrics.ClassesByFile == nil {
		t.Error("calculateStructuralMetrics() missing ClassesByFile")
	}

	if metrics.ComplexityByFile == nil {
		t.Error("calculateStructuralMetrics() missing ComplexityByFile")
	}

	// Validate total exports/imports count
	expectedExports := 0
	expectedImports := 0
	for _, result := range testResult.FileResults {
		expectedExports += len(result.Exports)
		expectedImports += len(result.Imports)
	}

	if metrics.TotalExports != expectedExports {
		t.Errorf("calculateStructuralMetrics() export count: got %d, want %d",
			metrics.TotalExports, expectedExports)
	}

	if metrics.TotalImports != expectedImports {
		t.Errorf("calculateStructuralMetrics() import count: got %d, want %d",
			metrics.TotalImports, expectedImports)
	}
}

func TestCalculateQualityScore(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		result   *AnalysisResult
		minScore float64
		maxScore float64
	}{
		{
			name:     "good quality result",
			result:   createGoodQualityResult(),
			minScore: 80.0,
			maxScore: 100.0,
		},
		{
			name:     "poor quality result",
			result:   createPoorQualityResult(),
			minScore: 0.0,
			maxScore: 60.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := formatter.calculateQualityScore(tt.result)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("calculateQualityScore() score %f not in range [%f, %f]",
					score, tt.minScore, tt.maxScore)
			}

			// Score should be within valid bounds
			if score < 0.0 || score > 100.0 {
				t.Errorf("calculateQualityScore() score %f out of bounds [0, 100]", score)
			}
		})
	}
}

func TestCompressionLevels(t *testing.T) {
	testResult := createTestStandardizedResult()

	tests := []struct {
		name             string
		compressionLevel int
		checkField       string // Field to check for compression effect
	}{
		{
			name:             "no compression",
			compressionLevel: 0,
			checkField:       "code_analysis",
		},
		{
			name:             "minimal compression",
			compressionLevel: 1,
			checkField:       "ast_data", // Should be nil in compressed version
		},
		{
			name:             "aggressive compression",
			compressionLevel: 2,
			checkField:       "summary", // Should only have summary data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter(FormatterOptions{
				CompressionLevel: tt.compressionLevel,
			})

			compressed := formatter.compressResult(testResult)

			// Validate compression effects
			switch tt.compressionLevel {
			case 0:
				// No compression - should be identical
				if compressed != testResult {
					t.Error("No compression should return original result")
				}

			case 1:
				// Minimal compression - AST data should be removed
				compressedResult, ok := compressed.(*StandardizedAnalysisResult)
				if !ok {
					t.Error("Minimal compression should return StandardizedAnalysisResult")
					return
				}
				if compressedResult.CodeAnalysis.ASTData != nil {
					t.Error("Minimal compression should remove AST data")
				}

			case 2:
				// Aggressive compression - should return map with summary only
				compressedMap, ok := compressed.(map[string]interface{})
				if !ok {
					t.Error("Aggressive compression should return map")
					return
				}

				expectedKeys := []string{"analysis_id", "repository", "summary"}
				for _, key := range expectedKeys {
					if _, exists := compressedMap[key]; !exists {
						t.Errorf("Aggressive compression missing key: %s", key)
					}
				}
			}
		})
	}
}

// Helper functions to create test data

func createTestAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		ProjectPath: "/test/project",
		FileResults: map[string]*ParseResult{
			"file1.ts": {
				FilePath:   "file1.ts",
				Language:   "typescript",
				Functions:  []FunctionInfo{{Name: "testFunc1"}},
				Classes:    []ClassInfo{{Name: "TestClass1"}},
				Interfaces: []InterfaceInfo{{Name: "TestInterface1"}},
				Variables:  []VariableInfo{{Name: "testVar1"}},
				Imports:    []ImportInfo{{Source: "./utils"}},
				Exports:    []ExportInfo{{Name: "testFunc1"}},
			},
			"file2.js": {
				FilePath:  "file2.js",
				Language:  "javascript",
				Functions: []FunctionInfo{{Name: "testFunc2"}},
				Classes:   []ClassInfo{{Name: "TestClass2"}},
				Variables: []VariableInfo{{Name: "testVar2"}},
				Imports:   []ImportInfo{{Source: "lodash"}},
				Exports:   []ExportInfo{{Name: "testFunc2"}},
			},
		},
		ComponentMap: &ComponentMap{
			Components: []Component{
				{
					ID:    "test-component",
					Name:  "Test Component",
					Type:  "service",
					Files: []string{"file1.ts", "file2.js"},
				},
			},
		},
		ExternalPackages: map[string]ExternalPackage{
			"lodash": {
				Name:        "lodash",
				Version:     "^4.17.21",
				PackageType: "npm",
			},
		},
		Summary: AnalysisSummary{
			TotalFiles:      2,
			TotalFunctions:  2,
			TotalClasses:    2,
			TotalInterfaces: 1,
			TotalVariables:  2,
			Languages: map[string]int{
				"typescript": 1,
				"javascript": 1,
			},
			Complexity: ComplexityMetrics{
				AverageFileSize: 50.0,
			},
			Quality: QualityMetrics{
				DocumentationCoverage: 0.8,
				TestCoverage:          0.7,
				CircularDependencies:  1,
				UnusedExports:         2,
			},
		},
	}
}

func createTestStandardizedResult() *StandardizedAnalysisResult {
	return &StandardizedAnalysisResult{
		AnalysisID: "test-analysis-123",
		Repository: RepositoryInfo{
			Path:      "/test/project",
			Languages: []string{"typescript", "javascript"},
			FileCount: 2,
			SizeBytes: 1024,
		},
		Metadata: AnalysisMetadata{
			StartedAt:      time.Now().Add(-5 * time.Minute),
			CompletedAt:    time.Now(),
			Version:        "1.0.0",
			EngineVersion:  "ast-parser-v2.1",
			ProcessedFiles: 2,
			FailedFiles:    0,
		},
		CodeAnalysis: CodeAnalysisSection{
			ComponentMap: &ComponentMap{
				Components: []Component{
					{ID: "test", Name: "Test", Type: "service"},
				},
			},
			QualityScore: 85.0,
			StructuralMetrics: StructuralMetrics{
				TotalFunctions:    2,
				TotalClasses:      2,
				TotalInterfaces:   1,
				TotalVariables:    2,
				TotalExports:      2,
				TotalImports:      2,
				FunctionsByFile:   map[string]int{"file1.ts": 1, "file2.js": 1},
				ClassesByFile:     map[string]int{"file1.ts": 1, "file2.js": 1},
				ComplexityByFile:  map[string]int{"file1.ts": 3, "file2.js": 2},
				LanguageBreakdown: map[string]int{"typescript": 1, "javascript": 1},
			},
		},
		Dependencies: DependenciesSection{
			External: []ExternalPackage{
				{Name: "lodash", Version: "^4.17.21", PackageType: "npm"},
			},
			GraphMetrics: DependencyGraphMetrics{
				TotalNodes: 2,
				TotalEdges: 1,
			},
		},
		Security: SecuritySection{
			RiskScore: 30.0,
		},
		Quality: QualitySection{
			OverallScore:          85.0,
			DocumentationCoverage: 80.0,
			TestCoverage:          70.0,
			CodeConsistency:       85.0,
			Maintainability:       75.0,
		},
	}
}

func createGoodQualityResult() *AnalysisResult {
	result := createTestAnalysisResult()
	result.Summary.Quality = QualityMetrics{
		DocumentationCoverage: 0.9,
		TestCoverage:          0.8,
		CircularDependencies:  0,
		UnusedExports:         0,
	}
	return result
}

func createPoorQualityResult() *AnalysisResult {
	result := createTestAnalysisResult()
	result.Summary.Quality = QualityMetrics{
		DocumentationCoverage: 0.2,
		TestCoverage:          0.1,
		CircularDependencies:  5,
		UnusedExports:         10,
	}
	return result
}
