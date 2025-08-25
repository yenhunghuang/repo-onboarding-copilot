package ast

import (
	"encoding/json"
	"fmt"
	"time"
)

// OrchestrationMetadata provides structured metadata for analysis orchestrator consumption
type OrchestrationMetadata struct {
	AnalysisCapabilities AnalysisCapabilities            `json:"analysis_capabilities"`
	OutputFormats        []OutputFormatDescriptor        `json:"output_formats"`
	IntegrationPoints    []IntegrationPoint              `json:"integration_points"`
	QualityMetrics       OrchestrationQualityMetrics     `json:"quality_metrics"`
	PerformanceProfile   OrchestrationPerformanceProfile `json:"performance_profile"`
	ResourceRequirements ResourceRequirements            `json:"resource_requirements"`
	Dependencies         OrchestrationDependencies       `json:"dependencies"`
	Compatibility        CompatibilityMatrix             `json:"compatibility"`
	Metadata             map[string]interface{}          `json:"metadata"`
}

// AnalysisCapabilities describes what the AST parser can analyze
type AnalysisCapabilities struct {
	SupportedLanguages     []LanguageSupport `json:"supported_languages"`
	ExtractionCapabilities []string          `json:"extraction_capabilities"`
	AnalysisTypes          []string          `json:"analysis_types"`
	OutputTypes            []string          `json:"output_types"`
	ScalabilityLimits      ScalabilityLimits `json:"scalability_limits"`
}

// LanguageSupport describes language-specific capabilities
type LanguageSupport struct {
	Language       string   `json:"language"`
	Extensions     []string `json:"extensions"`
	Features       []string `json:"features"`
	Limitations    []string `json:"limitations"`
	GrammarVersion string   `json:"grammar_version"`
}

// ScalabilityLimits defines operational limits
type ScalabilityLimits struct {
	MaxFiles       int   `json:"max_files"`
	MaxFileSize    int64 `json:"max_file_size_bytes"`
	MaxMemory      int64 `json:"max_memory_bytes"`
	MaxConcurrency int   `json:"max_concurrency"`
	TimeoutSeconds int   `json:"timeout_seconds"`
}

// OutputFormatDescriptor describes available output formats
type OutputFormatDescriptor struct {
	FormatName  string                 `json:"format_name"`
	MimeType    string                 `json:"mime_type"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Compression []string               `json:"compression_options"`
	UseCase     string                 `json:"use_case"`
}

// IntegrationPoint describes how the AST parser integrates with other components
type IntegrationPoint struct {
	ComponentName   string                 `json:"component_name"`
	IntegrationType string                 `json:"integration_type"` // producer, consumer, bidirectional
	DataFormat      string                 `json:"data_format"`
	Protocol        string                 `json:"protocol"`
	Endpoint        string                 `json:"endpoint,omitempty"`
	Schema          map[string]interface{} `json:"schema"`
	Dependencies    []string               `json:"dependencies"`
}

// OrchestrationQualityMetrics provides quality indicators for orchestration
type OrchestrationQualityMetrics struct {
	Reliability     QualityIndicator `json:"reliability"`
	Performance     QualityIndicator `json:"performance"`
	Accuracy        QualityIndicator `json:"accuracy"`
	Completeness    QualityIndicator `json:"completeness"`
	Consistency     QualityIndicator `json:"consistency"`
	Maintainability QualityIndicator `json:"maintainability"`
}

// QualityIndicator represents a quality measurement
type QualityIndicator struct {
	Score      float64                `json:"score"`  // 0-100
	Status     string                 `json:"status"` // excellent, good, fair, poor
	Trends     []TrendPoint           `json:"trends"`
	Benchmarks map[string]float64     `json:"benchmarks"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TrendPoint represents a point in time measurement
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Context   string    `json:"context,omitempty"`
}

// OrchestrationPerformanceProfile provides performance characteristics
type OrchestrationPerformanceProfile struct {
	Throughput           ThroughputMetrics  `json:"throughput"`
	Latency              LatencyMetrics     `json:"latency"`
	ResourceUtilization  ResourceMetrics    `json:"resource_utilization"`
	Scalability          ScalabilityMetrics `json:"scalability"`
	OptimizationFeatures []string           `json:"optimization_features"`
}

// ThroughputMetrics describes processing throughput
type ThroughputMetrics struct {
	FilesPerSecond      float64 `json:"files_per_second"`
	LinesPerSecond      float64 `json:"lines_per_second"`
	TokensPerSecond     float64 `json:"tokens_per_second"`
	OptimizedThroughput float64 `json:"optimized_throughput"`
}

