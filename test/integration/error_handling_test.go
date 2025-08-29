package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis"
)

// TestErrorHandlingValidation provides comprehensive error handling validation
func TestErrorHandlingValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, analysis.DependencyAnalyzerConfig)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "MissingPackageJSON",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()
				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
				}
				return testDir, config
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "package.json") ||
					strings.Contains(err.Error(), "not found") ||
					strings.Contains(err.Error(), "no such file")
			},
		},
		{
			name: "InvalidJSONSyntax",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()
				invalidJSON := `{
					"name": "test"
					"version": "1.0.0"  // Missing comma
					"dependencies": {
						"lodash": "^4.17.21"
					}
				}`

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(invalidJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write invalid JSON: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
				}
				return testDir, config
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "json") ||
					strings.Contains(err.Error(), "parse") ||
					strings.Contains(err.Error(), "syntax")
			},
		},
		{
			name: "CircularDependencies",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()
				// Create a scenario that might lead to circular dependency detection
				packageJSON := `{
					"name": "circular-test",
					"version": "1.0.0",
					"dependencies": {
						"circular-test": "^1.0.0"
					}
				}`

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write circular JSON: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
					MaxDependencyDepth:  3,
				}
				return testDir, config
			},
			expectError: false, // Should handle gracefully
			errorCheck:  nil,
		},
		{
			name: "InsufficientPermissions",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()
				packageJSON := `{
					"name": "permission-test",
					"version": "1.0.0"
				}`

				packageJSONPath := filepath.Join(testDir, "package.json")
				err := os.WriteFile(packageJSONPath, []byte(packageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write package.json: %v", err)
				}

				// Try to make the file unreadable (may not work on all systems)
				os.Chmod(packageJSONPath, 0000)

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
				}
				return testDir, config
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "permission") ||
					strings.Contains(err.Error(), "denied") ||
					strings.Contains(err.Error(), "access")
			},
		},
		{
			name: "ExtremelyLargeFile",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()

				// Create a large but valid package.json
				var deps []string
				for i := 0; i < 1000; i++ {
					deps = append(deps, fmt.Sprintf(`"package-%d": "^1.0.0"`, i))
				}

				largePackageJSON := fmt.Sprintf(`{
					"name": "large-test",
					"version": "1.0.0",
					"dependencies": {
						%s
					}
				}`, strings.Join(deps, ",\n\t\t"))

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(largePackageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write large JSON: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
					MaxDependencyDepth:  2, // Limit depth to prevent excessive processing
				}
				return testDir, config
			},
			expectError: false, // Should handle large files gracefully
			errorCheck:  nil,
		},
		{
			name: "MalformedLockFile",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()

				validPackageJSON := `{
					"name": "malformed-lock-test",
					"version": "1.0.0",
					"dependencies": {
						"lodash": "^4.17.21"
					}
				}`

				malformedLockFile := `{
					"name": "malformed-lock-test",
					"version": "1.0.0"
					"lockfileVersion": "invalid-version"
					"packages": {
						"": invalid-json-structure
					}
				}`

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(validPackageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write package.json: %v", err)
				}

				err = os.WriteFile(filepath.Join(testDir, "package-lock.json"), []byte(malformedLockFile), 0644)
				if err != nil {
					t.Fatalf("Failed to write malformed lock file: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json", "package-lock.json"},
				}
				return testDir, config
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return strings.Contains(err.Error(), "lock") ||
					strings.Contains(err.Error(), "parse") ||
					strings.Contains(err.Error(), "json")
			},
		},
		{
			name: "NetworkTimeout",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()

				packageJSON := `{
					"name": "network-test",
					"version": "1.0.0",
					"dependencies": {
						"lodash": "^4.17.21"
					}
				}`

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write package.json: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
					EnableVulnScanning:  true, // Enable network operations
				}
				return testDir, config
			},
			expectError: false, // Network errors should be handled gracefully
			errorCheck:  nil,
		},
		{
			name: "UnsupportedPackageFormat",
			setupFunc: func(t *testing.T) (string, analysis.DependencyAnalyzerConfig) {
				testDir := t.TempDir()

				// Create a valid JSON but with unsupported fields
				unsupportedPackageJSON := `{
					"name": "unsupported-test",
					"version": "1.0.0",
					"packageManager": "yarn@3.0.0",
					"exports": {
						".": {
							"import": "./index.mjs",
							"require": "./index.js"
						}
					},
					"dependencies": {
						"lodash": "^4.17.21"
					}
				}`

				err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(unsupportedPackageJSON), 0644)
				if err != nil {
					t.Fatalf("Failed to write unsupported JSON: %v", err)
				}

				config := analysis.DependencyAnalyzerConfig{
					ProjectRoot:         testDir,
					IncludePackageFiles: []string{"package.json"},
				}
				return testDir, config
			},
			expectError: false, // Should handle unknown fields gracefully
			errorCheck:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Restore permissions after each test
			defer func() {
				if tt.name == "InsufficientPermissions" {
					// Try to restore permissions for cleanup
					testDir, _ := tt.setupFunc(t)
					os.Chmod(filepath.Join(testDir, "package.json"), 0644)
				}
			}()

			testDir, config := tt.setupFunc(t)

			analyzer, err := analysis.NewDependencyAnalyzer(config)
			if err != nil {
				if tt.expectError {
					if tt.errorCheck != nil && !tt.errorCheck(err) {
						t.Errorf("Error didn't match expected pattern: %v", err)
					}
					return
				}
				t.Fatalf("Failed to create analyzer: %v", err)
			}

			// Use shorter timeout for error scenarios
			timeout := 30 * time.Second
			if tt.expectError {
				timeout = 10 * time.Second
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			result, err := analyzer.AnalyzeDependencies(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tt.errorCheck != nil && !tt.errorCheck(err) {
					t.Errorf("Error didn't match expected pattern: %v", err)
				}

				t.Logf("Got expected error: %v", err)
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Result should not be nil when no error occurred")
			}

			// Additional validation for successful cases
			validateErrorHandlingResult(t, result, testDir)
		})
	}
}

