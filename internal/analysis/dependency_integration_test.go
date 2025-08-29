package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// Helper function to get map keys for debugging
func getKeys(m map[string]*DependencyNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

var testProjectPackageJSON = `{
	"name": "test-integration-project",
	"version": "1.0.0",
	"description": "Integration test project",
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "^4.17.21"
	},
	"devDependencies": {
		"jest": "^28.1.3",
		"webpack": "^5.70.0"
	}
}`

var testProjectLockFile = `{
	"name": "test-integration-project",
	"version": "1.0.0",
	"lockfileVersion": 1,
	"dependencies": {
		"express": {
			"version": "4.18.2",
			"resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz",
			"integrity": "sha512-5/PsL6iGPdfQ/lKM1UuielYgv3BUoJfz1aUwU9vHZ+J7gyvwdQXFEBIEIaxeGf0GIcreATNyBExtalisDbuMqQ==",
			"requires": {
				"accepts": "~1.3.8",
				"cookie": "0.5.0"
			}
		},
		"lodash": {
			"version": "4.17.21",
			"resolved": "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz",
			"integrity": "sha512-v2kDEe57lecTulaDIuNTPy3Ry4gLGJ6Z1O3vE1krgXZNrsQ+LFTGHVxVjcXPs17LhbZVGedAJv8XZ1tvj5FvSg=="
		},
		"jest": {
			"version": "28.1.3",
			"resolved": "https://registry.npmjs.org/jest/-/jest-28.1.3.tgz",
			"integrity": "sha512-N4GT5on8UkZgH0O5LUavMRV1EDEhNTL0KEfRmDIeZHSV7p2XgLoY9t9VDUgL6o+yfdgYHVxuz81G8oB9VG5uyA==",
			"dev": true
		},
		"webpack": {
			"version": "5.70.0",
			"resolved": "https://registry.npmjs.org/webpack/-/webpack-5.70.0.tgz",
			"integrity": "sha512-ZMWWy8CeuTTjCxbeaQI21xSswseF2oNOwc70QSKNePvmxE7XW36i7vpBMYZDuKU5WR1RSHQAkBxjxPjK/pQN3Q==",
			"dev": true
		},
		"accepts": {
			"version": "1.3.8",
			"resolved": "https://registry.npmjs.org/accepts/-/accepts-1.3.8.tgz",
			"integrity": "sha512-PYAthTa2m2VKxuvSD3DPC/Gy+U+sOA1LAuT8mkmRuvw+NACSaeXEQ+NHcVF7rONl6qcaxV3Uuemwawk+7+SJLw=="
		},
		"cookie": {
			"version": "0.5.0",
			"resolved": "https://registry.npmjs.org/cookie/-/cookie-0.5.0.tgz",
			"integrity": "sha512-YZ3GUyn/o8gfKJlnlX7g7xq4gyO6OSuhGPKaaGssGB2qgDUS0gPgtTvoyZLTt9Ab6dC4hfc9dV5arkvc/OCmrw=="
		}
	}
}`

func TestAnalyzeDependenciesForProject(t *testing.T) {
	// Create temporary project structure
	tempDir := t.TempDir()

	// Write project files
	packagePath := filepath.Join(tempDir, "package.json")
	lockPath := filepath.Join(tempDir, "package-lock.json")

	err := os.WriteFile(packagePath, []byte(testProjectPackageJSON), 0644)
	require.NoError(t, err)

	err = os.WriteFile(lockPath, []byte(testProjectLockFile), 0644)
	require.NoError(t, err)

	// Analyze dependencies
	config := DependencyAnalyzerConfig{
		EnableVulnScanning:    false,
		EnableLicenseChecking: false,
		EnableUpdateChecking:  false,
		MaxDependencyDepth:    5,
	}

	tree, err := AnalyzeDependenciesForProject(tempDir, config)
	require.NoError(t, err)
	require.NotNil(t, tree)

	// Verify basic structure
	assert.Equal(t, "test-integration-project", tree.RootPackage.Name)
	assert.Equal(t, "1.0.0", tree.RootPackage.Version)

	// Debug: Print what we found (can be enabled for troubleshooting)
	// t.Logf("Found %d direct dependencies: %v", len(tree.DirectDeps), getKeys(tree.DirectDeps))
	// t.Logf("Found %d total dependencies: %v", len(tree.AllDependencies), getKeys(tree.AllDependencies))

	// Verify direct dependencies
	assert.Len(t, tree.DirectDeps, 4) // express, lodash, jest, webpack
	assert.Contains(t, tree.DirectDeps, "express")
	assert.Contains(t, tree.DirectDeps, "lodash")
	assert.Contains(t, tree.DirectDeps, "jest")
	assert.Contains(t, tree.DirectDeps, "webpack")

	// Verify dependency types
	assert.Equal(t, "production", tree.DirectDeps["express"].Type)
	assert.Equal(t, "production", tree.DirectDeps["lodash"].Type)
	assert.Equal(t, "development", tree.DirectDeps["jest"].Type)
	assert.Equal(t, "development", tree.DirectDeps["webpack"].Type)

	// Verify transitive dependencies
	assert.Contains(t, tree.AllDependencies, "accepts")
	assert.Contains(t, tree.AllDependencies, "cookie")
	assert.True(t, tree.AllDependencies["accepts"].IsTransitive)
	assert.True(t, tree.AllDependencies["cookie"].IsTransitive)

	// Verify statistics
	stats := tree.Statistics
	// t.Logf("Statistics: Total=%d, Direct=%d, Transitive=%d, Dev=%d",
	//	stats.TotalDependencies, stats.DirectDependencies, stats.TransitiveDependencies, stats.DevDependencies)
	assert.Equal(t, 6, stats.TotalDependencies)      // express, lodash, jest, webpack, accepts, cookie
	assert.Equal(t, 4, stats.DirectDependencies)     // express, lodash, jest, webpack
	assert.Equal(t, 2, stats.TransitiveDependencies) // accepts, cookie
	assert.Equal(t, 2, stats.DevDependencies)        // jest, webpack
	assert.True(t, stats.MaxDepth >= 1)

	// Verify dependency graph
	assert.NotNil(t, tree.Graph)
	assert.True(t, len(tree.Graph.Nodes) > 0)
	assert.True(t, len(tree.Graph.Edges) > 0)

	// Verify reports are initialized
	assert.NotNil(t, tree.BundleAnalysis)
	assert.NotNil(t, tree.SecurityReport)
	assert.NotNil(t, tree.LicenseReport)
	assert.NotNil(t, tree.UpdateReport)
}

func TestIntegrateWithAnalysisResult(t *testing.T) {
	// Create temporary project structure
	tempDir := t.TempDir()

	packagePath := filepath.Join(tempDir, "package.json")
	err := os.WriteFile(packagePath, []byte(testProjectPackageJSON), 0644)
	require.NoError(t, err)

	lockPath := filepath.Join(tempDir, "package-lock.json")
	err = os.WriteFile(lockPath, []byte(testProjectLockFile), 0644)
	require.NoError(t, err)

	// Create dependency analyzer
	config := DependencyAnalyzerConfig{
		ProjectRoot:           tempDir,
		EnableVulnScanning:    false,
		EnableLicenseChecking: false,
		EnableUpdateChecking:  false,
	}
	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Create mock AnalysisResult (from AST analyzer)
	existingResult := &ast.AnalysisResult{
		ProjectPath:      tempDir,
		FileResults:      make(map[string]*ast.ParseResult),
		ExternalPackages: make(map[string]ast.ExternalPackage),
		Summary: ast.AnalysisSummary{
			Languages:  make(map[string]int),
			Complexity: ast.ComplexityMetrics{},
		},
	}

	// Integrate dependency analysis
	ctx := context.Background()
	result, err := analyzer.IntegrateWithAnalysisResult(ctx, existingResult)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify integration
	assert.Equal(t, tempDir, result.ProjectPath)
	assert.True(t, len(result.ExternalPackages) > 0)

	// Verify external packages were populated with dependency data
	assert.Contains(t, result.ExternalPackages, "express")
	assert.Contains(t, result.ExternalPackages, "lodash")

	expressPackage := result.ExternalPackages["express"]
	assert.Equal(t, "express", expressPackage.Name)
	assert.Equal(t, "4.18.2", expressPackage.Version)
	assert.Equal(t, "npm", expressPackage.PackageType)
	assert.Contains(t, expressPackage.Metadata, "dependency_type")
	assert.Equal(t, "production", expressPackage.Metadata["dependency_type"])

	// Verify dependency depth was updated in complexity metrics
	assert.True(t, result.Summary.Complexity.DependencyDepth >= 1)
}

func TestGetDependencyMetrics(t *testing.T) {
	// Create minimal dependency tree for testing
	tree := &DependencyTree{
		Statistics: DependencyStats{
			TotalDependencies:      10,
			DirectDependencies:     3,
			TransitiveDependencies: 7,
			DevDependencies:        2,
			MaxDepth:               3,
			TotalSize:              500000,
		},
		SecurityReport: &SecurityReport{
			TotalVulnerabilities: 2,
			CriticalCount:        1,
			HighCount:            1,
			RiskScore:            75.5,
		},
		LicenseReport: &LicenseReport{
			CompatibilityIssues: []LicenseConflict{{Package1: "test"}},
			UnknownLicenses:     []string{"unknown"},
			ProprietaryPackages: []string{"proprietary"},
		},
		UpdateReport: &UpdateReport{
			OutdatedPackages: 5,
			SecurityUpdates:  1,
			BreakingUpdates:  2,
		},
		BundleAnalysis: &BundleAnalysis{
			LoadTimeEstimate: map[string]float64{
				"3g":   3.5,
				"wifi": 0.8,
			},
			PerformanceScore: 85.0,
		},
	}

	metrics := GetDependencyMetrics(tree)

	// Verify metrics extraction
	assert.Equal(t, 10, metrics["total_dependencies"])
	assert.Equal(t, 3, metrics["direct_dependencies"])
	assert.Equal(t, 7, metrics["transitive_dependencies"])
	assert.Equal(t, 2, metrics["dev_dependencies"])
	assert.Equal(t, 3, metrics["max_dependency_depth"])
	assert.Equal(t, int64(500000), metrics["estimated_bundle_size"])

	// Security metrics
	assert.Equal(t, 2, metrics["total_vulnerabilities"])
	assert.Equal(t, 1, metrics["critical_vulnerabilities"])
	assert.Equal(t, 1, metrics["high_vulnerabilities"])
	assert.Equal(t, 75.5, metrics["risk_score"])

	// License metrics
	assert.Equal(t, 1, metrics["license_issues"])
	assert.Equal(t, 1, metrics["unknown_licenses"])
	assert.Equal(t, 1, metrics["proprietary_packages"])

	// Update metrics
	assert.Equal(t, 5, metrics["outdated_packages"])
	assert.Equal(t, 1, metrics["security_updates_available"])
	assert.Equal(t, 2, metrics["breaking_updates_available"])

	// Performance metrics
	assert.Equal(t, 3.5, metrics["estimated_load_time_3g"])
	assert.Equal(t, 0.8, metrics["estimated_load_time_wifi"])
	assert.Equal(t, 85.0, metrics["performance_score"])
}

func TestDeterminePackageType(t *testing.T) {
	analyzer := &DependencyAnalyzer{}

	tests := []struct {
		name        string
		packageName string
		expected    string
	}{
		{
			name:        "scoped package",
			packageName: "@babel/core",
			expected:    "scoped",
		},
		{
			name:        "built-in module",
			packageName: "fs",
			expected:    "built-in",
		},
		{
			name:        "built-in module http",
			packageName: "http",
			expected:    "built-in",
		},
		{
			name:        "regular npm package",
			packageName: "express",
			expected:    "npm",
		},
		{
			name:        "regular npm package with numbers",
			packageName: "lodash4",
			expected:    "npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.determinePackageType(tt.packageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToAnalysisResultFormat(t *testing.T) {
	analyzer := &DependencyAnalyzer{}

	// Create test dependency tree
	directNode := &DependencyNode{
		Name:            "express",
		Version:         "4.18.2",
		Type:            "production",
		IsTransitive:    false,
		Vulnerabilities: []Vulnerability{{ID: "CVE-2023-1234", Severity: "high"}},
		License:         LicenseInfo{SPDX: "MIT"},
	}

	transitiveNode := &DependencyNode{
		Name:            "accepts",
		Version:         "1.3.8",
		Type:            "transitive",
		IsTransitive:    true,
		Vulnerabilities: []Vulnerability{},
		License:         LicenseInfo{SPDX: "MIT"},
	}

	tree := &DependencyTree{
		AllDependencies: map[string]*DependencyNode{
			"express": directNode,
			"accepts": transitiveNode,
		},
	}

	section := analyzer.convertToAnalysisResultFormat(tree)

	// Verify conversion
	assert.Len(t, section.Direct, 1)
	assert.Len(t, section.Transitive, 1)
	assert.Len(t, section.Vulnerabilities, 1)
	assert.Len(t, section.Licenses, 2)

	assert.Equal(t, "express", section.Direct[0].Name)
	assert.Equal(t, "accepts", section.Transitive[0].Name)
	assert.Equal(t, "CVE-2023-1234", section.Vulnerabilities[0].ID)
	assert.Equal(t, "MIT", section.Licenses[0].SPDX)
}
