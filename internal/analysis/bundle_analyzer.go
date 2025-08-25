// Package analysis provides bundle-level performance analysis
package analysis

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// BundleAnalyzer analyzes overall bundle performance and provides optimization recommendations
type BundleAnalyzer struct {
	performanceAnalyzer *PerformanceAnalyzer
	bundlerConfig       *BundlerConfig
	budgets            PerformanceBudgets
	bundlerConfigs     map[string]*BundlerConfig
}

// NewBundleAnalyzer creates a new bundle analyzer
func NewBundleAnalyzer() *BundleAnalyzer {
	defaultBudgets := getDefaultPerformanceBudgets()
	
	bundlerConfigs := map[string]*BundlerConfig{
		"webpack": {
			Type:                "webpack",
			TreeShakingEnabled:  true,
			CompressionEnabled:  true,
			MinificationEnabled: true,
			CodeSplitting:       true,
			TreeShakingRatio:    0.3,
			CompressionRatio:    0.7,
			OutputFormats:       []string{"js", "css"},
		},
		"rollup": {
			Type:                "rollup", 
			TreeShakingEnabled:  true,
			CompressionEnabled:  true,
			MinificationEnabled: false,
			CodeSplitting:       true,
			TreeShakingRatio:    0.4,
			CompressionRatio:    0.7,
			OutputFormats:       []string{"esm", "cjs"},
		},
		"esbuild": {
			Type:                "esbuild",
			TreeShakingEnabled:  true,
			CompressionEnabled:  false,
			MinificationEnabled: true,
			CodeSplitting:       false,
			TreeShakingRatio:    0.35,
			CompressionRatio:    0.8,
			OutputFormats:       []string{"js"},
		},
	}

	return &BundleAnalyzer{
		performanceAnalyzer: NewPerformanceAnalyzer(),
		bundlerConfig: bundlerConfigs["webpack"], // default to webpack
		budgets: *defaultBudgets,
		bundlerConfigs: bundlerConfigs,
	}
}

// AnalyzeBundle performs comprehensive bundle analysis
func (ba *BundleAnalyzer) AnalyzeBundle(ctx context.Context, dependencies []Dependency) (*BundleAnalysisResult, error) {
	result := &BundleAnalysisResult{
		PackageContributions: make([]PackageContribution, 0),
		Recommendations:      make([]PerformanceRecommendation, 0),
		GeneratedAt:          time.Now(),
	}

	// Analyze each package
	var packageImpacts []*PerformanceImpact
	totalSize := int64(0)
	totalMinified := int64(0)
	totalCompressed := int64(0)
	totalTreeShakable := int64(0)
	productionSize := int64(0)

	for _, dep := range dependencies {
		if dep.Type == "devDependencies" {
			continue // Skip dev dependencies in production bundle
		}

		impact := ba.estimatePackageImpactFromDependency(dep)

		packageImpacts = append(packageImpacts, impact)
		totalSize += impact.EstimatedSize
		totalMinified += impact.MinifiedSize
		totalCompressed += impact.CompressedSize
		totalTreeShakable += impact.TreeShakableSize
		productionSize += impact.EstimatedSize
	}

	// Calculate bundle totals
	result.TotalSize = totalSize
	result.MinifiedSize = totalMinified
	result.CompressedSize = totalCompressed
	result.TreeShakableSize = totalTreeShakable
	result.OptimizedSize = ba.calculateOptimizedSize(totalSize, totalTreeShakable)

	// Populate SizeAnalysis
	result.SizeAnalysis = &SizeAnalysis{
		TotalSize:      totalSize,
		ProductionSize: productionSize,
		InitialBundle:  totalSize, // For now, assume all is initial bundle
	}

	// Populate TreeShakingAnalysis
	result.TreeShakingAnalysis = ba.analyzeTreeShakingPotential(dependencies)

	// Calculate package contributions
	result.PackageContributions = ba.calculatePackageContributions(packageImpacts, totalSize)

	// Perform budget analysis
	result.BudgetAnalysis = ba.performBudgetAnalysis(result.SizeAnalysis)

	// Calculate load time analysis for the bundle
	result.LoadTimeAnalysis = ba.performanceAnalyzer.calculateLoadTimeImpact(
		result.TotalSize,
		result.MinifiedSize,
		result.CompressedSize,
	)

	// Generate size breakdown
	result.SizeBreakdown = ba.generateSizeBreakdown(packageImpacts)

	// Generate recommendations
	result.Recommendations = ba.generateBundleRecommendations(result.SizeAnalysis, result.BudgetAnalysis, result.TreeShakingAnalysis)

	return result, nil
}

