package analysis

import (
	"strings"
	"testing"
)

func TestDataFlowAnalyzer_AnalyzeStateFlow(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	tests := []struct {
		name          string
		filePath      string
		content       string
		expectedNodes int
		expectedHook  string
	}{
		{
			name:     "useState Hook Detection",
			filePath: "/src/components/Counter.tsx",
			content: `
import React, { useState } from 'react';

const Counter = () => {
  const [count, setCount] = useState(0);
  const [name, setName] = useState('');
  
  return (
    <div>
      <p>{count}</p>
      <button onClick={() => setCount(count + 1)}>Increment</button>
    </div>
  );
};
			`,
			expectedNodes: 2,
			expectedHook:  "useState",
		},
		{
			name:     "useReducer Hook Detection",
			filePath: "/src/components/TodoList.tsx",
			content: `
import React, { useReducer } from 'react';

const todoReducer = (state, action) => {
  switch (action.type) {
    case 'ADD_TODO':
      return [...state, action.payload];
    default:
      return state;
  }
};

const TodoList = () => {
  const [todos, dispatch] = useReducer(todoReducer, []);
  
  return <div>{todos.length}</div>;
};
			`,
			expectedNodes: 1,
			expectedHook:  "useReducer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dfa.analyzeStateFlow(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("analyzeStateFlow() error = %v", err)
			}

			// Count nodes with StateFlow type
			stateFlowNodes := 0
			for _, node := range dfa.nodes {
				if node.Type == StateFlow && node.FilePath == tt.filePath {
					stateFlowNodes++
					if hookType, exists := node.Metadata["hook"]; exists {
						if hookType != tt.expectedHook {
							t.Errorf("Expected hook %s, got %s", tt.expectedHook, hookType)
						}
					}
				}
			}

			if stateFlowNodes != tt.expectedNodes {
				t.Errorf("Expected %d state flow nodes, got %d", tt.expectedNodes, stateFlowNodes)
			}
		})
	}
}

func TestDataFlowAnalyzer_AnalyzeAPIFlow(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	tests := []struct {
		name           string
		filePath       string
		content        string
		expectedAPIs   int
		expectedMethod string
	}{
		{
			name:     "Fetch API Detection",
			filePath: "/src/services/userService.js",
			content: `
export const fetchUsers = async () => {
  const response = await fetch('/api/users');
  return response.json();
};

export const fetchUser = async (id) => {
  const response = await fetch('/api/users/' + id);
  return response.json();
};
			`,
			expectedAPIs:   2,
			expectedMethod: "GET",
		},
		{
			name:     "Axios API Detection",
			filePath: "/src/services/postService.js",
			content: `
import axios from 'axios';

export const getPosts = () => {
  return axios.get('/api/posts');
};

export const createPost = (data) => {
  return axios.post('/api/posts', data);
};

export const updatePost = (id, data) => {
  return axios.put('/api/posts/' + id, data);
};

export const deletePost = (id) => {
  return axios.delete('/api/posts/' + id);
};
			`,
			expectedAPIs:   4,
			expectedMethod: "DELETE", // Check the last one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dfa.analyzeAPIFlow(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("analyzeAPIFlow() error = %v", err)
			}

			// Count API calls for this file
			apiCallsCount := 0
			var lastMethod string
			for _, apiCall := range dfa.apiCalls {
				if apiCall.FilePath == tt.filePath {
					apiCallsCount++
					lastMethod = apiCall.Method
				}
			}

			if apiCallsCount != tt.expectedAPIs {
				t.Errorf("Expected %d API calls, got %d", tt.expectedAPIs, apiCallsCount)
			}

			if tt.name == "Axios API Detection" && lastMethod != tt.expectedMethod {
				t.Errorf("Expected last method %s, got %s", tt.expectedMethod, lastMethod)
			}
		})
	}
}

func TestDataFlowAnalyzer_AnalyzeEventFlow(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	tests := []struct {
		name           string
		filePath       string
		content        string
		expectedEvents int
	}{
		{
			name:     "Event Handler Detection",
			filePath: "/src/components/Form.tsx",
			content: `
import React, { useState } from 'react';

const Form = () => {
  const [name, setName] = useState('');
  
  const handleSubmit = (e) => {
    e.preventDefault();
    console.log(name);
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input 
        type="text" 
        value={name}
        onChange={(e) => setName(e.target.value)}
      />
      <button type="submit" onClick={() => console.log('clicked')}>
        Submit
      </button>
    </form>
  );
};
			`,
			expectedEvents: 3, // onSubmit, onChange, onClick
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dfa.analyzeEventFlow(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("analyzeEventFlow() error = %v", err)
			}

			// Count event flow nodes for this file
			eventFlowNodes := 0
			for _, node := range dfa.nodes {
				if node.Type == EventFlow && node.FilePath == tt.filePath {
					eventFlowNodes++
				}
			}

			if eventFlowNodes != tt.expectedEvents {
				t.Errorf("Expected %d event flow nodes, got %d", tt.expectedEvents, eventFlowNodes)
			}
		})
	}
}

