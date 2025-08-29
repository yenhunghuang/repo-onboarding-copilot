// Package analysis provides comprehensive bundle analysis tests
package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBundleAnalyzer(t *testing.T) {
	analyzer := NewBundleAnalyzer()
	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.performanceAnalyzer)

	// Test default budgets are set
	assert.Greater(t, analyzer.budgets.MaxBundleSize, int64(0))
	assert.Greater(t, analyzer.budgets.MaxInitialLoadTime, float64(0))
	assert.Greater(t, analyzer.budgets.MaxScriptEvalTime, float64(0))
}

func TestAnalyzeBundle(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []Dependency
		budgets      *PerformanceBudgets
		expectError  bool
		expectAlerts bool
	}{
		{
			name: "small bundle within budget",
			dependencies: []Dependency{
				{
					Name:    "lodash.get",
					Version: "4.4.2",
					Type:    "dependencies",
				},
				{
					Name:    "is-array",
					Version: "1.0.1",
					Type:    "dependencies",
				},
			},
			budgets: &PerformanceBudgets{
				MaxBundleSize:      2097152, // 2MB
				MaxInitialLoadTime: 3000,    // 3 seconds
				MaxScriptEvalTime:  1000,    // 1 second
			},
			expectError:  false,
			expectAlerts: false,
		},
		{
			name: "large bundle exceeding budget",
			dependencies: []Dependency{
				{
					Name:    "moment",
					Version: "2.29.4",
					Type:    "dependencies",
				},
				{
					Name:    "lodash",
					Version: "4.17.21",
					Type:    "dependencies",
				},
				{
					Name:    "rxjs",
					Version: "7.8.1",
					Type:    "dependencies",
				},
			},
			budgets: &PerformanceBudgets{
				MaxBundleSize:      524288, // 512KB - very tight budget
				MaxInitialLoadTime: 1000,   // 1 second
				MaxScriptEvalTime:  500,    // 0.5 second
			},
			expectError:  false,
			expectAlerts: true, // Should trigger budget violations
		},
		{
			name: "bundle with dev dependencies",
			dependencies: []Dependency{
				{
					Name:    "react",
					Version: "18.2.0",
					Type:    "dependencies",
				},
				{
					Name:    "webpack",
					Version: "5.88.2",
					Type:    "devDependencies", // Should be excluded from production bundle
				},
				{
					Name:    "typescript",
					Version: "5.1.6",
					Type:    "devDependencies",
				},
			},
			budgets: &PerformanceBudgets{
				MaxBundleSize:      1048576, // 1MB
				MaxInitialLoadTime: 3000,    // 3 seconds
				MaxScriptEvalTime:  1500,    // 1.5 seconds
			},
			expectError:  false,
			expectAlerts: false, // Dev deps shouldn't affect production budget
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewBundleAnalyzer()
			if tt.budgets != nil {
				analyzer.budgets = *tt.budgets
			}

			ctx := context.Background()
			result, err := analyzer.AnalyzeBundle(ctx, tt.dependencies)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			// Verify basic result structure
			assert.NotNil(t, result.SizeAnalysis)
			assert.NotNil(t, result.BudgetAnalysis)
			assert.NotNil(t, result.TreeShakingAnalysis)
			assert.NotEmpty(t, result.Recommendations)
			assert.True(t, result.GeneratedAt.After(time.Now().Add(-time.Minute)))

			// Check budget violations match expectations
			hasViolations := len(result.BudgetAnalysis.Violations) > 0
			if tt.expectAlerts {
				assert.True(t, hasViolations, "Expected budget violations but found none")
			}

			// Verify size analysis completeness
			assert.GreaterOrEqual(t, result.SizeAnalysis.TotalSize, int64(0))
			assert.GreaterOrEqual(t, result.SizeAnalysis.ProductionSize, int64(0))
			assert.LessOrEqual(t, result.SizeAnalysis.ProductionSize, result.SizeAnalysis.TotalSize)
		})
	}
}

