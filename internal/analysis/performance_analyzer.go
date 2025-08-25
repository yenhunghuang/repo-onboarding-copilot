// Package analysis provides performance impact assessment for dependencies
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MemoryCache provides simple in-memory caching with TTL
type MemoryCache struct {
	data map[string]cacheEntry
	mu   sync.RWMutex
}

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]cacheEntry),
	}
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	
	return entry.value, true
}

// Set stores a value in the cache with TTL
func (c *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.data[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

// PerformanceAnalyzer analyzes performance impact of dependencies
type PerformanceAnalyzer struct {
	client        *http.Client
	cache         *MemoryCache
	bundlerConfig *BundlerConfig
	budgets       *PerformanceBudgets
}

// BundlerConfig contains configuration for different bundlers
type BundlerConfig struct {
	Type                string   `json:"type"`                 // webpack, rollup, esbuild, vite
	TreeShakingEnabled  bool     `json:"tree_shaking_enabled"` 
	CompressionEnabled  bool     `json:"compression_enabled"`  // gzip/brotli
	MinificationEnabled bool     `json:"minification_enabled"`
	CodeSplitting       bool     `json:"code_splitting"`
	TreeShakingRatio    float64  `json:"tree_shaking_ratio"`   // 0.0 to 1.0, how much can be removed
	CompressionRatio    float64  `json:"compression_ratio"`    // typical compression ratio
	OutputFormats       []string `json:"output_formats"`       // supported output formats
}

// PerformanceBudgets defines performance thresholds
type PerformanceBudgets struct {
	MaxBundleSize        int64   `json:"max_bundle_size"`         // bytes
	MaxInitialLoadTime   float64 `json:"max_initial_load_time"`   // milliseconds
	MaxScriptEvalTime    float64 `json:"max_script_eval_time"`    // milliseconds
	MaxFirstContentfulPaint float64 `json:"max_first_contentful_paint"` // milliseconds
	MaxLargestContentfulPaint float64 `json:"max_largest_contentful_paint"` // milliseconds
	MaxCumulativeLayoutShift float64 `json:"max_cumulative_layout_shift"` // CLS score
	MaxFirstInputDelay   float64 `json:"max_first_input_delay"`   // milliseconds
	// Additional fields for compatibility with tests
	TotalSize            int64   `json:"total_size"`              // alias for max_bundle_size
	InitialSize          int64   `json:"initial_size"`            // maximum initial bundle size
	AssetSize            int64   `json:"asset_size"`              // maximum individual asset size
}

// PerformanceImpact represents the performance analysis results
type PerformanceImpact struct {
	PackageName        string                 `json:"package_name"`
	EstimatedSize      int64                  `json:"estimated_size"`      // raw size in bytes
	MinifiedSize       int64                  `json:"minified_size"`       // after minification
	CompressedSize     int64                  `json:"compressed_size"`     // after gzip/brotli
	TreeShakableSize   int64                  `json:"tree_shakable_size"`  // removable with tree shaking
	LoadTimeImpact     *LoadTimeAnalysis      `json:"load_time_impact"`
	BundleContribution float64                `json:"bundle_contribution"` // percentage of total bundle
	PerformanceScore   float64                `json:"performance_score"`   // 0-100 score
	Recommendations    []string               `json:"recommendations"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// LoadTimeAnalysis contains load time calculations for different scenarios
type LoadTimeAnalysis struct {
	Network3G     *NetworkImpact `json:"network_3g"`     // 3G connection
	Network4G     *NetworkImpact `json:"network_4g"`     // 4G connection
	NetworkWiFi   *NetworkImpact `json:"network_wifi"`   // WiFi connection
	NetworkCable  *NetworkImpact `json:"network_cable"`  // Wired broadband
	DeviceLowEnd  *DeviceImpact  `json:"device_low_end"` // Low-end device
	DeviceMidEnd  *DeviceImpact  `json:"device_mid_end"` // Mid-range device
	DeviceHighEnd *DeviceImpact  `json:"device_high_end"` // High-end device
}

// NetworkImpact represents network-specific performance metrics
type NetworkImpact struct {
	DownloadTime     float64 `json:"download_time"`     // milliseconds
	ParseTime        float64 `json:"parse_time"`        // milliseconds
	ExecutionTime    float64 `json:"execution_time"`    // milliseconds
	TotalTime        float64 `json:"total_time"`        // milliseconds
	Bandwidth        int64   `json:"bandwidth"`         // bytes per second
	Latency          float64 `json:"latency"`           // milliseconds
	PacketLoss       float64 `json:"packet_loss"`       // percentage
}

// DeviceImpact represents device-specific performance metrics
type DeviceImpact struct {
	ParseTime       float64 `json:"parse_time"`       // milliseconds
	CompileTime     float64 `json:"compile_time"`     // milliseconds
	ExecutionTime   float64 `json:"execution_time"`   // milliseconds
	MemoryUsage     int64   `json:"memory_usage"`     // bytes
	CPUUtilization  float64 `json:"cpu_utilization"`  // percentage
	TotalTime       float64 `json:"total_time"`       // milliseconds
	DeviceType      string  `json:"device_type"`      // low-end, mid-end, high-end
}

// BundleAnalysisResult contains overall bundle performance analysis
type BundleAnalysisResult struct {
	TotalSize             int64                         `json:"total_size"`
	MinifiedSize          int64                         `json:"minified_size"`
	CompressedSize        int64                         `json:"compressed_size"`
	TreeShakableSize      int64                         `json:"tree_shakable_size"`
	OptimizedSize         int64                         `json:"optimized_size"`      // after all optimizations
	SizeAnalysis          *SizeAnalysis                 `json:"size_analysis"`
	TreeShakingAnalysis   *TreeShakingAnalysis          `json:"tree_shaking_analysis"`
	PackageContributions  []PackageContribution         `json:"package_contributions"`
	BudgetAnalysis        *BudgetAnalysis               `json:"budget_analysis"`
	LoadTimeAnalysis      *LoadTimeAnalysis             `json:"load_time_analysis"`
	Recommendations       []PerformanceRecommendation   `json:"recommendations"`
	SizeBreakdown         *SizeBreakdown                `json:"size_breakdown"`
	GeneratedAt           time.Time                     `json:"generated_at"`
}

// SizeAnalysis groups size-related metrics for bundle analysis
type SizeAnalysis struct {
	TotalSize      int64            `json:"total_size"`      // total bundle size
	ProductionSize int64            `json:"production_size"` // production-only dependencies
	InitialBundle  int64            `json:"initial_bundle"`  // initial load bundle size
	ByCategory     map[string]int64 `json:"by_category"`     // breakdown by package category
}

// PackageContribution shows how much each package contributes to bundle size
type PackageContribution struct {
	PackageName    string  `json:"package_name"`
	Size           int64   `json:"size"`
	Percentage     float64 `json:"percentage"`
	IsTreeShakable bool    `json:"is_tree_shakable"`
	Impact         string  `json:"impact"` // high, medium, low
}

// BudgetAnalysis compares actual metrics against performance budgets
type BudgetAnalysis struct {
	BundleSizeStatus    string  `json:"bundle_size_status"`    // pass, warn, fail
	LoadTimeStatus      string  `json:"load_time_status"`      // pass, warn, fail
	OverBudgetBy        int64   `json:"over_budget_by"`        // bytes over budget
	BudgetUtilization   float64 `json:"budget_utilization"`    // percentage of budget used
	MaxSeverity         string  `json:"max_severity"`          // highest severity of violations
	Violations          []BudgetViolation `json:"violations"`
}

// BudgetViolation represents a performance budget violation
type BudgetViolation struct {
	Metric           string  `json:"metric"`
	BudgetType       string  `json:"budget_type"`       // type of budget violated
	Actual           float64 `json:"actual"`
	ActualSize       int64   `json:"actual_size"`       // actual size for size violations
	Budget           float64 `json:"budget"`
	BudgetSize       int64   `json:"budget_size"`       // budget size for size violations
	OveragePercentage float64 `json:"overage_percentage"` // percentage over budget
	Severity         string  `json:"severity"`          // critical, high, medium, low
	Impact           string  `json:"impact"`            // description of user impact
}

// Remove duplicate - using the one from dependency_stubs.go

// SizeBreakdown shows bundle composition
type SizeBreakdown struct {
	Libraries      int64            `json:"libraries"`       // third-party packages
	ApplicationCode int64           `json:"application_code"` // user code
	Polyfills      int64            `json:"polyfills"`       // browser compatibility
	Framework      int64            `json:"framework"`       // React, Vue, Angular, etc.
	Utilities      int64            `json:"utilities"`       // lodash, moment, etc.
	Assets         int64            `json:"assets"`          // images, fonts, etc.
	ByCategory     map[string]int64 `json:"by_category"`     // detailed breakdown
}

// PackageMetrics contains size and performance metrics for a package
type PackageMetrics struct {
	RawSize          int64             `json:"raw_size"`
	MinifiedSize     int64             `json:"minified_size"`
	GzippedSize      int64             `json:"gzipped_size"`
	BrotliSize       int64             `json:"brotli_size"`
	TreeShakingRatio float64           `json:"tree_shaking_ratio"`
	ParseTime        float64           `json:"parse_time"`
	ExecutionTime    float64           `json:"execution_time"`
	MemoryUsage      int64             `json:"memory_usage"`
	Dependencies     []string          `json:"dependencies"`
	IsTreeShakable   bool              `json:"is_tree_shakable"`
	HasSideEffects   bool              `json:"has_side_effects"`
	Metadata         map[string]string `json:"metadata"`
	LastUpdated      time.Time         `json:"last_updated"`
}

// NetworkProfile defines network characteristics
type NetworkProfile struct {
	Name       string  `json:"name"`
	Bandwidth  int64   `json:"bandwidth"`   // bytes per second
	Latency    float64 `json:"latency"`     // milliseconds
	PacketLoss float64 `json:"packet_loss"` // percentage (0-1)
}

// DeviceProfile defines device performance characteristics
type DeviceProfile struct {
	Name           string  `json:"name"`
	CPUMultiplier  float64 `json:"cpu_multiplier"`  // relative to baseline
	MemoryLimit    int64   `json:"memory_limit"`    // bytes
	ParseMultiplier float64 `json:"parse_multiplier"` // JS parse time multiplier
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: NewMemoryCache(), // 1 hour cache
		bundlerConfig: &BundlerConfig{
			Type:                "webpack", // default
			TreeShakingEnabled:  true,
			CompressionEnabled:  true,
			MinificationEnabled: true,
			CodeSplitting:       true,
			TreeShakingRatio:    0.3, // typical 30% reduction
			CompressionRatio:    0.7, // typical 30% reduction with gzip
		},
		budgets: getDefaultPerformanceBudgets(),
	}
}

// getDefaultPerformanceBudgets returns sensible default performance budgets
func getDefaultPerformanceBudgets() *PerformanceBudgets {
	return &PerformanceBudgets{
		MaxBundleSize:             500 * 1024,    // 500KB
		MaxInitialLoadTime:        3000,          // 3 seconds
		MaxScriptEvalTime:         1000,          // 1 second
		MaxFirstContentfulPaint:   1500,          // 1.5 seconds
		MaxLargestContentfulPaint: 2500,          // 2.5 seconds
		MaxCumulativeLayoutShift:  0.1,           // CLS score
		MaxFirstInputDelay:        100,           // 100ms
	}
}

// AnalyzePackagePerformance analyzes performance impact of a single package
func (pa *PerformanceAnalyzer) AnalyzePackagePerformance(ctx context.Context, pkg *GraphPackageInfo) (*PerformanceImpact, error) {
	// Get package metrics from cache or fetch
	metrics, err := pa.getPackageMetrics(ctx, pkg.Name, pkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get package metrics: %w", err)
	}

	// Calculate size estimates
	estimatedSize := pa.calculatePackageSize(metrics, pkg)
	minifiedSize := int64(float64(estimatedSize) * 0.7) // typical 30% reduction
	compressedSize := int64(float64(minifiedSize) * pa.bundlerConfig.CompressionRatio)
	
	var treeShakableSize int64
	if metrics.IsTreeShakable && pa.bundlerConfig.TreeShakingEnabled {
		treeShakableSize = int64(float64(estimatedSize) * pa.bundlerConfig.TreeShakingRatio)
	}

	// Calculate load time impact
	loadTimeAnalysis := pa.calculateLoadTimeImpact(estimatedSize, minifiedSize, compressedSize)

	// Calculate performance score (0-100)
	performanceScore := pa.calculatePerformanceScore(estimatedSize, metrics)

	// Generate recommendations
	recommendations := pa.generatePackageRecommendations(pkg, metrics, estimatedSize)

	return &PerformanceImpact{
		PackageName:        pkg.Name,
		EstimatedSize:      estimatedSize,
		MinifiedSize:       minifiedSize,
		CompressedSize:     compressedSize,
		TreeShakableSize:   treeShakableSize,
		LoadTimeImpact:     loadTimeAnalysis,
		BundleContribution: 0, // Will be calculated when analyzing full bundle
		PerformanceScore:   performanceScore,
		Recommendations:    recommendations,
		Metadata: map[string]interface{}{
			"version":         pkg.Version,
			"dependency_type": pkg.DependencyType,
			"has_side_effects": metrics.HasSideEffects,
			"is_tree_shakable": metrics.IsTreeShakable,
		},
	}, nil
}

// getPackageMetrics retrieves or estimates package metrics
func (pa *PerformanceAnalyzer) getPackageMetrics(ctx context.Context, name, version string) (*PackageMetrics, error) {
	cacheKey := fmt.Sprintf("metrics:%s@%s", name, version)
	
	// Check cache first
	if cached, exists := pa.cache.Get(cacheKey); exists {
		if metrics, ok := cached.(*PackageMetrics); ok {
			return metrics, nil
		}
	}

	// Fetch from npm registry or use estimation
	metrics, err := pa.fetchPackageMetrics(ctx, name, version)
	if err != nil {
		// Fall back to estimation
		metrics = pa.estimatePackageMetrics(name, version)
	}

	// Cache the result
	pa.cache.Set(cacheKey, metrics, time.Hour)
	
	return metrics, nil
}

// fetchPackageMetrics fetches actual metrics from npm registry
func (pa *PerformanceAnalyzer) fetchPackageMetrics(ctx context.Context, name, version string) (*PackageMetrics, error) {
	url := fmt.Sprintf("https://registry.npmjs.org/%s/%s", name, version)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pa.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var npmData map[string]interface{}
	if err := json.Unmarshal(body, &npmData); err != nil {
		return nil, err
	}

	return pa.parseNpmMetrics(npmData)
}

// parseNpmMetrics extracts metrics from npm registry response
func (pa *PerformanceAnalyzer) parseNpmMetrics(data map[string]interface{}) (*PackageMetrics, error) {
	metrics := &PackageMetrics{
		LastUpdated: time.Now(),
		Metadata:    make(map[string]string),
	}

	// Extract size information from dist
	if dist, ok := data["dist"].(map[string]interface{}); ok {
		if unpackedSize, ok := dist["unpackedSize"].(float64); ok {
			metrics.RawSize = int64(unpackedSize)
		}
	}

	// Estimate other sizes
	metrics.MinifiedSize = int64(float64(metrics.RawSize) * 0.7)
	metrics.GzippedSize = int64(float64(metrics.MinifiedSize) * 0.7)
	metrics.BrotliSize = int64(float64(metrics.MinifiedSize) * 0.65)

	// Check for side effects (affects tree shaking)
	if sideEffects, ok := data["sideEffects"].(bool); ok {
		metrics.HasSideEffects = sideEffects
	}
	metrics.IsTreeShakable = !metrics.HasSideEffects

	// Estimate performance metrics
	metrics.TreeShakingRatio = 0.3 // default 30%
	metrics.ParseTime = float64(metrics.RawSize) / 1000000 * 10 // ~10ms per MB
	metrics.ExecutionTime = metrics.ParseTime * 2 // execution typically 2x parse time
	metrics.MemoryUsage = metrics.RawSize * 3 // rough estimate

	return metrics, nil
}

// estimatePackageMetrics provides fallback size estimation
func (pa *PerformanceAnalyzer) estimatePackageMetrics(name, version string) *PackageMetrics {
	// Basic size estimation based on package name patterns
	baseSize := int64(50000) // 50KB default
	
	// Adjust based on common package patterns
	if strings.Contains(name, "react") || strings.Contains(name, "vue") || strings.Contains(name, "angular") {
		baseSize = 150000 // Frameworks are larger
	} else if strings.Contains(name, "lodash") || strings.Contains(name, "moment") {
		baseSize = 200000 // Utility libraries
	} else if strings.Contains(name, "babel") || strings.Contains(name, "webpack") {
		baseSize = 100000 // Build tools
	} else if strings.Contains(name, "types") || strings.HasPrefix(name, "@types/") {
		baseSize = 10000 // Type definitions are small
	}

	return &PackageMetrics{
		RawSize:          baseSize,
		MinifiedSize:     int64(float64(baseSize) * 0.7),
		GzippedSize:      int64(float64(baseSize) * 0.5),
		BrotliSize:       int64(float64(baseSize) * 0.45),
		TreeShakingRatio: 0.3,
		ParseTime:        float64(baseSize) / 1000000 * 10,
		ExecutionTime:    float64(baseSize) / 1000000 * 20,
		MemoryUsage:      baseSize * 3,
		IsTreeShakable:   true,
		HasSideEffects:   false,
		Dependencies:     []string{},
		Metadata:         make(map[string]string),
		LastUpdated:      time.Now(),
	}
}

// calculatePackageSize calculates the effective size of a package
func (pa *PerformanceAnalyzer) calculatePackageSize(metrics *PackageMetrics, pkg *GraphPackageInfo) int64 {
	size := metrics.RawSize
	
	// Adjust for development dependencies (not included in production builds)
	if pkg.DependencyType == "devDependencies" {
		return 0 // Dev dependencies don't contribute to bundle size
	}
	
	return size
}

// calculateLoadTimeImpact calculates load time impact across different scenarios
func (pa *PerformanceAnalyzer) calculateLoadTimeImpact(rawSize, minifiedSize, compressedSize int64) *LoadTimeAnalysis {
	// Define network profiles
	profiles := map[string]NetworkProfile{
		"3g":    {Name: "3G", Bandwidth: 400 * 1024, Latency: 400, PacketLoss: 0.02},      // 400KB/s, 400ms latency
		"4g":    {Name: "4G", Bandwidth: 1500 * 1024, Latency: 150, PacketLoss: 0.005},   // 1.5MB/s, 150ms latency
		"wifi":  {Name: "WiFi", Bandwidth: 5000 * 1024, Latency: 50, PacketLoss: 0.001},  // 5MB/s, 50ms latency
		"cable": {Name: "Cable", Bandwidth: 10000 * 1024, Latency: 20, PacketLoss: 0.0005}, // 10MB/s, 20ms latency
	}

	// Define device profiles
	devices := map[string]DeviceProfile{
		"low_end":  {Name: "Low-end", CPUMultiplier: 3.0, MemoryLimit: 512 * 1024 * 1024, ParseMultiplier: 4.0},
		"mid_end":  {Name: "Mid-range", CPUMultiplier: 1.5, MemoryLimit: 2048 * 1024 * 1024, ParseMultiplier: 2.0},
		"high_end": {Name: "High-end", CPUMultiplier: 1.0, MemoryLimit: 8192 * 1024 * 1024, ParseMultiplier: 1.0},
	}

	analysis := &LoadTimeAnalysis{}

	// Calculate network impacts
	analysis.Network3G = pa.calculateNetworkImpact(compressedSize, profiles["3g"])
	analysis.Network4G = pa.calculateNetworkImpact(compressedSize, profiles["4g"])
	analysis.NetworkWiFi = pa.calculateNetworkImpact(compressedSize, profiles["wifi"])
	analysis.NetworkCable = pa.calculateNetworkImpact(compressedSize, profiles["cable"])

	// Calculate device impacts (using minified size for parse/execute)
	analysis.DeviceLowEnd = pa.calculateDeviceImpact(minifiedSize, devices["low_end"])
	analysis.DeviceMidEnd = pa.calculateDeviceImpact(minifiedSize, devices["mid_end"])
	analysis.DeviceHighEnd = pa.calculateDeviceImpact(minifiedSize, devices["high_end"])

	return analysis
}

// calculateNetworkImpact calculates network-specific load time metrics
func (pa *PerformanceAnalyzer) calculateNetworkImpact(size int64, profile NetworkProfile) *NetworkImpact {
	// Download time = size / bandwidth + latency overhead
	downloadTime := float64(size)/float64(profile.Bandwidth)*1000 + profile.Latency
	
	// Adjust for packet loss (retransmissions)
	downloadTime *= (1 + profile.PacketLoss*5) // rough packet loss impact
	
	// Parse and execution are network-independent for this calculation
	parseTime := float64(size) / 1000000 * 10  // ~10ms per MB
	executionTime := parseTime * 2             // execution typically 2x parse
	
	totalTime := downloadTime + parseTime + executionTime

	return &NetworkImpact{
		DownloadTime:  downloadTime,
		ParseTime:     parseTime,
		ExecutionTime: executionTime,
		TotalTime:     totalTime,
		Bandwidth:     profile.Bandwidth,
		Latency:       profile.Latency,
		PacketLoss:    profile.PacketLoss,
	}
}

// calculateDeviceImpact calculates device-specific performance metrics
func (pa *PerformanceAnalyzer) calculateDeviceImpact(size int64, profile DeviceProfile) *DeviceImpact {
	// Base parse time scaled by device capability
	baseParseTime := float64(size) / 1000000 * 10 // ~10ms per MB
	parseTime := baseParseTime * profile.ParseMultiplier
	
	// Compile time (JIT compilation)
	compileTime := parseTime * 0.5 * profile.CPUMultiplier
	
	// Execution time
	executionTime := parseTime * 2 * profile.CPUMultiplier
	
	// Memory usage
	memoryUsage := size * 3 // rough estimate: 3x file size in memory
	
	// CPU utilization estimate
	cpuUtilization := math.Min(100, 20*profile.CPUMultiplier) // base 20% * multiplier
	
	totalTime := parseTime + compileTime + executionTime

	return &DeviceImpact{
		ParseTime:      parseTime,
		CompileTime:    compileTime,
		ExecutionTime:  executionTime,
		MemoryUsage:    memoryUsage,
		CPUUtilization: cpuUtilization,
		TotalTime:      totalTime,
		DeviceType:     profile.Name,
	}
}

// calculatePerformanceScore calculates a 0-100 performance score for a package
func (pa *PerformanceAnalyzer) calculatePerformanceScore(size int64, metrics *PackageMetrics) float64 {
	score := 100.0

	// Size penalty (larger packages score lower)
	sizeMB := float64(size) / (1024 * 1024)
	sizePenalty := math.Min(50, sizeMB*10) // Up to 50 points penalty for size
	score -= sizePenalty

	// Tree shaking bonus
	if metrics.IsTreeShakable {
		score += 10
	}

	// Side effects penalty
	if metrics.HasSideEffects {
		score -= 15
	}

	// Parse time penalty
	if metrics.ParseTime > 100 {
		score -= math.Min(20, (metrics.ParseTime-100)/10)
	}

	// Ensure score is within bounds
	return math.Max(0, math.Min(100, score))
}

// generatePackageRecommendations generates optimization recommendations for a package
func (pa *PerformanceAnalyzer) generatePackageRecommendations(pkg *GraphPackageInfo, metrics *PackageMetrics, size int64) []string {
	var recommendations []string

	// Size-based recommendations
	if size > 100*1024 { // > 100KB
		recommendations = append(recommendations, "Consider lazy loading this large package")
		if metrics.IsTreeShakable {
			recommendations = append(recommendations, "Enable tree shaking to reduce bundle size")
		}
	}

	// Tree shaking recommendations
	if !metrics.IsTreeShakable && size > 50*1024 {
		recommendations = append(recommendations, "Package is not tree-shakable - consider alternatives")
	}

	// Side effects recommendations
	if metrics.HasSideEffects {
		recommendations = append(recommendations, "Package has side effects - may prevent optimizations")
	}

	// Dependency type recommendations
	if pkg.DependencyType == "devDependencies" {
		recommendations = append(recommendations, "Dev dependency - should not impact production bundle")
	}

	// Alternative suggestions
	if strings.Contains(pkg.Name, "moment") {
		recommendations = append(recommendations, "Consider date-fns or dayjs as lighter alternatives to moment.js")
	} else if strings.Contains(pkg.Name, "lodash") && !strings.Contains(pkg.Name, "lodash.") {
		recommendations = append(recommendations, "Consider importing specific lodash functions instead of the entire library")
	}

	return recommendations
}

// GetNetworkProfiles returns available network profiles for testing
func GetNetworkProfiles() map[string]NetworkProfile {
	return map[string]NetworkProfile{
		"slow_3g":    {Name: "Slow 3G", Bandwidth: 50 * 1024, Latency: 2000, PacketLoss: 0.05},
		"3g":         {Name: "3G", Bandwidth: 400 * 1024, Latency: 400, PacketLoss: 0.02},
		"4g":         {Name: "4G", Bandwidth: 1500 * 1024, Latency: 150, PacketLoss: 0.005},
		"wifi":       {Name: "WiFi", Bandwidth: 5000 * 1024, Latency: 50, PacketLoss: 0.001},
		"cable":      {Name: "Cable", Bandwidth: 10000 * 1024, Latency: 20, PacketLoss: 0.0005},
		"fiber":      {Name: "Fiber", Bandwidth: 50000 * 1024, Latency: 5, PacketLoss: 0.0001},
	}
}

// GetDeviceProfiles returns available device profiles for testing
func GetDeviceProfiles() map[string]DeviceProfile {
	return map[string]DeviceProfile{
		"low_end": {
			Name:            "Low-end Mobile",
			CPUMultiplier:   4.0,
			MemoryLimit:     512 * 1024 * 1024, // 512MB
			ParseMultiplier: 6.0,
		},
		"mid_end": {
			Name:            "Mid-range Mobile",
			CPUMultiplier:   2.0,
			MemoryLimit:     2048 * 1024 * 1024, // 2GB
			ParseMultiplier: 3.0,
		},
		"high_end_mobile": {
			Name:            "High-end Mobile",
			CPUMultiplier:   1.5,
			MemoryLimit:     4096 * 1024 * 1024, // 4GB
			ParseMultiplier: 1.5,
		},
		"desktop": {
			Name:            "Desktop",
			CPUMultiplier:   1.0,
			MemoryLimit:     8192 * 1024 * 1024, // 8GB
			ParseMultiplier: 1.0,
		},
		"high_end_desktop": {
			Name:            "High-end Desktop",
			CPUMultiplier:   0.7,
			MemoryLimit:     16384 * 1024 * 1024, // 16GB
			ParseMultiplier: 0.8,
		},
	}
}