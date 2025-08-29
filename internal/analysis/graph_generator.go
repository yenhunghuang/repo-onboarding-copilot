package analysis

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GraphGenerator creates component relationship graphs and dependency visualizations
type GraphGenerator struct {
	componentIdentifier *ComponentIdentifier
	dataFlowAnalyzer    *DataFlowAnalyzer
	cycleDetector       *CycleDetector
	nodes               map[string]*ComponentGraphNode
	edges               []*ComponentGraphEdge
	clusters            map[string]*ComponentGraphCluster
}

// ComponentGraphNode represents a node in the component relationship graph
type ComponentGraphNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       GraphNodeType          `json:"type"`
	FilePath   string                 `json:"file_path"`
	Size       int                    `json:"size"`       // Lines of code or complexity measure
	Importance float64                `json:"importance"` // Importance score (0-1)
	Complexity float64                `json:"complexity"` // Complexity score (0-1)
	Position   *GraphPosition         `json:"position"`   // Visual position
	Style      *GraphNodeStyle        `json:"style"`      // Visual styling
	Metadata   map[string]interface{} `json:"metadata"`
}

// ComponentGraphEdge represents a relationship between graph nodes
type ComponentGraphEdge struct {
	ID        string                 `json:"id"`
	Source    string                 `json:"source"` // Source node ID
	Target    string                 `json:"target"` // Target node ID
	Type      GraphEdgeType          `json:"type"`
	Weight    float64                `json:"weight"` // Relationship strength (0-1)
	Direction EdgeDirection          `json:"direction"`
	Label     string                 `json:"label"`
	Style     *GraphEdgeStyle        `json:"style"` // Visual styling
	Metadata  map[string]interface{} `json:"metadata"`
}

// ComponentGraphCluster represents a group of related nodes
type ComponentGraphCluster struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        GraphClusterType       `json:"type"`
	NodeIDs     []string               `json:"node_ids"`
	Color       string                 `json:"color"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ComponentGraph represents the complete component relationship graph
type ComponentGraph struct {
	Nodes    []*ComponentGraphNode    `json:"nodes"`
	Edges    []*ComponentGraphEdge    `json:"edges"`
	Clusters []*ComponentGraphCluster `json:"clusters"`
	Metadata *GraphMetadata           `json:"metadata"`
	Layout   *GraphLayout             `json:"layout"`
	Stats    *ComponentGraphStats     `json:"stats"`
}

// Supporting types and enums
type GraphNodeType string

const (
	ComponentNode GraphNodeType = "component"
	ServiceNode   GraphNodeType = "service"
	UtilityNode   GraphNodeType = "utility"
	ConfigNode    GraphNodeType = "config"
	TestNode      GraphNodeType = "test"
	AssetNode     GraphNodeType = "asset"
)

type GraphEdgeType string

const (
	ImportEdge    GraphEdgeType = "import"    // Direct import relationship
	ComponentEdge GraphEdgeType = "component" // Component usage
	FlowEdge      GraphEdgeType = "dataflow"  // Data flow relationship
	APIEdge       GraphEdgeType = "api"       // API call relationship
	EventEdge     GraphEdgeType = "event"     // Event-based relationship
	ConfigEdge    GraphEdgeType = "config"    // Configuration dependency
)

type EdgeDirection string

const (
	DirectionalEdge   EdgeDirection = "directional"   // A -> B
	BidirectionalEdge EdgeDirection = "bidirectional" // A <-> B
	UndirectedEdge    EdgeDirection = "undirected"    // A -- B
)

type GraphClusterType string

const (
	FeatureCluster GraphClusterType = "feature" // Feature-based grouping
	LayerCluster   GraphClusterType = "layer"   // Architectural layer
	ModuleCluster  GraphClusterType = "module"  // Module/package grouping
	DomainCluster  GraphClusterType = "domain"  // Domain-driven grouping
	CycleCluster   GraphClusterType = "cycle"   // Circular dependency group
)

type GraphPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z,omitempty"` // For 3D layouts
}

