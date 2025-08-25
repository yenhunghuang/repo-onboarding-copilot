package ast

import (
	"fmt"
	"time"
)

// DocumentationPayload represents data prepared for Documentation Generation Engine
type DocumentationPayload struct {
	AnalysisID      string                 `json:"analysis_id"`
	ProjectPath     string                 `json:"project_path"`
	GeneratedAt     time.Time              `json:"generated_at"`
	ComponentMap    *ComponentMap          `json:"component_map"`
	StructuralData  *StructuralMetrics     `json:"structural_data"`
	DependencyGraph DependenciesSection    `json:"dependency_graph"`
	QualityMetrics  *QualitySection        `json:"quality_metrics"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DocumentationIntegrator handles integration with Documentation Generation Engine
type DocumentationIntegrator struct {
	formatter *OutputFormatter
	config    IntegratorConfig
}

// IntegratorConfig configures documentation integration behavior
type IntegratorConfig struct {
	EnableRunbookGeneration    bool   `json:"enable_runbook_generation"`
	EnableArchitectureDiagrams bool   `json:"enable_architecture_diagrams"`
	EnableLearningRoadmap      bool   `json:"enable_learning_roadmap"`
	EnableSecurityReport       bool   `json:"enable_security_report"`
	EnableQualityReport        bool   `json:"enable_quality_report"`
	OutputDirectory            string `json:"output_directory"`
	TemplateVersion            string `json:"template_version"`
}

// AnalysisOrchestrator defines the interface for orchestrating analysis results
type AnalysisOrchestrator interface {
	ProcessAnalysisResult(payload *DocumentationPayload) error
	GenerateRunbook(payload *DocumentationPayload) (*RunbookData, error)
	GenerateArchitectureDiagrams(payload *DocumentationPayload) ([]DiagramData, error)
	GenerateLearningRoadmap(payload *DocumentationPayload) (*RoadmapData, error)
	GetProcessingStatus(analysisID string) (*ProcessingStatus, error)
}

// RunbookData represents executable runbook content
type RunbookData struct {
	Title                string                 `json:"title"`
	Description          string                 `json:"description"`
	SetupSteps           []RunbookStep          `json:"setup_steps"`
	ValidationSteps      []RunbookStep          `json:"validation_steps"`
	TroubleshootingSteps []RunbookStep          `json:"troubleshooting_steps"`
	Scripts              map[string]ScriptData  `json:"scripts"`
	Prerequisites        []string               `json:"prerequisites"`
	EstimatedTime        string                 `json:"estimated_time"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// RunbookStep represents an individual step in a runbook
type RunbookStep struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Commands    []string          `json:"commands,omitempty"`
	Scripts     []string          `json:"scripts,omitempty"`
	Validation  []string          `json:"validation,omitempty"`
	Notes       []string          `json:"notes,omitempty"`
	References  []string          `json:"references,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ScriptData represents executable script information
type ScriptData struct {
	Name        string   `json:"name"`
	Language    string   `json:"language"`
	Content     string   `json:"content"`
	Parameters  []string `json:"parameters,omitempty"`
	Returns     string   `json:"returns,omitempty"`
	Description string   `json:"description"`
}

// DiagramData represents architecture diagram information
type DiagramData struct {
	Type        string                 `json:"type"` // component, sequence, dependency, flow
	Title       string                 `json:"title"`
	Content     string                 `json:"content"` // Diagram markup (mermaid, plantuml, etc.)
	Format      string                 `json:"format"`  // mermaid, plantuml, dot
	Metadata    map[string]interface{} `json:"metadata"`
	Components  []DiagramComponent     `json:"components"`
	Connections []DiagramConnection    `json:"connections"`
}

// DiagramComponent represents a component in an architecture diagram
type DiagramComponent struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Layer       string            `json:"layer"`
	Description string            `json:"description"`
	Properties  map[string]string `json:"properties"`
}

// DiagramConnection represents connections between diagram components
type DiagramConnection struct {
	From          string            `json:"from"`
	To            string            `json:"to"`
	Type          string            `json:"type"` // uses, extends, implements, contains
	Label         string            `json:"label,omitempty"`
	Bidirectional bool              `json:"bidirectional"`
	Properties    map[string]string `json:"properties"`
}

// RoadmapData represents learning roadmap content
type RoadmapData struct {
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	TotalDuration string                 `json:"total_duration"`
	Phases        []RoadmapPhase         `json:"phases"`
	Prerequisites []string               `json:"prerequisites"`
	Resources     []RoadmapResource      `json:"resources"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// RoadmapPhase represents a phase in the learning roadmap
type RoadmapPhase struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Duration    string             `json:"duration"`
	Description string             `json:"description"`
	Goals       []string           `json:"goals"`
	Milestones  []RoadmapMilestone `json:"milestones"`
	Activities  []RoadmapActivity  `json:"activities"`
}

