package analysis

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLicenseChecker(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	assert.NotNil(t, lc)
	assert.NotNil(t, lc.client)
	assert.NotNil(t, lc.licenseMatrix)
	assert.NotNil(t, lc.spdxDatabase)
	assert.NotNil(t, lc.cache)

	defer lc.Close()
}

func TestNormalizeToSPDX(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "exact_mit",
			input:    "MIT",
			expected: "MIT",
		},
		{
			name:     "case_insensitive_mit",
			input:    "mit",
			expected: "MIT",
		},
		{
			name:     "apache_2_0",
			input:    "Apache 2.0",
			expected: "Apache-2.0",
		},
		{
			name:     "apache_2_0_exact",
			input:    "Apache-2.0",
			expected: "Apache-2.0",
		},
		{
			name:     "gpl_3",
			input:    "GPL-3.0",
			expected: "GPL-3.0-only",
		},
		{
			name:     "bsd_3_clause",
			input:    "BSD-3-Clause",
			expected: "BSD-3-Clause",
		},
		{
			name:     "public_domain",
			input:    "Public Domain",
			expected: "CC0-1.0",
		},
		{
			name:     "unknown_license",
			input:    "Custom License",
			expected: "",
		},
		{
			name:     "empty_input",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lc.normalizeToSPDX(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyzeLicenseCharacteristics(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name         string
		spdxID       string
		expectedType string
		expectedRisk string
	}{
		{
			name:         "mit_license",
			spdxID:       "MIT",
			expectedType: "permissive",
			expectedRisk: "low",
		},
		{
			name:         "apache_license",
			spdxID:       "Apache-2.0",
			expectedType: "permissive",
			expectedRisk: "low",
		},
		{
			name:         "gpl_license",
			spdxID:       "GPL-3.0-only",
			expectedType: "strong-copyleft",
			expectedRisk: "high",
		},
		{
			name:         "lgpl_license",
			spdxID:       "LGPL-3.0-only",
			expectedType: "weak-copyleft",
			expectedRisk: "medium",
		},
		{
			name:         "public_domain",
			spdxID:       "CC0-1.0",
			expectedType: "public-domain",
			expectedRisk: "low",
		},
		{
			name:         "unknown_license",
			spdxID:       "",
			expectedType: "unknown",
			expectedRisk: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PackageLicenseInfo{
				SPDXIdentifier: tt.spdxID,
			}

			lc.analyzeLicenseCharacteristics(info)
			assert.Equal(t, tt.expectedType, info.LicenseType)
			assert.Equal(t, tt.expectedRisk, info.RiskLevel)
		})
	}
}

func TestCheckLicenseConflict(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name             string
		pkg1License      string
		pkg1Type         string
		pkg2License      string
		pkg2Type         string
		shouldConflict   bool
		expectedSeverity string
	}{
		{
			name:           "mit_mit_no_conflict",
			pkg1License:    "MIT",
			pkg1Type:       "permissive",
			pkg2License:    "MIT",
			pkg2Type:       "permissive",
			shouldConflict: false,
		},
		{
			name:           "mit_apache_no_conflict",
			pkg1License:    "MIT",
			pkg1Type:       "permissive",
			pkg2License:    "Apache-2.0",
			pkg2Type:       "permissive",
			shouldConflict: false,
		},
		{
			name:             "gpl_proprietary_conflict",
			pkg1License:      "GPL-3.0-only",
			pkg1Type:         "strong-copyleft",
			pkg2License:      "proprietary",
			pkg2Type:         "proprietary",
			shouldConflict:   true,
			expectedSeverity: "critical",
		},
		{
			name:           "mit_gpl_no_conflict",
			pkg1License:    "MIT",
			pkg1Type:       "permissive",
			pkg2License:    "GPL-3.0-only",
			pkg2Type:       "strong-copyleft",
			shouldConflict: false,
		},
		{
			name:           "unknown_license_no_conflict",
			pkg1License:    "",
			pkg1Type:       "unknown",
			pkg2License:    "MIT",
			pkg2Type:       "permissive",
			shouldConflict: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg1 := &PackageLicenseInfo{
				PackageName:    "package1",
				SPDXIdentifier: tt.pkg1License,
				LicenseType:    tt.pkg1Type,
			}

			pkg2 := &PackageLicenseInfo{
				PackageName:    "package2",
				SPDXIdentifier: tt.pkg2License,
				LicenseType:    tt.pkg2Type,
			}

			conflict := lc.checkLicenseConflict(pkg1, pkg2)

			if !tt.shouldConflict {
				assert.Nil(t, conflict)
			} else {
				require.NotNil(t, conflict)
				assert.Equal(t, "package1", conflict.Package1)
				assert.Equal(t, "package2", conflict.Package2)
				assert.Equal(t, tt.pkg1License, conflict.License1)
				assert.Equal(t, tt.pkg2License, conflict.License2)
				if tt.expectedSeverity != "" {
					assert.Equal(t, tt.expectedSeverity, conflict.Severity)
				}
			}
		})
	}
}

