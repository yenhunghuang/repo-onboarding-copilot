package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewMultiFormatExporter(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	if exporter == nil {
		t.Fatal("NewMultiFormatExporter returned nil")
	}
	
	if len(exporter.defaultOptions.Formats) == 0 {
		t.Error("Default formats should not be empty")
	}
	
	if exporter.defaultOptions.OutputDirectory == "" {
		t.Error("Default output directory should not be empty")
	}
	
	if exporter.defaultOptions.BaseFilename == "" {
		t.Error("Default base filename should not be empty")
	}
	
	if exporter.templates == nil {
		t.Error("Templates map should be initialized")
	}
}

func TestExportAll(t *testing.T) {
	// Create test payload
	payload := createTestPayload()
	
	// Create temporary directory for test outputs
	tempDir := t.TempDir()
	
	options := &ExportOptions{
		Formats:         []ExportFormatType{FormatJSON, FormatMarkdown},
		OutputDirectory: tempDir,
		BaseFilename:    "test_report",
		IncludeSections: map[string]bool{
			"summary":         true,
			"dependencies":    true,
			"vulnerabilities": true,
		},
	}
	
	exporter := NewMultiFormatExporter()
	results, err := exporter.ExportAll(payload, options)
	
	if err != nil {
		t.Fatalf("ExportAll failed: %v", err)
	}
	
	if len(results) != len(options.Formats) {
		t.Errorf("Expected %d results, got %d", len(options.Formats), len(results))
	}
	
	// Check that files were created
	for _, result := range results {
		if !result.Success {
			t.Errorf("Export failed for format %s: %s", result.Format, result.Error)
		}
		
		if result.FilePath == "" {
			t.Errorf("File path is empty for format %s", result.Format)
		}
		
		// Verify file exists
		if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", result.FilePath)
		}
		
		// Verify file size
		if result.FileSize <= 0 {
			t.Errorf("File size should be greater than 0 for format %s", result.Format)
		}
		
		// Verify processing time is recorded
		if result.ProcessingTimeMs < 0 {
			t.Errorf("Processing time should be non-negative for format %s", result.Format)
		}
	}
}

func TestExportJSON(t *testing.T) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	
	// Convert to export data
	exportData, err := exporter.convertPayloadToExportData(payload)
	if err != nil {
		t.Fatalf("Failed to convert payload: %v", err)
	}
	
	tempDir := t.TempDir()
	filepath := filepath.Join(tempDir, "test.json")
	
	err = exporter.exportJSON(exportData, filepath)
	if err != nil {
		t.Fatalf("JSON export failed: %v", err)
	}
	
	// Verify file exists and is valid JSON
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}
	
	// Verify required fields
	requiredFields := []string{"summary", "dependencies", "vulnerabilities", "performance", "generated_at"}
	for _, field := range requiredFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Required field %s missing from JSON output", field)
		}
	}
}

func TestExportMarkdown(t *testing.T) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	
	exportData, err := exporter.convertPayloadToExportData(payload)
	if err != nil {
		t.Fatalf("Failed to convert payload: %v", err)
	}
	
	tempDir := t.TempDir()
	filepath := filepath.Join(tempDir, "test.md")
	
	err = exporter.exportMarkdown(exportData, filepath)
	if err != nil {
		t.Fatalf("Markdown export failed: %v", err)
	}
	
	// Verify file exists and contains expected content
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read Markdown file: %v", err)
	}
	
	contentStr := string(content)
	
	// Check for expected markdown elements
	expectedElements := []string{
		"# Dependency Analysis Report",
		"## Summary",
		"**Project:**",
		"**Total Dependencies:**",
		"**Generated:**",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("Expected markdown element '%s' not found in output", element)
		}
	}
}

func TestExportHTML(t *testing.T) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	
	exportData, err := exporter.convertPayloadToExportData(payload)
	if err != nil {
		t.Fatalf("Failed to convert payload: %v", err)
	}
	
	tempDir := t.TempDir()
	filepath := filepath.Join(tempDir, "test.html")
	
	err = exporter.exportHTML(exportData, filepath)
	if err != nil {
		t.Fatalf("HTML export failed: %v", err)
	}
	
	// Verify file exists and contains expected HTML content
	content, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}
	
	contentStr := string(content)
	
	// Check for expected HTML elements
	expectedElements := []string{
		"<!DOCTYPE html>",
		"<title>Dependency Analysis Report</title>",
		"<h1>Dependency Analysis Report</h1>",
		"<h2>Summary</h2>",
		"</html>",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("Expected HTML element '%s' not found in output", element)
		}
	}
}

