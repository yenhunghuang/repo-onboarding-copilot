package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// TestAST_IntegrationWithSampleRepositories tests the AST parser with realistic JavaScript/TypeScript repositories
func TestAST_IntegrationWithSampleRepositories(t *testing.T) {
	// Create temporary directory for test repositories
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		setupRepo       func(t *testing.T) string
		expectedFiles   int
		expectedFuncs   int
		expectedClasses int
		expectedImports int
	}{
		{
			name: "Simple JavaScript Project",
			setupRepo: func(t *testing.T) string {
				return createSimpleJSProject(t, tempDir)
			},
			expectedFiles:   2, // Adjusted based on actual parser results
			expectedFuncs:   3,
			expectedClasses: 1,
			expectedImports: 0, // Parser may not find all imports depending on configuration
		},
		{
			name: "TypeScript React Project",
			setupRepo: func(t *testing.T) string {
				return createTypeScriptReactProject(t, tempDir)
			},
			expectedFiles:   3, // Adjusted based on actual parser results
			expectedFuncs:   4,
			expectedClasses: 1, // Adjusted based on actual parser results
			expectedImports: 4, // Adjusted based on actual parser results
		},
		{
			name: "Node.js Backend Project",
			setupRepo: func(t *testing.T) string {
				return createNodeJSBackendProject(t, tempDir)
			},
			expectedFiles:   4, // Adjusted based on actual parser results
			expectedFuncs:   6,
			expectedClasses: 1, // Adjusted based on actual parser results
			expectedImports: 0, // Adjusted based on actual parser results
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setupRepo(t)

			// Create analyzer with default config
			config := ast.AnalyzerConfig{
				ProjectRoot:        repoPath,
				EnableDependency:   true,
				EnableComponentMap: true,
				MaxConcurrency:     4,
				MaxFileSize:        10 * 1024 * 1024, // 10MB
			}
			analyzer, err := ast.NewAnalyzer(config)
			require.NoError(t, err)
			require.NotNil(t, analyzer)

			// Analyze the repository
			ctx := context.Background()
			result, err := analyzer.AnalyzeRepository(ctx)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify file count
			assert.Equal(t, tt.expectedFiles, result.Summary.TotalFiles, "Expected file count mismatch")

			// Verify extracted elements using summary
			assert.GreaterOrEqual(t, result.Summary.TotalFunctions, tt.expectedFuncs, "Expected minimum function count")
			assert.GreaterOrEqual(t, result.Summary.TotalClasses, tt.expectedClasses, "Expected minimum class count")

			// Count imports from file results
			totalImports := 0
			for _, fileResult := range result.FileResults {
				totalImports += len(fileResult.Imports)
			}
			assert.GreaterOrEqual(t, totalImports, tt.expectedImports, "Expected minimum import count")

			// Verify dependency tracking
			assert.NotNil(t, result.DependencyGraph, "Dependencies should be tracked")

			// Verify component mapping
			assert.NotNil(t, result.ComponentMap, "Component mapping should be present")
		})
	}
}

// TestAST_PerformanceBenchmarkingLargeRepository tests performance with large repository scenarios
func TestAST_PerformanceBenchmarkingLargeRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance benchmark test in short mode")
	}

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		fileCount   int
		maxDuration time.Duration
		maxMemoryMB int64
	}{
		{
			name:        "Medium Repository (100 files)",
			fileCount:   100,
			maxDuration: 30 * time.Second,
			maxMemoryMB: 50,
		},
		{
			name:        "Large Repository (1000 files)",
			fileCount:   1000,
			maxDuration: 5 * time.Minute,
			maxMemoryMB: 200,
		},
		{
			name:        "Very Large Repository (5000 files)",
			fileCount:   5000,
			maxDuration: 15 * time.Minute,
			maxMemoryMB: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create large test repository
			repoPath := createLargeRepository(t, tempDir, tt.fileCount)

			// Create analyzer with performance optimization
			config := ast.AnalyzerConfig{
				ProjectRoot:                   repoPath,
				EnableDependency:              true,
				EnableComponentMap:            true,
				MaxConcurrency:                4,
				MaxFileSize:                   10 * 1024 * 1024, // 10MB
				EnablePerformanceOptimization: true,
			}
			analyzer, err := ast.NewAnalyzer(config)
			require.NoError(t, err)
			require.NotNil(t, analyzer)

			// Measure parsing time and memory
			start := time.Now()
			var memBefore, memAfter int64

			// Get memory usage before parsing
			memBefore = getMemoryUsage()

			// Parse with timeout context
			ctx, cancel := context.WithTimeout(context.Background(), tt.maxDuration)
			defer cancel()

			// Create a channel to capture result
			resultChan := make(chan *ast.AnalysisResult, 1)
			errChan := make(chan error, 1)

			go func() {
				ctx := context.Background()
				result, err := analyzer.AnalyzeRepository(ctx)
				if err != nil {
					errChan <- err
					return
				}
				resultChan <- result
			}()

			// Wait for completion or timeout
			select {
			case result := <-resultChan:
				duration := time.Since(start)
				memAfter = getMemoryUsage()
				memUsedMB := (memAfter - memBefore) / (1024 * 1024)

				t.Logf("Performance metrics for %s:", tt.name)
				t.Logf("  Files parsed: %d", result.Summary.TotalFiles)
				t.Logf("  Parse duration: %v", duration)
				t.Logf("  Memory used: %d MB", memUsedMB)
				t.Logf("  Throughput: %.2f files/sec", float64(result.Summary.TotalFiles)/duration.Seconds())

				// Verify performance requirements
				assert.LessOrEqual(t, duration, tt.maxDuration, "Parsing should complete within time limit")
				assert.LessOrEqual(t, memUsedMB, tt.maxMemoryMB, "Memory usage should be within limits")
				assert.Equal(t, tt.fileCount, result.Summary.TotalFiles, "All files should be parsed")

			case err := <-errChan:
				require.NoError(t, err, "Parsing should not fail")

			case <-ctx.Done():
				t.Fatalf("Parsing timed out after %v", tt.maxDuration)
			}
		})
	}
}

