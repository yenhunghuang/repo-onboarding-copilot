package ast

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Analyzer orchestrates AST parsing and dependency analysis for repositories
type Analyzer struct {
	parser               *Parser
	dependencyTracker    *DependencyTracker
	config               AnalyzerConfig
	results              map[string]*ParseResult
	componentMap         *ComponentMap
	performanceOptimizer *PerformanceOptimizer
	mu                   sync.RWMutex
}

// AnalyzerConfig configures analysis behavior
type AnalyzerConfig struct {
	ProjectRoot        string   `json:"project_root"`
	IncludePatterns    []string `json:"include_patterns"`
	ExcludePatterns    []string `json:"exclude_patterns"`
	MaxFileSize        int64    `json:"max_file_size"` // bytes
	EnableDependency   bool     `json:"enable_dependency"`
	EnableComponentMap bool     `json:"enable_component_map"`
	MaxConcurrency     int      `json:"max_concurrency"`

	// Performance optimization settings
	EnablePerformanceOptimization bool              `json:"enable_performance_optimization"`
	PerformanceConfig             PerformanceConfig `json:"performance_config"`
	UseStreamingForLargeRepos     bool              `json:"use_streaming_for_large_repos"`
}

// ComponentMap represents architectural component relationships
type ComponentMap struct {
	Components []Component          `json:"components"`
	Relations  []ComponentRelation  `json:"relations"`
	Layers     []ArchitecturalLayer `json:"layers"`
	Stats      ComponentStats       `json:"stats"`
}

