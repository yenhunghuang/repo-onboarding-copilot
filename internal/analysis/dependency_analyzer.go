// Package analysis provides dependency analysis for project dependencies
// including package.json parsing, vulnerability assessment, and license analysis
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"
)

// DependencyAnalyzer analyzes project dependencies from package management files
type DependencyAnalyzer struct {
	projectRoot       string
	config            DependencyAnalyzerConfig
	vulnerabilityDB   *VulnerabilityDatabase
	licenseChecker    *LicenseChecker
	updateChecker     *UpdateChecker
	performanceAnalyzer *PerformanceAnalyzer
	bundleAnalyzer    *BundleAnalyzer
}

// DependencyAnalyzerConfig configures dependency analysis behavior
type DependencyAnalyzerConfig struct {
	ProjectRoot            string   `json:"project_root"`
	IncludePackageFiles    []string `json:"include_package_files"`    // package.json, package-lock.json, yarn.lock
	EnableVulnScanning     bool     `json:"enable_vuln_scanning"`
	EnableLicenseChecking  bool     `json:"enable_license_checking"`
	EnableUpdateChecking   bool     `json:"enable_update_checking"`
	EnablePerformanceAnalysis bool  `json:"enable_performance_analysis"`
	EnableBundleAnalysis   bool     `json:"enable_bundle_analysis"`
	MaxDependencyDepth     int      `json:"max_dependency_depth"`     // limit transitive dependency resolution
	BundleSizeThreshold    int64    `json:"bundle_size_threshold"`    // bytes
	PerformanceThreshold   int      `json:"performance_threshold"`   // ms
	CriticalVulnThreshold  float64  `json:"critical_vuln_threshold"` // CVSS score
}

// PackageManifest represents parsed package.json data
type PackageManifest struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	Main            string                 `json:"main"`
	Scripts         map[string]string      `json:"scripts"`
	Dependencies    map[string]string      `json:"dependencies"`
	DevDependencies map[string]string      `json:"devDependencies"`
	PeerDependencies map[string]string     `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	BundledDependencies []string           `json:"bundledDependencies"`
	Engines         map[string]string      `json:"engines"`
	Repository      interface{}            `json:"repository"`
	Author          interface{}            `json:"author"`
	License         interface{}            `json:"license"`
	Keywords        []string               `json:"keywords"`
	Homepage        string                 `json:"homepage"`
	Bugs            interface{}            `json:"bugs"`
	Private         bool                   `json:"private"`
	Workspaces      interface{}            `json:"workspaces"`
	Metadata        map[string]interface{} `json:"-"` // store additional fields
}

// LockFile represents parsed lock file data (npm/yarn)
type LockFile struct {
	Type         string                 `json:"type"`         // npm-lock, yarn-lock
	Version      string                 `json:"version"`      // lock file format version
	Dependencies map[string]LockEntry   `json:"dependencies"` // resolved dependencies
	Metadata     map[string]interface{} `json:"metadata"`     // additional lock file data
}

// LockEntry represents a resolved dependency in lock file
type LockEntry struct {
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved"`     // download URL
	Integrity    string            `json:"integrity"`    // checksum
	Dependencies map[string]string `json:"dependencies"` // transitive dependencies
	DevDep       bool              `json:"dev"`          // development dependency
	Optional     bool              `json:"optional"`     // optional dependency
	Bundled      bool              `json:"bundled"`      // bundled dependency
}

// DependencyNode represents a node in the dependency tree
type DependencyNode struct {
	Name             string                     `json:"name"`
	Version          string                     `json:"version"`
	RequestedVersion string                     `json:"requested_version"` // semver range from package.json
	ResolvedVersion  string                     `json:"resolved_version"`  // exact version from lock file
	Type             string                     `json:"type"`              // production, development, peer, optional
	Path             string                     `json:"path"`              // node_modules path
	Children         map[string]*DependencyNode `json:"children"`          // transitive dependencies
	Parent           *DependencyNode            `json:"-"`                 // parent reference (avoid circular JSON)
	Depth            int                        `json:"depth"`             // depth in dependency tree
	IsTransitive     bool                       `json:"is_transitive"`     // false for direct dependencies
	PackageInfo      *PackageInfo               `json:"package_info"`      // metadata about the package
	Vulnerabilities  []Vulnerability            `json:"vulnerabilities"`   // security vulnerabilities
	License          LicenseInfo                `json:"license"`           // license information
	UpdateInfo       *UpdateInfo                `json:"update_info"`       // available updates
}

// PackageInfo contains metadata about a package
type PackageInfo struct {
	Description    string            `json:"description"`
	Homepage       string            `json:"homepage"`
	Repository     string            `json:"repository"`
	Author         string            `json:"author"`
	Keywords       []string          `json:"keywords"`
	EstimatedSize  int64             `json:"estimated_size"`  // bytes
	DownloadCount  int               `json:"download_count"`  // weekly downloads
	LastModified   time.Time         `json:"last_modified"`
	Deprecated     bool              `json:"deprecated"`
	MaintenanceScore float64         `json:"maintenance_score"` // 0-1 scale
	Metadata       map[string]string `json:"metadata"`
}

