package analysis

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CycleDetector detects circular dependencies and dependency cycles in the codebase
type CycleDetector struct {
	componentIdentifier *ComponentIdentifier
	dependencyGraph     map[string][]string // file -> [dependencies]
	visitedNodes        map[string]bool
	recursionStack      map[string]bool
	cycles              []DependencyCycle
}

// DependencyCycle represents a circular dependency chain
type DependencyCycle struct {
	ID          string                 `json:"id"`
	Type        CycleType              `json:"type"`
	Severity    CycleSeverity          `json:"severity"`
	Files       []string               `json:"files"`
	Length      int                    `json:"length"`
	Description string                 `json:"description"`
	Impact      CycleImpact            `json:"impact"`
	Resolution  []ResolutionStrategy   `json:"resolution"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CycleType represents different types of dependency cycles
type CycleType string

const (
	ImportCycle    CycleType = "import"    // Direct import/require cycles
	ComponentCycle CycleType = "component" // React component dependency cycles
	ModuleCycle    CycleType = "module"    // Module-level circular dependencies
	ServiceCycle   CycleType = "service"   // Service layer circular dependencies
	UtilityCycle   CycleType = "utility"   // Utility function circular dependencies
	TypeCycle      CycleType = "type"      // TypeScript type/interface cycles
)

// CycleSeverity represents the severity of a dependency cycle
type CycleSeverity string

const (
	CriticalSeverity CycleSeverity = "critical" // Blocks build or causes runtime errors
	HighSeverity     CycleSeverity = "high"     // Significant architecture issues
	MediumSeverity   CycleSeverity = "medium"   // Maintainability concerns
	LowSeverity      CycleSeverity = "low"      // Minor coupling issues
)

// CycleImpact describes the impact of a dependency cycle
type CycleImpact struct {
	BuildImpact           string   `json:"build_impact"`           // Impact on build process
	RuntimeImpact         string   `json:"runtime_impact"`         // Runtime performance impact
	MaintainabilityImpact string   `json:"maintainability_impact"` // Code maintainability impact
	TestabilityImpact     string   `json:"testability_impact"`     // Impact on testing
	RiskFactors           []string `json:"risk_factors"`           // Specific risk factors
}

// ResolutionStrategy provides strategies for resolving dependency cycles
type ResolutionStrategy struct {
	Strategy    string `json:"strategy"`    // Strategy name
	Description string `json:"description"` // Detailed description
	Priority    int    `json:"priority"`    // Priority (1=highest)
	Effort      string `json:"effort"`      // Required effort (low, medium, high)
}

// CycleStats provides statistics about dependency cycles
type CycleStats struct {
	TotalCycles          int                   `json:"total_cycles"`
	CyclesBySeverity     map[CycleSeverity]int `json:"cycles_by_severity"`
	CyclesByType         map[CycleType]int     `json:"cycles_by_type"`
	AverageCycleLength   float64               `json:"average_cycle_length"`
	MostProblematicFiles []string              `json:"most_problematic_files"`
	CycleComplexityScore float64               `json:"cycle_complexity_score"`
}

// NewCycleDetector creates a new cycle detector
func NewCycleDetector(componentIdentifier *ComponentIdentifier) *CycleDetector {
	return &CycleDetector{
		componentIdentifier: componentIdentifier,
		dependencyGraph:     make(map[string][]string),
		visitedNodes:        make(map[string]bool),
		recursionStack:      make(map[string]bool),
		cycles:              make([]DependencyCycle, 0),
	}
}

// DetectCycles performs comprehensive cycle detection analysis
func (cd *CycleDetector) DetectCycles(filePath, content string) error {
	// Extract dependencies from the file
	dependencies, err := cd.extractDependencies(filePath, content)
	if err != nil {
		return fmt.Errorf("error extracting dependencies: %w", err)
	}

	// Add to dependency graph
	cd.dependencyGraph[filePath] = dependencies

	return nil
}

// AnalyzeCycles performs cycle detection using DFS algorithm
func (cd *CycleDetector) AnalyzeCycles() error {
	// Reset analysis state
	cd.visitedNodes = make(map[string]bool)
	cd.recursionStack = make(map[string]bool)
	cd.cycles = make([]DependencyCycle, 0)

	// Build a complete node set including all dependencies
	allNodes := make(map[string]bool)
	for file := range cd.dependencyGraph {
		allNodes[file] = true
		for _, dep := range cd.dependencyGraph[file] {
			allNodes[dep] = true
		}
	}

	// Perform DFS from each unvisited node
	for file := range allNodes {
		if !cd.visitedNodes[file] {
			cd.dfsDetectCycles(file, []string{})
		}
	}

	// Analyze and categorize detected cycles
	cd.categorizeCycles()

	return nil
}

// dfsDetectCycles performs depth-first search to detect cycles
func (cd *CycleDetector) dfsDetectCycles(file string, path []string) {
	cd.visitedNodes[file] = true
	cd.recursionStack[file] = true
	path = append(path, file)

	// Check all dependencies - only if they exist in the graph
	if dependencies, exists := cd.dependencyGraph[file]; exists {
		for _, dependency := range dependencies {
			if cd.recursionStack[dependency] {
				// Found a cycle - extract the cycle path
				cycleStart := cd.findCycleStart(dependency, path)
				if cycleStart != -1 {
					cyclePath := path[cycleStart:]
					cyclePath = append(cyclePath, dependency) // Close the cycle
					cd.recordCycle(cyclePath)
				}
			} else if !cd.visitedNodes[dependency] {
				cd.dfsDetectCycles(dependency, path)
			}
		}
	}

	cd.recursionStack[file] = false
}

// findCycleStart finds the start of the cycle in the path
func (cd *CycleDetector) findCycleStart(target string, path []string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == target {
			return i
		}
	}
	return -1
}

// recordCycle records a detected cycle
func (cd *CycleDetector) recordCycle(cyclePath []string) {
	// Avoid duplicate cycles by normalizing the path
	normalizedPath := cd.normalizeCyclePath(cyclePath)

	// Check if this cycle already exists
	for _, existingCycle := range cd.cycles {
		if cd.compareCyclePaths(existingCycle.Files, normalizedPath) {
			return // Duplicate cycle
		}
	}

	// Create cycle record
	cycle := DependencyCycle{
		ID:          cd.generateCycleID(normalizedPath),
		Files:       normalizedPath,
		Length:      len(normalizedPath) - 1, // Subtract 1 for the closing duplicate
		Description: cd.generateCycleDescription(normalizedPath),
		Metadata: map[string]interface{}{
			"detection_method": "dfs",
			"cycle_path":       strings.Join(normalizedPath, " -> "),
		},
	}

	cd.cycles = append(cd.cycles, cycle)
}

// extractDependencies extracts dependencies from file content
func (cd *CycleDetector) extractDependencies(filePath, content string) ([]string, error) {
	dependencies := make([]string, 0)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract import/require dependencies
		if deps := cd.extractImportDependencies(line, filePath); len(deps) > 0 {
			dependencies = append(dependencies, deps...)
		}
	}

	// Remove duplicates and self-references
	dependencies = cd.deduplicateDependencies(dependencies, filePath)

	return dependencies, nil
}

// extractImportDependencies extracts dependencies from import/require statements
func (cd *CycleDetector) extractImportDependencies(line string, currentFile string) []string {
	dependencies := make([]string, 0)

	// ES6 import patterns
	if strings.Contains(line, "import ") && strings.Contains(line, "from ") {
		if dep := cd.extractFromClause(line, currentFile); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	// CommonJS require patterns
	if strings.Contains(line, "require(") {
		if dep := cd.extractRequirePath(line, currentFile); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	// Dynamic imports
	if strings.Contains(line, "import(") {
		if dep := cd.extractDynamicImportPath(line, currentFile); dep != "" {
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies
}

// extractFromClause extracts the module path from 'from' clause
func (cd *CycleDetector) extractFromClause(line string, currentFile string) string {
	parts := strings.Split(line, "from ")
	if len(parts) < 2 {
		return ""
	}

	path := strings.TrimSpace(parts[1])
	path = strings.Trim(path, "\"'`;")

	// Only track relative imports for cycle detection
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return cd.resolvePath(path, currentFile)
	}

	return ""
}