func TestPerformBudgetAnalysis(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		name                string
		sizeAnalysis        *SizeAnalysis
		budgets             PerformanceBudgets
		expectedViolations  int
		expectedMaxSeverity string
	}{
		{
			name: "within all budgets",
			sizeAnalysis: &SizeAnalysis{
				TotalSize:      400000, // 400KB
				ProductionSize: 300000, // 300KB
				InitialBundle:  200000, // 200KB
			},
			budgets: PerformanceBudgets{
				TotalSize:   524288, // 512KB
				InitialSize: 262144, // 256KB
				AssetSize:   131072, // 128KB
			},
			expectedViolations:  0,
			expectedMaxSeverity: "",
		},
		{
			name: "exceeds total budget",
			sizeAnalysis: &SizeAnalysis{
				TotalSize:      600000, // 600KB
				ProductionSize: 500000, // 500KB
				InitialBundle:  400000, // 400KB
			},
			budgets: PerformanceBudgets{
				TotalSize:   524288, // 512KB
				InitialSize: 262144, // 256KB
				AssetSize:   131072, // 128KB
			},
			expectedViolations:  3, // All budgets exceeded
			expectedMaxSeverity: "critical",
		},
		{
			name: "exceeds only initial budget",
			sizeAnalysis: &SizeAnalysis{
				TotalSize:      400000, // 400KB
				ProductionSize: 300000, // 300KB
				InitialBundle:  300000, // 300KB - exceeds initial budget
			},
			budgets: PerformanceBudgets{
				TotalSize:   524288, // 512KB
				InitialSize: 262144, // 256KB
				AssetSize:   131072, // 128KB
			},
			expectedViolations:  1,
			expectedMaxSeverity: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer.budgets = tt.budgets
			budgetAnalysis := analyzer.performBudgetAnalysis(tt.sizeAnalysis)

			assert.NotNil(t, budgetAnalysis)
			assert.Len(t, budgetAnalysis.Violations, tt.expectedViolations)

			if tt.expectedViolations > 0 {
				assert.Equal(t, tt.expectedMaxSeverity, budgetAnalysis.MaxSeverity)

				// Verify violation details
				for _, violation := range budgetAnalysis.Violations {
					assert.NotEmpty(t, violation.BudgetType)
					assert.Greater(t, violation.ActualSize, violation.BudgetSize)
					assert.Greater(t, violation.OveragePercentage, 0.0)
					assert.NotEmpty(t, violation.Severity)
					assert.NotEmpty(t, violation.Impact)
				}
			} else {
				assert.Equal(t, "none", budgetAnalysis.MaxSeverity)
			}
		})
	}
}

func TestAnalyzeTreeShakingPotential(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		name                   string
		dependencies           []Dependency
		expectedTreeShakable   int
		expectedSavingsPercent float64
	}{
		{
			name: "highly tree-shakable packages",
			dependencies: []Dependency{
				{Name: "lodash", Version: "4.17.21", Type: "dependencies"},
				{Name: "rxjs", Version: "7.8.1", Type: "dependencies"},
				{Name: "date-fns", Version: "2.30.0", Type: "dependencies"},
			},
			expectedTreeShakable:   3,
			expectedSavingsPercent: 30.0, // Should have good savings potential
		},
		{
			name: "mixed tree-shaking potential",
			dependencies: []Dependency{
				{Name: "react", Version: "18.2.0", Type: "dependencies"},   // Limited tree-shaking
				{Name: "lodash", Version: "4.17.21", Type: "dependencies"}, // High tree-shaking
				{Name: "moment", Version: "2.29.4", Type: "dependencies"},  // No tree-shaking
			},
			expectedTreeShakable:   2,
			expectedSavingsPercent: 15.0,
		},
		{
			name: "no tree-shakable packages",
			dependencies: []Dependency{
				{Name: "jquery", Version: "3.7.0", Type: "dependencies"},
				{Name: "moment", Version: "2.29.4", Type: "dependencies"},
			},
			expectedTreeShakable:   0,
			expectedSavingsPercent: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := analyzer.analyzeTreeShakingPotential(tt.dependencies)

			assert.NotNil(t, analysis)
			treeShakableCount := 0
			for _, pkg := range analysis.PackageAnalysis {
				if pkg.IsTreeShakable {
					treeShakableCount++
				}
			}
			assert.GreaterOrEqual(t, treeShakableCount, tt.expectedTreeShakable)
			assert.GreaterOrEqual(t, analysis.PotentialSavings, tt.expectedSavingsPercent)

			// Verify tree-shakable packages have valid data
			for _, pkg := range analysis.PackageAnalysis {
				if pkg.IsTreeShakable {
					assert.NotEmpty(t, pkg.PackageName)
					assert.Greater(t, pkg.TreeShakableSize, int64(0))
					assert.NotEmpty(t, pkg.Recommendations)
				}
			}
		})
	}
}

