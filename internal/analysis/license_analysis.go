// Package analysis provides license analysis helper functions
package analysis

import (
	"fmt"
	"strings"
	"time"
)

// analyzeCompatibilityConflicts identifies conflicts between package licenses
func (lc *LicenseChecker) analyzeCompatibilityConflicts(packages []*PackageLicenseInfo) []LicenseConflict {
	var conflicts []LicenseConflict

	// Check for conflicts between copyleft and proprietary licenses
	for i, pkg1 := range packages {
		for j, pkg2 := range packages {
			if i >= j {
				continue // avoid duplicates and self-comparison
			}

			conflict := lc.checkLicenseConflict(pkg1, pkg2)
			if conflict != nil {
				conflicts = append(conflicts, *conflict)
			}
		}
	}

	return conflicts
}

// checkLicenseConflict checks for conflicts between two package licenses
func (lc *LicenseChecker) checkLicenseConflict(pkg1, pkg2 *PackageLicenseInfo) *LicenseConflict {
	// No conflict if either license is unknown
	if pkg1.SPDXIdentifier == "" || pkg2.SPDXIdentifier == "" {
		return nil
	}

	// Get compatibility rule from matrix
	rule := lc.getCompatibilityRule(pkg1.SPDXIdentifier, pkg2.SPDXIdentifier)
	if rule == nil || rule.Compatible {
		return nil
	}

	// Determine conflict severity
	severity := "medium"
	if pkg1.LicenseType == "strong-copyleft" || pkg2.LicenseType == "strong-copyleft" {
		if pkg1.LicenseType == "proprietary" || pkg2.LicenseType == "proprietary" {
			severity = "critical"
		} else {
			severity = "high"
		}
	}

	return &LicenseConflict{
		Package1:     pkg1.PackageName,
		Package2:     pkg2.PackageName,
		License1:     pkg1.SPDXIdentifier,
		License2:     pkg2.SPDXIdentifier,
		ConflictType: determineConflictType(pkg1.LicenseType, pkg2.LicenseType),
		Severity:     severity,
	}
}

// getCompatibilityRule gets compatibility rule between two licenses
func (lc *LicenseChecker) getCompatibilityRule(license1, license2 string) *CompatibilityRule {
	if lc.licenseMatrix.Rules == nil {
		return nil
	}

	// Check both directions
	if rules, exists := lc.licenseMatrix.Rules[license1]; exists {
		if rule, exists := rules[license2]; exists {
			return &rule
		}
	}

	if rules, exists := lc.licenseMatrix.Rules[license2]; exists {
		if rule, exists := rules[license1]; exists {
			return &rule
		}
	}

	// Use default compatibility based on license types
	return lc.getDefaultCompatibilityRule(license1, license2)
}

// getDefaultCompatibilityRule provides default compatibility based on license types
func (lc *LicenseChecker) getDefaultCompatibilityRule(license1, license2 string) *CompatibilityRule {
	type1 := lc.getLicenseType(license1)
	type2 := lc.getLicenseType(license2)

	// Permissive licenses are generally compatible with everything
	if type1 == "permissive" || type2 == "permissive" {
		return &CompatibilityRule{
			Compatible: true,
			RiskLevel:  "low",
		}
	}

	// Strong copyleft conflicts with proprietary
	if (type1 == "strong-copyleft" && type2 == "proprietary") ||
	   (type1 == "proprietary" && type2 == "strong-copyleft") {
		return &CompatibilityRule{
			Compatible: false,
			RiskLevel:  "critical",
			Conditions: []string{"Copyleft license incompatible with proprietary license"},
		}
	}

	// Default to compatible with medium risk
	return &CompatibilityRule{
		Compatible: true,
		RiskLevel:  "medium",
	}
}

