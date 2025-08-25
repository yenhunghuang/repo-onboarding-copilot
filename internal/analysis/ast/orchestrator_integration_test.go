package ast

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewOrchestrationIntegrator(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{
		ProjectRoot: "/test/project",
	})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)

	if integrator == nil {
		t.Error("NewOrchestrationIntegrator() returned nil")
		return
	}

	if integrator.analyzer != testAnalyzer {
		t.Error("NewOrchestrationIntegrator() analyzer mismatch")
	}

	if integrator.formatter == nil {
		t.Error("NewOrchestrationIntegrator() missing formatter")
	}

	if integrator.docIntegrator == nil {
		t.Error("NewOrchestrationIntegrator() missing document integrator")
	}
}

func TestGetOrchestrationMetadata(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	metadata := integrator.GetOrchestrationMetadata()

	if metadata == nil {
		t.Error("GetOrchestrationMetadata() returned nil")
		return
	}

	// Validate required sections
	if len(metadata.AnalysisCapabilities.SupportedLanguages) == 0 {
		t.Error("GetOrchestrationMetadata() missing supported languages")
	}

	if len(metadata.OutputFormats) == 0 {
		t.Error("GetOrchestrationMetadata() missing output formats")
	}

	if len(metadata.IntegrationPoints) == 0 {
		t.Error("GetOrchestrationMetadata() missing integration points")
	}

	// Validate analysis capabilities
	capabilities := metadata.AnalysisCapabilities

	// Check JavaScript support
	jsSupported := false
	tsSupported := false
	for _, lang := range capabilities.SupportedLanguages {
		if lang.Language == "javascript" {
			jsSupported = true
			if len(lang.Extensions) == 0 {
				t.Error("JavaScript language support missing extensions")
			}
			if len(lang.Features) == 0 {
				t.Error("JavaScript language support missing features")
			}
		}
		if lang.Language == "typescript" {
			tsSupported = true
			if len(lang.Extensions) == 0 {
				t.Error("TypeScript language support missing extensions")
			}
			if len(lang.Features) == 0 {
				t.Error("TypeScript language support missing features")
			}
		}
	}

	if !jsSupported {
		t.Error("GetOrchestrationMetadata() missing JavaScript support")
	}

	if !tsSupported {
		t.Error("GetOrchestrationMetadata() missing TypeScript support")
	}

	// Validate extraction capabilities
	expectedCapabilities := []string{
		"function_declarations", "class_declarations", "interface_declarations",
		"variable_declarations", "import_statements", "export_statements",
	}
	for _, expected := range expectedCapabilities {
		found := false
		for _, capability := range capabilities.ExtractionCapabilities {
			if capability == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetOrchestrationMetadata() missing extraction capability: %s", expected)
		}
	}

	// Validate scalability limits
	limits := capabilities.ScalabilityLimits
	if limits.MaxFiles <= 0 {
		t.Error("GetOrchestrationMetadata() invalid MaxFiles limit")
	}
	if limits.MaxFileSize <= 0 {
		t.Error("GetOrchestrationMetadata() invalid MaxFileSize limit")
	}
	if limits.MaxMemory <= 0 {
		t.Error("GetOrchestrationMetadata() invalid MaxMemory limit")
	}

	// Validate quality metrics
	quality := metadata.QualityMetrics
	if quality.Reliability.Score < 0 || quality.Reliability.Score > 100 {
		t.Error("GetOrchestrationMetadata() invalid reliability score")
	}
	if quality.Performance.Score < 0 || quality.Performance.Score > 100 {
		t.Error("GetOrchestrationMetadata() invalid performance score")
	}
	if quality.Accuracy.Score < 0 || quality.Accuracy.Score > 100 {
		t.Error("GetOrchestrationMetadata() invalid accuracy score")
	}

	// Validate performance profile
	perf := metadata.PerformanceProfile
	if perf.Throughput.FilesPerSecond <= 0 {
		t.Error("GetOrchestrationMetadata() invalid files per second")
	}
	if perf.Latency.AverageLatency <= 0 {
		t.Error("GetOrchestrationMetadata() invalid average latency")
	}

	// Validate resource requirements
	resources := metadata.ResourceRequirements
	if resources.Minimum.Memory <= 0 {
		t.Error("GetOrchestrationMetadata() invalid minimum memory requirement")
	}
	if resources.Recommended.Memory <= resources.Minimum.Memory {
		t.Error("GetOrchestrationMetadata() recommended memory should be greater than minimum")
	}
	if resources.Optimal.Memory <= resources.Recommended.Memory {
		t.Error("GetOrchestrationMetadata() optimal memory should be greater than recommended")
	}

	// Validate metadata
	if metadata.Metadata == nil {
		t.Error("GetOrchestrationMetadata() missing metadata section")
	} else {
		if version, ok := metadata.Metadata["version"].(string); !ok || version == "" {
			t.Error("GetOrchestrationMetadata() missing or invalid version")
		}
		if componentType, ok := metadata.Metadata["component_type"].(string); !ok || componentType != "ast_parser" {
			t.Error("GetOrchestrationMetadata() missing or invalid component_type")
		}
		if ready, ok := metadata.Metadata["orchestrator_ready"].(bool); !ok || !ready {
			t.Error("GetOrchestrationMetadata() orchestrator should be ready")
		}
	}
}