func TestConvertPayloadToExportData(t *testing.T) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	
	exportData, err := exporter.convertPayloadToExportData(payload)
	if err != nil {
		t.Fatalf("Failed to convert payload: %v", err)
	}
	
	// Verify export data structure
	expectedProjectName := payload.DependencyTree.RootPackage.Name
	if exportData.Summary.ProjectName != expectedProjectName {
		t.Errorf("Project name mismatch: expected %s, got %s", 
			expectedProjectName, exportData.Summary.ProjectName)
	}
	
	expectedTotalDeps := len(payload.DependencyTree.AllDependencies)
	if exportData.Summary.TotalDependencies != expectedTotalDeps {
		t.Errorf("Total dependencies mismatch: expected %d, got %d", 
			expectedTotalDeps, exportData.Summary.TotalDependencies)
	}
	
	if len(exportData.Dependencies.Direct) == 0 && len(exportData.Dependencies.Transitive) == 0 {
		t.Error("Should have either direct or transitive dependencies")
	}
	
	if exportData.GeneratedAt.IsZero() {
		t.Error("Generated timestamp should not be zero")
	}
	
	if len(exportData.Metadata) == 0 {
		t.Error("Metadata should not be empty")
	}
}

func TestConvertDependencies(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	// Test with nil tree
	deps := exporter.convertDependencies(nil)
	if len(deps.Direct) != 0 || len(deps.Transitive) != 0 {
		t.Error("Should handle nil dependency tree gracefully")
	}
	
	// Test with valid tree
	tree := &DependencyTree{
		RootPackage: &PackageManifest{
			Name:    "test-project",
			Version: "1.0.0",
		},
		DirectDeps: map[string]*DependencyNode{
			"react": {
				Name:    "react",
				Version: "18.2.0",
				License: LicenseInfo{Name: "MIT"},
				PackageInfo: &PackageInfo{EstimatedSize: 123456},
				Children: map[string]*DependencyNode{
					"prop-types": {
						Name:        "prop-types",
						Version:     "15.8.1",
						License:     LicenseInfo{Name: "MIT"},
						PackageInfo: &PackageInfo{EstimatedSize: 45678},
						IsTransitive: true,
					},
				},
			},
		},
		AllDependencies: map[string]*DependencyNode{
			"react": {
				Name:    "react",
				Version: "18.2.0",
				License: LicenseInfo{Name: "MIT"},
				PackageInfo: &PackageInfo{EstimatedSize: 123456},
				IsTransitive: false,
			},
			"prop-types": {
				Name:        "prop-types",
				Version:     "15.8.1",
				License:     LicenseInfo{Name: "MIT"},
				PackageInfo: &PackageInfo{EstimatedSize: 45678},
				IsTransitive: true,
			},
		},
	}
	
	deps = exporter.convertDependencies(tree)
	
	if len(deps.Direct) != 1 {
		t.Errorf("Expected 1 direct dependency, got %d", len(deps.Direct))
	}
	
	if len(deps.Transitive) != 1 {
		t.Errorf("Expected 1 transitive dependency, got %d", len(deps.Transitive))
	}
	
	// Verify direct dependency details
	if deps.Direct[0].Name != "react" {
		t.Errorf("Expected direct dependency name 'react', got '%s'", deps.Direct[0].Name)
	}
	
	if deps.Direct[0].Version != "18.2.0" {
		t.Errorf("Expected version '18.2.0', got '%s'", deps.Direct[0].Version)
	}
}

func TestConvertVulnerabilities(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	// Test with nil report
	vulns := exporter.convertVulnerabilities(nil)
	if vulns.Summary.Total != 0 {
		t.Error("Should handle nil vulnerability report gracefully")
	}
	
	// Test with valid report
	report := &SecurityReport{
		TotalVulnerabilities: 2,
		CriticalCount:       1,
		HighCount:          1,
		SeverityDistribution: map[string]int{"high": 1, "critical": 1},
		Vulnerabilities: []Vulnerability{
			{
				ID:          "CVE-2023-1234",
				Title:       "Test Vulnerability",
				Description: "A test vulnerability",
				Severity:    "high",
				CVSS:        7.5,
				References:  []string{"https://example.com/cve-2023-1234"},
				PatchedIn:   []string{"1.2.3"},
			},
			{
				ID:          "CVE-2023-5678",
				Title:       "Critical Vulnerability",
				Description: "A critical vulnerability",
				Severity:    "critical",
				CVSS:        9.0,
			},
		},
	}
	
	vulns = exporter.convertVulnerabilities(report)
	
	if vulns.Summary.Total != 2 {
		t.Errorf("Expected 2 total vulnerabilities, got %d", vulns.Summary.Total)
	}
	
	if vulns.BySeverity["high"] != 1 {
		t.Errorf("Expected 1 high severity vulnerability, got %d", vulns.BySeverity["high"])
	}
	
	if vulns.BySeverity["critical"] != 1 {
		t.Errorf("Expected 1 critical severity vulnerability, got %d", vulns.BySeverity["critical"])
	}
	
	if len(vulns.TopThreat) != 2 {
		t.Errorf("Expected 2 top threats (high and critical), got %d", len(vulns.TopThreat))
	}
}