// getLicenseType gets license type for SPDX identifier
func (lc *LicenseChecker) getLicenseType(spdx string) string {
	permissive := []string{"MIT", "BSD-2-Clause", "BSD-3-Clause", "ISC", "Apache-2.0"}
	weakCopyleft := []string{"LGPL-2.1-only", "LGPL-3.0-only", "MPL-2.0", "EPL-1.0", "EPL-2.0"}
	strongCopyleft := []string{"GPL-2.0-only", "GPL-3.0-only", "AGPL-3.0-only"}

	for _, license := range permissive {
		if license == spdx {
			return "permissive"
		}
	}

	for _, license := range weakCopyleft {
		if license == spdx {
			return "weak-copyleft"
		}
	}

	for _, license := range strongCopyleft {
		if license == spdx {
			return "strong-copyleft"
		}
	}

	if strings.Contains(strings.ToLower(spdx), "proprietary") {
		return "proprietary"
	}

	return "other"
}

// determineConflictType determines the type of license conflict
func determineConflictType(type1, type2 string) string {
	if (type1 == "strong-copyleft" && type2 == "proprietary") ||
	   (type1 == "proprietary" && type2 == "strong-copyleft") {
		return "copyleft-proprietary"
	}

	if (type1 == "strong-copyleft" && type2 == "weak-copyleft") ||
	   (type1 == "weak-copyleft" && type2 == "strong-copyleft") {
		return "copyleft-mismatch"
	}

	return "incompatible"
}

// assessOverallRisk assesses overall license risk for the project
func (lc *LicenseChecker) assessOverallRisk(packages []*PackageLicenseInfo, conflicts []LicenseConflict) string {
	riskScore := 0.0
	packageCount := float64(len(packages))

	if packageCount == 0 {
		return "unknown"
	}

	// Calculate weighted risk score
	for _, pkg := range packages {
		switch pkg.RiskLevel {
		case "critical":
			riskScore += 4.0
		case "high":
			riskScore += 3.0
		case "medium":
			riskScore += 2.0
		case "low":
			riskScore += 1.0
		}
	}

	// Add conflict penalty
	for _, conflict := range conflicts {
		switch conflict.Severity {
		case "critical":
			riskScore += 5.0
		case "high":
			riskScore += 3.0
		case "medium":
			riskScore += 2.0
		}
	}

	// Normalize by package count
	normalizedScore := riskScore / packageCount

	// Determine overall risk level
	switch {
	case normalizedScore >= 4.0:
		return "critical"
	case normalizedScore >= 3.0:
		return "high"
	case normalizedScore >= 2.0:
		return "medium"
	default:
		return "low"
	}
}

// generateLicenseRecommendations generates recommendations based on license analysis
func (lc *LicenseChecker) generateLicenseRecommendations(packages []*PackageLicenseInfo, conflicts []LicenseConflict) []string {
	var recommendations []string

	// Count license types
	licenseCounts := make(map[string]int)
	var unknownPackages []string
	var proprietaryPackages []string
	var copyleftPackages []string

	for _, pkg := range packages {
		licenseCounts[pkg.LicenseType]++
		
		switch pkg.LicenseType {
		case "unknown":
			unknownPackages = append(unknownPackages, pkg.PackageName)
		case "proprietary":
			proprietaryPackages = append(proprietaryPackages, pkg.PackageName)
		case "strong-copyleft":
			copyleftPackages = append(copyleftPackages, pkg.PackageName)
		}
	}

	// Generate specific recommendations
	if len(conflicts) > 0 {
		recommendations = append(recommendations, 
			"CRITICAL: License conflicts detected - review incompatible license combinations")
		
		for _, conflict := range conflicts {
			if conflict.Severity == "critical" {
				recommendations = append(recommendations,
					"Replace packages with conflicting licenses or obtain legal approval")
				break
			}
		}
	}

	if len(unknownPackages) > 0 {
		recommendations = append(recommendations,
			"Review packages with unknown licenses - manual license verification required")
	}

	if len(proprietaryPackages) > 0 {
		recommendations = append(recommendations,
			"Verify proprietary license terms for distribution rights")
	}

	if len(copyleftPackages) > 0 {
		recommendations = append(recommendations,
			"Review copyleft license obligations for source code disclosure requirements")
	}

	if licenseCounts["strong-copyleft"] > 0 && licenseCounts["proprietary"] > 0 {
		recommendations = append(recommendations,
			"Consider separating copyleft and proprietary components into different modules")
	}

	// Policy recommendations
	if len(packages) > 20 {
		recommendations = append(recommendations,
			"Consider implementing automated license policy enforcement")
	}

	// Default recommendation if no issues
	if len(recommendations) == 0 {
		recommendations = append(recommendations,
			"License compliance looks good - continue monitoring new dependencies")
	}

	return recommendations
}