// estimatePackageImpactFromDependency provides fallback impact estimation from Dependency
func (ba *BundleAnalyzer) estimatePackageImpactFromDependency(dep Dependency) *PerformanceImpact {
	// Basic estimation logic
	baseSize := int64(50000) // 50KB default
	
	// Adjust based on package name patterns
	if strings.Contains(dep.Name, "react") || strings.Contains(dep.Name, "vue") {
		baseSize = 150000
	} else if strings.Contains(dep.Name, "lodash") || strings.Contains(dep.Name, "moment") {
		baseSize = 200000
	} else if strings.HasPrefix(dep.Name, "@types/") {
		baseSize = 5000
	}

	minifiedSize := int64(float64(baseSize) * 0.7)
	compressedSize := int64(float64(minifiedSize) * 0.7)
	treeShakableSize := int64(float64(baseSize) * 0.3)

	return &PerformanceImpact{
		PackageName:      dep.Name,
		EstimatedSize:    baseSize,
		MinifiedSize:     minifiedSize,
		CompressedSize:   compressedSize,
		TreeShakableSize: treeShakableSize,
		LoadTimeImpact:   ba.performanceAnalyzer.calculateLoadTimeImpact(baseSize, minifiedSize, compressedSize),
		PerformanceScore: 75.0, // neutral score
		Recommendations:  []string{},
		Metadata:         make(map[string]interface{}),
	}
}

// estimatePackageImpact provides fallback impact estimation
func (ba *BundleAnalyzer) estimatePackageImpact(pkg *GraphPackageInfo) *PerformanceImpact {
	// Basic estimation logic
	baseSize := int64(50000) // 50KB default
	
	// Adjust based on package name patterns
	if strings.Contains(pkg.Name, "react") || strings.Contains(pkg.Name, "vue") {
		baseSize = 150000
	} else if strings.Contains(pkg.Name, "lodash") || strings.Contains(pkg.Name, "moment") {
		baseSize = 200000
	} else if strings.HasPrefix(pkg.Name, "@types/") {
		baseSize = 5000
	}

	minifiedSize := int64(float64(baseSize) * 0.7)
	compressedSize := int64(float64(minifiedSize) * 0.7)
	treeShakableSize := int64(float64(baseSize) * 0.3)

	return &PerformanceImpact{
		PackageName:      pkg.Name,
		EstimatedSize:    baseSize,
		MinifiedSize:     minifiedSize,
		CompressedSize:   compressedSize,
		TreeShakableSize: treeShakableSize,
		LoadTimeImpact:   ba.performanceAnalyzer.calculateLoadTimeImpact(baseSize, minifiedSize, compressedSize),
		PerformanceScore: 75.0, // neutral score
		Recommendations:  []string{},
		Metadata:         make(map[string]interface{}),
	}
}

// calculateOptimizedSize calculates the size after all optimizations
func (ba *BundleAnalyzer) calculateOptimizedSize(totalSize, treeShakableSize int64) int64 {
	optimizedSize := totalSize

	// Apply tree shaking if enabled
	if ba.bundlerConfig.TreeShakingEnabled {
		optimizedSize -= treeShakableSize
	}

	// Apply code splitting benefits (reduces initial load)
	if ba.bundlerConfig.CodeSplitting {
		optimizedSize = int64(float64(optimizedSize) * 0.7) // 30% reduction in initial load
	}

	return optimizedSize
}

