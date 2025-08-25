package ast

import (
	"encoding/json"
	"fmt"
	"time"
)

// OutputFormatter provides standardized output formatting for AST analysis results
type OutputFormatter struct {
	options FormatterOptions
}

// FormatterOptions configures output formatting behavior
type FormatterOptions struct {
	PrettyPrint               bool   `json:"pretty_print"`
	IncludeMetadata           bool   `json:"include_metadata"`
	IncludePerformanceMetrics bool   `json:"include_performance_metrics"`
	CompressionLevel          int    `json:"compression_level"` // 0=none, 1=minimal, 2=aggressive
	OutputFormat              string `json:"output_format"`     // json, msgpack, yaml
}

// StandardizedAnalysisResult represents the core analysis result schema
// compatible with the architecture/data-flow-architecture.md specification
type StandardizedAnalysisResult struct {
	AnalysisID   string              `json:"analysis_id"`
	Repository   RepositoryInfo      `json:"repository"`
	Metadata     AnalysisMetadata    `json:"analysis_metadata"`
	CodeAnalysis CodeAnalysisSection `json:"code_analysis"`
	Dependencies DependenciesSection `json:"dependencies"`
	Security     SecuritySection     `json:"security_findings"`
	Quality      QualitySection      `json:"quality_analysis"`
	Performance  *PerformanceSection `json:"performance_analysis,omitempty"`
}

// RepositoryInfo contains repository metadata
type RepositoryInfo struct {
	URL        string   `json:"url,omitempty"`
	Path       string   `json:"path"`
	CommitHash string   `json:"commit_hash,omitempty"`
	SizeBytes  int64    `json:"size_bytes"`
	Languages  []string `json:"languages"`
	Frameworks []string `json:"frameworks"`
	FileCount  int      `json:"file_count"`
}

// AnalysisMetadata contains analysis execution information
type AnalysisMetadata struct {
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	DurationSeconds float64   `json:"duration_seconds"`
	Version         string    `json:"version"`
	EngineVersion   string    `json:"engine_version"`
	ProcessedFiles  int       `json:"processed_files"`
	FailedFiles     int       `json:"failed_files"`
}

// CodeAnalysisSection contains AST and structural analysis results
type CodeAnalysisSection struct {
	ASTData           interface{}       `json:"ast_data"`           // Raw AST analysis results
	ComponentMap      *ComponentMap     `json:"component_map"`      // Architectural components
	ComplexityMetrics ComplexityMetrics `json:"complexity_metrics"` // Code complexity measurements
	QualityScore      float64           `json:"quality_score"`      // Overall quality rating (0-100)
	StructuralMetrics StructuralMetrics `json:"structural_metrics"` // Code structure analysis
}

// StructuralMetrics provides detailed code structure analysis
type StructuralMetrics struct {
	TotalFunctions    int            `json:"total_functions"`
	TotalClasses      int            `json:"total_classes"`
	TotalInterfaces   int            `json:"total_interfaces"`
	TotalVariables    int            `json:"total_variables"`
	TotalExports      int            `json:"total_exports"`
	TotalImports      int            `json:"total_imports"`
	FunctionsByFile   map[string]int `json:"functions_by_file"`
	ClassesByFile     map[string]int `json:"classes_by_file"`
	ComplexityByFile  map[string]int `json:"complexity_by_file"`
	LanguageBreakdown map[string]int `json:"language_breakdown"`
}

// DependenciesSection contains dependency analysis results
type DependenciesSection struct {
	Direct       []DependencyInfo       `json:"direct"`
	Transitive   []DependencyInfo       `json:"transitive"`
	External     []ExternalPackage      `json:"external"`
	Internal     []InternalDependency   `json:"internal"`
	Circular     []CircularDependency   `json:"circular_dependencies"`
	GraphMetrics DependencyGraphMetrics `json:"graph_metrics"`
}