func TestProcessForOrchestrator(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)

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
			payload, err := integrator.ProcessForOrchestrator(tt.result, tt.metadata)

			if tt.wantError {
				if err == nil {
					t.Error("ProcessForOrchestrator() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessForOrchestrator() unexpected error: %v", err)
				return
			}

			if payload == nil {
				t.Error("ProcessForOrchestrator() returned nil payload")
				return
			}

			// Validate payload structure
			if payload.AnalysisResult == nil {
				t.Error("ProcessForOrchestrator() missing analysis result")
			}

			if payload.DocumentationPayload == nil {
				t.Error("ProcessForOrchestrator() missing documentation payload")
			}

			if payload.OrchestrationMetadata == nil {
				t.Error("ProcessForOrchestrator() missing orchestration metadata")
			}

			if payload.ProcessingInstructions == nil {
				t.Error("ProcessForOrchestrator() missing processing instructions")
			}

			if payload.QualityAssurance == nil {
				t.Error("ProcessForOrchestrator() missing quality assurance")
			}

			if payload.IntegrationManifest == nil {
				t.Error("ProcessForOrchestrator() missing integration manifest")
			}

			// Validate cross-references
			if payload.AnalysisResult.AnalysisID != payload.DocumentationPayload.AnalysisID {
				t.Error("ProcessForOrchestrator() analysis ID mismatch between payloads")
			}
		})
	}
}

func TestProcessingInstructions(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	testResult := createTestStandardizedResult()

	instructions := integrator.getProcessingInstructions(testResult)

	if instructions == nil {
		t.Error("getProcessingInstructions() returned nil")
		return
	}

	// Validate recommended processing order
	if len(instructions.RecommendedProcessingOrder) == 0 {
		t.Error("getProcessingInstructions() missing recommended processing order")
	}

	expectedSteps := []string{
		"validate_input_data", "generate_component_map", "create_dependency_graph",
		"calculate_quality_metrics", "generate_documentation",
	}
	for _, expected := range expectedSteps {
		found := false
		for _, step := range instructions.RecommendedProcessingOrder {
			if step == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getProcessingInstructions() missing processing step: %s", expected)
		}
	}

	// Validate parallelization opportunities
	if len(instructions.ParallelizationOpportunities) == 0 {
		t.Error("getProcessingInstructions() missing parallelization opportunities")
	}

	for i, group := range instructions.ParallelizationOpportunities {
		if group.GroupName == "" {
			t.Errorf("Parallelization group %d missing name", i)
		}
		if len(group.Tasks) == 0 {
			t.Errorf("Parallelization group %d missing tasks", i)
		}
		if group.EstimatedTime <= 0 {
			t.Errorf("Parallelization group %d invalid estimated time", i)
		}
	}

	// Validate critical path
	if len(instructions.CriticalPath) == 0 {
		t.Error("getProcessingInstructions() missing critical path")
	}

	// Validate optimization hints
	if len(instructions.OptimizationHints) == 0 {
		t.Error("getProcessingInstructions() missing optimization hints")
	}

	for i, hint := range instructions.OptimizationHints {
		if hint.Component == "" {
			t.Errorf("Optimization hint %d missing component", i)
		}
		if hint.Optimization == "" {
			t.Errorf("Optimization hint %d missing optimization", i)
		}
		if hint.Impact == "" {
			t.Errorf("Optimization hint %d missing impact", i)
		}
		if hint.Priority == "" {
			t.Errorf("Optimization hint %d missing priority", i)
		}
	}
}