// calculatePackageContributions calculates how much each package contributes to the bundle
func (ba *BundleAnalyzer) calculatePackageContributions(impacts []*PerformanceImpact, totalSize int64) []PackageContribution {
	contributions := make([]PackageContribution, 0)

	for _, impact := range impacts {
		percentage := 0.0
		if totalSize > 0 {
			percentage = float64(impact.EstimatedSize) / float64(totalSize) * 100
		}

		impactLevel := "low"
		if percentage > 20 {
			impactLevel = "high"
		} else if percentage > 10 {
			impactLevel = "medium"
		}

		contribution := PackageContribution{
			PackageName:    impact.PackageName,
			Size:           impact.EstimatedSize,
			Percentage:     percentage,
			IsTreeShakable: impact.TreeShakableSize > 0,
			Impact:         impactLevel,
		}

		contributions = append(contributions, contribution)
	}

	// Sort by size (descending)
	sort.Slice(contributions, func(i, j int) bool {
		return contributions[i].Size > contributions[j].Size
	})

	return contributions
}

// performBudgetAnalysis compares actual metrics against performance budgets (SizeAnalysis version)
func (ba *BundleAnalyzer) performBudgetAnalysis(sizeAnalysis *SizeAnalysis) *BudgetAnalysis {
	analysis := &BudgetAnalysis{
		Violations:  make([]BudgetViolation, 0),
		MaxSeverity: "none",
	}

	// Check total size budget
	if sizeAnalysis.TotalSize > ba.budgets.TotalSize {
		severity := ba.calculateViolationSeverityFromSize(sizeAnalysis.TotalSize, ba.budgets.TotalSize)
		violation := BudgetViolation{
			BudgetType:        "total",
			ActualSize:        sizeAnalysis.TotalSize,
			BudgetSize:        ba.budgets.TotalSize,
			OveragePercentage: float64(sizeAnalysis.TotalSize-ba.budgets.TotalSize) / float64(ba.budgets.TotalSize) * 100,
			Severity:          severity,
			Impact:            ba.describeSizeImpact(sizeAnalysis.TotalSize, ba.budgets.TotalSize),
		}
		analysis.Violations = append(analysis.Violations, violation)
		if ba.isHigherSeverity(severity, analysis.MaxSeverity) {
			analysis.MaxSeverity = severity
		}
	}

	// Check initial size budget
	if sizeAnalysis.InitialBundle > ba.budgets.InitialSize {
		severity := ba.calculateViolationSeverityFromSize(sizeAnalysis.InitialBundle, ba.budgets.InitialSize)
		violation := BudgetViolation{
			BudgetType:        "initial",
			ActualSize:        sizeAnalysis.InitialBundle,
			BudgetSize:        ba.budgets.InitialSize,
			OveragePercentage: float64(sizeAnalysis.InitialBundle-ba.budgets.InitialSize) / float64(ba.budgets.InitialSize) * 100,
			Severity:          severity,
			Impact:            ba.describeSizeImpact(sizeAnalysis.InitialBundle, ba.budgets.InitialSize),
		}
		analysis.Violations = append(analysis.Violations, violation)
		if ba.isHigherSeverity(severity, analysis.MaxSeverity) {
			analysis.MaxSeverity = severity
		}
	}

	return analysis
}