func TestGenerateBundleRecommendations(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		name                       string
		sizeAnalysis               *SizeAnalysis
		budgetAnalysis             *BudgetAnalysis
		treeShakingAnalysis        *TreeShakingAnalysis
		expectedMinRecommendations int
	}{
		{
			name: "large bundle with violations",
			sizeAnalysis: &SizeAnalysis{
				TotalSize:      2000000, // 2MB
				ProductionSize: 1800000,
				InitialBundle:  1000000,
			},
			budgetAnalysis: &BudgetAnalysis{
				MaxSeverity: "critical",
				Violations: []BudgetViolation{
					{BudgetType: "total", Severity: "critical"},
					{BudgetType: "initial", Severity: "high"},
				},
			},
			treeShakingAnalysis: &TreeShakingAnalysis{
				PotentialSavings: 25.0,
				PackageAnalysis: []PackageTreeShakingInfo{
					{PackageName: "lodash", TreeShakableSize: 40000, IsTreeShakable: true},
				},
			},
			expectedMinRecommendations: 4, // Should have many recommendations
		},
		{
			name: "optimized bundle",
			sizeAnalysis: &SizeAnalysis{
				TotalSize:      200000, // 200KB
				ProductionSize: 180000,
				InitialBundle:  120000,
			},
			budgetAnalysis: &BudgetAnalysis{
				MaxSeverity: "none",
				Violations:  []BudgetViolation{},
			},
			treeShakingAnalysis: &TreeShakingAnalysis{
				PotentialSavings: 5.0,
				PackageAnalysis:  []PackageTreeShakingInfo{},
			},
			expectedMinRecommendations: 1, // Should have minimal recommendations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := analyzer.generateBundleRecommendations(
				tt.sizeAnalysis,
				tt.budgetAnalysis,
				tt.treeShakingAnalysis,
			)

			assert.GreaterOrEqual(t, len(recommendations), tt.expectedMinRecommendations)

			for _, rec := range recommendations {
				assert.NotEmpty(t, rec.Type)
				assert.NotEmpty(t, rec.Description)
				assert.NotEmpty(t, rec.Priority)
				assert.Greater(t, rec.ImpactScore, 0.0)
				assert.LessOrEqual(t, rec.ImpactScore, 100.0)
			}

			// Verify priority values are valid
			validPriorities := map[string]bool{
				"critical": true, "high": true, "medium": true, "low": true,
			}
			for _, rec := range recommendations {
				assert.True(t, validPriorities[rec.Priority], "Invalid priority: %s", rec.Priority)
			}
		})
	}
}

func TestCalculateSizeBreakdown(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	dependencies := []Dependency{
		{Name: "react", Version: "18.2.0", Type: "dependencies"},
		{Name: "lodash", Version: "4.17.21", Type: "dependencies"},
		{Name: "moment", Version: "2.29.4", Type: "dependencies"},
		{Name: "webpack", Version: "5.88.2", Type: "devDependencies"}, // Should be excluded
	}

	ctx := context.Background()
	breakdown := analyzer.calculateSizeBreakdown(ctx, dependencies)

	assert.NotNil(t, breakdown)
	assert.Greater(t, breakdown.TotalSize, int64(0))
	assert.Greater(t, breakdown.ProductionSize, int64(0))
	assert.LessOrEqual(t, breakdown.ProductionSize, breakdown.TotalSize)

	// Should have categorized packages
	assert.NotEmpty(t, breakdown.ByCategory)

	// Verify categories sum to total (within estimation error)
	var categoryTotal int64
	for _, size := range breakdown.ByCategory {
		categoryTotal += size
	}
	// Allow for some estimation variance
	assert.InDelta(t, float64(breakdown.TotalSize), float64(categoryTotal), float64(breakdown.TotalSize)*0.1)
}

func TestPackageCategorization(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		packageName      string
		expectedCategory string
	}{
		{"react", "framework"},
		{"vue", "framework"},
		{"angular", "framework"},
		{"lodash", "utilities"},
		{"ramda", "utilities"},
		{"axios", "networking"},
		{"fetch", "networking"},
		{"moment", "datetime"},
		{"date-fns", "datetime"},
		{"unknown-package", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.packageName, func(t *testing.T) {
			category := analyzer.categorizePackage(tt.packageName)
			assert.Equal(t, tt.expectedCategory, category)
		})
	}
}

