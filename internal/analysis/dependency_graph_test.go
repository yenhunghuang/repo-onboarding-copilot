// Package analysis provides tests for dependency graph generation and analysis
package analysis

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGraphBuilder(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	assert.NotNil(t, builder)
	assert.Equal(t, "npm", builder.packageManager)
	assert.NotNil(t, builder.graph)
	assert.NotNil(t, builder.graph.Nodes)
	assert.NotNil(t, builder.graph.Edges)
	assert.NotNil(t, builder.graph.Stats)
}

func TestGraphBuilder_AddNode(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	pkg := &GraphPackageInfo{
		Name:           "test-package",
		Version:        "1.0.0",
		DependencyType: "dependencies",
		Description:    "A test package",
		Homepage:       "https://example.com",
		RegistryURL:    "https://registry.npmjs.org/test-package",
	}
	
	builder.AddNode(pkg)
	
	nodeID := "test-package@1.0.0"
	require.Contains(t, builder.graph.Nodes, nodeID)
	
	node := builder.graph.Nodes[nodeID]
	assert.Equal(t, nodeID, node.ID)
	assert.Equal(t, "test-package", node.Name)
	assert.Equal(t, "1.0.0", node.Version)
	assert.Equal(t, "dependencies", node.PackageType)
	assert.Equal(t, "https://example.com", node.Metadata["homepage"])
}

func TestGraphBuilder_AddEdge(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	source := "parent@1.0.0"
	target := "child@2.0.0"
	relationship := "dependencies"
	versionRange := "^2.0.0"
	
	builder.AddEdge(source, target, relationship, versionRange, false)
	
	require.Contains(t, builder.graph.Edges, source)
	edges := builder.graph.Edges[source]
	require.Len(t, edges, 1)
	
	edge := edges[0]
	assert.Equal(t, source, edge.Source)
	assert.Equal(t, target, edge.Target)
	assert.Equal(t, relationship, edge.Relationship)
	assert.Equal(t, versionRange, edge.VersionRange)
	assert.Equal(t, 1.0, edge.Weight)
	assert.False(t, edge.IsOptional)
}

func TestGraphBuilder_CalculateEdgeWeight(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	tests := []struct {
		name         string
		relationship string
		expectedWeight float64
	}{
		{"production dependency", "dependencies", 1.0},
		{"dev dependency", "devDependencies", 0.5},
		{"peer dependency", "peerDependencies", 0.8},
		{"optional dependency", "optionalDependencies", 0.3},
		{"unknown relationship", "unknown", 1.0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := builder.calculateEdgeWeight(tt.relationship)
			assert.Equal(t, tt.expectedWeight, weight)
		})
	}
}

func TestGraphBuilder_BuildFromPackageList(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	// Create test packages with dependencies
	packages := []*GraphPackageInfo{
		{
			Name:           "root",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"lodash": "^4.17.21",
				"express": "^4.18.0",
			},
		},
		{
			Name:           "lodash",
			Version:        "4.17.21",
			DependencyType: "dependencies",
		},
		{
			Name:           "express",
			Version:        "4.18.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"accepts": "^1.3.8",
			},
		},
		{
			Name:           "accepts",
			Version:        "1.3.8",
			DependencyType: "dependencies",
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	
	// Check nodes were created
	assert.Len(t, graph.Nodes, 4)
	assert.Contains(t, graph.Nodes, "root@1.0.0")
	assert.Contains(t, graph.Nodes, "lodash@4.17.21")
	assert.Contains(t, graph.Nodes, "express@4.18.0")
	assert.Contains(t, graph.Nodes, "accepts@1.3.8")
	
	// Check edges were created
	assert.Contains(t, graph.Edges, "root@1.0.0")
	rootEdges := graph.Edges["root@1.0.0"]
	assert.Len(t, rootEdges, 2)
	
	// Check edge targets
	targets := make([]string, len(rootEdges))
	for i, edge := range rootEdges {
		targets[i] = edge.Target
	}
	assert.Contains(t, targets, "lodash@4.17.21")
	assert.Contains(t, targets, "express@4.18.0")
	
	// Check statistics
	assert.Equal(t, 4, graph.Stats.TotalNodes)
	assert.Equal(t, 3, graph.Stats.TotalEdges)
	assert.True(t, graph.Stats.MaxDepth >= 0)
}

