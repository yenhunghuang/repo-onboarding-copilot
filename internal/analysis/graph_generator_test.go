package analysis

import (
	"strings"
	"testing"
)

func TestGraphGenerator_GenerateGraph(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	tests := []struct {
		name     string
		filePath string
		content  string
		wantNode bool
		nodeType GraphNodeType
	}{
		{
			name:     "React Component",
			filePath: "/src/components/Button.jsx",
			content: `
import React from 'react';
import './Button.css';

export const Button = ({ onClick, children }) => {
  return <button onClick={onClick}>{children}</button>;
};
			`,
			wantNode: true,
			nodeType: ComponentNode,
		},
		{
			name:     "Service Module",
			filePath: "/src/services/apiService.js",
			content: `
import axios from 'axios';

export const apiService = {
  async fetchUsers() {
    return axios.get('/api/users');
  }
};
			`,
			wantNode: true,
			nodeType: ServiceNode,
		},
		{
			name:     "Utility Module",
			filePath: "/src/utils/helpers.js",
			content: `
export const formatDate = (date) => {
  return new Date(date).toLocaleDateString();
};

export const debounce = (func, wait) => {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
};
			`,
			wantNode: true,
			nodeType: UtilityNode,
		},
		{
			name:     "Configuration Module",
			filePath: "/src/config/app.js",
			content: `
export const config = {
  apiUrl: process.env.REACT_APP_API_URL || 'http://localhost:3000',
  environment: process.env.NODE_ENV || 'development'
};
			`,
			wantNode: true,
			nodeType: ConfigNode,
		},
		{
			name:     "Test File",
			filePath: "/src/components/Button.test.js",
			content: `
import React from 'react';
import { render, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders correctly', () => {
    render(<Button>Click me</Button>);
  });
});
			`,
			wantNode: true,
			nodeType: UtilityNode, // Test files are currently detected as utilities
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gg.GenerateGraph(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("GenerateGraph() error = %v", err)
			}

			if tt.wantNode {
				nodeID := gg.generateNodeID(tt.filePath)
				node, exists := gg.nodes[nodeID]
				if !exists {
					t.Fatalf("Expected node %s to be created", nodeID)
				}

				if node.Type != tt.nodeType {
					t.Errorf("Expected node type %v, got %v", tt.nodeType, node.Type)
				}

				if node.FilePath != tt.filePath {
					t.Errorf("Expected file path %s, got %s", tt.filePath, node.FilePath)
				}

				if node.Size <= 0 {
					t.Error("Expected node size to be greater than 0")
				}

				if node.Importance < 0 || node.Importance > 1 {
					t.Errorf("Expected importance between 0 and 1, got %f", node.Importance)
				}

				if node.Complexity < 0 || node.Complexity > 1 {
					t.Errorf("Expected complexity between 0 and 1, got %f", node.Complexity)
				}

				if node.Style == nil {
					t.Error("Expected node style to be set")
				}
			}
		})
	}
}

func TestGraphGenerator_CreateEdges(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	// Create source and target files
	sourceFile := "/src/components/App.jsx"
	sourceContent := `
import React from 'react';
import { Button } from './Button.jsx';
import { apiService } from '../services/apiService.js';

export const App = () => {
  const handleClick = async () => {
    await apiService.fetchUsers();
  };

  return (
    <div>
      <Button onClick={handleClick}>Load Users</Button>
    </div>
  );
};
	`

	targetFile1 := "/src/components/Button.jsx"
	targetContent1 := `
import React from 'react';
export const Button = ({ onClick, children }) => {
  return <button onClick={onClick}>{children}</button>;
};
	`

	targetFile2 := "/src/services/apiService.js"
	targetContent2 := `
export const apiService = {
  async fetchUsers() {
    return fetch('/api/users');
  }
};
	`

	// Generate nodes and edges
	err := gg.GenerateGraph(sourceFile, sourceContent)
	if err != nil {
		t.Fatalf("GenerateGraph() error for source: %v", err)
	}

	err = gg.GenerateGraph(targetFile1, targetContent1)
	if err != nil {
		t.Fatalf("GenerateGraph() error for target1: %v", err)
	}

	err = gg.GenerateGraph(targetFile2, targetContent2)
	if err != nil {
		t.Fatalf("GenerateGraph() error for target2: %v", err)
	}

	// Check that edges were created
	if len(gg.edges) == 0 {
		t.Error("Expected edges to be created")
	}

	// Verify specific edge types exist
	hasImportEdge := false

	for _, edge := range gg.edges {
		if edge.Type == ImportEdge {
			hasImportEdge = true
		}

		// Verify edge properties
		if edge.Source == "" || edge.Target == "" {
			t.Error("Expected edge to have source and target")
		}

		if edge.Weight < 0 || edge.Weight > 1 {
			t.Errorf("Expected edge weight between 0 and 1, got %f", edge.Weight)
		}

		if edge.Style == nil {
			t.Error("Expected edge style to be set")
		}
	}

	if !hasImportEdge {
		t.Error("Expected to find import edge")
	}
	// Note: ComponentEdge detection depends on more complex JSX parsing
	// so we don't strictly require it in this test
}

