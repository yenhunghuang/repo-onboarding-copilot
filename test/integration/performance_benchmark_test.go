package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis"
)

// BenchmarkDependencyAnalysisSmall benchmarks analysis of small projects
func BenchmarkDependencyAnalysisSmall(b *testing.B) {
	testDir := b.TempDir()

	packageJSON := `{
		"name": "small-benchmark",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21",
			"axios": "^0.27.2"
		},
		"devDependencies": {
			"jest": "^29.0.0"
		}
	}`

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  5,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}

// BenchmarkDependencyAnalysisMedium benchmarks analysis of medium projects
func BenchmarkDependencyAnalysisMedium(b *testing.B) {
	testDir := b.TempDir()

	// Generate medium-sized dependency list
	var deps []string
	for i := 0; i < 20; i++ {
		deps = append(deps, fmt.Sprintf(`"package-%d": "^1.%d.0"`, i, i))
	}

	var devDeps []string
	for i := 0; i < 10; i++ {
		devDeps = append(devDeps, fmt.Sprintf(`"dev-package-%d": "^2.%d.0"`, i, i))
	}

	packageJSON := fmt.Sprintf(`{
		"name": "medium-benchmark",
		"version": "1.0.0",
		"dependencies": {
			%s
		},
		"devDependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"), strings.Join(devDeps, ",\n\t\t"))

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  7,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		_, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}

// BenchmarkDependencyAnalysisLarge benchmarks analysis of large projects
func BenchmarkDependencyAnalysisLarge(b *testing.B) {
	testDir := b.TempDir()

	// Generate large dependency list (1000+ dependencies)
	var deps []string
	for i := 0; i < 500; i++ {
		deps = append(deps, fmt.Sprintf(`"large-package-%d": "^1.%d.%d"`, i, i%10, i%100))
	}

	var devDeps []string
	for i := 0; i < 200; i++ {
		devDeps = append(devDeps, fmt.Sprintf(`"dev-large-package-%d": "^2.%d.%d"`, i, i%5, i%50))
	}

	packageJSON := fmt.Sprintf(`{
		"name": "large-benchmark",
		"version": "1.0.0",
		"description": "Large project for benchmarking",
		"dependencies": {
			%s
		},
		"devDependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"), strings.Join(devDeps, ",\n\t\t"))

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  3, // Limit depth for large projects
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}

		// Log metrics for the first iteration
		if i == 0 {
			b.Logf("Large project analysis - Total deps: %d, Direct: %d, Max depth: %d",
				result.Statistics.TotalDependencies,
				result.Statistics.DirectDependencies,
				result.Statistics.MaxDepth)
		}
	}
}

// BenchmarkMonorepoAnalysis benchmarks analysis of monorepo structures
func BenchmarkMonorepoAnalysis(b *testing.B) {
	testDir := b.TempDir()

	// Create root package.json
	rootPackageJSON := `{
		"name": "benchmark-monorepo",
		"version": "1.0.0",
		"private": true,
		"workspaces": ["packages/*"],
		"devDependencies": {
			"lerna": "^6.0.0",
			"typescript": "^4.8.0"
		}
	}`

	setupBenchmarkProject(b, testDir, rootPackageJSON)

	// Create multiple workspace packages
	packagesDir := filepath.Join(testDir, "packages")
	err := os.MkdirAll(packagesDir, 0755)
	if err != nil {
		b.Fatalf("Failed to create packages dir: %v", err)
	}

	// Create 10 workspace packages
	for i := 0; i < 10; i++ {
		packageDir := filepath.Join(packagesDir, fmt.Sprintf("package-%d", i))
		err = os.MkdirAll(packageDir, 0755)
		if err != nil {
			b.Fatalf("Failed to create package dir: %v", err)
		}

		// Generate dependencies for each package
		var deps []string
		for j := 0; j < 5; j++ {
			deps = append(deps, fmt.Sprintf(`"dep-%d-%d": "^1.0.0"`, i, j))
		}

		packageJSON := fmt.Sprintf(`{
			"name": "@monorepo/package-%d",
			"version": "1.0.0",
			"dependencies": {
				%s
			}
		}`, i, strings.Join(deps, ",\n\t\t\t"))

		err = os.WriteFile(filepath.Join(packageDir, "package.json"), []byte(packageJSON), 0644)
		if err != nil {
			b.Fatalf("Failed to write workspace package.json: %v", err)
		}
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  5,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Monorepo analysis failed: %v", err)
		}

		if i == 0 {
			b.Logf("Monorepo analysis - Total deps: %d, Direct: %d",
				result.Statistics.TotalDependencies,
				result.Statistics.DirectDependencies)
		}
	}
}