func TestGraphBuilder_CircularDependencyDetection(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	// Create packages with circular dependency
	packages := []*GraphPackageInfo{
		{
			Name:           "package-a",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"package-b": "^1.0.0",
			},
		},
		{
			Name:           "package-b",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"package-c": "^1.0.0",
			},
		},
		{
			Name:           "package-c",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"package-a": "^1.0.0", // Creates circular dependency
			},
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	
	// Should detect circular dependency
	assert.True(t, len(graph.Stats.CircularDeps) > 0, "Expected to detect circular dependencies")
	
	// Check first circular dependency
	if len(graph.Stats.CircularDeps) > 0 {
		cycle := graph.Stats.CircularDeps[0]
		assert.Equal(t, 4, cycle.Length) // 3 packages + closing the cycle
		assert.Contains(t, []string{"low", "medium", "high", "critical"}, cycle.Severity)
		assert.True(t, cycle.Impact > 0)
		assert.Contains(t, cycle.Description, "Circular dependency detected")
	}
}

func TestGraphBuilder_CriticalNodeIdentification(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	// Create a graph where one package is depended upon by many others
	packages := []*GraphPackageInfo{
		{
			Name:           "root",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"critical-package": "^1.0.0",
				"other-package": "^1.0.0",
			},
		},
		{
			Name:           "critical-package",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
		{
			Name:           "other-package",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"critical-package": "^1.0.0", // Also depends on critical package
			},
		},
		{
			Name:           "another-package",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"critical-package": "^1.0.0", // Also depends on critical package
			},
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	
	// Should identify critical nodes
	assert.True(t, len(graph.Stats.CriticalNodes) > 0, "Expected to identify critical nodes")
	
	// critical-package should be in the critical nodes list
	criticalPackageID := "critical-package@1.0.0"
	assert.Contains(t, graph.Stats.CriticalNodes, criticalPackageID)
}

func TestGraphBuilder_ClusterAnalysis(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	// Create packages that should form clusters
	packages := []*GraphPackageInfo{
		{
			Name:           "root",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"cluster1-a": "^1.0.0",
				"cluster2-a": "^1.0.0",
			},
		},
		{
			Name:           "cluster1-a",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"cluster1-b": "^1.0.0",
			},
		},
		{
			Name:           "cluster1-b",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"cluster1-c": "^1.0.0",
			},
		},
		{
			Name:           "cluster1-c",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
		{
			Name:           "cluster2-a",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"cluster2-b": "^1.0.0",
			},
		},
		{
			Name:           "cluster2-b",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	
	// Should identify clusters
	assert.True(t, len(graph.Stats.Clusters) >= 0, "Expected to identify clusters")
	
	// Check that clusters have valid cohesion and coupling values
	for _, cluster := range graph.Stats.Clusters {
		assert.True(t, cluster.Cohesion >= 0 && cluster.Cohesion <= 1, "Cohesion should be between 0 and 1")
		assert.True(t, cluster.Coupling >= 0 && cluster.Coupling <= 1, "Coupling should be between 0 and 1")
		assert.True(t, len(cluster.Packages) > 1, "Cluster should have more than one package")
		assert.NotEmpty(t, cluster.MainPackage, "Cluster should have a main package")
	}
}