// initializeCompatibilityMatrix creates default license compatibility matrix
func initializeCompatibilityMatrix() *CompatibilityMatrix {
	matrix := &CompatibilityMatrix{
		Rules:         make(map[string]map[string]CompatibilityRule),
		LicenseTypes:  make(map[string]LicenseTypeInfo),
		DefaultPolicy: "permissive-preferred",
	}

	// Define license type information
	matrix.LicenseTypes["permissive"] = LicenseTypeInfo{
		Category:      "permissive",
		Description:   "Allows almost unrestricted freedom with code",
		Restrictions:  []string{"attribution-required"},
		Requirements:  []string{"include-copyright-notice"},
		CommercialUse: true,
		Copyleft:      false,
	}

	matrix.LicenseTypes["weak-copyleft"] = LicenseTypeInfo{
		Category:      "weak-copyleft",
		Description:   "Requires source availability for modifications to the library itself",
		Restrictions:  []string{"modification-disclosure"},
		Requirements:  []string{"source-availability", "same-license-for-modifications"},
		CommercialUse: true,
		Copyleft:      true,
	}

	matrix.LicenseTypes["strong-copyleft"] = LicenseTypeInfo{
		Category:      "strong-copyleft",
		Description:   "Requires entire work to be licensed under same copyleft terms",
		Restrictions:  []string{"proprietary-linking", "closed-source-distribution"},
		Requirements:  []string{"source-disclosure", "same-license", "patent-grant"},
		CommercialUse: true,
		Copyleft:      true,
	}

	matrix.LicenseTypes["proprietary"] = LicenseTypeInfo{
		Category:      "proprietary",
		Description:   "Commercial license with specific terms and restrictions",
		Restrictions:  []string{"distribution-limited", "modification-restricted"},
		Requirements:  []string{"license-agreement", "payment-required"},
		CommercialUse: false, // depends on specific license
		Copyleft:      false,
	}

	// Initialize compatibility rules for common combinations
	initializeCommonCompatibilityRules(matrix)

	return matrix
}

