// Package analysis provides update analysis helper functions
package analysis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Semantic version parsing and comparison methods

// Parse parses a semantic version string
func (svp *SemanticVersionParser) Parse(version string) (*SemanticVersion, error) {
	// Clean version string
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("empty version string")
	}

	matches := svp.versionRegex.FindStringSubmatch(version)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid semantic version format: %s", version)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	sv := &SemanticVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Raw:   version,
	}

	// Parse prerelease
	if len(matches) > 4 && matches[4] != "" {
		sv.Prerelease = matches[4]
	}

	// Parse build metadata
	if len(matches) > 5 && matches[5] != "" {
		sv.Build = matches[5]
	}

	return sv, nil
}

// Compare compares two semantic versions (-1, 0, 1)
func (svp *SemanticVersionParser) Compare(v1, v2 *SemanticVersion) int {
	// Compare major version
	if v1.Major < v2.Major {
		return -1
	} else if v1.Major > v2.Major {
		return 1
	}

	// Compare minor version
	if v1.Minor < v2.Minor {
		return -1
	} else if v1.Minor > v2.Minor {
		return 1
	}

	// Compare patch version
	if v1.Patch < v2.Patch {
		return -1
	} else if v1.Patch > v2.Patch {
		return 1
	}

	// Compare prerelease
	return svp.comparePrerelease(v1.Prerelease, v2.Prerelease)
}

// comparePrerelease compares prerelease versions
func (svp *SemanticVersionParser) comparePrerelease(pre1, pre2 string) int {
	// No prerelease > has prerelease
	if pre1 == "" && pre2 != "" {
		return 1
	}
	if pre1 != "" && pre2 == "" {
		return -1
	}
	if pre1 == "" && pre2 == "" {
		return 0
	}

	// Compare prerelease versions lexically (simplified)
	if pre1 < pre2 {
		return -1
	} else if pre1 > pre2 {
		return 1
	}
	return 0
}

// Update analysis helper methods

// parseRegistryData parses npm registry response data
func (uc *UpdateChecker) parseRegistryData(data map[string]interface{}) *PackageVersionData {
	versionData := &PackageVersionData{
		DistTags: make(map[string]string),
		Versions: make(map[string]PackageVersion),
		Time:     make(map[string]string),
	}

	// Parse name
	if name, ok := data["name"].(string); ok {
		versionData.Name = name
	}

	// Parse dist-tags
	if distTags, ok := data["dist-tags"].(map[string]interface{}); ok {
		for tag, version := range distTags {
			if versionStr, ok := version.(string); ok {
				versionData.DistTags[tag] = versionStr
			}
		}
	}

	// Parse versions
	if versions, ok := data["versions"].(map[string]interface{}); ok {
		for version, versionInfo := range versions {
			if versionMap, ok := versionInfo.(map[string]interface{}); ok {
				packageVersion := uc.parsePackageVersion(version, versionMap)
				versionData.Versions[version] = packageVersion
			}
		}
	}

	// Parse time
	if timeData, ok := data["time"].(map[string]interface{}); ok {
		for version, timestamp := range timeData {
			if timeStr, ok := timestamp.(string); ok {
				versionData.Time[version] = timeStr
			}
		}
	}

	return versionData
}

// parsePackageVersion parses individual package version data
func (uc *UpdateChecker) parsePackageVersion(version string, data map[string]interface{}) PackageVersion {
	pv := PackageVersion{
		Version: version,
	}

	// Parse basic fields
	if desc, ok := data["description"].(string); ok {
		pv.Description = desc
	}
	if homepage, ok := data["homepage"].(string); ok {
		pv.Homepage = homepage
	}
	if deprecated, ok := data["deprecated"].(string); ok {
		pv.Deprecated = deprecated
	}

	// Parse dependencies
	pv.Dependencies = uc.parseStringMap(data["dependencies"])
	pv.DevDependencies = uc.parseStringMap(data["devDependencies"])
	pv.PeerDependencies = uc.parseStringMap(data["peerDependencies"])
	pv.Engines = uc.parseStringMap(data["engines"])
	pv.Scripts = uc.parseStringMap(data["scripts"])

	// Parse keywords
	if keywords, ok := data["keywords"].([]interface{}); ok {
		for _, keyword := range keywords {
			if keywordStr, ok := keyword.(string); ok {
				pv.Keywords = append(pv.Keywords, keywordStr)
			}
		}
	}

	// Parse distribution info
	if dist, ok := data["dist"].(map[string]interface{}); ok {
		pv.Dist = uc.parseDistribution(dist)
	}

	// Store complex fields as-is
	pv.License = data["license"]
	pv.Repository = data["repository"]
	pv.Bugs = data["bugs"]
	pv.Author = data["author"]

	if maintainers, ok := data["maintainers"].([]interface{}); ok {
		pv.Maintainers = maintainers
	}

	return pv
}