// RoadmapMilestone represents a milestone in a roadmap phase
type RoadmapMilestone struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Deliverables []string `json:"deliverables"`
	Success      []string `json:"success_criteria"`
}

// RoadmapActivity represents an activity in a roadmap phase
type RoadmapActivity struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"` // reading, coding, research, practice
	Duration    string   `json:"duration"`
	Description string   `json:"description"`
	Resources   []string `json:"resources"`
	Outcomes    []string `json:"outcomes"`
}

// RoadmapResource represents a resource in the learning roadmap
type RoadmapResource struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Type        string            `json:"type"` // documentation, tutorial, tool, reference
	URL         string            `json:"url,omitempty"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags"`
	Metadata    map[string]string `json:"metadata"`
}

// ProcessingStatus represents the status of documentation generation
type ProcessingStatus struct {
	AnalysisID  string                 `json:"analysis_id"`
	Status      string                 `json:"status"`   // pending, processing, completed, failed
	Progress    float64                `json:"progress"` // 0-100
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Artifacts   []GeneratedArtifact    `json:"artifacts"`
	Errors      []string               `json:"errors,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GeneratedArtifact represents a generated documentation artifact
type GeneratedArtifact struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // runbook, diagram, roadmap, report
	Name        string                 `json:"name"`
	Path        string                 `json:"path,omitempty"`
	Size        int64                  `json:"size"`
	GeneratedAt time.Time              `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewDocumentationIntegrator creates a new documentation integrator
func NewDocumentationIntegrator(config IntegratorConfig) *DocumentationIntegrator {
	// Set default configuration
	if config.OutputDirectory == "" {
		config.OutputDirectory = "./docs/generated"
	}
	if config.TemplateVersion == "" {
		config.TemplateVersion = "1.0"
	}

	formatter := NewOutputFormatter(FormatterOptions{
		PrettyPrint:     true,
		IncludeMetadata: true,
		OutputFormat:    "json",
	})

	return &DocumentationIntegrator{
		formatter: formatter,
		config:    config,
	}
}

// PrepareForDocumentation converts analysis results to documentation payload
func (di *DocumentationIntegrator) PrepareForDocumentation(result *AnalysisResult, metadata *AnalysisMetadata) (*DocumentationPayload, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result cannot be nil")
	}

	// Format the analysis result first
	standardized, err := di.formatter.FormatAnalysisResult(result, metadata, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to format analysis result: %w", err)
	}

	// Create documentation payload
	payload, err := di.formatter.CreateDocumentationPayload(standardized)
	if err != nil {
		return nil, fmt.Errorf("failed to create documentation payload: %w", err)
	}

	return payload, nil
}

