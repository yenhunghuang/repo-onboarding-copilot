package analysis

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUpdateChecker(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	assert.NotNil(t, uc)
	assert.NotNil(t, uc.client)
	assert.NotNil(t, uc.cache)
	assert.NotNil(t, uc.semverParser)
	assert.Equal(t, "https://registry.npmjs.org", uc.registryURL)
	
	defer uc.Close()
}

func TestSemanticVersionParser_Parse(t *testing.T) {
	parser := &SemanticVersionParser{
		versionRegex: regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`),
	}

	tests := []struct {
		name            string
		version         string
		expectedMajor   int
		expectedMinor   int
		expectedPatch   int
		expectedPrerel  string
		expectedBuild   string
		shouldError     bool
	}{
		{
			name:          "simple_version",
			version:       "1.2.3",
			expectedMajor: 1,
			expectedMinor: 2,
			expectedPatch: 3,
		},
		{
			name:          "version_with_v_prefix",
			version:       "v2.0.1",
			expectedMajor: 2,
			expectedMinor: 0,
			expectedPatch: 1,
		},
		{
			name:           "prerelease_version",
			version:        "1.0.0-alpha.1",
			expectedMajor:  1,
			expectedMinor:  0,
			expectedPatch:  0,
			expectedPrerel: "alpha.1",
		},
		{
			name:          "build_metadata",
			version:       "1.0.0+20220101",
			expectedMajor: 1,
			expectedMinor: 0,
			expectedPatch: 0,
			expectedBuild: "20220101",
		},
		{
			name:           "prerelease_and_build",
			version:        "1.0.0-beta.1+exp.sha.5114f85",
			expectedMajor:  1,
			expectedMinor:  0,
			expectedPatch:  0,
			expectedPrerel: "beta.1",
			expectedBuild:  "exp.sha.5114f85",
		},
		{
			name:        "invalid_version",
			version:     "not.a.version",
			shouldError: true,
		},
		{
			name:        "empty_version",
			version:     "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.version)
			
			if tt.shouldError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMajor, result.Major)
			assert.Equal(t, tt.expectedMinor, result.Minor)
			assert.Equal(t, tt.expectedPatch, result.Patch)
			assert.Equal(t, tt.expectedPrerel, result.Prerelease)
			assert.Equal(t, tt.expectedBuild, result.Build)
			assert.Equal(t, tt.version, result.Raw)
		})
	}
}

func TestSemanticVersionParser_Compare(t *testing.T) {
	parser := &SemanticVersionParser{
		versionRegex: regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`),
	}

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal_versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "v1_greater_major",
			v1:       "2.0.0",
			v2:       "1.0.0",
			expected: 1,
		},
		{
			name:     "v1_lesser_major",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "v1_greater_minor",
			v1:       "1.2.0",
			v2:       "1.1.0",
			expected: 1,
		},
		{
			name:     "v1_greater_patch",
			v1:       "1.0.2",
			v2:       "1.0.1",
			expected: 1,
		},
		{
			name:     "prerelease_vs_release",
			v1:       "1.0.0",
			v2:       "1.0.0-alpha",
			expected: 1,
		},
		{
			name:     "prerelease_comparison",
			v1:       "1.0.0-beta",
			v2:       "1.0.0-alpha",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := parser.Parse(tt.v1)
			require.NoError(t, err)
			
			v2, err := parser.Parse(tt.v2)
			require.NoError(t, err)
			
			result := parser.Compare(v1, v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineUpdateType(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	defer uc.Close()

	tests := []struct {
		name     string
		current  string
		target   string
		expected string
	}{
		{
			name:     "major_update",
			current:  "1.0.0",
			target:   "2.0.0",
			expected: "major",
		},
		{
			name:     "minor_update",
			current:  "1.0.0",
			target:   "1.1.0",
			expected: "minor",
		},
		{
			name:     "patch_update",
			current:  "1.0.0",
			target:   "1.0.1",
			expected: "patch",
		},
		{
			name:     "prerelease_update",
			current:  "1.0.0",
			target:   "1.0.0-beta",
			expected: "prerelease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := uc.semverParser.Parse(tt.current)
			require.NoError(t, err)
			
			target, err := uc.semverParser.Parse(tt.target)
			require.NoError(t, err)
			
			result := uc.determineUpdateType(current, target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemoryUpdateCache(t *testing.T) {
	cache := &MemoryUpdateCache{
		cache: make(map[string]updateCacheEntry),
	}

	// Test empty cache
	data, found := cache.GetVersions("test-package")
	assert.False(t, found)
	assert.Nil(t, data)

	// Test set and get
	testData := &PackageVersionData{
		Name: "test-package",
		DistTags: map[string]string{
			"latest": "1.0.0",
		},
		Versions: map[string]PackageVersion{
			"1.0.0": {
				Version:     "1.0.0",
				Description: "Test package",
			},
		},
		LastFetched: time.Now(),
	}

	cache.SetVersions("test-package", testData, time.Hour)
	retrievedData, found := cache.GetVersions("test-package")
	assert.True(t, found)
	assert.Equal(t, testData, retrievedData)

	// Test expiration
	cache.SetVersions("expired-package", testData, -time.Hour) // Already expired
	_, found = cache.GetVersions("expired-package")
	assert.False(t, found)

	// Test clear
	cache.Clear()
	_, found = cache.GetVersions("test-package")
	assert.False(t, found)
}

func TestAssessCompatibility(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	defer uc.Close()

	// Create mock version data
	versionData := &PackageVersionData{
		Name: "test-package",
		Versions: map[string]PackageVersion{
			"1.0.0": {
				Version: "1.0.0",
				Dependencies: map[string]string{
					"dep1": "^1.0.0",
				},
			},
			"2.0.0": {
				Version: "2.0.0",
				Dependencies: map[string]string{
					"dep1": "^2.0.0",
				},
				Deprecated: "This version is deprecated",
			},
		},
	}

	ctx := context.Background()
	compatibility, err := uc.assessCompatibility(ctx, "test-package", "1.0.0", "2.0.0", versionData)
	require.NoError(t, err)

	// Should detect breaking changes due to major version bump
	assert.Equal(t, "breaking", compatibility.Level)
	assert.Greater(t, compatibility.RiskScore, 0.5)
	assert.NotEmpty(t, compatibility.BreakingChanges)

	// Should detect dependency conflicts
	assert.NotEmpty(t, compatibility.DependencyConflicts)

	// Should detect deprecation
	foundDeprecation := false
	for _, change := range compatibility.BreakingChanges {
		if change.Type == "deprecation" {
			foundDeprecation = true
			break
		}
	}
	assert.True(t, foundDeprecation)
}

func TestCalculateUpdatePriority(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	defer uc.Close()

	current, err := uc.semverParser.Parse("1.0.0")
	require.NoError(t, err)

	tests := []struct {
		name           string
		target         string
		compatibility  UpdateCompatibility
		expectedPrio   string
	}{
		{
			name:   "safe_patch_update",
			target: "1.0.1",
			compatibility: UpdateCompatibility{
				Level:     "safe",
				RiskScore: 0.1,
			},
			expectedPrio: "medium",
		},
		{
			name:   "breaking_major_update",
			target: "2.0.0",
			compatibility: UpdateCompatibility{
				Level:     "breaking",
				RiskScore: 0.8,
			},
			expectedPrio: "low",
		},
		{
			name:   "high_risk_minor",
			target: "1.1.0",
			compatibility: UpdateCompatibility{
				Level:     "minor-risk",
				RiskScore: 0.6,
			},
			expectedPrio: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := uc.semverParser.Parse(tt.target)
			require.NoError(t, err)
			
			priority := uc.calculateUpdatePriority("test-package", current, target, tt.compatibility)
			assert.Equal(t, tt.expectedPrio, priority)
		})
	}
}

func TestEstimateUpdateEffort(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	defer uc.Close()

	tests := []struct {
		name          string
		updateType    string
		compatibility UpdateCompatibility
		expected      string
	}{
		{
			name:       "patch_update",
			updateType: "patch",
			compatibility: UpdateCompatibility{
				RiskScore: 0.1,
			},
			expected: "low",
		},
		{
			name:       "minor_update_low_risk",
			updateType: "minor",
			compatibility: UpdateCompatibility{
				RiskScore: 0.3,
			},
			expected: "low",
		},
		{
			name:       "minor_update_high_risk",
			updateType: "minor",
			compatibility: UpdateCompatibility{
				RiskScore: 0.5,
			},
			expected: "medium",
		},
		{
			name:       "major_update_few_breaks",
			updateType: "major",
			compatibility: UpdateCompatibility{
				BreakingChanges: []BreakingChange{{Type: "semver"}},
			},
			expected: "medium",
		},
		{
			name:       "major_update_many_breaks",
			updateType: "major",
			compatibility: UpdateCompatibility{
				BreakingChanges: []BreakingChange{
					{Type: "api"}, {Type: "dependency"}, {Type: "behavior"},
				},
			},
			expected: "high",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.estimateUpdateEffort(tt.updateType, tt.compatibility)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSuggestTimeline(t *testing.T) {
	uc, err := NewUpdateChecker()
	require.NoError(t, err)
	defer uc.Close()

	tests := []struct {
		name      string
		priority  string
		riskScore float64
		expected  string
	}{
		{
			name:      "critical_priority",
			priority:  "critical",
			riskScore: 0.5,
			expected:  "immediate",
		},
		{
			name:      "high_priority_low_risk",
			priority:  "high",
			riskScore: 0.3,
			expected:  "short-term",
		},
		{
			name:      "high_risk",
			priority:  "medium",
			riskScore: 0.7,
			expected:  "long-term",
		},
		{
			name:      "normal_case",
			priority:  "medium",
			riskScore: 0.4,
			expected:  "short-term",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uc.suggestTimeline(tt.priority, tt.riskScore)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterCriticalUpdates(t *testing.T) {
	updates := []UpdateInfo{
		{
			Current:        "1.0.0",
			Latest:         "1.0.1",
			UpdatePriority: "low",
			Security:       false,
		},
		{
			Current:        "2.0.0",
			Latest:         "2.1.0",
			UpdatePriority: "critical",
			Security:       false,
		},
		{
			Current:        "3.0.0",
			Latest:         "3.0.1",
			UpdatePriority: "medium",
			Security:       true,
		},
	}

	critical := filterCriticalUpdates(updates)
	assert.Len(t, critical, 2)
	
	// Should include the critical priority update
	foundCritical := false
	foundSecurity := false
	for _, update := range critical {
		if update.UpdatePriority == "critical" {
			foundCritical = true
		}
		if update.Security {
			foundSecurity = true
		}
	}
	assert.True(t, foundCritical)
	assert.True(t, foundSecurity)
}