// DependencyInfo represents a dependency relationship
type DependencyInfo struct {
	Name       string            `json:"name"`
	Version    string            `json:"version,omitempty"`
	Type       string            `json:"type"`   // npm, built-in, local, etc.
	Source     string            `json:"source"` // file that declares the dependency
	UsageCount int               `json:"usage_count"`
	LastUsed   time.Time         `json:"last_used,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// InternalDependency represents internal module dependencies
type InternalDependency struct {
	FromModule   string `json:"from_module"`
	ToModule     string `json:"to_module"`
	ImportCount  int    `json:"import_count"`
	Relationship string `json:"relationship"` // uses, extends, implements
}

// CircularDependency represents circular dependency chains
type CircularDependency struct {
	Chain       []string `json:"chain"`
	Severity    string   `json:"severity"`    // low, medium, high
	Impact      string   `json:"impact"`      // description of potential issues
	Suggestions []string `json:"suggestions"` // remediation suggestions
}

// DependencyGraphMetrics provides dependency graph statistics
type DependencyGraphMetrics struct {
	TotalNodes           int     `json:"total_nodes"`
	TotalEdges           int     `json:"total_edges"`
	MaxDepth             int     `json:"max_depth"`
	AverageDepth         float64 `json:"average_depth"`
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	Modularity           float64 `json:"modularity"` // 0-1, higher = better modularization
}

// SecuritySection contains security analysis results (placeholder for future integration)
type SecuritySection struct {
	RiskScore        float64           `json:"risk_score"` // 0-100
	Vulnerabilities  []SecurityFinding `json:"vulnerabilities"`
	Recommendations  []string          `json:"recommendations"`
	ComplianceStatus map[string]string `json:"compliance_status"`
}

// SecurityFinding represents a security issue (placeholder)
type SecurityFinding struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Remediation string `json:"remediation"`
}

// QualitySection contains code quality analysis
type QualitySection struct {
	OverallScore          float64        `json:"overall_score"`          // 0-100
	DocumentationCoverage float64        `json:"documentation_coverage"` // 0-100
	TestCoverage          float64        `json:"test_coverage"`          // 0-100
	CodeConsistency       float64        `json:"code_consistency"`       // 0-100
	Maintainability       float64        `json:"maintainability"`        // 0-100
	Issues                []QualityIssue `json:"issues"`
	Recommendations       []string       `json:"recommendations"`
	TrendAnalysis         QualityTrend   `json:"trend_analysis"`
}

// QualityIssue represents a code quality issue
type QualityIssue struct {
	Type        string `json:"type"`     // complexity, duplication, style, etc.
	Severity    string `json:"severity"` // low, medium, high
	Description string `json:"description"`
	Location    string `json:"location"`
	Suggestion  string `json:"suggestion"`
}

// QualityTrend tracks quality metrics over time (placeholder for future versions)
type QualityTrend struct {
	Direction      string  `json:"direction"`       // improving, declining, stable
	ChangeRate     float64 `json:"change_rate"`     // percentage change
	PredictedScore float64 `json:"predicted_score"` // projected future score
}

// PerformanceSection contains performance analysis results
type PerformanceSection struct {
	ParseTime           float64                    `json:"parse_time_seconds"`
	MemoryUsage         int64                      `json:"memory_usage_bytes"`
	FilesPerSecond      float64                    `json:"files_per_second"`
	OptimizationUsed    bool                       `json:"optimization_used"`
	ResourceUtilization ResourceUtilizationMetrics `json:"resource_utilization"`
	BottleneckAnalysis  []PerformanceBottleneck    `json:"bottleneck_analysis"`
}

// ResourceUtilizationMetrics tracks resource usage during analysis
type ResourceUtilizationMetrics struct {
	MaxMemoryUsed int64   `json:"max_memory_used_bytes"`
	AvgCPUUsage   float64 `json:"avg_cpu_usage_percent"`
	PeakCPUUsage  float64 `json:"peak_cpu_usage_percent"`
	IOOperations  int64   `json:"io_operations"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
}

