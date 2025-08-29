package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis"
)

// TestDependencyAnalysisIntegration tests the complete dependency analysis workflow
func TestDependencyAnalysisIntegration(t *testing.T) {
	// Create test project directory
	testDir := t.TempDir()

	// Create a realistic package.json
	packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "description": "Integration test project",
  "main": "index.js",
  "dependencies": {
    "react": "^18.2.0",
    "lodash": "^4.17.21",
    "axios": "^0.27.2"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "@types/node": "^18.0.0",
    "typescript": "^4.8.0"
  }
}`

	packageJSONPath := filepath.Join(testDir, "package.json")
	err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create package-lock.json
	packageLockJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "lockfileVersion": 3,
  "requires": true,
  "packages": {
    "": {
      "name": "test-project",
      "version": "1.0.0",
      "dependencies": {
        "react": "^18.2.0",
        "lodash": "^4.17.21",
        "axios": "^0.27.2"
      },
      "devDependencies": {
        "jest": "^29.0.0",
        "@types/node": "^18.0.0",
        "typescript": "^4.8.0"
      }
    },
    "node_modules/react": {
      "version": "18.2.0",
      "resolved": "https://registry.npmjs.org/react/-/react-18.2.0.tgz",
      "integrity": "sha512-/3IjMdb2L9QbBdWiW5e3P2/npwMBaU9mHCSCUzNln0ZCYbcfTsGbTJrU/kGemdH2IWmB2ioZ+zkxtmq6g09fGQ==",
      "dependencies": {
        "loose-envify": "^1.1.0"
      },
      "engines": {
        "node": ">=0.10.0"
      }
    },
    "node_modules/lodash": {
      "version": "4.17.21",
      "resolved": "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz",
      "integrity": "sha512-v2kDEe57lecTulaDIuNTPy3Ry4gLGJ6Z1O3vE1krgXZNrsQ+LFTGHVxVjcXPs17LhbZVGedAJv8XZ1tvj5FvSg=="
    }
  }
}`

	packageLockPath := filepath.Join(testDir, "package-lock.json")
	err = os.WriteFile(packageLockPath, []byte(packageLockJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package-lock.json: %v", err)
	}

	// Initialize dependency analyzer
	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:               testDir,
		IncludePackageFiles:       []string{"package.json", "package-lock.json"},
		EnableVulnScanning:        true,
		EnableLicenseChecking:     true,
		EnableUpdateChecking:      true,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
		MaxDependencyDepth:        5,
		BundleSizeThreshold:       1048576, // 1MB
		PerformanceThreshold:      3000,    // 3 seconds
		CriticalVulnThreshold:     7.0,     // CVSS >= 7.0
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create dependency analyzer: %v", err)
	}

	// Run complete analysis
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Validate results
	validateDependencyAnalysisResults(t, result, testDir)
}