func TestConvertPerformance(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	// Test with nil report
	perf := exporter.convertPerformance(nil)
	if perf.PerformanceScore != 85.0 {
		t.Error("Should set default performance score")
	}
	
	// Test with valid report
	report := &PerformanceReport{
		AverageLoadTime: map[string]float64{
			"3G": 2500,
			"4G": 800,
		},
		TotalImpact: 75.0,
	}
	
	perf = exporter.convertPerformance(report)
	
	if perf.PerformanceScore != 75.0 {
		t.Errorf("Expected performance score 75.0, got %.1f", perf.PerformanceScore)
	}
	
	if len(perf.LoadTimes.Networks) != 2 {
		t.Errorf("Expected 2 network timings, got %d", len(perf.LoadTimes.Networks))
	}
	
	// Verify 3G timing
	if timing, exists := perf.LoadTimes.Networks["3G"]; exists {
		if timing.LoadTime != 2500 {
			t.Errorf("Expected 3G load time 2500ms, got %d", timing.LoadTime)
		}
		if timing.Acceptable {
			t.Error("3G load time should not be acceptable (>3000ms threshold)")
		}
	} else {
		t.Error("3G timing not found")
	}
	
	// Verify 4G timing  
	if timing, exists := perf.LoadTimes.Networks["4G"]; exists {
		if timing.LoadTime != 800 {
			t.Errorf("Expected 4G load time 800ms, got %d", timing.LoadTime)
		}
		if !timing.Acceptable {
			t.Error("4G load time should be acceptable (<3000ms threshold)")
		}
	} else {
		t.Error("4G timing not found")
	}
}

func TestValidateOptions(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	// Test nil options
	err := exporter.ValidateOptions(nil)
	if err == nil {
		t.Error("Should reject nil options")
	}
	
	// Test empty formats
	options := &ExportOptions{
		Formats: []ExportFormatType{},
	}
	err = exporter.ValidateOptions(options)
	if err == nil {
		t.Error("Should reject empty formats")
	}
	
	// Test unsupported format
	options = &ExportOptions{
		Formats: []ExportFormatType{"xml"},
	}
	err = exporter.ValidateOptions(options)
	if err == nil {
		t.Error("Should reject unsupported format")
	}
	
	// Test empty output directory
	options = &ExportOptions{
		Formats:         []ExportFormatType{FormatJSON},
		OutputDirectory: "",
	}
	err = exporter.ValidateOptions(options)
	if err == nil {
		t.Error("Should reject empty output directory")
	}
	
	// Test empty base filename
	options = &ExportOptions{
		Formats:         []ExportFormatType{FormatJSON},
		OutputDirectory: "/tmp",
		BaseFilename:    "",
	}
	err = exporter.ValidateOptions(options)
	if err == nil {
		t.Error("Should reject empty base filename")
	}
	
	// Test valid options
	options = &ExportOptions{
		Formats:         []ExportFormatType{FormatJSON, FormatMarkdown},
		OutputDirectory: "/tmp",
		BaseFilename:    "test",
	}
	err = exporter.ValidateOptions(options)
	if err != nil {
		t.Errorf("Should accept valid options: %v", err)
	}
}

func TestGetSupportedFormats(t *testing.T) {
	exporter := NewMultiFormatExporter()
	formats := exporter.GetSupportedFormats()
	
	if len(formats) == 0 {
		t.Error("Should return at least one supported format")
	}
	
	expectedFormats := []ExportFormatType{FormatJSON, FormatMarkdown, FormatHTML, FormatPDF}
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}
	
	formatMap := make(map[ExportFormatType]bool)
	for _, format := range formats {
		formatMap[format] = true
	}
	
	for _, expected := range expectedFormats {
		if !formatMap[expected] {
			t.Errorf("Expected format %s not found in supported formats", expected)
		}
	}
}