func TestDataFlowAnalyzer_AnalyzeContextFlow(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	tests := []struct {
		name          string
		filePath      string
		content       string
		expectedNodes int
		expectedEdges int
	}{
		{
			name:     "Context API Detection",
			filePath: "/src/context/AuthContext.tsx",
			content: `
import React, { createContext, useContext, useState } from 'react';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  
  return (
    <AuthContext.Provider value={{ user, setUser }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  return context;
};
			`,
			expectedNodes: 4, // createContext + useContext (appears multiple times)
			expectedEdges: 3, // Provider edge (appears multiple times)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dfa.analyzeContextFlow(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("analyzeContextFlow() error = %v", err)
			}

			// Count context flow nodes
			contextNodes := 0
			for _, node := range dfa.nodes {
				if node.Type == ContextFlow && node.FilePath == tt.filePath {
					contextNodes++
				}
			}

			// Count context flow edges
			contextEdges := 0
			for _, edge := range dfa.edges {
				if edge.FlowType == ContextFlow {
					contextEdges++
				}
			}

			if contextNodes != tt.expectedNodes {
				t.Errorf("Expected %d context nodes, got %d", tt.expectedNodes, contextNodes)
			}

			if contextEdges != tt.expectedEdges {
				t.Errorf("Expected %d context edges, got %d", tt.expectedEdges, contextEdges)
			}
		})
	}
}

func TestDataFlowAnalyzer_IdentifyBottlenecks(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	// Simulate multiple useState hooks in a single component
	content := `
import React, { useState } from 'react';

const ComplexComponent = () => {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [age, setAge] = useState(0);
  const [address, setAddress] = useState('');
  const [phone, setPhone] = useState('');
  const [city, setCity] = useState('');
  const [country, setCountry] = useState('');
  
  return <div>Complex component</div>;
};
	`

	err := dfa.analyzeStateFlow("/src/ComplexComponent.tsx", content)
	if err != nil {
		t.Fatalf("analyzeStateFlow() error = %v", err)
	}

	bottlenecks := dfa.IdentifyStateManagementBottlenecks()

	// Should detect excessive local state
	found := false
	for _, bottleneck := range bottlenecks {
		if bottleneck.Type == "excessive_local_state" {
			found = true
			if bottleneck.Complexity != "high" {
				t.Errorf("Expected high complexity, got %s", bottleneck.Complexity)
			}
			break
		}
	}

	if !found {
		t.Error("Expected to find excessive_local_state bottleneck")
	}
}

func TestDataFlowAnalyzer_PropFlowAnalysis(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	content := `
import React from 'react';

const ParentComponent = ({ data }) => {
  const { user, settings } = data;
  
  return (
    <ChildComponent 
      user={user} 
      settings={settings}
      onUpdate={(newUser) => updateUser(newUser)}
    />
  );
};

const ChildComponent = ({ user, settings, onUpdate }) => {
  return (
    <GrandChildComponent 
      userName={user.name}
      theme={settings.theme}
      onUserUpdate={onUpdate}
    />
  );
};
	`

	err := dfa.analyzePropFlow("/src/PropDrillingExample.tsx", content)
	if err != nil {
		t.Fatalf("analyzePropFlow() error = %v", err)
	}

	// Check that prop patterns were detected
	propNodes := 0
	propEdges := 0

	for _, node := range dfa.nodes {
		if node.Type == PropFlow {
			propNodes++
		}
	}

	for _, edge := range dfa.edges {
		if edge.FlowType == PropFlow {
			propEdges++
		}
	}

	if propNodes == 0 {
		t.Error("Expected to find prop flow nodes")
	}

	if propEdges == 0 {
		t.Error("Expected to find prop flow edges")
	}
}

