package ast

import (
	"fmt"
	"path/filepath"
	"strings"
)

// DependencyTracker tracks module relationships and dependencies
type DependencyTracker struct {
	fileResults  map[string]*ParseResult    // file path -> parse result
	dependencies map[string][]Dependency    // file path -> dependencies
	reverseDeps  map[string][]string        // module -> files that depend on it
	externalDeps map[string]ExternalPackage // package name -> package info
	moduleGraph  *ModuleGraph               // complete dependency graph
}

// Dependency represents a dependency relationship
type Dependency struct {
	SourceFile    string            `json:"source_file"`
	TargetModule  string            `json:"target_module"`
	ImportType    string            `json:"import_type"` // named, default, namespace, side-effect
	ImportedNames []string          `json:"imported_names"`
	LocalName     string            `json:"local_name"`
	IsExternal    bool              `json:"is_external"`
	IsResolved    bool              `json:"is_resolved"`
	ResolvedPath  string            `json:"resolved_path"`
	Line          int               `json:"line"`
	Metadata      map[string]string `json:"metadata"`
}

// ExternalPackage represents an external package dependency
type ExternalPackage struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	UsedBy           []string          `json:"used_by"`           // files using this package
	ImportedFeatures []string          `json:"imported_features"` // specific imports
	PackageType      string            `json:"package_type"`      // npm, built-in, scoped
	Metadata         map[string]string `json:"metadata"`
}

// ModuleGraph represents the complete dependency graph
type ModuleGraph struct {
	Nodes []ModuleNode `json:"nodes"`
	Edges []ModuleEdge `json:"edges"`
	Stats GraphStats   `json:"stats"`
}

// ModuleNode represents a node in the dependency graph
type ModuleNode struct {
	ID            string            `json:"id"`
	FilePath      string            `json:"file_path"`
	ModuleName    string            `json:"module_name"`
	NodeType      string            `json:"node_type"` // internal, external, entry-point
	ExportedItems []string          `json:"exported_items"`
	ImportCount   int               `json:"import_count"`
	ExportCount   int               `json:"export_count"`
	Metadata      map[string]string `json:"metadata"`
}

// ModuleEdge represents an edge in the dependency graph
type ModuleEdge struct {
	FromID         string            `json:"from_id"`
	ToID           string            `json:"to_id"`
	DependencyType string            `json:"dependency_type"` // import, export, re-export
	ImportType     string            `json:"import_type"`     // named, default, namespace
	ImportedNames  []string          `json:"imported_names"`
	Weight         int               `json:"weight"` // strength of dependency
	Metadata       map[string]string `json:"metadata"`
}

// GraphStats contains statistics about the dependency graph
type GraphStats struct {
	TotalNodes      int      `json:"total_nodes"`
	TotalEdges      int      `json:"total_edges"`
	ExternalNodes   int      `json:"external_nodes"`
	InternalNodes   int      `json:"internal_nodes"`
	CircularDeps    int      `json:"circular_deps"`
	MaxDepth        int      `json:"max_depth"`
	AvgDependencies float64  `json:"avg_dependencies"`
	TopDependencies []string `json:"top_dependencies"`
}

// NewDependencyTracker creates a new dependency tracker
func NewDependencyTracker() *DependencyTracker {
	return &DependencyTracker{
		fileResults:  make(map[string]*ParseResult),
		dependencies: make(map[string][]Dependency),
		reverseDeps:  make(map[string][]string),
		externalDeps: make(map[string]ExternalPackage),
		moduleGraph: &ModuleGraph{
			Nodes: []ModuleNode{},
			Edges: []ModuleEdge{},
			Stats: GraphStats{},
		},
	}
}

// AddParseResult adds a parse result for dependency analysis
func (dt *DependencyTracker) AddParseResult(filePath string, result *ParseResult) error {
	// Normalize file path
	normalizedPath := filepath.Clean(filePath)
	dt.fileResults[normalizedPath] = result

	// Extract dependencies from imports
	dependencies := make([]Dependency, 0, len(result.Imports))
	for _, imp := range result.Imports {
		dep := Dependency{
			SourceFile:    normalizedPath,
			TargetModule:  imp.Source,
			ImportType:    imp.ImportType,
			ImportedNames: imp.Specifiers,
			LocalName:     imp.LocalName,
			IsExternal:    imp.IsExternal,
			IsResolved:    false,
			Line:          imp.StartLine,
			Metadata:      make(map[string]string),
		}

		// Add metadata
		dep.Metadata["language"] = result.Language
		dep.Metadata["import_line"] = fmt.Sprintf("%d", imp.StartLine)

		dependencies = append(dependencies, dep)

		// Track reverse dependencies
		targetKey := dt.getModuleKey(imp.Source, imp.IsExternal)
		if dt.reverseDeps[targetKey] == nil {
			dt.reverseDeps[targetKey] = []string{}
		}
		dt.reverseDeps[targetKey] = append(dt.reverseDeps[targetKey], normalizedPath)

		// Track external packages
		if imp.IsExternal {
			dt.trackExternalPackage(imp, normalizedPath)
		}
	}

	dt.dependencies[normalizedPath] = dependencies
	return nil
}

