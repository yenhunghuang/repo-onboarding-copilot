package metrics

import (
	"context"
	"crypto/md5"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// DuplicationDetector identifies code duplication across JavaScript/TypeScript codebases
type DuplicationDetector struct {
	config DuplicationConfig
}

// DuplicationConfig defines thresholds and settings for duplication detection
type DuplicationConfig struct {
	MinLines           int     `yaml:"min_lines" json:"min_lines"`
	MinTokens          int     `yaml:"min_tokens" json:"min_tokens"`
	SimilarityThreshold float64 `yaml:"similarity_threshold" json:"similarity_threshold"`
	TokenSimilarityThreshold float64 `yaml:"token_similarity_threshold" json:"token_similarity_threshold"`
	MaxDistance        int     `yaml:"max_distance" json:"max_distance"`
	IgnoreWhitespace   bool    `yaml:"ignore_whitespace" json:"ignore_whitespace"`
	IgnoreComments     bool    `yaml:"ignore_comments" json:"ignore_comments"`
	IgnoreVariableNames bool   `yaml:"ignore_variable_names" json:"ignore_variable_names"`
	EnableCrossFile    bool    `yaml:"enable_cross_file" json:"enable_cross_file"`
	ReportTopN         int     `yaml:"report_top_n" json:"report_top_n"`
	WeightFactors      DuplicationWeights `yaml:"weight_factors" json:"weight_factors"`
}

// DuplicationWeights for different types of duplication impact
type DuplicationWeights struct {
	ExactDuplication      float64 `yaml:"exact_duplication" json:"exact_duplication"`
	StructuralSimilarity  float64 `yaml:"structural_similarity" json:"structural_similarity"`
	TokenSimilarity       float64 `yaml:"token_similarity" json:"token_similarity"`
	CrossFileImpact       float64 `yaml:"cross_file_impact" json:"cross_file_impact"`
	MaintenanceBurden     float64 `yaml:"maintenance_burden" json:"maintenance_burden"`
}

// DuplicationMetrics contains comprehensive duplication analysis results
type DuplicationMetrics struct {
	OverallScore         float64                    `json:"overall_score"`
	TotalDuplicatedLines int                        `json:"total_duplicated_lines"`
	DuplicationRatio     float64                    `json:"duplication_ratio"`
	ExactDuplicates      []DuplicationCluster       `json:"exact_duplicates"`
	StructuralDuplicates []DuplicationCluster       `json:"structural_duplicates"`
	TokenDuplicates      []DuplicationCluster       `json:"token_duplicates"`
	CrossFileDuplicates  []CrossFileDuplication     `json:"cross_file_duplicates"`
	DuplicationByFile    map[string]FileDuplication `json:"duplication_by_file"`
	ConsolidationOps     []ConsolidationOpportunity `json:"consolidation_opportunities"`
	ImpactAnalysis       DuplicationImpact          `json:"impact_analysis"`
	Recommendations      []DuplicationRecommendation `json:"recommendations"`
	Summary              DuplicationSummary          `json:"summary"`
}

// DuplicationCluster represents a group of similar code blocks
type DuplicationCluster struct {
	ID               string             `json:"id"`
	Type             string             `json:"type"` // exact, structural, token
	Instances        []DuplicationInstance `json:"instances"`
	SimilarityScore  float64            `json:"similarity_score"`
	LineCount        int                `json:"line_count"`
	TokenCount       int                `json:"token_count"`
	MaintenanceBurden float64           `json:"maintenance_burden"`
	RefactoringEffort string            `json:"refactoring_effort"`
	Priority         string             `json:"priority"`
	Recommendations  []string           `json:"recommendations"`
}

// DuplicationInstance represents a single occurrence of duplicated code
type DuplicationInstance struct {
	FilePath     string                 `json:"file_path"`
	StartLine    int                    `json:"start_line"`
	EndLine      int                    `json:"end_line"`
	FunctionName string                 `json:"function_name,omitempty"`
	ClassName    string                 `json:"class_name,omitempty"`
	Content      string                 `json:"content"`
	TokenizedContent string             `json:"tokenized_content"`
	StructuralHash string               `json:"structural_hash"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CrossFileDuplication tracks duplication across different files
type CrossFileDuplication struct {
	ClusterID       string                `json:"cluster_id"`
	FilePairs       []CrossFileInstance   `json:"file_pairs"`
	SharedFunctionality string            `json:"shared_functionality"`
	ConsolidationTarget string            `json:"consolidation_target"`
	EstimatedSavings    int               `json:"estimated_savings"`
	RefactoringStrategy string            `json:"refactoring_strategy"`
}

// CrossFileInstance represents duplication between specific files
type CrossFileInstance struct {
	File1        string  `json:"file1"`
	File2        string  `json:"file2"`
	Similarity   float64 `json:"similarity"`
	SharedLines  int     `json:"shared_lines"`
	SharedTokens int     `json:"shared_tokens"`
}

// FileDuplication aggregates duplication metrics at the file level
type FileDuplication struct {
	FilePath            string  `json:"file_path"`
	InternalDuplication int     `json:"internal_duplication"`
	ExternalDuplication int     `json:"external_duplication"`
	DuplicationRatio    float64 `json:"duplication_ratio"`
	HotspotScore        float64 `json:"hotspot_score"`
	RefactoringPriority string  `json:"refactoring_priority"`
}

// ConsolidationOpportunity identifies specific refactoring opportunities
type ConsolidationOpportunity struct {
	ID                  string   `json:"id"`
	Type                string   `json:"type"` // extract_function, create_utility, merge_classes
	Description         string   `json:"description"`
	AffectedFiles       []string `json:"affected_files"`
	AffectedFunctions   []string `json:"affected_functions"`
	EstimatedReduction  int      `json:"estimated_reduction"`
	ComplexityReduction float64  `json:"complexity_reduction"`
	MaintenanceImprovement string `json:"maintenance_improvement"`
	RefactoringSteps    []string `json:"refactoring_steps"`
	EstimatedEffort     int      `json:"estimated_effort_hours"`
	ROIScore            float64  `json:"roi_score"`
}

// DuplicationImpact analyzes the maintenance and technical debt impact
type DuplicationImpact struct {
	MaintenanceMultiplier float64                    `json:"maintenance_multiplier"`
	TechnicalDebtScore    float64                    `json:"technical_debt_score"`
	ChangeRiskFactor      float64                    `json:"change_risk_factor"`
	TestingBurden         int                        `json:"testing_burden"`
	CodebaseHealth        string                     `json:"codebase_health"`
	HotspotAnalysis       []DuplicationHotspot       `json:"hotspot_analysis"`
}

// DuplicationHotspot identifies areas with high duplication density
type DuplicationHotspot struct {
	Location          string  `json:"location"`
	DuplicationScore  float64 `json:"duplication_score"`
	AffectedFunctions int     `json:"affected_functions"`
	MaintenanceRisk   string  `json:"maintenance_risk"`
	RecommendedAction string  `json:"recommended_action"`
}

// DuplicationRecommendation provides actionable improvement suggestions
type DuplicationRecommendation struct {
	Priority        string   `json:"priority"`
	Category        string   `json:"category"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Impact          string   `json:"impact"`
	Effort          string   `json:"effort"`
	Clusters        []string `json:"clusters"`
	Techniques      []string `json:"techniques"`
	EstimatedHours  int      `json:"estimated_hours"`
	ExpectedReduction int    `json:"expected_reduction"`
}

// DuplicationSummary provides executive-level overview
type DuplicationSummary struct {
	HealthScore        float64 `json:"health_score"`
	RiskLevel          string  `json:"risk_level"`
	MaintenanceBurden  string  `json:"maintenance_burden"`
	RefactoringNeeded  int     `json:"refactoring_needed"`
	PotentialSavings   int     `json:"potential_savings"`
	RecommendedActions int     `json:"recommended_actions"`
}

// NewDuplicationDetector creates a new duplication detector with default configuration
func NewDuplicationDetector() *DuplicationDetector {
	return &DuplicationDetector{
		config: DuplicationConfig{
			MinLines:                 6,
			MinTokens:                50,
			SimilarityThreshold:      0.85,
			TokenSimilarityThreshold: 0.75,
			MaxDistance:              3,
			IgnoreWhitespace:         true,
			IgnoreComments:           true,
			IgnoreVariableNames:      false,
			EnableCrossFile:          true,
			ReportTopN:               15,
			WeightFactors: DuplicationWeights{
				ExactDuplication:     1.0,
				StructuralSimilarity: 0.8,
				TokenSimilarity:      0.6,
				CrossFileImpact:      1.2,
				MaintenanceBurden:    0.9,
			},
		},
	}
}

// NewDuplicationDetectorWithConfig creates detector with custom configuration
func NewDuplicationDetectorWithConfig(config DuplicationConfig) *DuplicationDetector {
	return &DuplicationDetector{
		config: config,
	}
}

// DetectDuplication performs comprehensive duplication analysis on parsed AST results
func (dd *DuplicationDetector) DetectDuplication(ctx context.Context, parseResults []*ast.ParseResult) (*DuplicationMetrics, error) {
	if len(parseResults) == 0 {
		return nil, fmt.Errorf("no parse results provided for duplication analysis")
	}

	metrics := &DuplicationMetrics{
		ExactDuplicates:      []DuplicationCluster{},
		StructuralDuplicates: []DuplicationCluster{},
		TokenDuplicates:      []DuplicationCluster{},
		CrossFileDuplicates:  []CrossFileDuplication{},
		DuplicationByFile:    make(map[string]FileDuplication),
		ConsolidationOps:     []ConsolidationOpportunity{},
		Recommendations:      []DuplicationRecommendation{},
	}

	// Extract code blocks for analysis
	codeBlocks, err := dd.extractCodeBlocks(parseResults)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code blocks: %w", err)
	}

	// Detect exact duplicates
	exactDuplicates := dd.findExactDuplicates(codeBlocks)
	metrics.ExactDuplicates = dd.clusterDuplicates(exactDuplicates, "exact")

	// Detect structural duplicates
	structuralDuplicates := dd.findStructuralDuplicates(codeBlocks)
	metrics.StructuralDuplicates = dd.clusterDuplicates(structuralDuplicates, "structural")

	// Detect token-based duplicates
	tokenDuplicates := dd.findTokenDuplicates(codeBlocks)
	metrics.TokenDuplicates = dd.clusterDuplicates(tokenDuplicates, "token")

	// Analyze cross-file duplication if enabled
	if dd.config.EnableCrossFile {
		metrics.CrossFileDuplicates = dd.analyzeCrossFileDuplication(parseResults, metrics)
	}

	// Calculate file-level metrics
	dd.calculateFileMetrics(parseResults, metrics)

	// Generate consolidation opportunities
	metrics.ConsolidationOps = dd.generateConsolidationOpportunities(metrics)

	// Perform impact analysis
	metrics.ImpactAnalysis = dd.analyzeImpact(metrics)

	// Calculate aggregate metrics
	dd.calculateAggregateMetrics(metrics)

	// Generate recommendations
	dd.generateRecommendations(metrics)

	// Generate summary
	dd.generateSummary(metrics)

	return metrics, nil
}

