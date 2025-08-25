package ast

import (
	"testing"
	"time"
)

func TestNewDocumentationIntegrator(t *testing.T) {
	tests := []struct {
		name     string
		config   IntegratorConfig
		expected IntegratorConfig
	}{
		{
			name:   "default config",
			config: IntegratorConfig{},
			expected: IntegratorConfig{
				OutputDirectory: "./docs/generated",
				TemplateVersion: "1.0",
			},
		},
		{
			name: "custom config",
			config: IntegratorConfig{
				EnableRunbookGeneration:    true,
				EnableArchitectureDiagrams: true,
				OutputDirectory:            "./custom/docs",
				TemplateVersion:            "2.0",
			},
			expected: IntegratorConfig{
				EnableRunbookGeneration:    true,
				EnableArchitectureDiagrams: true,
				OutputDirectory:            "./custom/docs",
				TemplateVersion:            "2.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integrator := NewDocumentationIntegrator(tt.config)

			if integrator == nil {
				t.Error("NewDocumentationIntegrator() returned nil")
				return
			}

			if integrator.config.OutputDirectory != tt.expected.OutputDirectory {
				t.Errorf("OutputDirectory = %v, want %v", integrator.config.OutputDirectory, tt.expected.OutputDirectory)
			}

			if integrator.config.TemplateVersion != tt.expected.TemplateVersion {
				t.Errorf("TemplateVersion = %v, want %v", integrator.config.TemplateVersion, tt.expected.TemplateVersion)
			}

			if integrator.formatter == nil {
				t.Error("NewDocumentationIntegrator() formatter is nil")
			}
		})
	}
}

func TestPrepareForDocumentation(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	testResult := createTestAnalysisResult()
	testMetadata := &AnalysisMetadata{
		StartedAt:      time.Now().Add(-5 * time.Minute),
		CompletedAt:    time.Now(),
		Version:        "1.0.0",
		ProcessedFiles: 2,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := integrator.PrepareForDocumentation(tt.result, tt.metadata)

			if tt.wantError {
				if err == nil {
					t.Error("PrepareForDocumentation() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("PrepareForDocumentation() unexpected error: %v", err)
				return
			}

			if payload == nil {
				t.Error("PrepareForDocumentation() returned nil payload")
				return
			}

			// Validate payload structure
			if payload.AnalysisID == "" {
				t.Error("PrepareForDocumentation() missing analysis ID")
			}

			if payload.ProjectPath != testResult.ProjectPath {
				t.Errorf("PrepareForDocumentation() project path mismatch: got %s, want %s",
					payload.ProjectPath, testResult.ProjectPath)
			}

			if payload.ComponentMap != testResult.ComponentMap {
				t.Error("PrepareForDocumentation() component map mismatch")
			}

			if payload.Metadata == nil {
				t.Error("PrepareForDocumentation() missing metadata")
			}
		})
	}
}

func TestGenerateRunbookStructure(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	testPayload := createTestDocumentationPayload()

	tests := []struct {
		name      string
		payload   *DocumentationPayload
		wantError bool
	}{
		{
			name:      "valid payload",
			payload:   testPayload,
			wantError: false,
		},
		{
			name:      "nil payload",
			payload:   nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runbook, err := integrator.GenerateRunbookStructure(tt.payload)

			if tt.wantError {
				if err == nil {
					t.Error("GenerateRunbookStructure() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateRunbookStructure() unexpected error: %v", err)
				return
			}

			if runbook == nil {
				t.Error("GenerateRunbookStructure() returned nil runbook")
				return
			}

			// Validate runbook structure
			if runbook.Title == "" {
				t.Error("GenerateRunbookStructure() missing title")
			}

			if runbook.Description == "" {
				t.Error("GenerateRunbookStructure() missing description")
			}

			if len(runbook.SetupSteps) == 0 {
				t.Error("GenerateRunbookStructure() missing setup steps")
			}

			if len(runbook.ValidationSteps) == 0 {
				t.Error("GenerateRunbookStructure() missing validation steps")
			}

			if len(runbook.TroubleshootingSteps) == 0 {
				t.Error("GenerateRunbookStructure() missing troubleshooting steps")
			}

			if len(runbook.Scripts) == 0 {
				t.Error("GenerateRunbookStructure() missing scripts")
			}

			if runbook.EstimatedTime == "" {
				t.Error("GenerateRunbookStructure() missing estimated time")
			}

			// Validate metadata
			if runbook.Metadata == nil {
				t.Error("GenerateRunbookStructure() missing metadata")
			} else {
				if runbook.Metadata["generated_from"] != "ast-analysis" {
					t.Error("GenerateRunbookStructure() incorrect generated_from metadata")
				}
			}
		})
	}
}

