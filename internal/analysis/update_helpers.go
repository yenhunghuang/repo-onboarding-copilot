// Package analysis provides helper functions for update analysis
package analysis

// extractPackageNameFromUpdate extracts package name from UpdateInfo
// Since UpdateInfo doesn't store package name, we need a different approach
func extractPackageNameFromUpdate(updateInfo UpdateInfo) string {
	// This is a limitation - UpdateInfo should ideally contain package name
	// For now, return empty string and handle this gracefully
	return ""
}

// filterCriticalUpdates filters updates that are critical priority or security updates
func filterCriticalUpdates(updates []UpdateInfo) []UpdateInfo {
	var critical []UpdateInfo
	for _, update := range updates {
		if update.UpdatePriority == "critical" || update.Security {
			critical = append(critical, update)
		}
	}
	return critical
}