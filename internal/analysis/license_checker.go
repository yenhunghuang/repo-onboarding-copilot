// Package analysis provides license compatibility analysis for npm packages
package analysis

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LicenseChecker analyzes package licenses and compatibility
type LicenseChecker struct {
	client         *http.Client
	licenseMatrix  *CompatibilityMatrix
	spdxDatabase   *SPDXDatabase
	customPolicies []LicensePolicy
	cache          LicenseCache
}

// LicenseCache provides caching for license data
type LicenseCache interface {
	GetLicense(packageName string) (*PackageLicenseInfo, bool)
	SetLicense(packageName string, info *PackageLicenseInfo, ttl time.Duration)
	Clear()
}

// PackageLicenseInfo represents comprehensive license information for a package
type PackageLicenseInfo struct {
	PackageName      string               `json:"package_name"`
	Version          string               `json:"version"`
	DeclaredLicense  string               `json:"declared_license"`  // from package.json
	DetectedLicenses []DetectedLicense    `json:"detected_licenses"` // from file analysis
	SPDXIdentifier   string               `json:"spdx_identifier"`
	LicenseType      string               `json:"license_type"` // permissive, copyleft, proprietary
	LicenseText      string               `json:"license_text"`
	LicenseURL       string               `json:"license_url"`
	Compatibility    LicenseCompatibility `json:"compatibility"`
	RiskLevel        string               `json:"risk_level"` // low, medium, high, critical
	PolicyViolations []PolicyViolation    `json:"policy_violations"`
	LastAnalyzed     time.Time            `json:"last_analyzed"`
}

// DetectedLicense represents a license detected from file analysis
type DetectedLicense struct {
	SPDX       string  `json:"spdx"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"` // 0-1 confidence score
	Source     string  `json:"source"`     // file path or source
	Text       string  `json:"text"`       // license text excerpt
}

// LicenseCompatibility represents compatibility assessment
type LicenseCompatibility struct {
	Compatible   bool                 `json:"compatible"`
	Conflicts    []LicenseConflict    `json:"conflicts"`
	Restrictions []LicenseRestriction `json:"restrictions"`
	Requirements []LicenseRequirement `json:"requirements"`
	RiskScore    float64              `json:"risk_score"` // 0-1 scale
}

// LicenseRestriction represents license-imposed restrictions
type LicenseRestriction struct {
	Type        string `json:"type"` // distribution, modification, commercial-use
	Description string `json:"description"`
	Severity    string `json:"severity"` // low, medium, high
	Mitigation  string `json:"mitigation"`
}

// LicenseRequirement represents license-imposed requirements
type LicenseRequirement struct {
	Type        string `json:"type"` // attribution, source-disclosure, same-license
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
	Scope       string `json:"scope"` // package, derivative-works, distribution
}

// PolicyViolation represents a license policy violation
type PolicyViolation struct {
	PolicyName    string `json:"policy_name"`
	ViolationType string `json:"violation_type"` // forbidden, restricted, requires-approval
	Description   string `json:"description"`
	Severity      string `json:"severity"`
	Resolution    string `json:"resolution"`
}

// LicensePolicy represents custom license policies
type LicensePolicy struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	AllowedLicenses   []string          `json:"allowed_licenses"`
	ForbiddenLicenses []string          `json:"forbidden_licenses"`
	RequiresApproval  []string          `json:"requires_approval"`
	Restrictions      map[string]string `json:"restrictions"`
	EnforcementLevel  string            `json:"enforcement_level"` // strict, moderate, advisory
}

// CompatibilityMatrix defines license compatibility rules
type CompatibilityMatrix struct {
	Rules         map[string]map[string]CompatibilityRule `json:"rules"`
	LicenseTypes  map[string]LicenseTypeInfo              `json:"license_types"`
	DefaultPolicy string                                  `json:"default_policy"`
}

// CompatibilityRule defines compatibility between two licenses
type CompatibilityRule struct {
	Compatible   bool     `json:"compatible"`
	Conditions   []string `json:"conditions"`
	Warnings     []string `json:"warnings"`
	Requirements []string `json:"requirements"`
	RiskLevel    string   `json:"risk_level"`
}

