package ast

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// extractImport extracts import statement information
func (p *Parser) extractImport(node *sitter.Node, content []byte, result *ParseResult) error {
	importInfo := ImportInfo{
		Specifiers: []string{},
		StartLine:  int(node.StartPoint().Row) + 1,
	}

	// Extract import source (from "module")
	if sourceNode := p.findChildByType(node, "string"); sourceNode != nil {
		source := p.getNodeText(sourceNode, content)
		// Remove quotes
		importInfo.Source = strings.Trim(source, `"'`)

		// Determine if external package
		importInfo.IsExternal = p.isExternalImport(importInfo.Source)
	}

	// Extract import clause
	if clauseNode := p.findChildByType(node, "import_clause"); clauseNode != nil {
		p.extractImportClause(clauseNode, content, &importInfo)
	}

	result.Imports = append(result.Imports, importInfo)
	return nil
}

// extractExport extracts export statement information
func (p *Parser) extractExport(node *sitter.Node, content []byte, result *ParseResult) error {
	exportInfo := ExportInfo{
		Specifiers: []string{},
		StartLine:  int(node.StartPoint().Row) + 1,
	}

	// Check export type
	if p.findChildByType(node, "default") != nil {
		exportInfo.ExportType = "default"

		// Extract default export name
		if identifier := p.findChildByType(node, "identifier"); identifier != nil {
			exportInfo.Name = p.getNodeText(identifier, content)
		}
	} else if clauseNode := p.findChildByType(node, "export_clause"); clauseNode != nil {
		exportInfo.ExportType = "named"
		// Extract specifiers from export_clause
		specifiers := p.findChildrenByType(clauseNode, "export_specifier")
		for _, spec := range specifiers {
			if identifier := p.findChildByType(spec, "identifier"); identifier != nil {
				specName := p.getNodeText(identifier, content)
				exportInfo.Specifiers = append(exportInfo.Specifiers, specName)
			}
		}
	} else if p.findChildByType(node, "*") != nil {
		exportInfo.ExportType = "all"
	}

	// Extract re-export source
	if sourceNode := p.findChildByType(node, "string"); sourceNode != nil {
		source := p.getNodeText(sourceNode, content)
		exportInfo.Source = strings.Trim(source, `"'`)
	}

	result.Exports = append(result.Exports, exportInfo)
	return nil
}

// extractImportClause extracts import specifiers and types
func (p *Parser) extractImportClause(clauseNode *sitter.Node, content []byte, importInfo *ImportInfo) {
	// Default import
	if identifier := p.findChildByType(clauseNode, "identifier"); identifier != nil {
		importInfo.ImportType = "default"
		importInfo.LocalName = p.getNodeText(identifier, content)
		importInfo.Specifiers = []string{importInfo.LocalName}
		return
	}

	// Namespace import (import * as name)
	if namespaceNode := p.findChildByType(clauseNode, "namespace_import"); namespaceNode != nil {
		importInfo.ImportType = "namespace"
		if identifier := p.findChildByType(namespaceNode, "identifier"); identifier != nil {
			importInfo.LocalName = p.getNodeText(identifier, content)
			importInfo.Specifiers = []string{importInfo.LocalName}
		}
		return
	}

	// Named imports
	if namedNode := p.findChildByType(clauseNode, "named_imports"); namedNode != nil {
		importInfo.ImportType = "named"
		p.extractNamedImports(namedNode, content, importInfo)
		return
	}

	// Side-effect import (no specifiers)
	if len(importInfo.Specifiers) == 0 {
		importInfo.ImportType = "side-effect"
	}
}

// extractNamedImports extracts named import specifiers
func (p *Parser) extractNamedImports(namedNode *sitter.Node, content []byte, importInfo *ImportInfo) {
	specifiers := p.findChildrenByType(namedNode, "import_specifier")
	for _, spec := range specifiers {
		if identifier := p.findChildByType(spec, "identifier"); identifier != nil {
			specName := p.getNodeText(identifier, content)
			importInfo.Specifiers = append(importInfo.Specifiers, specName)
		}
	}
}

// extractClassMembers extracts methods and properties from class body
func (p *Parser) extractClassMembers(bodyNode *sitter.Node, content []byte, class *ClassInfo) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)

		switch child.Type() {
		case "method_definition":
			method := p.extractMethodFromClass(child, content)
			class.Methods = append(class.Methods, method)

		case "field_definition", "property_signature", "public_field_definition":
			property := p.extractProperty(child, content)
			class.Properties = append(class.Properties, property)
		}
	}
}

// extractInterfaceMembers extracts properties and methods from interface body
func (p *Parser) extractInterfaceMembers(bodyNode *sitter.Node, content []byte, iface *InterfaceInfo) {
	for i := 0; i < int(bodyNode.ChildCount()); i++ {
		child := bodyNode.Child(i)

		switch child.Type() {
		case "property_signature":
			property := p.extractProperty(child, content)
			iface.Properties = append(iface.Properties, property)

		case "method_signature":
			method := p.extractMethodSignature(child, content)
			iface.Methods = append(iface.Methods, method)
		}
	}
}