// performBudgetAnalysisLegacy compares actual metrics against performance budgets (legacy BundleAnalysisResult version)
func (ba *BundleAnalyzer) performBudgetAnalysisLegacy(result *BundleAnalysisResult) *BudgetAnalysis {
	analysis := &BudgetAnalysis{
		Violations: make([]BudgetViolation, 0),
	}

	// Check bundle size budget
	budgetUtilization := float64(result.CompressedSize) / float64(ba.budgets.MaxBundleSize) * 100
	analysis.BudgetUtilization = budgetUtilization

	if result.CompressedSize > ba.budgets.MaxBundleSize {
		analysis.BundleSizeStatus = "fail"
		analysis.OverBudgetBy = result.CompressedSize - ba.budgets.MaxBundleSize
		
		violation := BudgetViolation{
			Metric:   "bundle_size",
			Actual:   float64(result.CompressedSize),
			Budget:   float64(ba.budgets.MaxBundleSize),
			Severity: ba.calculateViolationSeverity(budgetUtilization),
			Impact:   ba.describeBundleSizeImpact(result.CompressedSize, ba.budgets.MaxBundleSize),
		}
		analysis.Violations = append(analysis.Violations, violation)
	} else if budgetUtilization > 80 {
		analysis.BundleSizeStatus = "warn"
	} else {
		analysis.BundleSizeStatus = "pass"
	}

	// Check load time budget (using WiFi + mid-end device as baseline)
	if result.LoadTimeAnalysis != nil && result.LoadTimeAnalysis.NetworkWiFi != nil {
		loadTime := result.LoadTimeAnalysis.NetworkWiFi.TotalTime
		if loadTime > ba.budgets.MaxInitialLoadTime {
			analysis.LoadTimeStatus = "fail"
			
			violation := BudgetViolation{
				Metric:   "load_time",
				Actual:   loadTime,
				Budget:   ba.budgets.MaxInitialLoadTime,
				Severity: ba.calculateViolationSeverity(loadTime/ba.budgets.MaxInitialLoadTime*100),
				Impact:   ba.describeLoadTimeImpact(loadTime, ba.budgets.MaxInitialLoadTime),
			}
			analysis.Violations = append(analysis.Violations, violation)
		} else if loadTime > ba.budgets.MaxInitialLoadTime*0.8 {
			analysis.LoadTimeStatus = "warn"
		} else {
			analysis.LoadTimeStatus = "pass"
		}
	}

	return analysis
}

// calculateViolationSeverity determines the severity of a budget violation
func (ba *BundleAnalyzer) calculateViolationSeverity(utilizationPercent float64) string {
	if utilizationPercent > 200 {
		return "critical"
	} else if utilizationPercent > 150 {
		return "high"
	} else if utilizationPercent > 120 {
		return "medium"
	} else {
		return "low"
	}
}

// describeBundleSizeImpact describes the user impact of bundle size violations
func (ba *BundleAnalyzer) describeBundleSizeImpact(actual, budget int64) string {
	overBy := actual - budget
	percentOver := float64(overBy) / float64(budget) * 100

	if percentOver > 100 {
		return "Bundle is more than double the recommended size, causing significant load time increases on slower connections"
	} else if percentOver > 50 {
		return "Bundle size significantly exceeds budget, noticeably impacting load performance"
	} else if percentOver > 20 {
		return "Bundle size moderately exceeds budget, may affect performance on slower connections"
	} else {
		return "Bundle size slightly exceeds budget, minimal performance impact"
	}
}

// describeLoadTimeImpact describes the user impact of load time violations
func (ba *BundleAnalyzer) describeLoadTimeImpact(actual, budget float64) string {
	overBy := actual - budget
	percentOver := overBy / budget * 100

	if percentOver > 100 {
		return "Load time is more than double the recommended threshold, severely impacting user experience"
	} else if percentOver > 50 {
		return "Load time significantly exceeds budget, causing poor user experience"
	} else if percentOver > 20 {
		return "Load time moderately exceeds budget, may cause user frustration"
	} else {
		return "Load time slightly exceeds budget, minimal user impact"
	}
}