// DependencyTree represents the complete dependency hierarchy
type DependencyTree struct {
	RootPackage       *PackageManifest            `json:"root_package"`
	DirectDeps        map[string]*DependencyNode  `json:"direct_dependencies"`
	AllDependencies   map[string]*DependencyNode  `json:"all_dependencies"`      // flattened view
	LockData          *LockFile                   `json:"lock_data"`
	Statistics        DependencyStats             `json:"statistics"`
	Graph             *DependencyGraph            `json:"graph"`                 // graph representation
	BundleAnalysis    *BundleAnalysis             `json:"bundle_analysis"`
	SecurityReport    *SecurityReport             `json:"security_report"`
	LicenseReport     *LicenseReport              `json:"license_report"`
	UpdateReport      *UpdateReport               `json:"update_report"`
	PerformanceReport *PerformanceReport          `json:"performance_report"`
	BundleResult      *BundleAnalysisResult       `json:"bundle_result"`
}

// DependencyStats contains dependency tree statistics
type DependencyStats struct {
	TotalDependencies   int               `json:"total_dependencies"`
	DirectDependencies  int               `json:"direct_dependencies"`
	DevDependencies     int               `json:"dev_dependencies"`
	TransitiveDependencies int            `json:"transitive_dependencies"`
	MaxDepth            int               `json:"max_depth"`
	DuplicatePackages   int               `json:"duplicate_packages"`
	OutdatedPackages    int               `json:"outdated_packages"`
	VulnerablePackages  int               `json:"vulnerable_packages"`
	DeprecatedPackages  int               `json:"deprecated_packages"`
	TotalSize           int64             `json:"total_size"`         // estimated bytes
	TypeDistribution    map[string]int    `json:"type_distribution"`  // production, development, etc.
	LicenseDistribution map[string]int    `json:"license_distribution"`
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(config DependencyAnalyzerConfig) (*DependencyAnalyzer, error) {
	if config.ProjectRoot == "" {
		return nil, fmt.Errorf("project root is required")
	}

	// Set defaults
	if len(config.IncludePackageFiles) == 0 {
		config.IncludePackageFiles = []string{"package.json", "package-lock.json", "yarn.lock"}
	}
	if config.MaxDependencyDepth == 0 {
		config.MaxDependencyDepth = 10 // reasonable default to prevent infinite recursion
	}
	if config.BundleSizeThreshold == 0 {
		config.BundleSizeThreshold = 500 * 1024 // 500KB default
	}
	if config.PerformanceThreshold == 0 {
		config.PerformanceThreshold = 3000 // 3 seconds default
	}
	if config.CriticalVulnThreshold == 0.0 {
		config.CriticalVulnThreshold = 7.0 // CVSS 7.0+ considered critical
	}

	analyzer := &DependencyAnalyzer{
		projectRoot: config.ProjectRoot,
		config:      config,
	}

	// Initialize sub-components if enabled
	if config.EnableVulnScanning {
		vulnDB, err := NewVulnerabilityDatabase()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize vulnerability database: %w", err)
		}
		analyzer.vulnerabilityDB = vulnDB
	}

	if config.EnableLicenseChecking {
		licenseChecker, err := NewLicenseChecker()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize license checker: %w", err)
		}
		analyzer.licenseChecker = licenseChecker
	}

	if config.EnableUpdateChecking {
		updateChecker, err := NewUpdateChecker()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize update checker: %w", err)
		}
		analyzer.updateChecker = updateChecker
	}

	if config.EnablePerformanceAnalysis {
		analyzer.performanceAnalyzer = NewPerformanceAnalyzer()
	}

	if config.EnableBundleAnalysis {
		analyzer.bundleAnalyzer = NewBundleAnalyzer()
	}

	return analyzer, nil
}

// AnalyzeDependencies performs complete dependency analysis
func (da *DependencyAnalyzer) AnalyzeDependencies(ctx context.Context) (*DependencyTree, error) {
	// Step 1: Find and parse package management files
	packageFiles, err := da.findPackageFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to find package files: %w", err)
	}

	if len(packageFiles) == 0 {
		return nil, fmt.Errorf("no package management files found in project root: %s", da.projectRoot)
	}

	// Step 2: Parse package.json
	var manifest *PackageManifest
	var lockFile *LockFile

	for _, file := range packageFiles {
		switch filepath.Base(file) {
		case "package.json":
			manifest, err = da.parsePackageJSON(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse package.json: %w", err)
			}
		case "package-lock.json":
			lockFile, err = da.parseNPMLockFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse package-lock.json: %w", err)
			}
		case "yarn.lock":
			if lockFile == nil { // prefer npm lock if both exist
				lockFile, err = da.parseYarnLockFile(file)
				if err != nil {
					return nil, fmt.Errorf("failed to parse yarn.lock: %w", err)
				}
			}
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("package.json is required for dependency analysis")
	}

	// Step 3: Build dependency tree
	tree, err := da.buildDependencyTree(ctx, manifest, lockFile)
	if err != nil {
		return nil, fmt.Errorf("failed to build dependency tree: %w", err)
	}

	// Step 4: Additional analysis based on configuration
	if err := da.enrichDependencyTree(ctx, tree); err != nil {
		return nil, fmt.Errorf("failed to enrich dependency tree: %w", err)
	}

	return tree, nil
}

// findPackageFiles locates package management files in the project root
func (da *DependencyAnalyzer) findPackageFiles() ([]string, error) {
	var files []string

	for _, filename := range da.config.IncludePackageFiles {
		filePath := filepath.Join(da.projectRoot, filename)
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			files = append(files, filePath)
		}
	}

	return files, nil
}