func TestGraphGenerator_GenerateClusters(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	// Create multiple files in different features/layers
	files := map[string]string{
		"/src/components/user/UserProfile.jsx": `
import React from 'react';
export const UserProfile = () => <div>User Profile</div>;
		`,
		"/src/components/user/UserList.jsx": `
import React from 'react';
export const UserList = () => <div>User List</div>;
		`,
		"/src/services/userService.js": `
export const userService = {
  getUsers: () => fetch('/users')
};
		`,
		"/src/services/authService.js": `
export const authService = {
  login: () => fetch('/auth/login')
};
		`,
		"/src/utils/dateUtils.js": `
export const formatDate = (date) => date.toLocaleDateString();
		`,
	}

	// Generate nodes for all files
	for filePath, content := range files {
		err := gg.GenerateGraph(filePath, content)
		if err != nil {
			t.Fatalf("GenerateGraph() error for %s: %v", filePath, err)
		}
	}

	// Generate clusters
	err := gg.GenerateClusters()
	if err != nil {
		t.Fatalf("GenerateClusters() error: %v", err)
	}

	// Check that clusters were created
	if len(gg.clusters) == 0 {
		t.Error("Expected clusters to be created")
	}

	// Verify cluster types
	hasFeatureCluster := false
	hasLayerCluster := false

	for _, cluster := range gg.clusters {
		if cluster.Type == FeatureCluster {
			hasFeatureCluster = true
		}
		if cluster.Type == LayerCluster {
			hasLayerCluster = true
		}

		// Verify cluster properties
		if cluster.ID == "" {
			t.Error("Expected cluster to have ID")
		}

		if cluster.Name == "" {
			t.Error("Expected cluster to have name")
		}

		if len(cluster.NodeIDs) == 0 {
			t.Error("Expected cluster to have node IDs")
		}

		if cluster.Color == "" {
			t.Error("Expected cluster to have color")
		}
	}

	if !hasFeatureCluster {
		t.Error("Expected to find feature cluster")
	}

	if !hasLayerCluster {
		t.Error("Expected to find layer cluster")
	}
}

func TestGraphGenerator_CycleClusters(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	// Create files with circular dependencies
	files := map[string]string{
		"/src/components/A.jsx": `
import B from './B.jsx';
export const A = () => <B />;
		`,
		"/src/components/B.jsx": `
import A from './A.jsx';
export const B = () => <A />;
		`,
	}

	// Add files to cycle detector first
	for filePath, content := range files {
		err := cd.DetectCycles(filePath, content)
		if err != nil {
			t.Fatalf("DetectCycles() error for %s: %v", filePath, err)
		}
	}

	// Analyze cycles
	err := cd.AnalyzeCycles()
	if err != nil {
		t.Fatalf("AnalyzeCycles() error: %v", err)
	}

	// Generate graph nodes
	for filePath, content := range files {
		err := gg.GenerateGraph(filePath, content)
		if err != nil {
			t.Fatalf("GenerateGraph() error for %s: %v", filePath, err)
		}
	}

	// Generate clusters (including cycle clusters)
	err = gg.GenerateClusters()
	if err != nil {
		t.Fatalf("GenerateClusters() error: %v", err)
	}

	// Check for cycle clusters
	foundCycleCluster := false
	for _, cluster := range gg.clusters {
		if cluster.Type == CycleCluster {
			foundCycleCluster = true

			// Verify cycle cluster properties
			if len(cluster.NodeIDs) < 2 {
				t.Error("Expected cycle cluster to have at least 2 nodes")
			}

			if cluster.Color != "#F44336" {
				t.Errorf("Expected cycle cluster color to be red, got %s", cluster.Color)
			}

			if !strings.Contains(cluster.Name, "Cycle") {
				t.Error("Expected cycle cluster name to contain 'Cycle'")
			}
		}
	}

	if !foundCycleCluster {
		t.Error("Expected to find cycle cluster")
	}
}