// Component represents an architectural component
type Component struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`  // service, utility, component, model
	Layer        string            `json:"layer"` // presentation, business, data
	Files        []string          `json:"files"`
	Dependencies []string          `json:"dependencies"`
	Dependents   []string          `json:"dependents"`
	Exports      []string          `json:"exports"`
	Complexity   int               `json:"complexity"`
	Stability    float64           `json:"stability"` // 0-1 scale
	Metadata     map[string]string `json:"metadata"`
}

// ComponentRelation represents relationships between components
type ComponentRelation struct {
	FromComponent string            `json:"from_component"`
	ToComponent   string            `json:"to_component"`
	RelationType  string            `json:"relation_type"` // uses, extends, implements, contains
	Strength      int               `json:"strength"`      // 1-10 scale
	Metadata      map[string]string `json:"metadata"`
}

// ArchitecturalLayer represents system layers
type ArchitecturalLayer struct {
	Name        string   `json:"name"`
	Level       int      `json:"level"` // 0=lowest, higher=upper layers
	Components  []string `json:"components"`
	Description string   `json:"description"`
}

// ComponentStats contains component analysis statistics
type ComponentStats struct {
	TotalComponents   int            `json:"total_components"`
	TotalRelations    int            `json:"total_relations"`
	AvgComplexity     float64        `json:"avg_complexity"`
	LayerDistribution map[string]int `json:"layer_distribution"`
	TypeDistribution  map[string]int `json:"type_distribution"`
}

// AnalysisResult contains complete analysis results
type AnalysisResult struct {
	ProjectPath      string                     `json:"project_path"`
	FileResults      map[string]*ParseResult    `json:"file_results"`
	DependencyGraph  *ModuleGraph               `json:"dependency_graph"`
	ComponentMap     *ComponentMap              `json:"component_map"`
	ExternalPackages map[string]ExternalPackage `json:"external_packages"`
	Summary          AnalysisSummary            `json:"summary"`
}

// AnalysisSummary provides high-level analysis insights
type AnalysisSummary struct {
	TotalFiles      int               `json:"total_files"`
	TotalFunctions  int               `json:"total_functions"`
	TotalClasses    int               `json:"total_classes"`
	TotalInterfaces int               `json:"total_interfaces"`
	TotalVariables  int               `json:"total_variables"`
	Languages       map[string]int    `json:"languages"`
	Complexity      ComplexityMetrics `json:"complexity"`
	Quality         QualityMetrics    `json:"quality"`
}

// ComplexityMetrics contains code complexity measurements
type ComplexityMetrics struct {
	AverageFileSize     float64 `json:"average_file_size"`
	LargestFile         string  `json:"largest_file"`
	MostComplexFunction string  `json:"most_complex_function"`
	DependencyDepth     int     `json:"dependency_depth"`
}

// QualityMetrics contains code quality indicators
type QualityMetrics struct {
	DocumentationCoverage float64 `json:"documentation_coverage"`
	TestCoverage          float64 `json:"test_coverage"`
	CircularDependencies  int     `json:"circular_dependencies"`
	UnusedExports         int     `json:"unused_exports"`
}

// NewAnalyzer creates a new AST analyzer
func NewAnalyzer(config AnalyzerConfig) (*Analyzer, error) {
	parser, err := NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 4
	}

	if config.MaxFileSize <= 0 {
		config.MaxFileSize = 10 * 1024 * 1024 // 10MB default
	}

	// Initialize performance optimization if enabled
	var performanceOptimizer *PerformanceOptimizer
	if config.EnablePerformanceOptimization {
		// Use provided config or default
		perfConfig := config.PerformanceConfig
		if perfConfig.MaxWorkers == 0 {
			perfConfig = DefaultPerformanceConfig()
		}
		performanceOptimizer = NewPerformanceOptimizer(perfConfig)
	}

	return &Analyzer{
		parser:               parser,
		dependencyTracker:    NewDependencyTracker(),
		config:               config,
		results:              make(map[string]*ParseResult),
		componentMap:         &ComponentMap{},
		performanceOptimizer: performanceOptimizer,
	}, nil
}

// AnalyzeRepository performs complete repository analysis
func (a *Analyzer) AnalyzeRepository(ctx context.Context) (*AnalysisResult, error) {
	// Find all supported files
	files, err := a.findSupportedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	// Use performance-optimized parsing for large repositories or when enabled
	if a.shouldUsePerformanceOptimization(len(files)) {
		return a.analyzeRepositoryOptimized(ctx, files)
	}

	// Parse files concurrently (existing method)
	if err := a.parseFiles(ctx, files); err != nil {
		return nil, fmt.Errorf("failed to parse files: %w", err)
	}

	// Use the common method to build analysis result
	return a.buildAnalysisResult(nil)
}

// AnalyzeFile analyzes a single file
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (*ParseResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	if int64(len(content)) > a.config.MaxFileSize {
		return nil, fmt.Errorf("file %s exceeds maximum size limit", filePath)
	}

	result, err := a.parser.ParseFile(ctx, filePath, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	a.mu.Lock()
	a.results[filePath] = result
	a.mu.Unlock()

	return result, nil
}

// GetFileResult retrieves analysis result for a specific file
func (a *Analyzer) GetFileResult(filePath string) (*ParseResult, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result, exists := a.results[filePath]
	return result, exists
}

// GetDependencies returns dependency information for a file
func (a *Analyzer) GetDependencies(filePath string) ([]Dependency, bool) {
	return a.dependencyTracker.GetDependencies(filePath)
}

// GetDependents returns files that depend on the specified file
func (a *Analyzer) GetDependents(filePath string) []string {
	return a.dependencyTracker.GetDependents(filePath, false)
}

// Close releases analyzer resources
func (a *Analyzer) Close() error {
	if a.parser != nil {
		return a.parser.Close()
	}
	return nil
}

// shouldUsePerformanceOptimization determines if performance optimization should be used
func (a *Analyzer) shouldUsePerformanceOptimization(fileCount int) bool {
	// Use performance optimization if explicitly enabled or for large repositories
	if a.config.EnablePerformanceOptimization && a.performanceOptimizer != nil {
		return true
	}

	// Auto-enable for large repositories (>1000 files)
	if fileCount > 1000 {
		return true
	}

	// Use streaming mode setting
	return a.config.UseStreamingForLargeRepos && fileCount > 500
}

// analyzeRepositoryOptimized performs optimized repository analysis for large codebases
func (a *Analyzer) analyzeRepositoryOptimized(ctx context.Context, files []string) (*AnalysisResult, error) {
	if a.performanceOptimizer == nil {
		// Initialize performance optimizer with default config if not already done
		a.performanceOptimizer = NewPerformanceOptimizer(DefaultPerformanceConfig())
	}

	// Use performance optimizer for parsing
	perfMetrics, results, err := a.performanceOptimizer.ParseRepositoryOptimized(ctx, files, a.parser)
	if err != nil {
		return nil, fmt.Errorf("optimized parsing failed: %w", err)
	}

	// Store results in analyzer
	a.mu.Lock()
	a.results = results
	a.mu.Unlock()

	// Continue with dependency analysis and component mapping
	return a.buildAnalysisResult(perfMetrics)
}

// buildAnalysisResult constructs the final analysis result with optional performance metrics
func (a *Analyzer) buildAnalysisResult(perfMetrics *PerformanceMetrics) (*AnalysisResult, error) {
	// Analyze dependencies if enabled
	var dependencyGraph *ModuleGraph
	var externalPackages map[string]ExternalPackage

	if a.config.EnableDependency {
		// Add results to dependency tracker
		for path, result := range a.results {
			if err := a.dependencyTracker.AddParseResult(path, result); err != nil {
				return nil, fmt.Errorf("failed to add result for %s: %w", path, err)
			}
		}

		// Resolve internal dependencies
		if err := a.dependencyTracker.ResolveDependencies(a.config.ProjectRoot); err != nil {
			return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
		}

		// Build dependency graph
		var err error
		dependencyGraph, err = a.dependencyTracker.BuildModuleGraph()
		if err != nil {
			return nil, fmt.Errorf("failed to build module graph: %w", err)
		}

		externalPackages = a.dependencyTracker.GetExternalPackages()
	}

	// Generate component map if enabled
	if a.config.EnableComponentMap {
		if err := a.generateComponentMap(); err != nil {
			return nil, fmt.Errorf("failed to generate component map: %w", err)
		}
	}

	// Generate summary
	summary := a.generateSummary()

	result := &AnalysisResult{
		ProjectPath:      a.config.ProjectRoot,
		FileResults:      a.results,
		DependencyGraph:  dependencyGraph,
		ComponentMap:     a.componentMap,
		ExternalPackages: externalPackages,
		Summary:          summary,
	}

	// Add performance metrics to summary if available
	if perfMetrics != nil {
		if result.Summary.Complexity.AverageFileSize == 0 {
			result.Summary.Complexity.AverageFileSize = float64(perfMetrics.AverageProcessingTime.Nanoseconds()) / 1000000 // Convert to ms
		}
		// Additional performance metadata could be added to result.Summary.Quality or as separate field
	}

	return result, nil
}

// Private methods

func (a *Analyzer) findSupportedFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(a.config.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip excluded directories
			if a.isExcluded(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is supported and not excluded
		if a.parser.IsSupported(path) && !a.isExcluded(path) {
			// Check file size
			if info.Size() <= a.config.MaxFileSize {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

func (a *Analyzer) isExcluded(path string) bool {
	// Check common exclusions
	excludePatterns := []string{
		"node_modules", ".git", ".vscode", "dist", "build",
		"coverage", ".nyc_output", "*.min.js", "*.bundle.js",
	}

	// Add user-defined exclusions
	excludePatterns = append(excludePatterns, a.config.ExcludePatterns...)

	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

func (a *Analyzer) parseFiles(ctx context.Context, files []string) error {
	// Create worker pool for concurrent parsing
	semaphore := make(chan struct{}, a.config.MaxConcurrency)
	var wg sync.WaitGroup
	var parseErrors []error
	var errorMu sync.Mutex

	for _, filePath := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Create dedicated parser for this goroutine (thread safety)
			parser, err := NewParser()
			if err != nil {
				errorMu.Lock()
				parseErrors = append(parseErrors, fmt.Errorf("failed to create parser for %s: %w", path, err))
				errorMu.Unlock()
				return
			}
			defer parser.Close()

			// Read and parse file
			content, err := os.ReadFile(path)
			if err != nil {
				errorMu.Lock()
				parseErrors = append(parseErrors, fmt.Errorf("failed to read file %s: %w", path, err))
				errorMu.Unlock()
				return
			}

			if int64(len(content)) > a.config.MaxFileSize {
				errorMu.Lock()
				parseErrors = append(parseErrors, fmt.Errorf("file %s exceeds maximum size limit", path))
				errorMu.Unlock()
				return
			}

			result, err := parser.ParseFile(ctx, path, content)
			if err != nil {
				errorMu.Lock()
				parseErrors = append(parseErrors, fmt.Errorf("failed to parse file %s: %w", path, err))
				errorMu.Unlock()
				return
			}

			a.mu.Lock()
			a.results[path] = result
			a.mu.Unlock()

		}(filePath)
	}

	wg.Wait()

	// Return first error if any occurred
	if len(parseErrors) > 0 {
		return parseErrors[0]
	}

	return nil
}

func (a *Analyzer) generateComponentMap() error {
	// Component identification based on directory structure and naming patterns
	components := make(map[string]*Component)

	for filePath, result := range a.results {
		componentID := a.identifyComponent(filePath)

		if component, exists := components[componentID]; exists {
			// Add file to existing component
			component.Files = append(component.Files, filePath)
			component.Complexity += len(result.Functions) + len(result.Classes)
		} else {
			// Create new component
			component := &Component{
				ID:         componentID,
				Name:       a.getComponentName(componentID),
				Type:       a.getComponentType(filePath),
				Layer:      a.getComponentLayer(filePath),
				Files:      []string{filePath},
				Exports:    a.getComponentExports(result),
				Complexity: len(result.Functions) + len(result.Classes),
				Metadata:   make(map[string]string),
			}

			component.Metadata["language"] = result.Language
			components[componentID] = component
		}
	}

	// Convert map to slice
	a.componentMap.Components = make([]Component, 0, len(components))
	for _, component := range components {
		a.componentMap.Components = append(a.componentMap.Components, *component)
	}

	// Generate component relations from dependencies
	a.generateComponentRelations()

	// Calculate component statistics
	a.calculateComponentStats()

	return nil
}

func (a *Analyzer) identifyComponent(filePath string) string {
	// Simple component identification based on directory structure
	relativePath, _ := filepath.Rel(a.config.ProjectRoot, filePath)
	parts := strings.Split(relativePath, string(filepath.Separator))

	if len(parts) > 1 {
		return parts[0] // Use top-level directory as component
	}

	return "root"
}

func (a *Analyzer) getComponentName(componentID string) string {
	return strings.Title(componentID)
}

func (a *Analyzer) getComponentType(filePath string) string {
	// Determine component type based on file patterns
	if strings.Contains(filePath, "service") || strings.Contains(filePath, "api") {
		return "service"
	}
	if strings.Contains(filePath, "component") || strings.Contains(filePath, "ui") {
		return "component"
	}
	if strings.Contains(filePath, "util") || strings.Contains(filePath, "helper") {
		return "utility"
	}
	if strings.Contains(filePath, "model") || strings.Contains(filePath, "entity") {
		return "model"
	}

	return "module"
}

func (a *Analyzer) getComponentLayer(filePath string) string {
	// Determine architectural layer
	if strings.Contains(filePath, "ui") || strings.Contains(filePath, "component") || strings.Contains(filePath, "view") {
		return "presentation"
	}
	if strings.Contains(filePath, "service") || strings.Contains(filePath, "business") || strings.Contains(filePath, "logic") {
		return "business"
	}
	if strings.Contains(filePath, "data") || strings.Contains(filePath, "repository") || strings.Contains(filePath, "model") {
		return "data"
	}

	return "core"
}

func (a *Analyzer) getComponentExports(result *ParseResult) []string {
	exports := make([]string, 0)
	for _, export := range result.Exports {
		if export.ExportType == "default" && export.Name != "" {
			exports = append(exports, export.Name)
		} else {
			exports = append(exports, export.Specifiers...)
		}
	}
	return exports
}

func (a *Analyzer) generateComponentRelations() {
	// Generate relations based on dependency graph
	// This is a simplified implementation
	a.componentMap.Relations = []ComponentRelation{}
}

func (a *Analyzer) calculateComponentStats() {
	stats := &a.componentMap.Stats
	stats.TotalComponents = len(a.componentMap.Components)
	stats.TotalRelations = len(a.componentMap.Relations)
	stats.LayerDistribution = make(map[string]int)
	stats.TypeDistribution = make(map[string]int)

	totalComplexity := 0
	for _, component := range a.componentMap.Components {
		totalComplexity += component.Complexity
		stats.LayerDistribution[component.Layer]++
		stats.TypeDistribution[component.Type]++
	}

	if stats.TotalComponents > 0 {
		stats.AvgComplexity = float64(totalComplexity) / float64(stats.TotalComponents)
	}
}

func (a *Analyzer) generateSummary() AnalysisSummary {
	summary := AnalysisSummary{
		Languages: make(map[string]int),
	}

	totalFunctions := 0
	totalClasses := 0
	totalInterfaces := 0
	totalVariables := 0

	for _, result := range a.results {
		summary.TotalFiles++
		summary.Languages[result.Language]++

		totalFunctions += len(result.Functions)
		totalClasses += len(result.Classes)
		totalInterfaces += len(result.Interfaces)
		totalVariables += len(result.Variables)
	}

	summary.TotalFunctions = totalFunctions
	summary.TotalClasses = totalClasses
	summary.TotalInterfaces = totalInterfaces
	summary.TotalVariables = totalVariables

	return summary
}
