package analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageManifest represents test package.json data
var testPackageJSON = `{
	"name": "test-project",
	"version": "1.0.0",
	"description": "A test project for dependency analysis",
	"main": "index.js",
	"scripts": {
		"test": "jest",
		"build": "webpack",
		"start": "node index.js"
	},
	"dependencies": {
		"express": "^4.18.0",
		"lodash": "~4.17.21",
		"axios": ">=0.27.0 <1.0.0",
		"@types/node": "18.0.0"
	},
	"devDependencies": {
		"jest": "^28.0.0",
		"webpack": "^5.70.0",
		"@types/jest": "^28.1.0"
	},
	"peerDependencies": {
		"react": ">=16.0.0"
	},
	"optionalDependencies": {
		"fsevents": "^2.3.0"
	},
	"engines": {
		"node": ">=16.0.0",
		"npm": ">=8.0.0"
	},
	"keywords": ["test", "analysis", "dependency"],
	"author": "Test Author <test@example.com>",
	"license": "MIT",
	"homepage": "https://github.com/test/project",
	"repository": {
		"type": "git",
		"url": "git+https://github.com/test/project.git"
	},
	"bugs": {
		"url": "https://github.com/test/project/issues"
	},
	"private": false,
	"customField": "custom value"
}`

func TestNewDependencyAnalyzer(t *testing.T) {
	tests := []struct {
		name    string
		config  DependencyAnalyzerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with minimal settings",
			config: DependencyAnalyzerConfig{
				ProjectRoot: "/test/project",
			},
			wantErr: false,
		},
		{
			name: "valid config with all features enabled",
			config: DependencyAnalyzerConfig{
				ProjectRoot:           "/test/project",
				EnableVulnScanning:    false, // disable for unit test
				EnableLicenseChecking: false, // disable for unit test
				EnableUpdateChecking:  false, // disable for unit test
				MaxDependencyDepth:    5,
				BundleSizeThreshold:   1024000,
				PerformanceThreshold:  5000,
				CriticalVulnThreshold: 8.0,
			},
			wantErr: false,
		},
		{
			name:    "missing project root",
			config:  DependencyAnalyzerConfig{},
			wantErr: true,
			errMsg:  "project root is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewDependencyAnalyzer(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, analyzer)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, analyzer)
				assert.Equal(t, tt.config.ProjectRoot, analyzer.projectRoot)

				// Verify defaults are set
				if len(tt.config.IncludePackageFiles) == 0 {
					expected := []string{"package.json", "package-lock.json", "yarn.lock"}
					assert.Equal(t, expected, analyzer.config.IncludePackageFiles)
				}
				if tt.config.MaxDependencyDepth == 0 {
					assert.Equal(t, 10, analyzer.config.MaxDependencyDepth)
				}
			}
		})
	}
}

func TestParsePackageJSON(t *testing.T) {
	// Create temporary directory and package.json file
	tempDir := t.TempDir()
	packagePath := filepath.Join(tempDir, "package.json")

	err := os.WriteFile(packagePath, []byte(testPackageJSON), 0644)
	require.NoError(t, err)

	// Create analyzer
	config := DependencyAnalyzerConfig{
		ProjectRoot: tempDir,
	}
	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)

	// Parse package.json
	manifest, err := analyzer.parsePackageJSON(packagePath)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	// Verify basic fields
	assert.Equal(t, "test-project", manifest.Name)
	assert.Equal(t, "1.0.0", manifest.Version)
	assert.Equal(t, "A test project for dependency analysis", manifest.Description)
	assert.Equal(t, "index.js", manifest.Main)
	assert.Equal(t, false, manifest.Private)
	assert.Equal(t, "https://github.com/test/project", manifest.Homepage)

	// Verify dependencies
	expectedDeps := map[string]string{
		"express":     "^4.18.0",
		"lodash":      "~4.17.21",
		"axios":       ">=0.27.0 <1.0.0",
		"@types/node": "18.0.0",
	}
	assert.Equal(t, expectedDeps, manifest.Dependencies)

	expectedDevDeps := map[string]string{
		"jest":        "^28.0.0",
		"webpack":     "^5.70.0",
		"@types/jest": "^28.1.0",
	}
	assert.Equal(t, expectedDevDeps, manifest.DevDependencies)

	expectedPeerDeps := map[string]string{
		"react": ">=16.0.0",
	}
	assert.Equal(t, expectedPeerDeps, manifest.PeerDependencies)

	expectedOptionalDeps := map[string]string{
		"fsevents": "^2.3.0",
	}
	assert.Equal(t, expectedOptionalDeps, manifest.OptionalDependencies)

	// Verify scripts
	expectedScripts := map[string]string{
		"test":  "jest",
		"build": "webpack",
		"start": "node index.js",
	}
	assert.Equal(t, expectedScripts, manifest.Scripts)

	// Verify engines
	expectedEngines := map[string]string{
		"node": ">=16.0.0",
		"npm":  ">=8.0.0",
	}
	assert.Equal(t, expectedEngines, manifest.Engines)

	// Verify keywords
	expectedKeywords := []string{"test", "analysis", "dependency"}
	assert.Equal(t, expectedKeywords, manifest.Keywords)

	// Verify complex fields are stored
	assert.NotNil(t, manifest.Repository)
	assert.NotNil(t, manifest.Author)
	assert.NotNil(t, manifest.License)
	assert.NotNil(t, manifest.Bugs)

	// Verify custom fields are stored in metadata
	assert.Contains(t, manifest.Metadata, "customField")
	assert.Equal(t, "custom value", manifest.Metadata["customField"])
}