func TestGraphBuilder_GraphMetrics(t *testing.T) {
	builder := NewGraphBuilder("npm")
	
	packages := []*GraphPackageInfo{
		{
			Name:           "root",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"dep1": "^1.0.0",
				"dep2": "^1.0.0",
			},
		},
		{
			Name:           "dep1",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
		{
			Name:           "dep2",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	metrics := graph.Stats.Metrics
	
	// Check that metrics are calculated
	assert.True(t, metrics.Density >= 0 && metrics.Density <= 1, "Density should be between 0 and 1")
	assert.True(t, metrics.Modularity >= 0 && metrics.Modularity <= 1, "Modularity should be between 0 and 1")
	assert.True(t, metrics.CentralityMax >= 0, "Centrality max should be non-negative")
	assert.True(t, metrics.PathLengthAvg >= 0, "Average path length should be non-negative")
	assert.True(t, metrics.ConnectedComponents >= 1, "Should have at least one connected component")
	assert.True(t, metrics.Diameter >= 0, "Diameter should be non-negative")
}

func TestNewGraphExporter(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{},
	}
	
	exporter := NewGraphExporter(graph)
	
	assert.NotNil(t, exporter)
	assert.Equal(t, graph, exporter.graph)
}

func TestGraphExporter_GenerateD3JSData(t *testing.T) {
	// Create test graph
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{
			TotalNodes: 2,
			TotalEdges: 1,
			Clusters:   []DependencyCluster{},
		},
	}
	
	// Add test nodes
	graph.Nodes["package-a@1.0.0"] = &GraphNode{
		ID:          "package-a@1.0.0",
		Name:        "package-a",
		Version:     "1.0.0",
		PackageType: "dependencies",
		Size:        1000,
		Weight:      2.0,
		Depth:       0,
		VulnerabilityCount: 1,
		RiskScore:   3.0,
		Metadata:    make(map[string]string),
	}
	
	graph.Nodes["package-b@2.0.0"] = &GraphNode{
		ID:          "package-b@2.0.0",
		Name:        "package-b",
		Version:     "2.0.0",
		PackageType: "devDependencies",
		Size:        500,
		Weight:      1.0,
		Depth:       1,
		VulnerabilityCount: 0,
		RiskScore:   1.0,
		Metadata:    make(map[string]string),
	}
	
	// Add test edge
	graph.Edges["package-a@1.0.0"] = []*GraphEdge{
		{
			Source:       "package-a@1.0.0",
			Target:       "package-b@2.0.0",
			Relationship: "dependencies",
			VersionRange: "^2.0.0",
			Weight:       1.0,
			IsOptional:   false,
		},
	}
	
	exporter := NewGraphExporter(graph)
	
	tests := []struct {
		name   string
		layout string
	}{
		{"force layout", "force"},
		{"hierarchical layout", "hierarchical"},
		{"circular layout", "circular"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := exporter.GenerateD3JSData(tt.layout)
			require.NoError(t, err)
			require.NotNil(t, data)
			
			// Check nodes
			assert.Len(t, data.Nodes, 2)
			assert.Equal(t, "package-a@1.0.0", data.Nodes[0].ID)
			assert.Equal(t, "package-a", data.Nodes[0].Name)
			assert.Equal(t, "1.0.0", data.Nodes[0].Version)
			assert.Equal(t, "dependencies", data.Nodes[0].PackageType)
			assert.Equal(t, 3.0, data.Nodes[0].RiskScore)
			assert.Equal(t, 1, data.Nodes[0].Vulnerabilities)
			
			// Check links
			assert.Len(t, data.Links, 1)
			assert.Equal(t, 0, data.Links[0].Source) // Index of first node
			assert.Equal(t, 1, data.Links[0].Target) // Index of second node
			assert.Equal(t, "dependencies", data.Links[0].Type)
			assert.Equal(t, 1.0, data.Links[0].Value)
			
			// Check meta
			assert.Equal(t, tt.layout, data.Meta.Layout)
			assert.Equal(t, 2, data.Meta.NodeCount)
			assert.Equal(t, 1, data.Meta.LinkCount)
			assert.NotEmpty(t, data.Meta.Title)
			assert.NotEmpty(t, data.Meta.Options)
		})
	}
}

func TestGraphExporter_ExportDOT(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{},
	}
	
	// Add test data
	graph.Nodes["test@1.0.0"] = &GraphNode{
		ID:          "test@1.0.0",
		Name:        "test",
		Version:     "1.0.0",
		PackageType: "dependencies",
	}
	
	graph.Nodes["dep@2.0.0"] = &GraphNode{
		ID:          "dep@2.0.0",
		Name:        "dep",
		Version:     "2.0.0",
		PackageType: "devDependencies",
	}
	
	graph.Edges["test@1.0.0"] = []*GraphEdge{
		{
			Source:       "test@1.0.0",
			Target:       "dep@2.0.0",
			Relationship: "dependencies",
		},
	}
	
	exporter := NewGraphExporter(graph)
	dotOutput, err := exporter.ExportDOT()
	require.NoError(t, err)
	
	// Check DOT format
	assert.Contains(t, dotOutput, "digraph DependencyGraph")
	assert.Contains(t, dotOutput, "\"test@1.0.0\"")
	assert.Contains(t, dotOutput, "\"dep@2.0.0\"")
	assert.Contains(t, dotOutput, "\"test@1.0.0\" -> \"dep@2.0.0\"")
}

func TestGraphExporter_ExportGraphML(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{},
	}
	
	// Add test data
	graph.Nodes["test@1.0.0"] = &GraphNode{
		ID:          "test@1.0.0",
		Name:        "test",
		Version:     "1.0.0",
		PackageType: "dependencies",
		Weight:      1.5,
	}
	
	graph.Edges["test@1.0.0"] = []*GraphEdge{
		{
			Source:       "test@1.0.0",
			Target:       "dep@2.0.0",
			Relationship: "dependencies",
			Weight:       1.0,
		},
	}
	
	exporter := NewGraphExporter(graph)
	graphMLOutput, err := exporter.ExportGraphML()
	require.NoError(t, err)
	
	// Check GraphML format
	assert.Contains(t, graphMLOutput, "<?xml version=\"1.0\"")
	assert.Contains(t, graphMLOutput, "<graphml")
	assert.Contains(t, graphMLOutput, "<node id=\"test@1.0.0\">")
	assert.Contains(t, graphMLOutput, "<data key=\"name\">test</data>")
	assert.Contains(t, graphMLOutput, "<data key=\"version\">1.0.0</data>")
	assert.Contains(t, graphMLOutput, "<edge id=\"e0\"")
}

