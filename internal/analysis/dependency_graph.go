// Package analysis provides dependency graph generation and analysis functionality
package analysis

import (
	"fmt"
	"time"
)

// DependencyGraph represents a weighted directed graph of package dependencies
type DependencyGraph struct {
	Nodes map[string]*GraphNode `json:"nodes"`
	Edges map[string][]*GraphEdge `json:"edges"`
	Stats *GraphStats `json:"stats"`
}

// GraphNode represents a package in the dependency graph
type GraphNode struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	PackageType string            `json:"package_type"` // direct, dev, peer, optional
	Size        int64             `json:"size"`         // estimated bundle size
	Weight      float64           `json:"weight"`       // importance score
	Depth       int               `json:"depth"`        // distance from root
	Metadata    map[string]string `json:"metadata"`
	
	// Analysis results
	VulnerabilityCount int     `json:"vulnerability_count"`
	LicenseInfo        string  `json:"license_info"`
	UpdatesAvailable   int     `json:"updates_available"`
	RiskScore          float64 `json:"risk_score"`
}

// GraphEdge represents a dependency relationship between packages
type GraphEdge struct {
	Source      string  `json:"source"`
	Target      string  `json:"target"`
	Relationship string `json:"relationship"` // dependency, devDependency, peerDependency
	VersionRange string `json:"version_range"`
	Weight       float64 `json:"weight"` // relationship strength
	IsOptional   bool    `json:"is_optional"`
}

// GraphStats provides summary statistics about the dependency graph
type GraphStats struct {
	TotalNodes      int                    `json:"total_nodes"`
	TotalEdges      int                    `json:"total_edges"`
	MaxDepth        int                    `json:"max_depth"`
	CircularDeps    []CircularDependency   `json:"circular_dependencies"`
	CriticalNodes   []string               `json:"critical_nodes"`
	Clusters        []DependencyCluster    `json:"clusters"`
	Metrics         *GraphMetrics          `json:"metrics"`
	GeneratedAt     time.Time             `json:"generated_at"`
}

// CircularDependency represents a detected circular dependency
type CircularDependency struct {
	Cycle       []string `json:"cycle"`
	Length      int      `json:"length"`
	Severity    string   `json:"severity"` // low, medium, high, critical
	Impact      float64  `json:"impact"`   // impact score
	Description string   `json:"description"`
}

// DependencyCluster represents a group of tightly coupled dependencies
type DependencyCluster struct {
	ID        string   `json:"id"`
	Packages  []string `json:"packages"`
	Cohesion  float64  `json:"cohesion"`  // internal connectivity
	Coupling  float64  `json:"coupling"`  // external dependencies
	MainPackage string  `json:"main_package"`
}

// GraphMetrics provides detailed graph analysis metrics
type GraphMetrics struct {
	Density         float64 `json:"density"`          // edge density
	Modularity      float64 `json:"modularity"`       // clustering coefficient
	CentralityMax   float64 `json:"centrality_max"`   // highest centrality score
	PathLengthAvg   float64 `json:"path_length_avg"`  // average shortest path
	ConnectedComponents int `json:"connected_components"`
	Diameter        int     `json:"diameter"`         // longest shortest path
}

// D3JSVisualizationData represents data formatted for D3.js visualization
type D3JSVisualizationData struct {
	Nodes []D3JSNode `json:"nodes"`
	Links []D3JSLink `json:"links"`
	Meta  D3JSMeta   `json:"meta"`
}

// D3JSNode represents a node in D3.js format
type D3JSNode struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Group    int     `json:"group"`    // clustering group
	Size     float64 `json:"size"`     // visual size
	Color    string  `json:"color"`    // node color
	Label    string  `json:"label"`    // display label
	X        float64 `json:"x,omitempty"` // fixed position x
	Y        float64 `json:"y,omitempty"` // fixed position y
	Fixed    bool    `json:"fx,omitempty"` // is position fixed
	
	// Additional metadata for tooltips
	Version     string  `json:"version"`
	PackageType string  `json:"package_type"`
	RiskScore   float64 `json:"risk_score"`
	Vulnerabilities int `json:"vulnerabilities"`
}