// GenerateRunbookStructure creates runbook structure from AST analysis
func (di *DocumentationIntegrator) GenerateRunbookStructure(payload *DocumentationPayload) (*RunbookData, error) {
	if payload == nil {
		return nil, fmt.Errorf("documentation payload cannot be nil")
	}

	runbook := &RunbookData{
		Title:                fmt.Sprintf("Development Runbook - %s", di.getProjectName(payload.ProjectPath)),
		Description:          "Automated runbook generated from AST analysis for repository onboarding and development",
		SetupSteps:           di.generateSetupSteps(payload),
		ValidationSteps:      di.generateValidationSteps(payload),
		TroubleshootingSteps: di.generateTroubleshootingSteps(payload),
		Scripts:              di.generateScripts(payload),
		Prerequisites:        di.generatePrerequisites(payload),
		EstimatedTime:        di.calculateEstimatedTime(payload),
		Metadata: map[string]interface{}{
			"generated_from": "ast-analysis",
			"analysis_id":    payload.AnalysisID,
			"generated_at":   payload.GeneratedAt,
			"languages":      payload.Metadata["languages"],
			"frameworks":     payload.Metadata["frameworks"],
		},
	}

	return runbook, nil
}

// GenerateArchitectureDiagrams creates diagram specifications from component map
func (di *DocumentationIntegrator) GenerateArchitectureDiagrams(payload *DocumentationPayload) ([]DiagramData, error) {
	if payload == nil {
		return nil, fmt.Errorf("documentation payload cannot be nil")
	}

	var diagrams []DiagramData

	// Component architecture diagram
	if payload.ComponentMap != nil && len(payload.ComponentMap.Components) > 0 {
		componentDiagram := di.generateComponentDiagram(payload)
		diagrams = append(diagrams, componentDiagram)
	}

	// Dependency diagram
	if payload.DependencyGraph.GraphMetrics.TotalNodes > 0 {
		dependencyDiagram := di.generateDependencyDiagram(payload)
		diagrams = append(diagrams, dependencyDiagram)
	}

	// Layer architecture diagram
	if payload.ComponentMap != nil && len(payload.ComponentMap.Layers) > 0 {
		layerDiagram := di.generateLayerDiagram(payload)
		diagrams = append(diagrams, layerDiagram)
	}

	return diagrams, nil
}

// GenerateLearningRoadmap creates learning roadmap from analysis results
func (di *DocumentationIntegrator) GenerateLearningRoadmap(payload *DocumentationPayload) (*RoadmapData, error) {
	if payload == nil {
		return nil, fmt.Errorf("documentation payload cannot be nil")
	}

	roadmap := &RoadmapData{
		Title:         fmt.Sprintf("Learning Roadmap - %s", di.getProjectName(payload.ProjectPath)),
		Description:   "Automated learning roadmap generated from codebase analysis",
		TotalDuration: "90 days",
		Phases:        di.generateRoadmapPhases(payload),
		Prerequisites: di.generateLearningPrerequisites(payload),
		Resources:     di.generateLearningResources(payload),
		Metadata: map[string]interface{}{
			"generated_from":   "ast-analysis",
			"analysis_id":      payload.AnalysisID,
			"complexity_score": payload.QualityMetrics.OverallScore,
			"languages":        payload.Metadata["languages"],
		},
	}

	return roadmap, nil
}

// GetIntegrationMetadata returns metadata for analysis orchestrator consumption
func (di *DocumentationIntegrator) GetIntegrationMetadata(payload *DocumentationPayload) map[string]interface{} {
	// Base metadata structure that's always present
	metadata := map[string]interface{}{
		"integration_version": "2.1.0",
		"supported_formats":   []string{"runbook", "architecture_diagrams", "learning_roadmap"},
		"generation_capabilities": map[string]bool{
			"runbook_generation":    di.config.EnableRunbookGeneration,
			"architecture_diagrams": di.config.EnableArchitectureDiagrams,
			"learning_roadmap":      di.config.EnableLearningRoadmap,
			"security_report":       di.config.EnableSecurityReport,
			"quality_report":        di.config.EnableQualityReport,
		},
		"output_configuration": map[string]interface{}{
			"output_directory": di.config.OutputDirectory,
			"template_version": di.config.TemplateVersion,
		},
	}

	// Add payload-specific data if available
	if payload != nil {
		metadata["analysis_summary"] = map[string]interface{}{
			"component_count":  len(payload.ComponentMap.Components),
			"function_count":   payload.StructuralData.TotalFunctions,
			"class_count":      payload.StructuralData.TotalClasses,
			"interface_count":  payload.StructuralData.TotalInterfaces,
			"dependency_count": payload.DependencyGraph.GraphMetrics.TotalNodes,
		}
		metadata["quality_indicators"] = map[string]interface{}{
			"overall_score":          payload.QualityMetrics.OverallScore,
			"documentation_coverage": payload.QualityMetrics.DocumentationCoverage,
			"maintainability":        payload.QualityMetrics.Maintainability,
			"code_consistency":       payload.QualityMetrics.CodeConsistency,
		}
	} else {
		// Provide default values for nil payload
		metadata["analysis_summary"] = map[string]interface{}{
			"component_count":  0,
			"function_count":   0,
			"class_count":      0,
			"interface_count":  0,
			"dependency_count": 0,
		}
		metadata["quality_indicators"] = map[string]interface{}{
			"overall_score":          0.0,
			"documentation_coverage": 0.0,
			"maintainability":        0.0,
			"code_consistency":       0.0,
		}
	}

	return metadata
}

