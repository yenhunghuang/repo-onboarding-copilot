package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyTracker(t *testing.T) {
	tracker := NewDependencyTracker()

	assert.NotNil(t, tracker)
	assert.NotNil(t, tracker.fileResults)
	assert.NotNil(t, tracker.dependencies)
	assert.NotNil(t, tracker.reverseDeps)
	assert.NotNil(t, tracker.externalDeps)
	assert.NotNil(t, tracker.moduleGraph)
}

func TestDependencyTracker_AddParseResult(t *testing.T) {
	tracker := NewDependencyTracker()

	// Create mock parse result
	result := &ParseResult{
		FilePath: "src/utils.js",
		Language: "javascript",
		Imports: []ImportInfo{
			{
				Source:     "lodash",
				ImportType: "named",
				Specifiers: []string{"map", "filter"},
				StartLine:  1,
				IsExternal: true,
			},
			{
				Source:     "./helper.js",
				ImportType: "default",
				LocalName:  "helper",
				StartLine:  2,
				IsExternal: false,
			},
		},
		Exports: []ExportInfo{
			{
				ExportType: "named",
				Specifiers: []string{"processData"},
				StartLine:  10,
			},
		},
	}

	err := tracker.AddParseResult("src/utils.js", result)
	require.NoError(t, err)

	// Verify result was stored
	assert.Contains(t, tracker.fileResults, "src/utils.js")

	// Verify dependencies were extracted
	deps, exists := tracker.GetDependencies("src/utils.js")
	assert.True(t, exists)
	assert.Len(t, deps, 2)

	// Check lodash dependency
	lodashDep := deps[0]
	assert.Equal(t, "lodash", lodashDep.TargetModule)
	assert.Equal(t, "named", lodashDep.ImportType)
	assert.True(t, lodashDep.IsExternal)
	assert.Contains(t, lodashDep.ImportedNames, "map")
	assert.Contains(t, lodashDep.ImportedNames, "filter")

	// Check local dependency
	helperDep := deps[1]
	assert.Equal(t, "./helper.js", helperDep.TargetModule)
	assert.Equal(t, "default", helperDep.ImportType)
	assert.False(t, helperDep.IsExternal)
	assert.Equal(t, "helper", helperDep.LocalName)

	// Verify external package tracking
	externals := tracker.GetExternalPackages()
	assert.Contains(t, externals, "lodash")

	lodashPkg := externals["lodash"]
	assert.Equal(t, "lodash", lodashPkg.Name)
	assert.Contains(t, lodashPkg.UsedBy, "src/utils.js")
	assert.Contains(t, lodashPkg.ImportedFeatures, "map")
	assert.Contains(t, lodashPkg.ImportedFeatures, "filter")
}

func TestDependencyTracker_ResolveDependencies(t *testing.T) {
	tracker := NewDependencyTracker()

	// Add helper.js file
	helperResult := &ParseResult{
		FilePath:  "src/helper.js",
		Language:  "javascript",
		Functions: []FunctionInfo{{Name: "helperFunc"}},
		Exports:   []ExportInfo{{ExportType: "default", Name: "helperFunc"}},
	}
	tracker.AddParseResult("src/helper.js", helperResult)

	// Add utils.js that imports helper
	utilsResult := &ParseResult{
		FilePath: "src/utils.js",
		Language: "javascript",
		Imports: []ImportInfo{
			{
				Source:     "./helper.js",
				ImportType: "default",
				LocalName:  "helper",
				StartLine:  1,
				IsExternal: false,
			},
		},
	}
	tracker.AddParseResult("src/utils.js", utilsResult)

	// Resolve dependencies
	err := tracker.ResolveDependencies("/project")
	require.NoError(t, err)

	// Check that dependency was resolved
	deps, _ := tracker.GetDependencies("src/utils.js")
	helperDep := deps[0]
	assert.True(t, helperDep.IsResolved)
	assert.Equal(t, "src/helper.js", helperDep.ResolvedPath)
}

func TestDependencyTracker_BuildModuleGraph(t *testing.T) {
	tracker := NewDependencyTracker()

	// Add multiple files with dependencies
	files := map[string]*ParseResult{
		"src/main.js": {
			FilePath: "src/main.js",
			Language: "javascript",
			Imports: []ImportInfo{
				{Source: "react", ImportType: "default", LocalName: "React", IsExternal: true},
				{Source: "./utils.js", ImportType: "named", Specifiers: []string{"processData"}, IsExternal: false},
			},
			Functions: []FunctionInfo{{Name: "App"}},
			Exports:   []ExportInfo{{ExportType: "default", Name: "App"}},
		},
		"src/utils.js": {
			FilePath: "src/utils.js",
			Language: "javascript",
			Imports: []ImportInfo{
				{Source: "lodash", ImportType: "named", Specifiers: []string{"map"}, IsExternal: true},
			},
			Functions: []FunctionInfo{{Name: "processData"}},
			Exports:   []ExportInfo{{ExportType: "named", Specifiers: []string{"processData"}}},
		},
	}

	for path, result := range files {
		tracker.AddParseResult(path, result)
	}

	// Build module graph
	graph, err := tracker.BuildModuleGraph()
	require.NoError(t, err)
	require.NotNil(t, graph)

	// Verify nodes
	assert.GreaterOrEqual(t, len(graph.Nodes), 2) // At least 2 internal files

	// Find main.js node
	var mainNode *ModuleNode
	for i, node := range graph.Nodes {
		if node.FilePath == "src/main.js" {
			mainNode = &graph.Nodes[i]
			break
		}
	}
	require.NotNil(t, mainNode)
	assert.Equal(t, "internal", mainNode.NodeType)
	assert.Equal(t, "main", mainNode.ModuleName)
	assert.Contains(t, mainNode.ExportedItems, "default:App")

	// Find external react node
	var reactNode *ModuleNode
	for i, node := range graph.Nodes {
		if node.ModuleName == "react" {
			reactNode = &graph.Nodes[i]
			break
		}
	}
	require.NotNil(t, reactNode)
	assert.Equal(t, "external", reactNode.NodeType)

	// Verify edges (main->react, main->utils, utils->lodash)
	assert.GreaterOrEqual(t, len(graph.Edges), 2) // At least main->react, main->utils

	// Verify stats
	assert.Equal(t, len(graph.Nodes), graph.Stats.TotalNodes)
	assert.Equal(t, len(graph.Edges), graph.Stats.TotalEdges)
	assert.Greater(t, graph.Stats.ExternalNodes, 0)
	assert.Greater(t, graph.Stats.InternalNodes, 0)
}