// D3JSLink represents an edge in D3.js format
type D3JSLink struct {
	Source   interface{} `json:"source"` // can be index or id
	Target   interface{} `json:"target"` // can be index or id
	Value    float64     `json:"value"`  // link strength
	Type     string      `json:"type"`   // relationship type
	Distance float64     `json:"distance,omitempty"` // link distance
	Color    string      `json:"color,omitempty"`    // link color
}

// D3JSMeta contains metadata for D3.js visualization
type D3JSMeta struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	NodeCount   int               `json:"node_count"`
	LinkCount   int               `json:"link_count"`
	Layout      string            `json:"layout"` // force, hierarchical, circular
	Options     map[string]interface{} `json:"options"`
	GeneratedAt time.Time         `json:"generated_at"`
}

// GraphBuilder provides functionality to build dependency graphs
type GraphBuilder struct {
	graph *DependencyGraph
	packageManager string
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(packageManager string) *GraphBuilder {
	return &GraphBuilder{
		graph: &DependencyGraph{
			Nodes: make(map[string]*GraphNode),
			Edges: make(map[string][]*GraphEdge),
			Stats: &GraphStats{
				CircularDeps: make([]CircularDependency, 0),
				CriticalNodes: make([]string, 0),
				Clusters: make([]DependencyCluster, 0),
				Metrics: &GraphMetrics{},
				GeneratedAt: time.Now(),
			},
		},
		packageManager: packageManager,
	}
}

// GraphPackageInfo represents package information for graph building
type GraphPackageInfo struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	DependencyType   string            `json:"dependency_type"`
	Dependencies     map[string]string `json:"dependencies"`
	DevDependencies  map[string]string `json:"dev_dependencies"`
	PeerDependencies map[string]string `json:"peer_dependencies"`
	Description      string            `json:"description"`
	Homepage         string            `json:"homepage"`
	Repository       string            `json:"repository"`
	RegistryURL      string            `json:"registry_url"`
}

// AddNode adds a package node to the graph
func (gb *GraphBuilder) AddNode(pkg *GraphPackageInfo) {
	nodeID := fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
	
	node := &GraphNode{
		ID:          nodeID,
		Name:        pkg.Name,
		Version:     pkg.Version,
		PackageType: pkg.DependencyType,
		Size:        0, // Will be calculated later
		Weight:      1.0,
		Depth:       0, // Will be calculated during graph analysis
		Metadata:    make(map[string]string),
		VulnerabilityCount: 0,
		LicenseInfo: "",
		UpdatesAvailable: 0,
		RiskScore: 0.0,
	}
	
	// Add metadata
	node.Metadata["registry_url"] = pkg.RegistryURL
	node.Metadata["homepage"] = pkg.Homepage
	node.Metadata["repository"] = pkg.Repository
	if pkg.Description != "" {
		node.Metadata["description"] = pkg.Description
	}
	
	gb.graph.Nodes[nodeID] = node
}

// AddEdge adds a dependency relationship to the graph
func (gb *GraphBuilder) AddEdge(source, target, relationship, versionRange string, isOptional bool) {
	edge := &GraphEdge{
		Source:       source,
		Target:       target,
		Relationship: relationship,
		VersionRange: versionRange,
		Weight:       gb.calculateEdgeWeight(relationship),
		IsOptional:   isOptional,
	}
	
	gb.graph.Edges[source] = append(gb.graph.Edges[source], edge)
}

// calculateEdgeWeight calculates the weight of an edge based on relationship type
func (gb *GraphBuilder) calculateEdgeWeight(relationship string) float64 {
	weights := map[string]float64{
		"dependencies":     1.0,  // Production dependencies
		"devDependencies":  0.5,  // Development dependencies
		"peerDependencies": 0.8,  // Peer dependencies
		"optionalDependencies": 0.3, // Optional dependencies
	}
	
	if weight, exists := weights[relationship]; exists {
		return weight
	}
	return 1.0 // Default weight
}