// parseStringMap safely parses map[string]interface{} to map[string]string
func (uc *UpdateChecker) parseStringMap(data interface{}) map[string]string {
	result := make(map[string]string)
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			if str, ok := value.(string); ok {
				result[key] = str
			}
		}
	}
	return result
}

// parseDistribution parses package distribution information
func (uc *UpdateChecker) parseDistribution(data map[string]interface{}) PackageDistribution {
	dist := PackageDistribution{}

	if tarball, ok := data["tarball"].(string); ok {
		dist.Tarball = tarball
	}
	if shasum, ok := data["shasum"].(string); ok {
		dist.Shasum = shasum
	}
	if integrity, ok := data["integrity"].(string); ok {
		dist.Integrity = integrity
	}
	if fileCount, ok := data["fileCount"].(float64); ok {
		dist.FileCount = int(fileCount)
	}
	if unpackedSize, ok := data["unpackedSize"].(float64); ok {
		dist.UnpackedSize = int64(unpackedSize)
	}

	return dist
}

// determineUpdateType determines the type of update (patch, minor, major)
func (uc *UpdateChecker) determineUpdateType(current, target *SemanticVersion) string {
	if current.Major != target.Major {
		return "major"
	}
	if current.Minor != target.Minor {
		return "minor"
	}
	if current.Patch != target.Patch {
		return "patch"
	}
	return "prerelease"
}

// findRecommendedVersion finds the best version to update to
func (uc *UpdateChecker) findRecommendedVersion(current *SemanticVersion, versionData *PackageVersionData, updateType string) string {
	var candidates []string

	// Collect candidate versions based on update type preference
	for version := range versionData.Versions {
		candidate, err := uc.semverParser.Parse(version)
		if err != nil {
			continue
		}

		// Only consider versions newer than current
		if uc.semverParser.Compare(candidate, current) <= 0 {
			continue
		}

		// Skip prerelease versions unless specifically looking for them
		if candidate.Prerelease != "" && updateType != "prerelease" {
			continue
		}

		// Filter based on update type preference
		switch updateType {
		case "patch":
			if candidate.Major == current.Major && candidate.Minor == current.Minor {
				candidates = append(candidates, version)
			}
		case "minor":
			if candidate.Major == current.Major {
				candidates = append(candidates, version)
			}
		case "major":
			candidates = append(candidates, version)
		}
	}

	if len(candidates) == 0 {
		// Fallback to latest if no constrained version found
		return versionData.DistTags["latest"]
	}

	// Sort candidates and return the latest compatible version
	sort.Slice(candidates, func(i, j int) bool {
		v1, _ := uc.semverParser.Parse(candidates[i])
		v2, _ := uc.semverParser.Parse(candidates[j])
		return uc.semverParser.Compare(v1, v2) > 0
	})

	return candidates[0]
}

// assessCompatibility assesses update compatibility and risks
func (uc *UpdateChecker) assessCompatibility(ctx context.Context, packageName, currentVersion, targetVersion string, versionData *PackageVersionData) (UpdateCompatibility, error) {
	compatibility := UpdateCompatibility{
		BreakingChanges:     []BreakingChange{},
		DependencyConflicts: []DependencyConflict{},
		PeerConflicts:       []PeerDependencyIssue{},
		Recommendations:     []string{},
	}

	current, err := uc.semverParser.Parse(currentVersion)
	if err != nil {
		return compatibility, err
	}

	target, err := uc.semverParser.Parse(targetVersion)
	if err != nil {
		return compatibility, err
	}

	// Determine compatibility level based on version difference
	if current.Major != target.Major {
		compatibility.Level = "breaking"
		compatibility.RiskScore = 0.8
		compatibility.BreakingChanges = append(compatibility.BreakingChanges, BreakingChange{
			Type:        "semver",
			Description: "Major version change indicates breaking changes",
			Severity:    "high",
			Source:      "semver",
			Mitigation:  "Review changelog and test thoroughly before updating",
		})
	} else if current.Minor != target.Minor {
		compatibility.Level = "minor-risk"
		compatibility.RiskScore = 0.3
	} else {
		compatibility.Level = "safe"
		compatibility.RiskScore = 0.1
	}

	// Check for deprecated versions
	if targetVersionData, exists := versionData.Versions[targetVersion]; exists {
		if targetVersionData.Deprecated != "" {
			compatibility.BreakingChanges = append(compatibility.BreakingChanges, BreakingChange{
				Type:        "deprecation",
				Description: fmt.Sprintf("Version %s is deprecated: %s", targetVersion, targetVersionData.Deprecated),
				Severity:    "medium",
				Source:      "registry",
				Mitigation:  "Consider updating to a non-deprecated version",
			})
			compatibility.RiskScore += 0.2
		}
	}

	// Analyze dependency changes
	uc.analyzeDependencyChanges(&compatibility, versionData, currentVersion, targetVersion)

	// Generate recommendations based on compatibility
	uc.generateCompatibilityRecommendations(&compatibility)

	return compatibility, nil
}