func TestSetTemplate(t *testing.T) {
	exporter := NewMultiFormatExporter()
	
	// Test valid template
	template := `<html><body>{{.Summary.ProjectName}}</body></html>`
	err := exporter.SetTemplate(FormatHTML, template)
	if err != nil {
		t.Errorf("Should accept valid template: %v", err)
	}
	
	// Test invalid template
	invalidTemplate := `<html><body>{{.InvalidField</body></html>`
	err = exporter.SetTemplate(FormatHTML, invalidTemplate)
	if err == nil {
		t.Error("Should reject invalid template")
	}
}

// Helper function to create test payload
func createTestPayload() *DependencyOrchestrationPayload {
	return &DependencyOrchestrationPayload{
		DependencyTree: &DependencyTree{
			RootPackage: &PackageManifest{
				Name:    "test-project",
				Version: "1.0.0",
				Dependencies: map[string]string{
					"react":  "^18.2.0",
					"lodash": "^4.17.21",
				},
				DevDependencies: map[string]string{
					"jest": "^29.0.0",
				},
			},
			DirectDeps: map[string]*DependencyNode{
				"react": {
					Name:    "react",
					Version: "18.2.0",
					License: LicenseInfo{Name: "MIT"},
					PackageInfo: &PackageInfo{EstimatedSize: 123456},
					Children: map[string]*DependencyNode{
						"prop-types": {
							Name:        "prop-types",
							Version:     "15.8.1",
							License:     LicenseInfo{Name: "MIT"},
							PackageInfo: &PackageInfo{EstimatedSize: 45678},
							IsTransitive: true,
						},
					},
				},
				"lodash": {
					Name:        "lodash",
					Version:     "4.17.21",
					License:     LicenseInfo{Name: "MIT"},
					PackageInfo: &PackageInfo{EstimatedSize: 67890},
				},
			},
			AllDependencies: map[string]*DependencyNode{
				"react": {
					Name:        "react",
					Version:     "18.2.0",
					License:     LicenseInfo{Name: "MIT"},
					PackageInfo: &PackageInfo{EstimatedSize: 123456},
					IsTransitive: false,
				},
				"lodash": {
					Name:        "lodash",
					Version:     "4.17.21",
					License:     LicenseInfo{Name: "MIT"},
					PackageInfo: &PackageInfo{EstimatedSize: 67890},
					IsTransitive: false,
				},
				"prop-types": {
					Name:        "prop-types",
					Version:     "15.8.1",
					License:     LicenseInfo{Name: "MIT"},
					PackageInfo: &PackageInfo{EstimatedSize: 45678},
					IsTransitive: true,
				},
			},
		},
		SecurityReport: &SecurityReport{
			TotalVulnerabilities: 1,
			CriticalCount:       0,
			HighCount:          1,
			MediumCount:        0,
			LowCount:           0,
			RiskScore:          75.0,
			SeverityDistribution: map[string]int{"high": 1},
			Vulnerabilities: []Vulnerability{
				{
					ID:          "CVE-2023-1234",
					Title:       "Test Vulnerability",
					Description: "A test vulnerability description",
					Severity:    "high",
					CVSS:        7.5,
					References:  []string{"https://example.com/cve-2023-1234"},
					PatchedIn:   []string{"1.2.3"},
				},
			},
		},
		PerformanceReport: &PerformanceReport{
			AverageLoadTime: map[string]float64{
				"3G":   2500,
				"WiFi": 300,
			},
			TotalImpact: 85.0,
		},
	}
}

// Benchmark tests
func BenchmarkExportJSON(b *testing.B) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	exportData, _ := exporter.convertPayloadToExportData(payload)
	
	tempDir := b.TempDir()
	filepath := filepath.Join(tempDir, "benchmark.json")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exporter.exportJSON(exportData, filepath)
	}
}

func BenchmarkExportMarkdown(b *testing.B) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	exportData, _ := exporter.convertPayloadToExportData(payload)
	
	tempDir := b.TempDir()
	filepath := filepath.Join(tempDir, "benchmark.md")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exporter.exportMarkdown(exportData, filepath)
	}
}

func BenchmarkConvertPayload(b *testing.B) {
	exporter := NewMultiFormatExporter()
	payload := createTestPayload()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = exporter.convertPayloadToExportData(payload)
	}
}