// ResolveDependencies attempts to resolve internal file dependencies
func (dt *DependencyTracker) ResolveDependencies(projectRoot string) error {
	for filePath, deps := range dt.dependencies {
		for i, dep := range deps {
			if !dep.IsExternal && !dep.IsResolved {
				resolved, resolvedPath := dt.resolveInternalDependency(dep.TargetModule, filePath, projectRoot)
				if resolved {
					deps[i].IsResolved = true
					deps[i].ResolvedPath = resolvedPath
					deps[i].Metadata["resolved"] = "true"
				}
			}
		}
		dt.dependencies[filePath] = deps
	}
	return nil
}

// BuildModuleGraph constructs the complete module dependency graph
func (dt *DependencyTracker) BuildModuleGraph() (*ModuleGraph, error) {
	// Reset graph
	dt.moduleGraph.Nodes = []ModuleNode{}
	dt.moduleGraph.Edges = []ModuleEdge{}

	nodeMap := make(map[string]int) // module key -> node index

	// Create nodes for all files
	for filePath, result := range dt.fileResults {
		nodeID := dt.getNodeID(filePath, false)
		node := ModuleNode{
			ID:            nodeID,
			FilePath:      filePath,
			ModuleName:    dt.getModuleName(filePath),
			NodeType:      "internal",
			ExportedItems: dt.getExportedItems(result),
			ImportCount:   len(result.Imports),
			ExportCount:   len(result.Exports),
			Metadata:      make(map[string]string),
		}

		// Add metadata
		node.Metadata["language"] = result.Language
		node.Metadata["functions"] = fmt.Sprintf("%d", len(result.Functions))
		node.Metadata["classes"] = fmt.Sprintf("%d", len(result.Classes))
		node.Metadata["interfaces"] = fmt.Sprintf("%d", len(result.Interfaces))

		nodeMap[nodeID] = len(dt.moduleGraph.Nodes)
		dt.moduleGraph.Nodes = append(dt.moduleGraph.Nodes, node)
	}

	// Create nodes for external packages
	for packageName, packageInfo := range dt.externalDeps {
		nodeID := dt.getNodeID(packageName, true)
		node := ModuleNode{
			ID:            nodeID,
			FilePath:      "",
			ModuleName:    packageName,
			NodeType:      "external",
			ExportedItems: packageInfo.ImportedFeatures,
			ImportCount:   0,
			ExportCount:   len(packageInfo.ImportedFeatures),
			Metadata:      packageInfo.Metadata,
		}

		nodeMap[nodeID] = len(dt.moduleGraph.Nodes)
		dt.moduleGraph.Nodes = append(dt.moduleGraph.Nodes, node)
	}

	// Create edges from dependencies
	for filePath, deps := range dt.dependencies {
		fromID := dt.getNodeID(filePath, false)
		fromIndex, exists := nodeMap[fromID]
		if !exists {
			continue
		}

		for _, dep := range deps {
			toID := dt.getNodeID(dep.TargetModule, dep.IsExternal)
			toIndex, exists := nodeMap[toID]
			if !exists {
				continue
			}

			edge := ModuleEdge{
				FromID:         fromID,
				ToID:           toID,
				DependencyType: "import",
				ImportType:     dep.ImportType,
				ImportedNames:  dep.ImportedNames,
				Weight:         dt.calculateEdgeWeight(dep),
				Metadata:       make(map[string]string),
			}

			// Add edge metadata
			edge.Metadata["source_line"] = fmt.Sprintf("%d", dep.Line)
			if dep.IsResolved {
				edge.Metadata["resolved_path"] = dep.ResolvedPath
			}

			dt.moduleGraph.Edges = append(dt.moduleGraph.Edges, edge)

			// Update node import/export counts
			dt.moduleGraph.Nodes[fromIndex].ImportCount++
			dt.moduleGraph.Nodes[toIndex].ExportCount++
		}
	}

	// Calculate graph statistics
	dt.calculateGraphStats()

	return dt.moduleGraph, nil
}

// GetDependencies returns dependencies for a specific file
func (dt *DependencyTracker) GetDependencies(filePath string) ([]Dependency, bool) {
	deps, exists := dt.dependencies[filepath.Clean(filePath)]
	return deps, exists
}

// GetDependents returns files that depend on a specific module
func (dt *DependencyTracker) GetDependents(modulePath string, isExternal bool) []string {
	key := dt.getModuleKey(modulePath, isExternal)
	if dependents, exists := dt.reverseDeps[key]; exists {
		return dependents
	}
	return []string{}
}