// parsePackageJSON parses a package.json file
func (da *DependencyAnalyzer) parsePackageJSON(filePath string) (*PackageManifest, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(content, &rawData); err != nil {
		return nil, fmt.Errorf("invalid JSON in package.json: %w", err)
	}

	// Parse into structured format
	manifest := &PackageManifest{
		Metadata: make(map[string]interface{}),
	}

	// Use type assertions with safe handling
	if name, ok := rawData["name"].(string); ok {
		manifest.Name = name
	}
	if version, ok := rawData["version"].(string); ok {
		manifest.Version = version
	}
	if description, ok := rawData["description"].(string); ok {
		manifest.Description = description
	}
	if main, ok := rawData["main"].(string); ok {
		manifest.Main = main
	}
	if private, ok := rawData["private"].(bool); ok {
		manifest.Private = private
	}
	if homepage, ok := rawData["homepage"].(string); ok {
		manifest.Homepage = homepage
	}

	// Parse dependencies
	manifest.Dependencies = da.parseStringMap(rawData["dependencies"])
	manifest.DevDependencies = da.parseStringMap(rawData["devDependencies"])
	manifest.PeerDependencies = da.parseStringMap(rawData["peerDependencies"])
	manifest.OptionalDependencies = da.parseStringMap(rawData["optionalDependencies"])

	// Parse scripts
	manifest.Scripts = da.parseStringMap(rawData["scripts"])

	// Parse engines
	manifest.Engines = da.parseStringMap(rawData["engines"])

	// Parse keywords
	if keywords, ok := rawData["keywords"].([]interface{}); ok {
		for _, keyword := range keywords {
			if str, ok := keyword.(string); ok {
				manifest.Keywords = append(manifest.Keywords, str)
			}
		}
	}

	// Parse bundledDependencies
	if bundled, ok := rawData["bundledDependencies"].([]interface{}); ok {
		for _, dep := range bundled {
			if str, ok := dep.(string); ok {
				manifest.BundledDependencies = append(manifest.BundledDependencies, str)
			}
		}
	}

	// Store complex fields as-is for later processing
	manifest.Repository = rawData["repository"]
	manifest.Author = rawData["author"]
	manifest.License = rawData["license"]
	manifest.Bugs = rawData["bugs"]
	manifest.Workspaces = rawData["workspaces"]

	// Store all remaining fields in metadata
	knownFields := map[string]bool{
		"name": true, "version": true, "description": true, "main": true, "private": true,
		"homepage": true, "dependencies": true, "devDependencies": true, "peerDependencies": true,
		"optionalDependencies": true, "bundledDependencies": true, "scripts": true, "engines": true,
		"keywords": true, "repository": true, "author": true, "license": true, "bugs": true, "workspaces": true,
	}

	for key, value := range rawData {
		if !knownFields[key] {
			manifest.Metadata[key] = value
		}
	}

	return manifest, nil
}

// parseStringMap safely parses map[string]interface{} to map[string]string
func (da *DependencyAnalyzer) parseStringMap(data interface{}) map[string]string {
	result := make(map[string]string)
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			if str, ok := value.(string); ok {
				result[key] = str
			}
		}
	}
	return result
}

// Close releases analyzer resources
func (da *DependencyAnalyzer) Close() error {
	var errors []error

	if da.vulnerabilityDB != nil {
		if err := da.vulnerabilityDB.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close vulnerability database: %w", err))
		}
	}

	if da.licenseChecker != nil {
		if err := da.licenseChecker.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close license checker: %w", err))
		}
	}

	if da.updateChecker != nil {
		if err := da.updateChecker.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close update checker: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("multiple close errors: %v", errors)
	}

	return nil
}

// buildDependencyTree constructs the dependency tree from manifest and lock data
func (da *DependencyAnalyzer) buildDependencyTree(ctx context.Context, manifest *PackageManifest, lockFile *LockFile) (*DependencyTree, error) {
	tree := &DependencyTree{
		RootPackage:     manifest,
		DirectDeps:      make(map[string]*DependencyNode),
		AllDependencies: make(map[string]*DependencyNode),
		LockData:        lockFile,
		Statistics:      DependencyStats{TypeDistribution: make(map[string]int), LicenseDistribution: make(map[string]int)},
	}

	// Build direct dependencies from package.json
	for name, version := range manifest.Dependencies {
		node := da.createDependencyNode(name, version, "production", lockFile, 0, false)
		tree.DirectDeps[name] = node
		tree.AllDependencies[name] = node
		tree.Statistics.DirectDependencies++
	}

	for name, version := range manifest.DevDependencies {
		node := da.createDependencyNode(name, version, "development", lockFile, 0, false)
		tree.DirectDeps[name] = node
		tree.AllDependencies[name] = node
		tree.Statistics.DirectDependencies++
		tree.Statistics.DevDependencies++
	}

	for name, version := range manifest.PeerDependencies {
		node := da.createDependencyNode(name, version, "peer", lockFile, 0, false)
		tree.DirectDeps[name] = node
		tree.AllDependencies[name] = node
		tree.Statistics.DirectDependencies++
	}

	for name, version := range manifest.OptionalDependencies {
		node := da.createDependencyNode(name, version, "optional", lockFile, 0, false)
		tree.DirectDeps[name] = node
		tree.AllDependencies[name] = node
		tree.Statistics.DirectDependencies++
	}

	// Resolve transitive dependencies
	da.resolveTransitiveDependencies(tree, lockFile)

	// Calculate statistics
	da.calculateDependencyStats(tree)

	return tree, nil
}

