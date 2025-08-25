// Package analysis provides comprehensive update analysis for npm packages
package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// UpdateChecker analyzes package versions and provides update recommendations
type UpdateChecker struct {
	client        *http.Client
	registryURL   string
	cache         UpdateCache
	semverParser  *SemanticVersionParser
}

// UpdateCache provides caching for package version data
type UpdateCache interface {
	GetVersions(packageName string) (*PackageVersionData, bool)
	SetVersions(packageName string, data *PackageVersionData, ttl time.Duration)
	Clear()
}

// PackageVersionData represents cached version information from npm registry
type PackageVersionData struct {
	Name         string                    `json:"name"`
	DistTags     map[string]string         `json:"dist-tags"`
	Versions     map[string]PackageVersion `json:"versions"`
	Time         map[string]string         `json:"time"`
	LastFetched  time.Time                 `json:"last_fetched"`
}

// PackageVersion represents detailed version information
type PackageVersion struct {
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Dependencies map[string]string      `json:"dependencies"`
	DevDependencies map[string]string   `json:"devDependencies"`
	PeerDependencies map[string]string  `json:"peerDependencies"`
	Engines      map[string]string      `json:"engines"`
	Scripts      map[string]string      `json:"scripts"`
	Keywords     []string               `json:"keywords"`
	License      interface{}            `json:"license"`
	Repository   interface{}            `json:"repository"`
	Homepage     string                 `json:"homepage"`
	Bugs         interface{}            `json:"bugs"`
	Author       interface{}            `json:"author"`
	Maintainers  []interface{}          `json:"maintainers"`
	Dist         PackageDistribution    `json:"dist"`
	PublishTime  time.Time              `json:"publish_time"`
	Deprecated   string                 `json:"deprecated,omitempty"`
}

// PackageDistribution represents package distribution information
type PackageDistribution struct {
	Tarball    string `json:"tarball"`
	Shasum     string `json:"shasum"`
	Integrity  string `json:"integrity"`
	FileCount  int    `json:"fileCount"`
	UnpackedSize int64 `json:"unpackedSize"`
}

// SemanticVersionParser handles semantic version parsing and comparison
type SemanticVersionParser struct {
	versionRegex    *regexp.Regexp
	prereleaseRegex *regexp.Regexp
	buildRegex      *regexp.Regexp
}

// SemanticVersion represents a parsed semantic version
type SemanticVersion struct {
	Major      int      `json:"major"`
	Minor      int      `json:"minor"`
	Patch      int      `json:"patch"`
	Prerelease string   `json:"prerelease"`
	Build      string   `json:"build"`
	Raw        string   `json:"raw"`
	Tags       []string `json:"tags"`
}

// UpdateCompatibility represents compatibility assessment for updates
type UpdateCompatibility struct {
	Level              string                 `json:"level"`                // safe, minor-risk, major-risk, breaking
	BreakingChanges    []BreakingChange       `json:"breaking_changes"`
	DependencyConflicts []DependencyConflict  `json:"dependency_conflicts"`
	PeerConflicts      []PeerDependencyIssue `json:"peer_conflicts"`
	EngineCompatibility EngineCompatibility   `json:"engine_compatibility"`
	RiskScore          float64               `json:"risk_score"` // 0-1 scale
	Recommendations    []string              `json:"recommendations"`
}

// BreakingChange represents detected breaking changes
type BreakingChange struct {
	Type        string `json:"type"`        // api, dependency, behavior, deprecation
	Description string `json:"description"`
	Severity    string `json:"severity"`    // low, medium, high, critical
	Source      string `json:"source"`      // changelog, semver, analysis
	Mitigation  string `json:"mitigation"`
}

// DependencyConflict represents conflicts with other dependencies
type DependencyConflict struct {
	Package      string `json:"package"`
	CurrentRange string `json:"current_range"`
	RequiredRange string `json:"required_range"`
	ConflictType string `json:"conflict_type"` // version, peer, engine
	Resolution   string `json:"resolution"`
}

