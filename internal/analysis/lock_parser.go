package analysis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// parseNPMLockFile parses npm package-lock.json files
func (da *DependencyAnalyzer) parseNPMLockFile(filePath string) (*LockFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package-lock.json: %w", err)
	}

	var rawData map[string]interface{}
	if err := json.Unmarshal(content, &rawData); err != nil {
		return nil, fmt.Errorf("invalid JSON in package-lock.json: %w", err)
	}

	lockFile := &LockFile{
		Type:         "npm-lock",
		Dependencies: make(map[string]LockEntry),
		Metadata:     make(map[string]interface{}),
	}

	// Parse lock file version
	if version, ok := rawData["lockfileVersion"].(float64); ok {
		lockFile.Version = strconv.Itoa(int(version))
	} else if version, ok := rawData["lockfileVersion"].(string); ok {
		lockFile.Version = version
	}

	// Parse dependencies based on lock file version
	switch lockFile.Version {
	case "1":
		return da.parseNPMLockV1(rawData, lockFile)
	case "2", "3":
		return da.parseNPMLockV2V3(rawData, lockFile)
	default:
		// Attempt to parse as v1 by default
		return da.parseNPMLockV1(rawData, lockFile)
	}
}

// parseNPMLockV1 parses npm lock file version 1
func (da *DependencyAnalyzer) parseNPMLockV1(rawData map[string]interface{}, lockFile *LockFile) (*LockFile, error) {
	// V1 format has flat dependencies structure
	if deps, ok := rawData["dependencies"].(map[string]interface{}); ok {
		for name, depData := range deps {
			if depMap, ok := depData.(map[string]interface{}); ok {
				entry := LockEntry{}

				// Parse version
				if version, ok := depMap["version"].(string); ok {
					entry.Version = version
				}

				// Parse resolved URL
				if resolved, ok := depMap["resolved"].(string); ok {
					entry.Resolved = resolved
				}

				// Parse integrity hash
				if integrity, ok := depMap["integrity"].(string); ok {
					entry.Integrity = integrity
				}

				// Parse dev dependency flag
				if dev, ok := depMap["dev"].(bool); ok {
					entry.DevDep = dev
				}

				// Parse optional flag
				if optional, ok := depMap["optional"].(bool); ok {
					entry.Optional = optional
				}

				// Parse bundled flag
				if bundled, ok := depMap["bundled"].(bool); ok {
					entry.Bundled = bundled
				}

				// Parse requires (transitive dependencies)
				if requires, ok := depMap["requires"].(map[string]interface{}); ok {
					entry.Dependencies = make(map[string]string)
					for reqName, reqVersion := range requires {
						if versionStr, ok := reqVersion.(string); ok {
							entry.Dependencies[reqName] = versionStr
						}
					}
				}

				lockFile.Dependencies[name] = entry
			}
		}
	}

	// Store additional metadata
	for key, value := range rawData {
		if key != "dependencies" && key != "lockfileVersion" {
			lockFile.Metadata[key] = value
		}
	}

	return lockFile, nil
}

// parseNPMLockV2V3 parses npm lock file versions 2 and 3
func (da *DependencyAnalyzer) parseNPMLockV2V3(rawData map[string]interface{}, lockFile *LockFile) (*LockFile, error) {
	// V2/V3 format has packages structure
	if packages, ok := rawData["packages"].(map[string]interface{}); ok {
		for path, pkgData := range packages {
			if pkgMap, ok := pkgData.(map[string]interface{}); ok {
				// Extract package name from path
				name := da.extractPackageNameFromPath(path)
				if name == "" {
					continue // Skip root package entry
				}

				entry := LockEntry{}

				// Parse version
				if version, ok := pkgMap["version"].(string); ok {
					entry.Version = version
				}

				// Parse resolved URL
				if resolved, ok := pkgMap["resolved"].(string); ok {
					entry.Resolved = resolved
				}

				// Parse integrity hash
				if integrity, ok := pkgMap["integrity"].(string); ok {
					entry.Integrity = integrity
				}

				// Parse dev dependency flag
				if dev, ok := pkgMap["dev"].(bool); ok {
					entry.DevDep = dev
				} else if devOptional, ok := pkgMap["devOptional"].(bool); ok {
					entry.DevDep = devOptional
				}

				// Parse optional flag
				if optional, ok := pkgMap["optional"].(bool); ok {
					entry.Optional = optional
				}

				// Parse dependencies
				if deps, ok := pkgMap["dependencies"].(map[string]interface{}); ok {
					entry.Dependencies = make(map[string]string)
					for depName, depVersion := range deps {
						if versionStr, ok := depVersion.(string); ok {
							entry.Dependencies[depName] = versionStr
						}
					}
				}

				lockFile.Dependencies[name] = entry
			}
		}
	}

	// Store additional metadata
	for key, value := range rawData {
		if key != "packages" && key != "lockfileVersion" {
			lockFile.Metadata[key] = value
		}
	}

	return lockFile, nil
}

