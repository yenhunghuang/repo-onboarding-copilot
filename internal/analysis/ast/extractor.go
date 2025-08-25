package ast

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
)

// extractASTInfo traverses the AST and extracts structured information
func (p *Parser) extractASTInfo(node *sitter.Node, content []byte, result *ParseResult) error {
	// Track parsing context
	result.Metadata["node_count"] = 0
	result.Metadata["max_depth"] = 0

	// Start recursive extraction
	return p.walkNode(node, content, result, 0)
}

// walkNode recursively walks the AST tree and extracts information
func (p *Parser) walkNode(node *sitter.Node, content []byte, result *ParseResult, depth int) error {
	if node == nil {
		return nil
	}

	// Update metadata
	if nodeCount, ok := result.Metadata["node_count"].(int); ok {
		result.Metadata["node_count"] = nodeCount + 1
	}
	if maxDepth, ok := result.Metadata["max_depth"].(int); ok && depth > maxDepth {
		result.Metadata["max_depth"] = depth
	}

	nodeType := node.Type()

	// Extract information based on node type
	switch nodeType {
	case "function_declaration", "function_expression", "arrow_function", "method_definition":
		if err := p.extractFunction(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting function: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}

	case "class_declaration":
		if err := p.extractClass(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting class: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}

	case "interface_declaration":
		if err := p.extractInterface(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting interface: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}

	case "variable_declaration", "lexical_declaration":
		if err := p.extractVariables(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting variables: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}

	case "import_statement":
		if err := p.extractImport(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting import: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}

	case "export_statement":
		if err := p.extractExport(node, content, result); err != nil {
			result.Errors = append(result.Errors, ParseError{
				Message:  fmt.Sprintf("Error extracting export: %v", err),
				Line:     int(node.StartPoint().Row) + 1,
				Column:   int(node.StartPoint().Column) + 1,
				Severity: "warning",
				Context:  nodeType,
			})
		}
	}

	// Recursively process child nodes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if err := p.walkNode(child, content, result, depth+1); err != nil {
			return err
		}
	}

	return nil
}

// extractFunction extracts function information from AST node
func (p *Parser) extractFunction(node *sitter.Node, content []byte, result *ParseResult) error {
	function := FunctionInfo{
		Parameters: []ParameterInfo{},
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Metadata:   make(map[string]string),
	}

	// Extract function name
	if nameNode := p.findChildByType(node, "identifier"); nameNode != nil {
		function.Name = p.getNodeText(nameNode, content)
	}

	// Check if function is async
	if p.findChildByType(node, "async") != nil {
		function.IsAsync = true
	}

	// Extract parameters
	if paramsNode := p.findChildByType(node, "formal_parameters"); paramsNode != nil {
		function.Parameters = p.extractParameters(paramsNode, content)
	}

	// Extract return type (TypeScript)
	if typeAnnotation := p.findChildByType(node, "type_annotation"); typeAnnotation != nil {
		function.ReturnType = p.getNodeText(typeAnnotation, content)
	}

	// Check if exported
	function.IsExported = p.isExported(node)

	// Add metadata
	function.Metadata["node_type"] = node.Type()
	if function.IsAsync {
		function.Metadata["async"] = "true"
	}

	result.Functions = append(result.Functions, function)
	return nil
}

// extractClass extracts class information from AST node
func (p *Parser) extractClass(node *sitter.Node, content []byte, result *ParseResult) error {
	class := ClassInfo{
		Methods:    []FunctionInfo{},
		Properties: []PropertyInfo{},
		Implements: []string{},
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Metadata:   make(map[string]string),
	}

	// Extract class name (try both identifier for JS and type_identifier for TS)
	if nameNode := p.findChildByType(node, "identifier"); nameNode != nil {
		class.Name = p.getNodeText(nameNode, content)
	} else if nameNode := p.findChildByType(node, "type_identifier"); nameNode != nil {
		class.Name = p.getNodeText(nameNode, content)
	}

	// Extract extends clause
	if extendsNode := p.findChildByType(node, "class_heritage"); extendsNode != nil {
		// Try identifier first (simple class names)
		if extendsId := p.findChildByType(extendsNode, "identifier"); extendsId != nil {
			class.Extends = p.getNodeText(extendsId, content)
		} else if memberExpr := p.findChildByType(extendsNode, "member_expression"); memberExpr != nil {
			// Handle member expressions like React.Component
			class.Extends = p.getNodeText(memberExpr, content)
		}
	}

	// Extract implements clause (TypeScript)
	implementsNodes := p.findChildrenByType(node, "implements_clause")
	for _, implNode := range implementsNodes {
		if identifier := p.findChildByType(implNode, "identifier"); identifier != nil {
			class.Implements = append(class.Implements, p.getNodeText(identifier, content))
		}
	}

	// Extract class body
	if bodyNode := p.findChildByType(node, "class_body"); bodyNode != nil {
		p.extractClassMembers(bodyNode, content, &class)
	}

	// Check if exported
	class.IsExported = p.isExported(node)

	class.Metadata["node_type"] = node.Type()

	result.Classes = append(result.Classes, class)
	return nil
}

