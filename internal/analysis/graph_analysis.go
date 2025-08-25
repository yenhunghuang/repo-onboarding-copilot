// Package analysis provides advanced graph analysis algorithms
package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// detectCircularDependencies detects circular dependencies in the graph using DFS
func (gb *GraphBuilder) detectCircularDependencies() {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	cycles := make([][]string, 0)
	
	// DFS for each unvisited node
	for nodeID := range gb.graph.Nodes {
		if !visited[nodeID] {
			path := make([]string, 0)
			gb.dfsCircularDetection(nodeID, visited, recursionStack, path, &cycles)
		}
	}
	
	// Convert cycles to CircularDependency structs
	for _, cycle := range cycles {
		circularDep := CircularDependency{
			Cycle:       cycle,
			Length:      len(cycle),
			Severity:    gb.calculateCycleSeverity(cycle),
			Impact:      gb.calculateCycleImpact(cycle),
			Description: fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " → ")),
		}
		gb.graph.Stats.CircularDeps = append(gb.graph.Stats.CircularDeps, circularDep)
	}
}

// dfsCircularDetection performs DFS to detect cycles
func (gb *GraphBuilder) dfsCircularDetection(node string, visited, recursionStack map[string]bool, path []string, cycles *[][]string) {
	visited[node] = true
	recursionStack[node] = true
	path = append(path, node)
	
	if edges, exists := gb.graph.Edges[node]; exists {
		for _, edge := range edges {
			target := edge.Target
			
			if recursionStack[target] {
				// Found a cycle - extract the cycle from the path
				cycleStart := -1
				for i, pathNode := range path {
					if pathNode == target {
						cycleStart = i
						break
					}
				}
				
				if cycleStart != -1 {
					cycle := make([]string, 0)
					cycle = append(cycle, path[cycleStart:]...)
					cycle = append(cycle, target) // Complete the cycle
					*cycles = append(*cycles, cycle)
				}
			} else if !visited[target] {
				gb.dfsCircularDetection(target, visited, recursionStack, path, cycles)
			}
		}
	}
	
	recursionStack[node] = false
}

// calculateCycleSeverity determines the severity of a circular dependency
func (gb *GraphBuilder) calculateCycleSeverity(cycle []string) string {
	if len(cycle) <= 2 {
		return "critical" // Direct circular dependency
	} else if len(cycle) <= 4 {
		return "high"
	} else if len(cycle) <= 6 {
		return "medium"
	}
	return "low"
}

// calculateCycleImpact calculates the impact score of a circular dependency
func (gb *GraphBuilder) calculateCycleImpact(cycle []string) float64 {
	totalWeight := 0.0
	nodeCount := 0
	
	for _, nodeID := range cycle {
		if node, exists := gb.graph.Nodes[nodeID]; exists {
			totalWeight += node.Weight
			nodeCount++
		}
	}
	
	if nodeCount == 0 {
		return 0.0
	}
	
	// Impact based on average node weight and cycle length
	avgWeight := totalWeight / float64(nodeCount)
	lengthPenalty := 1.0 / float64(len(cycle)) // Shorter cycles are more impactful
	
	return avgWeight * lengthPenalty * 10.0 // Scale to 0-10 range
}

// identifyCriticalNodes identifies nodes that are critical to the dependency graph
func (gb *GraphBuilder) identifyCriticalNodes() {
	type nodeScore struct {
		nodeID string
		score  float64
	}
	
	scores := make([]nodeScore, 0)
	
	for nodeID, node := range gb.graph.Nodes {
		score := gb.calculateNodeCriticality(nodeID, node)
		scores = append(scores, nodeScore{nodeID: nodeID, score: score})
	}
	
	// Sort by score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Take top 10% or at least top 5 as critical nodes
	criticalCount := int(math.Max(float64(len(scores))*0.1, 5))
	if criticalCount > len(scores) {
		criticalCount = len(scores)
	}
	
	gb.graph.Stats.CriticalNodes = make([]string, 0, criticalCount)
	for i := 0; i < criticalCount; i++ {
		gb.graph.Stats.CriticalNodes = append(gb.graph.Stats.CriticalNodes, scores[i].nodeID)
	}
}