// createDependencyNode creates a dependency node with information from lock file
func (da *DependencyAnalyzer) createDependencyNode(name, requestedVersion, depType string, lockFile *LockFile, depth int, isTransitive bool) *DependencyNode {
	node := &DependencyNode{
		Name:             name,
		RequestedVersion: requestedVersion,
		Type:             depType,
		Depth:            depth,
		IsTransitive:     isTransitive,
		Children:         make(map[string]*DependencyNode),
		PackageInfo:      &PackageInfo{},
		Vulnerabilities:  []Vulnerability{},
		License:          LicenseInfo{},
	}

	// Populate from lock file if available
	if lockFile != nil && lockFile.Dependencies != nil {
		if lockEntry, exists := lockFile.Dependencies[name]; exists {
			node.Version = lockEntry.Version
			node.ResolvedVersion = lockEntry.Version
			// Set path based on lock file type
			if lockFile.Type == "npm-lock" {
				node.Path = "node_modules/" + name
			} else if lockFile.Type == "yarn-lock" {
				node.Path = "node_modules/" + name
			}
		}
	}

	// If no lock file or entry not found, use requested version
	if node.Version == "" {
		node.Version = requestedVersion
		node.ResolvedVersion = requestedVersion
	}

	return node
}

// resolveTransitiveDependencies resolves transitive dependencies up to max depth
func (da *DependencyAnalyzer) resolveTransitiveDependencies(tree *DependencyTree, lockFile *LockFile) {
	if lockFile == nil || lockFile.Dependencies == nil {
		return
	}

	// Track visited nodes to prevent infinite recursion
	visited := make(map[string]bool)

	// Resolve transitive dependencies for each direct dependency
	for _, node := range tree.DirectDeps {
		da.resolveNodeDependencies(node, tree, lockFile, visited, 1)
	}
}

// resolveNodeDependencies recursively resolves dependencies for a node
func (da *DependencyAnalyzer) resolveNodeDependencies(node *DependencyNode, tree *DependencyTree, lockFile *LockFile, visited map[string]bool, depth int) {
	// Prevent infinite recursion and respect max depth
	nodeKey := node.Name + "@" + node.Version
	if visited[nodeKey] || depth > da.config.MaxDependencyDepth {
		return
	}
	visited[nodeKey] = true

	// Find dependencies for this node in lock file
	if lockEntry, exists := lockFile.Dependencies[node.Name]; exists {
		for depName, depVersion := range lockEntry.Dependencies {
			// Create child node
			childNode := da.createDependencyNode(depName, depVersion, "transitive", lockFile, depth, true)
			node.Children[depName] = childNode
			childNode.Parent = node

			// Add to tree's all dependencies if not already present
			if _, exists := tree.AllDependencies[depName]; !exists {
				tree.AllDependencies[depName] = childNode
				tree.Statistics.TransitiveDependencies++
			}

			// Recursively resolve child dependencies
			da.resolveNodeDependencies(childNode, tree, lockFile, visited, depth+1)
		}
	}
}

// calculateDependencyStats calculates statistics for the dependency tree
func (da *DependencyAnalyzer) calculateDependencyStats(tree *DependencyTree) {
	stats := &tree.Statistics

	stats.TotalDependencies = len(tree.AllDependencies)

	// Calculate type distribution
	stats.TypeDistribution = make(map[string]int)
	for _, node := range tree.AllDependencies {
		stats.TypeDistribution[node.Type]++
	}

	// Calculate max depth
	stats.MaxDepth = da.calculateMaxDepth(tree)

	// Initialize other statistics (will be populated by enrichDependencyTree)
	stats.LicenseDistribution = make(map[string]int)
}