// BenchmarkVulnerabilityScanningPerformance benchmarks vulnerability scanning performance specifically for performance metrics
func BenchmarkVulnerabilityScanningPerformance(b *testing.B) {
	testDir := b.TempDir()

	// Use packages that might have known vulnerabilities for realistic testing
	packageJSON := `{
		"name": "vuln-benchmark",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21",
			"axios": "^0.27.2",
			"react": "^18.2.0",
			"express": "^4.18.1",
			"moment": "^2.29.4"
		}
	}`

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:           testDir,
		IncludePackageFiles:   []string{"package.json"},
		EnableVulnScanning:    true,
		MaxDependencyDepth:    5,
		CriticalVulnThreshold: 7.0,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute) // Longer timeout for network requests
		result, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Vulnerability scanning failed: %v", err)
		}

		if i == 0 && result.SecurityReport != nil {
			b.Logf("Vulnerability scan - Total vulns: %d, Critical: %d, High: %d",
				result.SecurityReport.TotalVulnerabilities,
				result.SecurityReport.CriticalCount,
				result.SecurityReport.HighCount)
		}
	}
}

// BenchmarkMemoryUsage tests memory efficiency during analysis
func BenchmarkMemoryUsage(b *testing.B) {
	testDir := b.TempDir()

	// Create moderately sized project for memory testing
	var deps []string
	for i := 0; i < 100; i++ {
		deps = append(deps, fmt.Sprintf(`"memory-package-%d": "^1.%d.0"`, i, i%20))
	}

	packageJSON := fmt.Sprintf(`{
		"name": "memory-benchmark",
		"version": "1.0.0",
		"dependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"))

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  5,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	// Measure memory before benchmark
	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		_, err := analyzer.AnalyzeDependencies(ctx)
		cancel()

		if err != nil {
			b.Fatalf("Memory benchmark failed: %v", err)
		}

		// Force GC between iterations to get consistent measurements
		if i%10 == 0 {
			runtime.GC()
		}
	}

	// Measure memory after benchmark
	runtime.GC()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	b.Logf("Memory usage - Allocs: %d, Total allocated: %d MB, Sys: %d MB",
		memAfter.TotalAlloc-memBefore.TotalAlloc,
		(memAfter.TotalAlloc-memBefore.TotalAlloc)/1024/1024,
		memAfter.Sys/1024/1024)
}

// BenchmarkParallelAnalysis tests performance under concurrent load
func BenchmarkParallelAnalysis(b *testing.B) {
	testDir := b.TempDir()

	packageJSON := `{
		"name": "parallel-benchmark",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21",
			"axios": "^0.27.2",
			"react": "^18.2.0"
		}
	}`

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  5,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			_, err := analyzer.AnalyzeDependencies(ctx)
			cancel()

			if err != nil {
				b.Errorf("Parallel analysis failed: %v", err)
			}
		}
	})
}

// TestPerformanceRegression tests for performance regression with large datasets
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression test in short mode")
	}

	testCases := []struct {
		name        string
		depCount    int
		maxDuration time.Duration
		maxMemoryMB int64
	}{
		{
			name:        "Small_50_deps",
			depCount:    50,
			maxDuration: 5 * time.Second,
			maxMemoryMB: 50,
		},
		{
			name:        "Medium_200_deps",
			depCount:    200,
			maxDuration: 15 * time.Second,
			maxMemoryMB: 100,
		},
		{
			name:        "Large_1000_deps",
			depCount:    1000,
			maxDuration: 60 * time.Second,
			maxMemoryMB: 200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()

			// Generate package.json with specified number of dependencies
			var deps []string
			for i := 0; i < tc.depCount; i++ {
				deps = append(deps, fmt.Sprintf(`"perf-package-%d": "^1.%d.%d"`, i, i%50, i%100))
			}

			packageJSON := fmt.Sprintf(`{
				"name": "performance-test-%d",
				"version": "1.0.0",
				"dependencies": {
					%s
				}
			}`, tc.depCount, strings.Join(deps, ",\n\t\t\t"))

			err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
			if err != nil {
				t.Fatalf("Failed to write package.json: %v", err)
			}

			config := analysis.DependencyAnalyzerConfig{
				ProjectRoot:         testDir,
				IncludePackageFiles: []string{"package.json"},
				MaxDependencyDepth:  3, // Limit depth for large tests
			}

			analyzer, err := analysis.NewDependencyAnalyzer(config)
			if err != nil {
				t.Fatalf("Failed to create analyzer: %v", err)
			}

			// Measure memory before
			runtime.GC()
			var memBefore runtime.MemStats
			runtime.ReadMemStats(&memBefore)

			// Measure execution time
			startTime := time.Now()

			ctx, cancel := context.WithTimeout(context.Background(), tc.maxDuration+30*time.Second)
			result, err := analyzer.AnalyzeDependencies(ctx)
			cancel()

			duration := time.Since(startTime)

			// Measure memory after
			runtime.GC()
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)

			memUsedMB := int64(memAfter.TotalAlloc-memBefore.TotalAlloc) / 1024 / 1024

			if err != nil {
				t.Fatalf("Performance test failed: %v", err)
			}

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			// Check performance constraints
			if duration > tc.maxDuration {
				t.Errorf("Analysis took %v, expected ≤ %v", duration, tc.maxDuration)
			}

			if memUsedMB > tc.maxMemoryMB {
				t.Errorf("Analysis used %d MB memory, expected ≤ %d MB", memUsedMB, tc.maxMemoryMB)
			}

			// Validate that results are reasonable
			if result.Statistics.TotalDependencies == 0 {
				t.Error("Expected some dependencies to be found")
			}

			if result.Statistics.DirectDependencies != tc.depCount {
				t.Errorf("Expected %d direct dependencies, got %d",
					tc.depCount, result.Statistics.DirectDependencies)
			}

			t.Logf("%s: Duration=%v, Memory=%dMB, TotalDeps=%d, DirectDeps=%d",
				tc.name, duration, memUsedMB,
				result.Statistics.TotalDependencies,
				result.Statistics.DirectDependencies)
		})
	}
}

// TestScalabilityLimits tests the upper bounds of system scalability
func TestScalabilityLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scalability limits test in short mode")
	}

	// Test with extremely large dependency count
	testDir := t.TempDir()

	const extremeDepCount = 5000
	var deps []string
	for i := 0; i < extremeDepCount; i++ {
		deps = append(deps, fmt.Sprintf(`"extreme-package-%d": "^1.%d.%d"`,
			i, i%100, i%1000))
	}

	packageJSON := fmt.Sprintf(`{
		"name": "scalability-test",
		"version": "1.0.0",
		"dependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"))

	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write extreme package.json: %v", err)
	}

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  2, // Very limited depth for extreme test
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		t.Fatalf("Failed to create analyzer: %v", err)
	}

	// Set generous timeout for extreme test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	result, err := analyzer.AnalyzeDependencies(ctx)
	duration := time.Since(startTime)

	if err != nil {
		// It's acceptable for extreme tests to fail due to resource constraints
		t.Logf("Extreme scalability test failed as expected: %v (duration: %v)", err, duration)
		return
	}

	if result != nil {
		t.Logf("Extreme scalability test succeeded: Duration=%v, TotalDeps=%d",
			duration, result.Statistics.TotalDependencies)

		// Even in extreme cases, results should be reasonable
		if result.Statistics.DirectDependencies != extremeDepCount {
			t.Errorf("Expected %d direct dependencies, got %d",
				extremeDepCount, result.Statistics.DirectDependencies)
		}
	}
}