func TestEvaluatePolicy(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	// Create test policy
	policy := LicensePolicy{
		Name:              "Test Policy",
		ForbiddenLicenses: []string{"GPL-3.0-only"},
		RequiresApproval:  []string{"LGPL-3.0-only"},
		EnforcementLevel:  "strict",
	}

	tests := []struct {
		name                  string
		spdxID                string
		expectedViolations    int
		expectedViolationType string
	}{
		{
			name:               "allowed_license",
			spdxID:             "MIT",
			expectedViolations: 0,
		},
		{
			name:                  "forbidden_license",
			spdxID:                "GPL-3.0-only",
			expectedViolations:    1,
			expectedViolationType: "forbidden",
		},
		{
			name:                  "requires_approval",
			spdxID:                "LGPL-3.0-only",
			expectedViolations:    1,
			expectedViolationType: "requires-approval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &PackageLicenseInfo{
				PackageName:    "test-package",
				SPDXIdentifier: tt.spdxID,
			}

			violations := lc.evaluatePolicy(info, policy)
			assert.Len(t, violations, tt.expectedViolations)

			if tt.expectedViolations > 0 {
				assert.Equal(t, tt.expectedViolationType, violations[0].ViolationType)
				assert.Equal(t, "Test Policy", violations[0].PolicyName)
			}
		})
	}
}

func TestAssessOverallRisk(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name         string
		packages     []*PackageLicenseInfo
		conflicts    []LicenseConflict
		expectedRisk string
	}{
		{
			name: "low_risk_scenario",
			packages: []*PackageLicenseInfo{
				{RiskLevel: "low"},
				{RiskLevel: "low"},
			},
			conflicts:    []LicenseConflict{},
			expectedRisk: "low",
		},
		{
			name: "mixed_risk_scenario",
			packages: []*PackageLicenseInfo{
				{RiskLevel: "low"},
				{RiskLevel: "high"},
				{RiskLevel: "medium"},
			},
			conflicts:    []LicenseConflict{},
			expectedRisk: "medium",
		},
		{
			name: "high_risk_with_conflicts",
			packages: []*PackageLicenseInfo{
				{RiskLevel: "high"},
				{RiskLevel: "critical"},
			},
			conflicts: []LicenseConflict{
				{Severity: "critical"},
			},
			expectedRisk: "critical",
		},
		{
			name:         "no_packages",
			packages:     []*PackageLicenseInfo{},
			conflicts:    []LicenseConflict{},
			expectedRisk: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := lc.assessOverallRisk(tt.packages, tt.conflicts)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

func TestGenerateLicenseRecommendations(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name                         string
		packages                     []*PackageLicenseInfo
		conflicts                    []LicenseConflict
		shouldContainConflictWarning bool
		shouldContainUnknownWarning  bool
		shouldContainCopyleftWarning bool
	}{
		{
			name: "no_issues",
			packages: []*PackageLicenseInfo{
				{PackageName: "pkg1", LicenseType: "permissive"},
				{PackageName: "pkg2", LicenseType: "permissive"},
			},
			conflicts: []LicenseConflict{},
		},
		{
			name: "has_conflicts",
			packages: []*PackageLicenseInfo{
				{PackageName: "pkg1", LicenseType: "strong-copyleft"},
				{PackageName: "pkg2", LicenseType: "proprietary"},
			},
			conflicts: []LicenseConflict{
				{Severity: "critical"},
			},
			shouldContainConflictWarning: true,
		},
		{
			name: "has_unknown_licenses",
			packages: []*PackageLicenseInfo{
				{PackageName: "pkg1", LicenseType: "unknown"},
				{PackageName: "pkg2", LicenseType: "permissive"},
			},
			conflicts:                   []LicenseConflict{},
			shouldContainUnknownWarning: true,
		},
		{
			name: "has_copyleft",
			packages: []*PackageLicenseInfo{
				{PackageName: "pkg1", LicenseType: "strong-copyleft"},
				{PackageName: "pkg2", LicenseType: "permissive"},
			},
			conflicts:                    []LicenseConflict{},
			shouldContainCopyleftWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := lc.generateLicenseRecommendations(tt.packages, tt.conflicts)
			assert.NotEmpty(t, recommendations)

			recommendationText := strings.Join(recommendations, " ")

			if tt.shouldContainConflictWarning {
				assert.Contains(t, strings.ToLower(recommendationText), "conflict")
			}
			if tt.shouldContainUnknownWarning {
				assert.Contains(t, strings.ToLower(recommendationText), "unknown")
			}
			if tt.shouldContainCopyleftWarning {
				assert.Contains(t, strings.ToLower(recommendationText), "copyleft")
			}
		})
	}
}