// calculateNodeCriticality calculates how critical a node is to the graph
func (gb *GraphBuilder) calculateNodeCriticality(nodeID string, node *GraphNode) float64 {
	// Factors that make a node critical:
	// 1. High in-degree (many packages depend on it)
	// 2. High out-degree (it depends on many packages)
	// 3. Low depth (closer to root dependencies)
	// 4. High vulnerability count
	// 5. High risk score
	
	inDegree := 0
	outDegree := len(gb.graph.Edges[nodeID])
	
	// Count incoming edges
	for _, edges := range gb.graph.Edges {
		for _, edge := range edges {
			if edge.Target == nodeID {
				inDegree++
			}
		}
	}
	
	// Calculate criticality score
	score := 0.0
	
	// In-degree factor (0-5 points)
	score += math.Min(float64(inDegree), 5.0)
	
	// Out-degree factor (0-3 points)
	score += math.Min(float64(outDegree)*0.1, 3.0)
	
	// Depth factor (0-2 points, lower depth = higher score)
	maxDepth := float64(gb.graph.Stats.MaxDepth)
	if maxDepth > 0 {
		depthScore := (maxDepth - float64(node.Depth)) / maxDepth * 2.0
		score += depthScore
	}
	
	// Vulnerability factor (0-3 points)
	vulnScore := math.Min(float64(node.VulnerabilityCount)*0.5, 3.0)
	score += vulnScore
	
	// Risk factor (0-2 points)
	score += math.Min(node.RiskScore/5.0, 2.0)
	
	// Production dependency bonus
	if node.PackageType == "dependencies" {
		score += 1.0
	}
	
	return score
}

// performClusterAnalysis identifies clusters of tightly coupled dependencies
func (gb *GraphBuilder) performClusterAnalysis() {
	// Simple clustering based on shared dependencies
	clusters := make(map[string][]string)
	
	// Group nodes by their dependency patterns
	for nodeID, node := range gb.graph.Nodes {
		clusterKey := gb.generateClusterKey(nodeID)
		clusters[clusterKey] = append(clusters[clusterKey], nodeID)
		_ = node // Avoid unused variable warning
	}
	
	// Convert to DependencyCluster structs
	clusterID := 0
	for clusterKey, nodes := range clusters {
		if len(nodes) > 1 { // Only consider actual clusters
			cluster := DependencyCluster{
				ID:          fmt.Sprintf("cluster_%d", clusterID),
				Packages:    nodes,
				Cohesion:    gb.calculateClusterCohesion(nodes),
				Coupling:    gb.calculateClusterCoupling(nodes),
				MainPackage: gb.findMainPackageInCluster(nodes),
			}
			gb.graph.Stats.Clusters = append(gb.graph.Stats.Clusters, cluster)
			clusterID++
		}
		_ = clusterKey // Avoid unused variable warning
	}
}

// generateClusterKey generates a key for clustering based on dependencies
func (gb *GraphBuilder) generateClusterKey(nodeID string) string {
	dependencies := make([]string, 0)
	
	if edges, exists := gb.graph.Edges[nodeID]; exists {
		for _, edge := range edges {
			if edge.Relationship == "dependencies" {
				dependencies = append(dependencies, edge.Target)
			}
		}
	}
	
	sort.Strings(dependencies)
	return strings.Join(dependencies, ",")
}

// calculateClusterCohesion calculates internal connectivity of a cluster
func (gb *GraphBuilder) calculateClusterCohesion(nodes []string) float64 {
	if len(nodes) <= 1 {
		return 1.0
	}
	
	internalEdges := 0
	possibleEdges := len(nodes) * (len(nodes) - 1)
	
	for _, sourceNode := range nodes {
		if edges, exists := gb.graph.Edges[sourceNode]; exists {
			for _, edge := range edges {
				for _, targetNode := range nodes {
					if edge.Target == targetNode {
						internalEdges++
						break
					}
				}
			}
		}
	}
	
	if possibleEdges == 0 {
		return 1.0
	}
	
	return float64(internalEdges) / float64(possibleEdges)
}

// calculateClusterCoupling calculates external dependencies of a cluster
func (gb *GraphBuilder) calculateClusterCoupling(nodes []string) float64 {
	externalEdges := 0
	totalEdges := 0
	
	for _, sourceNode := range nodes {
		if edges, exists := gb.graph.Edges[sourceNode]; exists {
			totalEdges += len(edges)
			for _, edge := range edges {
				isInternal := false
				for _, targetNode := range nodes {
					if edge.Target == targetNode {
						isInternal = true
						break
					}
				}
				if !isInternal {
					externalEdges++
				}
			}
		}
	}
	
	if totalEdges == 0 {
		return 0.0
	}
	
	return float64(externalEdges) / float64(totalEdges)
}