// extractInterface extracts TypeScript interface information
func (p *Parser) extractInterface(node *sitter.Node, content []byte, result *ParseResult) error {
	iface := InterfaceInfo{
		Extends:    []string{},
		Properties: []PropertyInfo{},
		Methods:    []FunctionInfo{},
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Metadata:   make(map[string]string),
	}

	// Extract interface name
	if nameNode := p.findChildByType(node, "type_identifier"); nameNode != nil {
		iface.Name = p.getNodeText(nameNode, content)
	}

	// Extract extends clause
	if extendsNode := p.findChildByType(node, "extends_type_clause"); extendsNode != nil {
		if identifier := p.findChildByType(extendsNode, "type_identifier"); identifier != nil {
			iface.Extends = append(iface.Extends, p.getNodeText(identifier, content))
		}
	}

	// Extract interface body
	if bodyNode := p.findChildByType(node, "interface_body"); bodyNode != nil {
		p.extractInterfaceMembers(bodyNode, content, &iface)
	}

	// Check if exported
	iface.IsExported = p.isExported(node)

	iface.Metadata["node_type"] = node.Type()

	result.Interfaces = append(result.Interfaces, iface)
	return nil
}

// extractVariables extracts variable declarations
func (p *Parser) extractVariables(node *sitter.Node, content []byte, result *ParseResult) error {
	// Extract variable kind (var, let, const)
	kind := "var"
	if node.Type() == "lexical_declaration" {
		if kindNode := p.findChildByType(node, "let"); kindNode != nil {
			kind = "let"
		} else if kindNode := p.findChildByType(node, "const"); kindNode != nil {
			kind = "const"
		}
	}

	// Extract variable declarators
	declarators := p.findChildrenByType(node, "variable_declarator")
	for _, declarator := range declarators {
		variable := VariableInfo{
			Kind:      kind,
			StartLine: int(declarator.StartPoint().Row) + 1,
			Metadata:  make(map[string]string),
		}

		// Extract variable name
		if nameNode := p.findChildByType(declarator, "identifier"); nameNode != nil {
			variable.Name = p.getNodeText(nameNode, content)
		}

		// Extract type annotation (TypeScript)
		if typeAnnotation := p.findChildByType(declarator, "type_annotation"); typeAnnotation != nil {
			variable.Type = p.getNodeText(typeAnnotation, content)
		}

		// Check if exported
		variable.IsExported = p.isExported(node)

		variable.Metadata["node_type"] = declarator.Type()

		result.Variables = append(result.Variables, variable)
	}

	return nil
}

// Helper methods for AST traversal
func (p *Parser) findChildByType(node *sitter.Node, nodeType string) *sitter.Node {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return child
		}
	}
	return nil
}

func (p *Parser) findChildrenByType(node *sitter.Node, nodeType string) []*sitter.Node {
	var children []*sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			children = append(children, child)
		}
	}
	return children
}

func (p *Parser) getNodeText(node *sitter.Node, content []byte) string {
	return string(content[node.StartByte():node.EndByte()])
}

func (p *Parser) isExported(node *sitter.Node) bool {
	// Check if node is preceded by export keyword
	parent := node.Parent()
	if parent != nil && (parent.Type() == "export_statement" || parent.Type() == "export_declaration") {
		return true
	}
	return false
}