// analyzeDependencyChanges analyzes changes in dependencies between versions
func (uc *UpdateChecker) analyzeDependencyChanges(compatibility *UpdateCompatibility, versionData *PackageVersionData, currentVersion, targetVersion string) {
	currentVersionData, currentExists := versionData.Versions[currentVersion]
	targetVersionData, targetExists := versionData.Versions[targetVersion]

	if !currentExists || !targetExists {
		return
	}

	// Check for dependency changes
	for depName, newRange := range targetVersionData.Dependencies {
		if oldRange, exists := currentVersionData.Dependencies[depName]; exists {
			if oldRange != newRange {
				compatibility.DependencyConflicts = append(compatibility.DependencyConflicts, DependencyConflict{
					Package:       depName,
					CurrentRange:  oldRange,
					RequiredRange: newRange,
					ConflictType:  "version",
					Resolution:    "Review dependency compatibility",
				})
			}
		}
	}

	// Check for peer dependency changes
	for peerName, newRange := range targetVersionData.PeerDependencies {
		if oldRange, exists := currentVersionData.PeerDependencies[peerName]; exists {
			if oldRange != newRange {
				compatibility.PeerConflicts = append(compatibility.PeerConflicts, PeerDependencyIssue{
					PeerPackage:   peerName,
					RequiredRange: newRange,
					Satisfied:     false, // Would need to check actual installed version
					Resolution:    "Update peer dependency if needed",
				})
			}
		}
	}
}

// generateCompatibilityRecommendations generates recommendations based on compatibility analysis
func (uc *UpdateChecker) generateCompatibilityRecommendations(compatibility *UpdateCompatibility) {
	if compatibility.Level == "breaking" {
		compatibility.Recommendations = append(compatibility.Recommendations,
			"Major version update detected - review changelog for breaking changes",
			"Test thoroughly in development environment before deploying",
			"Consider gradual rollout for production updates")
	}

	if len(compatibility.DependencyConflicts) > 0 {
		compatibility.Recommendations = append(compatibility.Recommendations,
			"Review dependency changes and update related packages if needed")
	}

	if len(compatibility.PeerConflicts) > 0 {
		compatibility.Recommendations = append(compatibility.Recommendations,
			"Check peer dependency requirements and update as necessary")
	}
}

// calculateUpdatePriority calculates update priority based on various factors
func (uc *UpdateChecker) calculateUpdatePriority(packageName string, current, target *SemanticVersion, compatibility UpdateCompatibility) string {
	score := 0.0

	// Security updates get high priority
	if uc.isSecurityUpdate(packageName, current.Raw, target.Raw) {
		score += 0.4
	}

	// Major versions get lower priority due to risk
	if current.Major != target.Major {
		score -= 0.2
	}

	// Recent releases get slight priority boost
	// This would require checking release date from registry data
	// score += uc.calculateRecencyBoost(target)

	// Risk factor
	score -= compatibility.RiskScore * 0.3

	// Convert score to priority category
	if score >= 0.7 {
		return "critical"
	} else if score >= 0.5 {
		return "high"
	} else if score >= 0.3 {
		return "medium"
	} else {
		return "low"
	}
}

// isSecurityUpdate checks if an update contains security fixes
func (uc *UpdateChecker) isSecurityUpdate(packageName, currentVersion, targetVersion string) bool {
	// This would integrate with vulnerability database to check if the update fixes known vulnerabilities
	// For now, simplified implementation
	return false
}

// identifyBenefits identifies benefits of updating to a new version
func (uc *UpdateChecker) identifyBenefits(current, target *SemanticVersion, versionData *PackageVersionData) []string {
	var benefits []string

	if current.Major != target.Major {
		benefits = append(benefits, "Major feature updates and improvements")
		benefits = append(benefits, "Latest API enhancements")
	} else if current.Minor != target.Minor {
		benefits = append(benefits, "New features and functionality")
		benefits = append(benefits, "Performance improvements")
	} else {
		benefits = append(benefits, "Bug fixes and stability improvements")
	}

	// Check if target version fixes deprecation warnings
	if targetVersionData, exists := versionData.Versions[target.Raw]; exists {
		if targetVersionData.Deprecated == "" {
			benefits = append(benefits, "Resolves deprecated version issues")
		}
	}

	return benefits
}