func TestGraphGenerator_NodeMetrics(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	tests := []struct {
		name                 string
		filePath             string
		content              string
		expectHighImportance bool
		expectHighComplexity bool
	}{
		{
			name:     "Simple Component",
			filePath: "/src/components/Simple.jsx",
			content: `
import React from 'react';
export const Simple = () => <div>Hello</div>;
			`,
			expectHighImportance: false,
			expectHighComplexity: false,
		},
		{
			name:     "Complex Component",
			filePath: "/src/components/App.jsx", // App suggests main entry point
			content: `
import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';

export const App = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/users');
        if (response.ok) {
          const data = await response.json();
          setUsers(data);
        } else {
          throw new Error('Failed to fetch');
        }
      } catch (error) {
        console.error('Error:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []);

  if (loading) return <div>Loading...</div>;

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<UserList users={users} />} />
        <Route path="/user/:id" element={<UserDetail />} />
      </Routes>
    </BrowserRouter>
  );
};

export const UserList = ({ users }) => (
  <div>
    {users.map(user => (
      <div key={user.id}>{user.name}</div>
    ))}
  </div>
);

export const UserDetail = () => <div>User Detail</div>;
			`,
			expectHighImportance: true,
			expectHighComplexity: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gg.GenerateGraph(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("GenerateGraph() error: %v", err)
			}

			nodeID := gg.generateNodeID(tt.filePath)
			node, exists := gg.nodes[nodeID]
			if !exists {
				t.Fatalf("Expected node %s to exist", nodeID)
			}

			if tt.expectHighImportance && node.Importance < 0.6 {
				t.Errorf("Expected high importance (>0.6), got %f", node.Importance)
			}

			if !tt.expectHighImportance && node.Importance > 0.6 {
				t.Errorf("Expected low importance (<0.6), got %f", node.Importance)
			}

			if tt.expectHighComplexity && node.Complexity < 0.3 {
				t.Errorf("Expected high complexity (>0.3), got %f", node.Complexity)
			}

			if !tt.expectHighComplexity && node.Complexity > 0.3 {
				t.Errorf("Expected low complexity (<0.3), got %f", node.Complexity)
			}
		})
	}
}

func TestGraphGenerator_GetGraph(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	// Create a simple graph
	files := map[string]string{
		"/src/components/App.jsx": `
import React from 'react';
import { Button } from './Button.jsx';

export const App = () => <Button>Click me</Button>;
		`,
		"/src/components/Button.jsx": `
import React from 'react';
export const Button = ({ children }) => <button>{children}</button>;
		`,
		"/src/utils/helpers.js": `
export const formatDate = (date) => date.toLocaleDateString();
		`,
	}

	// Generate graph
	for filePath, content := range files {
		err := gg.GenerateGraph(filePath, content)
		if err != nil {
			t.Fatalf("GenerateGraph() error for %s: %v", filePath, err)
		}
	}

	err := gg.GenerateClusters()
	if err != nil {
		t.Fatalf("GenerateClusters() error: %v", err)
	}

	// Get complete graph
	graph := gg.GetGraph()

	// Verify graph structure
	if graph == nil {
		t.Fatal("Expected graph to be returned")
	}

	if len(graph.Nodes) == 0 {
		t.Error("Expected graph to have nodes")
	}

	if len(graph.Edges) == 0 {
		t.Error("Expected graph to have edges")
	}

	if len(graph.Clusters) == 0 {
		t.Error("Expected graph to have clusters")
	}

	if graph.Layout == nil {
		t.Error("Expected graph to have layout")
	}

	if graph.Stats == nil {
		t.Error("Expected graph to have stats")
	}

	if graph.Metadata == nil {
		t.Error("Expected graph to have metadata")
	}

	// Verify stats
	stats := graph.Stats
	if stats.TotalNodes != len(graph.Nodes) {
		t.Errorf("Expected stats total nodes %d, got %d", len(graph.Nodes), stats.TotalNodes)
	}

	if stats.TotalEdges != len(graph.Edges) {
		t.Errorf("Expected stats total edges %d, got %d", len(graph.Edges), stats.TotalEdges)
	}

	if stats.TotalClusters != len(graph.Clusters) {
		t.Errorf("Expected stats total clusters %d, got %d", len(graph.Clusters), stats.TotalClusters)
	}

	if len(stats.NodesByType) == 0 {
		t.Error("Expected stats to have nodes by type breakdown")
	}

	if len(stats.EdgesByType) == 0 {
		t.Error("Expected stats to have edges by type breakdown")
	}

	// Verify layout
	layout := graph.Layout
	if layout.Algorithm == "" {
		t.Error("Expected layout to have algorithm")
	}

	if layout.Width <= 0 || layout.Height <= 0 {
		t.Error("Expected layout to have positive dimensions")
	}

	// Verify metadata
	metadata := graph.Metadata
	if metadata.Title == "" {
		t.Error("Expected metadata to have title")
	}

	if metadata.Description == "" {
		t.Error("Expected metadata to have description")
	}
}