// initializeCommonCompatibilityRules sets up rules for common license combinations
func initializeCommonCompatibilityRules(matrix *CompatibilityMatrix) {
	// MIT compatibility (very permissive)
	matrix.Rules["MIT"] = make(map[string]CompatibilityRule)
	matrix.Rules["MIT"]["MIT"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}
	matrix.Rules["MIT"]["Apache-2.0"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}
	matrix.Rules["MIT"]["BSD-3-Clause"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}
	matrix.Rules["MIT"]["GPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "medium", 
		Conditions: []string{"Combined work must be GPL-3.0"}}
	matrix.Rules["MIT"]["LGPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}

	// Apache-2.0 compatibility
	matrix.Rules["Apache-2.0"] = make(map[string]CompatibilityRule)
	matrix.Rules["Apache-2.0"]["Apache-2.0"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}
	matrix.Rules["Apache-2.0"]["GPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "medium",
		Conditions: []string{"Combined work must be GPL-3.0"}}
	matrix.Rules["Apache-2.0"]["GPL-2.0-only"] = CompatibilityRule{Compatible: false, RiskLevel: "high",
		Conditions: []string{"Apache-2.0 incompatible with GPL-2.0"}}

	// GPL-3.0 compatibility (strong copyleft)
	matrix.Rules["GPL-3.0-only"] = make(map[string]CompatibilityRule)
	matrix.Rules["GPL-3.0-only"]["GPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "medium"}
	matrix.Rules["GPL-3.0-only"]["LGPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "medium"}
	matrix.Rules["GPL-3.0-only"]["proprietary"] = CompatibilityRule{Compatible: false, RiskLevel: "critical",
		Conditions: []string{"GPL cannot be combined with proprietary code"}}

	// LGPL compatibility (weak copyleft)
	matrix.Rules["LGPL-3.0-only"] = make(map[string]CompatibilityRule)
	matrix.Rules["LGPL-3.0-only"]["LGPL-3.0-only"] = CompatibilityRule{Compatible: true, RiskLevel: "low"}
	matrix.Rules["LGPL-3.0-only"]["proprietary"] = CompatibilityRule{Compatible: true, RiskLevel: "medium",
		Conditions: []string{"Dynamic linking allowed, static linking requires source disclosure"}}
}

// initializeSPDXDatabase creates SPDX license database
func initializeSPDXDatabase() *SPDXDatabase {
	db := &SPDXDatabase{
		Identifiers: make(map[string]SPDXLicenseInfo),
	}

	// Add common SPDX licenses
	licenses := []SPDXLicenseInfo{
		{
			ID:            "MIT",
			Name:          "MIT License",
			Reference:     "https://opensource.org/licenses/MIT",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "permissive",
		},
		{
			ID:            "Apache-2.0",
			Name:          "Apache License 2.0",
			Reference:     "https://opensource.org/licenses/Apache-2.0",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "permissive",
		},
		{
			ID:            "GPL-3.0-only",
			Name:          "GNU General Public License v3.0 only",
			Reference:     "https://opensource.org/licenses/GPL-3.0",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "strong-copyleft",
		},
		{
			ID:            "LGPL-3.0-only",
			Name:          "GNU Lesser General Public License v3.0 only",
			Reference:     "https://opensource.org/licenses/LGPL-3.0",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "weak-copyleft",
		},
		{
			ID:            "BSD-3-Clause",
			Name:          "BSD 3-Clause \"New\" or \"Revised\" License",
			Reference:     "https://opensource.org/licenses/BSD-3-Clause",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "permissive",
		},
		{
			ID:            "ISC",
			Name:          "ISC License",
			Reference:     "https://opensource.org/licenses/ISC",
			IsOSIApproved: true,
			IsFSFLibre:    true,
			Category:      "permissive",
		},
	}

	for _, license := range licenses {
		db.Identifiers[license.ID] = license
	}

	return db
}

// AddCustomPolicy adds a custom license policy
func (lc *LicenseChecker) AddCustomPolicy(policy LicensePolicy) {
	lc.customPolicies = append(lc.customPolicies, policy)
}

// GetLicenseInfo retrieves detailed information about a specific license
func (lc *LicenseChecker) GetLicenseInfo(spdxID string) (*SPDXLicenseInfo, bool) {
	info, exists := lc.spdxDatabase.Identifiers[spdxID]
	return &info, exists
}

// ValidateLicenseCompliance validates overall license compliance
func (lc *LicenseChecker) ValidateLicenseCompliance(packages []*PackageLicenseInfo) (bool, []string) {
	var issues []string
	compliant := true

	// Check for critical conflicts
	conflicts := lc.analyzeCompatibilityConflicts(packages)
	for _, conflict := range conflicts {
		if conflict.Severity == "critical" {
			compliant = false
			issues = append(issues, fmt.Sprintf("Critical license conflict: %s (%s) vs %s (%s)", 
				conflict.Package1, conflict.License1, conflict.Package2, conflict.License2))
		}
	}

	// Check policy violations
	for _, pkg := range packages {
		for _, violation := range pkg.PolicyViolations {
			if violation.Severity == "high" && violation.ViolationType == "forbidden" {
				compliant = false
				issues = append(issues, fmt.Sprintf("Policy violation: %s", violation.Description))
			}
		}
	}

	return compliant, issues
}

// Memory cache implementation

func (mlc *MemoryLicenseCache) GetLicense(packageName string) (*PackageLicenseInfo, bool) {
	entry, exists := mlc.cache[packageName]
	if !exists || entry.expiry.Before(time.Now()) {
		delete(mlc.cache, packageName)
		return nil, false
	}
	return entry.info, true
}

func (mlc *MemoryLicenseCache) SetLicense(packageName string, info *PackageLicenseInfo, ttl time.Duration) {
	mlc.cache[packageName] = licenseCacheEntry{
		info:   info,
		expiry: time.Now().Add(ttl),
	}
}

func (mlc *MemoryLicenseCache) Clear() {
	mlc.cache = make(map[string]licenseCacheEntry)
}