func TestQualityAssuranceData(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	testResult := createTestStandardizedResult()

	qaData := integrator.getQualityAssuranceData(testResult)

	if qaData == nil {
		t.Error("getQualityAssuranceData() returned nil")
		return
	}

	// Validate data integrity checks
	if len(qaData.DataIntegrityChecks) == 0 {
		t.Error("getQualityAssuranceData() missing data integrity checks")
	}

	for i, check := range qaData.DataIntegrityChecks {
		if check.CheckName == "" {
			t.Errorf("Data integrity check %d missing name", i)
		}
		if check.CheckType == "" {
			t.Errorf("Data integrity check %d missing type", i)
		}
		if check.Status == "" {
			t.Errorf("Data integrity check %d missing status", i)
		}
		if check.Criteria == nil {
			t.Errorf("Data integrity check %d missing criteria", i)
		}
	}

	// Validate accuracy metrics
	if len(qaData.AccuracyMetrics) == 0 {
		t.Error("getQualityAssuranceData() missing accuracy metrics")
	}

	for i, metric := range qaData.AccuracyMetrics {
		if metric.MetricName == "" {
			t.Errorf("Accuracy metric %d missing name", i)
		}
		if metric.AccuracyScore < 0 || metric.AccuracyScore > 100 {
			t.Errorf("Accuracy metric %d invalid score: %f", i, metric.AccuracyScore)
		}
		if metric.SampleSize <= 0 {
			t.Errorf("Accuracy metric %d invalid sample size: %d", i, metric.SampleSize)
		}
	}

	// Validate completeness checks
	if len(qaData.CompletenessChecks) == 0 {
		t.Error("getQualityAssuranceData() missing completeness checks")
	}

	for i, check := range qaData.CompletenessChecks {
		if check.ComponentName == "" {
			t.Errorf("Completeness check %d missing component name", i)
		}
		if check.CompletenessRate < 0 || check.CompletenessRate > 100 {
			t.Errorf("Completeness check %d invalid rate: %f", i, check.CompletenessRate)
		}
	}
}

func TestIntegrationManifest(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	manifest := integrator.getIntegrationManifest()

	if manifest == nil {
		t.Error("getIntegrationManifest() returned nil")
		return
	}

	// Validate supported protocols
	if len(manifest.SupportedProtocols) == 0 {
		t.Error("getIntegrationManifest() missing supported protocols")
	}

	expectedProtocols := []string{"direct_call", "json_rpc", "rest_api"}
	for _, expected := range expectedProtocols {
		found := false
		for _, protocol := range manifest.SupportedProtocols {
			if protocol == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getIntegrationManifest() missing protocol: %s", expected)
		}
	}

	// Validate required interfaces
	if len(manifest.RequiredInterfaces) == 0 {
		t.Error("getIntegrationManifest() missing required interfaces")
	}

	for i, iface := range manifest.RequiredInterfaces {
		if iface.InterfaceName == "" {
			t.Errorf("Required interface %d missing name", i)
		}
		if iface.Version == "" {
			t.Errorf("Required interface %d missing version", i)
		}
		if iface.Protocol == "" {
			t.Errorf("Required interface %d missing protocol", i)
		}
		if len(iface.Operations) == 0 {
			t.Errorf("Required interface %d missing operations", i)
		}
	}

	// Validate provided interfaces
	if len(manifest.ProvidedInterfaces) == 0 {
		t.Error("getIntegrationManifest() missing provided interfaces")
	}

	// Validate configuration schema
	if manifest.ConfigurationSchema == nil {
		t.Error("getIntegrationManifest() missing configuration schema")
	}

	// Validate environment requirements
	if len(manifest.EnvironmentRequirements) == 0 {
		t.Error("getIntegrationManifest() missing environment requirements")
	}
}