// GetExternalPackages returns all external package dependencies
func (dt *DependencyTracker) GetExternalPackages() map[string]ExternalPackage {
	return dt.externalDeps
}

// Helper methods

func (dt *DependencyTracker) trackExternalPackage(imp ImportInfo, filePath string) {
	packageName := dt.getPackageName(imp.Source)

	if pkg, exists := dt.externalDeps[packageName]; exists {
		// Update existing package
		pkg.UsedBy = append(pkg.UsedBy, filePath)
		for _, spec := range imp.Specifiers {
			if !contains(pkg.ImportedFeatures, spec) {
				pkg.ImportedFeatures = append(pkg.ImportedFeatures, spec)
			}
		}
		dt.externalDeps[packageName] = pkg
	} else {
		// Create new package entry
		pkg := ExternalPackage{
			Name:             packageName,
			UsedBy:           []string{filePath},
			ImportedFeatures: imp.Specifiers,
			PackageType:      dt.getPackageType(packageName),
			Metadata:         make(map[string]string),
		}

		pkg.Metadata["import_type"] = imp.ImportType
		dt.externalDeps[packageName] = pkg
	}
}

func (dt *DependencyTracker) resolveInternalDependency(targetModule, sourceFile, projectRoot string) (bool, string) {
	basePath := filepath.Dir(sourceFile)

	// Try different resolution strategies
	candidatePaths := []string{
		filepath.Join(basePath, targetModule),
		filepath.Join(basePath, targetModule+".js"),
		filepath.Join(basePath, targetModule+".ts"),
		filepath.Join(basePath, targetModule+".jsx"),
		filepath.Join(basePath, targetModule+".tsx"),
		filepath.Join(basePath, targetModule, "index.js"),
		filepath.Join(basePath, targetModule, "index.ts"),
	}

	for _, candidatePath := range candidatePaths {
		normalizedPath := filepath.Clean(candidatePath)
		if _, exists := dt.fileResults[normalizedPath]; exists {
			return true, normalizedPath
		}
	}

	return false, ""
}

func (dt *DependencyTracker) getExportedItems(result *ParseResult) []string {
	items := []string{}

	for _, export := range result.Exports {
		if export.ExportType == "default" && export.Name != "" {
			items = append(items, "default:"+export.Name)
		} else {
			items = append(items, export.Specifiers...)
		}
	}

	return items
}

func (dt *DependencyTracker) calculateEdgeWeight(dep Dependency) int {
	weight := 1

	// Increase weight based on import type
	switch dep.ImportType {
	case "default":
		weight += 1
	case "namespace":
		weight += 2
	case "named":
		weight += len(dep.ImportedNames)
	}

	return weight
}

func (dt *DependencyTracker) calculateGraphStats() {
	stats := &dt.moduleGraph.Stats
	stats.TotalNodes = len(dt.moduleGraph.Nodes)
	stats.TotalEdges = len(dt.moduleGraph.Edges)

	// Count node types
	for _, node := range dt.moduleGraph.Nodes {
		if node.NodeType == "external" {
			stats.ExternalNodes++
		} else {
			stats.InternalNodes++
		}
	}

	// Calculate average dependencies
	if stats.TotalNodes > 0 {
		stats.AvgDependencies = float64(stats.TotalEdges) / float64(stats.TotalNodes)
	}

	// Find top dependencies (most imported modules)
	dependencyCounts := make(map[string]int)
	for _, edge := range dt.moduleGraph.Edges {
		dependencyCounts[edge.ToID]++
	}

	// Sort and get top 5
	// Implementation would sort dependencyCounts and take top entries
	stats.TopDependencies = []string{} // Simplified for now
}

func (dt *DependencyTracker) getModuleKey(modulePath string, isExternal bool) string {
	if isExternal {
		return "ext:" + dt.getPackageName(modulePath)
	}
	return "int:" + modulePath
}

func (dt *DependencyTracker) getNodeID(path string, isExternal bool) string {
	if isExternal {
		return "ext_" + strings.ReplaceAll(dt.getPackageName(path), "/", "_")
	}
	return "int_" + strings.ReplaceAll(path, "/", "_")
}

func (dt *DependencyTracker) getModuleName(filePath string) string {
	return filepath.Base(strings.TrimSuffix(filePath, filepath.Ext(filePath)))
}

func (dt *DependencyTracker) getPackageName(source string) string {
	// Handle scoped packages like @types/node
	if strings.HasPrefix(source, "@") {
		parts := strings.Split(source, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	}

	// Handle regular packages
	parts := strings.Split(source, "/")
	return parts[0]
}

func (dt *DependencyTracker) getPackageType(packageName string) string {
	if strings.HasPrefix(packageName, "@") {
		return "scoped"
	}

	builtIns := []string{"fs", "path", "http", "https", "util", "events", "stream", "crypto"}
	for _, builtIn := range builtIns {
		if packageName == builtIn {
			return "built-in"
		}
	}

	return "npm"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