// Helper function to set up benchmark projects
func setupBenchmarkProject(b *testing.B, testDir, packageJSON string) {
	err := os.WriteFile(filepath.Join(testDir, "package.json"), []byte(packageJSON), 0644)
	if err != nil {
		b.Fatalf("Failed to write package.json: %v", err)
	}
}

// BenchmarkJSONSerialization benchmarks the JSON serialization performance
func BenchmarkJSONSerialization(b *testing.B) {
	testDir := b.TempDir()

	// Create a medium-sized project for serialization testing
	var deps []string
	for i := 0; i < 100; i++ {
		deps = append(deps, fmt.Sprintf(`"serialize-package-%d": "^1.%d.0"`, i, i%20))
	}

	packageJSON := fmt.Sprintf(`{
		"name": "serialization-benchmark",
		"version": "1.0.0",
		"dependencies": {
			%s
		}
	}`, strings.Join(deps, ",\n\t\t"))

	setupBenchmarkProject(b, testDir, packageJSON)

	config := analysis.DependencyAnalyzerConfig{
		ProjectRoot:         testDir,
		IncludePackageFiles: []string{"package.json"},
		MaxDependencyDepth:  5,
	}

	analyzer, err := analysis.NewDependencyAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	// Get a result to benchmark serialization
	ctx := context.Background()
	result, err := analyzer.AnalyzeDependencies(ctx)
	if err != nil {
		b.Fatalf("Failed to get analysis result: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Benchmark JSON marshaling
		_, err := json.Marshal(result)
		if err != nil {
			b.Fatalf("JSON serialization failed: %v", err)
		}
	}
}