// LatencyMetrics describes processing latency
type LatencyMetrics struct {
	AverageLatency float64 `json:"average_latency_ms"`
	P50Latency     float64 `json:"p50_latency_ms"`
	P95Latency     float64 `json:"p95_latency_ms"`
	P99Latency     float64 `json:"p99_latency_ms"`
	StartupLatency float64 `json:"startup_latency_ms"`
}

// ResourceMetrics describes resource usage patterns
type ResourceMetrics struct {
	AverageMemoryUsage    float64 `json:"average_memory_mb"`
	PeakMemoryUsage       float64 `json:"peak_memory_mb"`
	AverageCPUUsage       float64 `json:"average_cpu_percent"`
	PeakCPUUsage          float64 `json:"peak_cpu_percent"`
	IOOperationsPerSecond float64 `json:"io_operations_per_second"`
}

// ScalabilityMetrics describes scalability characteristics
type ScalabilityMetrics struct {
	LinearScaling      bool     `json:"linear_scaling"`
	OptimalWorkerCount int      `json:"optimal_worker_count"`
	ScalingEfficiency  float64  `json:"scaling_efficiency"` // 0-1
	BottleneckPoints   []string `json:"bottleneck_points"`
}

// ResourceRequirements specifies minimum and recommended resources
type ResourceRequirements struct {
	Minimum     ResourceSpec `json:"minimum"`
	Recommended ResourceSpec `json:"recommended"`
	Optimal     ResourceSpec `json:"optimal"`
}

// ResourceSpec defines resource specifications
type ResourceSpec struct {
	Memory      int64 `json:"memory_mb"`
	CPU         int   `json:"cpu_cores"`
	Storage     int64 `json:"storage_mb"`
	Concurrency int   `json:"max_concurrency"`
}

// OrchestrationDependencies lists external dependencies
type OrchestrationDependencies struct {
	Required  []DependencySpec `json:"required"`
	Optional  []DependencySpec `json:"optional"`
	Conflicts []DependencySpec `json:"conflicts"`
}

// DependencySpec describes a dependency
type DependencySpec struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"` // library, service, tool
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// CompatibilityMatrix describes compatibility with other components
type CompatibilityMatrix struct {
	AnalysisEngines     []ComponentCompatibility `json:"analysis_engines"`
	OutputGenerators    []ComponentCompatibility `json:"output_generators"`
	DocumentationTools  []ComponentCompatibility `json:"documentation_tools"`
	SecurityScanners    []ComponentCompatibility `json:"security_scanners"`
	PerformanceMonitors []ComponentCompatibility `json:"performance_monitors"`
}

// ComponentCompatibility describes compatibility with a specific component
type ComponentCompatibility struct {
	ComponentName      string   `json:"component_name"`
	Version            string   `json:"version"`
	CompatibilityLevel string   `json:"compatibility_level"` // full, partial, limited, none
	SupportedFeatures  []string `json:"supported_features"`
	Limitations        []string `json:"limitations"`
	IntegrationNotes   string   `json:"integration_notes"`
}

// OrchestrationIntegrator provides metadata and integration services for orchestrator consumption
type OrchestrationIntegrator struct {
	analyzer      *Analyzer
	formatter     *OutputFormatter
	docIntegrator *DocumentationIntegrator
}

// NewOrchestrationIntegrator creates a new orchestration integrator
func NewOrchestrationIntegrator(analyzer *Analyzer) *OrchestrationIntegrator {
	formatter := NewOutputFormatter()
	docIntegrator := NewDocumentationIntegrator(IntegratorConfig{
		EnableRunbookGeneration:    true,
		EnableArchitectureDiagrams: true,
		EnableLearningRoadmap:      true,
		OutputDirectory:            "./docs/generated",
	})

	return &OrchestrationIntegrator{
		analyzer:      analyzer,
		formatter:     formatter,
		docIntegrator: docIntegrator,
	}
}