// findMainPackageInCluster identifies the main package in a cluster
func (gb *GraphBuilder) findMainPackageInCluster(nodes []string) string {
	maxWeight := 0.0
	mainPackage := ""
	
	for _, nodeID := range nodes {
		if node, exists := gb.graph.Nodes[nodeID]; exists {
			if node.Weight > maxWeight {
				maxWeight = node.Weight
				mainPackage = nodeID
			}
		}
	}
	
	return mainPackage
}

// calculateGraphMetrics calculates various graph metrics
func (gb *GraphBuilder) calculateGraphMetrics() {
	metrics := gb.graph.Stats.Metrics
	
	// Calculate density
	nodeCount := len(gb.graph.Nodes)
	edgeCount := gb.graph.Stats.TotalEdges
	maxPossibleEdges := nodeCount * (nodeCount - 1)
	
	if maxPossibleEdges > 0 {
		metrics.Density = float64(edgeCount) / float64(maxPossibleEdges)
	}
	
	// Calculate connected components using Union-Find
	metrics.ConnectedComponents = gb.calculateConnectedComponents()
	
	// Calculate modularity (simplified version)
	metrics.Modularity = gb.calculateModularity()
	
	// Calculate maximum centrality
	metrics.CentralityMax = gb.calculateMaxCentrality()
	
	// Calculate average path length and diameter
	metrics.PathLengthAvg, metrics.Diameter = gb.calculatePathMetrics()
}

// calculateConnectedComponents calculates the number of connected components
func (gb *GraphBuilder) calculateConnectedComponents() int {
	visited := make(map[string]bool)
	components := 0
	
	for nodeID := range gb.graph.Nodes {
		if !visited[nodeID] {
			gb.dfsComponentSearch(nodeID, visited)
			components++
		}
	}
	
	return components
}

// dfsComponentSearch performs DFS for connected component analysis
func (gb *GraphBuilder) dfsComponentSearch(nodeID string, visited map[string]bool) {
	visited[nodeID] = true
	
	// Visit all connected nodes (both directions)
	if edges, exists := gb.graph.Edges[nodeID]; exists {
		for _, edge := range edges {
			if !visited[edge.Target] {
				gb.dfsComponentSearch(edge.Target, visited)
			}
		}
	}
	
	// Also check incoming edges
	for _, edges := range gb.graph.Edges {
		for _, edge := range edges {
			if edge.Target == nodeID && !visited[edge.Source] {
				gb.dfsComponentSearch(edge.Source, visited)
			}
		}
	}
}

// calculateModularity calculates a simplified modularity score
func (gb *GraphBuilder) calculateModularity() float64 {
	// Simplified modularity based on cluster quality
	if len(gb.graph.Stats.Clusters) == 0 {
		return 0.0
	}
	
	totalCohesion := 0.0
	for _, cluster := range gb.graph.Stats.Clusters {
		totalCohesion += cluster.Cohesion
	}
	
	return totalCohesion / float64(len(gb.graph.Stats.Clusters))
}

// calculateMaxCentrality calculates the maximum centrality score
func (gb *GraphBuilder) calculateMaxCentrality() float64 {
	maxCentrality := 0.0
	
	for nodeID := range gb.graph.Nodes {
		centrality := gb.calculateNodeCentrality(nodeID)
		if centrality > maxCentrality {
			maxCentrality = centrality
		}
	}
	
	return maxCentrality
}

// calculateNodeCentrality calculates centrality score for a node
func (gb *GraphBuilder) calculateNodeCentrality(nodeID string) float64 {
	inDegree := 0
	outDegree := len(gb.graph.Edges[nodeID])
	
	for _, edges := range gb.graph.Edges {
		for _, edge := range edges {
			if edge.Target == nodeID {
				inDegree++
			}
		}
	}
	
	return float64(inDegree + outDegree)
}

// calculatePathMetrics calculates average path length and diameter
func (gb *GraphBuilder) calculatePathMetrics() (avgPath float64, diameter int) {
	// Simplified calculation - would need proper shortest path algorithm for accuracy
	// For now, use depth-based approximation
	
	totalDepth := 0
	maxDepth := 0
	nodeCount := 0
	
	for _, node := range gb.graph.Nodes {
		totalDepth += node.Depth
		if node.Depth > maxDepth {
			maxDepth = node.Depth
		}
		nodeCount++
	}
	
	if nodeCount > 0 {
		avgPath = float64(totalDepth) / float64(nodeCount)
	}
	diameter = maxDepth
	
	return avgPath, diameter
}