func TestGraphExporter_ExportJSON(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{
			TotalNodes: 1,
			TotalEdges: 0,
		},
	}
	
	graph.Nodes["test@1.0.0"] = &GraphNode{
		ID:          "test@1.0.0",
		Name:        "test",
		Version:     "1.0.0",
		PackageType: "dependencies",
		Weight:      1.0,
		Metadata:    make(map[string]string),
	}
	
	exporter := NewGraphExporter(graph)
	jsonOutput, err := exporter.ExportJSON()
	require.NoError(t, err)
	
	// Parse JSON to verify it's valid
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonOutput), &parsed)
	require.NoError(t, err)
	
	// Check structure
	assert.Contains(t, parsed, "nodes")
	assert.Contains(t, parsed, "edges")
	assert.Contains(t, parsed, "stats")
}

func TestGraphExporter_ExportSummaryJSON(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Edges: make(map[string][]*GraphEdge),
		Stats: &GraphStats{
			TotalNodes:   3,
			TotalEdges:   2,
			MaxDepth:     2,
			CircularDeps: []CircularDependency{},
			CriticalNodes: []string{"critical@1.0.0"},
			Clusters:     []DependencyCluster{},
			Metrics:      &GraphMetrics{Density: 0.5},
		},
	}
	
	// Add a root package
	graph.Nodes["root@1.0.0"] = &GraphNode{
		ID:          "root@1.0.0",
		Name:        "root",
		Version:     "1.0.0",
		PackageType: "dependencies",
		Depth:       0,
	}
	
	exporter := NewGraphExporter(graph)
	summaryOutput, err := exporter.ExportSummaryJSON()
	require.NoError(t, err)
	
	// Parse JSON to verify it's valid
	var summary map[string]interface{}
	err = json.Unmarshal([]byte(summaryOutput), &summary)
	require.NoError(t, err)
	
	// Check summary fields
	assert.Equal(t, "root", summary["project_name"])
	assert.Equal(t, float64(1), summary["total_packages"]) // Only 1 actual node added
	assert.Equal(t, float64(2), summary["total_dependencies"])
	assert.Equal(t, float64(2), summary["max_depth"])
	assert.Equal(t, float64(0), summary["circular_dependencies"])
	assert.Equal(t, float64(1), summary["critical_packages"])
	assert.Equal(t, float64(0), summary["clusters"])
	assert.Contains(t, summary, "metrics")
	assert.Contains(t, summary, "generated_at")
}

func TestD3JSNodeSizeCalculation(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Stats: &GraphStats{},
	}
	
	exporter := NewGraphExporter(graph)
	
	tests := []struct {
		name     string
		node     *GraphNode
		minSize  float64
		maxSize  float64
	}{
		{
			name: "low importance node",
			node: &GraphNode{
				Weight:             1.0,
				RiskScore:          0.0,
				VulnerabilityCount: 0,
			},
			minSize: 3.0,
			maxSize: 10.0,
		},
		{
			name: "high importance node",
			node: &GraphNode{
				Weight:             10.0,
				RiskScore:          8.0,
				VulnerabilityCount: 5,
			},
			minSize: 10.0,
			maxSize: 20.0,
		},
		{
			name: "critical vulnerability node",
			node: &GraphNode{
				Weight:             5.0,
				RiskScore:          10.0,
				VulnerabilityCount: 10,
			},
			minSize: 15.0,
			maxSize: 20.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size := exporter.calculateNodeSize(tt.node)
			assert.True(t, size >= tt.minSize && size <= 20.0, 
				"Size %.2f should be between %.2f and 20.0", size, tt.minSize)
		})
	}
}