// PerformanceBottleneck identifies performance issues
type PerformanceBottleneck struct {
	Component    string  `json:"component"`
	Issue        string  `json:"issue"`
	Impact       string  `json:"impact"`
	Severity     string  `json:"severity"`
	Suggestion   string  `json:"suggestion"`
	TimeImpact   float64 `json:"time_impact_seconds"`
	MemoryImpact int64   `json:"memory_impact_bytes"`
}

// NewOutputFormatter creates a new output formatter with default options
func NewOutputFormatter(opts ...FormatterOptions) *OutputFormatter {
	options := FormatterOptions{
		PrettyPrint:               true,
		IncludeMetadata:           true,
		IncludePerformanceMetrics: false,
		CompressionLevel:          0,
		OutputFormat:              "json",
	}

	if len(opts) > 0 {
		options = opts[0]
	}

	return &OutputFormatter{
		options: options,
	}
}

// FormatAnalysisResult converts internal AnalysisResult to standardized format
func (of *OutputFormatter) FormatAnalysisResult(result *AnalysisResult, metadata *AnalysisMetadata, perfMetrics *PerformanceMetrics) (*StandardizedAnalysisResult, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result cannot be nil")
	}

	// Generate unique analysis ID
	analysisID := fmt.Sprintf("ast-analysis-%d", time.Now().Unix())

	// Prepare metadata
	if metadata == nil {
		metadata = &AnalysisMetadata{
			StartedAt:     time.Now(),
			CompletedAt:   time.Now(),
			Version:       "1.0.0",
			EngineVersion: "ast-parser-v2.1",
		}
	}

	// Calculate duration
	metadata.DurationSeconds = metadata.CompletedAt.Sub(metadata.StartedAt).Seconds()

	// Create standardized result
	standardized := &StandardizedAnalysisResult{
		AnalysisID:   analysisID,
		Repository:   of.buildRepositoryInfo(result),
		Metadata:     *metadata,
		CodeAnalysis: of.buildCodeAnalysisSection(result),
		Dependencies: of.buildDependenciesSection(result),
		Security:     of.buildSecuritySection(result),
		Quality:      of.buildQualitySection(result),
	}

	// Add performance metrics if available and requested
	if of.options.IncludePerformanceMetrics && perfMetrics != nil {
		standardized.Performance = of.buildPerformanceSection(perfMetrics)
	}

	return standardized, nil
}

// ToJSON serializes the standardized result to JSON
func (of *OutputFormatter) ToJSON(result *StandardizedAnalysisResult) ([]byte, error) {
	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	if of.options.PrettyPrint {
		return json.MarshalIndent(result, "", "  ")
	}

	return json.Marshal(result)
}

// ToCompactJSON serializes with minimal formatting for efficiency
func (of *OutputFormatter) ToCompactJSON(result *StandardizedAnalysisResult) ([]byte, error) {
	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	// Create a compressed version based on compression level
	compressed := of.compressResult(result)
	return json.Marshal(compressed)
}

// CreateDocumentationPayload creates a payload specifically for Documentation Generation Engine
func (of *OutputFormatter) CreateDocumentationPayload(result *StandardizedAnalysisResult) (*DocumentationPayload, error) {
	if result == nil {
		return nil, fmt.Errorf("result cannot be nil")
	}

	payload := &DocumentationPayload{
		AnalysisID:      result.AnalysisID,
		ProjectPath:     result.Repository.Path,
		GeneratedAt:     time.Now(),
		ComponentMap:    result.CodeAnalysis.ComponentMap,
		StructuralData:  &result.CodeAnalysis.StructuralMetrics,
		DependencyGraph: result.Dependencies,
		QualityMetrics:  &result.Quality,
		Metadata: map[string]interface{}{
			"languages":        result.Repository.Languages,
			"frameworks":       result.Repository.Frameworks,
			"file_count":       result.Repository.FileCount,
			"processed_files":  result.Metadata.ProcessedFiles,
			"analysis_version": result.Metadata.Version,
		},
	}

	return payload, nil
}