// extractPackageNameFromPath extracts package name from npm v2/v3 path format
func (da *DependencyAnalyzer) extractPackageNameFromPath(path string) string {
	if path == "" || path == "." {
		return "" // Root package
	}

	// Remove "node_modules/" prefix
	path = strings.TrimPrefix(path, "node_modules/")

	// Handle scoped packages (@scope/package)
	if strings.HasPrefix(path, "@") {
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
	}

	// Handle nested packages (extract first component)
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[0]
	}

	return path
}

// parseYarnLockFile parses Yarn lock files
func (da *DependencyAnalyzer) parseYarnLockFile(filePath string) (*LockFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open yarn.lock: %w", err)
	}
	defer file.Close()

	lockFile := &LockFile{
		Type:         "yarn-lock",
		Dependencies: make(map[string]LockEntry),
		Metadata:     make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(file)
	var currentEntry *LockEntry
	var currentNames []string
	var indentLevel int
	var headerPattern = regexp.MustCompile(`^[^:\s]+.*@.*:$`)
	var versionPattern = regexp.MustCompile(`^\s+version\s+"([^"]+)"`)
	var resolvedPattern = regexp.MustCompile(`^\s+resolved\s+"([^"]+)"`)
	var integrityPattern = regexp.MustCompile(`^\s+integrity\s+(.+)$`)
	var dependenciesPattern = regexp.MustCompile(`^\s+dependencies:`)
	var dependencyPattern = regexp.MustCompile(`^\s+"?([^"\s]+)"?\s+"([^"]+)"`)

	lineNumber := 0
	inDependenciesSection := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++

		// Parse yarn version from header comment
		if lineNumber <= 5 && strings.Contains(line, "yarn lockfile v") {
			if matches := regexp.MustCompile(`yarn lockfile v(\d+)`).FindStringSubmatch(line); len(matches) > 1 {
				lockFile.Version = matches[1]
			}
		}

		// Skip comments and empty lines
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Calculate indent level
		currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Check for package header
		if headerPattern.MatchString(line) {
			// Save previous entry
			if currentEntry != nil && len(currentNames) > 0 {
				for _, name := range currentNames {
					lockFile.Dependencies[name] = *currentEntry
				}
			}

			// Start new entry
			currentEntry = &LockEntry{
				Dependencies: make(map[string]string),
			}
			currentNames = da.parseYarnPackageNames(line)
			indentLevel = currentIndent
			inDependenciesSection = false
			continue
		}

		// Skip if not in a package entry
		if currentEntry == nil {
			continue
		}

		// Parse package properties
		if currentIndent <= indentLevel+2 {
			if matches := versionPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentEntry.Version = matches[1]
			} else if matches := resolvedPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentEntry.Resolved = matches[1]
			} else if matches := integrityPattern.FindStringSubmatch(line); len(matches) > 1 {
				currentEntry.Integrity = matches[1]
			} else if dependenciesPattern.MatchString(line) {
				inDependenciesSection = true
			}
		}

		// Parse dependencies section
		if inDependenciesSection && currentIndent > indentLevel+2 {
			if matches := dependencyPattern.FindStringSubmatch(line); len(matches) > 2 {
				depName := matches[1]
				depVersion := matches[2]
				currentEntry.Dependencies[depName] = depVersion
			}
		}

		// End of dependencies section
		if inDependenciesSection && currentIndent <= indentLevel+2 && !dependenciesPattern.MatchString(line) {
			inDependenciesSection = false
		}
	}

	// Save the last entry
	if currentEntry != nil && len(currentNames) > 0 {
		for _, name := range currentNames {
			lockFile.Dependencies[name] = *currentEntry
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading yarn.lock: %w", err)
	}

	return lockFile, nil
}

// parseYarnPackageNames extracts package names from Yarn lock file header
func (da *DependencyAnalyzer) parseYarnPackageNames(header string) []string {
	// Remove trailing colon and quotes
	header = strings.TrimSuffix(strings.TrimSpace(header), ":")
	header = strings.Trim(header, `"`)

	// Split multiple package specifications
	// Example: "@babel/core@^7.0.0", "@babel/core@^7.1.0"
	var names []string
	parts := strings.Split(header, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)

		// Extract package name before @ symbol (handle scoped packages)
		var name string
		if strings.HasPrefix(part, "@") {
			// Scoped package: @scope/package@version
			atIndex := strings.Index(part[1:], "@")
			if atIndex != -1 {
				name = part[:atIndex+1]
			} else {
				name = part
			}
		} else {
			// Regular package: package@version
			atIndex := strings.Index(part, "@")
			if atIndex != -1 {
				name = part[:atIndex]
			} else {
				name = part
			}
		}

		if name != "" {
			names = append(names, name)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueNames []string
	for _, name := range names {
		if !seen[name] {
			seen[name] = true
			uniqueNames = append(uniqueNames, name)
		}
	}

	return uniqueNames
}