func validateErrorHandlingResult(t *testing.T, result *analysis.DependencyTree, testDir string) {
	// Basic structure validation
	if result == nil {
		return // Already checked in main test
	}

	// Ensure no nil pointers in critical paths
	if result.Statistics.TotalDependencies < 0 {
		t.Error("Total dependencies should not be negative")
	}

	if result.DirectDeps != nil {
		for name, dep := range result.DirectDeps {
			if dep == nil {
				t.Errorf("Direct dependency '%s' is nil", name)
				continue
			}

			if dep.Name == "" {
				t.Errorf("Dependency '%s' has empty name", name)
			}

			if dep.Version == "" {
				t.Errorf("Dependency '%s' has empty version", name)
			}
		}
	}

	// Validate that optional components handle nil gracefully
	if result.SecurityReport != nil {
		if result.SecurityReport.SeverityDistribution == nil {
			t.Error("Security report has nil severity distribution")
		}
	}

	if result.PerformanceReport != nil {
		if result.PerformanceReport.AverageLoadTime == nil {
			t.Error("Performance report has nil average load time")
		}
	}
}

// TestConcurrentAnalysis tests behavior under concurrent access
func TestConcurrentAnalysis(t *testing.T) {
	testDir := t.TempDir()

	packageJSON := `{
		"name": "concurrent-test",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21",
			"axios": "^0.27.2"
		}
	}`

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Run multiple concurrent analyses
	const numGoroutines = 5
	errChan := make(chan error, numGoroutines)
	resultChan := make(chan *analysis.DependencyTree, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := analyzer.AnalyzeDependencies(ctx)
			if err != nil {
				errChan <- fmt.Errorf("goroutine %d: %w", id, err)
				return
			}

			resultChan <- result
		}(i)
	}

	// Collect results
	var results []*analysis.DependencyTree
	var errors []error

	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errChan:
			errors = append(errors, err)
		case <-time.After(60 * time.Second):
			t.Fatal("Timeout waiting for concurrent analysis")
		}
	}

	// Check for errors
	if len(errors) > 0 {
		t.Errorf("Got %d errors from concurrent analysis:", len(errors))
		for _, err := range errors {
			t.Errorf("  %v", err)
		}
	}

	// Validate all results are consistent
	if len(results) > 1 {
		baseline := results[0]
		for i, result := range results[1:] {
			if result.RootPackage.Name != baseline.RootPackage.Name {
				t.Errorf("Result %d has different project name: %s vs %s",
					i+1, result.RootPackage.Name, baseline.RootPackage.Name)
			}

			if len(result.DirectDeps) != len(baseline.DirectDeps) {
				t.Errorf("Result %d has different number of direct deps: %d vs %d",
					i+1, len(result.DirectDeps), len(baseline.DirectDeps))
			}
		}
	}

	t.Logf("Successfully completed %d concurrent analyses", len(results))
}