func TestGenerateArchitectureDiagrams(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	testPayload := createTestDocumentationPayload()

	tests := []struct {
		name        string
		payload     *DocumentationPayload
		wantError   bool
		minDiagrams int
	}{
		{
			name:        "valid payload with components",
			payload:     testPayload,
			wantError:   false,
			minDiagrams: 2, // Component and dependency diagrams at minimum
		},
		{
			name:      "nil payload",
			payload:   nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagrams, err := integrator.GenerateArchitectureDiagrams(tt.payload)

			if tt.wantError {
				if err == nil {
					t.Error("GenerateArchitectureDiagrams() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateArchitectureDiagrams() unexpected error: %v", err)
				return
			}

			if len(diagrams) < tt.minDiagrams {
				t.Errorf("GenerateArchitectureDiagrams() got %d diagrams, want at least %d",
					len(diagrams), tt.minDiagrams)
				return
			}

			// Validate each diagram
			for i, diagram := range diagrams {
				if diagram.Type == "" {
					t.Errorf("Diagram %d missing type", i)
				}

				if diagram.Title == "" {
					t.Errorf("Diagram %d missing title", i)
				}

				if diagram.Format == "" {
					t.Errorf("Diagram %d missing format", i)
				}

				// Validate expected diagram types
				validTypes := []string{"component", "dependency", "layered"}
				typeValid := false
				for _, validType := range validTypes {
					if diagram.Type == validType {
						typeValid = true
						break
					}
				}
				if !typeValid {
					t.Errorf("Diagram %d has invalid type: %s", i, diagram.Type)
				}
			}
		})
	}
}

func TestGenerateLearningRoadmap(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	testPayload := createTestDocumentationPayload()

	tests := []struct {
		name      string
		payload   *DocumentationPayload
		wantError bool
	}{
		{
			name:      "valid payload",
			payload:   testPayload,
			wantError: false,
		},
		{
			name:      "nil payload",
			payload:   nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roadmap, err := integrator.GenerateLearningRoadmap(tt.payload)

			if tt.wantError {
				if err == nil {
					t.Error("GenerateLearningRoadmap() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateLearningRoadmap() unexpected error: %v", err)
				return
			}

			if roadmap == nil {
				t.Error("GenerateLearningRoadmap() returned nil roadmap")
				return
			}

			// Validate roadmap structure
			if roadmap.Title == "" {
				t.Error("GenerateLearningRoadmap() missing title")
			}

			if roadmap.Description == "" {
				t.Error("GenerateLearningRoadmap() missing description")
			}

			if roadmap.TotalDuration == "" {
				t.Error("GenerateLearningRoadmap() missing total duration")
			}

			if len(roadmap.Phases) == 0 {
				t.Error("GenerateLearningRoadmap() missing phases")
			} else {
				// Validate phases
				expectedPhases := 3 // Based on implementation
				if len(roadmap.Phases) != expectedPhases {
					t.Errorf("GenerateLearningRoadmap() got %d phases, want %d",
						len(roadmap.Phases), expectedPhases)
				}

				// Validate each phase
				for i, phase := range roadmap.Phases {
					if phase.ID == "" {
						t.Errorf("Phase %d missing ID", i)
					}
					if phase.Title == "" {
						t.Errorf("Phase %d missing title", i)
					}
					if phase.Duration == "" {
						t.Errorf("Phase %d missing duration", i)
					}
					if len(phase.Goals) == 0 {
						t.Errorf("Phase %d missing goals", i)
					}
				}
			}

			if len(roadmap.Prerequisites) == 0 {
				t.Error("GenerateLearningRoadmap() missing prerequisites")
			}

			if len(roadmap.Resources) == 0 {
				t.Error("GenerateLearningRoadmap() missing resources")
			}

			// Validate metadata
			if roadmap.Metadata == nil {
				t.Error("GenerateLearningRoadmap() missing metadata")
			} else {
				if roadmap.Metadata["generated_from"] != "ast-analysis" {
					t.Error("GenerateLearningRoadmap() incorrect generated_from metadata")
				}
			}
		})
	}
}

