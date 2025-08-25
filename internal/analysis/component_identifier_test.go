package analysis

import (
	"testing"
	"reflect"
)

func TestComponentIdentifier_IdentifyReactComponent(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectedType ComponentType
		expectedMetadata map[string]interface{}
	}{
		{
			name:     "Functional React Component",
			filePath: "/src/components/Button.tsx",
			content: `
import React from 'react';

const Button = ({ onClick, children }) => {
  return <button onClick={onClick}>{children}</button>;
};

export default Button;
			`,
			expectedType: ReactComponent,
			expectedMetadata: map[string]interface{}{
				"has_jsx": true,
				"is_functional": true,
				"is_class": false,
				"uses_hooks": false,
				"detection_confidence": "high",
			},
		},
		{
			name:     "Class React Component",
			filePath: "/src/components/Counter.tsx",
			content: `
import React from 'react';

class Counter extends React.Component {
  constructor(props) {
    super(props);
    this.state = { count: 0 };
  }

  render() {
    return <div>{this.state.count}</div>;
  }
}

export default Counter;
			`,
			expectedType: ReactComponent,
			expectedMetadata: map[string]interface{}{
				"has_jsx": true,
				"is_functional": false,
				"is_class": true,
				"uses_hooks": false,
				"detection_confidence": "high",
			},
		},
		{
			name:     "React Component with Hooks",
			filePath: "/src/components/UserProfile.tsx",
			content: `
import React, { useState, useEffect } from 'react';

const UserProfile = ({ userId }) => {
  const [user, setUser] = useState(null);
  
  useEffect(() => {
    fetchUser(userId).then(setUser);
  }, [userId]);

  return user ? <div>{user.name}</div> : <div>Loading...</div>;
};

export default UserProfile;
			`,
			expectedType: ReactComponent,
			expectedMetadata: map[string]interface{}{
				"has_jsx": true,
				"is_functional": true,
				"is_class": false,
				"uses_hooks": true,
				"detection_confidence": "high",
			},
		},
		{
			name:     "Custom Hook",
			filePath: "/src/hooks/useAuth.tsx",
			content: `
import { useState, useEffect } from 'react';

const useAuth = () => {
  const [user, setUser] = useState(null);
  
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      validateToken(token).then(setUser);
    }
  }, []);

  return { user, setUser };
};

export default useAuth;
			`,
			expectedType: ReactComponent,
			expectedMetadata: map[string]interface{}{
				"has_jsx": false,
				"is_functional": false,
				"is_class": false,
				"uses_hooks": true,
				"is_custom_hook": true,
				"detection_confidence": "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := ci.IdentifyComponent(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("IdentifyComponent() error = %v", err)
			}

			if component.Type != tt.expectedType {
				t.Errorf("IdentifyComponent() type = %v, want %v", component.Type, tt.expectedType)
			}

			for key, expectedValue := range tt.expectedMetadata {
				if actualValue, exists := component.Metadata[key]; !exists || actualValue != expectedValue {
					t.Errorf("IdentifyComponent() metadata[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestComponentIdentifier_IdentifyService(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectedType ComponentType
		expectedMetadata map[string]interface{}
	}{
		{
			name:     "API Service with Axios",
			filePath: "/src/services/userService.ts",
			content: `
import axios from 'axios';

export class UserService {
  private apiUrl = 'https://api.example.com/users';

  async getUser(id: string) {
    const response = await axios.get(` + "`${this.apiUrl}/${id}`" + `);
    return response.data;
  }

  async createUser(userData: any) {
    return await axios.post(this.apiUrl, userData);
  }
}
			`,
			expectedType: Service,
			expectedMetadata: map[string]interface{}{
				"has_http_client": true,
				"has_db_operations": false,
				"has_async_patterns": true,
				"detection_confidence": "high",
			},
		},
		{
			name:     "Database Repository",
			filePath: "/src/repositories/userRepository.ts",
			content: `
export class UserRepository {
  async findOne(id: string) {
    return await this.db.query('SELECT * FROM users WHERE id = ?', [id]);
  }

  async create(userData: any) {
    const result = await this.db.query(
      'INSERT INTO users (name, email) VALUES (?, ?)',
      [userData.name, userData.email]
    );
    return result.insertId;
  }

  async update(id: string, data: any) {
    return await this.db.query(
      'UPDATE users SET name = ?, email = ? WHERE id = ?',
      [data.name, data.email, id]
    );
  }
}
			`,
			expectedType: Service,
			expectedMetadata: map[string]interface{}{
				"has_http_client": false,
				"has_db_operations": true,
				"has_async_patterns": true,
				"detection_confidence": "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := ci.IdentifyComponent(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("IdentifyComponent() error = %v", err)
			}

			if component.Type != tt.expectedType {
				t.Errorf("IdentifyComponent() type = %v, want %v", component.Type, tt.expectedType)
			}

			for key, expectedValue := range tt.expectedMetadata {
				if actualValue, exists := component.Metadata[key]; !exists || actualValue != expectedValue {
					t.Errorf("IdentifyComponent() metadata[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestComponentIdentifier_IdentifyUtility(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectedType ComponentType
		expectedMetadata map[string]interface{}
	}{
		{
			name:     "Pure Utility Functions",
			filePath: "/src/utils/formatters.ts",
			content: `
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount);
}

export const formatDate = (date: Date): string => {
  return date.toLocaleDateString('en-US');
};

export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
			`,
			expectedType: Utility,
			expectedMetadata: map[string]interface{}{
				"has_pure_functions": true,
				"detection_confidence": "medium",
			},
		},
		{
			name:     "Helper Functions",
			filePath: "/src/helpers/calculations.ts",
			content: `
export const calculateTax = (amount: number, rate: number): number => {
  return amount * (rate / 100);
};

export function calculateDiscount(price: number, discountPercent: number): number {
  return price - (price * discountPercent / 100);
}

export const roundToTwoDecimals = (num: number): number => {
  return Math.round((num + Number.EPSILON) * 100) / 100;
};
			`,
			expectedType: Utility,
			expectedMetadata: map[string]interface{}{
				"has_pure_functions": true,
				"detection_confidence": "medium",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := ci.IdentifyComponent(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("IdentifyComponent() error = %v", err)
			}

			if component.Type != tt.expectedType {
				t.Errorf("IdentifyComponent() type = %v, want %v", component.Type, tt.expectedType)
			}

			for key, expectedValue := range tt.expectedMetadata {
				if actualValue, exists := component.Metadata[key]; !exists || actualValue != expectedValue {
					t.Errorf("IdentifyComponent() metadata[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestComponentIdentifier_IdentifyConfiguration(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectedType ComponentType
		expectedMetadata map[string]interface{}
	}{
		{
			name:     "Environment Configuration",
			filePath: "/src/config/environment.ts",
			content: `
export const config = {
  apiUrl: process.env.REACT_APP_API_URL || 'http://localhost:3000/api',
  environment: process.env.NODE_ENV || 'development',
  debugMode: process.env.REACT_APP_DEBUG === 'true',
  version: process.env.REACT_APP_VERSION || '1.0.0',
};

export default config;
			`,
			expectedType: Configuration,
			expectedMetadata: map[string]interface{}{
				"has_env_vars": true,
				"is_json": false,
				"detection_confidence": "high",
			},
		},
		{
			name:     "Constants File",
			filePath: "/src/constants/index.ts",
			content: `
export const API_ENDPOINTS = {
  USERS: '/api/users',
  POSTS: '/api/posts',
  COMMENTS: '/api/comments',
};

export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  NOT_FOUND: 404,
  INTERNAL_ERROR: 500,
};

export const VALIDATION_RULES = {
  MIN_PASSWORD_LENGTH: 8,
  MAX_USERNAME_LENGTH: 30,
  EMAIL_REGEX: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
};
			`,
			expectedType: Configuration,
			expectedMetadata: map[string]interface{}{
				"has_env_vars": false,
				"is_json": false,
				"detection_confidence": "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := ci.IdentifyComponent(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("IdentifyComponent() error = %v", err)
			}

			if component.Type != tt.expectedType {
				t.Errorf("IdentifyComponent() type = %v, want %v", component.Type, tt.expectedType)
			}

			for key, expectedValue := range tt.expectedMetadata {
				if actualValue, exists := component.Metadata[key]; !exists || actualValue != expectedValue {
					t.Errorf("IdentifyComponent() metadata[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestComponentIdentifier_IdentifyMiddleware(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectedType ComponentType
		expectedMetadata map[string]interface{}
	}{
		{
			name:     "Express Authentication Middleware",
			filePath: "/src/middleware/auth.ts",
			content: `
import jwt from 'jsonwebtoken';

export const authenticateToken = (req, res, next) => {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1];

  if (!token) {
    return res.sendStatus(401);
  }

  jwt.verify(token, process.env.ACCESS_TOKEN_SECRET, (err, user) => {
    if (err) return res.sendStatus(403);
    req.user = user;
    next();
  });
};
			`,
			expectedType: Middleware,
			expectedMetadata: map[string]interface{}{
				"is_express_middleware": true,
				"is_auth_middleware": true,
				"detection_confidence": "high",
			},
		},
		{
			name:     "Logging Middleware",
			filePath: "/src/middleware/logger.ts",
			content: `
export const requestLogger = (req, res, next) => {
  const start = Date.now();
  
  res.on('finish', () => {
    const duration = Date.now() - start;
    console.log(req.method + " " + req.path + " - " + res.statusCode + " - " + duration + "ms");
  });
  
  next();
};

export default requestLogger;
			`,
			expectedType: Middleware,
			expectedMetadata: map[string]interface{}{
				"is_express_middleware": true,
				"is_auth_middleware": false,
				"detection_confidence": "high",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := ci.IdentifyComponent(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("IdentifyComponent() error = %v", err)
			}

			if component.Type != tt.expectedType {
				t.Errorf("IdentifyComponent() type = %v, want %v", component.Type, tt.expectedType)
			}

			for key, expectedValue := range tt.expectedMetadata {
				if actualValue, exists := component.Metadata[key]; !exists || actualValue != expectedValue {
					t.Errorf("IdentifyComponent() metadata[%s] = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestComponentIdentifier_ExtractExports(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Multiple Export Types",
			content: `
export function calculateTax(amount) {
  return amount * 0.1;
}

export const formatCurrency = (amount) => {
  return "$" + amount.toFixed(2);
};

export default UserComponent;

module.exports = config;
			`,
			expected: []string{"calculateTax", "formatCurrency", "default", "module"},
		},
		{
			name: "Named Exports Only",
			content: `
export const API_URL = 'https://api.example.com';
export function fetchUsers() {
  return fetch(API_URL + '/users');
}
			`,
			expected: []string{"API_URL", "fetchUsers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ci.extractExports(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractExports() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComponentIdentifier_ExtractDependencies(t *testing.T) {
	ci := NewComponentIdentifier()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "ES6 and CommonJS Imports",
			content: `
import React from 'react';
import { useState, useEffect } from 'react';
import axios from 'axios';
import './styles.css';

const config = require('./config');
const utils = require('../utils/helpers');
			`,
			expected: []string{"react", "react", "axios", "./styles.css", "./config", "../utils/helpers"},
		},
		{
			name: "Mixed Import Styles",
			content: `
import lodash from 'lodash';
const express = require('express');
import * as path from 'path';
			`,
			expected: []string{"lodash", "express", "path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ci.extractDependencies(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractDependencies() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComponentIdentifier_ExtractComponentName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "React Component File",
			filePath: "/src/components/UserProfile.tsx",
			expected: "UserProfile",
		},
		{
			name:     "Service File",
			filePath: "/src/services/userService.ts",
			expected: "userService",
		},
		{
			name:     "Index File - Use Parent Directory",
			filePath: "/src/components/Button/index.tsx",
			expected: "Button",
		},
		{
			name:     "Utility File with Extension",
			filePath: "/src/utils/formatters.util.ts",
			expected: "formatters",
		},
		{
			name:     "Configuration File",
			filePath: "/src/config/database.config.js",
			expected: "database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentName(tt.filePath)
			if result != tt.expected {
				t.Errorf("extractComponentName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestComponentIdentifier_GetComponentStats(t *testing.T) {
	ci := NewComponentIdentifier()

	// Add test components
	ci.AddComponent(Component{Type: ReactComponent, Name: "Button"})
	ci.AddComponent(Component{Type: ReactComponent, Name: "Modal"})
	ci.AddComponent(Component{Type: Service, Name: "UserService"})
	ci.AddComponent(Component{Type: Utility, Name: "formatters"})
	ci.AddComponent(Component{Type: Configuration, Name: "config"})

	expected := map[ComponentType]int{
		ReactComponent: 2,
		Service:       1,
		Utility:       1,
		Configuration: 1,
		Middleware:    0,
	}

	result := ci.GetComponentStats()
	
	for componentType, expectedCount := range expected {
		if result[componentType] != expectedCount {
			t.Errorf("GetComponentStats()[%s] = %d, want %d", componentType, result[componentType], expectedCount)
		}
	}
}

func TestComponentIdentifier_Caching(t *testing.T) {
	ci := NewComponentIdentifier()

	filePath := "/test/component.tsx"
	content := `
import React from 'react';
export const TestComponent = () => <div>Test</div>;
	`

	// First identification
	component1, err1 := ci.IdentifyComponent(filePath, content)
	if err1 != nil {
		t.Fatalf("First IdentifyComponent() error = %v", err1)
	}

	// Second identification (should use cache)
	component2, err2 := ci.IdentifyComponent(filePath, content)
	if err2 != nil {
		t.Fatalf("Second IdentifyComponent() error = %v", err2)
	}

	// Verify both results are identical (cached)
	if !reflect.DeepEqual(component1, component2) {
		t.Errorf("Cached component should be identical to original")
	}

	// Verify cache contains the component
	if cached, exists := ci.cache[filePath]; !exists {
		t.Errorf("Component should be cached")
	} else if cached.Name != component1.Name {
		t.Errorf("Cached component name mismatch")
	}
}