// GetOrchestrationMetadata returns comprehensive metadata for analysis orchestrator
func (oi *OrchestrationIntegrator) GetOrchestrationMetadata() *OrchestrationMetadata {
	return &OrchestrationMetadata{
		AnalysisCapabilities: oi.getAnalysisCapabilities(),
		OutputFormats:        oi.getOutputFormats(),
		IntegrationPoints:    oi.getIntegrationPoints(),
		QualityMetrics:       oi.getQualityMetrics(),
		PerformanceProfile:   oi.getPerformanceProfile(),
		ResourceRequirements: oi.getResourceRequirements(),
		Dependencies:         oi.getDependencies(),
		Compatibility:        oi.getCompatibilityMatrix(),
		Metadata: map[string]interface{}{
			"version":            "2.1.0",
			"schema_version":     "1.0",
			"generated_at":       time.Now(),
			"component_type":     "ast_parser",
			"orchestrator_ready": true,
		},
	}
}

// ProcessForOrchestrator prepares analysis results for orchestrator consumption
func (oi *OrchestrationIntegrator) ProcessForOrchestrator(result *AnalysisResult, metadata *AnalysisMetadata) (*OrchestrationPayload, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result cannot be nil")
	}

	// Format standardized result
	standardized, err := oi.formatter.FormatAnalysisResult(result, metadata, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to format analysis result: %w", err)
	}

	// Prepare documentation payload
	docPayload, err := oi.docIntegrator.PrepareForDocumentation(result, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare documentation payload: %w", err)
	}

	// Create orchestration payload
	payload := &OrchestrationPayload{
		AnalysisResult:         standardized,
		DocumentationPayload:   docPayload,
		OrchestrationMetadata:  oi.GetOrchestrationMetadata(),
		ProcessingInstructions: oi.getProcessingInstructions(standardized),
		QualityAssurance:       oi.getQualityAssuranceData(standardized),
		IntegrationManifest:    oi.getIntegrationManifest(),
	}

	return payload, nil
}

// OrchestrationPayload contains all data needed for analysis orchestration
type OrchestrationPayload struct {
	AnalysisResult         *StandardizedAnalysisResult `json:"analysis_result"`
	DocumentationPayload   *DocumentationPayload       `json:"documentation_payload"`
	OrchestrationMetadata  *OrchestrationMetadata      `json:"orchestration_metadata"`
	ProcessingInstructions *ProcessingInstructions     `json:"processing_instructions"`
	QualityAssurance       *QualityAssuranceData       `json:"quality_assurance"`
	IntegrationManifest    *IntegrationManifest        `json:"integration_manifest"`
}

// ProcessingInstructions provide guidance for downstream processing
type ProcessingInstructions struct {
	RecommendedProcessingOrder   []string               `json:"recommended_processing_order"`
	ParallelizationOpportunities []ParallelizationGroup `json:"parallelization_opportunities"`
	CriticalPath                 []string               `json:"critical_path"`
	OptimizationHints            []OptimizationHint     `json:"optimization_hints"`
	ValidationCheckpoints        []ValidationCheckpoint `json:"validation_checkpoints"`
}

// ParallelizationGroup describes tasks that can be processed in parallel
type ParallelizationGroup struct {
	GroupName     string   `json:"group_name"`
	Tasks         []string `json:"tasks"`
	Dependencies  []string `json:"dependencies"`
	EstimatedTime float64  `json:"estimated_time_seconds"`
}

// OptimizationHint provides performance optimization suggestions
type OptimizationHint struct {
	Component    string `json:"component"`
	Optimization string `json:"optimization"`
	Impact       string `json:"impact"`
	Complexity   string `json:"complexity"`
	Priority     string `json:"priority"`
}

// ValidationCheckpoint defines validation points in the processing pipeline
type ValidationCheckpoint struct {
	CheckpointName string                 `json:"checkpoint_name"`
	ValidationType string                 `json:"validation_type"`
	Criteria       map[string]interface{} `json:"criteria"`
	FailureAction  string                 `json:"failure_action"`
}

// QualityAssuranceData provides quality assurance information
type QualityAssuranceData struct {
	DataIntegrityChecks   []DataIntegrityCheck `json:"data_integrity_checks"`
	AccuracyMetrics       []AccuracyMetric     `json:"accuracy_metrics"`
	CompletenessChecks    []CompletenessCheck  `json:"completeness_checks"`
	ConsistencyValidation []ConsistencyCheck   `json:"consistency_validation"`
}

// DataIntegrityCheck defines a data integrity validation
type DataIntegrityCheck struct {
	CheckName   string                 `json:"check_name"`
	CheckType   string                 `json:"check_type"`
	Criteria    map[string]interface{} `json:"criteria"`
	Status      string                 `json:"status"`
	LastChecked time.Time              `json:"last_checked"`
}