func TestCompatibilityMatrix(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	compatibility := integrator.getCompatibilityMatrix()

	// Validate analysis engines compatibility
	if len(compatibility.AnalysisEngines) == 0 {
		t.Error("getCompatibilityMatrix() missing analysis engines compatibility")
	}

	// Validate output generators compatibility
	if len(compatibility.OutputGenerators) == 0 {
		t.Error("getCompatibilityMatrix() missing output generators compatibility")
	}

	// Check specific compatibility entries
	for i, comp := range compatibility.OutputGenerators {
		if comp.ComponentName == "" {
			t.Errorf("Output generator compatibility %d missing component name", i)
		}
		if comp.Version == "" {
			t.Errorf("Output generator compatibility %d missing version", i)
		}
		if comp.CompatibilityLevel == "" {
			t.Errorf("Output generator compatibility %d missing compatibility level", i)
		}

		validLevels := []string{"full", "partial", "limited", "none"}
		levelValid := false
		for _, valid := range validLevels {
			if comp.CompatibilityLevel == valid {
				levelValid = true
				break
			}
		}
		if !levelValid {
			t.Errorf("Output generator compatibility %d invalid compatibility level: %s",
				i, comp.CompatibilityLevel)
		}
	}
}

func TestOrchestrationMetadataJSONSerialization(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	metadata := integrator.GetOrchestrationMetadata()

	// Test pretty JSON serialization
	prettyJSON, err := metadata.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() unexpected error: %v", err)
		return
	}

	if len(prettyJSON) == 0 {
		t.Error("ToJSON() returned empty JSON")
		return
	}

	// Validate JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(prettyJSON, &parsed); err != nil {
		t.Errorf("ToJSON() produced invalid JSON: %v", err)
		return
	}

	// Check required top-level fields
	requiredFields := []string{
		"analysis_capabilities", "output_formats", "integration_points",
		"quality_metrics", "performance_profile", "resource_requirements",
		"dependencies", "compatibility", "metadata",
	}
	for _, field := range requiredFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("ToJSON() missing required field: %s", field)
		}
	}

	// Test compact JSON serialization
	compactJSON, err := metadata.ToCompactJSON()
	if err != nil {
		t.Errorf("ToCompactJSON() unexpected error: %v", err)
		return
	}

	if len(compactJSON) == 0 {
		t.Error("ToCompactJSON() returned empty JSON")
		return
	}

	// Compact JSON should be smaller (no pretty formatting)
	if len(compactJSON) >= len(prettyJSON) {
		t.Log("ToCompactJSON() may not be significantly more compact for test data")
	}

	// Validate compact JSON structure
	var compactParsed map[string]interface{}
	if err := json.Unmarshal(compactJSON, &compactParsed); err != nil {
		t.Errorf("ToCompactJSON() produced invalid JSON: %v", err)
		return
	}
}