// extractRequirePath extracts the path from require() calls
func (cd *CycleDetector) extractRequirePath(line string, currentFile string) string {
	start := strings.Index(line, "require(")
	if start == -1 {
		return ""
	}

	start += 8 // Move past "require("
	end := strings.Index(line[start:], ")")
	if end == -1 {
		return ""
	}

	path := strings.TrimSpace(line[start : start+end])
	path = strings.Trim(path, "\"'`")

	// Only track relative imports for cycle detection
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return cd.resolvePath(path, currentFile)
	}

	return ""
}

// extractDynamicImportPath extracts the path from dynamic import() calls
func (cd *CycleDetector) extractDynamicImportPath(line string, currentFile string) string {
	start := strings.Index(line, "import(")
	if start == -1 {
		return ""
	}

	start += 7 // Move past "import("
	end := strings.Index(line[start:], ")")
	if end == -1 {
		return ""
	}

	path := strings.TrimSpace(line[start : start+end])
	path = strings.Trim(path, "\"'`")

	// Only track relative imports for cycle detection
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return cd.resolvePath(path, currentFile)
	}

	return ""
}

// resolvePath resolves relative paths to absolute paths
func (cd *CycleDetector) resolvePath(relativePath string, currentFile string) string {
	// For testing purposes, we need to resolve relative paths correctly
	// This is a simplified implementation

	if strings.HasPrefix(relativePath, "./") {
		// Same directory - replace ./ with current file's directory
		dir := cd.getDirectory(currentFile)
		resolved := dir + "/" + relativePath[2:]
		return cd.normalizeAndAddExtension(resolved)
	}

	if strings.HasPrefix(relativePath, "../") {
		// Parent directory - need to resolve relative to current file
		dir := cd.getDirectory(currentFile)
		// For simplicity, resolve .. by going up one level
		parentDir := cd.getParentDirectory(dir)
		resolved := parentDir + "/" + relativePath[3:]
		return cd.normalizeAndAddExtension(resolved)
	}

	return cd.normalizeAndAddExtension(relativePath)
}