type GraphNodeStyle struct {
	Color       string  `json:"color"`
	Size        float64 `json:"size"`
	Shape       string  `json:"shape"`
	BorderColor string  `json:"border_color"`
	BorderWidth float64 `json:"border_width"`
	Opacity     float64 `json:"opacity"`
}

type GraphEdgeStyle struct {
	Color     string  `json:"color"`
	Width     float64 `json:"width"`
	Style     string  `json:"style"` // solid, dashed, dotted
	Opacity   float64 `json:"opacity"`
	Curvature float64 `json:"curvature"` // For curved edges
}

type GraphLayout struct {
	Algorithm string                 `json:"algorithm"` // force-directed, hierarchical, circular
	Width     float64                `json:"width"`
	Height    float64                `json:"height"`
	Options   map[string]interface{} `json:"options"`
}

type GraphMetadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Version     string `json:"version"`
	ProjectPath string `json:"project_path"`
}

type ComponentGraphStats struct {
	TotalNodes          int                      `json:"total_nodes"`
	TotalEdges          int                      `json:"total_edges"`
	TotalClusters       int                      `json:"total_clusters"`
	NodesByType         map[GraphNodeType]int    `json:"nodes_by_type"`
	EdgesByType         map[GraphEdgeType]int    `json:"edges_by_type"`
	ClustersByType      map[GraphClusterType]int `json:"clusters_by_type"`
	AverageConnectivity float64                  `json:"average_connectivity"`
	MaxDepth            int                      `json:"max_depth"`
	CriticalPaths       []string                 `json:"critical_paths"`
}

// NewGraphGenerator creates a new graph generator
func NewGraphGenerator(ci *ComponentIdentifier, dfa *DataFlowAnalyzer, cd *CycleDetector) *GraphGenerator {
	return &GraphGenerator{
		componentIdentifier: ci,
		dataFlowAnalyzer:    dfa,
		cycleDetector:       cd,
		nodes:               make(map[string]*ComponentGraphNode),
		edges:               make([]*ComponentGraphEdge, 0),
		clusters:            make(map[string]*ComponentGraphCluster),
	}
}

// GenerateGraph creates a comprehensive component relationship graph
func (gg *GraphGenerator) GenerateGraph(filePath, content string) error {
	// Create node for the current file
	if err := gg.createNode(filePath, content); err != nil {
		return fmt.Errorf("error creating node: %w", err)
	}

	// Analyze relationships and create edges
	if err := gg.analyzeRelationships(filePath, content); err != nil {
		return fmt.Errorf("error analyzing relationships: %w", err)
	}

	return nil
}

// createNode creates a graph node from a file
func (gg *GraphGenerator) createNode(filePath, content string) error {
	// Identify component type
	component, err := gg.componentIdentifier.IdentifyComponent(filePath, content)
	if err != nil {
		return fmt.Errorf("error identifying component: %w", err)
	}

	// Calculate node metrics
	size := gg.calculateSize(content)
	importance := gg.calculateImportance(filePath, content)
	complexity := gg.calculateComplexity(content)

	// Determine node type
	nodeType := gg.mapComponentToNodeType(component.Type)

	// Create node
	node := &ComponentGraphNode{
		ID:         gg.generateNodeID(filePath),
		Name:       gg.extractNodeName(filePath),
		Type:       nodeType,
		FilePath:   filePath,
		Size:       size,
		Importance: importance,
		Complexity: complexity,
		Style:      gg.createNodeStyle(nodeType, importance, complexity),
		Metadata: map[string]interface{}{
			"component_type": component.Type,
			"exports":        component.Exports,
			"dependencies":   component.Dependencies,
		},
	}

	gg.nodes[node.ID] = node
	return nil
}