// generateSizeBreakdown creates a detailed size breakdown of the bundle
func (ba *BundleAnalyzer) generateSizeBreakdown(impacts []*PerformanceImpact) *SizeBreakdown {
	breakdown := &SizeBreakdown{
		ByCategory: make(map[string]int64),
	}

	for _, impact := range impacts {
		category := ba.categorizePackage(impact.PackageName)
		breakdown.ByCategory[category] += impact.EstimatedSize

		// Update main categories
		switch category {
		case "framework":
			breakdown.Framework += impact.EstimatedSize
		case "utilities":
			breakdown.Utilities += impact.EstimatedSize
		case "libraries":
			breakdown.Libraries += impact.EstimatedSize
		case "polyfills":
			breakdown.Polyfills += impact.EstimatedSize
		default:
			breakdown.Libraries += impact.EstimatedSize
		}
	}

	return breakdown
}

// categorizePackage determines the category of a package
func (ba *BundleAnalyzer) categorizePackage(packageName string) string {
	name := strings.ToLower(packageName)

	// Framework packages
	if strings.Contains(name, "react") || strings.Contains(name, "vue") || 
	   strings.Contains(name, "angular") || strings.Contains(name, "svelte") {
		return "framework"
	}

	// Utility libraries
	if strings.Contains(name, "lodash") || strings.Contains(name, "moment") ||
	   strings.Contains(name, "date-fns") || strings.Contains(name, "ramda") {
		return "utilities"
	}

	// Polyfills
	if strings.Contains(name, "polyfill") || strings.Contains(name, "core-js") ||
	   strings.Contains(name, "babel-runtime") {
		return "polyfills"
	}

	// Build tools and dev dependencies (shouldn't be in bundle)
	if strings.Contains(name, "webpack") || strings.Contains(name, "babel") ||
	   strings.Contains(name, "eslint") || strings.Contains(name, "typescript") {
		return "build-tools"
	}

	// Default to libraries
	return "libraries"
}

// generateBundleRecommendations generates optimization recommendations for the entire bundle
func (ba *BundleAnalyzer) generateBundleRecommendations(sizeAnalysis *SizeAnalysis, budgetAnalysis *BudgetAnalysis, treeShakingAnalysis *TreeShakingAnalysis) []PerformanceRecommendation {
	recommendations := make([]PerformanceRecommendation, 0)

	// Bundle size recommendations based on budget violations
	if budgetAnalysis.MaxSeverity == "critical" {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "bundle-splitting",
			Priority:    "critical",
			Description: fmt.Sprintf("Implement code splitting to reduce initial bundle size by %d KB", (sizeAnalysis.TotalSize-ba.budgets.TotalSize)/1024),
			ImpactScore: 90.0,
		})
	}

	// Tree shaking recommendations
	if treeShakingAnalysis != nil && treeShakingAnalysis.PotentialSavings > 20 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "tree-shaking",
			Priority:    "high", 
			Description: fmt.Sprintf("Enable tree shaking to achieve %.1f%% bundle size reduction", treeShakingAnalysis.PotentialSavings),
			ImpactScore: 80.0,
		})
	}

	// Budget violation recommendations
	for _, violation := range budgetAnalysis.Violations {
		if violation.Severity == "critical" || violation.Severity == "high" {
			recommendations = append(recommendations, PerformanceRecommendation{
				Type:        "budget-optimization",
				Priority:    violation.Severity,
				Description: fmt.Sprintf("Optimize %s budget - currently %.1f%% over limit", violation.BudgetType, violation.OveragePercentage),
				ImpactScore: 85.0,
			})
		}
	}

	// General optimization recommendations
	if len(budgetAnalysis.Violations) == 0 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Type:        "maintenance",
			Priority:    "low",
			Description: "Bundle is within performance budgets - continue monitoring",
			ImpactScore: 20.0,
		})
	}

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priorities := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priorities[recommendations[i].Priority] > priorities[recommendations[j].Priority]
	})

	return recommendations
}

