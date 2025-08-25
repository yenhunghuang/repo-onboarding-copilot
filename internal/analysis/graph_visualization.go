// Package analysis provides graph visualization and export functionality
package analysis

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// GraphExporter provides functionality to export graphs in various formats
type GraphExporter struct {
	graph *DependencyGraph
}

// NewGraphExporter creates a new graph exporter
func NewGraphExporter(graph *DependencyGraph) *GraphExporter {
	return &GraphExporter{graph: graph}
}

// GenerateD3JSData generates visualization data compatible with D3.js
func (ge *GraphExporter) GenerateD3JSData(layout string) (*D3JSVisualizationData, error) {
	nodes := make([]D3JSNode, 0, len(ge.graph.Nodes))
	links := make([]D3JSLink, 0, ge.graph.Stats.TotalEdges)
	
	// Color schemes for different node types and risk levels
	nodeColors := map[string]string{
		"dependencies":         "#2E86C1", // Blue for production dependencies
		"devDependencies":      "#28B463", // Green for dev dependencies
		"peerDependencies":     "#F39C12", // Orange for peer dependencies
		"optionalDependencies": "#8E44AD", // Purple for optional dependencies
		"unknown":              "#95A5A6", // Gray for unknown
	}
	
	riskColors := map[string]string{
		"low":      "#27AE60", // Green
		"medium":   "#F39C12", // Orange
		"high":     "#E74C3C", // Red
		"critical": "#8E44AD", // Purple
	}
	
	// Create node index mapping for D3.js links
	nodeIndexMap := make(map[string]int)
	
	// Generate nodes
	for i, nodeID := range ge.getSortedNodeIDs() {
		node := ge.graph.Nodes[nodeID]
		nodeIndexMap[nodeID] = i
		
		// Determine node group (for clustering visualization)
		group := ge.getNodeGroup(nodeID)
		
		// Calculate visual size based on importance
		size := ge.calculateNodeSize(node)
		
		// Determine color based on package type or risk
		color := nodeColors[node.PackageType]
		if color == "" {
			// Use risk-based coloring if package type is unknown
			riskLevel := ge.getRiskLevel(node.RiskScore)
			color = riskColors[riskLevel]
		}
		
		// Create label (truncate long names)
		label := node.Name
		if len(label) > 20 {
			label = label[:17] + "..."
		}
		
		d3Node := D3JSNode{
			ID:              nodeID,
			Name:            node.Name,
			Group:           group,
			Size:            size,
			Color:           color,
			Label:           label,
			Version:         node.Version,
			PackageType:     node.PackageType,
			RiskScore:       node.RiskScore,
			Vulnerabilities: node.VulnerabilityCount,
		}
		
		// Add fixed positions for hierarchical layout
		if layout == "hierarchical" {
			d3Node.X, d3Node.Y = ge.calculateHierarchicalPosition(node, i)
			d3Node.Fixed = true
		}
		
		nodes = append(nodes, d3Node)
	}
	
	// Generate links
	for sourceID, edges := range ge.graph.Edges {
		sourceIndex := nodeIndexMap[sourceID]
		
		for _, edge := range edges {
			targetIndex, exists := nodeIndexMap[edge.Target]
			if !exists {
				continue // Skip if target node doesn't exist
			}
			
			// Determine link color based on relationship type
			linkColor := ge.getLinkColor(edge.Relationship)
			
			// Calculate link distance for force-directed layout
			distance := ge.calculateLinkDistance(edge)
			
			d3Link := D3JSLink{
				Source:   sourceIndex,
				Target:   targetIndex,
				Value:    edge.Weight,
				Type:     edge.Relationship,
				Distance: distance,
				Color:    linkColor,
			}
			
			links = append(links, d3Link)
		}
	}
	
	// Generate metadata
	meta := D3JSMeta{
		Title:       fmt.Sprintf("Dependency Graph - %s", ge.getProjectName()),
		Description: fmt.Sprintf("Interactive dependency graph with %d packages and %d relationships", len(nodes), len(links)),
		NodeCount:   len(nodes),
		LinkCount:   len(links),
		Layout:      layout,
		Options:     ge.generateLayoutOptions(layout),
		GeneratedAt: time.Now(),
	}
	
	return &D3JSVisualizationData{
		Nodes: nodes,
		Links: links,
		Meta:  meta,
	}, nil
}

