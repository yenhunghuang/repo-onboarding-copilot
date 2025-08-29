// Package analysis provides integration tests for performance analysis workflow
package analysis

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceAnalysisIntegration(t *testing.T) {
	// Create temporary directory with test package.json
	tmpDir, err := ioutil.TempDir("", "performance-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test package.json
	packageJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "4.17.21",
			"moment": "2.29.4",
			"react": "18.2.0"
		},
		"devDependencies": {
			"webpack": "5.88.2",
			"typescript": "5.1.6"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create dependency analyzer with performance analysis enabled
	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
		EnableVulnScanning:        false,
		EnableLicenseChecking:     false,
		EnableUpdateChecking:      false,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Run dependency analysis
	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Verify performance analysis results
	assert.NotNil(t, tree.PerformanceReport)
	assert.NotEmpty(t, tree.PerformanceReport.Packages)
	assert.NotEmpty(t, tree.PerformanceReport.AverageLoadTime)

	// Should have performance analysis for production dependencies only
	expectedProdPackages := []string{"lodash", "moment", "react"}
	foundPackages := make(map[string]bool)
	for _, impact := range tree.PerformanceReport.Packages {
		foundPackages[impact.PackageName] = true

		// Verify performance impact structure
		assert.NotEmpty(t, impact.PackageName)
		assert.Greater(t, impact.EstimatedSize, int64(0))
		assert.GreaterOrEqual(t, impact.PerformanceScore, 0.0)
		assert.LessOrEqual(t, impact.PerformanceScore, 100.0)
		assert.NotNil(t, impact.LoadTimeImpact)
		assert.NotEmpty(t, impact.Recommendations)
	}

	// Should have found all production packages
	for _, pkg := range expectedProdPackages {
		assert.True(t, foundPackages[pkg], "Should have performance analysis for %s", pkg)
	}

	// Should have load time analysis by network type
	assert.Contains(t, tree.PerformanceReport.AverageLoadTime, "3G")
	assert.Contains(t, tree.PerformanceReport.AverageLoadTime, "WiFi")
	assert.Greater(t, tree.PerformanceReport.TotalImpact, 0.0)

	// Verify bundle analysis results
	assert.NotNil(t, tree.BundleResult)
	assert.NotNil(t, tree.BundleResult.SizeAnalysis)
	assert.Greater(t, tree.BundleResult.SizeAnalysis.TotalSize, int64(0))
	assert.Greater(t, tree.BundleResult.SizeAnalysis.ProductionSize, int64(0))

	// Verify budget analysis
	assert.NotNil(t, tree.BundleResult.BudgetAnalysis)
	assert.NotEmpty(t, tree.BundleResult.BudgetAnalysis.MaxSeverity)

	// Verify tree-shaking analysis
	assert.NotNil(t, tree.BundleResult.TreeShakingAnalysis)
	assert.GreaterOrEqual(t, tree.BundleResult.TreeShakingAnalysis.PotentialSavings, 0.0)

	// Should have recommendations
	assert.NotEmpty(t, tree.BundleResult.Recommendations)
	for _, rec := range tree.BundleResult.Recommendations {
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.Priority)
		assert.Greater(t, rec.ImpactScore, 0.0)
	}

	// Verify backward compatibility - legacy bundle analysis should be updated
	assert.Greater(t, tree.BundleAnalysis.EstimatedSize, int64(0))
	assert.Greater(t, tree.BundleAnalysis.CompressedSize, int64(0))
	assert.NotEmpty(t, tree.BundleAnalysis.LoadTimeEstimate)
	assert.NotEmpty(t, tree.BundleAnalysis.Recommendations)
}