// LicenseTypeInfo provides information about license categories
type LicenseTypeInfo struct {
	Category      string   `json:"category"` // permissive, weak-copyleft, strong-copyleft, proprietary
	Description   string   `json:"description"`
	Restrictions  []string `json:"restrictions"`
	Requirements  []string `json:"requirements"`
	CommercialUse bool     `json:"commercial_use"`
	Copyleft      bool     `json:"copyleft"`
}

// SPDXDatabase provides SPDX license identifier resolution
type SPDXDatabase struct {
	Identifiers map[string]SPDXLicenseInfo `json:"identifiers"`
	LastUpdate  time.Time                  `json:"last_update"`
}

// SPDXLicenseInfo represents SPDX license information
type SPDXLicenseInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Reference     string   `json:"reference"`
	IsDeprecated  bool     `json:"is_deprecated"`
	IsOSIApproved bool     `json:"is_osi_approved"`
	IsFSFLibre    bool     `json:"is_fsf_libre"`
	SeeAlso       []string `json:"see_also"`
	Category      string   `json:"category"`
}

// MemoryLicenseCache provides in-memory caching for license data
type MemoryLicenseCache struct {
	cache map[string]licenseCacheEntry
}

type licenseCacheEntry struct {
	info   *PackageLicenseInfo
	expiry time.Time
}

// NewLicenseChecker creates a new license checker
func NewLicenseChecker() (*LicenseChecker, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize compatibility matrix with common licenses
	matrix := initializeCompatibilityMatrix()

	// Initialize SPDX database
	spdxDB := initializeSPDXDatabase()

	cache := &MemoryLicenseCache{
		cache: make(map[string]licenseCacheEntry),
	}

	lc := &LicenseChecker{
		client:         client,
		licenseMatrix:  matrix,
		spdxDatabase:   spdxDB,
		customPolicies: []LicensePolicy{},
		cache:          cache,
	}

	return lc, nil
}

// CheckLicenses analyzes license compatibility for packages
func (lc *LicenseChecker) CheckLicenses(ctx context.Context, packages []string) (*LicenseReport, error) {
	report := &LicenseReport{
		LicenseDistribution: make(map[string]int),
		CompatibilityIssues: []LicenseConflict{},
		UnknownLicenses:     []string{},
		ProprietaryPackages: []string{},
		CopyleftPackages:    []string{},
		Recommendations:     []string{},
	}

	var packageLicenses []*PackageLicenseInfo
	processed := make(map[string]bool)

	// Analyze each package
	for _, pkg := range packages {
		if processed[pkg] {
			continue
		}
		processed[pkg] = true

		name, version, err := parsePackageSpec(pkg)
		if err != nil {
			continue
		}

		// Get or analyze license information
		licenseInfo, err := lc.analyzeLicense(ctx, name, version)
		if err != nil {
			continue
		}

		packageLicenses = append(packageLicenses, licenseInfo)

		// Update distribution
		if licenseInfo.SPDXIdentifier != "" {
			report.LicenseDistribution[licenseInfo.SPDXIdentifier]++
		} else if licenseInfo.DeclaredLicense != "" {
			report.LicenseDistribution[licenseInfo.DeclaredLicense]++
		} else {
			report.LicenseDistribution["UNKNOWN"]++
			report.UnknownLicenses = append(report.UnknownLicenses, name)
		}

		// Categorize packages
		switch licenseInfo.LicenseType {
		case "proprietary":
			report.ProprietaryPackages = append(report.ProprietaryPackages, name)
		case "copyleft", "strong-copyleft":
			report.CopyleftPackages = append(report.CopyleftPackages, name)
		}
	}

	// Analyze compatibility between packages
	conflicts := lc.analyzeCompatibilityConflicts(packageLicenses)
	report.CompatibilityIssues = conflicts

	// Assess overall risk
	report.RiskAssessment = lc.assessOverallRisk(packageLicenses, conflicts)

	// Generate recommendations
	report.Recommendations = lc.generateLicenseRecommendations(packageLicenses, conflicts)

	// Update totals
	report.TotalPackages = len(packageLicenses)

	return report, nil
}