// PeerDependencyIssue represents peer dependency compatibility issues
type PeerDependencyIssue struct {
	PeerPackage    string `json:"peer_package"`
	RequiredRange  string `json:"required_range"`
	CurrentVersion string `json:"current_version"`
	Satisfied      bool   `json:"satisfied"`
	Resolution     string `json:"resolution"`
}

// EngineCompatibility represents Node.js/npm engine compatibility
type EngineCompatibility struct {
	NodeCompatible bool   `json:"node_compatible"`
	NPMCompatible  bool   `json:"npm_compatible"`
	NodeCurrent    string `json:"node_current"`
	NodeRequired   string `json:"node_required"`
	NPMCurrent     string `json:"npm_current"`
	NPMRequired    string `json:"npm_required"`
	Issues         []string `json:"issues"`
}

// UpdateRecommendation represents a comprehensive update recommendation
type UpdateRecommendation struct {
	Package           string               `json:"package"`
	CurrentVersion    string               `json:"current_version"`
	RecommendedVersion string              `json:"recommended_version"`
	LatestVersion     string               `json:"latest_version"`
	UpdateType        string               `json:"update_type"`     // patch, minor, major
	Priority          string               `json:"priority"`        // low, medium, high, critical
	SecurityUpdate    bool                 `json:"security_update"`
	Compatibility     UpdateCompatibility  `json:"compatibility"`
	Benefits          []string             `json:"benefits"`
	Risks             []string             `json:"risks"`
	EstimatedEffort   string               `json:"estimated_effort"` // low, medium, high
	Timeline          string               `json:"timeline"`         // immediate, short-term, long-term
}

// MemoryUpdateCache provides in-memory caching for update data
type MemoryUpdateCache struct {
	cache map[string]updateCacheEntry
}