func TestMemoryLicenseCache(t *testing.T) {
	cache := &MemoryLicenseCache{
		cache: make(map[string]licenseCacheEntry),
	}

	// Test empty cache
	info, found := cache.GetLicense("test-package")
	assert.False(t, found)
	assert.Nil(t, info)

	// Test set and get
	testInfo := &PackageLicenseInfo{
		PackageName:    "test-package",
		Version:        "1.0.0",
		SPDXIdentifier: "MIT",
		LicenseType:    "permissive",
		RiskLevel:      "low",
	}

	cache.SetLicense("test-package", testInfo, time.Hour)
	retrievedInfo, found := cache.GetLicense("test-package")
	assert.True(t, found)
	assert.Equal(t, testInfo, retrievedInfo)

	// Test expiration
	cache.SetLicense("expired-package", testInfo, -time.Hour) // Already expired
	_, found = cache.GetLicense("expired-package")
	assert.False(t, found)

	// Test clear
	cache.Clear()
	_, found = cache.GetLicense("test-package")
	assert.False(t, found)
}

func TestGetLicenseType(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	tests := []struct {
		name     string
		spdxID   string
		expected string
	}{
		{"mit", "MIT", "permissive"},
		{"apache", "Apache-2.0", "permissive"},
		{"bsd", "BSD-3-Clause", "permissive"},
		{"isc", "ISC", "permissive"},
		{"lgpl", "LGPL-3.0-only", "weak-copyleft"},
		{"mpl", "MPL-2.0", "weak-copyleft"},
		{"gpl", "GPL-3.0-only", "strong-copyleft"},
		{"agpl", "AGPL-3.0-only", "strong-copyleft"},
		{"proprietary", "proprietary-license", "proprietary"},
		{"unknown", "Custom-License", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lc.getLicenseType(tt.spdxID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddCustomPolicy(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	initialPolicyCount := len(lc.customPolicies)

	policy := LicensePolicy{
		Name:              "Custom Policy",
		ForbiddenLicenses: []string{"GPL-3.0-only"},
		RequiresApproval:  []string{"LGPL-3.0-only"},
	}

	lc.AddCustomPolicy(policy)
	assert.Len(t, lc.customPolicies, initialPolicyCount+1)
	assert.Equal(t, "Custom Policy", lc.customPolicies[len(lc.customPolicies)-1].Name)
}

func TestGetLicenseInfo(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	// Test existing license
	info, exists := lc.GetLicenseInfo("MIT")
	assert.True(t, exists)
	assert.Equal(t, "MIT", info.ID)
	assert.Equal(t, "MIT License", info.Name)
	assert.True(t, info.IsOSIApproved)

	// Test non-existing license
	_, exists = lc.GetLicenseInfo("NON-EXISTENT")
	assert.False(t, exists)
}

// Integration test
func TestCheckLicensesIntegration(t *testing.T) {
	lc, err := NewLicenseChecker()
	require.NoError(t, err)
	defer lc.Close()

	ctx := context.Background()
	packages := []string{
		"react@18.0.0",
		"lodash@4.17.21",
		"express@4.18.0",
	}

	report, err := lc.CheckLicenses(ctx, packages)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 3, report.TotalPackages)
	assert.NotEmpty(t, report.LicenseDistribution)
	assert.NotEmpty(t, report.Recommendations)
}