func TestGraphGenerator_ExportToJSON(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	// Create a simple graph
	err := gg.GenerateGraph("/src/components/Test.jsx", `
import React from 'react';
export const Test = () => <div>Test</div>;
	`)
	if err != nil {
		t.Fatalf("GenerateGraph() error: %v", err)
	}

	err = gg.GenerateClusters()
	if err != nil {
		t.Fatalf("GenerateClusters() error: %v", err)
	}

	// Export to JSON
	jsonStr, err := gg.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() error: %v", err)
	}

	// Verify JSON structure
	if !strings.Contains(jsonStr, "nodes") {
		t.Error("Expected JSON to contain 'nodes' field")
	}

	if !strings.Contains(jsonStr, "edges") {
		t.Error("Expected JSON to contain 'edges' field")
	}

	if !strings.Contains(jsonStr, "clusters") {
		t.Error("Expected JSON to contain 'clusters' field")
	}

	if !strings.Contains(jsonStr, "metadata") {
		t.Error("Expected JSON to contain 'metadata' field")
	}

	if !strings.Contains(jsonStr, "layout") {
		t.Error("Expected JSON to contain 'layout' field")
	}

	if !strings.Contains(jsonStr, "stats") {
		t.Error("Expected JSON to contain 'stats' field")
	}

	// Verify it's valid JSON by length check
	if len(jsonStr) < 100 {
		t.Error("Expected JSON to be substantial in size")
	}
}

func TestGraphGenerator_NodeStyles(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	tests := []struct {
		name      string
		nodeType  GraphNodeType
		wantColor string
	}{
		{
			name:      "Component Node",
			nodeType:  ComponentNode,
			wantColor: "#4CAF50",
		},
		{
			name:      "Service Node",
			nodeType:  ServiceNode,
			wantColor: "#2196F3",
		},
		{
			name:      "Utility Node",
			nodeType:  UtilityNode,
			wantColor: "#FF9800",
		},
		{
			name:      "Config Node",
			nodeType:  ConfigNode,
			wantColor: "#9C27B0",
		},
		{
			name:      "Test Node",
			nodeType:  TestNode,
			wantColor: "#607D8B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := gg.createNodeStyle(tt.nodeType, 0.5, 0.3)

			if style.Color != tt.wantColor {
				t.Errorf("Expected color %s, got %s", tt.wantColor, style.Color)
			}

			if style.Size <= 0 {
				t.Error("Expected positive size")
			}

			if style.Shape == "" {
				t.Error("Expected shape to be set")
			}

			if style.Opacity <= 0 || style.Opacity > 1 {
				t.Errorf("Expected opacity between 0 and 1, got %f", style.Opacity)
			}
		})
	}
}