func TestDependencyTracker_GetDependents(t *testing.T) {
	tracker := NewDependencyTracker()

	// Set up files where multiple files depend on utils.js
	files := map[string]*ParseResult{
		"src/main.js": {
			FilePath: "src/main.js",
			Imports: []ImportInfo{
				{Source: "./utils.js", ImportType: "named", Specifiers: []string{"helper"}, IsExternal: false},
			},
		},
		"src/component.js": {
			FilePath: "src/component.js",
			Imports: []ImportInfo{
				{Source: "./utils.js", ImportType: "default", LocalName: "utils", IsExternal: false},
			},
		},
		"src/utils.js": {
			FilePath:  "src/utils.js",
			Functions: []FunctionInfo{{Name: "helper"}},
			Exports:   []ExportInfo{{ExportType: "named", Specifiers: []string{"helper"}}},
		},
	}

	for path, result := range files {
		tracker.AddParseResult(path, result)
	}

	// Get dependents of utils.js
	dependents := tracker.GetDependents("./utils.js", false)

	assert.Len(t, dependents, 2)
	assert.Contains(t, dependents, "src/main.js")
	assert.Contains(t, dependents, "src/component.js")
}

func TestDependencyTracker_ExternalPackageTypes(t *testing.T) {
	tracker := NewDependencyTracker()

	result := &ParseResult{
		FilePath: "src/test.js",
		Imports: []ImportInfo{
			{Source: "react", ImportType: "default", IsExternal: true},
			{Source: "@types/node", ImportType: "named", Specifiers: []string{"Buffer"}, IsExternal: true},
			{Source: "fs", ImportType: "named", Specifiers: []string{"readFile"}, IsExternal: true},
		},
	}

	tracker.AddParseResult("src/test.js", result)

	externals := tracker.GetExternalPackages()

	// Check React (npm package)
	react := externals["react"]
	assert.Equal(t, "npm", react.PackageType)

	// Check @types/node (scoped package)
	typesNode := externals["@types/node"]
	assert.Equal(t, "scoped", typesNode.PackageType)

	// Check fs (built-in package)
	fs := externals["fs"]
	assert.Equal(t, "built-in", fs.PackageType)
}

func TestDependencyTracker_EdgeWeightCalculation(t *testing.T) {
	tracker := NewDependencyTracker()

	testCases := []struct {
		dep            Dependency
		expectedWeight int
	}{
		{
			dep:            Dependency{ImportType: "side-effect"},
			expectedWeight: 1,
		},
		{
			dep:            Dependency{ImportType: "default"},
			expectedWeight: 2,
		},
		{
			dep:            Dependency{ImportType: "namespace"},
			expectedWeight: 3,
		},
		{
			dep:            Dependency{ImportType: "named", ImportedNames: []string{"a", "b", "c"}},
			expectedWeight: 4, // 1 + 3 names
		},
	}

	for _, tc := range testCases {
		weight := tracker.calculateEdgeWeight(tc.dep)
		assert.Equal(t, tc.expectedWeight, weight, "Failed for import type: %s", tc.dep.ImportType)
	}
}

func TestDependencyTracker_PackageNameExtraction(t *testing.T) {
	tracker := NewDependencyTracker()

	testCases := []struct {
		source      string
		packageName string
	}{
		{"react", "react"},
		{"@types/node", "@types/node"},
		{"@babel/core", "@babel/core"},
		{"lodash/map", "lodash"},
		{"react-dom/client", "react-dom"},
		{"@types/node/fs", "@types/node"},
	}

	for _, tc := range testCases {
		result := tracker.getPackageName(tc.source)
		assert.Equal(t, tc.packageName, result, "Failed for source: %s", tc.source)
	}
}

func TestDependencyTracker_ModuleNameExtraction(t *testing.T) {
	tracker := NewDependencyTracker()

	testCases := []struct {
		filePath   string
		moduleName string
	}{
		{"src/utils.js", "utils"},
		{"components/Button.tsx", "Button"},
		{"lib/index.ts", "index"},
		{"src/services/api.service.js", "api.service"},
	}

	for _, tc := range testCases {
		result := tracker.getModuleName(tc.filePath)
		assert.Equal(t, tc.moduleName, result, "Failed for path: %s", tc.filePath)
	}
}