// extractCodeBlocks extracts analyzable code blocks from parsed results
func (dd *DuplicationDetector) extractCodeBlocks(parseResults []*ast.ParseResult) ([]DuplicationInstance, error) {
	blocks := []DuplicationInstance{}
	blockID := 0

	for _, parseResult := range parseResults {
		// Extract function blocks
		for _, function := range parseResult.Functions {
			if dd.isBlockSizeValid(function.StartLine, function.EndLine) {
				block := DuplicationInstance{
					FilePath:     parseResult.FilePath,
					StartLine:    function.StartLine,
					EndLine:      function.EndLine,
					FunctionName: function.Name,
					Content:      dd.generateBlockContent(function.StartLine, function.EndLine),
					Metadata: map[string]interface{}{
						"id":          blockID,
						"type":        "function",
						"async":       function.IsAsync,
						"exported":    function.IsExported,
						"param_count": len(function.Parameters),
					},
				}
				
				// Generate tokenized and structural representations
				block.TokenizedContent = dd.tokenizeContent(block.Content)
				block.StructuralHash = dd.generateStructuralHash(block.Content)
				
				blocks = append(blocks, block)
				blockID++
			}
		}

		// Extract class method blocks
		for _, class := range parseResult.Classes {
			for _, method := range class.Methods {
				if dd.isBlockSizeValid(method.StartLine, method.EndLine) {
					block := DuplicationInstance{
						FilePath:     parseResult.FilePath,
						StartLine:    method.StartLine,
						EndLine:      method.EndLine,
						FunctionName: method.Name,
						ClassName:    class.Name,
						Content:      dd.generateBlockContent(method.StartLine, method.EndLine),
						Metadata: map[string]interface{}{
							"id":          blockID,
							"type":        "method",
							"class":       class.Name,
							"async":       method.IsAsync,
							"param_count": len(method.Parameters),
						},
					}
					
					block.TokenizedContent = dd.tokenizeContent(block.Content)
					block.StructuralHash = dd.generateStructuralHash(block.Content)
					
					blocks = append(blocks, block)
					blockID++
				}
			}
		}
	}

	return blocks, nil
}