func TestPerformanceProfileValidation(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	profile := integrator.getPerformanceProfile()

	// Validate throughput metrics
	throughput := profile.Throughput
	if throughput.FilesPerSecond <= 0 {
		t.Error("Invalid files per second throughput")
	}
	if throughput.LinesPerSecond <= 0 {
		t.Error("Invalid lines per second throughput")
	}
	if throughput.TokensPerSecond <= 0 {
		t.Error("Invalid tokens per second throughput")
	}
	if throughput.OptimizedThroughput <= throughput.FilesPerSecond {
		t.Error("Optimized throughput should be higher than regular throughput")
	}

	// Validate latency metrics
	latency := profile.Latency
	if latency.AverageLatency <= 0 {
		t.Error("Invalid average latency")
	}
	if latency.P50Latency <= 0 {
		t.Error("Invalid P50 latency")
	}
	if latency.P95Latency <= latency.P50Latency {
		t.Error("P95 latency should be higher than P50")
	}
	if latency.P99Latency <= latency.P95Latency {
		t.Error("P99 latency should be higher than P95")
	}

	// Validate resource utilization
	resources := profile.ResourceUtilization
	if resources.AverageMemoryUsage <= 0 {
		t.Error("Invalid average memory usage")
	}
	if resources.PeakMemoryUsage <= resources.AverageMemoryUsage {
		t.Error("Peak memory usage should be higher than average")
	}

	// Validate scalability metrics
	scalability := profile.Scalability
	if scalability.OptimalWorkerCount <= 0 {
		t.Error("Invalid optimal worker count")
	}
	if scalability.ScalingEfficiency <= 0 || scalability.ScalingEfficiency > 1 {
		t.Error("Scaling efficiency should be between 0 and 1")
	}

	// Validate optimization features
	if len(profile.OptimizationFeatures) == 0 {
		t.Error("Missing optimization features")
	}

	expectedFeatures := []string{
		"multi_threaded_parsing", "memory_monitoring", "timeout_management",
	}
	for _, expected := range expectedFeatures {
		found := false
		for _, feature := range profile.OptimizationFeatures {
			if feature == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing optimization feature: %s", expected)
		}
	}
}

func TestResourceRequirementsValidation(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	requirements := integrator.getResourceRequirements()

	// Validate minimum requirements
	min := requirements.Minimum
	if min.Memory <= 0 {
		t.Error("Invalid minimum memory requirement")
	}
	if min.CPU <= 0 {
		t.Error("Invalid minimum CPU requirement")
	}
	if min.Storage <= 0 {
		t.Error("Invalid minimum storage requirement")
	}
	if min.Concurrency <= 0 {
		t.Error("Invalid minimum concurrency requirement")
	}

	// Validate recommended requirements are higher than minimum
	recommended := requirements.Recommended
	if recommended.Memory <= min.Memory {
		t.Error("Recommended memory should be higher than minimum")
	}
	if recommended.CPU <= min.CPU {
		t.Error("Recommended CPU should be higher than minimum")
	}
	if recommended.Concurrency <= min.Concurrency {
		t.Error("Recommended concurrency should be higher than minimum")
	}

	// Validate optimal requirements are higher than recommended
	optimal := requirements.Optimal
	if optimal.Memory <= recommended.Memory {
		t.Error("Optimal memory should be higher than recommended")
	}
	if optimal.CPU <= recommended.CPU {
		t.Error("Optimal CPU should be higher than recommended")
	}
	if optimal.Concurrency <= recommended.Concurrency {
		t.Error("Optimal concurrency should be higher than recommended")
	}
}

func TestDependenciesValidation(t *testing.T) {
	testAnalyzer, err := NewAnalyzer(AnalyzerConfig{})
	if err != nil {
		t.Fatalf("Failed to create test analyzer: %v", err)
	}
	defer testAnalyzer.Close()

	integrator := NewOrchestrationIntegrator(testAnalyzer)
	dependencies := integrator.getDependencies()

	// Validate required dependencies
	if len(dependencies.Required) == 0 {
		t.Error("Missing required dependencies")
	}

	for i, dep := range dependencies.Required {
		if dep.Name == "" {
			t.Errorf("Required dependency %d missing name", i)
		}
		if dep.Version == "" {
			t.Errorf("Required dependency %d missing version", i)
		}
		if dep.Type == "" {
			t.Errorf("Required dependency %d missing type", i)
		}
		if !dep.Required {
			t.Errorf("Required dependency %d should be marked as required", i)
		}
	}

	// Check for expected tree-sitter dependency
	treeSSitterFound := false
	for _, dep := range dependencies.Required {
		if dep.Name == "go-tree-sitter" {
			treeSSitterFound = true
			if dep.Type != "library" {
				t.Error("go-tree-sitter should be marked as library type")
			}
		}
	}
	if !treeSSitterFound {
		t.Error("Missing required go-tree-sitter dependency")
	}

	// Validate optional dependencies
	for i, dep := range dependencies.Optional {
		if dep.Name == "" {
			t.Errorf("Optional dependency %d missing name", i)
		}
		if dep.Required {
			t.Errorf("Optional dependency %d should not be marked as required", i)
		}
	}
}