// PerformTreeShakingAnalysis analyzes the tree-shaking potential of the bundle
func (ba *BundleAnalyzer) PerformTreeShakingAnalysis(ctx context.Context, packages []*GraphPackageInfo) (*TreeShakingAnalysis, error) {
	analysis := &TreeShakingAnalysis{
		TotalSize:         0,
		TreeShakableSize:  0,
		NonTreeShakableSize: 0,
		PackageAnalysis:   make([]PackageTreeShakingInfo, 0),
		Recommendations:   make([]string, 0),
	}

	for _, pkg := range packages {
		if pkg.DependencyType == "devDependencies" {
			continue
		}

		// Get package metrics
		metrics, err := ba.performanceAnalyzer.getPackageMetrics(ctx, pkg.Name, pkg.Version)
		if err != nil {
			// Use estimation
			metrics = ba.performanceAnalyzer.estimatePackageMetrics(pkg.Name, pkg.Version)
		}

		info := PackageTreeShakingInfo{
			PackageName:       pkg.Name,
			TotalSize:         metrics.RawSize,
			IsTreeShakable:    metrics.IsTreeShakable,
			HasSideEffects:    metrics.HasSideEffects,
			TreeShakableSize:  0,
			TreeShakingRatio:  0,
			Recommendations:   make([]string, 0),
		}

		if metrics.IsTreeShakable {
			info.TreeShakableSize = int64(float64(metrics.RawSize) * metrics.TreeShakingRatio)
			info.TreeShakingRatio = metrics.TreeShakingRatio
			analysis.TreeShakableSize += info.TreeShakableSize
		} else {
			analysis.NonTreeShakableSize += metrics.RawSize
			info.Recommendations = append(info.Recommendations, "Package is not tree-shakable due to side effects")
		}

		analysis.TotalSize += metrics.RawSize
		analysis.PackageAnalysis = append(analysis.PackageAnalysis, info)
	}

	// Calculate potential savings
	if analysis.TotalSize > 0 {
		analysis.PotentialSavings = float64(analysis.TreeShakableSize) / float64(analysis.TotalSize) * 100
	}

	// Generate overall recommendations
	if analysis.PotentialSavings > 20 {
		analysis.Recommendations = append(analysis.Recommendations, "Enable tree shaking to achieve significant bundle size reduction")
	}
	if analysis.NonTreeShakableSize > analysis.TreeShakableSize {
		analysis.Recommendations = append(analysis.Recommendations, "Consider replacing non-tree-shakable packages with alternatives")
	}

	return analysis, nil
}

// TreeShakingAnalysis contains tree-shaking analysis results
type TreeShakingAnalysis struct {
	TotalSize           int64                      `json:"total_size"`
	TreeShakableSize    int64                      `json:"tree_shakable_size"`
	NonTreeShakableSize int64                      `json:"non_tree_shakable_size"`
	PotentialSavings    float64                    `json:"potential_savings"` // percentage
	PackageAnalysis     []PackageTreeShakingInfo   `json:"package_analysis"`
	Recommendations     []string                   `json:"recommendations"`
}

// PackageTreeShakingInfo contains tree-shaking info for a specific package
type PackageTreeShakingInfo struct {
	PackageName      string   `json:"package_name"`
	TotalSize        int64    `json:"total_size"`
	IsTreeShakable   bool     `json:"is_tree_shakable"`
	HasSideEffects   bool     `json:"has_side_effects"`
	TreeShakableSize int64    `json:"tree_shakable_size"`
	TreeShakingRatio float64  `json:"tree_shaking_ratio"`
	Recommendations  []string `json:"recommendations"`
}

// SetBundlerConfig allows customizing the bundler configuration
func (ba *BundleAnalyzer) SetBundlerConfig(config *BundlerConfig) {
	ba.bundlerConfig = config
}

