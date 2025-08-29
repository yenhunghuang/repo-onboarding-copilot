package analysis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportFormatType defines the available export formats
type ExportFormatType string

const (
	FormatJSON     ExportFormatType = "json"
	FormatMarkdown ExportFormatType = "markdown"
	FormatPDF      ExportFormatType = "pdf"
	FormatHTML     ExportFormatType = "html"
)

// ExportOptions configures export behavior
type ExportOptions struct {
	Formats         []ExportFormatType     `json:"formats"`
	OutputDirectory string                 `json:"output_directory"`
	BaseFilename    string                 `json:"base_filename"`
	IncludeSections map[string]bool        `json:"include_sections"`
	Metadata        map[string]interface{} `json:"metadata"`
	Template        string                 `json:"template,omitempty"`
}

// ExportResult contains the result of an export operation
type ExportResult struct {
	Format           ExportFormatType `json:"format"`
	FilePath         string           `json:"file_path"`
	FileSize         int64            `json:"file_size"`
	Success          bool             `json:"success"`
	Error            string           `json:"error,omitempty"`
	GeneratedAt      time.Time        `json:"generated_at"`
	ProcessingTimeMs int64            `json:"processing_time_ms"`
}

// MultiFormatExporter handles exporting dependency analysis results to multiple formats
type MultiFormatExporter struct {
	defaultOptions ExportOptions
	templates      map[ExportFormatType]*template.Template
}

// NewMultiFormatExporter creates a new multi-format exporter
func NewMultiFormatExporter() *MultiFormatExporter {
	exporter := &MultiFormatExporter{
		defaultOptions: ExportOptions{
			Formats:         []ExportFormatType{FormatJSON, FormatMarkdown},
			OutputDirectory: "./reports",
			BaseFilename:    "dependency_analysis",
			IncludeSections: map[string]bool{
				"summary":         true,
				"dependencies":    true,
				"vulnerabilities": true,
				"performance":     true,
				"recommendations": true,
				"metadata":        true,
			},
		},
		templates: make(map[ExportFormatType]*template.Template),
	}

	// Initialize templates
	exporter.initializeTemplates()
	return exporter
}

