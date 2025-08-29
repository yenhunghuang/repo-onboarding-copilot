package analysis

import (
	"context"
	"fmt"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// DependencySection represents the dependency analysis section for AnalysisResult
// This matches the expected structure from the story requirements
type DependencySection struct {
	Direct          []DependencyNode `json:"direct"`          // Direct dependencies from package.json
	Transitive      []DependencyNode `json:"transitive"`      // All transitive dependencies
	Vulnerabilities []Vulnerability  `json:"vulnerabilities"` // Security vulnerabilities found
	Licenses        []LicenseInfo    `json:"licenses"`        // License information for all dependencies
}

// IntegrateWithAnalysisResult adds dependency analysis to the existing AnalysisResult
// This allows the dependency analyzer to work with the AST analysis pipeline
func (da *DependencyAnalyzer) IntegrateWithAnalysisResult(ctx context.Context, existingResult *ast.AnalysisResult) (*ast.AnalysisResult, error) {
	// Perform dependency analysis
	dependencyTree, err := da.AnalyzeDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("dependency analysis failed: %w", err)
	}

	// Convert dependency tree to the expected format for AnalysisResult
	_ = da.convertToAnalysisResultFormat(dependencyTree) // Reserved for future use

	// Add dependencies section to existing result
	// Note: This would require modifying the AnalysisResult struct to include a dependencies field
	// For now, we'll store it in the ExternalPackages field and extend that to include our data
	if existingResult.ExternalPackages == nil {
		existingResult.ExternalPackages = make(map[string]ast.ExternalPackage)
	}

	// Convert our dependency nodes to ExternalPackages format for compatibility
	for name, node := range dependencyTree.AllDependencies {
		if !node.IsTransitive || node.Type == "production" || node.Type == "development" {
			existingPackage := ast.ExternalPackage{
				Name:             name,
				Version:          node.Version,
				UsedBy:           []string{da.projectRoot}, // Could be more specific with actual usage
				ImportedFeatures: []string{},               // Would need additional analysis
				PackageType:      da.determinePackageType(name),
				Metadata:         make(map[string]string),
			}

			// Add metadata from our analysis
			existingPackage.Metadata["dependency_type"] = node.Type
			existingPackage.Metadata["requested_version"] = node.RequestedVersion
			existingPackage.Metadata["resolved_version"] = node.ResolvedVersion
			existingPackage.Metadata["depth"] = fmt.Sprintf("%d", node.Depth)

			if node.PackageInfo != nil {
				existingPackage.Metadata["description"] = node.PackageInfo.Description
				existingPackage.Metadata["homepage"] = node.PackageInfo.Homepage
				existingPackage.Metadata["estimated_size"] = fmt.Sprintf("%d", node.PackageInfo.EstimatedSize)
			}

			if len(node.Vulnerabilities) > 0 {
				existingPackage.Metadata["vulnerability_count"] = fmt.Sprintf("%d", len(node.Vulnerabilities))
				existingPackage.Metadata["has_vulnerabilities"] = "true"
			}

			existingPackage.Metadata["license"] = node.License.SPDX

			existingResult.ExternalPackages[name] = existingPackage
		}
	}

	// Update summary with dependency information
	if existingResult.Summary.Languages == nil {
		existingResult.Summary.Languages = make(map[string]int)
	}

	// Add dependency-related metrics to summary
	existingResult.Summary.Complexity.DependencyDepth = dependencyTree.Statistics.MaxDepth

	// Add dependency tree as metadata (could be extended to be a first-class field)
	// This preserves all our detailed analysis for downstream consumers

	return existingResult, nil
}