// analyzeRelationships analyzes relationships between components
func (gg *GraphGenerator) analyzeRelationships(filePath, content string) error {
	currentNodeID := gg.generateNodeID(filePath)

	// Analyze import relationships
	if err := gg.analyzeImportRelationships(filePath, content, currentNodeID); err != nil {
		return fmt.Errorf("error analyzing imports: %w", err)
	}

	// Analyze data flow relationships
	if err := gg.analyzeDataFlowRelationships(filePath, content, currentNodeID); err != nil {
		return fmt.Errorf("error analyzing data flow: %w", err)
	}

	// Analyze component usage relationships
	if err := gg.analyzeComponentRelationships(filePath, content, currentNodeID); err != nil {
		return fmt.Errorf("error analyzing components: %w", err)
	}

	return nil
}

// analyzeImportRelationships creates edges for import dependencies
func (gg *GraphGenerator) analyzeImportRelationships(filePath, content, currentNodeID string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// ES6 imports
		if strings.Contains(line, "import ") && strings.Contains(line, "from ") {
			if targetPath := gg.extractImportPath(line, filePath); targetPath != "" {
				targetNodeID := gg.generateNodeID(targetPath)
				gg.createEdge(currentNodeID, targetNodeID, ImportEdge, 0.8, DirectionalEdge,
					fmt.Sprintf("imports from %s", targetPath), lineNum+1)
			}
		}

		// CommonJS requires
		if strings.Contains(line, "require(") {
			if targetPath := gg.extractRequirePath(line, filePath); targetPath != "" {
				targetNodeID := gg.generateNodeID(targetPath)
				gg.createEdge(currentNodeID, targetNodeID, ImportEdge, 0.7, DirectionalEdge,
					fmt.Sprintf("requires %s", targetPath), lineNum+1)
			}
		}
	}

	return nil
}

// analyzeDataFlowRelationships creates edges for data flow
func (gg *GraphGenerator) analyzeDataFlowRelationships(filePath, content, currentNodeID string) error {
	// Use data flow analyzer to get flow information
	if err := gg.dataFlowAnalyzer.AnalyzeDataFlow(filePath, content); err != nil {
		return err
	}

	// Get data flow graph and create edges
	flowGraph := gg.dataFlowAnalyzer.GetDataFlowGraph()
	if flowGraph.Nodes != nil && len(flowGraph.Nodes) > 0 {
		for _, edge := range flowGraph.Edges {
			// Create data flow edge if it's between different files
			if gg.isDifferentFile(edge.From, edge.To, filePath) {
				fromNodeID := gg.resolveNodeID(edge.From, filePath)
				toNodeID := gg.resolveNodeID(edge.To, filePath)

				if fromNodeID != "" && toNodeID != "" {
					gg.createEdge(fromNodeID, toNodeID, FlowEdge, 0.6, DirectionalEdge,
						fmt.Sprintf("data flow: %s", edge.FlowType), 0)
				}
			}
		}
	}

	return nil
}

// analyzeComponentRelationships creates edges for component usage
func (gg *GraphGenerator) analyzeComponentRelationships(filePath, content, currentNodeID string) error {
	// Detect component usage patterns
	if gg.containsJSX(content) {
		// React component relationships
		componentRefs := gg.extractComponentReferences(content)
		for _, ref := range componentRefs {
			targetNodeID := gg.resolveComponentNodeID(ref, filePath)
			if targetNodeID != "" && targetNodeID != currentNodeID {
				gg.createEdge(currentNodeID, targetNodeID, ComponentEdge, 0.9, DirectionalEdge,
					fmt.Sprintf("uses component %s", ref), 0)
			}
		}
	}

	// API call relationships
	apiCalls := gg.extractAPICalls(content)
	for _, apiCall := range apiCalls {
		// Create virtual API node if needed
		apiNodeID := gg.createOrGetAPINode(apiCall)
		gg.createEdge(currentNodeID, apiNodeID, APIEdge, 0.5, DirectionalEdge,
			fmt.Sprintf("calls %s", apiCall), 0)
	}

	return nil
}