// Private helper methods

func (di *DocumentationIntegrator) getProjectName(projectPath string) string {
	// Extract project name from path
	if projectPath == "" {
		return "Unknown Project"
	}
	// Simple implementation - would be enhanced with actual path parsing
	return "Repository Project"
}

func (di *DocumentationIntegrator) generateSetupSteps(payload *DocumentationPayload) []RunbookStep {
	return []RunbookStep{
		{
			ID:          "environment-setup",
			Title:       "Environment Setup",
			Description: "Setup development environment based on detected languages and frameworks",
			Commands:    di.generateSetupCommands(payload),
			Validation:  []string{"Verify all dependencies are installed", "Check environment configuration"},
		},
		{
			ID:          "dependency-installation",
			Title:       "Dependency Installation",
			Description: "Install project dependencies",
			Commands:    []string{"npm install", "go mod download"},
			Validation:  []string{"Check for dependency conflicts", "Verify installation success"},
		},
	}
}

func (di *DocumentationIntegrator) generateValidationSteps(payload *DocumentationPayload) []RunbookStep {
	return []RunbookStep{
		{
			ID:          "build-validation",
			Title:       "Build Validation",
			Description: "Validate that the project builds successfully",
			Commands:    []string{"make build", "go build ./..."},
			Validation:  []string{"Build completes without errors", "Binary/artifacts generated"},
		},
	}
}

func (di *DocumentationIntegrator) generateTroubleshootingSteps(payload *DocumentationPayload) []RunbookStep {
	return []RunbookStep{
		{
			ID:          "common-issues",
			Title:       "Common Issues Resolution",
			Description: "Solutions for frequently encountered issues",
			Notes: []string{
				"Check Go version compatibility",
				"Verify GOPATH and GOMODULES settings",
				"Review dependency versions",
			},
		},
	}
}

func (di *DocumentationIntegrator) generateScripts(payload *DocumentationPayload) map[string]ScriptData {
	return map[string]ScriptData{
		"build": {
			Name:        "build.sh",
			Language:    "bash",
			Content:     "#!/bin/bash\nset -e\ngo build -o bin/app ./cmd/main.go",
			Description: "Build the application",
		},
		"test": {
			Name:        "test.sh",
			Language:    "bash",
			Content:     "#!/bin/bash\nset -e\ngo test ./...",
			Description: "Run all tests",
		},
	}
}

func (di *DocumentationIntegrator) generatePrerequisites(payload *DocumentationPayload) []string {
	prerequisites := []string{}

	if languages, ok := payload.Metadata["languages"].([]string); ok {
		for _, lang := range languages {
			switch lang {
			case "go":
				prerequisites = append(prerequisites, "Go 1.21+")
			case "javascript", "typescript":
				prerequisites = append(prerequisites, "Node.js 18+", "npm or yarn")
			}
		}
	}

	return prerequisites
}

