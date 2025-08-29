package analysis

import (
	"encoding/json"
	"strings"
)

// DataFlowType represents different types of data flow patterns
type DataFlowType string

const (
	StateFlow   DataFlowType = "state_flow"
	PropFlow    DataFlowType = "prop_flow"
	APIFlow     DataFlowType = "api_flow"
	EventFlow   DataFlowType = "event_flow"
	ContextFlow DataFlowType = "context_flow"
)

// DataFlowNode represents a node in the data flow graph
type DataFlowNode struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     DataFlowType           `json:"type"`
	FilePath string                 `json:"file_path"`
	Location DataFlowLocation       `json:"location"`
	Metadata map[string]interface{} `json:"metadata"`
}

// DataFlowLocation represents the location of a data flow element
type DataFlowLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// DataFlowEdge represents a connection between data flow nodes
type DataFlowEdge struct {
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	FlowType    DataFlowType           `json:"flow_type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DataFlowGraph represents the complete data flow analysis result
type DataFlowGraph struct {
	Nodes []DataFlowNode `json:"nodes"`
	Edges []DataFlowEdge `json:"edges"`
}

// StateManagementPattern represents detected state management patterns
type StateManagementPattern struct {
	Type       string                 `json:"type"`
	Location   string                 `json:"location"`
	Complexity string                 `json:"complexity"`
	Issues     []string               `json:"issues"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// APICall represents an API call pattern detected in the code
type APICall struct {
	Method    string                 `json:"method"`
	URL       string                 `json:"url"`
	FilePath  string                 `json:"file_path"`
	Location  DataFlowLocation       `json:"location"`
	TriggerBy string                 `json:"trigger_by"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DataFlowAnalyzer analyzes data flow patterns in the application
type DataFlowAnalyzer struct {
	nodes               []DataFlowNode
	edges               []DataFlowEdge
	statePatterns       []StateManagementPattern
	apiCalls            []APICall
	componentIdentifier *ComponentIdentifier
}

// NewDataFlowAnalyzer creates a new data flow analyzer instance
func NewDataFlowAnalyzer(ci *ComponentIdentifier) *DataFlowAnalyzer {
	return &DataFlowAnalyzer{
		nodes:               make([]DataFlowNode, 0),
		edges:               make([]DataFlowEdge, 0),
		statePatterns:       make([]StateManagementPattern, 0),
		apiCalls:            make([]APICall, 0),
		componentIdentifier: ci,
	}
}

// AnalyzeDataFlow performs comprehensive data flow analysis on a file
func (dfa *DataFlowAnalyzer) AnalyzeDataFlow(filePath string, content string) error {
	// Analyze different types of data flows
	if err := dfa.analyzeStateFlow(filePath, content); err != nil {
		return err
	}

	if err := dfa.analyzePropFlow(filePath, content); err != nil {
		return err
	}

	if err := dfa.analyzeAPIFlow(filePath, content); err != nil {
		return err
	}

	if err := dfa.analyzeEventFlow(filePath, content); err != nil {
		return err
	}

	if err := dfa.analyzeContextFlow(filePath, content); err != nil {
		return err
	}

	return nil
}

// analyzeStateFlow analyzes useState/useReducer patterns and state management
func (dfa *DataFlowAnalyzer) analyzeStateFlow(filePath string, content string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Detect useState hooks
		if dfa.containsUseState(line) {
			node, edge := dfa.extractUseStatePattern(filePath, line, lineNum)
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
				if edge != nil {
					dfa.edges = append(dfa.edges, *edge)
				}
			}
		}

		// Detect useReducer hooks
		if dfa.containsUseReducer(line) {
			node, edge := dfa.extractUseReducerPattern(filePath, line, lineNum)
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
				if edge != nil {
					dfa.edges = append(dfa.edges, *edge)
				}
			}
		}

		// Detect state setters usage
		if dfa.containsStateSetter(line) {
			edge := dfa.extractStateSetterPattern(filePath, line, lineNum)
			if edge != nil {
				dfa.edges = append(dfa.edges, *edge)
			}
		}
	}

	return nil
}

// analyzePropFlow analyzes prop drilling and component prop patterns
func (dfa *DataFlowAnalyzer) analyzePropFlow(filePath string, content string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Detect prop destructuring
		if dfa.containsPropDestructuring(line) {
			node := dfa.extractPropPattern(filePath, line, lineNum)
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}

		// Detect prop passing in JSX
		if dfa.containsPropPassing(line) {
			edge := dfa.extractPropFlowPattern(filePath, line, lineNum)
			if edge != nil {
				dfa.edges = append(dfa.edges, *edge)
			}
		}
	}

	return nil
}

// analyzeAPIFlow analyzes API calls and external data fetching patterns
func (dfa *DataFlowAnalyzer) analyzeAPIFlow(filePath string, content string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Detect fetch calls
		if strings.Contains(line, "fetch(") {
			apiCall := dfa.extractFetchPattern(filePath, line, lineNum)
			if apiCall != nil {
				dfa.apiCalls = append(dfa.apiCalls, *apiCall)
				node := dfa.createAPINode(*apiCall)
				dfa.nodes = append(dfa.nodes, node)
			}
		}

		// Detect axios calls
		if strings.Contains(line, "axios.") || strings.Contains(line, "axios(") {
			apiCall := dfa.extractAxiosPattern(filePath, line, lineNum)
			if apiCall != nil {
				dfa.apiCalls = append(dfa.apiCalls, *apiCall)
				node := dfa.createAPINode(*apiCall)
				dfa.nodes = append(dfa.nodes, node)
			}
		}
	}

	return nil
}

// analyzeEventFlow analyzes event handlers and user interactions
func (dfa *DataFlowAnalyzer) analyzeEventFlow(filePath string, content string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Detect onClick handlers
		if strings.Contains(line, "onClick") {
			node := dfa.extractEventHandlerPattern(filePath, line, lineNum, "click")
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}

		// Detect onChange handlers
		if strings.Contains(line, "onChange") {
			node := dfa.extractEventHandlerPattern(filePath, line, lineNum, "change")
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}

		// Detect onSubmit handlers
		if strings.Contains(line, "onSubmit") {
			node := dfa.extractEventHandlerPattern(filePath, line, lineNum, "submit")
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}
	}

	return nil
}

// analyzeContextFlow analyzes React Context API usage
func (dfa *DataFlowAnalyzer) analyzeContextFlow(filePath string, content string) error {
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Detect Context creation
		if strings.Contains(line, "createContext") {
			node := dfa.extractContextCreationPattern(filePath, line, lineNum)
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}

		// Detect useContext usage
		if strings.Contains(line, "useContext") {
			node := dfa.extractContextUsagePattern(filePath, line, lineNum)
			if node != nil {
				dfa.nodes = append(dfa.nodes, *node)
			}
		}

		// Detect Context Provider
		if strings.Contains(line, "Provider") {
			edge := dfa.extractContextProviderPattern(filePath, line, lineNum)
			if edge != nil {
				dfa.edges = append(dfa.edges, *edge)
			}
		}
	}

	return nil
}

// Helper methods for pattern detection

func (dfa *DataFlowAnalyzer) containsUseState(line string) bool {
	return strings.Contains(line, "useState") && (strings.Contains(line, "const ") || strings.Contains(line, "let "))
}

func (dfa *DataFlowAnalyzer) containsUseReducer(line string) bool {
	return strings.Contains(line, "useReducer") && (strings.Contains(line, "const ") || strings.Contains(line, "let "))
}

func (dfa *DataFlowAnalyzer) containsStateSetter(line string) bool {
	return strings.Contains(line, "set") && (strings.Contains(line, "(") && strings.Contains(line, ")"))
}

func (dfa *DataFlowAnalyzer) containsPropDestructuring(line string) bool {
	return strings.Contains(line, "const {") || strings.Contains(line, "let {") ||
		strings.Contains(line, "= ({ ") || strings.Contains(line, "= (props")
}

func (dfa *DataFlowAnalyzer) containsPropPassing(line string) bool {
	return strings.Contains(line, "=") && strings.Contains(line, "{") && strings.Contains(line, "}")
}

// Pattern extraction methods

func (dfa *DataFlowAnalyzer) extractUseStatePattern(filePath, line string, lineNum int) (*DataFlowNode, *DataFlowEdge) {
	// Extract state variable and setter names
	parts := strings.Split(line, "useState")
	if len(parts) < 2 {
		return nil, nil
	}

	leftPart := strings.TrimSpace(parts[0])
	// Simple extraction - in production, use proper AST parsing
	if strings.Contains(leftPart, "[") && strings.Contains(leftPart, "]") {
		start := strings.Index(leftPart, "[")
		end := strings.Index(leftPart, "]")
		if start < end {
			variables := leftPart[start+1 : end]
			vars := strings.Split(variables, ",")
			if len(vars) >= 2 {
				stateVar := strings.TrimSpace(vars[0])
				setter := strings.TrimSpace(vars[1])

				node := &DataFlowNode{
					ID:       filePath + ":" + stateVar,
					Name:     stateVar,
					Type:     StateFlow,
					FilePath: filePath,
					Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
					Metadata: map[string]interface{}{
						"setter": setter,
						"hook":   "useState",
					},
				}

				return node, nil
			}
		}
	}

	return nil, nil
}

func (dfa *DataFlowAnalyzer) extractUseReducerPattern(filePath, line string, lineNum int) (*DataFlowNode, *DataFlowEdge) {
	// Similar to useState but for useReducer
	parts := strings.Split(line, "useReducer")
	if len(parts) < 2 {
		return nil, nil
	}

	leftPart := strings.TrimSpace(parts[0])
	if strings.Contains(leftPart, "[") && strings.Contains(leftPart, "]") {
		start := strings.Index(leftPart, "[")
		end := strings.Index(leftPart, "]")
		if start < end {
			variables := leftPart[start+1 : end]
			vars := strings.Split(variables, ",")
			if len(vars) >= 2 {
				stateVar := strings.TrimSpace(vars[0])
				dispatch := strings.TrimSpace(vars[1])

				node := &DataFlowNode{
					ID:       filePath + ":" + stateVar,
					Name:     stateVar,
					Type:     StateFlow,
					FilePath: filePath,
					Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
					Metadata: map[string]interface{}{
						"dispatch": dispatch,
						"hook":     "useReducer",
					},
				}

				return node, nil
			}
		}
	}

	return nil, nil
}

func (dfa *DataFlowAnalyzer) extractStateSetterPattern(filePath, line string, lineNum int) *DataFlowEdge {
	// Extract state setter calls like setCount(5) or setUser({...})
	if strings.Contains(line, "set") {
		// Simple pattern matching - in production, use AST
		return &DataFlowEdge{
			From:        "unknown",
			To:          filePath + ":" + "state",
			FlowType:    StateFlow,
			Description: "State update",
			Metadata: map[string]interface{}{
				"line":   lineNum + 1,
				"source": line,
			},
		}
	}
	return nil
}

func (dfa *DataFlowAnalyzer) extractPropPattern(filePath, line string, lineNum int) *DataFlowNode {
	// Extract prop destructuring patterns
	return &DataFlowNode{
		ID:       filePath + ":" + "props:" + string(rune(lineNum)),
		Name:     "props",
		Type:     PropFlow,
		FilePath: filePath,
		Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
		Metadata: map[string]interface{}{
			"pattern": "destructuring",
			"source":  line,
		},
	}
}

func (dfa *DataFlowAnalyzer) extractPropFlowPattern(filePath, line string, lineNum int) *DataFlowEdge {
	// Extract prop passing patterns in JSX
	return &DataFlowEdge{
		From:        filePath,
		To:          "child_component",
		FlowType:    PropFlow,
		Description: "Prop passing",
		Metadata: map[string]interface{}{
			"line":   lineNum + 1,
			"source": line,
		},
	}
}

func (dfa *DataFlowAnalyzer) extractFetchPattern(filePath, line string, lineNum int) *APICall {
	// Extract fetch() API calls
	start := strings.Index(line, "fetch(")
	if start == -1 {
		return nil
	}

	remaining := line[start+6:] // Skip "fetch("
	end := strings.Index(remaining, ")")
	if end == -1 {
		return nil
	}

	url := strings.TrimSpace(remaining[:end])
	url = strings.Trim(url, "'\"")

	return &APICall{
		Method:    "GET", // Default for fetch
		URL:       url,
		FilePath:  filePath,
		Location:  DataFlowLocation{Line: lineNum + 1, Column: start},
		TriggerBy: "fetch",
		Metadata: map[string]interface{}{
			"client": "fetch",
		},
	}
}

func (dfa *DataFlowAnalyzer) extractAxiosPattern(filePath, line string, lineNum int) *APICall {
	// Extract axios API calls
	var method, url string

	if strings.Contains(line, "axios.get") {
		method = "GET"
	} else if strings.Contains(line, "axios.post") {
		method = "POST"
	} else if strings.Contains(line, "axios.put") {
		method = "PUT"
	} else if strings.Contains(line, "axios.delete") {
		method = "DELETE"
	} else {
		method = "UNKNOWN"
	}

	// Extract URL (simplified)
	start := strings.Index(line, "(")
	if start != -1 {
		remaining := line[start+1:]
		end := strings.Index(remaining, ")")
		if end != -1 {
			url = strings.TrimSpace(remaining[:end])
			url = strings.Trim(url, "'\"")
		}
	}

	return &APICall{
		Method:    method,
		URL:       url,
		FilePath:  filePath,
		Location:  DataFlowLocation{Line: lineNum + 1, Column: 0},
		TriggerBy: "axios",
		Metadata: map[string]interface{}{
			"client": "axios",
		},
	}
}

func (dfa *DataFlowAnalyzer) extractEventHandlerPattern(filePath, line string, lineNum int, eventType string) *DataFlowNode {
	return &DataFlowNode{
		ID:       filePath + ":" + eventType + ":" + string(rune(lineNum)),
		Name:     "on" + strings.Title(eventType),
		Type:     EventFlow,
		FilePath: filePath,
		Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
		Metadata: map[string]interface{}{
			"event_type": eventType,
			"source":     line,
		},
	}
}

func (dfa *DataFlowAnalyzer) extractContextCreationPattern(filePath, line string, lineNum int) *DataFlowNode {
	return &DataFlowNode{
		ID:       filePath + ":context:" + string(rune(lineNum)),
		Name:     "Context",
		Type:     ContextFlow,
		FilePath: filePath,
		Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
		Metadata: map[string]interface{}{
			"pattern": "creation",
			"source":  line,
		},
	}
}

func (dfa *DataFlowAnalyzer) extractContextUsagePattern(filePath, line string, lineNum int) *DataFlowNode {
	return &DataFlowNode{
		ID:       filePath + ":context_usage:" + string(rune(lineNum)),
		Name:     "useContext",
		Type:     ContextFlow,
		FilePath: filePath,
		Location: DataFlowLocation{Line: lineNum + 1, Column: 0},
		Metadata: map[string]interface{}{
			"pattern": "usage",
			"source":  line,
		},
	}
}

func (dfa *DataFlowAnalyzer) extractContextProviderPattern(filePath, line string, lineNum int) *DataFlowEdge {
	return &DataFlowEdge{
		From:        filePath + ":provider",
		To:          filePath + ":consumers",
		FlowType:    ContextFlow,
		Description: "Context value provision",
		Metadata: map[string]interface{}{
			"line":   lineNum + 1,
			"source": line,
		},
	}
}

func (dfa *DataFlowAnalyzer) createAPINode(apiCall APICall) DataFlowNode {
	return DataFlowNode{
		ID:       apiCall.FilePath + ":api:" + apiCall.Method + ":" + apiCall.URL,
		Name:     apiCall.Method + " " + apiCall.URL,
		Type:     APIFlow,
		FilePath: apiCall.FilePath,
		Location: apiCall.Location,
		Metadata: map[string]interface{}{
			"method": apiCall.Method,
			"url":    apiCall.URL,
			"client": apiCall.TriggerBy,
		},
	}
}

// Analysis result methods

func (dfa *DataFlowAnalyzer) GetDataFlowGraph() DataFlowGraph {
	return DataFlowGraph{
		Nodes: dfa.nodes,
		Edges: dfa.edges,
	}
}

func (dfa *DataFlowAnalyzer) GetStateManagementPatterns() []StateManagementPattern {
	return dfa.statePatterns
}

func (dfa *DataFlowAnalyzer) GetAPICalls() []APICall {
	return dfa.apiCalls
}

func (dfa *DataFlowAnalyzer) IdentifyStateManagementBottlenecks() []StateManagementPattern {
	// Identify patterns that indicate bottlenecks or anti-patterns
	bottlenecks := make([]StateManagementPattern, 0)

	// Analyze for excessive prop drilling
	propDrillingDepth := dfa.calculatePropDrillingDepth()
	if propDrillingDepth > 3 {
		bottlenecks = append(bottlenecks, StateManagementPattern{
			Type:       "prop_drilling",
			Location:   "multiple_components",
			Complexity: "high",
			Issues:     []string{"Deep prop drilling detected", "Consider using Context or state management"},
			Metadata:   map[string]interface{}{"depth": propDrillingDepth},
		})
	}

	// Analyze for too many useState hooks in single component
	stateHooksCount := dfa.countStateHooksPerComponent()
	for filePath, count := range stateHooksCount {
		if count > 5 {
			bottlenecks = append(bottlenecks, StateManagementPattern{
				Type:       "excessive_local_state",
				Location:   filePath,
				Complexity: "high",
				Issues:     []string{"Too many useState hooks", "Consider useReducer or external state management"},
				Metadata:   map[string]interface{}{"hooks_count": count},
			})
		}
	}

	return bottlenecks
}

func (dfa *DataFlowAnalyzer) calculatePropDrillingDepth() int {
	// Simple heuristic - count max depth of prop flow edges
	maxDepth := 0
	// Implementation would trace the actual prop flow paths
	// This is a simplified version
	propFlowCount := 0
	for _, edge := range dfa.edges {
		if edge.FlowType == PropFlow {
			propFlowCount++
		}
	}

	// Rough estimate - in real implementation, build actual flow graph
	if propFlowCount > 10 {
		maxDepth = 4
	} else if propFlowCount > 5 {
		maxDepth = 3
	} else if propFlowCount > 2 {
		maxDepth = 2
	}

	return maxDepth
}

func (dfa *DataFlowAnalyzer) countStateHooksPerComponent() map[string]int {
	counts := make(map[string]int)

	for _, node := range dfa.nodes {
		if node.Type == StateFlow {
			counts[node.FilePath]++
		}
	}

	return counts
}

func (dfa *DataFlowAnalyzer) ExportToJSON() ([]byte, error) {
	result := map[string]interface{}{
		"data_flow_graph": dfa.GetDataFlowGraph(),
		"state_patterns":  dfa.statePatterns,
		"api_calls":       dfa.apiCalls,
		"bottlenecks":     dfa.IdentifyStateManagementBottlenecks(),
	}

	return json.MarshalIndent(result, "", "  ")
}