// createEdge creates a graph edge between two nodes
func (gg *GraphGenerator) createEdge(sourceID, targetID string, edgeType GraphEdgeType,
	weight float64, direction EdgeDirection, label string, lineNumber int) {

	edgeID := fmt.Sprintf("%s-%s-%s", sourceID, targetID, string(edgeType))

	// Check if edge already exists
	for _, edge := range gg.edges {
		if edge.ID == edgeID {
			// Update existing edge weight
			edge.Weight = (edge.Weight + weight) / 2
			return
		}
	}

	edge := &ComponentGraphEdge{
		ID:        edgeID,
		Source:    sourceID,
		Target:    targetID,
		Type:      edgeType,
		Weight:    weight,
		Direction: direction,
		Label:     label,
		Style:     gg.createEdgeStyle(edgeType, weight),
		Metadata: map[string]interface{}{
			"line_number": lineNumber,
			"created_by":  "graph_generator",
		},
	}

	gg.edges = append(gg.edges, edge)
}

// GenerateClusters creates logical groupings of nodes
func (gg *GraphGenerator) GenerateClusters() error {
	// Feature-based clustering
	if err := gg.createFeatureClusters(); err != nil {
		return fmt.Errorf("error creating feature clusters: %w", err)
	}

	// Layer-based clustering
	if err := gg.createLayerClusters(); err != nil {
		return fmt.Errorf("error creating layer clusters: %w", err)
	}

	// Cycle-based clustering
	if err := gg.createCycleClusters(); err != nil {
		return fmt.Errorf("error creating cycle clusters: %w", err)
	}

	return nil
}

// createFeatureClusters groups nodes by feature/domain
func (gg *GraphGenerator) createFeatureClusters() error {
	featureGroups := make(map[string][]string)

	// Group nodes by directory structure (assuming feature-based organization)
	for nodeID, node := range gg.nodes {
		feature := gg.extractFeature(node.FilePath)
		if featureGroups[feature] == nil {
			featureGroups[feature] = make([]string, 0)
		}
		featureGroups[feature] = append(featureGroups[feature], nodeID)
	}

	// Create clusters for each feature
	for feature, nodeIDs := range featureGroups {
		if len(nodeIDs) > 1 { // Only create clusters with multiple nodes
			cluster := &ComponentGraphCluster{
				ID:          fmt.Sprintf("feature-%s", feature),
				Name:        fmt.Sprintf("Feature: %s", feature),
				Type:        FeatureCluster,
				NodeIDs:     nodeIDs,
				Color:       gg.generateFeatureColor(feature),
				Description: fmt.Sprintf("Feature cluster containing %d components", len(nodeIDs)),
				Metadata: map[string]interface{}{
					"feature":    feature,
					"node_count": len(nodeIDs),
				},
			}
			gg.clusters[cluster.ID] = cluster
		}
	}

	return nil
}

// createLayerClusters groups nodes by architectural layer
func (gg *GraphGenerator) createLayerClusters() error {
	layerGroups := make(map[string][]string)

	// Group nodes by architectural layer
	for nodeID, node := range gg.nodes {
		layer := gg.determineArchitecturalLayer(node.FilePath, node.Type)
		if layerGroups[layer] == nil {
			layerGroups[layer] = make([]string, 0)
		}
		layerGroups[layer] = append(layerGroups[layer], nodeID)
	}

	// Create clusters for each layer
	layerColors := map[string]string{
		"presentation":   "#4CAF50",
		"business":       "#2196F3",
		"data":           "#FF9800",
		"infrastructure": "#9C27B0",
	}

	for layer, nodeIDs := range layerGroups {
		if len(nodeIDs) > 1 {
			cluster := &ComponentGraphCluster{
				ID:          fmt.Sprintf("layer-%s", layer),
				Name:        fmt.Sprintf("Layer: %s", layer),
				Type:        LayerCluster,
				NodeIDs:     nodeIDs,
				Color:       layerColors[layer],
				Description: fmt.Sprintf("Architectural layer containing %d components", len(nodeIDs)),
				Metadata: map[string]interface{}{
					"layer":      layer,
					"node_count": len(nodeIDs),
				},
			}
			gg.clusters[cluster.ID] = cluster
		}
	}

	return nil
}