// convertToAnalysisResultFormat converts DependencyTree to the expected AnalysisResult format
func (da *DependencyAnalyzer) convertToAnalysisResultFormat(tree *DependencyTree) *DependencySection {
	section := &DependencySection{
		Direct:          []DependencyNode{},
		Transitive:      []DependencyNode{},
		Vulnerabilities: []Vulnerability{},
		Licenses:        []LicenseInfo{},
	}

	// Collect direct and transitive dependencies
	for _, node := range tree.AllDependencies {
		if node.IsTransitive {
			section.Transitive = append(section.Transitive, *node)
		} else {
			section.Direct = append(section.Direct, *node)
		}

		// Collect vulnerabilities
		section.Vulnerabilities = append(section.Vulnerabilities, node.Vulnerabilities...)

		// Collect license information
		if node.License.SPDX != "" {
			section.Licenses = append(section.Licenses, node.License)
		}
	}

	return section
}

// determinePackageType determines the package type for ExternalPackage compatibility
func (da *DependencyAnalyzer) determinePackageType(packageName string) string {
	if packageName[0] == '@' {
		return "scoped"
	}

	// Check if it's a built-in Node.js module (basic check)
	builtins := map[string]bool{
		"fs": true, "path": true, "http": true, "https": true, "crypto": true,
		"util": true, "events": true, "stream": true, "buffer": true, "os": true,
		"url": true, "querystring": true, "child_process": true, "cluster": true,
		"dgram": true, "dns": true, "net": true, "tls": true, "zlib": true,
	}

	if builtins[packageName] {
		return "built-in"
	}

	return "npm"
}

// GetDependencyTree provides access to the full dependency tree for external consumers
// This method can be used by other analyzers or report generators
func (da *DependencyAnalyzer) GetDependencyTree(ctx context.Context) (*DependencyTree, error) {
	return da.AnalyzeDependencies(ctx)
}

// AnalyzeDependenciesForProject provides a high-level interface for dependency analysis
// This is the main entry point that other components should use
func AnalyzeDependenciesForProject(projectRoot string, config DependencyAnalyzerConfig) (*DependencyTree, error) {
	config.ProjectRoot = projectRoot

	analyzer, err := NewDependencyAnalyzer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dependency analyzer: %w", err)
	}
	defer analyzer.Close()

	ctx := context.Background()
	return analyzer.AnalyzeDependencies(ctx)
}

// GetDependencyMetrics extracts key metrics from dependency analysis
// This provides a summary view for reporting and monitoring
func GetDependencyMetrics(tree *DependencyTree) map[string]interface{} {
	metrics := make(map[string]interface{})

	stats := tree.Statistics
	metrics["total_dependencies"] = stats.TotalDependencies
	metrics["direct_dependencies"] = stats.DirectDependencies
	metrics["transitive_dependencies"] = stats.TransitiveDependencies
	metrics["dev_dependencies"] = stats.DevDependencies
	metrics["max_dependency_depth"] = stats.MaxDepth
	metrics["estimated_bundle_size"] = stats.TotalSize

	if tree.SecurityReport != nil {
		metrics["total_vulnerabilities"] = tree.SecurityReport.TotalVulnerabilities
		metrics["critical_vulnerabilities"] = tree.SecurityReport.CriticalCount
		metrics["high_vulnerabilities"] = tree.SecurityReport.HighCount
		metrics["risk_score"] = tree.SecurityReport.RiskScore
	}

	if tree.LicenseReport != nil {
		metrics["license_issues"] = len(tree.LicenseReport.CompatibilityIssues)
		metrics["unknown_licenses"] = len(tree.LicenseReport.UnknownLicenses)
		metrics["proprietary_packages"] = len(tree.LicenseReport.ProprietaryPackages)
	}

	if tree.UpdateReport != nil {
		metrics["outdated_packages"] = tree.UpdateReport.OutdatedPackages
		metrics["security_updates_available"] = tree.UpdateReport.SecurityUpdates
		metrics["breaking_updates_available"] = tree.UpdateReport.BreakingUpdates
	}

	if tree.BundleAnalysis != nil {
		metrics["estimated_load_time_3g"] = tree.BundleAnalysis.LoadTimeEstimate["3g"]
		metrics["estimated_load_time_wifi"] = tree.BundleAnalysis.LoadTimeEstimate["wifi"]
		metrics["performance_score"] = tree.BundleAnalysis.PerformanceScore
	}

	return metrics
}