// getSortedNodeIDs returns node IDs sorted by importance
func (ge *GraphExporter) getSortedNodeIDs() []string {
	type nodeWeight struct {
		id     string
		weight float64
	}
	
	weights := make([]nodeWeight, 0, len(ge.graph.Nodes))
	for id, node := range ge.graph.Nodes {
		weights = append(weights, nodeWeight{id: id, weight: node.Weight})
	}
	
	// Sort by weight (descending)
	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight > weights[j].weight
	})
	
	result := make([]string, len(weights))
	for i, w := range weights {
		result[i] = w.id
	}
	
	return result
}

// getNodeGroup determines the cluster group for a node
func (ge *GraphExporter) getNodeGroup(nodeID string) int {
	// Find which cluster this node belongs to
	for i, cluster := range ge.graph.Stats.Clusters {
		for _, packageID := range cluster.Packages {
			if packageID == nodeID {
				return i
			}
		}
	}
	
	// Default group based on package type
	node := ge.graph.Nodes[nodeID]
	switch node.PackageType {
	case "dependencies":
		return 0
	case "devDependencies":
		return 1
	case "peerDependencies":
		return 2
	case "optionalDependencies":
		return 3
	default:
		return 4
	}
}

// calculateNodeSize calculates the visual size of a node
func (ge *GraphExporter) calculateNodeSize(node *GraphNode) float64 {
	// Base size
	baseSize := 5.0
	
	// Size based on weight (importance)
	weightSize := math.Log(node.Weight + 1) * 3.0
	
	// Size based on risk (higher risk = larger size for visibility)
	riskSize := node.RiskScore * 0.5
	
	// Vulnerability size bonus
	vulnSize := float64(node.VulnerabilityCount) * 0.3
	
	total := baseSize + weightSize + riskSize + vulnSize
	
	// Clamp between 3 and 20
	return math.Max(3.0, math.Min(20.0, total))
}

// getRiskLevel converts risk score to categorical risk level
func (ge *GraphExporter) getRiskLevel(riskScore float64) string {
	if riskScore >= 8.0 {
		return "critical"
	} else if riskScore >= 6.0 {
		return "high"
	} else if riskScore >= 3.0 {
		return "medium"
	}
	return "low"
}

// calculateHierarchicalPosition calculates fixed positions for hierarchical layout
func (ge *GraphExporter) calculateHierarchicalPosition(node *GraphNode, index int) (x, y float64) {
	// Arrange nodes in levels based on depth
	levelWidth := 800.0 // Total width for each level
	levelHeight := 100.0 // Vertical spacing between levels
	
	// Count nodes at this depth level
	nodesAtDepth := 0
	currentNodeIndex := 0
	
	for _, otherNode := range ge.graph.Nodes {
		if otherNode.Depth == node.Depth {
			if otherNode == node {
				currentNodeIndex = nodesAtDepth
			}
			nodesAtDepth++
		}
	}
	
	// Calculate x position (spread evenly across level width)
	if nodesAtDepth > 1 {
		x = (float64(currentNodeIndex) / float64(nodesAtDepth-1)) * levelWidth
	} else {
		x = levelWidth / 2
	}
	
	// Calculate y position based on depth
	y = float64(node.Depth) * levelHeight
	
	return x, y
}

// getLinkColor determines the color for a link based on relationship type
func (ge *GraphExporter) getLinkColor(relationship string) string {
	colors := map[string]string{
		"dependencies":         "#2C3E50", // Dark blue
		"devDependencies":      "#27AE60", // Green
		"peerDependencies":     "#E67E22", // Orange
		"optionalDependencies": "#9B59B6", // Purple
	}
	
	if color, exists := colors[relationship]; exists {
		return color
	}
	return "#BDC3C7" // Light gray for unknown
}