// createCycleClusters groups nodes involved in dependency cycles
func (gg *GraphGenerator) createCycleClusters() error {
	cycles := gg.cycleDetector.GetCycles()

	for i, cycle := range cycles {
		nodeIDs := make([]string, 0)
		for _, filePath := range cycle.Files {
			nodeID := gg.generateNodeID(filePath)
			if _, exists := gg.nodes[nodeID]; exists {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}

		if len(nodeIDs) > 1 {
			cluster := &ComponentGraphCluster{
				ID:          fmt.Sprintf("cycle-%d", i),
				Name:        fmt.Sprintf("Dependency Cycle %d", i+1),
				Type:        CycleCluster,
				NodeIDs:     nodeIDs,
				Color:       "#F44336", // Red for cycles
				Description: fmt.Sprintf("Circular dependency involving %d components", len(nodeIDs)),
				Metadata: map[string]interface{}{
					"cycle_id":   cycle.ID,
					"cycle_type": cycle.Type,
					"severity":   cycle.Severity,
					"node_count": len(nodeIDs),
				},
			}
			gg.clusters[cluster.ID] = cluster
		}
	}

	return nil
}

// GetGraph returns the complete generated graph
func (gg *GraphGenerator) GetGraph() *ComponentGraph {
	// Convert maps to slices
	nodes := make([]*ComponentGraphNode, 0, len(gg.nodes))
	for _, node := range gg.nodes {
		nodes = append(nodes, node)
	}

	clusters := make([]*ComponentGraphCluster, 0, len(gg.clusters))
	for _, cluster := range gg.clusters {
		clusters = append(clusters, cluster)
	}

	// Generate layout
	layout := gg.generateLayout()

	// Calculate statistics
	stats := gg.calculateGraphStats(nodes, gg.edges, clusters)

	return &ComponentGraph{
		Nodes:    nodes,
		Edges:    gg.edges,
		Clusters: clusters,
		Layout:   layout,
		Stats:    stats,
		Metadata: &GraphMetadata{
			Title:       "Component Relationship Graph",
			Description: "Generated component dependency and relationship graph",
			Version:     "1.0",
		},
	}
}

// Helper functions

// calculateSize estimates the size/complexity of a component
func (gg *GraphGenerator) calculateSize(content string) int {
	return len(strings.Split(content, "\n"))
}

// calculateImportance calculates the importance score of a component
func (gg *GraphGenerator) calculateImportance(filePath, content string) float64 {
	score := 0.5 // Base score

	// Boost for certain file types
	if strings.Contains(filePath, "index.") {
		score += 0.2 // Entry points are important
	}
	if strings.Contains(filePath, "App.") || strings.Contains(filePath, "main.") {
		score += 0.3 // Main application files
	}

	// Boost for larger files (more functionality)
	lines := len(strings.Split(content, "\n"))
	if lines > 100 {
		score += 0.1
	}
	if lines > 500 {
		score += 0.1
	}

	// Boost for files with many exports
	exports := strings.Count(content, "export ")
	score += float64(exports) * 0.02

	if score > 1.0 {
		score = 1.0
	}
	return score
}

// calculateComplexity estimates the complexity of a component
func (gg *GraphGenerator) calculateComplexity(content string) float64 {
	complexity := 0.0

	// Count control structures
	complexity += float64(strings.Count(content, " if ")) * 0.1
	complexity += float64(strings.Count(content, " for ")) * 0.1
	complexity += float64(strings.Count(content, " while ")) * 0.1
	complexity += float64(strings.Count(content, " switch ")) * 0.2
	complexity += float64(strings.Count(content, " try ")) * 0.1

	// Count function definitions
	complexity += float64(strings.Count(content, "function ")) * 0.05
	complexity += float64(strings.Count(content, "=> ")) * 0.05

	// Count async patterns
	complexity += float64(strings.Count(content, "async ")) * 0.1
	complexity += float64(strings.Count(content, "await ")) * 0.05

	if complexity > 1.0 {
		complexity = 1.0
	}
	return complexity
}

// mapComponentToNodeType maps component types to graph node types
func (gg *GraphGenerator) mapComponentToNodeType(componentType ComponentType) GraphNodeType {
	switch componentType {
	case ReactComponent:
		return ComponentNode
	case Service:
		return ServiceNode
	case Utility:
		return UtilityNode
	case Configuration:
		return ConfigNode
	case Middleware:
		return ServiceNode
	default:
		return ComponentNode
	}
}

// generateNodeID creates a unique ID for a node
func (gg *GraphGenerator) generateNodeID(filePath string) string {
	// Use file path as base, clean it up for ID
	id := strings.ReplaceAll(filePath, "/", "_")
	id = strings.ReplaceAll(id, ".", "_")
	return id
}

// extractNodeName extracts a display name from file path
func (gg *GraphGenerator) extractNodeName(filePath string) string {
	parts := strings.Split(filePath, "/")
	name := parts[len(parts)-1]

	// Remove extension
	if dotIndex := strings.LastIndex(name, "."); dotIndex != -1 {
		name = name[:dotIndex]
	}

	return name
}

// Additional helper functions would be implemented here...
// (createNodeStyle, createEdgeStyle, extractImportPath, etc.)

// Placeholder implementations for remaining helper functions
func (gg *GraphGenerator) createNodeStyle(nodeType GraphNodeType, importance, complexity float64) *GraphNodeStyle {
	colorMap := map[GraphNodeType]string{
		ComponentNode: "#4CAF50",
		ServiceNode:   "#2196F3",
		UtilityNode:   "#FF9800",
		ConfigNode:    "#9C27B0",
		TestNode:      "#607D8B",
		AssetNode:     "#795548",
	}

	return &GraphNodeStyle{
		Color:       colorMap[nodeType],
		Size:        10 + (importance * 20),
		Shape:       "circle",
		BorderColor: "#000000",
		BorderWidth: 1.0,
		Opacity:     0.8 + (complexity * 0.2),
	}
}

func (gg *GraphGenerator) createEdgeStyle(edgeType GraphEdgeType, weight float64) *GraphEdgeStyle {
	colorMap := map[GraphEdgeType]string{
		ImportEdge:    "#333333",
		ComponentEdge: "#4CAF50",
		FlowEdge:      "#2196F3",
		APIEdge:       "#FF9800",
		EventEdge:     "#9C27B0",
		ConfigEdge:    "#607D8B",
	}

	return &GraphEdgeStyle{
		Color:     colorMap[edgeType],
		Width:     1 + (weight * 3),
		Style:     "solid",
		Opacity:   0.6 + (weight * 0.4),
		Curvature: 0.2,
	}
}

// Simplified implementations for other helper methods
func (gg *GraphGenerator) extractImportPath(line, currentFile string) string {
	// Simplified import path extraction
	if idx := strings.Index(line, "from '"); idx != -1 {
		start := idx + 6
		if end := strings.Index(line[start:], "'"); end != -1 {
			return line[start : start+end]
		}
	}
	return ""
}

func (gg *GraphGenerator) extractRequirePath(line, currentFile string) string {
	// Simplified require path extraction
	if idx := strings.Index(line, "require('"); idx != -1 {
		start := idx + 9
		if end := strings.Index(line[start:], "'"); end != -1 {
			return line[start : start+end]
		}
	}
	return ""
}

func (gg *GraphGenerator) isDifferentFile(from, to, currentFile string) bool {
	return true // Simplified - assume different files
}

func (gg *GraphGenerator) resolveNodeID(reference, currentFile string) string {
	return gg.generateNodeID(reference) // Simplified resolution
}

func (gg *GraphGenerator) containsJSX(content string) bool {
	return strings.Contains(content, "<") && strings.Contains(content, ">")
}

func (gg *GraphGenerator) extractComponentReferences(content string) []string {
	// Simplified component reference extraction
	return []string{}
}

func (gg *GraphGenerator) resolveComponentNodeID(ref, currentFile string) string {
	return gg.generateNodeID(ref)
}

func (gg *GraphGenerator) extractAPICalls(content string) []string {
	calls := make([]string, 0)
	if strings.Contains(content, "fetch(") {
		calls = append(calls, "fetch-api")
	}
	if strings.Contains(content, "axios.") {
		calls = append(calls, "axios-api")
	}
	return calls
}

func (gg *GraphGenerator) createOrGetAPINode(apiCall string) string {
	nodeID := fmt.Sprintf("api-%s", apiCall)
	if _, exists := gg.nodes[nodeID]; !exists {
		gg.nodes[nodeID] = &ComponentGraphNode{
			ID:   nodeID,
			Name: apiCall,
			Type: ServiceNode,
			Style: &GraphNodeStyle{
				Color: "#FF5722",
				Size:  8,
				Shape: "diamond",
			},
		}
	}
	return nodeID
}

func (gg *GraphGenerator) extractFeature(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) > 2 {
		return parts[2] // Assume /src/feature/...
	}
	return "core"
}