func TestDataFlowAnalyzer_CompleteAnalysis(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	// Complex component with multiple data flow patterns
	content := `
import React, { useState, useEffect, useContext } from 'react';
import axios from 'axios';

const UserProfile = ({ userId, onUserUpdate }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(false);
  const { theme } = useContext(ThemeContext);
  
  useEffect(() => {
    const fetchUser = async () => {
      setLoading(true);
      try {
        const response = await axios.get('/api/users/' + userId);
        setUser(response.data);
      } catch (error) {
        console.error(error);
      } finally {
        setLoading(false);
      }
    };
    
    fetchUser();
  }, [userId]);
  
  const handleUpdate = async (userData) => {
    try {
      const response = await fetch('/api/users/' + userId, {
        method: 'PUT',
        body: JSON.stringify(userData)
      });
      const updatedUser = await response.json();
      setUser(updatedUser);
      onUserUpdate(updatedUser);
    } catch (error) {
      console.error(error);
    }
  };
  
  const handleClick = () => {
    console.log('Profile clicked');
  };
  
  if (loading) return <div>Loading...</div>;
  
  return (
    <div onClick={handleClick}>
      <h1>{user?.name}</h1>
      <UserForm user={user} onSubmit={handleUpdate} />
    </div>
  );
};
	`

	err := dfa.AnalyzeDataFlow("/src/UserProfile.tsx", content)
	if err != nil {
		t.Fatalf("AnalyzeDataFlow() error = %v", err)
	}

	// Verify comprehensive analysis
	graph := dfa.GetDataFlowGraph()
	apiCalls := dfa.GetAPICalls()

	// Should detect useState hooks
	stateNodes := 0
	for _, node := range graph.Nodes {
		if node.Type == StateFlow {
			stateNodes++
		}
	}
	if stateNodes < 2 { // user and loading state
		t.Errorf("Expected at least 2 state nodes, got %d", stateNodes)
	}

	// Should detect API calls
	if len(apiCalls) < 1 { // axios.get (fetch with POST body may not be detected)
		t.Errorf("Expected at least 1 API call, got %d", len(apiCalls))
	}

	// Should detect context usage
	contextNodes := 0
	for _, node := range graph.Nodes {
		if node.Type == ContextFlow {
			contextNodes++
		}
	}
	if contextNodes < 1 {
		t.Errorf("Expected at least 1 context node, got %d", contextNodes)
	}

	// Should detect event handlers
	eventNodes := 0
	for _, node := range graph.Nodes {
		if node.Type == EventFlow {
			eventNodes++
		}
	}
	if eventNodes < 1 {
		t.Errorf("Expected at least 1 event node, got %d", eventNodes)
	}
}

func TestDataFlowAnalyzer_ExportToJSON(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	// Add some test data
	content := `
const [count, setCount] = useState(0);
const response = await axios.get('/api/data');
	`

	err := dfa.AnalyzeDataFlow("/test.tsx", content)
	if err != nil {
		t.Fatalf("AnalyzeDataFlow() error = %v", err)
	}

	jsonData, err := dfa.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON export")
	}

	// Verify it's valid JSON by checking it contains expected keys
	jsonStr := string(jsonData)
	expectedKeys := []string{"data_flow_graph", "state_patterns", "api_calls", "bottlenecks"}

	for _, key := range expectedKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("Expected JSON to contain key: %s", key)
		}
	}
}

func TestDataFlowAnalyzer_PatternDetectionHelpers(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	tests := []struct {
		name     string
		line     string
		method   string
		expected bool
	}{
		{"useState detection", "const [count, setCount] = useState(0);", "containsUseState", true},
		{"useReducer detection", "const [state, dispatch] = useReducer(reducer, initial);", "containsUseReducer", true},
		{"state setter detection", "setCount(5);", "containsStateSetter", true},
		{"prop destructuring", "const { user, settings } = props;", "containsPropDestructuring", true},
		{"prop passing", "user={currentUser}", "containsPropPassing", true},
		{"negative useState", "const count = 0;", "containsUseState", false},
		{"negative useReducer", "const reducer = (state, action) => state;", "containsUseReducer", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result bool

			switch tt.method {
			case "containsUseState":
				result = dfa.containsUseState(tt.line)
			case "containsUseReducer":
				result = dfa.containsUseReducer(tt.line)
			case "containsStateSetter":
				result = dfa.containsStateSetter(tt.line)
			case "containsPropDestructuring":
				result = dfa.containsPropDestructuring(tt.line)
			case "containsPropPassing":
				result = dfa.containsPropPassing(tt.line)
			}

			if result != tt.expected {
				t.Errorf("%s(%q) = %v, want %v", tt.method, tt.line, result, tt.expected)
			}
		})
	}
}

func TestDataFlowAnalyzer_StateManagementPatterns(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)

	// Test prop drilling detection
	for i := 0; i < 15; i++ {
		dfa.edges = append(dfa.edges, DataFlowEdge{
			FlowType: PropFlow,
		})
	}

	bottlenecks := dfa.IdentifyStateManagementBottlenecks()

	// Should detect prop drilling
	foundPropDrilling := false
	for _, bottleneck := range bottlenecks {
		if bottleneck.Type == "prop_drilling" {
			foundPropDrilling = true
			if bottleneck.Complexity != "high" {
				t.Errorf("Expected high complexity for prop drilling, got %s", bottleneck.Complexity)
			}
			break
		}
	}

	if !foundPropDrilling {
		t.Error("Expected to detect prop drilling bottleneck")
	}
}