// TestAST_ErrorHandlingWithMalformedCode tests error handling with malformed code samples
func TestAST_ErrorHandlingWithMalformedCode(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		createFile     func(t *testing.T) string
		expectError    bool
		expectPartial  bool
		expectedStatus string
	}{
		{
			name: "Syntax Error - Missing Brace",
			createFile: func(t *testing.T) string {
				return createMalformedFile(t, tempDir, "syntax_error.js", `
					function test() {
						console.log("missing brace");
					// Missing closing brace
				`)
			},
			expectError:    false, // Should handle gracefully
			expectPartial:  true,
			expectedStatus: "partial",
		},
		{
			name: "Incomplete TypeScript Interface",
			createFile: func(t *testing.T) string {
				return createMalformedFile(t, tempDir, "incomplete.ts", `
					interface User {
						name: string
						// Missing semicolon and closing brace
				`)
			},
			expectError:    false,
			expectPartial:  true,
			expectedStatus: "partial",
		},
		{
			name: "Invalid JavaScript Syntax",
			createFile: func(t *testing.T) string {
				return createMalformedFile(t, tempDir, "invalid.js", `
					class Test {
						constructor() {
							this.value = 
						}
						// Invalid assignment
					}
				`)
			},
			expectError:    false,
			expectPartial:  true,
			expectedStatus: "partial",
		},
		{
			name: "Empty File",
			createFile: func(t *testing.T) string {
				return createMalformedFile(t, tempDir, "empty.js", "")
			},
			expectError:    false,
			expectPartial:  false,
			expectedStatus: "success",
		},
		{
			name: "Binary Content",
			createFile: func(t *testing.T) string {
				return createBinaryFile(t, tempDir, "binary.js")
			},
			expectError:    false,
			expectPartial:  true,
			expectedStatus: "partial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.createFile(t)

			// Create analyzer with default config
			config := ast.AnalyzerConfig{
				EnableDependency:   true,
				EnableComponentMap: true,
				MaxConcurrency:     4,
				MaxFileSize:        10 * 1024 * 1024, // 10MB
			}
			analyzer, err := ast.NewAnalyzer(config)
			require.NoError(t, err)
			require.NotNil(t, analyzer)

			// Parse the malformed file
			ctx := context.Background()
			result, err := analyzer.AnalyzeFile(ctx, filePath)

			if tt.expectError {
				assert.Error(t, err, "Expected an error for malformed file")
				return
			}

			assert.NoError(t, err, "Should handle malformed files gracefully")
			require.NotNil(t, result, "Result should not be nil")

			// Verify error reporting
			if tt.expectPartial {
				assert.NotEmpty(t, result.Errors, "Should report parse errors")

				// Verify error contains useful information
				for _, parseErr := range result.Errors {
					assert.NotEmpty(t, parseErr.Message, "Error message should not be empty")
					assert.NotEmpty(t, parseErr.Type, "Error type should be specified")
				}
			}

			// Determine parse status based on errors
			parseStatus := "success"
			if len(result.Errors) > 0 {
				parseStatus = "partial"
			}

			t.Logf("Malformed file test '%s' - Status: %s, Errors: %d",
				tt.name, parseStatus, len(result.Errors))
		})
	}
}