func (di *DocumentationIntegrator) generateSetupCommands(payload *DocumentationPayload) []string {
	commands := []string{}

	if languages, ok := payload.Metadata["languages"].([]string); ok {
		for _, lang := range languages {
			switch lang {
			case "go":
				commands = append(commands, "go mod download", "go mod tidy")
			case "javascript", "typescript":
				commands = append(commands, "npm install")
			}
		}
	}

	return commands
}

func (di *DocumentationIntegrator) calculateEstimatedTime(payload *DocumentationPayload) string {
	// Simple estimation based on complexity
	if payload.StructuralData.TotalFunctions > 100 {
		return "30-45 minutes"
	} else if payload.StructuralData.TotalFunctions > 50 {
		return "15-30 minutes"
	}
	return "10-15 minutes"
}

func (di *DocumentationIntegrator) generateComponentDiagram(payload *DocumentationPayload) DiagramData {
	return DiagramData{
		Type:    "component",
		Title:   "Component Architecture",
		Content: "// Mermaid diagram content would be generated here",
		Format:  "mermaid",
		Metadata: map[string]interface{}{
			"component_count": len(payload.ComponentMap.Components),
			"layer_count":     len(payload.ComponentMap.Layers),
		},
	}
}

func (di *DocumentationIntegrator) generateDependencyDiagram(payload *DocumentationPayload) DiagramData {
	return DiagramData{
		Type:    "dependency",
		Title:   "Dependency Graph",
		Content: "// Dependency graph content would be generated here",
		Format:  "mermaid",
		Metadata: map[string]interface{}{
			"node_count": payload.DependencyGraph.GraphMetrics.TotalNodes,
			"edge_count": payload.DependencyGraph.GraphMetrics.TotalEdges,
		},
	}
}

func (di *DocumentationIntegrator) generateLayerDiagram(payload *DocumentationPayload) DiagramData {
	return DiagramData{
		Type:    "layered",
		Title:   "Layered Architecture",
		Content: "// Layer architecture content would be generated here",
		Format:  "mermaid",
		Metadata: map[string]interface{}{
			"layer_count": len(payload.ComponentMap.Layers),
		},
	}
}

func (di *DocumentationIntegrator) generateRoadmapPhases(payload *DocumentationPayload) []RoadmapPhase {
	return []RoadmapPhase{
		{
			ID:          "phase-1-orientation",
			Title:       "Orientation & Setup (Days 1-30)",
			Duration:    "30 days",
			Description: "Get familiar with the codebase structure and setup development environment",
			Goals:       []string{"Understand project structure", "Setup development environment", "Run first builds"},
		},
		{
			ID:          "phase-2-comprehension",
			Title:       "Code Comprehension (Days 31-60)",
			Duration:    "30 days",
			Description: "Deep dive into code understanding and component relationships",
			Goals:       []string{"Understand component interactions", "Map data flows", "Identify key patterns"},
		},
		{
			ID:          "phase-3-contribution",
			Title:       "Active Contribution (Days 61-90)",
			Duration:    "30 days",
			Description: "Start making meaningful contributions to the codebase",
			Goals:       []string{"Fix bugs", "Add features", "Improve documentation", "Optimize performance"},
		},
	}
}

func (di *DocumentationIntegrator) generateLearningPrerequisites(payload *DocumentationPayload) []string {
	return []string{
		"Basic programming knowledge",
		"Familiarity with version control (Git)",
		"Understanding of software development lifecycle",
	}
}

func (di *DocumentationIntegrator) generateLearningResources(payload *DocumentationPayload) []RoadmapResource {
	return []RoadmapResource{
		{
			ID:          "official-docs",
			Title:       "Official Documentation",
			Type:        "documentation",
			Description: "Project's official documentation and API references",
			Tags:        []string{"documentation", "reference"},
		},
		{
			ID:          "code-examples",
			Title:       "Code Examples",
			Type:        "reference",
			Description: "Key code examples and patterns from the analysis",
			Tags:        []string{"examples", "patterns"},
		},
	}
}