// calculateMaxDepth finds the maximum depth in the dependency tree
func (da *DependencyAnalyzer) calculateMaxDepth(tree *DependencyTree) int {
	maxDepth := 0
	for _, node := range tree.DirectDeps {
		depth := da.getNodeMaxDepth(node, 0)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

// getNodeMaxDepth recursively calculates the maximum depth for a node
func (da *DependencyAnalyzer) getNodeMaxDepth(node *DependencyNode, currentDepth int) int {
	maxDepth := currentDepth
	for _, child := range node.Children {
		depth := da.getNodeMaxDepth(child, currentDepth+1)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

// enrichDependencyTree adds additional analysis data to the dependency tree
func (da *DependencyAnalyzer) enrichDependencyTree(ctx context.Context, tree *DependencyTree) error {
	// Create basic reports (stubs will be populated by later tasks)
	tree.BundleAnalysis = &BundleAnalysis{
		SizeByType:       make(map[string]int64),
		LargestPackages:  []PackageSize{},
		Recommendations:  []string{},
		LoadTimeEstimate: make(map[string]float64),
	}

	tree.SecurityReport = &SecurityReport{
		VulnerablePackages:   []string{},
		Vulnerabilities:      []Vulnerability{},
		SeverityDistribution: make(map[string]int),
		Recommendations:      []string{},
	}

	tree.LicenseReport = &LicenseReport{
		LicenseDistribution:  make(map[string]int),
		CompatibilityIssues:  []LicenseConflict{},
		UnknownLicenses:      []string{},
		ProprietaryPackages:  []string{},
		CopyleftPackages:     []string{},
		Recommendations:      []string{},
	}

	tree.UpdateReport = &UpdateReport{
		Updates:           []UpdateInfo{},
		UpdatesByType:     make(map[string]int),
		UpdatesByPriority: make(map[string]int),
		Recommendations:   []string{},
	}

	// Initialize performance analysis results
	tree.PerformanceReport = &PerformanceReport{
		Packages:          []PerformanceImpact{},
		AverageLoadTime:   make(map[string]float64),
		TotalImpact:       0.0,
		Recommendations:   []PerformanceRecommendation{},
	}

	tree.BundleResult = &BundleAnalysisResult{
		TotalSize:            0,
		MinifiedSize:         0,
		CompressedSize:       0,
		TreeShakableSize:     0,
		OptimizedSize:        0,
		PackageContributions: []PackageContribution{},
		BudgetAnalysis:       &BudgetAnalysis{},
		LoadTimeAnalysis:     &LoadTimeAnalysis{},
		Recommendations:      []PerformanceRecommendation{},
		SizeBreakdown:        &SizeBreakdown{},
	}

	// Create basic dependency graph
	tree.Graph = da.createDependencyGraph(tree)

	// Extend analysis for monorepo structures
	if err := da.extendAnalysisForMonorepo(ctx, tree); err != nil {
		return fmt.Errorf("monorepo analysis failed: %w", err)
	}

	// Run enabled analysis components
	if da.config.EnableVulnScanning && da.vulnerabilityDB != nil {
		if err := da.runVulnerabilityAnalysis(ctx, tree); err != nil {
			return fmt.Errorf("vulnerability analysis failed: %w", err)
		}
	}

	if da.config.EnableLicenseChecking && da.licenseChecker != nil {
		if err := da.runLicenseAnalysis(ctx, tree); err != nil {
			return fmt.Errorf("license analysis failed: %w", err)
		}
	}

	if da.config.EnableUpdateChecking && da.updateChecker != nil {
		if err := da.runUpdateAnalysis(ctx, tree); err != nil {
			return fmt.Errorf("update analysis failed: %w", err)
		}
	}

	// Run performance analysis if enabled
	if da.config.EnablePerformanceAnalysis && da.performanceAnalyzer != nil {
		if err := da.runPerformanceAnalysis(ctx, tree); err != nil {
			return fmt.Errorf("performance analysis failed: %w", err)
		}
	}

	// Run bundle analysis if enabled
	if da.config.EnableBundleAnalysis && da.bundleAnalyzer != nil {
		if err := da.runBundleAnalysis(ctx, tree); err != nil {
			return fmt.Errorf("bundle analysis failed: %w", err)
		}
	} else {
		// Run basic bundle estimation as fallback
		da.runBasicBundleEstimation(tree)
	}

	return nil
}

// createDependencyGraph creates an advanced graph representation of the dependency tree
func (da *DependencyAnalyzer) createDependencyGraph(tree *DependencyTree) *DependencyGraph {
	// Create packages list from dependency tree
	packages := make([]*GraphPackageInfo, 0, len(tree.AllDependencies))
	
	for name, node := range tree.AllDependencies {
		// Convert DependencyNode to GraphPackageInfo format
		pkg := &GraphPackageInfo{
			Name:         name,
			Version:      node.Version,
			DependencyType: node.Type,
			Dependencies: make(map[string]string),
			DevDependencies: make(map[string]string),
			PeerDependencies: make(map[string]string),
			Description:  getNodeDescription(node),
			Homepage:     getNodeHomepage(node),
			Repository:   getNodeRepository(node),
			RegistryURL:  fmt.Sprintf("https://registry.npmjs.org/%s", name),
		}
		
		// Add dependencies based on type
		for childName, childNode := range node.Children {
			switch node.Type {
			case "devDependencies":
				pkg.DevDependencies[childName] = childNode.RequestedVersion
			case "peerDependencies":
				pkg.PeerDependencies[childName] = childNode.RequestedVersion
			default:
				pkg.Dependencies[childName] = childNode.RequestedVersion
			}
		}
		
		packages = append(packages, pkg)
	}
	
	// Use GraphBuilder to create advanced graph
	builder := NewGraphBuilder("npm") // Default to npm for now
	err := builder.BuildFromPackageList(packages)
	if err != nil {
		// Fallback to basic graph on error
		return da.createBasicGraph(tree)
	}
	
	graph := builder.GetGraph()
	
	// Enrich graph with analysis results from tree
	da.enrichGraphWithAnalysisData(graph, tree)
	
	return graph
}

// createBasicGraph creates a basic graph as fallback
func (da *DependencyAnalyzer) createBasicGraph(tree *DependencyTree) *DependencyGraph {
	builder := NewGraphBuilder("npm") // Default to npm
	graph := builder.GetGraph()
	
	// Add basic nodes
	for name, node := range tree.AllDependencies {
		nodeID := fmt.Sprintf("%s@%s", name, node.Version)
		graphNode := &GraphNode{
			ID:          nodeID,
			Name:        name,
			Version:     node.Version,
			PackageType: node.Type,
			Size:        0,
			Weight:      1.0,
			Depth:       node.Depth,
			Metadata:    make(map[string]string),
			VulnerabilityCount: len(node.Vulnerabilities),
			LicenseInfo: node.License.SPDX,
			UpdatesAvailable: 0,
			RiskScore: 0.0,
		}
		
		graph.Nodes[nodeID] = graphNode
	}
	
	return graph
}

// enrichGraphWithAnalysisData enriches the graph with analysis results
func (da *DependencyAnalyzer) enrichGraphWithAnalysisData(graph *DependencyGraph, tree *DependencyTree) {
	// Add vulnerability counts, license info, and risk scores from dependency analysis
	for _, graphNode := range graph.Nodes {
		if depNode, exists := tree.AllDependencies[graphNode.Name]; exists {
			graphNode.VulnerabilityCount = len(depNode.Vulnerabilities)
			graphNode.LicenseInfo = depNode.License.SPDX
			
			// Calculate risk score based on vulnerabilities and license issues
			riskScore := 0.0
			if len(depNode.Vulnerabilities) > 0 {
				for _, vuln := range depNode.Vulnerabilities {
					// Add to risk score based on severity
					switch vuln.Severity {
					case "critical":
						riskScore += 4.0
					case "high":
						riskScore += 3.0
					case "moderate", "medium":
						riskScore += 2.0
					case "low":
						riskScore += 1.0
					}
				}
			}
			
			// Add license risk - basic check for now
			// TODO: Implement license risk assessment based on license compatibility analysis
			if depNode.License.SPDX == "" {
				riskScore += 0.5 // Unknown license adds some risk
			}
			
			graphNode.RiskScore = math.Min(riskScore, 10.0) // Cap at 10
		}
	}
}

// Helper functions to safely extract metadata from DependencyNode
func getNodeDescription(node *DependencyNode) string {
	if node.PackageInfo != nil {
		return node.PackageInfo.Description
	}
	return ""
}

func getNodeHomepage(node *DependencyNode) string {
	if node.PackageInfo != nil {
		return node.PackageInfo.Homepage
	}
	return ""
}

func getNodeRepository(node *DependencyNode) string {
	if node.PackageInfo != nil {
		return node.PackageInfo.Repository
	}
	return ""
}

// Stub implementations for analysis methods (will be implemented in later tasks)

func (da *DependencyAnalyzer) runVulnerabilityAnalysis(ctx context.Context, tree *DependencyTree) error {
	if da.vulnerabilityDB == nil {
		return fmt.Errorf("vulnerability database not initialized")
	}

	// Collect all packages for vulnerability scanning
	var packages []string
	for name, node := range tree.AllDependencies {
		packageSpec := fmt.Sprintf("%s@%s", name, node.Version)
		packages = append(packages, packageSpec)
	}

	// Scan for vulnerabilities
	vulnerabilities, err := da.vulnerabilityDB.CheckVulnerabilities(ctx, packages)
	if err != nil {
		return fmt.Errorf("failed to check vulnerabilities: %w", err)
	}

	// Initialize security report
	report := tree.SecurityReport
	report.Vulnerabilities = vulnerabilities
	report.TotalVulnerabilities = len(vulnerabilities)
	report.SeverityDistribution = make(map[string]int)

	// Process vulnerabilities and match them to specific packages
	vulnerablePackages := make(map[string]bool)
	var criticalAlerts []*VulnerabilityAlert

	for _, vuln := range vulnerabilities {
		// Count by severity
		report.SeverityDistribution[vuln.Severity]++
		
		switch vuln.Severity {
		case "critical":
			report.CriticalCount++
		case "high":
			report.HighCount++
		case "medium":
			report.MediumCount++
		case "low":
			report.LowCount++
		}

		// Find affected packages and create vulnerability matches
		for packageName, node := range tree.AllDependencies {
			matches, err := da.vulnerabilityDB.MatchVulnerabilities(ctx, packageName, node.Version)
			if err != nil {
				continue // Log error and continue with other packages
			}

			for _, match := range matches {
				if match.Vulnerability.ID == vuln.ID {
					// Add vulnerability to the node
					node.Vulnerabilities = append(node.Vulnerabilities, vuln)
					vulnerablePackages[packageName] = true

					// Generate alert for critical vulnerabilities
					if vuln.CVSS >= da.config.CriticalVulnThreshold {
						alert := da.vulnerabilityDB.GenerateAlert(match)
						criticalAlerts = append(criticalAlerts, alert)
					}
				}
			}
		}
	}

	// Update vulnerable packages list
	report.VulnerablePackages = make([]string, 0, len(vulnerablePackages))
	for pkg := range vulnerablePackages {
		report.VulnerablePackages = append(report.VulnerablePackages, pkg)
	}

	// Generate recommendations
	report.Recommendations = da.generateSecurityRecommendations(vulnerabilities, criticalAlerts)

	// Calculate overall risk score
	report.RiskScore = da.calculateRiskScore(report)

	// Update tree statistics
	tree.Statistics.VulnerablePackages = len(vulnerablePackages)

	// Store critical alerts in metadata for immediate attention
	if tree.RootPackage.Metadata == nil {
		tree.RootPackage.Metadata = make(map[string]interface{})
	}
	if len(criticalAlerts) > 0 {
		tree.RootPackage.Metadata["critical_vulnerability_alerts"] = criticalAlerts
	}

	return nil
}

func (da *DependencyAnalyzer) runLicenseAnalysis(ctx context.Context, tree *DependencyTree) error {
	if da.licenseChecker == nil {
		return fmt.Errorf("license checker not initialized")
	}

	// Collect all packages for license analysis
	var packages []string
	for name, node := range tree.AllDependencies {
		packageSpec := fmt.Sprintf("%s@%s", name, node.Version)
		packages = append(packages, packageSpec)
	}

	// Analyze licenses
	licenseReport, err := da.licenseChecker.CheckLicenses(ctx, packages)
	if err != nil {
		return fmt.Errorf("failed to check licenses: %w", err)
	}

	// Update the tree's license report
	tree.LicenseReport = licenseReport

	// Update individual dependency nodes with license information
	for packageName, node := range tree.AllDependencies {
		if licenseInfo, err := da.licenseChecker.analyzeLicense(ctx, packageName, node.Version); err == nil {
			node.License = LicenseInfo{
				SPDX:         licenseInfo.SPDXIdentifier,
				Name:         licenseInfo.DeclaredLicense,
				Type:         licenseInfo.LicenseType,
				URL:          licenseInfo.LicenseURL,
				Compatibility: fmt.Sprintf("%t", licenseInfo.Compatibility.Compatible),
				Risk:         licenseInfo.RiskLevel,
				Metadata:     map[string]string{
					"declared_license": licenseInfo.DeclaredLicense,
					"spdx_id":         licenseInfo.SPDXIdentifier,
					"license_type":    licenseInfo.LicenseType,
				},
			}
		}
	}

	// Update tree statistics
	tree.Statistics.LicenseDistribution = licenseReport.LicenseDistribution

	// Store license compliance issues in metadata
	if len(licenseReport.CompatibilityIssues) > 0 {
		if tree.RootPackage.Metadata == nil {
			tree.RootPackage.Metadata = make(map[string]interface{})
		}
		tree.RootPackage.Metadata["license_conflicts"] = licenseReport.CompatibilityIssues
	}

	return nil
}

func (da *DependencyAnalyzer) runUpdateAnalysis(ctx context.Context, tree *DependencyTree) error {
	if da.updateChecker == nil {
		return fmt.Errorf("update checker not initialized")
	}

	// Collect all packages for update analysis
	var packages []string
	for name, node := range tree.AllDependencies {
		packageSpec := fmt.Sprintf("%s@%s", name, node.Version)
		packages = append(packages, packageSpec)
	}

	// Check for available updates
	updateReport, err := da.updateChecker.CheckUpdates(ctx, packages)
	if err != nil {
		return fmt.Errorf("failed to check updates: %w", err)
	}

	// Update the tree's update report
	tree.UpdateReport = updateReport

	// Update individual dependency nodes with update information
	for _, updateInfo := range updateReport.Updates {
		packageName := extractPackageNameFromUpdate(updateInfo)
		if node, exists := tree.AllDependencies[packageName]; exists {
			node.UpdateInfo = &updateInfo
		}
	}

	// Update tree statistics
	tree.Statistics.OutdatedPackages = updateReport.OutdatedPackages

	// Store critical updates in metadata
	criticalUpdates := filterCriticalUpdates(updateReport.Updates)
	if len(criticalUpdates) > 0 {
		if tree.RootPackage.Metadata == nil {
			tree.RootPackage.Metadata = make(map[string]interface{})
		}
		tree.RootPackage.Metadata["critical_updates"] = criticalUpdates
	}

	return nil
}

// runPerformanceAnalysis analyzes performance impact of individual packages
func (da *DependencyAnalyzer) runPerformanceAnalysis(ctx context.Context, tree *DependencyTree) error {
	if da.performanceAnalyzer == nil {
		return fmt.Errorf("performance analyzer not initialized")
	}

	var allPerformanceImpacts []PerformanceImpact
	var totalImpactScore float64
	loadTimesByNetwork := make(map[string]float64)

	// Analyze each dependency for performance impact
	for name, node := range tree.AllDependencies {
		// Skip development dependencies for production performance analysis
		if node.Type == "development" {
			continue
		}

		// Create GraphPackageInfo for the performance analyzer
		pkg := &GraphPackageInfo{
			Name:    name,
			Version: node.Version,
			DependencyType: node.Type,
		}
		
		impact, err := da.performanceAnalyzer.AnalyzePackagePerformance(ctx, pkg)
		if err != nil {
			// Log error but continue with other packages
			continue
		}

		allPerformanceImpacts = append(allPerformanceImpacts, *impact)
		totalImpactScore += impact.PerformanceScore

		// Aggregate load times by network type
		if impact.LoadTimeImpact != nil {
			if impact.LoadTimeImpact.Network3G != nil {
				if existing, exists := loadTimesByNetwork["3G"]; exists {
					loadTimesByNetwork["3G"] = existing + impact.LoadTimeImpact.Network3G.TotalTime
				} else {
					loadTimesByNetwork["3G"] = impact.LoadTimeImpact.Network3G.TotalTime
				}
			}
			if impact.LoadTimeImpact.NetworkWiFi != nil {
				if existing, exists := loadTimesByNetwork["WiFi"]; exists {
					loadTimesByNetwork["WiFi"] = existing + impact.LoadTimeImpact.NetworkWiFi.TotalTime
				} else {
					loadTimesByNetwork["WiFi"] = impact.LoadTimeImpact.NetworkWiFi.TotalTime
				}
			}
		}

		// Update the dependency node with performance info
		if node.PackageInfo != nil {
			node.PackageInfo.EstimatedSize = impact.EstimatedSize
		}
	}

	// Update the performance report
	tree.PerformanceReport.Packages = allPerformanceImpacts
	tree.PerformanceReport.AverageLoadTime = loadTimesByNetwork
	if len(allPerformanceImpacts) > 0 {
		tree.PerformanceReport.TotalImpact = totalImpactScore / float64(len(allPerformanceImpacts))
	}

	// Generate overall performance recommendations
	tree.PerformanceReport.Recommendations = da.generatePerformanceRecommendations(allPerformanceImpacts, loadTimesByNetwork)

	return nil
}

// runBundleAnalysis performs comprehensive bundle analysis
func (da *DependencyAnalyzer) runBundleAnalysis(ctx context.Context, tree *DependencyTree) error {
	if da.bundleAnalyzer == nil {
		return fmt.Errorf("bundle analyzer not initialized")
	}

	// Convert dependency tree to bundle analyzer format
	var dependencies []Dependency
	for name, node := range tree.AllDependencies {
		dep := Dependency{
			Name:    name,
			Version: node.Version,
			Type:    node.Type,
		}
		dependencies = append(dependencies, dep)
	}

	// Run bundle analysis
	result, err := da.bundleAnalyzer.AnalyzeBundle(ctx, dependencies)
	if err != nil {
		return fmt.Errorf("failed to analyze bundle: %w", err)
	}

	// Update the tree with bundle analysis results
	tree.BundleResult = result

	// Update legacy bundle analysis for backward compatibility
	tree.BundleAnalysis.EstimatedSize = result.TotalSize
	tree.BundleAnalysis.CompressedSize = result.CompressedSize
	tree.Statistics.TotalSize = result.TotalSize

	// Extract recommendations as strings for backward compatibility
	tree.BundleAnalysis.Recommendations = make([]string, len(result.Recommendations))
	for i, rec := range result.Recommendations {
		tree.BundleAnalysis.Recommendations[i] = rec.Description
	}

	// Update load time estimates
	tree.BundleAnalysis.LoadTimeEstimate = make(map[string]float64)
	tree.BundleAnalysis.LoadTimeEstimate["3G"] = da.estimateLoadTime(result.CompressedSize, "3G")
	tree.BundleAnalysis.LoadTimeEstimate["WiFi"] = da.estimateLoadTime(result.CompressedSize, "WiFi")

	return nil
}

// runBasicBundleEstimation provides basic bundle estimation when full analysis is disabled
func (da *DependencyAnalyzer) runBasicBundleEstimation(tree *DependencyTree) {
	// Basic size estimation based on package count
	estimatedSize := int64(len(tree.AllDependencies) * 50000) // rough estimate of 50KB per package
	tree.BundleAnalysis.EstimatedSize = estimatedSize
	tree.BundleAnalysis.CompressedSize = estimatedSize / 3 // rough gzip ratio
	tree.Statistics.TotalSize = estimatedSize

	// Basic load time estimates
	tree.BundleAnalysis.LoadTimeEstimate = make(map[string]float64)
	tree.BundleAnalysis.LoadTimeEstimate["3G"] = da.estimateLoadTime(estimatedSize, "3G")
	tree.BundleAnalysis.LoadTimeEstimate["WiFi"] = da.estimateLoadTime(estimatedSize, "WiFi")
}

// generatePerformanceRecommendations generates overall performance recommendations
func (da *DependencyAnalyzer) generatePerformanceRecommendations(impacts []PerformanceImpact, loadTimes map[string]float64) []PerformanceRecommendation {
	var recommendations []PerformanceRecommendation

	// Check for overall performance issues
	if loadTimes["3G"] > 10.0 { // More than 10 seconds on 3G
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "load-time",
			Description: "Total load time on 3G networks exceeds 10 seconds. Consider bundle splitting and lazy loading.",
			Priority:    "high",
			ImpactScore: 85.0,
		})
	}

	// Find packages with low performance scores
	var poorPerformers []string
	for _, impact := range impacts {
		if impact.PerformanceScore < 30.0 {
			poorPerformers = append(poorPerformers, impact.PackageName)
		}
	}

	if len(poorPerformers) > 0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "package-alternatives",
			Description: fmt.Sprintf("Consider alternatives for large packages: %v", poorPerformers),
			Priority:    "medium",
			ImpactScore: 70.0,
		})
	}

	// Bundle size recommendation
	totalPackages := len(impacts)
	if totalPackages > 50 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "bundle-optimization",
			Description: fmt.Sprintf("Large number of dependencies (%d). Consider tree-shaking and bundle splitting.", totalPackages),
			Priority:    "medium",
			ImpactScore: 60.0,
		})
	}

	return recommendations
}

// estimateLoadTime provides basic load time estimation
func (da *DependencyAnalyzer) estimateLoadTime(sizeBytes int64, networkType string) float64 {
	// Basic load time calculation based on size and network
	sizeMB := float64(sizeBytes) / 1048576.0 // Convert to MB
	
	switch networkType {
	case "3G":
		return sizeMB * 8.0 // ~8 seconds per MB on 3G
	case "WiFi":
		return sizeMB * 0.5 // ~0.5 seconds per MB on WiFi
	default:
		return sizeMB * 2.0 // Default estimate
	}
}