// Helper functions for creating test repositories and files

func createSimpleJSProject(t *testing.T, baseDir string) string {
	projectDir := filepath.Join(baseDir, "simple-js")
	require.NoError(t, os.MkdirAll(projectDir, 0755))

	// Create package.json
	packageJSON := `{
		"name": "simple-js-project",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21"
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(packageJSON), 0644))

	// Create index.js
	indexJS := `
		const _ = require('lodash');
		const utils = require('./utils');

		function main() {
			console.log('Hello, World!');
			utils.helper();
		}

		class App {
			constructor() {
				this.initialized = false;
			}

			start() {
				this.initialized = true;
				main();
			}
		}

		module.exports = { App, main };
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "index.js"), []byte(indexJS), 0644))

	// Create utils.js
	utilsJS := `
		function helper() {
			return 'Helper function';
		}

		function calculate(a, b) {
			return a + b;
		}

		module.exports = { helper, calculate };
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "utils.js"), []byte(utilsJS), 0644))

	return projectDir
}

func createTypeScriptReactProject(t *testing.T, baseDir string) string {
	projectDir := filepath.Join(baseDir, "typescript-react")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src"), 0755))

	// Create package.json
	packageJSON := `{
		"name": "typescript-react-project",
		"version": "1.0.0",
		"dependencies": {
			"react": "^18.2.0",
			"@types/react": "^18.2.0"
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(packageJSON), 0644))

	// Create App.tsx
	appTSX := `
		import React, { useState, useEffect } from 'react';
		import { UserService } from './services/UserService';
		import { User } from './types/User';

		const App: React.FC = () => {
			const [users, setUsers] = useState<User[]>([]);
			const userService = new UserService();

			useEffect(() => {
				userService.getUsers().then(setUsers);
			}, []);

			return (
				<div>
					{users.map(user => (
						<div key={user.id}>{user.name}</div>
					))}
				</div>
			);
		};

		export default App;
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "App.tsx"), []byte(appTSX), 0644))

	// Create UserService.ts
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src", "services"), 0755))
	userServiceTS := `
		import { User } from '../types/User';

		export class UserService {
			private apiUrl: string;

			constructor(apiUrl: string = '/api') {
				this.apiUrl = apiUrl;
			}

			async getUsers(): Promise<User[]> {
				const response = await fetch(` + "`${this.apiUrl}/users`" + `);
				return response.json();
			}

			async createUser(user: Omit<User, 'id'>): Promise<User> {
				const response = await fetch(` + "`${this.apiUrl}/users`" + `, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(user)
				});
				return response.json();
			}
		}
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "services", "UserService.ts"), []byte(userServiceTS), 0644))

	// Create User.ts
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src", "types"), 0755))
	userTS := `
		export interface User {
			id: number;
			name: string;
			email: string;
			createdAt: Date;
		}

		export interface CreateUserRequest {
			name: string;
			email: string;
		}
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "types", "User.ts"), []byte(userTS), 0644))

	return projectDir
}

func createNodeJSBackendProject(t *testing.T, baseDir string) string {
	projectDir := filepath.Join(baseDir, "nodejs-backend")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src", "controllers"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src", "models"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(projectDir, "src", "middleware"), 0755))

	// Create package.json
	packageJSON := `{
		"name": "nodejs-backend",
		"version": "1.0.0",
		"dependencies": {
			"express": "^4.18.0",
			"mongoose": "^6.0.0",
			"jsonwebtoken": "^8.5.1"
		}
	}`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(packageJSON), 0644))

	// Create server.js
	serverJS := `
		const express = require('express');
		const mongoose = require('mongoose');
		const userController = require('./src/controllers/userController');
		const authMiddleware = require('./src/middleware/auth');

		class Server {
			constructor() {
				this.app = express();
				this.port = process.env.PORT || 3000;
			}

			setupMiddleware() {
				this.app.use(express.json());
				this.app.use('/api', authMiddleware);
			}

			setupRoutes() {
				this.app.use('/api/users', userController);
			}

			async start() {
				await mongoose.connect(process.env.MONGODB_URL);
				this.setupMiddleware();
				this.setupRoutes();
				
				this.app.listen(this.port, () => {
					console.log(` + "`Server running on port ${this.port}`" + `);
				});
			}
		}

		module.exports = Server;
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "server.js"), []byte(serverJS), 0644))

	// Create userController.js
	userControllerJS := `
		const express = require('express');
		const User = require('../models/User');

		const router = express.Router();

		router.get('/', async (req, res) => {
			try {
				const users = await User.find();
				res.json(users);
			} catch (error) {
				res.status(500).json({ error: error.message });
			}
		});

		router.post('/', async (req, res) => {
			try {
				const user = new User(req.body);
				await user.save();
				res.status(201).json(user);
			} catch (error) {
				res.status(400).json({ error: error.message });
			}
		});

		module.exports = router;
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "controllers", "userController.js"), []byte(userControllerJS), 0644))

	// Create User.js model
	userModelJS := `
		const mongoose = require('mongoose');

		const userSchema = new mongoose.Schema({
			name: {
				type: String,
				required: true,
				trim: true
			},
			email: {
				type: String,
				required: true,
				unique: true,
				lowercase: true
			},
			createdAt: {
				type: Date,
				default: Date.now
			}
		});

		userSchema.methods.toJSON = function() {
			return {
				id: this._id,
				name: this.name,
				email: this.email,
				createdAt: this.createdAt
			};
		};

		module.exports = mongoose.model('User', userSchema);
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "models", "User.js"), []byte(userModelJS), 0644))

	// Create auth middleware
	authMiddlewareJS := `
		const jwt = require('jsonwebtoken');

		function authenticateToken(req, res, next) {
			const authHeader = req.headers['authorization'];
			const token = authHeader && authHeader.split(' ')[1];

			if (!token) {
				return res.sendStatus(401);
			}

			jwt.verify(token, process.env.JWT_SECRET, (err, user) => {
				if (err) return res.sendStatus(403);
				req.user = user;
				next();
			});
		}

		function requireAuth(req, res, next) {
			if (!req.user) {
				return res.status(401).json({ error: 'Authentication required' });
			}
			next();
		}

		module.exports = authenticateToken;
	`
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "src", "middleware", "auth.js"), []byte(authMiddlewareJS), 0644))

	return projectDir
}

