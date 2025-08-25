package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// BenchmarkAST_SmallRepository benchmarks parsing performance with small repositories
func BenchmarkAST_SmallRepository(b *testing.B) {
	tempDir := b.TempDir()
	repoPath := createSmallBenchmarkRepo(b, tempDir, 10) // 10 files

	config := ast.AnalyzerConfig{
		ProjectRoot:                   repoPath,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
	result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result.Summary.TotalFiles != 10 {
			b.Fatalf("Expected 10 files, got %d", result.Summary.TotalFiles)
		}
	}
}

// BenchmarkAST_MediumRepository benchmarks parsing performance with medium repositories
func BenchmarkAST_MediumRepository(b *testing.B) {
	tempDir := b.TempDir()
	repoPath := createSmallBenchmarkRepo(b, tempDir, 100) // 100 files

	config := ast.AnalyzerConfig{
		ProjectRoot:                   repoPath,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
	result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result.Summary.TotalFiles != 100 {
			b.Fatalf("Expected 100 files, got %d", result.Summary.TotalFiles)
		}
	}
}

// BenchmarkAST_LargeRepository benchmarks parsing performance with large repositories
func BenchmarkAST_LargeRepository(b *testing.B) {
	tempDir := b.TempDir()
	repoPath := createSmallBenchmarkRepo(b, tempDir, 1000) // 1000 files

	config := ast.AnalyzerConfig{
		ProjectRoot:                   repoPath,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
	result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result.Summary.TotalFiles != 1000 {
			b.Fatalf("Expected 1000 files, got %d", result.Summary.TotalFiles)
		}
	}
}

// BenchmarkAST_SingleFileJavaScript benchmarks single JavaScript file parsing
func BenchmarkAST_SingleFileJavaScript(b *testing.B) {
	tempDir := b.TempDir()
	filePath := createBenchmarkJSFile(b, tempDir)

	config := ast.AnalyzerConfig{
		ProjectRoot:                   tempDir,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		result, err := analyzer.AnalyzeFile(ctx, filePath)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result == nil {
			b.Fatal("Result should not be nil")
		}
	}
}

// BenchmarkAST_SingleFileTypeScript benchmarks single TypeScript file parsing
func BenchmarkAST_SingleFileTypeScript(b *testing.B) {
	tempDir := b.TempDir()
	filePath := createBenchmarkTSFile(b, tempDir)

	config := ast.AnalyzerConfig{
		ProjectRoot:                   tempDir,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		result, err := analyzer.AnalyzeFile(ctx, filePath)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result == nil {
			b.Fatal("Result should not be nil")
		}
	}
}

// BenchmarkAST_MemoryUsage measures memory usage during parsing
func BenchmarkAST_MemoryUsage(b *testing.B) {
	tempDir := b.TempDir()
	repoPath := createSmallBenchmarkRepo(b, tempDir, 500) // 500 files

	config := ast.AnalyzerConfig{
		ProjectRoot:                   repoPath,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	var memStart, memEnd runtime.MemStats

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		runtime.GC()
		runtime.ReadMemStats(&memStart)

		ctx := context.Background()
	result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result == nil {
			b.Fatal("Result should not be nil")
		}

		runtime.ReadMemStats(&memEnd)

		// Report memory usage
		memUsed := memEnd.Alloc - memStart.Alloc
		b.ReportMetric(float64(memUsed), "bytes/op")
	}
}

// BenchmarkAST_ParsingThroughput measures parsing throughput (files per second)
func BenchmarkAST_ParsingThroughput(b *testing.B) {
	tempDir := b.TempDir()
	repoPath := createSmallBenchmarkRepo(b, tempDir, 200) // 200 files

	config := ast.AnalyzerConfig{
		ProjectRoot:                   repoPath,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	
	start := time.Now()
	totalFiles := 0

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
	result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		totalFiles += result.Summary.TotalFiles
	}

	duration := time.Since(start)
	throughput := float64(totalFiles) / duration.Seconds()

	b.ReportMetric(throughput, "files/sec")
}

// BenchmarkAST_DifferentFileSizes benchmarks parsing files of different sizes
func BenchmarkAST_DifferentFileSizes(b *testing.B) {
	sizes := []struct {
		name string
		size int // number of functions to generate
	}{
		{"Small", 10},
		{"Medium", 50},
		{"Large", 200},
		{"XLarge", 500},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			tempDir := b.TempDir()
			filePath := createLargeJSFile(b, tempDir, size.size)

			config := ast.AnalyzerConfig{
		ProjectRoot:                   tempDir,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				ctx := context.Background()
		result, err := analyzer.AnalyzeFile(ctx, filePath)
				if err != nil {
					b.Fatalf("Analysis failed: %v", err)
				}
				if len(result.Functions) < size.size {
					b.Fatalf("Expected at least %d functions, got %d", size.size, len(result.Functions))
				}
			}
		})
	}
}

