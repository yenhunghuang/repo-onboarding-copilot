package ast

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	require.NotNil(t, parser)

	defer parser.Close()

	// Test that parsers are initialized
	assert.NotNil(t, parser.jsParser)
	assert.NotNil(t, parser.tsParser)
	assert.NotNil(t, parser.tsxParser)
}

func TestParser_IsSupported(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	tests := []struct {
		filePath string
		expected bool
	}{
		{"test.js", true},
		{"test.jsx", true},
		{"test.ts", true},
		{"test.tsx", true},
		{"test.py", false},
		{"test.go", false},
		{"test.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := parser.IsSupported(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_ParseFile_JavaScript(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	jsCode := `
// Simple JavaScript function
function greet(name) {
    return "Hello, " + name;
}

// Arrow function
const add = (a, b) => a + b;

// Class
class Calculator {
    constructor() {
        this.value = 0;
    }
    
    add(x) {
        this.value += x;
        return this;
    }
}

// Variable declarations
var oldVar = "old";
let newVar = "new";
const CONSTANT = 42;

// Export
export { greet, Calculator };
export default add;
`

	result, err := parser.ParseFile(context.Background(), "test.js", []byte(jsCode))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify basic metadata
	assert.Equal(t, "test.js", result.FilePath)
	assert.Equal(t, "javascript", result.Language)

	// Verify functions were extracted (including class methods)
	assert.GreaterOrEqual(t, len(result.Functions), 2) // greet, add, and class methods

	// Verify greet function
	greetFunc := findFunctionByName(result.Functions, "greet")
	require.NotNil(t, greetFunc)
	assert.Equal(t, "greet", greetFunc.Name)
	assert.Len(t, greetFunc.Parameters, 1)
	assert.Equal(t, "name", greetFunc.Parameters[0].Name)
	assert.False(t, greetFunc.IsAsync)

	// Verify classes were extracted
	assert.Len(t, result.Classes, 1)
	calcClass := result.Classes[0]
	assert.Equal(t, "Calculator", calcClass.Name)
	assert.Len(t, calcClass.Methods, 2) // constructor and add

	// Verify variables were extracted
	assert.GreaterOrEqual(t, len(result.Variables), 3) // oldVar, newVar, CONSTANT

	// Verify exports were extracted
	assert.GreaterOrEqual(t, len(result.Exports), 2) // named and default exports
}

func TestParser_ParseFile_TypeScript(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	tsCode := `
import { Component } from 'react';
import axios from 'axios';

// Interface
interface User {
    id: number;
    name: string;
    email?: string;
}

// Type alias
type Status = 'pending' | 'completed' | 'failed';

// Function with types
async function fetchUser(id: number): Promise<User> {
    const response = await axios.get<User>(\"/api/users/\" + id);
    return response.data;
}

// Class with TypeScript features
class UserService {
    private apiUrl: string;
    
    constructor(apiUrl: string) {
        this.apiUrl = apiUrl;
    }
    
    async getUser(id: number): Promise<User> {
        return fetchUser(id);
    }
    
    static validateUser(user: User): boolean {
        return !!user.id && !!user.name;
    }
}

// Export
export { UserService, fetchUser };
export type { User, Status };
`

	result, err := parser.ParseFile(context.Background(), "test.ts", []byte(tsCode))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify basic metadata
	assert.Equal(t, "test.ts", result.FilePath)
	assert.Equal(t, "typescript", result.Language)

	// Verify interfaces were extracted
	assert.Len(t, result.Interfaces, 1)
	userInterface := result.Interfaces[0]
	assert.Equal(t, "User", userInterface.Name)
	assert.Len(t, userInterface.Properties, 3) // id, name, email

	// Verify functions with async
	assert.GreaterOrEqual(t, len(result.Functions), 1)
	fetchFunc := findFunctionByName(result.Functions, "fetchUser")
	require.NotNil(t, fetchFunc)
	assert.True(t, fetchFunc.IsAsync)
	assert.Contains(t, fetchFunc.ReturnType, "Promise")

	// Verify classes
	assert.Len(t, result.Classes, 1)
	userService := result.Classes[0]
	assert.Equal(t, "UserService", userService.Name)
	assert.GreaterOrEqual(t, len(userService.Methods), 2)    // constructor, getUser, validateUser
	assert.GreaterOrEqual(t, len(userService.Properties), 1) // apiUrl

	// Verify imports
	assert.GreaterOrEqual(t, len(result.Imports), 2)

	// Check for external vs internal imports
	reactImport := findImportBySource(result.Imports, "react")
	require.NotNil(t, reactImport)
	assert.True(t, reactImport.IsExternal)
}

func TestParser_ParseFile_JSX(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	jsxCode := `
import React from 'react';

// React component
function Welcome(props) {
    return <h1>Hello, {props.name}!</h1>;
}

// Class component
class Button extends React.Component {
    handleClick() {
        console.log('Button clicked');
    }
    
    render() {
        return (
            <button onClick={this.handleClick}>
                {this.props.children}
            </button>
        );
    }
}

export default Welcome;
export { Button };
`

	result, err := parser.ParseFile(context.Background(), "test.jsx", []byte(jsxCode))
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify language detection
	assert.Equal(t, "javascript", result.Language)

	// Verify components are detected as functions/classes
	assert.GreaterOrEqual(t, len(result.Functions), 1)
	assert.GreaterOrEqual(t, len(result.Classes), 1)

	welcomeFunc := findFunctionByName(result.Functions, "Welcome")
	require.NotNil(t, welcomeFunc)
	assert.Equal(t, "Welcome", welcomeFunc.Name)

	buttonClass := findClassByName(result.Classes, "Button")
	require.NotNil(t, buttonClass)
	assert.Equal(t, "Button", buttonClass.Name)
	assert.Equal(t, "React.Component", buttonClass.Extends)
}

func TestParser_ParseFile_ErrorHandling(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	// Test with malformed code
	malformedCode := `
function incomplete( {
    return "missing closing brace"
// missing closing brace
`

	result, err := parser.ParseFile(context.Background(), "malformed.js", []byte(malformedCode))

	// Parser should handle malformed code gracefully
	// It might succeed with partial parsing or return an error
	if err != nil {
		// Error is acceptable for malformed code
		assert.Contains(t, err.Error(), "malformed.js")
	} else {
		// If parsing succeeded, check if errors were recorded
		assert.NotNil(t, result)
	}
}

func TestParser_ParseFile_UnsupportedFile(t *testing.T) {
	parser, err := NewParser()
	require.NoError(t, err)
	defer parser.Close()

	result, err := parser.ParseFile(context.Background(), "test.py", []byte("print('hello')"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported file type")
}

// Helper functions for tests
func findFunctionByName(functions []FunctionInfo, name string) *FunctionInfo {
	for i := range functions {
		if functions[i].Name == name {
			return &functions[i]
		}
	}
	return nil
}

func findClassByName(classes []ClassInfo, name string) *ClassInfo {
	for i := range classes {
		if classes[i].Name == name {
			return &classes[i]
		}
	}
	return nil
}

func findImportBySource(imports []ImportInfo, source string) *ImportInfo {
	for i := range imports {
		if imports[i].Source == source {
			return &imports[i]
		}
	}
	return nil
}