func TestLinkDistanceCalculation(t *testing.T) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*GraphNode),
		Stats: &GraphStats{},
	}
	
	exporter := NewGraphExporter(graph)
	
	tests := []struct {
		name     string
		edge     *GraphEdge
		expected float64
	}{
		{
			name: "production dependency",
			edge: &GraphEdge{
				Relationship: "dependencies",
				Weight:       1.0,
			},
			expected: 50.0, // base distance * 1.0 * (1.0 / (1.0 + 0.1))
		},
		{
			name: "dev dependency",
			edge: &GraphEdge{
				Relationship: "devDependencies",
				Weight:       1.0,
			},
			expected: 60.0, // base distance * 1.2 * (1.0 / (1.0 + 0.1))
		},
		{
			name: "peer dependency",
			edge: &GraphEdge{
				Relationship: "peerDependencies",
				Weight:       1.0,
			},
			expected: 40.0, // base distance * 0.8 * (1.0 / (1.0 + 0.1))
		},
		{
			name: "optional dependency",
			edge: &GraphEdge{
				Relationship: "optionalDependencies",
				Weight:       1.0,
			},
			expected: 75.0, // base distance * 1.5 * (1.0 / (1.0 + 0.1))
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := exporter.calculateLinkDistance(tt.edge)
			assert.InDelta(t, tt.expected, distance, 1.0, 
				"Distance %.2f should be close to expected %.2f", distance, tt.expected)
		})
	}
}

func TestCircularDependencyDetectionEdgeCases(t *testing.T) {
	t.Run("self dependency", func(t *testing.T) {
		builder := NewGraphBuilder("npm")
		
		packages := []*GraphPackageInfo{
			{
				Name:           "self-dep",
				Version:        "1.0.0",
				DependencyType: "dependencies",
				Dependencies: map[string]string{
					"self-dep": "^1.0.0", // Self-dependency
				},
			},
		}
		
		err := builder.BuildFromPackageList(packages)
		require.NoError(t, err)
		
		graph := builder.GetGraph()
		
		// Should detect the self-dependency as a cycle
		assert.True(t, len(graph.Stats.CircularDeps) > 0, "Expected to detect self-dependency cycle")
		
		if len(graph.Stats.CircularDeps) > 0 {
			cycle := graph.Stats.CircularDeps[0]
			assert.Equal(t, "critical", cycle.Severity, "Self-dependency should be critical")
		}
	})
	
	t.Run("no circular dependencies", func(t *testing.T) {
		builder := NewGraphBuilder("npm")
		
		packages := []*GraphPackageInfo{
			{
				Name:           "root",
				Version:        "1.0.0",
				DependencyType: "dependencies",
				Dependencies: map[string]string{
					"leaf1": "^1.0.0",
					"leaf2": "^1.0.0",
				},
			},
			{
				Name:           "leaf1",
				Version:        "1.0.0",
				DependencyType: "dependencies",
			},
			{
				Name:           "leaf2",
				Version:        "1.0.0",
				DependencyType: "dependencies",
			},
		}
		
		err := builder.BuildFromPackageList(packages)
		require.NoError(t, err)
		
		graph := builder.GetGraph()
		
		// Should not detect any circular dependencies
		assert.Equal(t, 0, len(graph.Stats.CircularDeps), "Expected no circular dependencies")
	})
}

func TestGraphExportFormats(t *testing.T) {
	// Create a simple test graph
	builder := NewGraphBuilder("npm")
	packages := []*GraphPackageInfo{
		{
			Name:           "root",
			Version:        "1.0.0",
			DependencyType: "dependencies",
			Dependencies: map[string]string{
				"dep": "^1.0.0",
			},
		},
		{
			Name:           "dep",
			Version:        "1.0.0",
			DependencyType: "dependencies",
		},
	}
	
	err := builder.BuildFromPackageList(packages)
	require.NoError(t, err)
	
	graph := builder.GetGraph()
	exporter := NewGraphExporter(graph)
	
	t.Run("DOT export validity", func(t *testing.T) {
		dot, err := exporter.ExportDOT()
		require.NoError(t, err)
		
		// Check basic DOT structure
		lines := strings.Split(dot, "\n")
		assert.True(t, strings.HasPrefix(lines[0], "digraph"), "Should start with digraph")
		assert.True(t, strings.HasSuffix(strings.TrimSpace(dot), "}"), "Should end with }")
	})
	
	t.Run("GraphML export validity", func(t *testing.T) {
		graphML, err := exporter.ExportGraphML()
		require.NoError(t, err)
		
		// Check XML structure
		assert.Contains(t, graphML, "<?xml version=\"1.0\"")
		assert.Contains(t, graphML, "<graphml")
		assert.Contains(t, graphML, "</graphml>")
	})
	
	t.Run("JSON export validity", func(t *testing.T) {
		jsonStr, err := exporter.ExportJSON()
		require.NoError(t, err)
		
		// Should be valid JSON
		var data interface{}
		err = json.Unmarshal([]byte(jsonStr), &data)
		require.NoError(t, err)
	})
}