func TestGetIntegrationMetadata(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{
		EnableRunbookGeneration:    true,
		EnableArchitectureDiagrams: true,
		EnableLearningRoadmap:      true,
		OutputDirectory:            "./test/docs",
		TemplateVersion:            "test-v1.0",
	})

	testPayload := createTestDocumentationPayload()

	tests := []struct {
		name    string
		payload *DocumentationPayload
	}{
		{
			name:    "valid payload",
			payload: testPayload,
		},
		{
			name:    "nil payload",
			payload: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := integrator.GetIntegrationMetadata(tt.payload)

			if metadata == nil {
				t.Error("GetIntegrationMetadata() returned nil")
				return
			}

			// Validate required fields
			requiredFields := []string{
				"integration_version", "supported_formats", "analysis_summary",
				"quality_indicators", "generation_capabilities", "output_configuration",
			}

			for _, field := range requiredFields {
				if _, exists := metadata[field]; !exists {
					t.Errorf("GetIntegrationMetadata() missing field: %s", field)
				}
			}

			// Validate generation capabilities
			if capabilities, ok := metadata["generation_capabilities"].(map[string]bool); ok {
				if !capabilities["runbook_generation"] {
					t.Error("GetIntegrationMetadata() runbook_generation should be enabled")
				}
				if !capabilities["architecture_diagrams"] {
					t.Error("GetIntegrationMetadata() architecture_diagrams should be enabled")
				}
				if !capabilities["learning_roadmap"] {
					t.Error("GetIntegrationMetadata() learning_roadmap should be enabled")
				}
			} else {
				t.Error("GetIntegrationMetadata() invalid generation_capabilities format")
			}

			// Validate output configuration
			if config, ok := metadata["output_configuration"].(map[string]interface{}); ok {
				if config["output_directory"] != "./test/docs" {
					t.Error("GetIntegrationMetadata() incorrect output_directory")
				}
				if config["template_version"] != "test-v1.0" {
					t.Error("GetIntegrationMetadata() incorrect template_version")
				}
			} else {
				t.Error("GetIntegrationMetadata() invalid output_configuration format")
			}

			// For non-nil payload, validate analysis summary
			if tt.payload != nil {
				if summary, ok := metadata["analysis_summary"].(map[string]interface{}); ok {
					expectedComponentCount := len(tt.payload.ComponentMap.Components)
					if actualCount, ok := summary["component_count"].(int); !ok || actualCount != expectedComponentCount {
						t.Errorf("GetIntegrationMetadata() component_count: got %v, want %d", summary["component_count"], expectedComponentCount)
					}
				}
			}
		})
	}
}

func TestRunbookStepGeneration(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})
	testPayload := createTestDocumentationPayload()

	// Test setup steps generation
	setupSteps := integrator.generateSetupSteps(testPayload)

	if len(setupSteps) == 0 {
		t.Error("generateSetupSteps() returned no steps")
	}

	for i, step := range setupSteps {
		if step.ID == "" {
			t.Errorf("Setup step %d missing ID", i)
		}
		if step.Title == "" {
			t.Errorf("Setup step %d missing title", i)
		}
		if step.Description == "" {
			t.Errorf("Setup step %d missing description", i)
		}
	}

	// Test validation steps generation
	validationSteps := integrator.generateValidationSteps(testPayload)

	if len(validationSteps) == 0 {
		t.Error("generateValidationSteps() returned no steps")
	}

	// Test troubleshooting steps generation
	troubleshootingSteps := integrator.generateTroubleshootingSteps(testPayload)

	if len(troubleshootingSteps) == 0 {
		t.Error("generateTroubleshootingSteps() returned no steps")
	}
}