// SetPerformanceBudgets allows customizing performance budgets
func (ba *BundleAnalyzer) SetPerformanceBudgets(budgets *PerformanceBudgets) {
	ba.budgets = *budgets
}

// GetBundlerConfig returns the current bundler configuration
func (ba *BundleAnalyzer) GetBundlerConfig() *BundlerConfig {
	return ba.bundlerConfig
}

// GetPerformanceBudgets returns the current performance budgets
func (ba *BundleAnalyzer) GetPerformanceBudgets() *PerformanceBudgets {
	return &ba.budgets
}

// analyzeTreeShakingPotential analyzes tree-shaking potential from dependencies
func (ba *BundleAnalyzer) analyzeTreeShakingPotential(dependencies []Dependency) *TreeShakingAnalysis {
	analysis := &TreeShakingAnalysis{
		PackageAnalysis: make([]PackageTreeShakingInfo, 0),
		PotentialSavings: 0.0,
		TotalSize: 0,
		TreeShakableSize: 0,
		NonTreeShakableSize: 0,
		Recommendations: make([]string, 0),
	}

	totalSavings := 0.0

	for _, dep := range dependencies {
		if dep.Type == "devDependencies" {
			continue
		}

		// Estimate package characteristics
		isTreeShakable := ba.isPackageTreeShakable(dep.Name)
		estimatedSize := ba.estimatePackageSize(dep.Name)
		analysis.TotalSize += estimatedSize

		packageInfo := PackageTreeShakingInfo{
			PackageName:      dep.Name,
			TotalSize:        estimatedSize,
			IsTreeShakable:   isTreeShakable,
			HasSideEffects:   !isTreeShakable,
			TreeShakableSize: 0,
			TreeShakingRatio: 0,
			Recommendations:  ba.getTreeShakingTechniques(dep.Name),
		}

		if isTreeShakable {
			savingsPercent := ba.calculateTreeShakingPotential(dep.Name)
			packageInfo.TreeShakableSize = int64(float64(estimatedSize) * savingsPercent / 100)
			packageInfo.TreeShakingRatio = savingsPercent / 100
			analysis.TreeShakableSize += packageInfo.TreeShakableSize
			totalSavings += savingsPercent
		} else {
			analysis.NonTreeShakableSize += estimatedSize
		}

		analysis.PackageAnalysis = append(analysis.PackageAnalysis, packageInfo)
	}

	if analysis.TotalSize > 0 {
		analysis.PotentialSavings = float64(analysis.TreeShakableSize) / float64(analysis.TotalSize) * 100
	}

	// Generate recommendations based on analysis
	if analysis.PotentialSavings > 20 {
		analysis.Recommendations = append(analysis.Recommendations, "Enable tree shaking to achieve significant bundle size reduction")
	}
	if analysis.NonTreeShakableSize > analysis.TreeShakableSize {
		analysis.Recommendations = append(analysis.Recommendations, "Consider replacing non-tree-shakable packages with alternatives")
	}

	return analysis
}

// Helper methods for tree-shaking analysis
func (ba *BundleAnalyzer) isPackageTreeShakable(packageName string) bool {
	name := strings.ToLower(packageName)
	
	// Known tree-shakable packages
	treeShakablePackages := map[string]bool{
		"lodash":   true,
		"rxjs":     true,
		"date-fns": true,
		"ramda":    true,
	}

	// Known non-tree-shakable packages  
	nonTreeShakablePackages := map[string]bool{
		"moment": true,
		"jquery": true,
	}

	if isTreeShakable, exists := treeShakablePackages[name]; exists {
		return isTreeShakable
	}
	if isTreeShakable, exists := nonTreeShakablePackages[name]; exists {
		return !isTreeShakable
	}

	// Default heuristics
	return !strings.Contains(name, "polyfill") && !strings.Contains(name, "shim")
}