// getDirectory extracts the directory from a file path
func (cd *CycleDetector) getDirectory(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// getParentDirectory gets the parent directory
func (cd *CycleDetector) getParentDirectory(dir string) string {
	parts := strings.Split(dir, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// normalizeAndAddExtension normalizes path and adds extension if needed
func (cd *CycleDetector) normalizeAndAddExtension(path string) string {
	// Clean up double slashes
	path = strings.ReplaceAll(path, "//", "/")

	// If path already has an extension, return as-is
	if strings.Contains(path, ".") {
		return path
	}

	// Add extension if not present (for cycle detection matching)
	// Check if we're in a types directory, prefer .d.ts
	if strings.Contains(path, "/types/") {
		return path + ".d.ts"
	}
	return path + ".js"
}

// deduplicateDependencies removes duplicates and self-references
func (cd *CycleDetector) deduplicateDependencies(dependencies []string, filePath string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, dep := range dependencies {
		if !seen[dep] && dep != filePath && !cd.isSelfReference(dep, filePath) {
			seen[dep] = true
			result = append(result, dep)
		}
	}

	return result
}

// isSelfReference checks if a dependency is a self-reference
func (cd *CycleDetector) isSelfReference(dependency, filePath string) bool {
	// Extract filename from both paths for comparison
	depBase := cd.getBaseName(dependency)
	fileBase := cd.getBaseName(filePath)

	return depBase == fileBase
}

// getBaseName extracts the base name from a file path
func (cd *CycleDetector) getBaseName(path string) string {
	// Remove directory path
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]

	// Remove extension
	if dotIndex := strings.LastIndex(name, "."); dotIndex != -1 {
		name = name[:dotIndex]
	}

	return name
}

// categorizeCycles analyzes and categorizes detected cycles
func (cd *CycleDetector) categorizeCycles() {
	for i := range cd.cycles {
		cycle := &cd.cycles[i]

		// Determine cycle type
		cycle.Type = cd.determineCycleType(cycle.Files)

		// Assess severity
		cycle.Severity = cd.assessCycleSeverity(cycle)

		// Analyze impact
		cycle.Impact = cd.analyzeCycleImpact(cycle)

		// Generate resolution strategies
		cycle.Resolution = cd.generateResolutionStrategies(cycle)
	}
}

// determineCycleType determines the type of dependency cycle
func (cd *CycleDetector) determineCycleType(files []string) CycleType {
	// Analyze file types and patterns to determine cycle type
	hasComponents := false
	hasServices := false
	hasUtilities := false
	hasTypes := false

	for _, file := range files {
		if strings.Contains(file, "component") || strings.Contains(file, "Component") {
			hasComponents = true
		}
		if strings.Contains(file, "service") || strings.Contains(file, "Service") {
			hasServices = true
		}
		if strings.Contains(file, "util") || strings.Contains(file, "helper") {
			hasUtilities = true
		}
		if strings.Contains(file, "type") || strings.Contains(file, ".d.ts") {
			hasTypes = true
		}
	}

	// Prioritize cycle type based on analysis
	if hasTypes {
		return TypeCycle
	}
	if hasComponents {
		return ComponentCycle
	}
	if hasServices {
		return ServiceCycle
	}
	if hasUtilities {
		return UtilityCycle
	}

	return ImportCycle // Default to import cycle
}

// assessCycleSeverity assesses the severity of a dependency cycle
func (cd *CycleDetector) assessCycleSeverity(cycle *DependencyCycle) CycleSeverity {
	// Base severity on cycle length and type
	length := cycle.Length
	cycleType := cycle.Type

	// Critical severity conditions
	if length >= 5 || cycleType == TypeCycle {
		return CriticalSeverity
	}

	// High severity conditions
	if length >= 3 || cycleType == ComponentCycle || cycleType == ServiceCycle {
		return HighSeverity
	}

	// Medium severity conditions
	if length >= 2 || cycleType == ModuleCycle {
		return MediumSeverity
	}

	return LowSeverity
}

// analyzeCycleImpact analyzes the impact of a dependency cycle
func (cd *CycleDetector) analyzeCycleImpact(cycle *DependencyCycle) CycleImpact {
	impact := CycleImpact{
		RiskFactors: make([]string, 0),
	}

	// Analyze build impact
	switch cycle.Severity {
	case CriticalSeverity:
		impact.BuildImpact = "May cause build failures or circular dependency errors"
		impact.RiskFactors = append(impact.RiskFactors, "build_failure_risk")
	case HighSeverity:
		impact.BuildImpact = "May cause warnings or slow builds"
		impact.RiskFactors = append(impact.RiskFactors, "build_performance_risk")
	default:
		impact.BuildImpact = "Minimal build impact"
	}

	// Analyze runtime impact
	switch cycle.Type {
	case ComponentCycle:
		impact.RuntimeImpact = "May cause rendering issues or infinite re-renders"
		impact.RiskFactors = append(impact.RiskFactors, "rendering_risk")
	case ServiceCycle:
		impact.RuntimeImpact = "May cause service initialization problems"
		impact.RiskFactors = append(impact.RiskFactors, "service_risk")
	default:
		impact.RuntimeImpact = "May cause module loading issues"
	}

	// Analyze maintainability impact
	impact.MaintainabilityImpact = fmt.Sprintf("Reduces code maintainability due to tight coupling (%d files involved)", cycle.Length)
	impact.RiskFactors = append(impact.RiskFactors, "maintainability_risk")

	// Analyze testability impact
	impact.TestabilityImpact = "Makes unit testing difficult due to interdependencies"
	impact.RiskFactors = append(impact.RiskFactors, "testability_risk")

	return impact
}

// generateResolutionStrategies generates strategies for resolving dependency cycles
func (cd *CycleDetector) generateResolutionStrategies(cycle *DependencyCycle) []ResolutionStrategy {
	strategies := make([]ResolutionStrategy, 0)

	// Strategy 1: Extract shared dependencies
	strategies = append(strategies, ResolutionStrategy{
		Strategy:    "Extract Common Dependencies",
		Description: "Move shared code to a separate module that both files can depend on",
		Priority:    1,
		Effort:      "medium",
	})

	// Strategy 2: Use dependency injection
	if cycle.Type == ServiceCycle || cycle.Type == ComponentCycle {
		strategies = append(strategies, ResolutionStrategy{
			Strategy:    "Implement Dependency Injection",
			Description: "Use dependency injection to break the circular dependency",
			Priority:    2,
			Effort:      "high",
		})
	}

	// Strategy 3: Merge modules
	if cycle.Length <= 2 {
		strategies = append(strategies, ResolutionStrategy{
			Strategy:    "Merge Modules",
			Description: "Consider merging the modules if they are tightly coupled",
			Priority:    3,
			Effort:      "low",
		})
	}

	// Strategy 4: Use interfaces/abstractions
	if cycle.Type == TypeCycle {
		strategies = append(strategies, ResolutionStrategy{
			Strategy:    "Use Type Abstractions",
			Description: "Create type interfaces to break type dependencies",
			Priority:    1,
			Effort:      "low",
		})
	}

	return strategies
}

// normalizeCyclePath normalizes a cycle path for comparison
func (cd *CycleDetector) normalizeCyclePath(cyclePath []string) []string {
	if len(cyclePath) <= 1 {
		return cyclePath
	}

	// Find the lexicographically smallest element as starting point
	minIndex := 0
	for i, file := range cyclePath[:len(cyclePath)-1] { // Exclude the duplicate at the end
		if file < cyclePath[minIndex] {
			minIndex = i
		}
	}

	// Rotate the cycle to start with the smallest element
	normalized := make([]string, len(cyclePath))
	for i := 0; i < len(cyclePath)-1; i++ {
		normalized[i] = cyclePath[(minIndex+i)%(len(cyclePath)-1)]
	}
	normalized[len(normalized)-1] = normalized[0] // Close the cycle

	return normalized
}

// compareCyclePaths compares two cycle paths for equality
func (cd *CycleDetector) compareCyclePaths(path1, path2 []string) bool {
	if len(path1) != len(path2) {
		return false
	}

	for i := range path1 {
		if path1[i] != path2[i] {
			return false
		}
	}

	return true
}

// generateCycleID generates a unique ID for a dependency cycle
func (cd *CycleDetector) generateCycleID(cyclePath []string) string {
	if len(cyclePath) == 0 {
		return "cycle-empty"
	}

	// Create a stable ID based on the normalized path
	pathStr := strings.Join(cyclePath[:len(cyclePath)-1], "-")
	return fmt.Sprintf("cycle-%s-%d", strings.ReplaceAll(pathStr, "/", "_"), len(cyclePath)-1)
}

// generateCycleDescription generates a human-readable description
func (cd *CycleDetector) generateCycleDescription(cyclePath []string) string {
	if len(cyclePath) <= 1 {
		return "Empty cycle"
	}

	pathStr := strings.Join(cyclePath, " → ")
	return fmt.Sprintf("Circular dependency: %s", pathStr)
}

// GetCycles returns all detected dependency cycles
func (cd *CycleDetector) GetCycles() []DependencyCycle {
	return cd.cycles
}

// GetCyclesBySeverity returns cycles filtered by severity
func (cd *CycleDetector) GetCyclesBySeverity(severity CycleSeverity) []DependencyCycle {
	filtered := make([]DependencyCycle, 0)
	for _, cycle := range cd.cycles {
		if cycle.Severity == severity {
			filtered = append(filtered, cycle)
		}
	}
	return filtered
}

// GetCyclesByType returns cycles filtered by type
func (cd *CycleDetector) GetCyclesByType(cycleType CycleType) []DependencyCycle {
	filtered := make([]DependencyCycle, 0)
	for _, cycle := range cd.cycles {
		if cycle.Type == cycleType {
			filtered = append(filtered, cycle)
		}
	}
	return filtered
}

// GetCycleStats returns comprehensive statistics about detected cycles
func (cd *CycleDetector) GetCycleStats() CycleStats {
	stats := CycleStats{
		TotalCycles:      len(cd.cycles),
		CyclesBySeverity: make(map[CycleSeverity]int),
		CyclesByType:     make(map[CycleType]int),
	}

	// Count cycles by severity and type
	totalLength := 0
	fileOccurrences := make(map[string]int)

	for _, cycle := range cd.cycles {
		stats.CyclesBySeverity[cycle.Severity]++
		stats.CyclesByType[cycle.Type]++
		totalLength += cycle.Length

		// Count file occurrences
		for _, file := range cycle.Files {
			fileOccurrences[file]++
		}
	}

	// Calculate average cycle length
	if len(cd.cycles) > 0 {
		stats.AverageCycleLength = float64(totalLength) / float64(len(cd.cycles))
	}

	// Find most problematic files
	type fileCount struct {
		file  string
		count int
	}

	fileCounts := make([]fileCount, 0, len(fileOccurrences))
	for file, count := range fileOccurrences {
		fileCounts = append(fileCounts, fileCount{file, count})
	}

	sort.Slice(fileCounts, func(i, j int) bool {
		return fileCounts[i].count > fileCounts[j].count
	})

	// Get top 5 most problematic files
	maxFiles := 5
	if len(fileCounts) < maxFiles {
		maxFiles = len(fileCounts)
	}

	stats.MostProblematicFiles = make([]string, maxFiles)
	for i := 0; i < maxFiles; i++ {
		stats.MostProblematicFiles[i] = fileCounts[i].file
	}

	// Calculate cycle complexity score (0-100)
	complexityScore := 0.0
	for _, cycle := range cd.cycles {
		switch cycle.Severity {
		case CriticalSeverity:
			complexityScore += 25.0
		case HighSeverity:
			complexityScore += 15.0
		case MediumSeverity:
			complexityScore += 8.0
		case LowSeverity:
			complexityScore += 3.0
		}
	}

	if complexityScore > 100 {
		complexityScore = 100
	}
	stats.CycleComplexityScore = complexityScore

	return stats
}

// ExportToJSON exports cycle detection results to JSON
func (cd *CycleDetector) ExportToJSON() (string, error) {
	data := map[string]interface{}{
		"cycles": cd.cycles,
		"stats":  cd.GetCycleStats(),
		"graph":  cd.dependencyGraph,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %w", err)
	}

	return string(jsonData), nil
}