// analyzeLicense analyzes license information for a specific package
func (lc *LicenseChecker) analyzeLicense(ctx context.Context, packageName, version string) (*PackageLicenseInfo, error) {
	// Check cache first
	cacheKey := packageName + "@" + version
	if info, found := lc.cache.GetLicense(cacheKey); found {
		return info, nil
	}

	// Create license info structure
	info := &PackageLicenseInfo{
		PackageName:      packageName,
		Version:          version,
		DetectedLicenses: []DetectedLicense{},
		PolicyViolations: []PolicyViolation{},
		LastAnalyzed:     time.Now(),
	}

	// Extract license from package metadata (would integrate with npm registry)
	declaredLicense := lc.extractDeclaredLicense(ctx, packageName, version)
	info.DeclaredLicense = declaredLicense

	// Normalize to SPDX identifier
	info.SPDXIdentifier = lc.normalizeToSPDX(declaredLicense)

	// Determine license type and characteristics
	lc.analyzeLicenseCharacteristics(info)

	// Assess compatibility and risk
	lc.assessLicenseCompatibility(info)

	// Check policy violations
	lc.checkPolicyViolations(info)

	// Cache the result
	lc.cache.SetLicense(cacheKey, info, 24*time.Hour)

	return info, nil
}

// extractDeclaredLicense extracts declared license from package metadata
func (lc *LicenseChecker) extractDeclaredLicense(ctx context.Context, packageName, version string) string {
	// This would integrate with npm registry to get package.json license field
	// For now, return mock data based on common patterns

	// Common licenses for popular packages
	commonLicenses := map[string]string{
		"react":      "MIT",
		"vue":        "MIT",
		"angular":    "MIT",
		"lodash":     "MIT",
		"express":    "MIT",
		"typescript": "Apache-2.0",
		"webpack":    "MIT",
		"babel":      "MIT",
		"eslint":     "MIT",
		"jest":       "MIT",
	}

	if license, exists := commonLicenses[packageName]; exists {
		return license
	}

	// Default for unknown packages
	return "UNKNOWN"
}

// normalizeToSPDX normalizes license string to SPDX identifier
func (lc *LicenseChecker) normalizeToSPDX(licenseString string) string {
	if licenseString == "" {
		return ""
	}

	// Common license mappings to SPDX
	mappings := map[string]string{
		"MIT":           "MIT",
		"Apache-2.0":    "Apache-2.0",
		"Apache 2.0":    "Apache-2.0",
		"GPL-2.0":       "GPL-2.0-only",
		"GPL-3.0":       "GPL-3.0-only",
		"LGPL-2.1":      "LGPL-2.1-only",
		"LGPL-3.0":      "LGPL-3.0-only",
		"BSD-2-Clause":  "BSD-2-Clause",
		"BSD-3-Clause":  "BSD-3-Clause",
		"ISC":           "ISC",
		"Unlicense":     "Unlicense",
		"CC0-1.0":       "CC0-1.0",
		"WTFPL":         "WTFPL",
		"Artistic-2.0":  "Artistic-2.0",
		"EPL-1.0":       "EPL-1.0",
		"EPL-2.0":       "EPL-2.0",
		"MPL-2.0":       "MPL-2.0",
		"CDDL-1.0":      "CDDL-1.0",
		"Public Domain": "CC0-1.0",
		"UNKNOWN":       "",
	}

	// Normalize case and spaces
	normalized := strings.ToUpper(strings.TrimSpace(licenseString))

	// Try exact matches first
	for key, spdx := range mappings {
		if strings.ToUpper(key) == normalized {
			return spdx
		}
	}

	// Try partial matches
	for key, spdx := range mappings {
		if strings.Contains(normalized, strings.ToUpper(key)) {
			return spdx
		}
	}

	return ""
}

// analyzeLicenseCharacteristics determines license type and properties
func (lc *LicenseChecker) analyzeLicenseCharacteristics(info *PackageLicenseInfo) {
	spdx := info.SPDXIdentifier
	if spdx == "" {
		info.LicenseType = "unknown"
		info.RiskLevel = "medium"
		return
	}

	// License categorization based on SPDX identifier
	switch spdx {
	case "MIT", "BSD-2-Clause", "BSD-3-Clause", "ISC", "Apache-2.0":
		info.LicenseType = "permissive"
		info.RiskLevel = "low"
	case "LGPL-2.1-only", "LGPL-3.0-only", "MPL-2.0", "EPL-1.0", "EPL-2.0":
		info.LicenseType = "weak-copyleft"
		info.RiskLevel = "medium"
	case "GPL-2.0-only", "GPL-3.0-only", "AGPL-3.0-only":
		info.LicenseType = "strong-copyleft"
		info.RiskLevel = "high"
	case "CC0-1.0", "Unlicense":
		info.LicenseType = "public-domain"
		info.RiskLevel = "low"
	default:
		if strings.Contains(strings.ToLower(spdx), "proprietary") {
			info.LicenseType = "proprietary"
			info.RiskLevel = "critical"
		} else {
			info.LicenseType = "other"
			info.RiskLevel = "medium"
		}
	}
}