func (gg *GraphGenerator) generateFeatureColor(feature string) string {
	colors := []string{"#E91E63", "#9C27B0", "#673AB7", "#3F51B5", "#2196F3", "#00BCD4", "#009688", "#4CAF50", "#8BC34A", "#CDDC39", "#FFEB3B", "#FFC107", "#FF9800", "#FF5722"}
	hash := 0
	for _, c := range feature {
		hash += int(c)
	}
	return colors[hash%len(colors)]
}

func (gg *GraphGenerator) determineArchitecturalLayer(filePath string, nodeType GraphNodeType) string {
	if strings.Contains(filePath, "component") || nodeType == ComponentNode {
		return "presentation"
	}
	if strings.Contains(filePath, "service") || nodeType == ServiceNode {
		return "business"
	}
	if strings.Contains(filePath, "data") || strings.Contains(filePath, "model") {
		return "data"
	}
	return "infrastructure"
}

func (gg *GraphGenerator) generateLayout() *GraphLayout {
	return &GraphLayout{
		Algorithm: "force-directed",
		Width:     1200,
		Height:    800,
		Options: map[string]interface{}{
			"iterations": 100,
			"cooling":    0.95,
		},
	}
}

func (gg *GraphGenerator) calculateGraphStats(nodes []*ComponentGraphNode, edges []*ComponentGraphEdge, clusters []*ComponentGraphCluster) *ComponentGraphStats {
	stats := &ComponentGraphStats{
		TotalNodes:     len(nodes),
		TotalEdges:     len(edges),
		TotalClusters:  len(clusters),
		NodesByType:    make(map[GraphNodeType]int),
		EdgesByType:    make(map[GraphEdgeType]int),
		ClustersByType: make(map[GraphClusterType]int),
	}

	// Count by types
	for _, node := range nodes {
		stats.NodesByType[node.Type]++
	}
	for _, edge := range edges {
		stats.EdgesByType[edge.Type]++
	}
	for _, cluster := range clusters {
		stats.ClustersByType[cluster.Type]++
	}

	// Calculate connectivity
	if len(nodes) > 0 {
		stats.AverageConnectivity = float64(len(edges)) / float64(len(nodes))
	}

	return stats
}

// ExportToJSON exports the graph to JSON format
func (gg *GraphGenerator) ExportToJSON() (string, error) {
	graph := gg.GetGraph()
	jsonData, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling graph to JSON: %w", err)
	}
	return string(jsonData), nil
}