// calculateLinkDistance calculates the ideal distance for a link
func (ge *GraphExporter) calculateLinkDistance(edge *GraphEdge) float64 {
	// Base distance
	baseDistance := 50.0
	
	// Adjust based on relationship type
	typeMultiplier := map[string]float64{
		"dependencies":         1.0, // Normal distance
		"devDependencies":      1.2, // Slightly longer
		"peerDependencies":     0.8, // Closer (peer relationships)
		"optionalDependencies": 1.5, // Longer (less important)
	}
	
	multiplier := typeMultiplier[edge.Relationship]
	if multiplier == 0 {
		multiplier = 1.0
	}
	
	// Adjust based on edge weight (stronger relationships are closer)
	weightMultiplier := 1.0 / (edge.Weight + 0.1)
	
	return baseDistance * multiplier * weightMultiplier
}

// generateLayoutOptions generates layout-specific options
func (ge *GraphExporter) generateLayoutOptions(layout string) map[string]interface{} {
	options := make(map[string]interface{})
	
	switch layout {
	case "force":
		options["charge"] = -300
		options["linkDistance"] = 50
		options["gravity"] = 0.1
		options["friction"] = 0.9
		
	case "hierarchical":
		options["rankDir"] = "TB" // Top to bottom
		options["nodeSpacing"] = 50
		options["levelSpacing"] = 100
		
	case "circular":
		options["radius"] = 200
		options["spacing"] = 10
	}
	
	// Common options
	options["width"] = 1200
	options["height"] = 800
	options["showLabels"] = true
	options["showLegend"] = true
	options["allowZoom"] = true
	options["allowPan"] = true
	
	return options
}

// getProjectName attempts to determine the project name
func (ge *GraphExporter) getProjectName() string {
	// Look for root package (depth 0, production dependency)
	for _, node := range ge.graph.Nodes {
		if node.Depth == 0 && node.PackageType == "dependencies" {
			return node.Name
		}
	}
	
	// Fallback to first node
	for _, node := range ge.graph.Nodes {
		return node.Name
	}
	
	return "Unknown Project"
}

// ExportDOT exports the graph in DOT format for Graphviz
func (ge *GraphExporter) ExportDOT() (string, error) {
	var builder strings.Builder
	
	builder.WriteString("digraph DependencyGraph {\n")
	builder.WriteString("  rankdir=TB;\n")
	builder.WriteString("  node [shape=box, style=filled];\n")
	builder.WriteString("  edge [color=gray];\n\n")
	
	// Export nodes
	for nodeID, node := range ge.graph.Nodes {
		color := ge.getDotColor(node.PackageType)
		label := fmt.Sprintf("%s\\n%s", node.Name, node.Version)
		
		builder.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", fillcolor=\"%s\"];\n", 
			nodeID, label, color))
	}
	
	builder.WriteString("\n")
	
	// Export edges
	for sourceID, edges := range ge.graph.Edges {
		for _, edge := range edges {
			style := ge.getDotEdgeStyle(edge.Relationship)
			builder.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n", 
				sourceID, edge.Target, style))
		}
	}
	
	builder.WriteString("}\n")
	
	return builder.String(), nil
}

// getDotColor returns DOT format color for node types
func (ge *GraphExporter) getDotColor(packageType string) string {
	colors := map[string]string{
		"dependencies":         "lightblue",
		"devDependencies":      "lightgreen",
		"peerDependencies":     "orange",
		"optionalDependencies": "plum",
	}
	
	if color, exists := colors[packageType]; exists {
		return color
	}
	return "lightgray"
}

// getDotEdgeStyle returns DOT format edge style for relationship types
func (ge *GraphExporter) getDotEdgeStyle(relationship string) string {
	styles := map[string]string{
		"dependencies":         "style=solid, color=blue",
		"devDependencies":      "style=dashed, color=green",
		"peerDependencies":     "style=dotted, color=orange",
		"optionalDependencies": "style=dashed, color=purple",
	}
	
	if style, exists := styles[relationship]; exists {
		return style
	}
	return "style=solid, color=gray"
}