// isBlockSizeValid checks if a code block meets minimum size requirements
func (dd *DuplicationDetector) isBlockSizeValid(startLine, endLine int) bool {
	lineCount := endLine - startLine + 1
	return lineCount >= dd.config.MinLines
}

// generateBlockContent creates content representation for a code block
func (dd *DuplicationDetector) generateBlockContent(startLine, endLine int) string {
	// In a real implementation, this would read the actual file content
	// For this implementation, we'll generate representative content
	lines := make([]string, 0, endLine-startLine+1)
	for i := startLine; i <= endLine; i++ {
		lines = append(lines, fmt.Sprintf("  // Line %d content", i))
	}
	return strings.Join(lines, "\n")
}

// tokenizeContent creates a normalized token representation
func (dd *DuplicationDetector) tokenizeContent(content string) string {
	tokenized := content
	
	// Remove comments first if configured
	if dd.config.IgnoreComments {
		// Remove single-line comments
		singleLineCommentRegex := regexp.MustCompile(`//.*`)
		tokenized = singleLineCommentRegex.ReplaceAllString(tokenized, "")
		
		// Remove multi-line comments
		multiLineCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
		tokenized = multiLineCommentRegex.ReplaceAllString(tokenized, "")
	}
	
	// Normalize whitespace if configured
	if dd.config.IgnoreWhitespace {
		whitespaceRegex := regexp.MustCompile(`\s+`)
		tokenized = whitespaceRegex.ReplaceAllString(tokenized, " ")
		tokenized = strings.TrimSpace(tokenized)
	}
	
	// Replace identifiers with generic tokens if configured
	if dd.config.IgnoreVariableNames {
		// Replace identifiers with generic tokens
		identifierRegex := regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`)
		tokenized = identifierRegex.ReplaceAllString(tokenized, "IDENTIFIER")
		
		// Replace string literals
		stringRegex := regexp.MustCompile(`["'][^"']*["']`)
		tokenized = stringRegex.ReplaceAllString(tokenized, "STRING_LITERAL")
		
		// Replace numeric literals
		numberRegex := regexp.MustCompile(`\b\d+(\.\d+)?\b`)
		tokenized = numberRegex.ReplaceAllString(tokenized, "NUMBER_LITERAL")
	}
	
	return tokenized
}

// generateStructuralHash creates a hash based on code structure
func (dd *DuplicationDetector) generateStructuralHash(content string) string {
	// Extract structural elements (keywords, operators, brackets)
	structuralRegex := regexp.MustCompile(`\b(if|else|for|while|function|class|return|var|let|const)\b|[{}\[\]();,.]`)
	matches := structuralRegex.FindAllString(content, -1)
	structural := strings.Join(matches, "")
	
	// Generate MD5 hash
	hash := md5.Sum([]byte(structural))
	return fmt.Sprintf("%x", hash)
}

// findExactDuplicates identifies exact code duplicates
func (dd *DuplicationDetector) findExactDuplicates(blocks []DuplicationInstance) [][]DuplicationInstance {
	clusters := [][]DuplicationInstance{}
	duplicateMap := make(map[string][]DuplicationInstance)
	
	// Group blocks by exact content
	for _, block := range blocks {
		content := block.Content
		if dd.config.IgnoreWhitespace {
			content = strings.ReplaceAll(strings.ReplaceAll(content, " ", ""), "\t", "")
		}
		duplicateMap[content] = append(duplicateMap[content], block)
	}
	
	// Extract clusters with multiple instances
	for _, instances := range duplicateMap {
		if len(instances) > 1 {
			clusters = append(clusters, instances)
		}
	}
	
	return clusters
}

// findStructuralDuplicates identifies structurally similar code
func (dd *DuplicationDetector) findStructuralDuplicates(blocks []DuplicationInstance) [][]DuplicationInstance {
	clusters := [][]DuplicationInstance{}
	hashGroups := make(map[string][]DuplicationInstance)
	
	// Group blocks by structural hash
	for _, block := range blocks {
		hashGroups[block.StructuralHash] = append(hashGroups[block.StructuralHash], block)
	}
	
	// Extract clusters with multiple instances and high similarity
	for _, instances := range hashGroups {
		if len(instances) > 1 {
			// Verify similarity threshold
			validCluster := dd.validateStructuralCluster(instances)
			if validCluster {
				clusters = append(clusters, instances)
			}
		}
	}
	
	return clusters
}

// findTokenDuplicates identifies token-based similar code
func (dd *DuplicationDetector) findTokenDuplicates(blocks []DuplicationInstance) [][]DuplicationInstance {
	clusters := [][]DuplicationInstance{}
	
	// Compare all pairs of blocks for token similarity
	for i := 0; i < len(blocks); i++ {
		cluster := []DuplicationInstance{blocks[i]}
		
		for j := i + 1; j < len(blocks); j++ {
			similarity := dd.calculateTokenSimilarity(blocks[i].TokenizedContent, blocks[j].TokenizedContent)
			if similarity >= dd.config.TokenSimilarityThreshold {
				cluster = append(cluster, blocks[j])
			}
		}
		
		if len(cluster) > 1 {
			// Check if this cluster overlaps with existing ones
			if !dd.clusterExists(clusters, cluster) {
				clusters = append(clusters, cluster)
			}
		}
	}
	
	return clusters
}

// validateStructuralCluster verifies that a structural cluster meets similarity requirements
func (dd *DuplicationDetector) validateStructuralCluster(instances []DuplicationInstance) bool {
	if len(instances) < 2 {
		return false
	}
	
	// Calculate average similarity within cluster
	totalSimilarity := 0.0
	comparisons := 0
	
	for i := 0; i < len(instances); i++ {
		for j := i + 1; j < len(instances); j++ {
			similarity := dd.calculateContentSimilarity(instances[i].Content, instances[j].Content)
			totalSimilarity += similarity
			comparisons++
		}
	}
	
	avgSimilarity := totalSimilarity / float64(comparisons)
	return avgSimilarity >= dd.config.SimilarityThreshold
}

// calculateTokenSimilarity computes similarity between tokenized content
func (dd *DuplicationDetector) calculateTokenSimilarity(content1, content2 string) float64 {
	tokens1 := strings.Fields(content1)
	tokens2 := strings.Fields(content2)
	
	if len(tokens1) == 0 && len(tokens2) == 0 {
		return 1.0
	}
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}
	
	// Calculate Jaccard similarity
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, token := range tokens1 {
		set1[token] = true
	}
	for _, token := range tokens2 {
		set2[token] = true
	}
	
	intersection := 0
	union := len(set1)
	
	for token := range set2 {
		if set1[token] {
			intersection++
		} else {
			union++
		}
	}
	
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}

// calculateContentSimilarity computes similarity between raw content
func (dd *DuplicationDetector) calculateContentSimilarity(content1, content2 string) float64 {
	if content1 == content2 {
		return 1.0
	}
	
	// Use Levenshtein distance for similarity calculation
	distance := dd.levenshteinDistance(content1, content2)
	maxLen := math.Max(float64(len(content1)), float64(len(content2)))
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - (float64(distance) / maxLen)
}

// levenshteinDistance calculates edit distance between two strings
func (dd *DuplicationDetector) levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	matrix := make([][]int, len1+1)
	
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
		matrix[i][0] = i
	}
	
	for j := 1; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			matrix[i][j] = dd.min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len1][len2]
}

// min3 returns the minimum of three integers
func (dd *DuplicationDetector) min3(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// clusterExists checks if a similar cluster already exists
func (dd *DuplicationDetector) clusterExists(clusters [][]DuplicationInstance, newCluster []DuplicationInstance) bool {
	for _, existingCluster := range clusters {
		if dd.clustersOverlap(existingCluster, newCluster) {
			return true
		}
	}
	return false
}

// clustersOverlap checks if two clusters have significant overlap
func (dd *DuplicationDetector) clustersOverlap(cluster1, cluster2 []DuplicationInstance) bool {
	overlap := 0
	for _, instance1 := range cluster1 {
		for _, instance2 := range cluster2 {
			if dd.instancesEqual(instance1, instance2) {
				overlap++
			}
		}
	}
	
	// Consider overlap significant if > 50% of smaller cluster overlaps
	minSize := len(cluster1)
	if len(cluster2) < minSize {
		minSize = len(cluster2)
	}
	
	return float64(overlap)/float64(minSize) > 0.5
}

// instancesEqual checks if two duplication instances are the same
func (dd *DuplicationDetector) instancesEqual(instance1, instance2 DuplicationInstance) bool {
	return instance1.FilePath == instance2.FilePath &&
		instance1.StartLine == instance2.StartLine &&
		instance1.EndLine == instance2.EndLine
}

// clusterDuplicates converts duplicate groups into structured clusters
func (dd *DuplicationDetector) clusterDuplicates(duplicateGroups [][]DuplicationInstance, clusterType string) []DuplicationCluster {
	clusters := []DuplicationCluster{}
	
	for i, group := range duplicateGroups {
		if len(group) < 2 {
			continue
		}
		
		cluster := DuplicationCluster{
			ID:               fmt.Sprintf("%s_%d", clusterType, i),
			Type:             clusterType,
			Instances:        group,
			LineCount:        group[0].EndLine - group[0].StartLine + 1,
			Recommendations:  []string{},
		}
		
		// Calculate similarity score
		cluster.SimilarityScore = dd.calculateClusterSimilarity(group)
		
		// Estimate token count
		cluster.TokenCount = dd.estimateTokenCount(group[0].Content)
		
		// Calculate maintenance burden
		cluster.MaintenanceBurden = dd.calculateMaintenanceBurden(cluster)
		
		// Assess refactoring effort
		cluster.RefactoringEffort = dd.assessRefactoringEffort(cluster)
		
		// Determine priority
		cluster.Priority = dd.determinePriority(cluster)
		
		// Generate recommendations
		cluster.Recommendations = dd.generateClusterRecommendations(cluster)
		
		clusters = append(clusters, cluster)
	}
	
	// Sort clusters by priority and impact
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].MaintenanceBurden > clusters[j].MaintenanceBurden
	})
	
	return clusters
}

// calculateClusterSimilarity computes average similarity within a cluster
func (dd *DuplicationDetector) calculateClusterSimilarity(instances []DuplicationInstance) float64 {
	if len(instances) < 2 {
		return 0.0
	}
	
	totalSimilarity := 0.0
	comparisons := 0
	
	for i := 0; i < len(instances); i++ {
		for j := i + 1; j < len(instances); j++ {
			similarity := dd.calculateContentSimilarity(instances[i].Content, instances[j].Content)
			totalSimilarity += similarity
			comparisons++
		}
	}
	
	return totalSimilarity / float64(comparisons)
}

// estimateTokenCount estimates the number of tokens in content
func (dd *DuplicationDetector) estimateTokenCount(content string) int {
	// Simple estimation based on words and symbols
	tokens := strings.Fields(content)
	symbolRegex := regexp.MustCompile(`[{}\[\]();,.]`)
	symbols := symbolRegex.FindAllString(content, -1)
	
	return len(tokens) + len(symbols)
}

// calculateMaintenanceBurden assesses the maintenance impact of duplication
func (dd *DuplicationDetector) calculateMaintenanceBurden(cluster DuplicationCluster) float64 {
	burden := float64(len(cluster.Instances)) * float64(cluster.LineCount)
	
	// Apply weight factors
	switch cluster.Type {
	case "exact":
		burden *= dd.config.WeightFactors.ExactDuplication
	case "structural":
		burden *= dd.config.WeightFactors.StructuralSimilarity
	case "token":
		burden *= dd.config.WeightFactors.TokenSimilarity
	}
	
	return burden * dd.config.WeightFactors.MaintenanceBurden
}

// assessRefactoringEffort evaluates the difficulty of refactoring
func (dd *DuplicationDetector) assessRefactoringEffort(cluster DuplicationCluster) string {
	crossFileCount := dd.countCrossFileInstances(cluster.Instances)
	
	if crossFileCount > len(cluster.Instances)/2 {
		return "high" // Cross-file refactoring is complex
	} else if cluster.LineCount > 50 {
		return "medium" // Large blocks require careful extraction
	} else if len(cluster.Instances) > 5 {
		return "medium" // Many instances increase coordination complexity
	}
	
	return "low"
}

// countCrossFileInstances counts instances across different files
func (dd *DuplicationDetector) countCrossFileInstances(instances []DuplicationInstance) int {
	files := make(map[string]bool)
	for _, instance := range instances {
		files[instance.FilePath] = true
	}
	return len(files)
}

// determinePriority assigns priority based on impact and effort
func (dd *DuplicationDetector) determinePriority(cluster DuplicationCluster) string {
	if cluster.MaintenanceBurden > 100 && cluster.RefactoringEffort == "low" {
		return "critical"
	} else if cluster.MaintenanceBurden > 50 && cluster.RefactoringEffort != "high" {
		return "high"
	} else if cluster.MaintenanceBurden > 20 {
		return "medium"
	}
	return "low"
}

// generateClusterRecommendations creates specific recommendations for a cluster
func (dd *DuplicationDetector) generateClusterRecommendations(cluster DuplicationCluster) []string {
	recommendations := []string{}
	
	if cluster.Type == "exact" {
		recommendations = append(recommendations, "Extract common functionality into a shared utility function")
		recommendations = append(recommendations, "Consider creating a base class or mixin for shared behavior")
	}
	
	if cluster.Type == "structural" {
		recommendations = append(recommendations, "Identify common patterns and create template functions")
		recommendations = append(recommendations, "Use design patterns like Strategy or Template Method")
	}
	
	if cluster.Type == "token" {
		recommendations = append(recommendations, "Standardize variable naming and code formatting")
		recommendations = append(recommendations, "Consider refactoring to use consistent abstractions")
	}
	
	if cluster.RefactoringEffort == "high" {
		recommendations = append(recommendations, "Plan refactoring in phases to minimize risk")
		recommendations = append(recommendations, "Ensure comprehensive test coverage before refactoring")
	}
	
	return recommendations
}

// analyzeCrossFileDuplication identifies duplication patterns across files
func (dd *DuplicationDetector) analyzeCrossFileDuplication(parseResults []*ast.ParseResult, metrics *DuplicationMetrics) []CrossFileDuplication {
	crossFileInstances := []CrossFileDuplication{}
	
	// Analyze each cluster for cross-file patterns
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	
	for _, cluster := range allClusters {
		crossFile := dd.analyzeCrossFileCluster(cluster)
		if crossFile != nil {
			crossFileInstances = append(crossFileInstances, *crossFile)
		}
	}
	
	return crossFileInstances
}

// analyzeCrossFileCluster analyzes a cluster for cross-file implications
func (dd *DuplicationDetector) analyzeCrossFileCluster(cluster DuplicationCluster) *CrossFileDuplication {
	fileGroups := make(map[string][]DuplicationInstance)
	
	// Group instances by file
	for _, instance := range cluster.Instances {
		fileGroups[instance.FilePath] = append(fileGroups[instance.FilePath], instance)
	}
	
	// Only consider clusters with cross-file duplication
	if len(fileGroups) < 2 {
		return nil
	}
	
	crossFile := &CrossFileDuplication{
		ClusterID:           cluster.ID,
		FilePairs:           []CrossFileInstance{},
		SharedFunctionality: dd.identifySharedFunctionality(cluster),
		EstimatedSavings:    cluster.LineCount * (len(cluster.Instances) - 1),
		RefactoringStrategy: dd.determineRefactoringStrategy(cluster),
	}
	
	// Analyze all file pairs
	files := make([]string, 0, len(fileGroups))
	for file := range fileGroups {
		files = append(files, file)
	}
	
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			pair := CrossFileInstance{
				File1:        files[i],
				File2:        files[j],
				Similarity:   cluster.SimilarityScore,
				SharedLines:  cluster.LineCount,
				SharedTokens: cluster.TokenCount,
			}
			crossFile.FilePairs = append(crossFile.FilePairs, pair)
		}
	}
	
	// Determine consolidation target
	crossFile.ConsolidationTarget = dd.determineConsolidationTarget(fileGroups)
	
	return crossFile
}

// identifySharedFunctionality analyzes what functionality is being duplicated
func (dd *DuplicationDetector) identifySharedFunctionality(cluster DuplicationCluster) string {
	// Analyze function names and content to identify common functionality
	functionTypes := make(map[string]int)
	
	for _, instance := range cluster.Instances {
		if instance.FunctionName != "" {
			// Extract common patterns from function names
			if strings.Contains(strings.ToLower(instance.FunctionName), "validate") {
				functionTypes["validation"]++
			} else if strings.Contains(strings.ToLower(instance.FunctionName), "format") {
				functionTypes["formatting"]++
			} else if strings.Contains(strings.ToLower(instance.FunctionName), "parse") {
				functionTypes["parsing"]++
			} else if strings.Contains(strings.ToLower(instance.FunctionName), "transform") {
				functionTypes["transformation"]++
			} else {
				functionTypes["utility"]++
			}
		}
	}
	
	// Return most common functionality type
	maxCount := 0
	mostCommon := "utility"
	for funcType, count := range functionTypes {
		if count > maxCount {
			maxCount = count
			mostCommon = funcType
		}
	}
	
	return mostCommon
}

// determineRefactoringStrategy suggests an appropriate refactoring approach
func (dd *DuplicationDetector) determineRefactoringStrategy(cluster DuplicationCluster) string {
	crossFileCount := dd.countCrossFileInstances(cluster.Instances)
	
	if cluster.Type == "exact" && crossFileCount > 1 {
		return "extract_to_shared_module"
	} else if cluster.Type == "structural" {
		return "create_template_function"
	} else if cluster.Type == "token" {
		return "standardize_and_refactor"
	} else if crossFileCount == 1 {
		return "extract_local_function"
	}
	
	return "manual_review_required"
}

// determineConsolidationTarget identifies the best file for consolidation
func (dd *DuplicationDetector) determineConsolidationTarget(fileGroups map[string][]DuplicationInstance) string {
	// Simple heuristic: choose file with most instances
	maxInstances := 0
	targetFile := ""
	
	// Get sorted file names for deterministic behavior
	var files []string
	for file := range fileGroups {
		files = append(files, file)
	}
	sort.Strings(files)
	
	for _, file := range files {
		instances := fileGroups[file]
		if len(instances) > maxInstances {
			maxInstances = len(instances)
			targetFile = file
		} else if len(instances) == maxInstances && targetFile == "" {
			// In case of tie, first file alphabetically wins (deterministic)
			targetFile = file
		}
	}
	
	// If no clear winner, suggest creating a new utility file
	if maxInstances <= 1 {
		return "new_utility_file"
	}
	
	return targetFile
}

// calculateFileMetrics computes duplication metrics for each file
func (dd *DuplicationDetector) calculateFileMetrics(parseResults []*ast.ParseResult, metrics *DuplicationMetrics) {
	for _, parseResult := range parseResults {
		filePath := parseResult.FilePath
		
		fileDuplication := FileDuplication{
			FilePath: filePath,
		}
		
		// Count internal and external duplications
		allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
		
		internalLines := 0
		externalLines := 0
		
		for _, cluster := range allClusters {
			fileInstances := 0
			otherFileInstances := 0
			
			for _, instance := range cluster.Instances {
				if instance.FilePath == filePath {
					fileInstances++
					internalLines += cluster.LineCount
				} else {
					otherFileInstances++
				}
			}
			
			if fileInstances > 0 && otherFileInstances > 0 {
				externalLines += cluster.LineCount * fileInstances
			}
		}
		
		fileDuplication.InternalDuplication = internalLines
		fileDuplication.ExternalDuplication = externalLines
		
		// Calculate total lines in file (estimated)
		totalLines := dd.estimateTotalLines(parseResult)
		if totalLines > 0 {
			fileDuplication.DuplicationRatio = float64(internalLines+externalLines) / float64(totalLines)
		}
		
		// Calculate hotspot score
		fileDuplication.HotspotScore = dd.calculateHotspotScore(fileDuplication)
		
		// Determine refactoring priority
		fileDuplication.RefactoringPriority = dd.determineFileRefactoringPriority(fileDuplication)
		
		metrics.DuplicationByFile[filePath] = fileDuplication
	}
}

// estimateTotalLines estimates the total lines of code in a file
func (dd *DuplicationDetector) estimateTotalLines(parseResult *ast.ParseResult) int {
	maxLine := 0
	
	for _, function := range parseResult.Functions {
		if function.EndLine > maxLine {
			maxLine = function.EndLine
		}
	}
	
	for _, class := range parseResult.Classes {
		for _, method := range class.Methods {
			if method.EndLine > maxLine {
				maxLine = method.EndLine
			}
		}
	}
	
	return maxLine
}

// calculateHotspotScore calculates duplication density score
func (dd *DuplicationDetector) calculateHotspotScore(fileDuplication FileDuplication) float64 {
	return fileDuplication.DuplicationRatio * 100
}

// determineFileRefactoringPriority determines priority for file-level refactoring
func (dd *DuplicationDetector) determineFileRefactoringPriority(fileDuplication FileDuplication) string {
	if fileDuplication.DuplicationRatio > 0.3 {
		return "critical"
	} else if fileDuplication.DuplicationRatio > 0.2 {
		return "high"
	} else if fileDuplication.DuplicationRatio > 0.1 {
		return "medium"
	}
	return "low"
}

// generateConsolidationOpportunities identifies specific refactoring opportunities
func (dd *DuplicationDetector) generateConsolidationOpportunities(metrics *DuplicationMetrics) []ConsolidationOpportunity {
	opportunities := []ConsolidationOpportunity{}
	opportunityID := 0
	
	// Analyze each cluster for consolidation potential
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	
	for _, cluster := range allClusters {
		if cluster.Priority == "critical" || cluster.Priority == "high" {
			opportunity := dd.generateConsolidationOpportunity(cluster, opportunityID)
			opportunities = append(opportunities, opportunity)
			opportunityID++
		}
	}
	
	// Sort by ROI score
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].ROIScore > opportunities[j].ROIScore
	})
	
	return opportunities
}

// generateConsolidationOpportunity creates a specific consolidation opportunity
func (dd *DuplicationDetector) generateConsolidationOpportunity(cluster DuplicationCluster, id int) ConsolidationOpportunity {
	opportunity := ConsolidationOpportunity{
		ID:                 fmt.Sprintf("consolidation_%d", id),
		AffectedFiles:      dd.extractAffectedFiles(cluster.Instances),
		AffectedFunctions:  dd.extractAffectedFunctions(cluster.Instances),
		EstimatedReduction: cluster.LineCount * (len(cluster.Instances) - 1),
	}
	
	// Determine consolidation type and strategy
	if cluster.Type == "exact" {
		opportunity.Type = "extract_function"
		opportunity.Description = "Extract identical code blocks into a shared utility function"
		opportunity.MaintenanceImprovement = "Eliminates exact duplicates, centralizes logic"
		opportunity.RefactoringSteps = []string{
			"Identify common parameters and return values",
			"Create new utility function",
			"Replace duplicated code with function calls",
			"Add comprehensive tests for new function",
		}
		opportunity.EstimatedEffort = len(cluster.Instances) * 2 // 2 hours per instance
	} else if cluster.Type == "structural" {
		opportunity.Type = "create_template"
		opportunity.Description = "Create template function or pattern for structurally similar code"
		opportunity.MaintenanceImprovement = "Standardizes patterns, reduces cognitive overhead"
		opportunity.RefactoringSteps = []string{
			"Analyze structural differences",
			"Design template or strategy pattern",
			"Implement template function",
			"Refactor instances to use template",
			"Update tests and documentation",
		}
		opportunity.EstimatedEffort = len(cluster.Instances) * 3 // 3 hours per instance
	} else {
		opportunity.Type = "standardize_code"
		opportunity.Description = "Standardize code patterns and naming conventions"
		opportunity.MaintenanceImprovement = "Improves consistency and readability"
		opportunity.RefactoringSteps = []string{
			"Define coding standards",
			"Refactor instances to follow standards",
			"Update linting rules",
			"Review and test changes",
		}
		opportunity.EstimatedEffort = len(cluster.Instances) * 1 // 1 hour per instance
	}
	
	// Calculate complexity reduction
	opportunity.ComplexityReduction = float64(opportunity.EstimatedReduction) / 10.0
	
	// Calculate ROI score
	maintenanceSavings := cluster.MaintenanceBurden * 0.2 // 20% maintenance reduction per year
	opportunity.ROIScore = maintenanceSavings / float64(opportunity.EstimatedEffort)
	
	return opportunity
}

// extractAffectedFiles gets unique file paths from instances
func (dd *DuplicationDetector) extractAffectedFiles(instances []DuplicationInstance) []string {
	fileSet := make(map[string]bool)
	for _, instance := range instances {
		fileSet[instance.FilePath] = true
	}
	
	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}
	
	return files
}

// extractAffectedFunctions gets function names from instances
func (dd *DuplicationDetector) extractAffectedFunctions(instances []DuplicationInstance) []string {
	functionSet := make(map[string]bool)
	for _, instance := range instances {
		if instance.FunctionName != "" {
			functionName := instance.FunctionName
			if instance.ClassName != "" {
				functionName = fmt.Sprintf("%s.%s", instance.ClassName, functionName)
			}
			functionSet[functionName] = true
		}
	}
	
	functions := make([]string, 0, len(functionSet))
	for function := range functionSet {
		functions = append(functions, function)
	}
	
	return functions
}

// analyzeImpact performs comprehensive impact analysis
func (dd *DuplicationDetector) analyzeImpact(metrics *DuplicationMetrics) DuplicationImpact {
	impact := DuplicationImpact{
		HotspotAnalysis: []DuplicationHotspot{},
	}
	
	// Calculate maintenance multiplier
	totalInstances := 0
	totalUniqueFunctions := 0
	
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	for _, cluster := range allClusters {
		totalInstances += len(cluster.Instances)
		totalUniqueFunctions++ // Each cluster represents one unique function that was duplicated
	}
	
	if totalUniqueFunctions > 0 {
		impact.MaintenanceMultiplier = float64(totalInstances) / float64(totalUniqueFunctions)
	}
	
	// Calculate technical debt score
	totalBurden := 0.0
	for _, cluster := range allClusters {
		totalBurden += cluster.MaintenanceBurden
	}
	impact.TechnicalDebtScore = totalBurden / 100.0 // Normalize to 0-100 scale
	
	// Calculate change risk factor
	impact.ChangeRiskFactor = math.Min(impact.MaintenanceMultiplier * 0.2, 2.0)
	
	// Estimate testing burden
	impact.TestingBurden = totalInstances * 2 // Assume 2 tests per duplicated instance
	
	// Determine overall codebase health
	if impact.TechnicalDebtScore > 50 {
		impact.CodebaseHealth = "critical"
	} else if impact.TechnicalDebtScore > 30 {
		impact.CodebaseHealth = "poor"
	} else if impact.TechnicalDebtScore > 15 {
		impact.CodebaseHealth = "fair"
	} else {
		impact.CodebaseHealth = "good"
	}
	
	// Generate hotspot analysis
	impact.HotspotAnalysis = dd.generateHotspotAnalysis(metrics)
	
	return impact
}

// generateHotspotAnalysis identifies high-duplication areas
func (dd *DuplicationDetector) generateHotspotAnalysis(metrics *DuplicationMetrics) []DuplicationHotspot {
	hotspots := []DuplicationHotspot{}
	
	// Analyze file-level hotspots
	for filePath, fileDuplication := range metrics.DuplicationByFile {
		if fileDuplication.HotspotScore > 20 { // More than 20% duplication
			hotspot := DuplicationHotspot{
				Location:          filePath,
				DuplicationScore:  fileDuplication.HotspotScore,
				AffectedFunctions: dd.countAffectedFunctionsInFile(filePath, metrics),
				MaintenanceRisk:   fileDuplication.RefactoringPriority,
				RecommendedAction: dd.generateHotspotRecommendation(fileDuplication),
			}
			hotspots = append(hotspots, hotspot)
		}
	}
	
	// Sort by duplication score
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].DuplicationScore > hotspots[j].DuplicationScore
	})
	
	return hotspots
}

// countAffectedFunctionsInFile counts functions affected by duplication in a file
func (dd *DuplicationDetector) countAffectedFunctionsInFile(filePath string, metrics *DuplicationMetrics) int {
	functionSet := make(map[string]bool)
	
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	for _, cluster := range allClusters {
		for _, instance := range cluster.Instances {
			if instance.FilePath == filePath && instance.FunctionName != "" {
				functionSet[instance.FunctionName] = true
			}
		}
	}
	
	return len(functionSet)
}

// generateHotspotRecommendation creates specific recommendation for a hotspot
func (dd *DuplicationDetector) generateHotspotRecommendation(fileDuplication FileDuplication) string {
	if fileDuplication.RefactoringPriority == "critical" {
		return "Immediate refactoring required - consider breaking file into modules"
	} else if fileDuplication.RefactoringPriority == "high" {
		return "High priority refactoring - extract common functionality"
	} else if fileDuplication.RefactoringPriority == "medium" {
		return "Medium priority - review for consolidation opportunities"
	}
	return "Monitor for increasing duplication"
}

// calculateAggregateMetrics computes overall duplication statistics
func (dd *DuplicationDetector) calculateAggregateMetrics(metrics *DuplicationMetrics) {
	totalDuplicatedLines := 0
	totalLines := 0
	
	// Calculate totals from all clusters
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	for _, cluster := range allClusters {
		clusterLines := cluster.LineCount * len(cluster.Instances)
		totalDuplicatedLines += clusterLines
	}
	
	// Estimate total lines from file metrics
	for _, fileDuplication := range metrics.DuplicationByFile {
		// Rough estimation based on duplication ratio
		if fileDuplication.DuplicationRatio > 0 {
			estimatedFileLines := int(float64(fileDuplication.InternalDuplication+fileDuplication.ExternalDuplication) / fileDuplication.DuplicationRatio)
			totalLines += estimatedFileLines
		} else {
			totalLines += 100 // Default estimate for files without duplication
		}
	}
	
	metrics.TotalDuplicatedLines = totalDuplicatedLines
	
	if totalLines > 0 {
		metrics.DuplicationRatio = float64(totalDuplicatedLines) / float64(totalLines)
	}
	
	// Calculate overall score (inverse of duplication ratio)
	metrics.OverallScore = math.Max(0, 100*(1-metrics.DuplicationRatio*2))
}

// generateRecommendations creates prioritized improvement recommendations
func (dd *DuplicationDetector) generateRecommendations(metrics *DuplicationMetrics) {
	recommendations := []DuplicationRecommendation{}
	
	// High-priority exact duplicate recommendations
	if len(metrics.ExactDuplicates) > 0 {
		criticalDuplicates := []string{}
		totalEstimatedHours := 0
		totalReduction := 0
		
		for _, cluster := range metrics.ExactDuplicates {
			if cluster.Priority == "critical" || cluster.Priority == "high" {
				criticalDuplicates = append(criticalDuplicates, cluster.ID)
				totalEstimatedHours += len(cluster.Instances) * 2
				totalReduction += cluster.LineCount * (len(cluster.Instances) - 1)
			}
		}
		
		if len(criticalDuplicates) > 0 {
			recommendations = append(recommendations, DuplicationRecommendation{
				Priority:          "critical",
				Category:          "refactoring",
				Title:             "Eliminate High-Impact Exact Duplicates",
				Description:       fmt.Sprintf("Refactor %d clusters of exact code duplication", len(criticalDuplicates)),
				Impact:            "high",
				Effort:            "medium",
				Clusters:          criticalDuplicates,
				Techniques:        []string{"extract_method", "create_utility_functions", "consolidate_logic"},
				EstimatedHours:    totalEstimatedHours,
				ExpectedReduction: totalReduction,
			})
		}
	}
	
	// Cross-file duplication recommendations
	if len(metrics.CrossFileDuplicates) > 0 {
		crossFileIds := []string{}
		for _, crossFile := range metrics.CrossFileDuplicates {
			crossFileIds = append(crossFileIds, crossFile.ClusterID)
		}
		
		recommendations = append(recommendations, DuplicationRecommendation{
			Priority:          "high",
			Category:          "architecture",
			Title:             "Address Cross-File Duplication",
			Description:       "Create shared modules for functionality duplicated across files",
			Impact:            "high",
			Effort:            "high",
			Clusters:          crossFileIds,
			Techniques:        []string{"create_shared_modules", "extract_common_utilities", "establish_code_patterns"},
			EstimatedHours:    len(metrics.CrossFileDuplicates) * 4,
			ExpectedReduction: dd.calculateCrossFileReduction(metrics.CrossFileDuplicates),
		})
	}
	
	// Structural duplication recommendations
	if len(metrics.StructuralDuplicates) > 0 {
		structuralIds := []string{}
		for _, cluster := range metrics.StructuralDuplicates {
			if cluster.Priority == "high" || cluster.Priority == "medium" {
				structuralIds = append(structuralIds, cluster.ID)
			}
		}
		
		if len(structuralIds) > 0 {
			recommendations = append(recommendations, DuplicationRecommendation{
				Priority:          "medium",
				Category:          "patterns",
				Title:             "Standardize Code Patterns",
				Description:       "Create templates and patterns for structurally similar code",
				Impact:            "medium",
				Effort:            "medium",
				Clusters:          structuralIds,
				Techniques:        []string{"template_method_pattern", "strategy_pattern", "code_generation"},
				EstimatedHours:    len(structuralIds) * 3,
				ExpectedReduction: dd.calculateStructuralReduction(metrics.StructuralDuplicates),
			})
		}
	}
	
	metrics.Recommendations = recommendations
}

// calculateCrossFileReduction estimates line reduction from cross-file consolidation
func (dd *DuplicationDetector) calculateCrossFileReduction(crossFiles []CrossFileDuplication) int {
	totalReduction := 0
	for _, crossFile := range crossFiles {
		totalReduction += crossFile.EstimatedSavings
	}
	return totalReduction
}

// calculateStructuralReduction estimates reduction from structural improvements
func (dd *DuplicationDetector) calculateStructuralReduction(structuralClusters []DuplicationCluster) int {
	totalReduction := 0
	for _, cluster := range structuralClusters {
		if cluster.Priority == "high" || cluster.Priority == "medium" {
			// Structural improvements typically reduce 30-50% of duplication
			reduction := int(float64(cluster.LineCount*len(cluster.Instances)) * 0.4)
			totalReduction += reduction
		}
	}
	return totalReduction
}

// generateSummary creates executive-level summary
func (dd *DuplicationDetector) generateSummary(metrics *DuplicationMetrics) {
	summary := DuplicationSummary{
		HealthScore: metrics.OverallScore,
	}
	
	// Determine risk level
	if metrics.DuplicationRatio > 0.25 {
		summary.RiskLevel = "critical"
		summary.MaintenanceBurden = "critical"
	} else if metrics.DuplicationRatio > 0.15 {
		summary.RiskLevel = "high"
		summary.MaintenanceBurden = "high"
	} else if metrics.DuplicationRatio > 0.08 {
		summary.RiskLevel = "medium"
		summary.MaintenanceBurden = "medium"
	} else {
		summary.RiskLevel = "low"
		summary.MaintenanceBurden = "low"
	}
	
	// Count clusters needing refactoring
	allClusters := append(append(metrics.ExactDuplicates, metrics.StructuralDuplicates...), metrics.TokenDuplicates...)
	for _, cluster := range allClusters {
		if cluster.Priority == "critical" || cluster.Priority == "high" {
			summary.RefactoringNeeded++
		}
	}
	
	// Calculate potential savings
	for _, opportunity := range metrics.ConsolidationOps {
		summary.PotentialSavings += opportunity.EstimatedReduction
	}
	
	summary.RecommendedActions = len(metrics.Recommendations)
	
	metrics.Summary = summary
}