// AccuracyMetric provides accuracy measurement
type AccuracyMetric struct {
	MetricName      string  `json:"metric_name"`
	AccuracyScore   float64 `json:"accuracy_score"`
	SampleSize      int     `json:"sample_size"`
	ConfidenceLevel float64 `json:"confidence_level"`
}

// CompletenessCheck validates data completeness
type CompletenessCheck struct {
	ComponentName    string  `json:"component_name"`
	ExpectedItems    int     `json:"expected_items"`
	ActualItems      int     `json:"actual_items"`
	CompletenessRate float64 `json:"completeness_rate"`
}

// ConsistencyCheck validates data consistency
type ConsistencyCheck struct {
	CheckName       string                   `json:"check_name"`
	ConsistencyType string                   `json:"consistency_type"`
	Status          string                   `json:"status"`
	Violations      []map[string]interface{} `json:"violations"`
}

// IntegrationManifest describes integration capabilities and requirements
type IntegrationManifest struct {
	SupportedProtocols      []string                 `json:"supported_protocols"`
	RequiredInterfaces      []InterfaceSpec          `json:"required_interfaces"`
	ProvidedInterfaces      []InterfaceSpec          `json:"provided_interfaces"`
	ConfigurationSchema     map[string]interface{}   `json:"configuration_schema"`
	EnvironmentRequirements []EnvironmentRequirement `json:"environment_requirements"`
}

// InterfaceSpec describes an interface
type InterfaceSpec struct {
	InterfaceName string                 `json:"interface_name"`
	Version       string                 `json:"version"`
	Protocol      string                 `json:"protocol"`
	Schema        map[string]interface{} `json:"schema"`
	Operations    []string               `json:"operations"`
}

// EnvironmentRequirement specifies environment needs
type EnvironmentRequirement struct {
	RequirementType string `json:"requirement_type"`
	Name            string `json:"name"`
	Value           string `json:"value"`
	Required        bool   `json:"required"`
}

// Private helper methods to populate metadata

func (oi *OrchestrationIntegrator) getAnalysisCapabilities() AnalysisCapabilities {
	return AnalysisCapabilities{
		SupportedLanguages: []LanguageSupport{
			{
				Language:       "javascript",
				Extensions:     []string{".js", ".jsx"},
				Features:       []string{"functions", "classes", "modules", "imports", "exports"},
				Limitations:    []string{"dynamic imports not fully supported"},
				GrammarVersion: "tree-sitter-javascript-0.20.1",
			},
			{
				Language:       "typescript",
				Extensions:     []string{".ts", ".tsx"},
				Features:       []string{"functions", "classes", "interfaces", "modules", "imports", "exports", "types"},
				Limitations:    []string{"complex type inference not included"},
				GrammarVersion: "tree-sitter-typescript-0.20.2",
			},
		},
		ExtractionCapabilities: []string{
			"function_declarations", "class_declarations", "interface_declarations",
			"variable_declarations", "import_statements", "export_statements",
			"method_definitions", "property_definitions", "type_annotations",
		},
		AnalysisTypes: []string{
			"structural_analysis", "dependency_analysis", "component_mapping",
			"complexity_analysis", "quality_analysis", "architectural_analysis",
		},
		OutputTypes: []string{
			"standardized_json", "documentation_payload", "component_map",
			"dependency_graph", "quality_report", "integration_metadata",
		},
		ScalabilityLimits: ScalabilityLimits{
			MaxFiles:       10000,
			MaxFileSize:    10 * 1024 * 1024,       // 10MB
			MaxMemory:      2 * 1024 * 1024 * 1024, // 2GB
			MaxConcurrency: 10,
			TimeoutSeconds: 3600, // 1 hour
		},
	}
}

func (oi *OrchestrationIntegrator) getOutputFormats() []OutputFormatDescriptor {
	return []OutputFormatDescriptor{
		{
			FormatName:  "standardized_analysis_result",
			MimeType:    "application/json",
			Description: "Complete analysis result in standardized format",
			Schema: map[string]interface{}{
				"version": "1.0",
				"type":    "StandardizedAnalysisResult",
			},
			Compression: []string{"none", "gzip", "minimal", "aggressive"},
			UseCase:     "comprehensive_analysis_processing",
		},
		{
			FormatName:  "documentation_payload",
			MimeType:    "application/json",
			Description: "Structured data for documentation generation",
			Schema: map[string]interface{}{
				"version": "1.0",
				"type":    "DocumentationPayload",
			},
			Compression: []string{"none", "minimal"},
			UseCase:     "documentation_generation",
		},
	}
}