// BuildFromPackageList builds a graph from a list of packages
func (gb *GraphBuilder) BuildFromPackageList(packages []*GraphPackageInfo) error {
	// First pass: Add all nodes
	for _, pkg := range packages {
		gb.AddNode(pkg)
	}
	
	// Second pass: Add edges based on dependencies
	for _, pkg := range packages {
		sourceID := fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)
		
		// Add production dependencies
		for depName, depVersion := range pkg.Dependencies {
			targetID := gb.findBestMatchingNode(depName, depVersion)
			if targetID != "" {
				gb.AddEdge(sourceID, targetID, "dependencies", depVersion, false)
			}
		}
		
		// Add dev dependencies
		for depName, depVersion := range pkg.DevDependencies {
			targetID := gb.findBestMatchingNode(depName, depVersion)
			if targetID != "" {
				gb.AddEdge(sourceID, targetID, "devDependencies", depVersion, false)
			}
		}
		
		// Add peer dependencies
		for depName, depVersion := range pkg.PeerDependencies {
			targetID := gb.findBestMatchingNode(depName, depVersion)
			if targetID != "" {
				gb.AddEdge(sourceID, targetID, "peerDependencies", depVersion, false)
			}
		}
	}
	
	// Calculate graph metrics and analysis
	gb.calculateNodeDepths()
	gb.calculateNodeWeights()
	gb.detectCircularDependencies()
	gb.identifyCriticalNodes()
	gb.performClusterAnalysis()
	gb.calculateGraphMetrics()
	gb.updateGraphStats()
	
	return nil
}

// findBestMatchingNode finds the best matching node for a dependency
func (gb *GraphBuilder) findBestMatchingNode(name, versionRange string) string {
	// Look for exact matches first
	for nodeID, node := range gb.graph.Nodes {
		if node.Name == name {
			// TODO: Implement semver range matching
			// For now, return the first match
			return nodeID
		}
	}
	return ""
}

// calculateNodeDepths calculates the depth of each node from root dependencies
func (gb *GraphBuilder) calculateNodeDepths() {
	// Find root nodes (nodes with no incoming edges)
	incomingEdges := make(map[string]int)
	for _, edges := range gb.graph.Edges {
		for _, edge := range edges {
			incomingEdges[edge.Target]++
		}
	}
	
	// BFS to calculate depths
	queue := make([]string, 0)
	depths := make(map[string]int)
	
	// Start with root nodes
	for nodeID := range gb.graph.Nodes {
		if incomingEdges[nodeID] == 0 {
			queue = append(queue, nodeID)
			depths[nodeID] = 0
		}
	}
	
	// BFS traversal
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentDepth := depths[current]
		
		if edges, exists := gb.graph.Edges[current]; exists {
			for _, edge := range edges {
				if _, visited := depths[edge.Target]; !visited {
					depths[edge.Target] = currentDepth + 1
					queue = append(queue, edge.Target)
				}
			}
		}
	}
	
	// Update node depths
	maxDepth := 0
	for nodeID, depth := range depths {
		if node, exists := gb.graph.Nodes[nodeID]; exists {
			node.Depth = depth
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}
	
	gb.graph.Stats.MaxDepth = maxDepth
}

// calculateNodeWeights calculates importance weights for nodes
func (gb *GraphBuilder) calculateNodeWeights() {
	// Calculate centrality-based weights
	for nodeID, node := range gb.graph.Nodes {
		// Count incoming and outgoing edges
		inDegree := 0
		outDegree := len(gb.graph.Edges[nodeID])
		
		for _, edges := range gb.graph.Edges {
			for _, edge := range edges {
				if edge.Target == nodeID {
					inDegree++
				}
			}
		}
		
		// Calculate weight based on centrality
		// Higher weight for nodes with more connections
		weight := float64(inDegree+outDegree) / 2.0
		if weight == 0 {
			weight = 1.0 // Minimum weight
		}
		
		node.Weight = weight
	}
}

// GetGraph returns the built dependency graph
func (gb *GraphBuilder) GetGraph() *DependencyGraph {
	return gb.graph
}

// updateGraphStats updates the graph statistics
func (gb *GraphBuilder) updateGraphStats() {
	gb.graph.Stats.TotalNodes = len(gb.graph.Nodes)
	
	totalEdges := 0
	for _, edges := range gb.graph.Edges {
		totalEdges += len(edges)
	}
	gb.graph.Stats.TotalEdges = totalEdges
	gb.graph.Stats.GeneratedAt = time.Now()
}