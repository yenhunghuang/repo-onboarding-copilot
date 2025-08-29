// Package analysis provides monorepo and workspace support for dependency analysis
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceConfig represents workspace configuration
type WorkspaceConfig struct {
	Packages []string               `json:"packages"` // workspace package patterns
	NohoisT  []string               `json:"nohoist"`  // packages that shouldn't be hoisted
	Metadata map[string]interface{} `json:"metadata"`
}

// WorkspacePackage represents a package within a workspace
type WorkspacePackage struct {
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Path          string            `json:"path"` // relative path from workspace root
	AbsolutePath  string            `json:"absolute_path"`
	Manifest      *PackageManifest  `json:"manifest"`
	Dependencies  map[string]string `json:"dependencies"`   // all dependencies combined
	WorkspaceDeps map[string]string `json:"workspace_deps"` // dependencies within workspace
	ExternalDeps  map[string]string `json:"external_deps"`  // external dependencies
	IsPrivate     bool              `json:"is_private"`
}

// MonorepoStructure represents the complete monorepo structure
type MonorepoStructure struct {
	RootPath         string                       `json:"root_path"`
	RootManifest     *PackageManifest             `json:"root_manifest"`
	WorkspaceConfig  *WorkspaceConfig             `json:"workspace_config"`
	Packages         map[string]*WorkspacePackage `json:"packages"`      // keyed by package name
	PackagePaths     []string                     `json:"package_paths"` // all discovered package paths
	LernaConfig      *LernaConfig                 `json:"lerna_config,omitempty"`
	Type             string                       `json:"type"`               // npm-workspaces, yarn-workspaces, lerna
	CrossPackageDeps map[string][]string          `json:"cross_package_deps"` // cross-workspace dependencies
}

