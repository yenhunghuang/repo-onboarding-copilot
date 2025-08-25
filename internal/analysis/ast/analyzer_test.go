package ast

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnalyzer(t *testing.T) {
	config := AnalyzerConfig{
		ProjectRoot:        "/tmp/test",
		MaxConcurrency:     2,
		EnableDependency:   true,
		EnableComponentMap: true,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	require.NotNil(t, analyzer)
	defer analyzer.Close()

	assert.Equal(t, 2, analyzer.config.MaxConcurrency)
	assert.NotNil(t, analyzer.parser)
	assert.NotNil(t, analyzer.dependencyTracker)
	assert.NotNil(t, analyzer.componentMap)
}

func TestAnalyzer_AnalyzeFile(t *testing.T) {
	// Create temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.js")

	testCode := `
import React from 'react';
import { useState } from 'react';
import utils from './utils.js';

function TestComponent() {
    const [count, setCount] = useState(0);
    return React.createElement('div', null, count);
}

export default TestComponent;
export { useState };
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	config := AnalyzerConfig{
		ProjectRoot: tempDir,
		MaxFileSize: 1024 * 1024,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Analyze the file
	result, err := analyzer.AnalyzeFile(context.Background(), testFile)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify parsing results
	assert.Equal(t, testFile, result.FilePath)
	assert.Equal(t, "javascript", result.Language)
	assert.GreaterOrEqual(t, len(result.Functions), 1) // TestComponent function
	assert.GreaterOrEqual(t, len(result.Imports), 3)   // React imports + utils
	assert.GreaterOrEqual(t, len(result.Exports), 2)   // default + named export

	// Verify result is stored
	storedResult, exists := analyzer.GetFileResult(testFile)
	assert.True(t, exists)
	assert.Equal(t, result, storedResult)
}

func TestAnalyzer_AnalyzeRepository(t *testing.T) {
	// Create temporary test repository
	tempDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"src/main.js": `
import React from 'react';
import App from './App.js';
import './styles.css';

function main() {
    return React.createElement(App);
}

export default main;
`,
		"src/App.js": `
import React from 'react';
import { Button } from './components/Button.js';
import { utils } from './utils/index.js';

class App extends React.Component {
    render() {
        return React.createElement(Button);
    }
}

export default App;
`,
		"src/components/Button.js": `
import React from 'react';

function Button({ onClick, children }) {
    return React.createElement('button', { onClick }, children);
}

export { Button };
export default Button;
`,
		"src/utils/index.js": `
export function formatDate(date) {
    return date.toISOString();
}

export const utils = {
    formatDate
};
`,
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Configure analyzer
	config := AnalyzerConfig{
		ProjectRoot:        tempDir,
		MaxConcurrency:     2,
		EnableDependency:   true,
		EnableComponentMap: true,
		MaxFileSize:        1024 * 1024,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Analyze repository
	result, err := analyzer.AnalyzeRepository(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify analysis results
	assert.Equal(t, tempDir, result.ProjectPath)
	assert.Equal(t, 4, len(result.FileResults)) // 4 JavaScript files

	// Verify summary
	assert.Equal(t, 4, result.Summary.TotalFiles)
	assert.Greater(t, result.Summary.TotalFunctions, 0)
	assert.Contains(t, result.Summary.Languages, "javascript")

	// Verify dependency graph
	require.NotNil(t, result.DependencyGraph)
	assert.Greater(t, len(result.DependencyGraph.Nodes), 0)
	assert.Greater(t, len(result.DependencyGraph.Edges), 0)

	// Check for internal and external nodes
	hasInternalNode := false
	hasExternalNode := false
	for _, node := range result.DependencyGraph.Nodes {
		if node.NodeType == "internal" {
			hasInternalNode = true
		}
		if node.NodeType == "external" {
			hasExternalNode = true
		}
	}
	assert.True(t, hasInternalNode, "Should have internal nodes")
	assert.True(t, hasExternalNode, "Should have external nodes (React)")

	// Verify external packages
	assert.Contains(t, result.ExternalPackages, "react")
	reactPkg := result.ExternalPackages["react"]
	assert.Equal(t, "react", reactPkg.Name)
	assert.Greater(t, len(reactPkg.UsedBy), 0)

	// Verify component map
	require.NotNil(t, result.ComponentMap)
	assert.Greater(t, len(result.ComponentMap.Components), 0)
	assert.Greater(t, result.ComponentMap.Stats.TotalComponents, 0)
}

func TestAnalyzer_GetDependencies(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with dependencies
	mainFile := filepath.Join(tempDir, "main.js")
	utilsFile := filepath.Join(tempDir, "utils.js")

	mainContent := `
import { helper } from './utils.js';
import React from 'react';

function main() {
    return helper();
}
`

	utilsContent := `
export function helper() {
    return 'help';
}
`

	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(utilsFile, []byte(utilsContent), 0644)
	require.NoError(t, err)

	config := AnalyzerConfig{
		ProjectRoot:      tempDir,
		EnableDependency: true,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Analyze files
	_, err = analyzer.AnalyzeFile(context.Background(), mainFile)
	require.NoError(t, err)
	_, err = analyzer.AnalyzeFile(context.Background(), utilsFile)
	require.NoError(t, err)

	// Add to dependency tracker
	for path, result := range analyzer.results {
		err = analyzer.dependencyTracker.AddParseResult(path, result)
		require.NoError(t, err)
	}

	// Get dependencies
	deps, exists := analyzer.GetDependencies(mainFile)
	assert.True(t, exists)
	assert.Len(t, deps, 2) // utils.js and react

	// Check utils dependency
	var utilsDep *Dependency
	for i, dep := range deps {
		if dep.TargetModule == "./utils.js" {
			utilsDep = &deps[i]
			break
		}
	}
	require.NotNil(t, utilsDep)
	assert.False(t, utilsDep.IsExternal)
	assert.Equal(t, "named", utilsDep.ImportType)
	assert.Contains(t, utilsDep.ImportedNames, "helper")

	// Check react dependency
	var reactDep *Dependency
	for i, dep := range deps {
		if dep.TargetModule == "react" {
			reactDep = &deps[i]
			break
		}
	}
	require.NotNil(t, reactDep)
	assert.True(t, reactDep.IsExternal)
}

func TestAnalyzer_GetDependents(t *testing.T) {
	tempDir := t.TempDir()

	// Create files where multiple files depend on utils
	files := map[string]string{
		"main.js":  `import { helper } from './utils.js';`,
		"app.js":   `import utils from './utils.js';`,
		"utils.js": `export function helper() {} export default {};`,
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
	}

	config := AnalyzerConfig{
		ProjectRoot:      tempDir,
		EnableDependency: true,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Analyze all files
	for name := range files {
		path := filepath.Join(tempDir, name)
		_, err = analyzer.AnalyzeFile(context.Background(), path)
		require.NoError(t, err)
	}

	// Add to dependency tracker
	for path, result := range analyzer.results {
		err = analyzer.dependencyTracker.AddParseResult(path, result)
		require.NoError(t, err)
	}

	// Get dependents of utils.js
	dependents := analyzer.GetDependents("./utils.js")

	assert.Len(t, dependents, 2) // main.js and app.js depend on utils.js
}

func TestAnalyzer_ComponentMapping(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with component structure
	files := map[string]string{
		"components/Button.js": `
export function Button() {}
export default Button;
`,
		"services/api.js": `
export class ApiService {}
export default ApiService;
`,
		"utils/helper.js": `
export function helper() {}
`,
		"models/User.js": `
export class User {}
`,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	config := AnalyzerConfig{
		ProjectRoot:        tempDir,
		EnableComponentMap: true,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Analyze repository
	result, err := analyzer.AnalyzeRepository(context.Background())
	require.NoError(t, err)

	// Verify component map
	require.NotNil(t, result.ComponentMap)
	assert.Greater(t, len(result.ComponentMap.Components), 0)

	// Check component types
	componentTypes := make(map[string]bool)
	for _, component := range result.ComponentMap.Components {
		componentTypes[component.Type] = true
	}

	// Should have different component types
	assert.True(t, len(componentTypes) > 1, "Should identify different component types")

	// Verify stats
	stats := result.ComponentMap.Stats
	assert.Equal(t, len(result.ComponentMap.Components), stats.TotalComponents)
	assert.Greater(t, len(stats.TypeDistribution), 0)
	assert.Greater(t, len(stats.LayerDistribution), 0)
}

func TestAnalyzer_FileFiltering(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with different patterns
	files := map[string]string{
		"src/main.js":                 `console.log('main');`,
		"src/utils.ts":                `export const utils = {};`,
		"node_modules/react/index.js": `module.exports = {};`,
		"dist/bundle.js":              `var bundle = {};`,
		"test.min.js":                 `var minified = {};`,
		".git/config":                 `[core]`,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(tempDir, filePath)
		dir := filepath.Dir(fullPath)

		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	config := AnalyzerConfig{
		ProjectRoot: tempDir,
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Find supported files
	files_found, err := analyzer.findSupportedFiles()
	require.NoError(t, err)

	// Should only include main.js and utils.ts (excluding node_modules, dist, etc.)
	assert.Len(t, files_found, 2)

	hasMainJs := false
	hasUtilsTs := false
	for _, file := range files_found {
		if filepath.Base(file) == "main.js" {
			hasMainJs = true
		}
		if filepath.Base(file) == "utils.ts" {
			hasUtilsTs = true
		}
	}

	assert.True(t, hasMainJs, "Should include main.js")
	assert.True(t, hasUtilsTs, "Should include utils.ts")
}

func TestAnalyzer_MaxFileSize(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.js")

	// Create file larger than limit
	largeContent := make([]byte, 1024) // 1KB content
	for i := range largeContent {
		largeContent[i] = 'x'
	}

	err := os.WriteFile(testFile, largeContent, 0644)
	require.NoError(t, err)

	config := AnalyzerConfig{
		ProjectRoot: tempDir,
		MaxFileSize: 512, // 512 bytes limit
	}

	analyzer, err := NewAnalyzer(config)
	require.NoError(t, err)
	defer analyzer.Close()

	// Should fail to analyze large file
	_, err = analyzer.AnalyzeFile(context.Background(), testFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum size limit")
}

func BenchmarkAnalyzer_ParseFiles(b *testing.B) {
	tempDir := b.TempDir()

	// Create multiple test files
	for i := 0; i < 10; i++ {
		content := `
import React from 'react';
function Component${i}() {
    return React.createElement('div');
}
export default Component${i};
`
		fileName := filepath.Join(tempDir, "component"+string(rune(i))+"_test.js")
		os.WriteFile(fileName, []byte(content), 0644)
	}

	config := AnalyzerConfig{
		ProjectRoot:    tempDir,
		MaxConcurrency: 4,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		analyzer, _ := NewAnalyzer(config)
		analyzer.AnalyzeRepository(context.Background())
		analyzer.Close()
	}
}