// ExportGraphML exports the graph in GraphML format
func (ge *GraphExporter) ExportGraphML() (string, error) {
	var builder strings.Builder
	
	builder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	builder.WriteString("<graphml xmlns=\"http://graphml.graphdrawing.org/xmlns\" ")
	builder.WriteString("xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" ")
	builder.WriteString("xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns ")
	builder.WriteString("http://graphml.graphdrawing.org/xmlns/1.0/graphml.xsd\">\n")
	
	// Define attributes
	builder.WriteString("  <key id=\"name\" for=\"node\" attr.name=\"name\" attr.type=\"string\"/>\n")
	builder.WriteString("  <key id=\"version\" for=\"node\" attr.name=\"version\" attr.type=\"string\"/>\n")
	builder.WriteString("  <key id=\"type\" for=\"node\" attr.name=\"packageType\" attr.type=\"string\"/>\n")
	builder.WriteString("  <key id=\"weight\" for=\"node\" attr.name=\"weight\" attr.type=\"double\"/>\n")
	builder.WriteString("  <key id=\"relationship\" for=\"edge\" attr.name=\"relationship\" attr.type=\"string\"/>\n")
	builder.WriteString("  <key id=\"edgeWeight\" for=\"edge\" attr.name=\"weight\" attr.type=\"double\"/>\n")
	
	builder.WriteString("  <graph id=\"DependencyGraph\" edgedefault=\"directed\">\n")
	
	// Export nodes
	for nodeID, node := range ge.graph.Nodes {
		builder.WriteString(fmt.Sprintf("    <node id=\"%s\">\n", nodeID))
		builder.WriteString(fmt.Sprintf("      <data key=\"name\">%s</data>\n", node.Name))
		builder.WriteString(fmt.Sprintf("      <data key=\"version\">%s</data>\n", node.Version))
		builder.WriteString(fmt.Sprintf("      <data key=\"type\">%s</data>\n", node.PackageType))
		builder.WriteString(fmt.Sprintf("      <data key=\"weight\">%.2f</data>\n", node.Weight))
		builder.WriteString("    </node>\n")
	}
	
	// Export edges
	edgeID := 0
	for sourceID, edges := range ge.graph.Edges {
		for _, edge := range edges {
			builder.WriteString(fmt.Sprintf("    <edge id=\"e%d\" source=\"%s\" target=\"%s\">\n", 
				edgeID, sourceID, edge.Target))
			builder.WriteString(fmt.Sprintf("      <data key=\"relationship\">%s</data>\n", edge.Relationship))
			builder.WriteString(fmt.Sprintf("      <data key=\"edgeWeight\">%.2f</data>\n", edge.Weight))
			builder.WriteString("    </edge>\n")
			edgeID++
		}
	}
	
	builder.WriteString("  </graph>\n")
	builder.WriteString("</graphml>\n")
	
	return builder.String(), nil
}

// ExportJSON exports the full graph data as JSON
func (ge *GraphExporter) ExportJSON() (string, error) {
	data, err := json.MarshalIndent(ge.graph, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal graph to JSON: %w", err)
	}
	
	return string(data), nil
}

// ExportSummaryJSON exports a summary of the graph analysis
func (ge *GraphExporter) ExportSummaryJSON() (string, error) {
	summary := map[string]interface{}{
		"project_name":         ge.getProjectName(),
		"total_packages":       len(ge.graph.Nodes),
		"total_dependencies":   ge.graph.Stats.TotalEdges,
		"max_depth":           ge.graph.Stats.MaxDepth,
		"circular_dependencies": len(ge.graph.Stats.CircularDeps),
		"critical_packages":    len(ge.graph.Stats.CriticalNodes),
		"clusters":            len(ge.graph.Stats.Clusters),
		"metrics":             ge.graph.Stats.Metrics,
		"generated_at":        ge.graph.Stats.GeneratedAt,
	}
	
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal summary to JSON: %w", err)
	}
	
	return string(data), nil
}