func validateDependencyAnalysisResults(t *testing.T, result *analysis.DependencyTree, testDir string) {
	// Test basic structure
	if result == nil {
		t.Fatal("Analysis result is nil")
	}

	if result.RootPackage == nil {
		t.Fatal("Root package is nil")
	}

	if result.RootPackage.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", result.RootPackage.Name)
	}

	if result.RootPackage.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", result.RootPackage.Version)
	}

	// Test dependencies
	if result.DirectDeps == nil {
		t.Fatal("Direct dependencies is nil")
	}

	expectedDirectDeps := []string{"react", "lodash", "axios", "jest", "@types/node", "typescript"}
	if len(result.DirectDeps) < 3 { // At least production deps
		t.Errorf("Expected at least 3 direct dependencies, got %d", len(result.DirectDeps))
	}

	// Check for specific dependencies using expectedDirectDeps slice
	productionDeps := []string{"react", "lodash", "axios"}
	for _, depName := range productionDeps {
		if _, exists := result.DirectDeps[depName]; !exists {
			t.Errorf("Expected direct dependency '%s' not found", depName)
		}
	}
	
	// Verify we have most of our expected dependencies
	foundCount := 0
	for _, expected := range expectedDirectDeps {
		if _, exists := result.DirectDeps[expected]; exists {
			foundCount++
		}
	}
	if foundCount < len(productionDeps) {
		t.Errorf("Expected to find at least %d of the expected dependencies, found %d", len(productionDeps), foundCount)
	}

	// Test dependency tree structure
	if result.AllDependencies == nil {
		t.Fatal("All dependencies is nil")
	}

	if len(result.AllDependencies) == 0 {
		t.Error("Expected some dependencies in AllDependencies")
	}

	// Test statistics
	if result.Statistics.TotalDependencies == 0 {
		t.Error("Expected non-zero total dependencies")
	}

	if result.Statistics.DirectDependencies == 0 {
		t.Error("Expected non-zero direct dependencies")
	}

	// Test lock file integration
	if result.LockData == nil {
		t.Error("Expected lock file data to be parsed")
	}

	// Validate dependency nodes have required fields
	for name, node := range result.DirectDeps {
		if node.Name != name {
			t.Errorf("Dependency name mismatch: expected '%s', got '%s'", name, node.Name)
		}

		if node.Version == "" {
			t.Errorf("Dependency '%s' has empty version", name)
		}

		if node.PackageInfo == nil {
			t.Errorf("Dependency '%s' has nil PackageInfo", name)
		}

		// Test that direct dependencies are not marked as transitive
		if node.IsTransitive {
			t.Errorf("Direct dependency '%s' is marked as transitive", name)
		}
	}
}

// TestMonorepoIntegration tests dependency analysis in a monorepo structure
func TestMonorepoIntegration(t *testing.T) {
	testDir := t.TempDir()

	// Create root package.json with workspaces
	rootPackageJSON := `{
  "name": "monorepo-root",
  "version": "1.0.0",
  "private": true,
  "workspaces": [
    "packages/*"
  ],
  "devDependencies": {
    "lerna": "^6.0.0",
    "typescript": "^4.8.0"
  }
}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(rootPackageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create root package.json: %v", err)
	}

	// Create packages directory
	packagesDir := filepath.Join(testDir, "packages")
	err = os.MkdirAll(packagesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create packages directory: %v", err)
	}

	// Create first workspace package
	packageADir := filepath.Join(packagesDir, "package-a")
	err = os.MkdirAll(packageADir, 0755)
	if err != nil {
		t.Fatalf("Failed to create package-a directory: %v", err)
	}

	packageAJSON := `{
  "name": "@monorepo/package-a",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}`

	err = os.WriteFile(filepath.Join(packageADir, "package.json"), []byte(packageAJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package-a package.json: %v", err)
	}

	// Create second workspace package with cross-workspace dependency
	packageBDir := filepath.Join(packagesDir, "package-b")
	err = os.MkdirAll(packageBDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create package-b directory: %v", err)
	}

	packageBJSON := `{
  "name": "@monorepo/package-b",
  "version": "1.0.0",
  "dependencies": {
    "@monorepo/package-a": "^1.0.0",
    "axios": "^0.27.2"
  }
}`

	err = os.WriteFile(filepath.Join(packageBDir, "package.json"), []byte(packageBJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package-b package.json: %v", err)
	}

	// Initialize dependency analyzer with monorepo support
	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:               testDir,
		IncludePackageFiles:       []string{"package.json"},
		EnableVulnScanning:        false, // Disable for faster test
		EnableLicenseChecking:     false,
		EnableUpdateChecking:      false,
		EnablePerformanceAnalysis: false,
		EnableBundleAnalysis:      false,
		MaxDependencyDepth:        3,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create dependency analyzer: %v", err)
	}

	// Run analysis
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze monorepo: %v", err)
	}

	// Validate monorepo structure
	if result == nil {
		t.Fatal("Analysis result is nil")
	}

	if result.RootPackage == nil {
		t.Fatal("Root package is nil")
	}

	if result.RootPackage.Name != "monorepo-root" {
		t.Errorf("Expected root name 'monorepo-root', got '%s'", result.RootPackage.Name)
	}

	// Validate workspace detection
	if result.RootPackage.Workspaces == nil {
		t.Error("Expected workspaces to be detected")
	}
}

// TestVulnerabilityIntegration tests vulnerability scanning integration
func TestVulnerabilityIntegration(t *testing.T) {
	testDir := t.TempDir()

	// Create package.json with potentially vulnerable packages
	packageJSON := `{
  "name": "vuln-test-project",
  "version": "1.0.0",
  "dependencies": {
    "minimist": "1.2.5",
    "lodash": "4.17.20"
  }
}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:               testDir,
		IncludePackageFiles:       []string{"package.json"},
		EnableVulnScanning:        true,
		EnableLicenseChecking:     false,
		EnableUpdateChecking:      false,
		EnablePerformanceAnalysis: false,
		EnableBundleAnalysis:      false,
		MaxDependencyDepth:        3,
		CriticalVulnThreshold:     7.0,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create dependency analyzer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for network requests
	defer cancel()

	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Validate vulnerability scanning results
	if result.SecurityReport == nil {
		t.Fatal("Security report is nil")
	}

	// Note: In a real scenario, we might find vulnerabilities
	// For testing purposes, we just verify the structure is correct
	if result.SecurityReport.SeverityDistribution == nil {
		t.Error("Expected severity distribution to be initialized")
	}

	if result.SecurityReport.Vulnerabilities == nil {
		t.Error("Expected vulnerabilities slice to be initialized")
	}

	// Test that vulnerability data is properly attached to dependencies
	for _, dep := range result.DirectDeps {
		if dep.Vulnerabilities == nil {
			t.Errorf("Vulnerability slice for dependency '%s' is nil", dep.Name)
		}
	}
}