// Private helper methods

func (of *OutputFormatter) buildRepositoryInfo(result *AnalysisResult) RepositoryInfo {
	// Calculate repository size and detect frameworks
	var sizeBytes int64
	languages := make(map[string]bool)
	frameworks := []string{} // Framework detection logic would be implemented here

	for _, fileResult := range result.FileResults {
		// Estimate file size from metadata or use a default estimation
		if metadata, ok := fileResult.Metadata["file_size"]; ok {
			if size, ok := metadata.(int64); ok {
				sizeBytes += size
			} else {
				sizeBytes += 1024 // Default estimation
			}
		} else {
			sizeBytes += 1024 // Default estimation
		}
		languages[fileResult.Language] = true
	}

	// Convert languages map to slice
	langSlice := make([]string, 0, len(languages))
	for lang := range languages {
		langSlice = append(langSlice, lang)
	}

	return RepositoryInfo{
		Path:       result.ProjectPath,
		SizeBytes:  sizeBytes,
		Languages:  langSlice,
		Frameworks: frameworks,
		FileCount:  len(result.FileResults),
	}
}

func (of *OutputFormatter) buildCodeAnalysisSection(result *AnalysisResult) CodeAnalysisSection {
	// Build structural metrics
	structuralMetrics := of.calculateStructuralMetrics(result)

	// Calculate overall quality score
	qualityScore := of.calculateQualityScore(result)

	return CodeAnalysisSection{
		ASTData:           result.FileResults, // Raw AST data
		ComponentMap:      result.ComponentMap,
		ComplexityMetrics: result.Summary.Complexity,
		QualityScore:      qualityScore,
		StructuralMetrics: structuralMetrics,
	}
}

func (of *OutputFormatter) calculateStructuralMetrics(result *AnalysisResult) StructuralMetrics {
	metrics := StructuralMetrics{
		TotalFunctions:    result.Summary.TotalFunctions,
		TotalClasses:      result.Summary.TotalClasses,
		TotalInterfaces:   result.Summary.TotalInterfaces,
		TotalVariables:    result.Summary.TotalVariables,
		FunctionsByFile:   make(map[string]int),
		ClassesByFile:     make(map[string]int),
		ComplexityByFile:  make(map[string]int),
		LanguageBreakdown: result.Summary.Languages,
	}

	// Calculate per-file metrics
	for filePath, fileResult := range result.FileResults {
		metrics.FunctionsByFile[filePath] = len(fileResult.Functions)
		metrics.ClassesByFile[filePath] = len(fileResult.Classes)
		metrics.ComplexityByFile[filePath] = len(fileResult.Functions) + len(fileResult.Classes) + len(fileResult.Interfaces)

		// Count exports and imports
		metrics.TotalExports += len(fileResult.Exports)
		metrics.TotalImports += len(fileResult.Imports)
	}

	return metrics
}

func (of *OutputFormatter) calculateQualityScore(result *AnalysisResult) float64 {
	// Simple quality scoring algorithm
	// This would be enhanced with more sophisticated metrics in a real implementation
	score := 100.0

	// Deduct points for various issues
	if result.Summary.Quality.CircularDependencies > 0 {
		score -= float64(result.Summary.Quality.CircularDependencies) * 5.0
	}

	if result.Summary.Quality.UnusedExports > 0 {
		score -= float64(result.Summary.Quality.UnusedExports) * 2.0
	}

	// Boost score for good documentation coverage
	if result.Summary.Quality.DocumentationCoverage > 0.8 {
		score += 5.0
	}

	// Ensure score stays within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func (of *OutputFormatter) buildDependenciesSection(result *AnalysisResult) DependenciesSection {
	section := DependenciesSection{
		External: make([]ExternalPackage, 0, len(result.ExternalPackages)),
		Internal: []InternalDependency{},
		Circular: []CircularDependency{},
	}

	// Convert external packages
	for _, pkg := range result.ExternalPackages {
		section.External = append(section.External, pkg)
	}

	// Build dependency graph metrics
	if result.DependencyGraph != nil {
		section.GraphMetrics = DependencyGraphMetrics{
			TotalNodes: len(result.DependencyGraph.Nodes),
			TotalEdges: len(result.DependencyGraph.Edges),
			// Additional metrics would be calculated here
		}
	}

	return section
}

