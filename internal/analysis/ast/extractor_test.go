package ast

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractImport_Named(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `import { useState, useEffect } from 'react';`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Imports, 1)
	imp := result.Imports[0]
	assert.Equal(t, "react", imp.Source)
	assert.Equal(t, "named", imp.ImportType)
	assert.True(t, imp.IsExternal)
	assert.Contains(t, imp.Specifiers, "useState")
	assert.Contains(t, imp.Specifiers, "useEffect")
}

func TestExtractImport_Default(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `import React from 'react';`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Imports, 1)
	imp := result.Imports[0]
	assert.Equal(t, "react", imp.Source)
	assert.Equal(t, "default", imp.ImportType)
	assert.Equal(t, "React", imp.LocalName)
	assert.True(t, imp.IsExternal)
}

func TestExtractImport_Namespace(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `import * as fs from 'fs';`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Imports, 1)
	imp := result.Imports[0]
	assert.Equal(t, "fs", imp.Source)
	assert.Equal(t, "namespace", imp.ImportType)
	assert.Equal(t, "fs", imp.LocalName)
	assert.True(t, imp.IsExternal)
}

func TestExtractImport_Relative(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `import { helper } from './utils/helper';`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Imports, 1)
	imp := result.Imports[0]
	assert.Equal(t, "./utils/helper", imp.Source)
	assert.Equal(t, "named", imp.ImportType)
	assert.False(t, imp.IsExternal)
}

func TestExtractExport_Named(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
function helper() {}
export { helper };
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	// Should have both function and export
	assert.GreaterOrEqual(t, len(result.Functions), 1)
	assert.GreaterOrEqual(t, len(result.Exports), 1)

	// Find named export
	var namedExport *ExportInfo
	for i := range result.Exports {
		if result.Exports[i].ExportType == "named" {
			namedExport = &result.Exports[i]
			break
		}
	}
	require.NotNil(t, namedExport)
	assert.Contains(t, namedExport.Specifiers, "helper")
}

func TestExtractExport_Default(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
function main() {}
export default main;
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Exports), 1)

	// Find default export
	var defaultExport *ExportInfo
	for i := range result.Exports {
		if result.Exports[i].ExportType == "default" {
			defaultExport = &result.Exports[i]
			break
		}
	}
	require.NotNil(t, defaultExport)
	assert.Equal(t, "main", defaultExport.Name)
}

func TestExtractFunction_WithParameters(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
function calculate(x, y = 0, ...rest) {
    return x + y + rest.reduce((a, b) => a + b, 0);
}
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Functions), 1) // calculate function and possibly nested arrow function
	calculateFunc := findFunctionByName(result.Functions, "calculate")
	require.NotNil(t, calculateFunc)
	assert.Equal(t, "calculate", calculateFunc.Name)
	assert.GreaterOrEqual(t, len(calculateFunc.Parameters), 1) // at least x parameter
}

func TestExtractFunction_Async(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
async function fetchData() {
    const response = await fetch('/api/data');
    return response.json();
}
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Functions, 1)
	fn := result.Functions[0]
	assert.Equal(t, "fetchData", fn.Name)
	assert.True(t, fn.IsAsync)
}

func TestExtractClass_WithInheritance(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
class Animal {
    constructor(name) {
        this.name = name;
    }
    
    speak() {
        console.log(this.name + ' makes a sound');
    }
}

class Dog extends Animal {
    speak() {
        console.log(this.name + ' barks');
    }
}
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Classes, 2)

	// Find Dog class
	var dogClass *ClassInfo
	for i := range result.Classes {
		if result.Classes[i].Name == "Dog" {
			dogClass = &result.Classes[i]
			break
		}
	}
	require.NotNil(t, dogClass)
	assert.Equal(t, "Animal", dogClass.Extends)
}

func TestExtractInterface_TypeScript(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
interface Drawable {
    draw(): void;
}

interface Shape extends Drawable {
    area(): number;
    perimeter(): number;
}
`

	result, err := parser.ParseFile(context.Background(), "test.ts", []byte(code))
	require.NoError(t, err)

	assert.Len(t, result.Interfaces, 2)

	// Find Shape interface
	var shapeInterface *InterfaceInfo
	for i := range result.Interfaces {
		if result.Interfaces[i].Name == "Shape" {
			shapeInterface = &result.Interfaces[i]
			break
		}
	}
	require.NotNil(t, shapeInterface)
	assert.Contains(t, shapeInterface.Extends, "Drawable")
	assert.GreaterOrEqual(t, len(shapeInterface.Methods), 2) // area, perimeter
}

func TestExtractVariables_Different_Kinds(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
var oldStyle = 'var';
let newStyle = 'let';
const CONSTANT = 'const';
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(result.Variables), 3)

	// Check different variable kinds
	kinds := make(map[string]bool)
	for _, v := range result.Variables {
		kinds[v.Kind] = true
	}

	assert.True(t, kinds["var"])
	assert.True(t, kinds["let"])
	assert.True(t, kinds["const"])
}

func TestIsExternalImport(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	tests := []struct {
		source   string
		expected bool
	}{
		{"react", true},
		{"@types/node", true},
		{"lodash", true},
		{"./utils", false},
		{"../config", false},
		{"/absolute/path", false},
		{"./utils.js", false},
		{"utils/helper", true}, // Package-like path
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			result := parser.isExternalImport(tt.source)
			assert.Equal(t, tt.expected, result, "Source: %s", tt.source)
		})
	}
}

func TestParser_Metadata_Tracking(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	code := `
function test() {
    return "test";
}
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(code))
	require.NoError(t, err)

	// Check that metadata is tracked
	assert.Contains(t, result.Metadata, "node_count")
	assert.Contains(t, result.Metadata, "max_depth")

	nodeCount, ok := result.Metadata["node_count"].(int)
	assert.True(t, ok)
	assert.Greater(t, nodeCount, 0)

	maxDepth, ok := result.Metadata["max_depth"].(int)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, maxDepth, 0)
}