// TestLargeProjectPerformance tests performance with a larger dependency tree
func TestLargeProjectPerformance(t *testing.T) {
	testDir := t.TempDir()

	// Create package.json with many dependencies (simulating a large project)
	packageJSON := `{
  "name": "large-test-project",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.3.0",
    "@reduxjs/toolkit": "^1.8.5",
    "react-redux": "^8.0.2",
    "axios": "^0.27.2",
    "lodash": "^4.17.21",
    "moment": "^2.29.4",
    "classnames": "^2.3.1",
    "prop-types": "^15.8.1",
    "styled-components": "^5.3.5",
    "@material-ui/core": "^4.12.4",
    "@material-ui/icons": "^4.11.3"
  },
  "devDependencies": {
    "@types/react": "^18.0.17",
    "@types/react-dom": "^18.0.6",
    "@types/node": "^18.0.0",
    "typescript": "^4.8.0",
    "jest": "^29.0.0",
    "@testing-library/react": "^13.3.0",
    "@testing-library/jest-dom": "^5.16.4",
    "eslint": "^8.22.0",
    "prettier": "^2.7.1"
  }
}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:               testDir,
		IncludePackageFiles:       []string{"package.json"},
		EnableVulnScanning:        false, // Disable for performance test
		EnableLicenseChecking:     false,
		EnableUpdateChecking:      false,
		EnablePerformanceAnalysis: true,
		EnableBundleAnalysis:      true,
		MaxDependencyDepth:        10,
		BundleSizeThreshold:       2097152, // 2MB
		PerformanceThreshold:      5000,    // 5 seconds
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create dependency analyzer: %v", err)
	}

	// Measure analysis time
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze large project: %v", err)
	}

	analysisTime := time.Since(startTime)

	// Performance assertions
	if analysisTime > time.Minute {
		t.Errorf("Analysis took too long: %v (expected < 1 minute)", analysisTime)
	}

	if result.Statistics.TotalDependencies == 0 {
		t.Error("Expected some dependencies to be found")
	}

	// Validate that performance analysis was conducted
	if result.PerformanceReport == nil {
		t.Error("Expected performance report to be generated")
	}

	// Log performance metrics for visibility
	t.Logf("Analysis completed in %v", analysisTime)
	t.Logf("Total dependencies found: %d", result.Statistics.TotalDependencies)
	t.Logf("Direct dependencies: %d", result.Statistics.DirectDependencies)
	t.Logf("Max dependency depth: %d", result.Statistics.MaxDepth)
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	// Test with non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		config := analysis.DependencyAnalyzerConfig{
			ProjectRoot:         "/non/existent/path",
			IncludePackageFiles: []string{"package.json"},
		}

		analyzer, err := analysis.NewDependencyAnalyzer(config)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		ctx := context.Background()
		_, err = analyzer.AnalyzeDependencies(ctx)
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	// Test with malformed package.json
	t.Run("MalformedPackageJSON", func(t *testing.T) {
		testDir := t.TempDir()

		malformedJSON := `{
  "name": "test-project"
  "version": "1.0.0",  // Invalid JSON: missing comma
  "dependencies": {
    "lodash": "^4.17.21"
  }
}`

		err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(malformedJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create malformed package.json: %v", err)
		}

		config := analysis.DependencyAnalyzerConfig{
			ProjectRoot:         testDir,
			IncludePackageFiles: []string{"package.json"},
		}

		analyzer, err := analysis.NewDependencyAnalyzer(config)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		ctx := context.Background()
		_, err = analyzer.AnalyzeDependencies(ctx)
		if err == nil {
			t.Error("Expected error for malformed JSON")
		}
	})

	// Test with empty package.json
	t.Run("EmptyPackageJSON", func(t *testing.T) {
		testDir := t.TempDir()

		emptyJSON := `{}`

		err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(emptyJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty package.json: %v", err)
		}

		config := analysis.DependencyAnalyzerConfig{
			ProjectRoot:         testDir,
			IncludePackageFiles: []string{"package.json"},
		}

		analyzer, err := analysis.NewDependencyAnalyzer(config)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		ctx := context.Background()
		result, err := analyzer.AnalyzeDependencies(ctx)
		if err != nil {
			t.Errorf("Should handle empty package.json gracefully: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result for empty package.json")
		}
	})

	// Test context cancellation
	t.Run("ContextCancellation", func(t *testing.T) {
		testDir := t.TempDir()

		packageJSON := `{
  "name": "test-project",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21"
  }
}`

		err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to create package.json: %v", err)
		}

		config := analysis.DependencyAnalyzerConfig{
			ProjectRoot:         testDir,
			IncludePackageFiles: []string{"package.json"},
			EnableVulnScanning:  true, // Enable to make operation take longer
		}

		analyzer, err := analysis.NewDependencyAnalyzer(config)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err = analyzer.AnalyzeDependencies(ctx)
		if err == nil {
			t.Error("Expected context cancellation error")
		}

		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got: %v", ctx.Err())
		}
	})
}

// TestOutputFormat tests the JSON serialization of results
func TestOutputFormat(t *testing.T) {
	testDir := t.TempDir()

	packageJSON := `{
  "name": "format-test",
  "version": "1.0.0",
  "dependencies": {
    "lodash": "^4.17.21"
  }
}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Test JSON serialization
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatalf("Failed to serialize result to JSON: %v", err)
	}

	// Validate JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	// Check for required fields
	requiredFields := []string{"root_package", "direct_dependencies", "all_dependencies", "statistics"}
	for _, field := range requiredFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("Required field '%s' not found in JSON output", field)
		}
	}

	// Validate that the JSON is not empty
	if len(jsonData) < 100 { // Reasonable minimum size
		t.Errorf("JSON output seems too small: %d bytes", len(jsonData))
	}

	t.Logf("JSON output size: %d bytes", len(jsonData))
}

// BenchmarkDependencyAnalysis benchmarks the dependency analysis performance
func BenchmarkDependencyAnalysis(b *testing.B) {
	testDir := b.TempDir()

	packageJSON := `{
  "name": "benchmark-test",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.2.0",
    "lodash": "^4.17.21",
    "axios": "^0.27.2"
  },
  "devDependencies": {
    "jest": "^29.0.0",
    "typescript": "^4.8.0"
  }
}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		b.Fatalf("Failed to create package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:               testDir,
		IncludePackageFiles:       []string{"package.json"},
		EnableVulnScanning:        false, // Disable for consistent benchmarking
		EnableLicenseChecking:     false,
		EnableUpdateChecking:      false,
		EnablePerformanceAnalysis: false,
		EnableBundleAnalysis:      false,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := analyzer.AnalyzeDependencies(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}