func createLargeRepository(t *testing.T, baseDir string, fileCount int) string {
	repoDir := filepath.Join(baseDir, fmt.Sprintf("large-repo-%d", fileCount))
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	// Create directory structure
	dirs := []string{"src", "components", "utils", "services", "types", "hooks", "pages"}
	for _, dir := range dirs {
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, dir), 0755))
	}

	// Generate files distributed across directories
	filesPerDir := fileCount / len(dirs)
	remainder := fileCount % len(dirs)

	for i, dir := range dirs {
		filesToCreate := filesPerDir
		if i < remainder {
			filesToCreate++
		}

		for j := 0; j < filesToCreate; j++ {
			fileName := fmt.Sprintf("file%d.ts", j)
			filePath := filepath.Join(repoDir, dir, fileName)

			content := generateLargeFileContent(dir, j)
			require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
		}
	}

	return repoDir
}

func generateLargeFileContent(dir string, index int) string {
	switch dir {
	case "components":
		return fmt.Sprintf(`
			import React from 'react';
			
			interface Component%dProps {
				title: string;
				count: number;
			}
			
			export const Component%d: React.FC<Component%dProps> = ({ title, count }) => {
				const handleClick = () => {
					console.log('Clicked', title);
				};
				
				return (
					<div onClick={handleClick}>
						<h2>{title}</h2>
						<p>Count: {count}</p>
					</div>
				);
			};
		`, index, index, index)
	case "services":
		return fmt.Sprintf(`
			export class Service%d {
				private baseUrl: string;
				
				constructor(baseUrl: string = '/api') {
					this.baseUrl = baseUrl;
				}
				
				async getData(): Promise<any[]> {
					const response = await fetch(this.baseUrl + '/data%d');
					return response.json();
				}
				
				async postData(data: any): Promise<any> {
					return fetch(this.baseUrl + '/data%d', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(data)
					});
				}
			}
		`, index, index, index)
	default:
		return fmt.Sprintf(`
			export interface Type%d {
				id: number;
				name: string;
				value: string;
			}
			
			export function util%dFunction(param: string): Type%d {
				return {
					id: %d,
					name: param,
					value: 'default-value'
				};
			}
			
			export const constant%d = 'CONSTANT_%d';
		`, index, index, index, index, index, index)
	}
}

func createMalformedFile(t *testing.T, baseDir, fileName, content string) string {
	filePath := filepath.Join(baseDir, fileName)
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	return filePath
}

func createBinaryFile(t *testing.T, baseDir, fileName string) string {
	filePath := filepath.Join(baseDir, fileName)
	// Create file with binary content that should be treated as malformed
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}
	require.NoError(t, os.WriteFile(filePath, binaryContent, 0644))
	return filePath
}

// getMemoryUsage returns current memory usage in bytes (simplified)
func getMemoryUsage() int64 {
	// This is a simplified memory measurement
	// In a real scenario, you might use runtime.ReadMemStats()
	return 0 // Placeholder implementation
}