func (ba *BundleAnalyzer) estimatePackageSize(packageName string) int64 {
	name := strings.ToLower(packageName)
	
	// Size estimates for common packages
	if strings.Contains(name, "react") || strings.Contains(name, "vue") {
		return 150000 // 150KB
	} else if strings.Contains(name, "lodash") || strings.Contains(name, "moment") {
		return 200000 // 200KB
	} else if strings.HasPrefix(name, "@types/") {
		return 5000 // 5KB
	}
	
	return 50000 // 50KB default
}

func (ba *BundleAnalyzer) calculateTreeShakingPotential(packageName string) float64 {
	name := strings.ToLower(packageName)
	
	// Tree-shaking potential estimates
	if strings.Contains(name, "lodash") {
		return 40.0 // Can save ~40% with selective imports
	} else if strings.Contains(name, "rxjs") {
		return 30.0 // Can save ~30% with operator imports
	} else if strings.Contains(name, "date-fns") {
		return 35.0 // Can save ~35% with selective imports
	}
	
	return 15.0 // Default 15% savings
}

func (ba *BundleAnalyzer) getTreeShakingTechniques(packageName string) []string {
	name := strings.ToLower(packageName)
	
	if strings.Contains(name, "lodash") {
		return []string{
			"Import specific functions",
			"Use lodash-es",
			"Use babel-plugin-lodash",
		}
	} else if strings.Contains(name, "rxjs") {
		return []string{
			"Import operators individually",
			"Use custom builds",
		}
	} else if strings.Contains(name, "date-fns") {
		return []string{
			"Import specific functions",
			"Use babel-plugin-date-fns",
		}
	} else if strings.Contains(name, "moment") {
		return []string{} // Not tree-shakable
	}
	
	return []string{"Use ES6 imports", "Enable tree-shaking in bundler"}
}

// calculateSizeBreakdown calculates size breakdown for bundle
func (ba *BundleAnalyzer) calculateSizeBreakdown(ctx context.Context, dependencies []Dependency) *SizeAnalysis {
	totalSize := int64(0)
	productionSize := int64(0)
	byCategory := make(map[string]int64)

	for _, dep := range dependencies {
		estimatedSize := ba.estimatePackageSize(dep.Name)
		totalSize += estimatedSize
		
		// Categorize package
		category := ba.categorizePackage(dep.Name)
		byCategory[category] += estimatedSize
		
		if dep.Type != "devDependencies" {
			productionSize += estimatedSize
		}
	}

	return &SizeAnalysis{
		TotalSize:      totalSize,
		ProductionSize: productionSize,
		InitialBundle:  productionSize, // Assume all production deps in initial bundle
		ByCategory:     byCategory,
	}
}

// Additional helper methods for budget analysis
func (ba *BundleAnalyzer) calculateViolationSeverityFromSize(actual, budget int64) string {
	if actual <= budget {
		return "none"
	}

	overagePercent := float64(actual-budget) / float64(budget) * 100
	
	if overagePercent > 150 {
		return "critical"
	} else if overagePercent > 50 {
		return "high" 
	} else if overagePercent > 10 {
		return "medium"
	}
	
	return "low"
}


func (ba *BundleAnalyzer) isHigherSeverity(newSeverity, currentMax string) bool {
	severityLevels := map[string]int{
		"none":     0,
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}
	
	return severityLevels[newSeverity] > severityLevels[currentMax]
}

func (ba *BundleAnalyzer) describeSizeImpact(actual, budget int64) string {
	if actual <= budget {
		return "Within budget"
	}
	
	overBy := actual - budget
	percentOver := float64(overBy) / float64(budget) * 100

	if percentOver > 100 {
		return "Size is more than double the recommended budget, causing significant performance impact"
	} else if percentOver > 50 {
		return "Size significantly exceeds budget, noticeably impacting performance"
	} else if percentOver > 20 {
		return "Size moderately exceeds budget, may affect performance on slower connections"
	}
	
	return "Size slightly exceeds budget, minimal performance impact"
}