// TestMemoryConstraints tests behavior under memory pressure
func TestMemoryConstraints(t *testing.T) {
	testDir := t.TempDir()

	// Create a package.json with many dependencies to test memory usage
	var deps []string
	for i := 0; i < 500; i++ {
		deps = append(deps, fmt.Sprintf(`"package-%d": "^1.0.%d"`, i, i%10))
	}

	largePackageJSON := fmt.Sprintf(`{
		"name": "memory-test",
		"version": "1.0.0",
		"dependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"))

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(largePackageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write large package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  3, // Limit depth to control memory usage
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		t.Fatalf("Analysis failed under memory constraints: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Validate that the result structure is reasonable
	if result.Statistics.TotalDependencies == 0 {
		t.Error("Expected some dependencies to be found")
	}

	if len(result.DirectDeps) == 0 {
		t.Error("Expected some direct dependencies")
	}

	t.Logf("Memory constraint test completed with %d dependencies",
		result.Statistics.TotalDependencies)
}

// TestRecoveryMechanisms tests the system's ability to recover from various failure modes
func TestRecoveryMechanisms(t *testing.T) {
	testCases := []struct {
		name          string
		setupError    func(string) error
		cleanupFunc   func(string) error
		shouldRecover bool
	}{
		{
			name: "TemporaryFilePermission",
			setupError: func(testDir string) error {
				// Create a file and temporarily make it unreadable
				packageJSON := `{"name": "temp-test", "version": "1.0.0"}`
				filePath := filepath.Join(testDir, "package.json")
				err := os.WriteFile(filePath, []byte(packageJSON), 0644)
				if err != nil {
					return err
				}
				return os.Chmod(filePath, 0000)
			},
			cleanupFunc: func(testDir string) error {
				return os.Chmod(filepath.Join(testDir, "package.json"), 0644)
			},
			shouldRecover: true,
		},
		{
			name: "FileDisappears",
			setupError: func(testDir string) error {
				// File will be missing, no setup needed
				return nil
			},
			cleanupFunc: func(testDir string) error {
				// Create the file that was missing
				packageJSON := `{"name": "recovery-test", "version": "1.0.0"}`
				return os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
			},
			shouldRecover: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()

			config := analysis.DependencyAnalyzerConfig{
				ProjectRoot:         testDir,
				IncludePackageFiles: []string{"package.json"},
			}

			analyzer, err := analysis.NewDependencyAnalyzer(config)
			if err != nil {
				t.Fatalf("Failed to create analyzer: %v", err)
			}

			// Set up the error condition
			if err := tc.setupError(testDir); err != nil {
				t.Fatalf("Failed to setup error condition: %v", err)
			}

			// First analysis should fail
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			result1, err1 := analyzer.AnalyzeDependencies(ctx)
			cancel()

			if err1 == nil {
				t.Error("Expected first analysis to fail")
			}

			// Clean up the error condition
			if err := tc.cleanupFunc(testDir); err != nil {
				t.Fatalf("Failed to clean up error condition: %v", err)
			}

			if tc.shouldRecover {
				// Second analysis should succeed
				ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
				result2, err2 := analyzer.AnalyzeDependencies(ctx)
				cancel()

				if err2 != nil {
					t.Errorf("Expected recovery but got error: %v", err2)
				}

				if result2 == nil {
					t.Error("Expected result after recovery")
				}

				t.Logf("Successfully recovered from %s", tc.name)
			}

			// Ensure first result was nil due to error
			if result1 != nil {
				t.Error("First result should be nil when error occurred")
			}
		})
	}
}