// assessLicenseCompatibility assesses compatibility with project requirements
func (lc *LicenseChecker) assessLicenseCompatibility(info *PackageLicenseInfo) {
	compatibility := LicenseCompatibility{
		Compatible:   true,
		Conflicts:    []LicenseConflict{},
		Restrictions: []LicenseRestriction{},
		Requirements: []LicenseRequirement{},
		RiskScore:    0.0,
	}

	// Assess based on license type
	switch info.LicenseType {
	case "permissive":
		compatibility.RiskScore = 0.1
		compatibility.Requirements = append(compatibility.Requirements, LicenseRequirement{
			Type:        "attribution",
			Description: "Must include copyright notice and license text",
			Mandatory:   true,
			Scope:       "distribution",
		})

	case "weak-copyleft":
		compatibility.RiskScore = 0.4
		compatibility.Requirements = append(compatibility.Requirements, LicenseRequirement{
			Type:        "source-disclosure",
			Description: "Must provide source code for modifications to the library",
			Mandatory:   true,
			Scope:       "derivative-works",
		})

	case "strong-copyleft":
		compatibility.RiskScore = 0.8
		compatibility.Requirements = append(compatibility.Requirements, LicenseRequirement{
			Type:        "same-license",
			Description: "Entire work must be licensed under the same copyleft license",
			Mandatory:   true,
			Scope:       "distribution",
		})
		compatibility.Restrictions = append(compatibility.Restrictions, LicenseRestriction{
			Type:        "commercial-use",
			Description: "May require special consideration for commercial use",
			Severity:    "high",
			Mitigation:  "Consult legal counsel for commercial distribution",
		})

	case "proprietary":
		compatibility.RiskScore = 1.0
		compatibility.Compatible = false
		compatibility.Restrictions = append(compatibility.Restrictions, LicenseRestriction{
			Type:        "distribution",
			Description: "Proprietary license may prohibit distribution",
			Severity:    "critical",
			Mitigation:  "Review license terms or find alternative package",
		})
	}

	info.Compatibility = compatibility
}

// checkPolicyViolations checks against custom license policies
func (lc *LicenseChecker) checkPolicyViolations(info *PackageLicenseInfo) {
	for _, policy := range lc.customPolicies {
		violations := lc.evaluatePolicy(info, policy)
		info.PolicyViolations = append(info.PolicyViolations, violations...)
	}
}

// evaluatePolicy evaluates a package against a license policy
func (lc *LicenseChecker) evaluatePolicy(info *PackageLicenseInfo, policy LicensePolicy) []PolicyViolation {
	var violations []PolicyViolation

	spdx := info.SPDXIdentifier
	if spdx == "" {
		spdx = info.DeclaredLicense
	}

	// Check forbidden licenses
	for _, forbidden := range policy.ForbiddenLicenses {
		if spdx == forbidden || strings.Contains(strings.ToLower(spdx), strings.ToLower(forbidden)) {
			violations = append(violations, PolicyViolation{
				PolicyName:    policy.Name,
				ViolationType: "forbidden",
				Description:   fmt.Sprintf("License %s is forbidden by policy %s", spdx, policy.Name),
				Severity:      "high",
				Resolution:    "Replace package or obtain policy exception",
			})
		}
	}

	// Check if requires approval
	for _, requiresApproval := range policy.RequiresApproval {
		if spdx == requiresApproval || strings.Contains(strings.ToLower(spdx), strings.ToLower(requiresApproval)) {
			violations = append(violations, PolicyViolation{
				PolicyName:    policy.Name,
				ViolationType: "requires-approval",
				Description:   fmt.Sprintf("License %s requires approval under policy %s", spdx, policy.Name),
				Severity:      "medium",
				Resolution:    "Obtain approval from legal/compliance team",
			})
		}
	}

	return violations
}

// Close releases license checker resources
func (lc *LicenseChecker) Close() error {
	lc.cache.Clear()
	return nil
}

// Helper methods for compatibility analysis will be implemented in the next file