// identifyRisks identifies risks associated with the update
func (uc *UpdateChecker) identifyRisks(compatibility UpdateCompatibility) []string {
	var risks []string

	if compatibility.Level == "breaking" {
		risks = append(risks, "Potential breaking changes requiring code updates")
	}

	if len(compatibility.DependencyConflicts) > 0 {
		risks = append(risks, "Dependency version conflicts may occur")
	}

	if len(compatibility.PeerConflicts) > 0 {
		risks = append(risks, "Peer dependency issues may need resolution")
	}

	if compatibility.RiskScore > 0.5 {
		risks = append(risks, "Moderate to high risk of introducing issues")
	}

	return risks
}

// estimateUpdateEffort estimates the effort required for the update
func (uc *UpdateChecker) estimateUpdateEffort(updateType string, compatibility UpdateCompatibility) string {
	switch updateType {
	case "patch":
		return "low"
	case "minor":
		if compatibility.RiskScore > 0.4 {
			return "medium"
		}
		return "low"
	case "major":
		if len(compatibility.BreakingChanges) > 2 {
			return "high"
		}
		return "medium"
	default:
		return "medium"
	}
}

// suggestTimeline suggests when the update should be applied
func (uc *UpdateChecker) suggestTimeline(priority string, riskScore float64) string {
	if priority == "critical" {
		return "immediate"
	}
	if priority == "high" && riskScore < 0.5 {
		return "short-term"
	}
	if riskScore > 0.6 {
		return "long-term"
	}
	return "short-term"
}

// convertToUpdateInfo converts UpdateRecommendation to UpdateInfo for compatibility
func (uc *UpdateChecker) convertToUpdateInfo(rec UpdateRecommendation) UpdateInfo {
	return UpdateInfo{
		Current:        rec.CurrentVersion,
		Latest:         rec.LatestVersion,
		Wanted:         rec.RecommendedVersion,
		Type:           rec.UpdateType,
		Breaking:       rec.Compatibility.Level == "breaking",
		Security:       rec.SecurityUpdate,
		Deprecated:     false, // Would check deprecated status
		UpdatePriority: rec.Priority,
		ChangelogURL:   "", // Would extract from registry data
	}
}

// generateUpdateRecommendations generates overall recommendations
func (uc *UpdateChecker) generateUpdateRecommendations(recommendations []UpdateRecommendation) []string {
	var advice []string

	criticalCount := 0
	securityCount := 0
	majorCount := 0

	for _, rec := range recommendations {
		if rec.Priority == "critical" {
			criticalCount++
		}
		if rec.SecurityUpdate {
			securityCount++
		}
		if rec.UpdateType == "major" {
			majorCount++
		}
	}

	if criticalCount > 0 {
		advice = append(advice, fmt.Sprintf("URGENT: %d critical updates require immediate attention", criticalCount))
	}

	if securityCount > 0 {
		advice = append(advice, fmt.Sprintf("SECURITY: %d updates contain security fixes", securityCount))
	}

	if majorCount > 0 {
		advice = append(advice, fmt.Sprintf("BREAKING: %d major version updates may require code changes", majorCount))
	}

	if len(recommendations) > 10 {
		advice = append(advice, "Consider implementing automated dependency updates for better maintenance")
	}

	return advice
}

// determineUpdateStrategy determines the overall update strategy
func (uc *UpdateChecker) determineUpdateStrategy(recommendations []UpdateRecommendation) string {
	highRiskCount := 0
	securityCount := 0

	for _, rec := range recommendations {
		if rec.Compatibility.RiskScore > 0.6 {
			highRiskCount++
		}
		if rec.SecurityUpdate {
			securityCount++
		}
	}

	if securityCount > 0 {
		return "security-first"
	}
	if highRiskCount > len(recommendations)/2 {
		return "conservative"
	}
	return "balanced"
}

// Memory cache implementation

func (muc *MemoryUpdateCache) GetVersions(packageName string) (*PackageVersionData, bool) {
	entry, exists := muc.cache[packageName]
	if !exists || time.Now().After(entry.expiry) {
		delete(muc.cache, packageName)
		return nil, false
	}
	return entry.data, true
}

func (muc *MemoryUpdateCache) SetVersions(packageName string, data *PackageVersionData, ttl time.Duration) {
	muc.cache[packageName] = updateCacheEntry{
		data:   data,
		expiry: time.Now().Add(ttl),
	}
}

func (muc *MemoryUpdateCache) Clear() {
	muc.cache = make(map[string]updateCacheEntry)
}