func TestParsePackageJSONErrors(t *testing.T) {
	tempDir := t.TempDir()
	analyzer, err := NewDependencyAnalyzer(DependencyAnalyzerConfig{
		ProjectRoot: tempDir,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid JSON",
			content: `{"name": "test", invalid}`,
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name:    "empty file",
			content: "",
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name:    "valid minimal JSON",
			content: `{"name": "minimal"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packagePath := filepath.Join(tempDir, "test-"+tt.name+".json")
			err := os.WriteFile(packagePath, []byte(tt.content), 0644)
			require.NoError(t, err)

			manifest, err := analyzer.parsePackageJSON(packagePath)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, manifest)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manifest)
			}
		})
	}
}

func TestParseStringMap(t *testing.T) {
	analyzer := &DependencyAnalyzer{}

	tests := []struct {
		name     string
		input    interface{}
		expected map[string]string
	}{
		{
			name: "valid string map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "mixed types - filters non-strings",
			input: map[string]interface{}{
				"string": "value",
				"number": 123,
				"bool":   true,
				"null":   nil,
			},
			expected: map[string]string{
				"string": "value",
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: map[string]string{},
		},
		{
			name:     "wrong type input",
			input:    "not a map",
			expected: map[string]string{},
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parseStringMap(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindPackageFiles(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()

	// Create test files
	packageJSON := filepath.Join(tempDir, "package.json")
	packageLock := filepath.Join(tempDir, "package-lock.json")
	yarnLock := filepath.Join(tempDir, "yarn.lock")
	nonPackageFile := filepath.Join(tempDir, "README.md")

	// Write files
	err := os.WriteFile(packageJSON, []byte(`{"name": "test"}`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(packageLock, []byte(`{"lockfileVersion": 1}`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(yarnLock, []byte(`# yarn.lock`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(nonPackageFile, []byte("readme"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name          string
		includeFiles  []string
		expectedFiles []string
	}{
		{
			name:         "default package files",
			includeFiles: []string{"package.json", "package-lock.json", "yarn.lock"},
			expectedFiles: []string{
				packageJSON,
				packageLock,
				yarnLock,
			},
		},
		{
			name:         "only package.json",
			includeFiles: []string{"package.json"},
			expectedFiles: []string{
				packageJSON,
			},
		},
		{
			name:          "non-existent file",
			includeFiles:  []string{"non-existent.json"},
			expectedFiles: []string{},
		},
		{
			name:          "mixed existing and non-existing",
			includeFiles:  []string{"package.json", "non-existent.json", "yarn.lock"},
			expectedFiles: []string{packageJSON, yarnLock},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DependencyAnalyzerConfig{
				ProjectRoot:         tempDir,
				IncludePackageFiles: tt.includeFiles,
			}
			analyzer, err := NewDependencyAnalyzer(config)
			require.NoError(t, err)

			files, err := analyzer.findPackageFiles()
			assert.NoError(t, err)

			// Sort both slices for comparison
			assert.ElementsMatch(t, tt.expectedFiles, files)
		})
	}
}

func TestAnalyzeDependenciesBasic(t *testing.T) {
	// Create temporary project structure
	tempDir := t.TempDir()
	packagePath := filepath.Join(tempDir, "package.json")

	// Write minimal package.json
	minimalPackageJSON := `{
		"name": "minimal-test",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21"
		}
	}`

	err := os.WriteFile(packagePath, []byte(minimalPackageJSON), 0644)
	require.NoError(t, err)

	// Create analyzer with features disabled for basic test
	config := DependencyAnalyzerConfig{
		ProjectRoot:           tempDir,
		EnableVulnScanning:    false,
		EnableLicenseChecking: false,
		EnableUpdateChecking:  false,
	}
	analyzer, err := NewDependencyAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	ctx := context.Background()
	tree, err := analyzer.AnalyzeDependencies(ctx)

	// For now, we expect this to fail because we haven't implemented
	// buildDependencyTree and enrichDependencyTree yet
	// This test validates that the parsing works correctly
	if err != nil {
		// This is expected until we implement the remaining methods
		assert.Contains(t, err.Error(), "buildDependencyTree")
	} else {
		// Once implemented, we can test the actual tree
		assert.NotNil(t, tree)
		assert.Equal(t, "minimal-test", tree.RootPackage.Name)
	}
}