func TestPerformanceAnalysisDisabled(t *testing.T) {
	// Create temporary directory with test package.json
	tmpDir, err := ioutil.TempDir("", "performance-disabled-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create minimal package.json
	packageJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "4.17.21"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create dependency analyzer with performance analysis disabled
	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: false,
		EnableBundleAnalysis:      false,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Run dependency analysis
	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should have empty performance analysis
	assert.Empty(t, tree.PerformanceReport.Packages)
	assert.Empty(t, tree.PerformanceReport.AverageLoadTime)
	assert.Equal(t, 0.0, tree.PerformanceReport.TotalImpact)

	// Should have basic bundle estimation
	assert.Greater(t, tree.BundleAnalysis.EstimatedSize, int64(0))
	assert.NotEmpty(t, tree.BundleAnalysis.LoadTimeEstimate)
}

func TestBundleAnalysisWithBudgetViolations(t *testing.T) {
	// Create temporary directory with large dependency list
	tmpDir, err := ioutil.TempDir("", "bundle-budget-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create package.json with many large dependencies
	packageJSON := `{
		"name": "large-project",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "4.17.21",
			"moment": "2.29.4",
			"react": "18.2.0",
			"rxjs": "7.8.1",
			"antd": "5.8.4",
			"d3": "7.8.5"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	// Create analyzer with bundle analysis and tight budgets
	config := DependencyAnalyzerConfig{
		ProjectRoot:          tmpDir,
		EnableBundleAnalysis: true,
		BundleSizeThreshold:  100 * 1024, // 100KB - very tight budget
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Override bundle analyzer with tight budgets for testing
	analyzer.bundleAnalyzer = NewBundleAnalyzer()
	analyzer.bundleAnalyzer.budgets = PerformanceBudgets{
		TotalSize:   512 * 1024, // 512KB
		InitialSize: 256 * 1024, // 256KB
		AssetSize:   128 * 1024, // 128KB
	}

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should detect budget violations due to many large packages
	assert.NotEmpty(t, tree.BundleResult.BudgetAnalysis.Violations)
	assert.NotEqual(t, "none", tree.BundleResult.BudgetAnalysis.MaxSeverity)

	// Should have optimization recommendations
	hasOptimizationRec := false
	for _, rec := range tree.BundleResult.Recommendations {
		if rec.Type == "bundle-size" || rec.Type == "tree-shaking" {
			hasOptimizationRec = true
			break
		}
	}
	assert.True(t, hasOptimizationRec, "Should have bundle optimization recommendations")
}

func TestTreeShakingAnalysis(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "tree-shaking-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Package.json with tree-shakable dependencies
	packageJSON := `{
		"name": "tree-shaking-project",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "4.17.21",
			"rxjs": "7.8.1",
			"date-fns": "2.30.0"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	config := DependencyAnalyzerConfig{
		ProjectRoot:          tmpDir,
		EnableBundleAnalysis: true,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should identify tree-shakable packages
	treeShaking := tree.BundleResult.TreeShakingAnalysis
	assert.Greater(t, treeShaking.PotentialSavings, 0.0)
	assert.NotEmpty(t, treeShaking.PackageAnalysis)

	// Check for tree-shakable packages
	treeShakableNames := make(map[string]bool)
	for _, pkg := range treeShaking.PackageAnalysis {
		if pkg.IsTreeShakable {
			treeShakableNames[pkg.PackageName] = true
			assert.Greater(t, pkg.TreeShakableSize, int64(0))
			assert.NotEmpty(t, pkg.Recommendations)
		}
	}

	// lodash should be identified as tree-shakable
	assert.True(t, treeShakableNames["lodash"], "lodash should be tree-shakable")
}

func TestPerformanceRecommendationGeneration(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "recommendations-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create package.json with performance issues
	packageJSON := `{
		"name": "slow-project",
		"version": "1.0.0",
		"dependencies": {
			"moment": "2.29.4",
			"lodash": "4.17.21"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should have performance recommendations
	assert.NotEmpty(t, tree.PerformanceReport.Recommendations)

	// Verify recommendation structure
	for _, rec := range tree.PerformanceReport.Recommendations {
		assert.NotEmpty(t, rec.Type)
		assert.NotEmpty(t, rec.Description)
		assert.NotEmpty(t, rec.Priority)
		assert.Greater(t, rec.ImpactScore, 0.0)
		assert.LessOrEqual(t, rec.ImpactScore, 100.0)

		// Priority should be valid
		validPriorities := map[string]bool{
			"high": true, "medium": true, "low": true,
		}
		assert.True(t, validPriorities[rec.Priority], "Invalid priority: %s", rec.Priority)
	}
}

func TestAnalysisResultSerialization(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "serialization-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	packageJSON := `{
		"name": "serialization-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "18.2.0"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Test JSON serialization of complete dependency tree
	jsonData, err := json.MarshalIndent(tree, "", "  ")
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify that key performance analysis fields are present
	var result map[string]interface{}
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)

	// Check performance report structure
	perfReport, exists := result["performance_report"]
	assert.True(t, exists)

	perfMap, ok := perfReport.(map[string]interface{})
	require.True(t, ok)

	_, hasPackages := perfMap["packages"]
	_, hasLoadTime := perfMap["average_load_time"]
	_, hasRecommendations := perfMap["recommendations"]

	assert.True(t, hasPackages, "Should have performance packages data")
	assert.True(t, hasLoadTime, "Should have load time data")
	assert.True(t, hasRecommendations, "Should have performance recommendations")

	// Check bundle result structure
	bundleResult, exists := result["bundle_result"]
	assert.True(t, exists)

	bundleMap, ok := bundleResult.(map[string]interface{})
	require.True(t, ok)

	_, hasSizeAnalysis := bundleMap["size_analysis"]
	_, hasBudgetAnalysis := bundleMap["budget_analysis"]
	_, hasTreeShaking := bundleMap["tree_shaking_analysis"]

	assert.True(t, hasSizeAnalysis, "Should have size analysis data")
	assert.True(t, hasBudgetAnalysis, "Should have budget analysis data")
	assert.True(t, hasTreeShaking, "Should have tree-shaking analysis data")
}

func TestPerformanceAnalysisWithEmptyDependencies(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "empty-deps-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Package.json with no dependencies
	packageJSON := `{
		"name": "empty-project",
		"version": "1.0.0",
		"dependencies": {}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should handle empty dependencies gracefully
	assert.Empty(t, tree.PerformanceReport.Packages)
	assert.Equal(t, 0.0, tree.PerformanceReport.TotalImpact)
	assert.Equal(t, int64(0), tree.BundleResult.SizeAnalysis.TotalSize)
	assert.Equal(t, int64(0), tree.BundleResult.SizeAnalysis.ProductionSize)
	assert.Equal(t, "none", tree.BundleResult.BudgetAnalysis.MaxSeverity)
}

func TestPerformanceAnalysisTimeout(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "timeout-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	packageJSON := `{
		"name": "timeout-project",
		"version": "1.0.0",
		"dependencies": {
			"unknown-package": "1.0.0"
		}
	}`

	packagePath := filepath.Join(tmpDir, "package.json")
	err = ioutil.WriteFile(packagePath, []byte(packageJSON), 0644)
	require.NoError(t, err)

	config := DependencyAnalyzerConfig{
		ProjectRoot:               tmpDir,
		EnablePerformanceAnalysis: true,
	}

	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Use short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// Should still complete analysis even with timeout (fallback estimation)
	tree, err := analyzer.AnalyzeDependencies(ctx)
	require.NoError(t, err)

	// Should have empty or fallback performance analysis
	assert.NotNil(t, tree.PerformanceReport)
}