type updateCacheEntry struct {
	data   *PackageVersionData
	expiry time.Time
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker() (*UpdateChecker, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	cache := &MemoryUpdateCache{
		cache: make(map[string]updateCacheEntry),
	}

	parser := &SemanticVersionParser{
		versionRegex:    regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`),
		prereleaseRegex: regexp.MustCompile(`^([0-9A-Za-z\-\.]+)$`),
		buildRegex:      regexp.MustCompile(`^([0-9A-Za-z\-\.]+)$`),
	}

	uc := &UpdateChecker{
		client:       client,
		registryURL:  "https://registry.npmjs.org",
		cache:        cache,
		semverParser: parser,
	}

	return uc, nil
}

// CheckUpdates analyzes packages for available updates
func (uc *UpdateChecker) CheckUpdates(ctx context.Context, packages []string) (*UpdateReport, error) {
	report := &UpdateReport{
		Updates:           []UpdateInfo{},
		UpdatesByType:     make(map[string]int),
		UpdatesByPriority: make(map[string]int),
		Recommendations:   []string{},
	}

	var allRecommendations []UpdateRecommendation
	processed := make(map[string]bool)

	for _, pkg := range packages {
		if processed[pkg] {
			continue
		}
		processed[pkg] = true

		name, version, err := parsePackageSpec(pkg)
		if err != nil {
			continue
		}

		// Get package version data
		versionData, err := uc.getPackageVersionData(ctx, name)
		if err != nil {
			continue
		}

		// Analyze updates for this package
		recommendation, err := uc.analyzePackageUpdates(ctx, name, version, versionData)
		if err != nil {
			continue
		}

		if recommendation != nil {
			allRecommendations = append(allRecommendations, *recommendation)

			// Convert to UpdateInfo for compatibility
			updateInfo := uc.convertToUpdateInfo(*recommendation)
			report.Updates = append(report.Updates, updateInfo)

			// Update statistics
			report.UpdatesByType[updateInfo.Type]++
			report.UpdatesByPriority[updateInfo.UpdatePriority]++

			if updateInfo.Security {
				report.SecurityUpdates++
			}
			if updateInfo.Breaking {
				report.BreakingUpdates++
			}
		}
	}

	// Update totals
	report.TotalPackages = len(processed)
	report.OutdatedPackages = len(report.Updates)

	// Generate recommendations
	report.Recommendations = uc.generateUpdateRecommendations(allRecommendations)
	report.UpdateStrategy = uc.determineUpdateStrategy(allRecommendations)

	return report, nil
}

// getPackageVersionData retrieves version data from npm registry with caching
func (uc *UpdateChecker) getPackageVersionData(ctx context.Context, packageName string) (*PackageVersionData, error) {
	// Check cache first
	if data, found := uc.cache.GetVersions(packageName); found {
		return data, nil
	}

	// Fetch from registry
	url := fmt.Sprintf("%s/%s", uc.registryURL, packageName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "repo-onboarding-copilot/1.0")

	resp, err := uc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	var registryData map[string]interface{}
	if err := json.Unmarshal(body, &registryData); err != nil {
		return nil, fmt.Errorf("failed to parse registry response: %w", err)
	}

	// Parse registry data
	versionData := uc.parseRegistryData(registryData)
	versionData.LastFetched = time.Now()

	// Cache the data
	uc.cache.SetVersions(packageName, versionData, 4*time.Hour)

	return versionData, nil
}

// analyzePackageUpdates analyzes available updates for a specific package
func (uc *UpdateChecker) analyzePackageUpdates(ctx context.Context, packageName, currentVersion string, versionData *PackageVersionData) (*UpdateRecommendation, error) {
	// Parse current version
	current, err := uc.semverParser.Parse(currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	// Find latest version
	latestVersion := versionData.DistTags["latest"]
	if latestVersion == "" {
		return nil, fmt.Errorf("no latest version found")
	}

	latest, err := uc.semverParser.Parse(latestVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latest version: %w", err)
	}

	// Check if update is needed
	if uc.semverParser.Compare(current, latest) >= 0 {
		return nil, nil // Already up to date
	}

	// Determine update type
	updateType := uc.determineUpdateType(current, latest)

	// Find best update version based on constraints
	recommendedVersion := uc.findRecommendedVersion(current, versionData, updateType)

	// Assess compatibility
	compatibility, err := uc.assessCompatibility(ctx, packageName, currentVersion, recommendedVersion, versionData)
	if err != nil {
		compatibility = UpdateCompatibility{Level: "unknown", RiskScore: 0.5}
	}

	// Parse recommended version for comparison
	recommended, err := uc.semverParser.Parse(recommendedVersion)
	if err != nil {
		recommended = latest // fallback to latest if parsing fails
	}

	// Calculate priority
	priority := uc.calculateUpdatePriority(packageName, current, recommended, compatibility)

	// Check for security updates
	securityUpdate := uc.isSecurityUpdate(packageName, currentVersion, recommendedVersion)

	recommendation := &UpdateRecommendation{
		Package:            packageName,
		CurrentVersion:     currentVersion,
		RecommendedVersion: recommendedVersion,
		LatestVersion:      latestVersion,
		UpdateType:         updateType,
		Priority:           priority,
		SecurityUpdate:     securityUpdate,
		Compatibility:      compatibility,
		Benefits:           uc.identifyBenefits(current, recommended, versionData),
		Risks:              uc.identifyRisks(compatibility),
		EstimatedEffort:    uc.estimateUpdateEffort(updateType, compatibility),
		Timeline:           uc.suggestTimeline(priority, compatibility.RiskScore),
	}

	return recommendation, nil
}

// Close releases update checker resources
func (uc *UpdateChecker) Close() error {
	uc.cache.Clear()
	return nil
}

// Helper methods will be implemented in the next file to keep this manageable
// These include semver parsing, compatibility assessment, and recommendation generation