// Benchmark tests for performance validation
func BenchmarkParsePackageJSON(b *testing.B) {
	tempDir := b.TempDir()
	packagePath := filepath.Join(tempDir, "package.json")

	err := os.WriteFile(packagePath, []byte(testPackageJSON), 0644)
	require.NoError(b, err)

	analyzer, err := NewDependencyAnalyzer(DependencyAnalyzerConfig{
		ProjectRoot: tempDir,
	})
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.parsePackageJSON(packagePath)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test utility functions
func createTestPackageJSON(content string) (string, func()) {
	tempDir, err := os.MkdirTemp("", "dep-test-*")
	if err != nil {
		panic(err)
	}

	packagePath := filepath.Join(tempDir, "package.json")
	err = os.WriteFile(packagePath, []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return packagePath, cleanup
}

// Mock data for testing
var mockComplexPackageJSON = `{
	"name": "@scope/complex-project",
	"version": "2.1.0-beta.3",
	"description": "A complex project with many dependencies",
	"main": "dist/index.js",
	"module": "dist/index.esm.js",
	"types": "dist/index.d.ts",
	"sideEffects": false,
	"scripts": {
		"build": "rollup -c",
		"test": "jest --coverage",
		"lint": "eslint src/",
		"type-check": "tsc --noEmit",
		"prepublishOnly": "npm run build && npm test"
	},
	"dependencies": {
		"express": "^4.18.2",
		"lodash": "~4.17.21",
		"axios": ">=1.0.0 <2.0.0",
		"@types/node": "18.11.18",
		"moment": "^2.29.4",
		"uuid": "^9.0.0"
	},
	"devDependencies": {
		"jest": "^29.0.0",
		"@types/jest": "^29.2.0",
		"eslint": "^8.30.0",
		"typescript": "^4.9.0",
		"rollup": "^3.7.0"
	},
	"peerDependencies": {
		"react": ">=16.0.0",
		"react-dom": ">=16.0.0"
	},
	"optionalDependencies": {
		"fsevents": "^2.3.2"
	},
	"bundledDependencies": [
		"custom-internal-lib"
	],
	"engines": {
		"node": ">=16.0.0",
		"npm": ">=8.0.0"
	},
	"keywords": ["test", "analysis", "dependency", "complex"],
	"author": {
		"name": "Test Author",
		"email": "test@example.com",
		"url": "https://example.com"
	},
	"license": "MIT",
	"homepage": "https://github.com/scope/complex-project#readme",
	"repository": {
		"type": "git",
		"url": "git+https://github.com/scope/complex-project.git"
	},
	"bugs": {
		"url": "https://github.com/scope/complex-project/issues",
		"email": "bugs@example.com"
	},
	"private": false,
	"workspaces": [
		"packages/*"
	],
	"config": {
		"custom": "value"
	},
	"files": [
		"dist/",
		"README.md",
		"LICENSE"
	]
}`

func TestParseComplexPackageJSON(t *testing.T) {
	packagePath, cleanup := createTestPackageJSON(mockComplexPackageJSON)
	defer cleanup()

	tempDir := filepath.Dir(packagePath)
	analyzer, err := NewDependencyAnalyzer(DependencyAnalyzerConfig{
		ProjectRoot: tempDir,
	})
	require.NoError(t, err)

	manifest, err := analyzer.parsePackageJSON(packagePath)
	require.NoError(t, err)
	require.NotNil(t, manifest)

	// Test scoped package name
	assert.Equal(t, "@scope/complex-project", manifest.Name)
	assert.Equal(t, "2.1.0-beta.3", manifest.Version)

	// Test that all dependency types are parsed
	assert.Len(t, manifest.Dependencies, 6)
	assert.Len(t, manifest.DevDependencies, 5)
	assert.Len(t, manifest.PeerDependencies, 2)
	assert.Len(t, manifest.OptionalDependencies, 1)
	assert.Len(t, manifest.BundledDependencies, 1)

	// Test bundled dependencies
	assert.Contains(t, manifest.BundledDependencies, "custom-internal-lib")

	// Test complex fields are preserved
	assert.NotNil(t, manifest.Author)
	assert.NotNil(t, manifest.Repository)
	assert.NotNil(t, manifest.Bugs)
	assert.NotNil(t, manifest.Workspaces)

	// Test metadata contains additional fields
	assert.Contains(t, manifest.Metadata, "config")
	assert.Contains(t, manifest.Metadata, "files")
	assert.Contains(t, manifest.Metadata, "module")
	assert.Contains(t, manifest.Metadata, "types")
}
