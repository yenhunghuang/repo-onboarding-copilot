package analysis

import (
	"time"
)

// Stub implementations for components that will be implemented in later tasks
// These allow the basic dependency parsing functionality to compile and be tested

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`    // low, medium, high, critical
	CVSS        float64           `json:"cvss"`        // CVSS score 0-10
	CWE         []string          `json:"cwe"`         // Common Weakness Enumeration IDs
	References  []string          `json:"references"`  // URLs to vulnerability details
	PatchedIn   []string          `json:"patched_in"`  // Versions that fix the vulnerability
	AffectedIn  []string          `json:"affected_in"` // Version ranges affected
	Metadata    map[string]string `json:"metadata"`
}

// LicenseInfo represents license information
type LicenseInfo struct {
	SPDX         string            `json:"spdx"`         // SPDX license identifier
	Name         string            `json:"name"`         // License name
	Type         string            `json:"type"`         // permissive, copyleft, proprietary
	URL          string            `json:"url"`          // License text URL
	Compatibility string           `json:"compatibility"` // compatible, incompatible, unknown
	Risk         string            `json:"risk"`         // low, medium, high
	Metadata     map[string]string `json:"metadata"`
}

// UpdateInfo represents available package updates
type UpdateInfo struct {
	Current        string    `json:"current"`
	Latest         string    `json:"latest"`
	Wanted         string    `json:"wanted"`      // Highest version satisfying semver
	Type           string    `json:"type"`        // major, minor, patch
	Breaking       bool      `json:"breaking"`    // Indicates breaking changes
	Security       bool      `json:"security"`    // Security update
	Deprecated     bool      `json:"deprecated"`  // Current version deprecated
	ReleaseDate    time.Time `json:"release_date"`
	ChangelogURL   string    `json:"changelog_url"`
	UpdatePriority string    `json:"update_priority"` // low, medium, high, critical
}

// Note: DependencyGraph, GraphNode, GraphEdge, and GraphStats are now defined in dependency_graph.go
// with advanced analysis capabilities including circular dependency detection, impact analysis,
// and D3.js visualization export functionality.

// BundleAnalysis represents bundle size analysis
type BundleAnalysis struct {
	EstimatedSize    int64               `json:"estimated_size"`    // total estimated bundle size
	CompressedSize   int64               `json:"compressed_size"`   // gzipped size estimate
	TreeShakable     int64               `json:"tree_shakable"`     // size that can be tree-shaken
	SizeByType       map[string]int64    `json:"size_by_type"`      // size breakdown by dependency type
	LargestPackages  []PackageSize       `json:"largest_packages"`  // packages contributing most to bundle
	Recommendations  []string            `json:"recommendations"`   // optimization suggestions
	PerformanceScore float64             `json:"performance_score"` // 0-100 performance score
	LoadTimeEstimate map[string]float64  `json:"load_time_estimate"` // estimated load times by connection
}

// PackageSize represents a package's contribution to bundle size
type PackageSize struct {
	Name           string  `json:"name"`
	Size           int64   `json:"size"`
	CompressedSize int64   `json:"compressed_size"`
	Percentage     float64 `json:"percentage"` // percentage of total bundle
	TreeShakable   bool    `json:"tree_shakable"`
}

// SecurityReport represents vulnerability analysis results
type SecurityReport struct {
	TotalVulnerabilities int               `json:"total_vulnerabilities"`
	CriticalCount        int               `json:"critical_count"`
	HighCount            int               `json:"high_count"`
	MediumCount          int               `json:"medium_count"`
	LowCount             int               `json:"low_count"`
	VulnerablePackages   []string          `json:"vulnerable_packages"`
	Vulnerabilities      []Vulnerability   `json:"vulnerabilities"`
	SeverityDistribution map[string]int    `json:"severity_distribution"`
	Recommendations      []string          `json:"recommendations"`
	RiskScore            float64           `json:"risk_score"` // 0-100 overall risk score
}

// LicenseReport represents license analysis results
type LicenseReport struct {
	TotalPackages        int               `json:"total_packages"`
	LicenseDistribution  map[string]int    `json:"license_distribution"`
	CompatibilityIssues  []LicenseConflict `json:"compatibility_issues"`
	UnknownLicenses      []string          `json:"unknown_licenses"`
	ProprietaryPackages  []string          `json:"proprietary_packages"`
	CopyleftPackages     []string          `json:"copyleft_packages"`
	RiskAssessment       string            `json:"risk_assessment"` // low, medium, high
	Recommendations      []string          `json:"recommendations"`
}

// LicenseConflict represents a license compatibility issue
type LicenseConflict struct {
	Package1    string `json:"package1"`
	Package2    string `json:"package2"`
	License1    string `json:"license1"`
	License2    string `json:"license2"`
	ConflictType string `json:"conflict_type"` // incompatible, restrictive, unknown
	Severity    string `json:"severity"`      // low, medium, high
}

// UpdateReport represents available updates analysis
type UpdateReport struct {
	TotalPackages      int               `json:"total_packages"`
	OutdatedPackages   int               `json:"outdated_packages"`
	SecurityUpdates    int               `json:"security_updates"`
	BreakingUpdates    int               `json:"breaking_updates"`
	Updates            []UpdateInfo      `json:"updates"`
	UpdatesByType      map[string]int    `json:"updates_by_type"` // patch, minor, major
	UpdatesByPriority  map[string]int    `json:"updates_by_priority"` // low, medium, high, critical
	Recommendations    []string          `json:"recommendations"`
	UpdateStrategy     string            `json:"update_strategy"` // conservative, moderate, aggressive
}

// PerformanceReport represents performance analysis results
type PerformanceReport struct {
	Packages          []PerformanceImpact          `json:"packages"`
	AverageLoadTime   map[string]float64           `json:"average_load_time"`   // by network type
	TotalImpact       float64                      `json:"total_impact"`        // overall performance score
	Recommendations   []PerformanceRecommendation  `json:"recommendations"`
}

// PerformanceRecommendation represents a performance optimization recommendation
type PerformanceRecommendation struct {
	Type        string  `json:"type"`        // load-time, bundle-size, package-alternatives
	Description string  `json:"description"`
	Priority    string  `json:"priority"`    // high, medium, low
	ImpactScore float64 `json:"impact_score"` // 0-100 potential improvement
}

// Dependency represents a package dependency for bundle analysis
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"` // dependencies, devDependencies, etc.
}

// Stub implementations of the analyzers - these will be implemented in later tasks
// Note: VulnerabilityDatabase is now implemented in vulnerability_scanner.go

// Note: LicenseChecker is now implemented in license_checker.go

// Note: UpdateChecker is now implemented in update_checker.go