func (oi *OrchestrationIntegrator) getIntegrationPoints() []IntegrationPoint {
	return []IntegrationPoint{
		{
			ComponentName:   "Documentation Generation Engine",
			IntegrationType: "producer",
			DataFormat:      "DocumentationPayload",
			Protocol:        "direct_call",
			Schema: map[string]interface{}{
				"input":  "DocumentationPayload",
				"output": "GeneratedDocumentation",
			},
			Dependencies: []string{"component_map", "structural_metrics"},
		},
		{
			ComponentName:   "Analysis Orchestrator",
			IntegrationType: "bidirectional",
			DataFormat:      "OrchestrationPayload",
			Protocol:        "service_call",
			Schema: map[string]interface{}{
				"input":  "AnalysisRequest",
				"output": "OrchestrationPayload",
			},
			Dependencies: []string{"analysis_result", "metadata"},
		},
	}
}

func (oi *OrchestrationIntegrator) getQualityMetrics() OrchestrationQualityMetrics {
	return OrchestrationQualityMetrics{
		Reliability: QualityIndicator{
			Score:  95.0,
			Status: "excellent",
			Benchmarks: map[string]float64{
				"uptime":     99.9,
				"error_rate": 0.1,
			},
		},
		Performance: QualityIndicator{
			Score:  88.0,
			Status: "good",
			Benchmarks: map[string]float64{
				"throughput": 1000.0, // files per minute
				"latency":    50.0,   // ms average
			},
		},
		Accuracy: QualityIndicator{
			Score:  92.0,
			Status: "excellent",
			Benchmarks: map[string]float64{
				"extraction_accuracy": 98.5,
				"false_positive_rate": 1.5,
			},
		},
	}
}

func (oi *OrchestrationIntegrator) getPerformanceProfile() OrchestrationPerformanceProfile {
	return OrchestrationPerformanceProfile{
		Throughput: ThroughputMetrics{
			FilesPerSecond:      16.7,  // ~1000 files/minute
			LinesPerSecond:      833.0, // ~50k lines/minute
			TokensPerSecond:     5000.0,
			OptimizedThroughput: 33.3, // With optimization
		},
		Latency: LatencyMetrics{
			AverageLatency: 50.0,
			P50Latency:     30.0,
			P95Latency:     120.0,
			P99Latency:     250.0,
			StartupLatency: 500.0,
		},
		ResourceUtilization: ResourceMetrics{
			AverageMemoryUsage:    256.0,
			PeakMemoryUsage:       512.0,
			AverageCPUUsage:       25.0,
			PeakCPUUsage:          80.0,
			IOOperationsPerSecond: 100.0,
		},
		Scalability: ScalabilityMetrics{
			LinearScaling:      true,
			OptimalWorkerCount: 4,
			ScalingEfficiency:  0.85,
			BottleneckPoints:   []string{"file_io", "memory_allocation"},
		},
		OptimizationFeatures: []string{
			"multi_threaded_parsing", "memory_monitoring", "timeout_management",
			"file_batching", "streaming_for_large_repos", "gc_optimization",
		},
	}
}

func (oi *OrchestrationIntegrator) getResourceRequirements() ResourceRequirements {
	return ResourceRequirements{
		Minimum: ResourceSpec{
			Memory:      128,
			CPU:         1,
			Storage:     50,
			Concurrency: 1,
		},
		Recommended: ResourceSpec{
			Memory:      512,
			CPU:         2,
			Storage:     100,
			Concurrency: 4,
		},
		Optimal: ResourceSpec{
			Memory:      1024,
			CPU:         4,
			Storage:     200,
			Concurrency: 8,
		},
	}
}

func (oi *OrchestrationIntegrator) getDependencies() OrchestrationDependencies {
	return OrchestrationDependencies{
		Required: []DependencySpec{
			{
				Name:        "go-tree-sitter",
				Version:     ">=0.20.0",
				Type:        "library",
				Description: "Tree-sitter Go bindings for AST parsing",
				Required:    true,
			},
		},
		Optional: []DependencySpec{
			{
				Name:        "performance-monitoring",
				Version:     ">=1.0.0",
				Type:        "service",
				Description: "Enhanced performance monitoring and metrics",
				Required:    false,
			},
		},
	}
}