// BenchmarkAST_ConcurrentParsing benchmarks concurrent file parsing
func BenchmarkAST_ConcurrentParsing(b *testing.B) {
	tempDir := b.TempDir()
	
	// Create multiple files for concurrent parsing
	files := make([]string, 20)
	for i := 0; i < 20; i++ {
		files[i] = createBenchmarkJSFile(b, filepath.Join(tempDir, fmt.Sprintf("file%d", i)))
	}

	config := ast.AnalyzerConfig{
		ProjectRoot:                   tempDir,
		EnableDependency:              true,
		EnableComponentMap:            true,
		MaxConcurrency:               4,
		MaxFileSize:                  10 * 1024 * 1024, // 10MB
		EnablePerformanceOptimization: true,
	}
	analyzer, err := ast.NewAnalyzer(config)
	if err != nil {
		b.Fatalf("Failed to create analyzer: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Parse files concurrently using repository analysis
		ctx := context.Background()
		result, err := analyzer.AnalyzeRepository(ctx)
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
		if result.Summary.TotalFiles != 20 {
			b.Fatalf("Expected 20 files, got %d", result.Summary.TotalFiles)
		}
	}
}

// Helper functions for benchmark test setup

func createSmallBenchmarkRepo(b *testing.B, baseDir string, fileCount int) string {
	repoDir := filepath.Join(baseDir, fmt.Sprintf("benchmark-repo-%d", fileCount))
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		b.Fatalf("Failed to create repo dir: %v", err)
	}

	// Create package.json
	packageJSON := `{
		"name": "benchmark-repo",
		"version": "1.0.0",
		"dependencies": {
			"lodash": "^4.17.21",
			"react": "^18.2.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(repoDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		b.Fatalf("Failed to create package.json: %v", err)
	}

	// Create directories
	dirs := []string{"src", "components", "utils", "services"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(repoDir, dir), 0755); err != nil {
			b.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	// Distribute files across directories
	filesPerDir := fileCount / len(dirs)
	remainder := fileCount % len(dirs)

	for i, dir := range dirs {
		filesToCreate := filesPerDir
		if i < remainder {
			filesToCreate++
		}

		for j := 0; j < filesToCreate; j++ {
			fileName := fmt.Sprintf("file%d.js", j)
			filePath := filepath.Join(repoDir, dir, fileName)

			content := generateBenchmarkFileContent(dir, j)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				b.Fatalf("Failed to create file %s: %v", filePath, err)
			}
		}
	}

	return repoDir
}

func createBenchmarkJSFile(b *testing.B, baseDir string) string {
	filePath := filepath.Join(baseDir, "benchmark.js")
	content := `
		const lodash = require('lodash');
		const utils = require('./utils');

		class BenchmarkClass {
			constructor() {
				this.initialized = false;
				this.data = [];
			}

			initialize() {
				this.initialized = true;
				this.data = [1, 2, 3, 4, 5];
			}

			processData() {
				return this.data.map(item => item * 2);
			}

			async fetchData(url) {
				const response = await fetch(url);
				return response.json();
			}
		}

		function benchmarkFunction(param) {
			return param.toString().toUpperCase();
		}

		function complexCalculation(numbers) {
			return numbers.reduce((acc, num) => {
				if (num % 2 === 0) {
					return acc + num;
				}
				return acc - num;
			}, 0);
		}

		const arrow = (x, y) => {
			return x + y;
		};

		module.exports = {
			BenchmarkClass,
			benchmarkFunction,
			complexCalculation,
			arrow
		};
	`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create benchmark JS file: %v", err)
	}

	return filePath
}

func createBenchmarkTSFile(b *testing.B, baseDir string) string {
	filePath := filepath.Join(baseDir, "benchmark.ts")
	content := `
		import { Observable } from 'rxjs';

		interface BenchmarkInterface {
			id: number;
			name: string;
			data: any[];
		}

		class BenchmarkService implements BenchmarkInterface {
			id: number;
			name: string;
			data: any[];
			private readonly apiUrl: string;

			constructor(apiUrl: string) {
				this.id = Math.random();
				this.name = 'BenchmarkService';
				this.data = [];
				this.apiUrl = apiUrl;
			}

			public async getData<T>(endpoint: string): Promise<T[]> {
				const response = await fetch(` + "`${this.apiUrl}/${endpoint}`" + `);
				return response.json();
			}

			public processItems<T>(items: T[], predicate: (item: T) => boolean): T[] {
				return items.filter(predicate);
			}

			protected validateData(data: any): boolean {
				return data !== null && data !== undefined;
			}
		}

		function genericFunction<T>(param: T): T {
			return param;
		}

		function complexGenericFunction<T extends Record<string, any>>(
			obj: T, 
			key: keyof T
		): T[keyof T] {
			return obj[key];
		}

		const arrowWithTypes = <T>(items: T[]): number => {
			return items.length;
		};

		export {
			BenchmarkInterface,
			BenchmarkService,
			genericFunction,
			complexGenericFunction,
			arrowWithTypes
		};
	`

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create benchmark TS file: %v", err)
	}

	return filePath
}

func createLargeJSFile(b *testing.B, baseDir string, functionCount int) string {
	filePath := filepath.Join(baseDir, "large.js")
	
	content := "// Large JavaScript file for benchmarking\n\n"
	
	// Generate many functions
	for i := 0; i < functionCount; i++ {
		content += fmt.Sprintf(`
			function function%d(param) {
				const result = param * %d;
				return result + Math.random();
			}
		`, i, i+1)
	}

	// Add a class with many methods
	content += `
		class LargeClass {
			constructor() {
				this.value = 0;
			}
	`

	for i := 0; i < functionCount/10; i++ {
		content += fmt.Sprintf(`
			method%d() {
				return this.value + %d;
			}
		`, i, i)
	}

	content += `
		}

		module.exports = {
			LargeClass,
	`

	// Export all functions
	for i := 0; i < functionCount; i++ {
		if i == functionCount-1 {
			content += fmt.Sprintf("			function%d", i)
		} else {
			content += fmt.Sprintf("			function%d,\n", i)
		}
	}

	content += "\n		};"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create large JS file: %v", err)
	}

	return filePath
}

func generateBenchmarkFileContent(dir string, index int) string {
	switch dir {
	case "components":
		return fmt.Sprintf(`
			const React = require('react');

			class Component%d extends React.Component {
				constructor(props) {
					super(props);
					this.state = { value: %d };
				}

				handleClick() {
					this.setState({ value: this.state.value + 1 });
				}

				render() {
					return React.createElement('div', {
						onClick: this.handleClick.bind(this)
					}, this.state.value);
				}
			}

			function FunctionalComponent%d(props) {
				const [count, setCount] = React.useState(0);
				
				return React.createElement('div', null, 
					React.createElement('button', {
						onClick: () => setCount(count + 1)
					}, 'Count: ' + count)
				);
			}

			module.exports = { Component%d, FunctionalComponent%d };
		`, index, index, index, index, index)

	case "services":
		return fmt.Sprintf(`
			class Service%d {
				constructor(baseUrl) {
					this.baseUrl = baseUrl || '/api';
					this.cache = new Map();
				}

				async get(id) {
					if (this.cache.has(id)) {
						return this.cache.get(id);
					}

					const response = await fetch(` + "`${this.baseUrl}/items/${id}`" + `);
					const data = await response.json();
					this.cache.set(id, data);
					return data;
				}

				async create(data) {
					const response = await fetch(` + "`${this.baseUrl}/items`" + `, {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(data)
					});
					return response.json();
				}

				clearCache() {
					this.cache.clear();
				}
			}

			module.exports = Service%d;
		`, index, index)

	case "utils":
		return fmt.Sprintf(`
			function util%d(input) {
				if (typeof input === 'string') {
					return input.toLowerCase().trim();
				}
				return String(input);
			}

			function calculate%d(a, b, operation) {
				switch (operation) {
					case 'add': return a + b;
					case 'subtract': return a - b;
					case 'multiply': return a * b;
					case 'divide': return b !== 0 ? a / b : 0;
					default: return 0;
				}
			}

			const constants%d = {
				MAX_VALUE: %d,
				MIN_VALUE: 0,
				DEFAULT_TIMEOUT: 5000
			};

			module.exports = {
				util%d,
				calculate%d,
				constants%d
			};
		`, index, index, index, index*100, index, index, index)

	default: // src
		return fmt.Sprintf(`
			const lodash = require('lodash');

			class Module%d {
				constructor() {
					this.id = %d;
					this.timestamp = Date.now();
				}

				process(data) {
					return lodash.map(data, item => ({
						...item,
						processed: true,
						moduleId: this.id
					}));
				}

				validate(input) {
					return input && typeof input === 'object' && input.id;
				}
			}

			function transform%d(data) {
				return data.map(item => ({
					id: item.id,
					value: item.value * 2,
					computed: Date.now()
				}));
			}

			const config%d = {
				enabled: true,
				maxItems: %d,
				timeout: 1000
			};

			module.exports = {
				Module%d,
				transform%d,
				config%d
			};
		`, index, index, index, index, index*10, index, index, index)
	}
}