// LernaConfig represents lerna.json configuration
type LernaConfig struct {
	Version       string                 `json:"version"`
	Packages      []string               `json:"packages"`
	Command       map[string]interface{} `json:"command"`
	NPMClient     string                 `json:"npmClient"`
	UseWorkspaces bool                   `json:"useWorkspaces"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// detectMonorepoStructure detects if the project is a monorepo and analyzes its structure
func (da *DependencyAnalyzer) detectMonorepoStructure() (*MonorepoStructure, error) {
	structure := &MonorepoStructure{
		RootPath:         da.projectRoot,
		Packages:         make(map[string]*WorkspacePackage),
		PackagePaths:     []string{},
		CrossPackageDeps: make(map[string][]string),
	}

	// Try to parse root package.json
	rootPackagePath := filepath.Join(da.projectRoot, "package.json")
	if manifest, err := da.parsePackageJSON(rootPackagePath); err == nil {
		structure.RootManifest = manifest

		// Check for workspace configuration
		if workspaceConfig, err := da.parseWorkspaceConfig(manifest); err == nil {
			structure.WorkspaceConfig = workspaceConfig
			structure.Type = "npm-workspaces"

			// Detect if it's actually Yarn workspaces
			if _, err := os.Stat(filepath.Join(da.projectRoot, "yarn.lock")); err == nil {
				structure.Type = "yarn-workspaces"
			}
		}
	}

	// Check for lerna.json
	if lernaConfig, err := da.parseLernaConfig(); err == nil {
		structure.LernaConfig = lernaConfig

		// If lerna uses workspaces and we have workspace config, it's lerna-workspaces
		if lernaConfig.UseWorkspaces && structure.WorkspaceConfig != nil {
			structure.Type = "lerna-workspaces"
		} else if structure.WorkspaceConfig == nil {
			// Pure lerna without workspace config
			structure.Type = "lerna"
		}
		// If we have workspaces but lerna doesn't use them, keep the workspace type
	}

	// If no workspace configuration found, return nil (not a monorepo)
	if structure.WorkspaceConfig == nil && structure.LernaConfig == nil {
		return nil, nil
	}

	// Discover workspace packages
	if err := da.discoverWorkspacePackages(structure); err != nil {
		return nil, fmt.Errorf("failed to discover workspace packages: %w", err)
	}

	// Analyze cross-package dependencies
	da.analyzeCrossPackageDependencies(structure)

	return structure, nil
}

// parseWorkspaceConfig parses workspace configuration from package.json
func (da *DependencyAnalyzer) parseWorkspaceConfig(manifest *PackageManifest) (*WorkspaceConfig, error) {
	if manifest.Workspaces == nil {
		return nil, fmt.Errorf("no workspaces configuration found")
	}

	config := &WorkspaceConfig{
		Metadata: make(map[string]interface{}),
	}

	switch ws := manifest.Workspaces.(type) {
	case []interface{}:
		// Simple array format: ["packages/*", "apps/*"]
		for _, pkg := range ws {
			if pkgStr, ok := pkg.(string); ok {
				config.Packages = append(config.Packages, pkgStr)
			}
		}
	case map[string]interface{}:
		// Object format: {"packages": ["packages/*"], "nohoist": [...]}
		if packages, ok := ws["packages"].([]interface{}); ok {
			for _, pkg := range packages {
				if pkgStr, ok := pkg.(string); ok {
					config.Packages = append(config.Packages, pkgStr)
				}
			}
		}
		if nohoist, ok := ws["nohoist"].([]interface{}); ok {
			for _, nh := range nohoist {
				if nhStr, ok := nh.(string); ok {
					config.NohoisT = append(config.NohoisT, nhStr)
				}
			}
		}
		// Store other fields in metadata
		for key, value := range ws {
			if key != "packages" && key != "nohoist" {
				config.Metadata[key] = value
			}
		}
	default:
		return nil, fmt.Errorf("unsupported workspaces configuration format")
	}

	return config, nil
}

// parseLernaConfig parses lerna.json configuration
func (da *DependencyAnalyzer) parseLernaConfig() (*LernaConfig, error) {
	lernaPath := filepath.Join(da.projectRoot, "lerna.json")

	content, err := os.ReadFile(lernaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lerna.json: %w", err)
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(content, &rawConfig); err != nil {
		return nil, fmt.Errorf("invalid JSON in lerna.json: %w", err)
	}

	config := &LernaConfig{
		Metadata: make(map[string]interface{}),
	}

	// Parse known fields
	if version, ok := rawConfig["version"].(string); ok {
		config.Version = version
	}
	if npmClient, ok := rawConfig["npmClient"].(string); ok {
		config.NPMClient = npmClient
	}
	if useWorkspaces, ok := rawConfig["useWorkspaces"].(bool); ok {
		config.UseWorkspaces = useWorkspaces
	}

	// Parse packages array
	if packages, ok := rawConfig["packages"].([]interface{}); ok {
		for _, pkg := range packages {
			if pkgStr, ok := pkg.(string); ok {
				config.Packages = append(config.Packages, pkgStr)
			}
		}
	}

	// Parse command configuration
	if command, ok := rawConfig["command"].(map[string]interface{}); ok {
		config.Command = command
	}

	// Store other fields in metadata
	knownFields := map[string]bool{
		"version": true, "packages": true, "command": true, "npmClient": true, "useWorkspaces": true,
	}
	for key, value := range rawConfig {
		if !knownFields[key] {
			config.Metadata[key] = value
		}
	}

	return config, nil
}

// discoverWorkspacePackages discovers all packages in the workspace
func (da *DependencyAnalyzer) discoverWorkspacePackages(structure *MonorepoStructure) error {
	var patterns []string

	// Get patterns from workspace config
	if structure.WorkspaceConfig != nil {
		patterns = append(patterns, structure.WorkspaceConfig.Packages...)
	}

	// Get patterns from lerna config
	if structure.LernaConfig != nil {
		patterns = append(patterns, structure.LernaConfig.Packages...)
	}

	// If no patterns, use common defaults
	if len(patterns) == 0 {
		patterns = []string{"packages/*", "apps/*", "libs/*"}
	}

	// Discover packages using patterns
	for _, pattern := range patterns {
		if err := da.discoverPackagesByPattern(structure, pattern); err != nil {
			return fmt.Errorf("failed to discover packages with pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// discoverPackagesByPattern discovers packages matching a specific pattern
func (da *DependencyAnalyzer) discoverPackagesByPattern(structure *MonorepoStructure, pattern string) error {
	// Convert workspace pattern to file glob
	searchPath := filepath.Join(da.projectRoot, pattern)

	// Handle different pattern types
	if strings.Contains(pattern, "*") {
		// Glob pattern - find directories that match
		matches, err := filepath.Glob(searchPath)
		if err != nil {
			return fmt.Errorf("glob pattern failed: %w", err)
		}

		for _, match := range matches {
			if err := da.processPackageDirectory(structure, match); err != nil {
				// Log warning but continue processing other packages
				continue
			}
		}
	} else {
		// Direct path
		if err := da.processPackageDirectory(structure, searchPath); err != nil {
			return err
		}
	}

	return nil
}

// processPackageDirectory processes a potential package directory
func (da *DependencyAnalyzer) processPackageDirectory(structure *MonorepoStructure, dirPath string) error {
	// Check if directory exists and is a directory
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a valid directory: %s", dirPath)
	}

	// Check for package.json in this directory
	packageJSONPath := filepath.Join(dirPath, "package.json")
	if _, err := os.Stat(packageJSONPath); err != nil {
		// No package.json, not a package
		return fmt.Errorf("no package.json found in %s", dirPath)
	}

	// Parse the package.json
	manifest, err := da.parsePackageJSON(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to parse package.json in %s: %w", dirPath, err)
	}

	// Get relative path from workspace root
	relPath, err := filepath.Rel(da.projectRoot, dirPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Create workspace package
	workspacePackage := &WorkspacePackage{
		Name:          manifest.Name,
		Version:       manifest.Version,
		Path:          relPath,
		AbsolutePath:  dirPath,
		Manifest:      manifest,
		Dependencies:  make(map[string]string),
		WorkspaceDeps: make(map[string]string),
		ExternalDeps:  make(map[string]string),
		IsPrivate:     manifest.Private,
	}

	// Combine all dependencies
	for name, version := range manifest.Dependencies {
		workspacePackage.Dependencies[name] = version
	}
	for name, version := range manifest.DevDependencies {
		workspacePackage.Dependencies[name] = version
	}
	for name, version := range manifest.PeerDependencies {
		workspacePackage.Dependencies[name] = version
	}
	for name, version := range manifest.OptionalDependencies {
		workspacePackage.Dependencies[name] = version
	}

	// Add to structure
	structure.Packages[manifest.Name] = workspacePackage
	structure.PackagePaths = append(structure.PackagePaths, relPath)

	return nil
}

// analyzeCrossPackageDependencies analyzes dependencies between workspace packages
func (da *DependencyAnalyzer) analyzeCrossPackageDependencies(structure *MonorepoStructure) {
	// Build a map of package names for quick lookup
	packageNames := make(map[string]bool)
	for name := range structure.Packages {
		packageNames[name] = true
	}

	// Analyze each package's dependencies
	for packageName, pkg := range structure.Packages {
		crossDeps := []string{}

		for depName, depVersion := range pkg.Dependencies {
			if packageNames[depName] {
				// This is a workspace dependency
				pkg.WorkspaceDeps[depName] = depVersion
				crossDeps = append(crossDeps, depName)
			} else {
				// This is an external dependency
				pkg.ExternalDeps[depName] = depVersion
			}
		}

		if len(crossDeps) > 0 {
			structure.CrossPackageDeps[packageName] = crossDeps
		}
	}
}

// integrateMonorepoWithDependencyTree integrates monorepo structure with dependency analysis
func (da *DependencyAnalyzer) integrateMonorepoWithDependencyTree(ctx context.Context, tree *DependencyTree, monorepo *MonorepoStructure) error {
	// Add monorepo information to the tree
	if tree.Statistics.TypeDistribution == nil {
		tree.Statistics.TypeDistribution = make(map[string]int)
	}

	// Add workspace packages as special dependency nodes
	for _, pkg := range monorepo.Packages {
		// Create a workspace dependency node
		workspaceNode := &DependencyNode{
			Name:             pkg.Name,
			Version:          pkg.Version,
			RequestedVersion: pkg.Version,
			ResolvedVersion:  pkg.Version,
			Type:             "workspace",
			Path:             pkg.Path,
			Depth:            0,
			IsTransitive:     false,
			Children:         make(map[string]*DependencyNode),
			PackageInfo: &PackageInfo{
				Description: pkg.Manifest.Description,
				Homepage:    pkg.Manifest.Homepage,
				Keywords:    pkg.Manifest.Keywords,
				Deprecated:  false,
			},
			Vulnerabilities: []Vulnerability{},
			License:         LicenseInfo{},
		}

		// Add workspace dependencies as children
		for depName, depVersion := range pkg.WorkspaceDeps {
			if depPkg, exists := monorepo.Packages[depName]; exists {
				childNode := &DependencyNode{
					Name:             depName,
					Version:          depPkg.Version,
					RequestedVersion: depVersion,
					ResolvedVersion:  depPkg.Version,
					Type:             "workspace",
					Path:             depPkg.Path,
					Depth:            1,
					IsTransitive:     false,
					Children:         make(map[string]*DependencyNode),
					Parent:           workspaceNode,
					PackageInfo: &PackageInfo{
						Description: depPkg.Manifest.Description,
						Homepage:    depPkg.Manifest.Homepage,
						Keywords:    depPkg.Manifest.Keywords,
					},
					Vulnerabilities: []Vulnerability{},
					License:         LicenseInfo{},
				}
				workspaceNode.Children[depName] = childNode
			}
		}

		// Add to tree
		tree.AllDependencies[pkg.Name] = workspaceNode
		tree.Statistics.TypeDistribution["workspace"]++
	}

	return nil
}

// extendAnalysisForMonorepo extends the dependency analysis to support monorepo structures
func (da *DependencyAnalyzer) extendAnalysisForMonorepo(ctx context.Context, tree *DependencyTree) error {
	// Detect monorepo structure
	monorepo, err := da.detectMonorepoStructure()
	if err != nil {
		return fmt.Errorf("failed to detect monorepo structure: %w", err)
	}

	// If not a monorepo, return early
	if monorepo == nil {
		return nil
	}

	// Integrate monorepo structure with dependency tree
	if err := da.integrateMonorepoWithDependencyTree(ctx, tree, monorepo); err != nil {
		return fmt.Errorf("failed to integrate monorepo with dependency tree: %w", err)
	}

	// Store monorepo information in tree metadata
	if tree.RootPackage.Metadata == nil {
		tree.RootPackage.Metadata = make(map[string]interface{})
	}
	tree.RootPackage.Metadata["monorepo_structure"] = monorepo
	tree.RootPackage.Metadata["is_monorepo"] = true
	tree.RootPackage.Metadata["monorepo_type"] = monorepo.Type

	return nil
}