// ExportData represents the structured data to be exported
type ExportData struct {
	Summary         ExportSummary          `json:"summary"`
	Dependencies    ExportDependencies     `json:"dependencies"`
	Vulnerabilities ExportVulnerabilities  `json:"vulnerabilities"`
	Performance     ExportPerformance      `json:"performance"`
	Recommendations []ExportRecommendation `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// ExportSummary provides high-level analysis summary
type ExportSummary struct {
	ProjectName        string             `json:"project_name"`
	TotalDependencies  int                `json:"total_dependencies"`
	DirectDependencies int                `json:"direct_dependencies"`
	CriticalVulns      int                `json:"critical_vulnerabilities"`
	HighVulns          int                `json:"high_vulnerabilities"`
	OutdatedPackages   int                `json:"outdated_packages"`
	LicenseIssues      int                `json:"license_issues"`
	OverallRiskScore   float64            `json:"overall_risk_score"`
	QualityGrade       string             `json:"quality_grade"`
	Metrics            map[string]float64 `json:"metrics"`
}

// ExportDependencies contains dependency tree information
type ExportDependencies struct {
	Direct     []ExportPackage      `json:"direct"`
	Transitive []ExportPackage      `json:"transitive"`
	Tree       ExportDependencyTree `json:"tree"`
	Statistics ExportDepStats       `json:"statistics"`
}

// ExportPackage represents a single package in exports
type ExportPackage struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	LatestVersion   string                 `json:"latest_version,omitempty"`
	License         string                 `json:"license"`
	Size            int64                  `json:"size"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	Vulnerabilities []ExportVulnerability  `json:"vulnerabilities,omitempty"`
	UpdateStatus    string                 `json:"update_status"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ExportDependencyTree represents the dependency tree structure
type ExportDependencyTree struct {
	Nodes []ExportTreeNode `json:"nodes"`
	Edges []ExportTreeEdge `json:"edges"`
	Depth int              `json:"max_depth"`
}

// ExportTreeNode represents a node in the dependency tree
type ExportTreeNode struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Version  string                 `json:"version"`
	Type     string                 `json:"type"` // direct, transitive
	Level    int                    `json:"level"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ExportTreeEdge represents an edge in the dependency tree
type ExportTreeEdge struct {
	From         string `json:"from"`
	To           string `json:"to"`
	Relationship string `json:"relationship"` // depends, devDepends, peerDepends
}

// ExportDepStats provides dependency statistics
type ExportDepStats struct {
	TotalSize         int64  `json:"total_size"`
	AverageSize       int64  `json:"average_size"`
	LargestPackage    string `json:"largest_package"`
	MostDependencies  string `json:"most_dependencies"`
	CircularDeps      int    `json:"circular_dependencies"`
	DuplicatePackages int    `json:"duplicate_packages"`
}

// ExportVulnerabilities contains vulnerability information
type ExportVulnerabilities struct {
	Summary    ExportVulnSummary     `json:"summary"`
	ByPackage  []ExportPackageVulns  `json:"by_package"`
	BySeverity map[string]int        `json:"by_severity"`
	TopThreat  []ExportVulnerability `json:"top_threats"`
}

// ExportVulnSummary provides vulnerability summary
type ExportVulnSummary struct {
	Total        int       `json:"total"`
	Critical     int       `json:"critical"`
	High         int       `json:"high"`
	Medium       int       `json:"medium"`
	Low          int       `json:"low"`
	Fixed        int       `json:"fixed"`
	RiskScore    float64   `json:"risk_score"`
	LastScanDate time.Time `json:"last_scan_date"`
}

// ExportPackageVulns represents vulnerabilities for a specific package
type ExportPackageVulns struct {
	PackageName     string                `json:"package_name"`
	Version         string                `json:"version"`
	Vulnerabilities []ExportVulnerability `json:"vulnerabilities"`
	RiskLevel       string                `json:"risk_level"`
}

// ExportVulnerability represents a single vulnerability
type ExportVulnerability struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	CVSSScore   float64   `json:"cvss_score"`
	CVE         string    `json:"cve,omitempty"`
	URL         string    `json:"url"`
	FixVersion  string    `json:"fix_version,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}

// ExportPerformance contains performance analysis information
type ExportPerformance struct {
	BundleAnalysis   ExportBundleAnalysis       `json:"bundle_analysis"`
	LoadTimes        ExportLoadTimes            `json:"load_times"`
	Recommendations  []ExportPerfRecommendation `json:"recommendations"`
	PerformanceScore float64                    `json:"performance_score"`
}

// ExportBundleAnalysis represents bundle size analysis
type ExportBundleAnalysis struct {
	TotalSize       int64                 `json:"total_size"`
	GzippedSize     int64                 `json:"gzipped_size"`
	ByCategory      map[string]int64      `json:"by_category"`
	LargestPackages []ExportPackageSize   `json:"largest_packages"`
	TreeShaking     ExportTreeShakingInfo `json:"tree_shaking"`
}

// ExportPackageSize represents package size information
type ExportPackageSize struct {
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Percentage   float64 `json:"percentage"`
	TreeShakable bool    `json:"tree_shakable"`
}

// ExportTreeShakingInfo provides tree-shaking analysis
type ExportTreeShakingInfo struct {
	Potential     int64    `json:"potential_savings"`
	Percentage    float64  `json:"percentage_reduction"`
	Opportunities []string `json:"opportunities"`
}

// ExportLoadTimes represents load time analysis
type ExportLoadTimes struct {
	Networks map[string]ExportNetworkTiming `json:"networks"`
	Devices  map[string]ExportDeviceTiming  `json:"devices"`
}

// ExportNetworkTiming represents timing for different networks
type ExportNetworkTiming struct {
	Name       string `json:"name"`
	Bandwidth  string `json:"bandwidth"`
	LoadTime   int64  `json:"load_time_ms"`
	Grade      string `json:"grade"`
	Acceptable bool   `json:"acceptable"`
}

// ExportDeviceTiming represents timing for different devices
type ExportDeviceTiming struct {
	Name          string  `json:"name"`
	CPUMultiplier float64 `json:"cpu_multiplier"`
	ParseTime     int64   `json:"parse_time_ms"`
	ExecuteTime   int64   `json:"execute_time_ms"`
	TotalTime     int64   `json:"total_time_ms"`
}

// ExportPerfRecommendation represents a performance recommendation
type ExportPerfRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
	Savings     int64  `json:"potential_savings_bytes,omitempty"`
}

// ExportRecommendation represents a general recommendation
type ExportRecommendation struct {
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	Category    string    `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`
	Steps       []string  `json:"steps"`
	Resources   []string  `json:"resources,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExportAll exports data to all configured formats
func (mfe *MultiFormatExporter) ExportAll(payload *DependencyOrchestrationPayload, options *ExportOptions) ([]ExportResult, error) {
	if options == nil {
		options = &mfe.defaultOptions
	}

	// Convert payload to export data
	exportData, err := mfe.convertPayloadToExportData(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert payload to export data: %w", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(options.OutputDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	var results []ExportResult
	for _, format := range options.Formats {
		result := mfe.exportToFormat(exportData, format, options)
		results = append(results, result)
	}

	return results, nil
}

// convertPayloadToExportData converts orchestration payload to export format
func (mfe *MultiFormatExporter) convertPayloadToExportData(payload *DependencyOrchestrationPayload) (*ExportData, error) {
	data := &ExportData{
		GeneratedAt: time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Convert summary information
	if payload.DependencyTree != nil {
		projectName := "Unknown Project"
		totalDeps := 0
		if payload.DependencyTree.RootPackage != nil {
			projectName = payload.DependencyTree.RootPackage.Name
		}
		if payload.DependencyTree.AllDependencies != nil {
			totalDeps = len(payload.DependencyTree.AllDependencies)
		}

		data.Summary = ExportSummary{
			ProjectName:        projectName,
			TotalDependencies:  totalDeps,
			DirectDependencies: mfe.countDirectDeps(payload.DependencyTree),
			QualityGrade:       "B+", // Default grade
			Metrics:            make(map[string]float64),
		}
	}

	// Convert dependencies
	data.Dependencies = mfe.convertDependencies(payload.DependencyTree)

	// Convert vulnerabilities
	data.Vulnerabilities = mfe.convertVulnerabilities(payload.SecurityReport)

	// Convert performance data
	data.Performance = mfe.convertPerformance(payload.PerformanceReport)

	// Convert recommendations
	data.Recommendations = mfe.convertRecommendations(payload)

	// Add metadata
	data.Metadata["export_version"] = "1.0"
	data.Metadata["analyzer_version"] = "2.2.0"

	return data, nil
}

// Helper methods for data conversion
func (mfe *MultiFormatExporter) countDirectDeps(tree *DependencyTree) int {
	if tree == nil || tree.RootPackage == nil {
		return 0
	}

	directCount := 0
	if tree.RootPackage.Dependencies != nil {
		directCount += len(tree.RootPackage.Dependencies)
	}
	if tree.RootPackage.DevDependencies != nil {
		directCount += len(tree.RootPackage.DevDependencies)
	}
	return directCount
}

func (mfe *MultiFormatExporter) convertDependencies(tree *DependencyTree) ExportDependencies {
	deps := ExportDependencies{
		Direct:     []ExportPackage{},
		Transitive: []ExportPackage{},
		Tree:       ExportDependencyTree{Nodes: []ExportTreeNode{}, Edges: []ExportTreeEdge{}},
		Statistics: ExportDepStats{},
	}

	if tree == nil {
		return deps
	}

	// Convert direct dependencies
	if tree.DirectDeps != nil {
		for _, depNode := range tree.DirectDeps {
			exportPkg := ExportPackage{
				Name:         depNode.Name,
				Version:      depNode.Version,
				License:      depNode.License.Name,
				Size:         depNode.PackageInfo.EstimatedSize,
				Dependencies: mfe.getDependencyNames(depNode.Children),
				UpdateStatus: "current", // Default status
				Metadata:     make(map[string]interface{}),
			}

			// Convert vulnerabilities if present
			for _, vuln := range depNode.Vulnerabilities {
				exportVuln := ExportVulnerability{
					ID:          vuln.ID,
					Title:       vuln.Title,
					Description: vuln.Description,
					Severity:    vuln.Severity,
					CVSSScore:   vuln.CVSS,
					CVE:         "", // CVE field not in our Vulnerability struct
					URL:         "",
					FixVersion:  "",
					PublishedAt: time.Now(), // Default timestamp
				}

				if len(vuln.References) > 0 {
					exportVuln.URL = vuln.References[0]
				}
				if len(vuln.PatchedIn) > 0 {
					exportVuln.FixVersion = vuln.PatchedIn[0]
				}

				exportPkg.Vulnerabilities = append(exportPkg.Vulnerabilities, exportVuln)
			}

			deps.Direct = append(deps.Direct, exportPkg)
		}
	}

	// Convert all dependencies (includes transitive)
	if tree.AllDependencies != nil {
		for _, depNode := range tree.AllDependencies {
			if !depNode.IsTransitive {
				continue // Skip direct deps, already processed
			}

			exportPkg := ExportPackage{
				Name:         depNode.Name,
				Version:      depNode.Version,
				License:      depNode.License.Name,
				Size:         depNode.PackageInfo.EstimatedSize,
				Dependencies: mfe.getDependencyNames(depNode.Children),
				UpdateStatus: "current", // Default status
				Metadata:     make(map[string]interface{}),
			}

			deps.Transitive = append(deps.Transitive, exportPkg)
		}
	}

	// Add statistics if available
	if tree.Statistics.TotalDependencies > 0 {
		deps.Statistics = ExportDepStats{
			TotalSize:         tree.Statistics.TotalSize,
			AverageSize:       tree.Statistics.TotalSize / int64(tree.Statistics.TotalDependencies),
			DuplicatePackages: tree.Statistics.DuplicatePackages,
		}
	}

	return deps
}

// getDependencyNames extracts dependency names from children map
func (mfe *MultiFormatExporter) getDependencyNames(children map[string]*DependencyNode) []string {
	if children == nil {
		return []string{}
	}

	names := make([]string, 0, len(children))
	for name := range children {
		names = append(names, name)
	}
	return names
}

func (mfe *MultiFormatExporter) convertVulnerabilities(report *SecurityReport) ExportVulnerabilities {
	vulns := ExportVulnerabilities{
		Summary:    ExportVulnSummary{LastScanDate: time.Now()},
		ByPackage:  []ExportPackageVulns{},
		BySeverity: make(map[string]int),
		TopThreat:  []ExportVulnerability{},
	}

	if report == nil {
		return vulns
	}

	// Convert vulnerability summary
	vulns.Summary.Total = report.TotalVulnerabilities
	vulns.Summary.Critical = report.CriticalCount
	vulns.Summary.High = report.HighCount
	vulns.Summary.Medium = report.MediumCount
	vulns.Summary.Low = report.LowCount
	vulns.Summary.RiskScore = report.RiskScore
	vulns.BySeverity = report.SeverityDistribution

	for _, vuln := range report.Vulnerabilities {
		exportVuln := ExportVulnerability{
			ID:          vuln.ID,
			Title:       vuln.Title,
			Description: vuln.Description,
			Severity:    vuln.Severity,
			CVSSScore:   vuln.CVSS,
			CVE:         "", // Not available in our Vulnerability struct
			URL:         "",
			FixVersion:  "",
			PublishedAt: time.Now(), // Default timestamp
		}

		if len(vuln.References) > 0 {
			exportVuln.URL = vuln.References[0]
		}
		if len(vuln.PatchedIn) > 0 {
			exportVuln.FixVersion = vuln.PatchedIn[0]
		}

		if vuln.Severity == "critical" || vuln.Severity == "high" {
			vulns.TopThreat = append(vulns.TopThreat, exportVuln)
		}
	}

	return vulns
}

func (mfe *MultiFormatExporter) convertPerformance(report *PerformanceReport) ExportPerformance {
	perf := ExportPerformance{
		BundleAnalysis:   ExportBundleAnalysis{},
		LoadTimes:        ExportLoadTimes{Networks: make(map[string]ExportNetworkTiming), Devices: make(map[string]ExportDeviceTiming)},
		Recommendations:  []ExportPerfRecommendation{},
		PerformanceScore: 85.0, // Default score
	}

	if report == nil {
		return perf
	}

	// Convert performance data from available fields
	if len(report.AverageLoadTime) > 0 {
		for network, loadTime := range report.AverageLoadTime {
			perf.LoadTimes.Networks[network] = ExportNetworkTiming{
				Name:       network,
				LoadTime:   int64(loadTime),
				Acceptable: loadTime < 3000, // 3 second threshold
				Bandwidth:  network,         // Use network type as bandwidth info
				Grade:      mfe.getLoadTimeGrade(loadTime),
			}
		}
	}

	// Use total impact as performance score
	perf.PerformanceScore = report.TotalImpact

	return perf
}

// getLoadTimeGrade assigns a grade based on load time
func (mfe *MultiFormatExporter) getLoadTimeGrade(loadTime float64) string {
	switch {
	case loadTime < 1000:
		return "A"
	case loadTime < 2000:
		return "B"
	case loadTime < 3000:
		return "C"
	case loadTime < 5000:
		return "D"
	default:
		return "F"
	}
}

func (mfe *MultiFormatExporter) convertRecommendations(payload *DependencyOrchestrationPayload) []ExportRecommendation {
	var recommendations []ExportRecommendation

	// Add default recommendations based on analysis
	if payload.SecurityReport != nil && len(payload.SecurityReport.Vulnerabilities) > 0 {
		recommendations = append(recommendations, ExportRecommendation{
			Type:        "security",
			Priority:    "high",
			Category:    "vulnerability",
			Title:       "Address Security Vulnerabilities",
			Description: fmt.Sprintf("Found %d vulnerabilities that should be addressed", len(payload.SecurityReport.Vulnerabilities)),
			Impact:      "high",
			Effort:      "medium",
			Steps:       []string{"Review vulnerability report", "Update affected packages", "Test changes"},
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}

// exportToFormat exports data to a specific format
func (mfe *MultiFormatExporter) exportToFormat(data *ExportData, format ExportFormatType, options *ExportOptions) ExportResult {
	startTime := time.Now()
	result := ExportResult{
		Format:      format,
		GeneratedAt: startTime,
		Success:     false,
	}

	filename := fmt.Sprintf("%s.%s", options.BaseFilename, string(format))
	filepath := filepath.Join(options.OutputDirectory, filename)
	result.FilePath = filepath

	var err error
	switch format {
	case FormatJSON:
		err = mfe.exportJSON(data, filepath)
	case FormatMarkdown:
		err = mfe.exportMarkdown(data, filepath)
	case FormatHTML:
		err = mfe.exportHTML(data, filepath)
	case FormatPDF:
		err = mfe.exportPDF(data, filepath)
	default:
		err = fmt.Errorf("unsupported format: %s", format)
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Get file size
	if fileInfo, statErr := os.Stat(filepath); statErr == nil {
		result.FileSize = fileInfo.Size()
	}

	result.Success = true
	return result
}

// exportJSON exports data as JSON
func (mfe *MultiFormatExporter) exportJSON(data *ExportData, filepath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filepath, jsonData, 0644)
}

// exportMarkdown exports data as Markdown
func (mfe *MultiFormatExporter) exportMarkdown(data *ExportData, filepath string) error {
	var buf bytes.Buffer

	// Write markdown content
	buf.WriteString(fmt.Sprintf("# Dependency Analysis Report\n\n"))
	buf.WriteString(fmt.Sprintf("**Generated:** %s\n\n", data.GeneratedAt.Format("January 2, 2006 at 3:04 PM")))

	// Summary section
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- **Project:** %s\n", data.Summary.ProjectName))
	buf.WriteString(fmt.Sprintf("- **Total Dependencies:** %d\n", data.Summary.TotalDependencies))
	buf.WriteString(fmt.Sprintf("- **Direct Dependencies:** %d\n", data.Summary.DirectDependencies))
	buf.WriteString(fmt.Sprintf("- **Overall Risk Score:** %.1f\n\n", data.Summary.OverallRiskScore))

	// Vulnerabilities section
	if data.Vulnerabilities.Summary.Total > 0 {
		buf.WriteString("## Security Vulnerabilities\n\n")
		buf.WriteString(fmt.Sprintf("- **Total Vulnerabilities:** %d\n", data.Vulnerabilities.Summary.Total))
		buf.WriteString(fmt.Sprintf("- **Critical:** %d\n", data.Vulnerabilities.Summary.Critical))
		buf.WriteString(fmt.Sprintf("- **High:** %d\n", data.Vulnerabilities.Summary.High))
		buf.WriteString(fmt.Sprintf("- **Medium:** %d\n", data.Vulnerabilities.Summary.Medium))
		buf.WriteString(fmt.Sprintf("- **Low:** %d\n\n", data.Vulnerabilities.Summary.Low))
	}

	// Performance section
	buf.WriteString("## Performance Analysis\n\n")
	buf.WriteString(fmt.Sprintf("- **Bundle Size:** %d bytes\n", data.Performance.BundleAnalysis.TotalSize))
	buf.WriteString(fmt.Sprintf("- **Gzipped Size:** %d bytes\n", data.Performance.BundleAnalysis.GzippedSize))
	buf.WriteString(fmt.Sprintf("- **Performance Score:** %.1f/100\n\n", data.Performance.PerformanceScore))

	// Recommendations section
	if len(data.Recommendations) > 0 {
		buf.WriteString("## Recommendations\n\n")
		for i, rec := range data.Recommendations {
			buf.WriteString(fmt.Sprintf("%d. **%s** (%s priority)\n", i+1, rec.Title, rec.Priority))
			buf.WriteString(fmt.Sprintf("   %s\n\n", rec.Description))
		}
	}

	return os.WriteFile(filepath, buf.Bytes(), 0644)
}

// exportHTML exports data as HTML
func (mfe *MultiFormatExporter) exportHTML(data *ExportData, filepath string) error {
	tmpl := mfe.templates[FormatHTML]
	if tmpl == nil {
		return fmt.Errorf("HTML template not initialized")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return os.WriteFile(filepath, buf.Bytes(), 0644)
}

// exportPDF exports data as PDF (placeholder implementation)
func (mfe *MultiFormatExporter) exportPDF(data *ExportData, filepath string) error {
	// For now, we'll create a simple text-based PDF representation
	// In a real implementation, you would use a PDF library like gofpdf

	htmlPath := strings.Replace(filepath, ".pdf", ".html", 1)
	if err := mfe.exportHTML(data, htmlPath); err != nil {
		return fmt.Errorf("failed to create HTML for PDF conversion: %w", err)
	}

	// Create a simple PDF placeholder file
	pdfContent := fmt.Sprintf("PDF Export - Dependency Analysis Report\nGenerated: %s\nProject: %s\n\nThis is a placeholder PDF export.\nFor full PDF support, integrate with a PDF generation library.",
		data.GeneratedAt.Format("2006-01-02 15:04:05"),
		data.Summary.ProjectName)

	return os.WriteFile(filepath, []byte(pdfContent), 0644)
}

// initializeTemplates sets up templates for different formats
func (mfe *MultiFormatExporter) initializeTemplates() {
	// HTML template
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Dependency Analysis Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 10px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px 20px 10px 0; }
        .vulnerability { background: #fff3cd; padding: 10px; margin: 5px 0; border-left: 4px solid #ffc107; }
        .critical { border-left-color: #dc3545; background: #f8d7da; }
        .high { border-left-color: #fd7e14; background: #ffeaa7; }
        .recommendation { background: #d1ecf1; padding: 10px; margin: 5px 0; border-left: 4px solid #17a2b8; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Dependency Analysis Report</h1>
        <p><strong>Generated:</strong> {{.GeneratedAt.Format "January 2, 2006 at 3:04 PM"}}</p>
        <p><strong>Project:</strong> {{.Summary.ProjectName}}</p>
    </div>

    <div class="section">
        <h2>Summary</h2>
        <div class="metric"><strong>Total Dependencies:</strong> {{.Summary.TotalDependencies}}</div>
        <div class="metric"><strong>Direct Dependencies:</strong> {{.Summary.DirectDependencies}}</div>
        <div class="metric"><strong>Risk Score:</strong> {{printf "%.1f" .Summary.OverallRiskScore}}</div>
        <div class="metric"><strong>Quality Grade:</strong> {{.Summary.QualityGrade}}</div>
    </div>

    {{if gt .Vulnerabilities.Summary.Total 0}}
    <div class="section">
        <h2>Security Vulnerabilities</h2>
        <div class="metric"><strong>Total:</strong> {{.Vulnerabilities.Summary.Total}}</div>
        <div class="metric"><strong>Critical:</strong> {{.Vulnerabilities.Summary.Critical}}</div>
        <div class="metric"><strong>High:</strong> {{.Vulnerabilities.Summary.High}}</div>
        <div class="metric"><strong>Medium:</strong> {{.Vulnerabilities.Summary.Medium}}</div>
        <div class="metric"><strong>Low:</strong> {{.Vulnerabilities.Summary.Low}}</div>
    </div>
    {{end}}

    <div class="section">
        <h2>Performance Analysis</h2>
        <div class="metric"><strong>Bundle Size:</strong> {{.Performance.BundleAnalysis.TotalSize}} bytes</div>
        <div class="metric"><strong>Gzipped Size:</strong> {{.Performance.BundleAnalysis.GzippedSize}} bytes</div>
        <div class="metric"><strong>Performance Score:</strong> {{printf "%.1f" .Performance.PerformanceScore}}/100</div>
    </div>

    {{if .Recommendations}}
    <div class="section">
        <h2>Recommendations</h2>
        {{range .Recommendations}}
        <div class="recommendation">
            <strong>{{.Title}}</strong> ({{.Priority}} priority)
            <p>{{.Description}}</p>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err == nil {
		mfe.templates[FormatHTML] = tmpl
	}
}

// SetTemplate allows custom template setting
func (mfe *MultiFormatExporter) SetTemplate(format ExportFormatType, templateContent string) error {
	tmpl, err := template.New(string(format)).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	mfe.templates[format] = tmpl
	return nil
}

// LoadTemplateFromFile loads template from file
func (mfe *MultiFormatExporter) LoadTemplateFromFile(format ExportFormatType, filepath string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	return mfe.SetTemplate(format, string(content))
}

// GetSupportedFormats returns list of supported export formats
func (mfe *MultiFormatExporter) GetSupportedFormats() []ExportFormatType {
	return []ExportFormatType{FormatJSON, FormatMarkdown, FormatHTML, FormatPDF}
}

// ValidateOptions validates export options
func (mfe *MultiFormatExporter) ValidateOptions(options *ExportOptions) error {
	if options == nil {
		return fmt.Errorf("export options cannot be nil")
	}

	if len(options.Formats) == 0 {
		return fmt.Errorf("at least one export format must be specified")
	}

	supported := mfe.GetSupportedFormats()
	supportedMap := make(map[ExportFormatType]bool)
	for _, format := range supported {
		supportedMap[format] = true
	}

	for _, format := range options.Formats {
		if !supportedMap[format] {
			return fmt.Errorf("unsupported format: %s", format)
		}
	}

	if options.OutputDirectory == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	if options.BaseFilename == "" {
		return fmt.Errorf("base filename cannot be empty")
	}

	return nil
}