func (of *OutputFormatter) buildSecuritySection(result *AnalysisResult) SecuritySection {
	// Placeholder - security analysis would be integrated from security components
	return SecuritySection{
		RiskScore:       50.0, // Default neutral score
		Vulnerabilities: []SecurityFinding{},
		Recommendations: []string{
			"Run security scanning tools",
			"Validate input handling",
			"Review dependency vulnerabilities",
		},
		ComplianceStatus: map[string]string{
			"owasp": "not-assessed",
		},
	}
}

func (of *OutputFormatter) buildQualitySection(result *AnalysisResult) QualitySection {
	return QualitySection{
		OverallScore:          of.calculateQualityScore(result),
		DocumentationCoverage: result.Summary.Quality.DocumentationCoverage * 100,
		TestCoverage:          result.Summary.Quality.TestCoverage * 100,
		CodeConsistency:       85.0, // Placeholder
		Maintainability:       75.0, // Placeholder
		Issues:                []QualityIssue{},
		Recommendations: []string{
			"Improve documentation coverage",
			"Add unit tests for critical functions",
			"Reduce circular dependencies",
		},
		TrendAnalysis: QualityTrend{
			Direction:      "stable",
			ChangeRate:     0.0,
			PredictedScore: of.calculateQualityScore(result),
		},
	}
}

func (of *OutputFormatter) buildPerformanceSection(metrics *PerformanceMetrics) *PerformanceSection {
	return &PerformanceSection{
		ParseTime:        metrics.TotalDuration.Seconds(),
		MemoryUsage:      metrics.PeakMemoryUsage,
		FilesPerSecond:   metrics.ThroughputFPS,
		OptimizationUsed: true, // Optimization is used when performance metrics are available
		ResourceUtilization: ResourceUtilizationMetrics{
			MaxMemoryUsed: metrics.PeakMemoryUsage,
			AvgCPUUsage:   0.0, // Would be calculated from system metrics
			PeakCPUUsage:  0.0, // Would be calculated from system metrics
			IOOperations:  metrics.TotalFilesProcessed,
			CacheHitRate:  0.0, // Would be calculated from caching metrics
		},
		BottleneckAnalysis: []PerformanceBottleneck{},
	}
}

func (of *OutputFormatter) compressResult(result *StandardizedAnalysisResult) interface{} {
	switch of.options.CompressionLevel {
	case 1: // Minimal compression
		return of.minimalCompression(result)
	case 2: // Aggressive compression
		return of.aggressiveCompression(result)
	default:
		return result
	}
}

func (of *OutputFormatter) minimalCompression(result *StandardizedAnalysisResult) interface{} {
	// Remove detailed metadata but keep core structure
	compressed := *result
	compressed.CodeAnalysis.ASTData = nil // Remove raw AST data
	return &compressed
}

func (of *OutputFormatter) aggressiveCompression(result *StandardizedAnalysisResult) interface{} {
	// Keep only essential metrics
	return map[string]interface{}{
		"analysis_id": result.AnalysisID,
		"repository": map[string]interface{}{
			"path":       result.Repository.Path,
			"languages":  result.Repository.Languages,
			"file_count": result.Repository.FileCount,
		},
		"summary": map[string]interface{}{
			"quality_score":   result.CodeAnalysis.QualityScore,
			"total_functions": result.CodeAnalysis.StructuralMetrics.TotalFunctions,
			"total_classes":   result.CodeAnalysis.StructuralMetrics.TotalClasses,
			"component_count": len(result.CodeAnalysis.ComponentMap.Components),
		},
	}
}