func TestBudgetViolationSeverity(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		name             string
		actualSize       int64
		budgetSize       int64
		expectedSeverity string
	}{
		{
			name:             "minor overage",
			actualSize:       110000, // 10% over
			budgetSize:       100000,
			expectedSeverity: "medium",
		},
		{
			name:             "significant overage",
			actualSize:       150000, // 50% over
			budgetSize:       100000,
			expectedSeverity: "high",
		},
		{
			name:             "critical overage",
			actualSize:       250000, // 150% over
			budgetSize:       100000,
			expectedSeverity: "critical",
		},
		{
			name:             "within budget",
			actualSize:       90000, // 10% under
			budgetSize:       100000,
			expectedSeverity: "none", // No violation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := BudgetViolation{
				ActualSize: tt.actualSize,
				BudgetSize: tt.budgetSize,
			}

			severity := analyzer.calculateViolationSeverityFromSize(violation.ActualSize, violation.BudgetSize)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestTreeShakingTechniques(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	tests := []struct {
		packageName        string
		expectedTechniques []string
		minTechniques      int
	}{
		{
			packageName:        "lodash",
			expectedTechniques: []string{"Import specific functions", "Use lodash-es", "Use babel-plugin-lodash"},
			minTechniques:      3,
		},
		{
			packageName:        "rxjs",
			expectedTechniques: []string{"Import operators individually", "Use custom builds"},
			minTechniques:      2,
		},
		{
			packageName:        "moment",
			expectedTechniques: []string{}, // Not tree-shakable
			minTechniques:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.packageName, func(t *testing.T) {
			techniques := analyzer.getTreeShakingTechniques(tt.packageName)
			assert.GreaterOrEqual(t, len(techniques), tt.minTechniques)

			if len(tt.expectedTechniques) > 0 {
				for _, expected := range tt.expectedTechniques {
					assert.Contains(t, techniques, expected)
				}
			}
		})
	}
}

func TestBundleConfigurationDefaults(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	assert.NotNil(t, analyzer.bundlerConfigs)

	// Test that default bundler configs are set
	webpack := analyzer.bundlerConfigs["webpack"]
	assert.NotNil(t, webpack)
	assert.True(t, webpack.TreeShakingEnabled)
	assert.True(t, webpack.MinificationEnabled)
	assert.NotEmpty(t, webpack.OutputFormats)

	rollup := analyzer.bundlerConfigs["rollup"]
	assert.NotNil(t, rollup)
	assert.True(t, rollup.TreeShakingEnabled)

	esbuild := analyzer.bundlerConfigs["esbuild"]
	assert.NotNil(t, esbuild)
	assert.True(t, esbuild.MinificationEnabled)
}

func TestPerformanceBudgetDefaults(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	// Test that default budgets are reasonable
	assert.Equal(t, int64(2097152), analyzer.budgets.TotalSize)  // 2MB
	assert.Equal(t, int64(524288), analyzer.budgets.InitialSize) // 512KB
	assert.Equal(t, int64(262144), analyzer.budgets.AssetSize)   // 256KB

	assert.Greater(t, analyzer.budgets.TotalSize, analyzer.budgets.InitialSize)
	assert.Greater(t, analyzer.budgets.InitialSize, analyzer.budgets.AssetSize)
}

func TestBundleAnalysisWithEmptyDependencies(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	ctx := context.Background()
	result, err := analyzer.AnalyzeBundle(ctx, []Dependency{})

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.SizeAnalysis.TotalSize)
	assert.Equal(t, int64(0), result.SizeAnalysis.ProductionSize)
	assert.Len(t, result.BudgetAnalysis.Violations, 0)
	assert.Equal(t, 0.0, result.TreeShakingAnalysis.PotentialSavings)
}

func TestRecommendationPrioritization(t *testing.T) {
	analyzer := NewBundleAnalyzer()

	// Generate recommendations for a problematic bundle
	sizeAnalysis := &SizeAnalysis{
		TotalSize:      5000000, // 5MB - very large
		ProductionSize: 4500000,
		InitialBundle:  2000000,
	}

	budgetAnalysis := &BudgetAnalysis{
		MaxSeverity: "critical",
		Violations: []BudgetViolation{
			{BudgetType: "total", Severity: "critical"},
			{BudgetType: "initial", Severity: "critical"},
		},
	}

	treeShakingAnalysis := &TreeShakingAnalysis{
		PotentialSavings: 40.0,
		PackageAnalysis: []PackageTreeShakingInfo{
			{PackageName: "lodash", TreeShakableSize: 50000, IsTreeShakable: true},
		},
	}

	recommendations := analyzer.generateBundleRecommendations(
		sizeAnalysis,
		budgetAnalysis,
		treeShakingAnalysis,
	)

	// Should have high-priority recommendations first
	assert.True(t, len(recommendations) > 0)

	// Check that high-impact recommendations are prioritized
	highPriorityFound := false
	for _, rec := range recommendations {
		if rec.Priority == "critical" || rec.Priority == "high" {
			highPriorityFound = true
			break
		}
	}
	assert.True(t, highPriorityFound, "Should have high-priority recommendations for problematic bundle")
}