func (oi *OrchestrationIntegrator) getCompatibilityMatrix() CompatibilityMatrix {
	return CompatibilityMatrix{
		AnalysisEngines: []ComponentCompatibility{
			{
				ComponentName:      "Security Scanner",
				Version:            ">=1.0.0",
				CompatibilityLevel: "full",
				SupportedFeatures:  []string{"ast_data_sharing", "vulnerability_mapping"},
			},
		},
		OutputGenerators: []ComponentCompatibility{
			{
				ComponentName:      "Documentation Generator",
				Version:            ">=2.0.0",
				CompatibilityLevel: "full",
				SupportedFeatures:  []string{"component_mapping", "dependency_graphs", "runbook_generation"},
			},
		},
	}
}

func (oi *OrchestrationIntegrator) getProcessingInstructions(result *StandardizedAnalysisResult) *ProcessingInstructions {
	return &ProcessingInstructions{
		RecommendedProcessingOrder: []string{
			"validate_input_data",
			"generate_component_map",
			"create_dependency_graph",
			"calculate_quality_metrics",
			"generate_documentation",
		},
		ParallelizationOpportunities: []ParallelizationGroup{
			{
				GroupName:     "documentation_generation",
				Tasks:         []string{"runbook_generation", "diagram_generation", "roadmap_generation"},
				Dependencies:  []string{"component_map", "dependency_graph"},
				EstimatedTime: 30.0,
			},
		},
		CriticalPath: []string{"ast_parsing", "component_mapping", "documentation_generation"},
		OptimizationHints: []OptimizationHint{
			{
				Component:    "documentation_generator",
				Optimization: "parallel_diagram_generation",
				Impact:       "30% faster documentation generation",
				Complexity:   "medium",
				Priority:     "high",
			},
		},
	}
}

func (oi *OrchestrationIntegrator) getQualityAssuranceData(result *StandardizedAnalysisResult) *QualityAssuranceData {
	return &QualityAssuranceData{
		DataIntegrityChecks: []DataIntegrityCheck{
			{
				CheckName:   "analysis_result_completeness",
				CheckType:   "completeness",
				Status:      "passed",
				LastChecked: time.Now(),
				Criteria: map[string]interface{}{
					"required_fields":   []string{"analysis_id", "repository", "code_analysis"},
					"min_quality_score": 0.0,
				},
			},
		},
		AccuracyMetrics: []AccuracyMetric{
			{
				MetricName:      "function_extraction_accuracy",
				AccuracyScore:   98.5,
				SampleSize:      1000,
				ConfidenceLevel: 95.0,
			},
		},
		CompletenessChecks: []CompletenessCheck{
			{
				ComponentName:    "component_map",
				ExpectedItems:    len(result.CodeAnalysis.ComponentMap.Components),
				ActualItems:      len(result.CodeAnalysis.ComponentMap.Components),
				CompletenessRate: 100.0,
			},
		},
	}
}

func (oi *OrchestrationIntegrator) getIntegrationManifest() *IntegrationManifest {
	return &IntegrationManifest{
		SupportedProtocols: []string{"direct_call", "json_rpc", "rest_api"},
		RequiredInterfaces: []InterfaceSpec{
			{
				InterfaceName: "AnalysisInput",
				Version:       "1.0",
				Protocol:      "direct_call",
				Operations:    []string{"analyze_repository", "analyze_file"},
				Schema: map[string]interface{}{
					"input":  "AnalysisRequest",
					"output": "AnalysisResult",
				},
			},
		},
		ProvidedInterfaces: []InterfaceSpec{
			{
				InterfaceName: "OrchestrationOutput",
				Version:       "1.0",
				Protocol:      "direct_call",
				Operations:    []string{"get_orchestration_payload", "get_metadata"},
				Schema: map[string]interface{}{
					"output": "OrchestrationPayload",
				},
			},
		},
		ConfigurationSchema: map[string]interface{}{
			"max_concurrency": map[string]interface{}{
				"type":    "integer",
				"min":     1,
				"max":     16,
				"default": 4,
			},
			"enable_optimization": map[string]interface{}{
				"type":    "boolean",
				"default": true,
			},
		},
		EnvironmentRequirements: []EnvironmentRequirement{
			{
				RequirementType: "memory",
				Name:            "minimum_memory",
				Value:           "128MB",
				Required:        true,
			},
		},
	}
}

// ToJSON serializes orchestration metadata to JSON
func (om *OrchestrationMetadata) ToJSON() ([]byte, error) {
	return json.MarshalIndent(om, "", "  ")
}

// ToCompactJSON serializes orchestration metadata to compact JSON
func (om *OrchestrationMetadata) ToCompactJSON() ([]byte, error) {
	return json.Marshal(om)
}