func TestScriptGeneration(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})
	testPayload := createTestDocumentationPayload()

	scripts := integrator.generateScripts(testPayload)

	if len(scripts) == 0 {
		t.Error("generateScripts() returned no scripts")
	}

	// Check for expected scripts
	expectedScripts := []string{"build", "test"}
	for _, expected := range expectedScripts {
		if script, exists := scripts[expected]; !exists {
			t.Errorf("generateScripts() missing script: %s", expected)
		} else {
			if script.Name == "" {
				t.Errorf("Script %s missing name", expected)
			}
			if script.Language == "" {
				t.Errorf("Script %s missing language", expected)
			}
			if script.Content == "" {
				t.Errorf("Script %s missing content", expected)
			}
			if script.Description == "" {
				t.Errorf("Script %s missing description", expected)
			}
		}
	}
}

func TestPrerequisiteGeneration(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	tests := []struct {
		name      string
		languages []string
		expected  []string
	}{
		{
			name:      "Go project",
			languages: []string{"go"},
			expected:  []string{"Go 1.21+"},
		},
		{
			name:      "JavaScript project",
			languages: []string{"javascript"},
			expected:  []string{"Node.js 18+", "npm or yarn"},
		},
		{
			name:      "TypeScript project",
			languages: []string{"typescript"},
			expected:  []string{"Node.js 18+", "npm or yarn"},
		},
		{
			name:      "Mixed project",
			languages: []string{"go", "javascript"},
			expected:  []string{"Go 1.21+", "Node.js 18+", "npm or yarn"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := createTestDocumentationPayload()
			payload.Metadata["languages"] = tt.languages

			prerequisites := integrator.generatePrerequisites(payload)

			for _, expected := range tt.expected {
				found := false
				for _, prereq := range prerequisites {
					if prereq == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("generatePrerequisites() missing prerequisite: %s", expected)
				}
			}
		})
	}
}

func TestEstimatedTimeCalculation(t *testing.T) {
	integrator := NewDocumentationIntegrator(IntegratorConfig{})

	tests := []struct {
		name          string
		functionCount int
		expectedTime  string
	}{
		{
			name:          "small project",
			functionCount: 10,
			expectedTime:  "10-15 minutes",
		},
		{
			name:          "medium project",
			functionCount: 75,
			expectedTime:  "15-30 minutes",
		},
		{
			name:          "large project",
			functionCount: 200,
			expectedTime:  "30-45 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := createTestDocumentationPayload()
			payload.StructuralData.TotalFunctions = tt.functionCount

			estimatedTime := integrator.calculateEstimatedTime(payload)

			if estimatedTime != tt.expectedTime {
				t.Errorf("calculateEstimatedTime() = %v, want %v", estimatedTime, tt.expectedTime)
			}
		})
	}
}

// Helper function to create test documentation payload
func createTestDocumentationPayload() *DocumentationPayload {
	return &DocumentationPayload{
		AnalysisID:  "test-analysis-123",
		ProjectPath: "/test/project",
		GeneratedAt: time.Now(),
		ComponentMap: &ComponentMap{
			Components: []Component{
				{
					ID:    "test-component-1",
					Name:  "Test Component 1",
					Type:  "service",
					Files: []string{"file1.ts"},
				},
				{
					ID:    "test-component-2",
					Name:  "Test Component 2",
					Type:  "utility",
					Files: []string{"file2.js"},
				},
			},
			Layers: []ArchitecturalLayer{
				{
					Name:       "Business Logic",
					Level:      1,
					Components: []string{"test-component-1"},
				},
			},
		},
		StructuralData: &StructuralMetrics{
			TotalFunctions:  50,
			TotalClasses:    10,
			TotalInterfaces: 5,
			TotalVariables:  100,
			TotalExports:    20,
			TotalImports:    15,
		},
		DependencyGraph: DependenciesSection{
			GraphMetrics: DependencyGraphMetrics{
				TotalNodes: 5,
				TotalEdges: 8,
			},
		},
		QualityMetrics: &QualitySection{
			OverallScore:          85.0,
			DocumentationCoverage: 80.0,
			TestCoverage:          70.0,
			CodeConsistency:       85.0,
			Maintainability:       75.0,
		},
		Metadata: map[string]interface{}{
			"languages":        []string{"typescript", "javascript"},
			"frameworks":       []string{"react", "express"},
			"file_count":       10,
			"processed_files":  10,
			"analysis_version": "2.1.0",
		},
	}
}