// extractMethodFromClass extracts method information from class
func (p *Parser) extractMethodFromClass(node *sitter.Node, content []byte) FunctionInfo {
	method := FunctionInfo{
		Parameters: []ParameterInfo{},
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Metadata:   make(map[string]string),
	}

	// Extract method name
	if nameNode := p.findChildByType(node, "property_identifier"); nameNode != nil {
		method.Name = p.getNodeText(nameNode, content)
	}

	// Check if async
	if p.findChildByType(node, "async") != nil {
		method.IsAsync = true
	}

	// Extract parameters
	if paramsNode := p.findChildByType(node, "formal_parameters"); paramsNode != nil {
		method.Parameters = p.extractParameters(paramsNode, content)
	}

	// Extract return type
	if typeAnnotation := p.findChildByType(node, "type_annotation"); typeAnnotation != nil {
		method.ReturnType = p.getNodeText(typeAnnotation, content)
	}

	// Check modifiers
	if p.findChildByType(node, "static") != nil {
		method.Metadata["static"] = "true"
	}
	if p.findChildByType(node, "private") != nil {
		method.Metadata["private"] = "true"
	}
	if p.findChildByType(node, "protected") != nil {
		method.Metadata["protected"] = "true"
	}

	method.Metadata["node_type"] = node.Type()

	return method
}

// extractMethodSignature extracts method signature from interface
func (p *Parser) extractMethodSignature(node *sitter.Node, content []byte) FunctionInfo {
	method := FunctionInfo{
		Parameters: []ParameterInfo{},
		StartLine:  int(node.StartPoint().Row) + 1,
		EndLine:    int(node.EndPoint().Row) + 1,
		Metadata:   make(map[string]string),
	}

	// Extract method name
	if nameNode := p.findChildByType(node, "property_identifier"); nameNode != nil {
		method.Name = p.getNodeText(nameNode, content)
	}

	// Extract parameters
	if paramsNode := p.findChildByType(node, "formal_parameters"); paramsNode != nil {
		method.Parameters = p.extractParameters(paramsNode, content)
	}

	// Extract return type
	if typeAnnotation := p.findChildByType(node, "type_annotation"); typeAnnotation != nil {
		method.ReturnType = p.getNodeText(typeAnnotation, content)
	}

	method.Metadata["node_type"] = node.Type()
	method.Metadata["signature_only"] = "true"

	return method
}

// extractProperty extracts property information
func (p *Parser) extractProperty(node *sitter.Node, content []byte) PropertyInfo {
	property := PropertyInfo{}

	// Extract property name
	if nameNode := p.findChildByType(node, "property_identifier"); nameNode != nil {
		property.Name = p.getNodeText(nameNode, content)
	}

	// Extract type annotation
	if typeAnnotation := p.findChildByType(node, "type_annotation"); typeAnnotation != nil {
		property.Type = p.getNodeText(typeAnnotation, content)
	}

	// Check modifiers
	property.IsStatic = p.findChildByType(node, "static") != nil
	property.IsReadonly = p.findChildByType(node, "readonly") != nil

	// Check for accessibility modifiers (TypeScript)
	if accessibilityNode := p.findChildByType(node, "accessibility_modifier"); accessibilityNode != nil {
		if p.findChildByType(accessibilityNode, "private") != nil {
			property.IsPrivate = true
		}
	} else {
		// Direct check for older patterns
		property.IsPrivate = p.findChildByType(node, "private") != nil
	}

	return property
}

// extractParameters extracts function parameters
func (p *Parser) extractParameters(paramsNode *sitter.Node, content []byte) []ParameterInfo {
	var parameters []ParameterInfo

	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		child := paramsNode.Child(i)

		if child.Type() == "identifier" || child.Type() == "required_parameter" || child.Type() == "optional_parameter" {
			param := ParameterInfo{}

			// Extract parameter name
			if child.Type() == "identifier" {
				param.Name = p.getNodeText(child, content)
			} else {
				if identifier := p.findChildByType(child, "identifier"); identifier != nil {
					param.Name = p.getNodeText(identifier, content)
				}
			}

			// Extract type annotation
			if typeAnnotation := p.findChildByType(child, "type_annotation"); typeAnnotation != nil {
				param.Type = p.getNodeText(typeAnnotation, content)
			}

			// Check if optional
			param.IsOptional = child.Type() == "optional_parameter" || strings.Contains(p.getNodeText(child, content), "?")

			// Extract default value
			if defaultValue := p.findChildByType(child, "assignment_pattern"); defaultValue != nil {
				if valueNode := defaultValue.Child(1); valueNode != nil {
					param.DefaultValue = p.getNodeText(valueNode, content)
				}
			}

			parameters = append(parameters, param)
		}
	}

	return parameters
}

// isExternalImport determines if an import is from an external package
func (p *Parser) isExternalImport(source string) bool {
	// External if doesn't start with . or / (relative paths)
	if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") {
		return false
	}

	// External if doesn't have file extension
	if filepath.Ext(source) == "" {
		return true
	}

	// Consider it external if it looks like a package name
	return !strings.Contains(source, "/") || !strings.HasPrefix(source, "./") || !strings.HasPrefix(source, "../")
}