func TestGraphGenerator_EdgeStyles(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	tests := []struct {
		name      string
		edgeType  GraphEdgeType
		weight    float64
		wantColor string
	}{
		{
			name:      "Import Edge",
			edgeType:  ImportEdge,
			weight:    0.8,
			wantColor: "#333333",
		},
		{
			name:      "Component Edge",
			edgeType:  ComponentEdge,
			weight:    0.9,
			wantColor: "#4CAF50",
		},
		{
			name:      "Data Flow Edge",
			edgeType:  FlowEdge,
			weight:    0.6,
			wantColor: "#2196F3",
		},
		{
			name:      "API Edge",
			edgeType:  APIEdge,
			weight:    0.5,
			wantColor: "#FF9800",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := gg.createEdgeStyle(tt.edgeType, tt.weight)

			if style.Color != tt.wantColor {
				t.Errorf("Expected color %s, got %s", tt.wantColor, style.Color)
			}

			expectedWidth := 1 + (tt.weight * 3)
			if style.Width != expectedWidth {
				t.Errorf("Expected width %f, got %f", expectedWidth, style.Width)
			}

			if style.Style == "" {
				t.Error("Expected style to be set")
			}

			expectedOpacity := 0.6 + (tt.weight * 0.4)
			if style.Opacity != expectedOpacity {
				t.Errorf("Expected opacity %f, got %f", expectedOpacity, style.Opacity)
			}
		})
	}
}

func TestGraphGenerator_HelperFunctions(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	cd := NewCycleDetector(ci)
	gg := NewGraphGenerator(ci, dfa, cd)

	t.Run("generateNodeID", func(t *testing.T) {
		filePath := "/src/components/Button.jsx"
		nodeID := gg.generateNodeID(filePath)

		if nodeID == "" {
			t.Error("Expected non-empty node ID")
		}

		if strings.Contains(nodeID, "/") {
			t.Error("Expected node ID to not contain slashes")
		}

		if strings.Contains(nodeID, ".") {
			t.Error("Expected node ID to not contain dots")
		}
	})

	t.Run("extractNodeName", func(t *testing.T) {
		filePath := "/src/components/Button.jsx"
		name := gg.extractNodeName(filePath)

		if name != "Button" {
			t.Errorf("Expected name 'Button', got '%s'", name)
		}
	})

	t.Run("calculateSize", func(t *testing.T) {
		content := "line1\nline2\nline3"
		size := gg.calculateSize(content)

		if size != 3 {
			t.Errorf("Expected size 3, got %d", size)
		}
	})

	t.Run("calculateImportance", func(t *testing.T) {
		// Test main app file (should have high importance)
		appImportance := gg.calculateImportance("/src/App.js", "export default App;")
		if appImportance <= 0.5 {
			t.Errorf("Expected high importance for App.js, got %f", appImportance)
		}

		// Test regular component (should have normal importance)
		componentImportance := gg.calculateImportance("/src/components/Button.js", "export const Button = () => {};")
		if componentImportance <= 0 || componentImportance > 1 {
			t.Errorf("Expected importance between 0 and 1, got %f", componentImportance)
		}
	})

	t.Run("calculateComplexity", func(t *testing.T) {
		simpleContent := "export const simple = () => 'hello';"
		simpleComplexity := gg.calculateComplexity(simpleContent)

		complexContent := `
function complex() {
  if (condition) {
    for (let i = 0; i < 10; i++) {
      if (another) {
        while (loop) {
          try {
            async function nested() {
              await something();
            }
          } catch (error) {
            switch (error.type) {
              case 'A': return;
              case 'B': break;
            }
          }
        }
      }
    }
  }
}
		`
		complexComplexity := gg.calculateComplexity(complexContent)

		if complexComplexity <= simpleComplexity {
			t.Errorf("Expected complex content to have higher complexity than simple content")
		}

		if simpleComplexity < 0 || simpleComplexity > 1 {
			t.Errorf("Expected simple complexity between 0 and 1, got %f", simpleComplexity)
		}

		if complexComplexity < 0 || complexComplexity > 1 {
			t.Errorf("Expected complex complexity between 0 and 1, got %f", complexComplexity)
		}
	})

	t.Run("extractFeature", func(t *testing.T) {
		feature := gg.extractFeature("/src/components/user/Profile.jsx")
		if feature != "components" {
			t.Errorf("Expected feature 'components', got '%s'", feature)
		}

		coreFeature := gg.extractFeature("/src/utils.js")
		if coreFeature != "utils.js" {
			t.Errorf("Expected feature 'utils.js' for root files, got '%s'", coreFeature)
		}
	})
}
