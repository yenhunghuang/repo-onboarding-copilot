package metrics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/internal/analysis/ast"
)

// QualityReporter generates comprehensive quality reports by aggregating all analysis components
type QualityReporter struct {
	config              QualityReportConfig
	complexityAnalyzer  *ComplexityAnalyzer
	duplicationDetector *DuplicationDetector
	debtScorer          *DebtScorer
	coverageAnalyzer    *CoverageAnalyzer
	performanceAnalyzer *PerformanceAnalyzer
	maintainabilityCalc *MaintainabilityCalculator
}

// QualityReportConfig defines configuration for quality reporting
type QualityReportConfig struct {
	ReportFormat            ReportFormat      `yaml:"report_format" json:"report_format"`
	IncludeExecutiveSummary bool              `yaml:"include_executive_summary" json:"include_executive_summary"`
	IncludeTrendAnalysis    bool              `yaml:"include_trend_analysis" json:"include_trend_analysis"`
	MaxRecommendations      int               `yaml:"max_recommendations" json:"max_recommendations"`
	EffortEstimationModel   string            `yaml:"effort_estimation_model" json:"effort_estimation_model"`
	RoadmapTimeframe        int               `yaml:"roadmap_timeframe" json:"roadmap_timeframe"` // weeks
	Thresholds              QualityThresholds `yaml:"thresholds" json:"thresholds"`
	WeightingFactors        QualityWeights    `yaml:"weighting_factors" json:"weighting_factors"`
}

// QualityThresholds defines quality score thresholds
type QualityThresholds struct {
	Excellent float64 `yaml:"excellent" json:"excellent"` // >= 90
	Good      float64 `yaml:"good" json:"good"`           // >= 75
	Fair      float64 `yaml:"fair" json:"fair"`           // >= 60
	Poor      float64 `yaml:"poor" json:"poor"`           // < 60
}

// QualityWeights defines weights for different quality aspects
type QualityWeights struct {
	Complexity      float64 `yaml:"complexity" json:"complexity"`           // 20%
	Duplication     float64 `yaml:"duplication" json:"duplication"`         // 15%
	TechnicalDebt   float64 `yaml:"technical_debt" json:"technical_debt"`   // 25%
	Coverage        float64 `yaml:"coverage" json:"coverage"`               // 20%
	Performance     float64 `yaml:"performance" json:"performance"`         // 10%
	Maintainability float64 `yaml:"maintainability" json:"maintainability"` // 10%
}

// ReportFormat defines the output format for quality reports
type ReportFormat string

const (
	FormatJSON     ReportFormat = "json"
	FormatMarkdown ReportFormat = "markdown"
	FormatHTML     ReportFormat = "html"
	FormatConsole  ReportFormat = "console"
)

// QualityReport represents the comprehensive quality analysis report
type QualityReport struct {
	GeneratedAt      time.Time               `json:"generated_at"`
	ProjectName      string                  `json:"project_name"`
	OverallScore     float64                 `json:"overall_score"`
	QualityGrade     string                  `json:"quality_grade"`
	ComponentScores  ComponentScores         `json:"component_scores"`
	Dashboard        QualityDashboard        `json:"dashboard"`
	Recommendations  []QualityRecommendation `json:"recommendations"`
	Roadmap          QualityRoadmap          `json:"roadmap"`
	ExecutiveSummary *ExecutiveSummary       `json:"executive_summary,omitempty"`
	TrendAnalysis    *QualityTrend           `json:"trend_analysis,omitempty"`
	DetailedMetrics  DetailedMetrics         `json:"detailed_metrics"`
}

// ComponentScores contains scores for each analysis component
type ComponentScores struct {
	Complexity      float64 `json:"complexity"`
	Duplication     float64 `json:"duplication"`
	TechnicalDebt   float64 `json:"technical_debt"`
	Coverage        float64 `json:"coverage"`
	Performance     float64 `json:"performance"`
	Maintainability float64 `json:"maintainability"`
}

// QualityDashboard provides visual indicators and trend analysis
type QualityDashboard struct {
	OverallHealth      HealthIndicator            `json:"overall_health"`
	ComponentHealth    map[string]HealthIndicator `json:"component_health"`
	TrendIndicators    []TrendIndicator           `json:"trend_indicators"`
	AlertsAndWarnings  []QualityAlert             `json:"alerts_and_warnings"`
	KeyMetrics         []KeyMetric                `json:"key_metrics"`
	ProgressIndicators []ProgressIndicator        `json:"progress_indicators"`
}

// HealthIndicator represents the health status of a component
type HealthIndicator struct {
	Score       float64 `json:"score"`
	Status      string  `json:"status"` // excellent, good, fair, poor
	Color       string  `json:"color"`  // green, yellow, orange, red
	Icon        string  `json:"icon"`   // visual indicator
	Description string  `json:"description"`
}

// TrendIndicator shows trend analysis for quality metrics
type TrendIndicator struct {
	Component    string  `json:"component"`
	Trend        string  `json:"trend"`        // improving, stable, degrading
	ChangeRate   float64 `json:"change_rate"`  // percentage change
	Direction    string  `json:"direction"`    // up, down, stable
	Significance string  `json:"significance"` // high, medium, low
}

// QualityAlert represents warnings and critical issues
type QualityAlert struct {
	Severity       string `json:"severity"` // critical, warning, info
	Component      string `json:"component"`
	Message        string `json:"message"`
	Impact         string `json:"impact"` // high, medium, low
	ActionRequired string `json:"action_required"`
}

// KeyMetric represents important quality metrics for dashboard
type KeyMetric struct {
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Target      float64 `json:"target"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}

// ProgressIndicator shows progress towards quality goals
type ProgressIndicator struct {
	Goal     string  `json:"goal"`
	Current  float64 `json:"current"`
	Target   float64 `json:"target"`
	Progress float64 `json:"progress"` // percentage
	Timeline string  `json:"timeline"`
	Status   string  `json:"status"`
}

// QualityRecommendation represents actionable recommendations
type QualityRecommendation struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Category     RecommendationCategory `json:"category"`
	Priority     Priority               `json:"priority"`
	Impact       ImpactLevel            `json:"impact"`
	Effort       EffortLevel            `json:"effort"`
	EffortHours  float64                `json:"effort_hours"`
	ROI          float64                `json:"roi"`
	Component    string                 `json:"component"`
	Files        []string               `json:"files"`
	Actions      []RecommendationAction `json:"actions"`
	Benefits     []string               `json:"benefits"`
	Risks        []string               `json:"risks"`
	Dependencies []string               `json:"dependencies"`
	Timeline     string                 `json:"timeline"`
}

// RecommendationCategory categorizes recommendations
type RecommendationCategory string

const (
	CategoryQuickWins               RecommendationCategory = "quick_wins"
	CategoryStrategicImprovements   RecommendationCategory = "strategic_improvements"
	CategoryLongTermGoals           RecommendationCategory = "long_term_goals"
	CategoryCriticalFixes           RecommendationCategory = "critical_fixes"
	CategoryPerformanceOptimization RecommendationCategory = "performance_optimization"
	CategoryTechnicalDebtReduction  RecommendationCategory = "technical_debt_reduction"
)

// Priority levels for recommendations
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// ImpactLevel represents the potential impact of a recommendation
type ImpactLevel string

const (
	ImpactHigh   ImpactLevel = "high"
	ImpactMedium ImpactLevel = "medium"
	ImpactLow    ImpactLevel = "low"
)

// EffortLevel represents the effort required for a recommendation
type EffortLevel string

const (
	EffortHigh   EffortLevel = "high"
	EffortMedium EffortLevel = "medium"
	EffortLow    EffortLevel = "low"
)

// RecommendationAction represents specific actions to take
type RecommendationAction struct {
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	Command        string   `json:"command,omitempty"`
	Files          []string `json:"files,omitempty"`
	EstimatedHours float64  `json:"estimated_hours"`
}

// QualityRoadmap provides improvement planning with milestones
type QualityRoadmap struct {
	Overview       string             `json:"overview"`
	TimeframeWeeks int                `json:"timeframe_weeks"`
	Milestones     []QualityMilestone `json:"milestones"`
	Phases         []ImprovementPhase `json:"phases"`
	ResourcePlan   ResourcePlan       `json:"resource_plan"`
	RiskAssessment []RoadmapRisk      `json:"risk_assessment"`
	SuccessMetrics []SuccessMetric    `json:"success_metrics"`
}

// QualityMilestone represents a significant quality improvement milestone
type QualityMilestone struct {
	Name            string    `json:"name"`
	TargetDate      time.Time `json:"target_date"`
	Description     string    `json:"description"`
	Goals           []string  `json:"goals"`
	Deliverables    []string  `json:"deliverables"`
	SuccessCriteria []string  `json:"success_criteria"`
	Dependencies    []string  `json:"dependencies"`
	EstimatedHours  float64   `json:"estimated_hours"`
}

// ImprovementPhase represents phases of quality improvement
type ImprovementPhase struct {
	Name            string             `json:"name"`
	Duration        string             `json:"duration"`
	Focus           []string           `json:"focus"`
	Recommendations []string           `json:"recommendations"`
	ExpectedImpact  map[string]float64 `json:"expected_impact"`
	ResourceNeeds   PhaseResourceNeeds `json:"resource_needs"`
}

// PhaseResourceNeeds represents resource requirements for a phase
type PhaseResourceNeeds struct {
	DeveloperHours float64  `json:"developer_hours"`
	QAHours        float64  `json:"qa_hours"`
	ReviewHours    float64  `json:"review_hours"`
	Specialists    []string `json:"specialists"`
}

// ResourcePlan provides overall resource planning
type ResourcePlan struct {
	TotalDeveloperHours float64  `json:"total_developer_hours"`
	TotalQAHours        float64  `json:"total_qa_hours"`
	TotalReviewHours    float64  `json:"total_review_hours"`
	EstimatedCost       float64  `json:"estimated_cost"`
	TeamSize            int      `json:"team_size"`
	Duration            string   `json:"duration"`
	SkillsNeeded        []string `json:"skills_needed"`
}

// RoadmapRisk represents risks in the improvement roadmap
type RoadmapRisk struct {
	Description string `json:"description"`
	Probability string `json:"probability"` // high, medium, low
	Impact      string `json:"impact"`      // high, medium, low
	Mitigation  string `json:"mitigation"`
	Owner       string `json:"owner"`
}

// SuccessMetric defines success criteria for quality improvements
type SuccessMetric struct {
	Name        string  `json:"name"`
	Current     float64 `json:"current"`
	Target      float64 `json:"target"`
	Unit        string  `json:"unit"`
	Timeline    string  `json:"timeline"`
	Measurement string  `json:"measurement"`
}

// ExecutiveSummary provides high-level overview for stakeholders
type ExecutiveSummary struct {
	OverallAssessment  string            `json:"overall_assessment"`
	KeyFindings        []string          `json:"key_findings"`
	CriticalIssues     []string          `json:"critical_issues"`
	TopRecommendations []string          `json:"top_recommendations"`
	BusinessImpact     BusinessImpact    `json:"business_impact"`
	InvestmentRequired InvestmentSummary `json:"investment_required"`
	ExpectedOutcomes   []ExpectedOutcome `json:"expected_outcomes"`
	NextSteps          []string          `json:"next_steps"`
}

// BusinessImpact describes business implications of quality issues
type BusinessImpact struct {
	RiskLevel           string  `json:"risk_level"`
	MaintenanceCost     float64 `json:"maintenance_cost"`
	DevelopmentVelocity string  `json:"development_velocity"`
	TechnicalDebtCost   float64 `json:"technical_debt_cost"`
	QualityRisk         string  `json:"quality_risk"`
	CustomerImpact      string  `json:"customer_impact"`
}

// InvestmentSummary provides cost/benefit analysis
type InvestmentSummary struct {
	TotalInvestmentHours float64 `json:"total_investment_hours"`
	EstimatedCost        float64 `json:"estimated_cost"`
	ExpectedSavings      float64 `json:"expected_savings"`
	PaybackPeriod        string  `json:"payback_period"`
	ROI                  float64 `json:"roi"`
}

// ExpectedOutcome describes expected results from improvements
type ExpectedOutcome struct {
	Area       string `json:"area"`
	Current    string `json:"current"`
	Improved   string `json:"improved"`
	Timeline   string `json:"timeline"`
	Confidence string `json:"confidence"`
}

// QualityTrend provides trend analysis over time
type QualityTrend struct {
	Period           string                    `json:"period"`
	OverallTrend     TrendDirection            `json:"overall_trend"`
	ComponentTrends  map[string]TrendDirection `json:"component_trends"`
	HistoricalData   []HistoricalDataPoint     `json:"historical_data"`
	TrendPredictions []TrendPrediction         `json:"trend_predictions"`
	SeasonalPatterns []SeasonalPattern         `json:"seasonal_patterns"`
	InflectionPoints []InflectionPoint         `json:"inflection_points"`
}

// TrendDirection represents the direction of a trend
type TrendDirection struct {
	Direction    string  `json:"direction"`    // improving, stable, degrading
	Velocity     float64 `json:"velocity"`     // rate of change
	Confidence   float64 `json:"confidence"`   // confidence in trend
	Significance string  `json:"significance"` // high, medium, low
}

// HistoricalDataPoint represents a point in quality history
type HistoricalDataPoint struct {
	Timestamp time.Time       `json:"timestamp"`
	Scores    ComponentScores `json:"scores"`
	Events    []QualityEvent  `json:"events"`
}

// QualityEvent represents events that affected quality
type QualityEvent struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// TrendPrediction provides predictions for future quality trends
type TrendPrediction struct {
	Component   string   `json:"component"`
	TimeFrame   string   `json:"time_frame"`
	Prediction  float64  `json:"prediction"`
	Confidence  float64  `json:"confidence"`
	Assumptions []string `json:"assumptions"`
}

// SeasonalPattern represents seasonal quality patterns
type SeasonalPattern struct {
	Pattern     string  `json:"pattern"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Timing      string  `json:"timing"`
}

// InflectionPoint represents significant changes in quality trends
type InflectionPoint struct {
	Date      time.Time `json:"date"`
	Component string    `json:"component"`
	Change    float64   `json:"change"`
	Cause     string    `json:"cause"`
	Impact    string    `json:"impact"`
}

// DetailedMetrics contains all detailed metrics from individual analyzers
type DetailedMetrics struct {
	Complexity      *ComplexityMetrics      `json:"complexity,omitempty"`
	Duplication     *DuplicationMetrics     `json:"duplication,omitempty"`
	TechnicalDebt   *TechnicalDebtMetrics   `json:"technical_debt,omitempty"`
	Coverage        *CoverageMetrics        `json:"coverage,omitempty"`
	Performance     *PerformanceMetrics     `json:"performance,omitempty"`
	Maintainability *MaintainabilityMetrics `json:"maintainability,omitempty"`
}

// NewQualityReporter creates a new quality reporter with all analyzers
func NewQualityReporter(config QualityReportConfig) *QualityReporter {
	if config.MaxRecommendations == 0 {
		config.MaxRecommendations = 20
	}
	if config.RoadmapTimeframe == 0 {
		config.RoadmapTimeframe = 12 // 12 weeks default
	}
	if config.EffortEstimationModel == "" {
		config.EffortEstimationModel = "complexity_based"
	}

	// Set default thresholds
	if config.Thresholds.Excellent == 0 {
		config.Thresholds = QualityThresholds{
			Excellent: 90.0,
			Good:      75.0,
			Fair:      60.0,
			Poor:      60.0,
		}
	}

	// Set default weights
	if config.WeightingFactors.Complexity == 0 {
		config.WeightingFactors = QualityWeights{
			Complexity:      0.20,
			Duplication:     0.15,
			TechnicalDebt:   0.25,
			Coverage:        0.20,
			Performance:     0.10,
			Maintainability: 0.10,
		}
	}

	return &QualityReporter{
		config:              config,
		complexityAnalyzer:  NewComplexityAnalyzer(),
		duplicationDetector: NewDuplicationDetector(),
		debtScorer:          NewDebtScorer(),
		coverageAnalyzer:    NewCoverageAnalyzer(),
		performanceAnalyzer: NewPerformanceAnalyzer(),
		maintainabilityCalc: NewMaintainabilityCalculator(),
	}
}

// GenerateQualityReport creates a comprehensive quality report
func (qr *QualityReporter) GenerateQualityReport(ctx context.Context, fileContents map[string]string) (*QualityReport, error) {
	if len(fileContents) == 0 {
		return nil, fmt.Errorf("no files provided for analysis")
	}

	// Run all analysis components in parallel for performance
	type analysisResult struct {
		complexity      *ComplexityMetrics
		duplication     *DuplicationMetrics
		technicalDebt   *TechnicalDebtMetrics
		coverage        *CoverageMetrics
		performance     *PerformanceMetrics
		maintainability *MaintainabilityMetrics
		err             error
	}

	resultChan := make(chan analysisResult, 1)

	go func() {
		defer close(resultChan)

		var result analysisResult

		// Parse files into parse results
		parseResults, err := qr.parseFiles(fileContents)
		if err != nil {
			result.err = fmt.Errorf("failed to parse files: %w", err)
			resultChan <- result
			return
		}

		// Run all analyses
		if result.complexity, err = qr.complexityAnalyzer.AnalyzeComplexity(ctx, parseResults); err != nil {
			result.err = fmt.Errorf("complexity analysis failed: %w", err)
			resultChan <- result
			return
		}

		if result.duplication, err = qr.duplicationDetector.DetectDuplication(ctx, parseResults); err != nil {
			result.err = fmt.Errorf("duplication detection failed: %w", err)
			resultChan <- result
			return
		}

		if result.technicalDebt, err = qr.debtScorer.AnalyzeDebt(ctx, parseResults, result.complexity, result.duplication); err != nil {
			result.err = fmt.Errorf("technical debt analysis failed: %w", err)
			resultChan <- result
			return
		}

		if result.coverage, err = qr.coverageAnalyzer.AnalyzeCoverage(ctx, parseResults, result.complexity); err != nil {
			result.err = fmt.Errorf("coverage analysis failed: %w", err)
			resultChan <- result
			return
		}

		if result.performance, err = qr.performanceAnalyzer.AnalyzePerformance(ctx, parseResults, result.complexity); err != nil {
			result.err = fmt.Errorf("performance analysis failed: %w", err)
			resultChan <- result
			return
		}

		if result.maintainability, err = qr.maintainabilityCalc.AnalyzeMaintainability(ctx, parseResults, result.complexity); err != nil {
			result.err = fmt.Errorf("maintainability calculation failed: %w", err)
			resultChan <- result
			return
		}

		resultChan <- result
	}()

	// Wait for results with context cancellation
	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, result.err
		}

		// Generate comprehensive report
		return qr.generateReport(
			result.complexity,
			result.duplication,
			result.technicalDebt,
			result.coverage,
			result.performance,
			result.maintainability,
		), nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// parseFiles converts file contents to parse results
func (qr *QualityReporter) parseFiles(fileContents map[string]string) ([]*ast.ParseResult, error) {
	var parseResults []*ast.ParseResult

	// Create parser instance
	parser, err := ast.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %v", err)
	}

	for filename, content := range fileContents {
		result, err := parser.ParseFile(context.Background(), filename, []byte(content))
		if err != nil {
			// Log warning but continue with other files
			continue
		}
		parseResults = append(parseResults, result)
	}

	if len(parseResults) == 0 {
		return nil, fmt.Errorf("no files could be parsed")
	}

	return parseResults, nil
}

// generateReport creates the comprehensive quality report from all analysis results
func (qr *QualityReporter) generateReport(
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) *QualityReport {
	now := time.Now()

	// Calculate component scores
	componentScores := qr.calculateComponentScores(complexity, duplication, technicalDebt, coverage, performance, maintainability)

	// Calculate overall score
	overallScore := qr.calculateOverallScore(componentScores)

	// Generate quality grade
	qualityGrade := qr.determineQualityGrade(overallScore)

	// Generate dashboard
	dashboard := qr.generateDashboard(componentScores, complexity, duplication, technicalDebt, coverage, performance, maintainability)

	// Generate recommendations
	recommendations := qr.generateRecommendations(complexity, duplication, technicalDebt, coverage, performance, maintainability)

	// Sort and limit recommendations
	recommendations = qr.rankAndLimitRecommendations(recommendations)

	// Generate roadmap
	roadmap := qr.generateRoadmap(recommendations, componentScores)

	// Generate executive summary if requested
	var executiveSummary *ExecutiveSummary
	if qr.config.IncludeExecutiveSummary {
		executiveSummary = qr.generateExecutiveSummary(overallScore, qualityGrade, componentScores, recommendations)
	}

	// Generate trend analysis if requested
	var trendAnalysis *QualityTrend
	if qr.config.IncludeTrendAnalysis {
		trendAnalysis = qr.generateTrendAnalysis(componentScores)
	}

	return &QualityReport{
		GeneratedAt:      now,
		ProjectName:      "Repository Analysis", // Could be made configurable
		OverallScore:     overallScore,
		QualityGrade:     qualityGrade,
		ComponentScores:  componentScores,
		Dashboard:        dashboard,
		Recommendations:  recommendations,
		Roadmap:          roadmap,
		ExecutiveSummary: executiveSummary,
		TrendAnalysis:    trendAnalysis,
		DetailedMetrics: DetailedMetrics{
			Complexity:      complexity,
			Duplication:     duplication,
			TechnicalDebt:   technicalDebt,
			Coverage:        coverage,
			Performance:     performance,
			Maintainability: maintainability,
		},
	}
}

// calculateComponentScores calculates normalized scores for each component
func (qr *QualityReporter) calculateComponentScores(
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) ComponentScores {
	return ComponentScores{
		Complexity:      qr.normalizeScore(complexity.OverallScore),
		Duplication:     qr.normalizeScore(duplication.OverallScore),
		TechnicalDebt:   qr.normalizeScore(technicalDebt.OverallScore),
		Coverage:        qr.normalizeScore(coverage.OverallScore),
		Performance:     qr.normalizeScore(performance.OverallScore),
		Maintainability: qr.normalizeScore(maintainability.OverallIndex),
	}
}

// normalizeScore ensures scores are between 0-100
func (qr *QualityReporter) normalizeScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// calculateOverallScore computes weighted overall quality score
func (qr *QualityReporter) calculateOverallScore(scores ComponentScores) float64 {
	weights := qr.config.WeightingFactors

	overallScore := scores.Complexity*weights.Complexity +
		scores.Duplication*weights.Duplication +
		scores.TechnicalDebt*weights.TechnicalDebt +
		scores.Coverage*weights.Coverage +
		scores.Performance*weights.Performance +
		scores.Maintainability*weights.Maintainability

	return math.Round(overallScore*100) / 100
}

// determineQualityGrade assigns a grade based on overall score
func (qr *QualityReporter) determineQualityGrade(score float64) string {
	thresholds := qr.config.Thresholds

	switch {
	case score >= thresholds.Excellent:
		return "Excellent"
	case score >= thresholds.Good:
		return "Good"
	case score >= thresholds.Fair:
		return "Fair"
	default:
		return "Poor"
	}
}

// generateDashboard creates visual indicators and trend analysis
func (qr *QualityReporter) generateDashboard(
	scores ComponentScores,
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) QualityDashboard {
	overallScore := qr.calculateOverallScore(scores)

	// Generate health indicators
	overallHealth := qr.createHealthIndicator(overallScore, "Overall Quality")
	componentHealth := map[string]HealthIndicator{
		"complexity":      qr.createHealthIndicator(scores.Complexity, "Code Complexity"),
		"duplication":     qr.createHealthIndicator(scores.Duplication, "Code Duplication"),
		"technical_debt":  qr.createHealthIndicator(scores.TechnicalDebt, "Technical Debt"),
		"coverage":        qr.createHealthIndicator(scores.Coverage, "Test Coverage"),
		"performance":     qr.createHealthIndicator(scores.Performance, "Performance"),
		"maintainability": qr.createHealthIndicator(scores.Maintainability, "Maintainability"),
	}

	// Generate trend indicators
	trendIndicators := qr.generateTrendIndicators(scores)

	// Generate alerts and warnings
	alerts := qr.generateQualityAlerts(scores, complexity, duplication, technicalDebt, coverage, performance, maintainability)

	// Generate key metrics
	keyMetrics := qr.generateKeyMetrics(scores, complexity, duplication, technicalDebt, coverage, performance, maintainability)

	// Generate progress indicators
	progressIndicators := qr.generateProgressIndicators(scores)

	return QualityDashboard{
		OverallHealth:      overallHealth,
		ComponentHealth:    componentHealth,
		TrendIndicators:    trendIndicators,
		AlertsAndWarnings:  alerts,
		KeyMetrics:         keyMetrics,
		ProgressIndicators: progressIndicators,
	}
}

// createHealthIndicator creates a health indicator for a given score
func (qr *QualityReporter) createHealthIndicator(score float64, description string) HealthIndicator {
	var status, color, icon string

	thresholds := qr.config.Thresholds
	switch {
	case score >= thresholds.Excellent:
		status, color, icon = "excellent", "green", "✅"
	case score >= thresholds.Good:
		status, color, icon = "good", "yellow", "⚠️"
	case score >= thresholds.Fair:
		status, color, icon = "fair", "orange", "🔶"
	default:
		status, color, icon = "poor", "red", "❌"
	}

	return HealthIndicator{
		Score:       score,
		Status:      status,
		Color:       color,
		Icon:        icon,
		Description: description,
	}
}

// generateTrendIndicators creates trend analysis indicators
func (qr *QualityReporter) generateTrendIndicators(scores ComponentScores) []TrendIndicator {
	// For now, return stable trends as we don't have historical data
	// In a real implementation, this would analyze historical trends
	components := map[string]float64{
		"complexity":      scores.Complexity,
		"duplication":     scores.Duplication,
		"technical_debt":  scores.TechnicalDebt,
		"coverage":        scores.Coverage,
		"performance":     scores.Performance,
		"maintainability": scores.Maintainability,
	}

	var indicators []TrendIndicator
	for component, score := range components {
		// Simulate trend analysis based on current score
		trend := "stable"
		direction := "stable"
		changeRate := 0.0
		significance := "low"

		if score < 60 {
			trend = "needs_attention"
			direction = "down"
			changeRate = -2.5
			significance = "high"
		} else if score > 85 {
			trend = "improving"
			direction = "up"
			changeRate = 1.2
			significance = "medium"
		}

		indicators = append(indicators, TrendIndicator{
			Component:    component,
			Trend:        trend,
			ChangeRate:   changeRate,
			Direction:    direction,
			Significance: significance,
		})
	}

	return indicators
}

// generateQualityAlerts creates alerts and warnings based on analysis results
func (qr *QualityReporter) generateQualityAlerts(
	scores ComponentScores,
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) []QualityAlert {
	var alerts []QualityAlert

	// Critical issues
	if scores.TechnicalDebt < 40 {
		alerts = append(alerts, QualityAlert{
			Severity:       "critical",
			Component:      "technical_debt",
			Message:        "Critical technical debt level detected",
			Impact:         "high",
			ActionRequired: "Immediate technical debt reduction required",
		})
	}

	if scores.Performance < 50 {
		alerts = append(alerts, QualityAlert{
			Severity:       "critical",
			Component:      "performance",
			Message:        "Performance issues detected",
			Impact:         "high",
			ActionRequired: "Performance optimization required",
		})
	}

	// Warnings
	if scores.Complexity < 60 {
		alerts = append(alerts, QualityAlert{
			Severity:       "warning",
			Component:      "complexity",
			Message:        "High complexity detected in multiple functions",
			Impact:         "medium",
			ActionRequired: "Consider refactoring complex functions",
		})
	}

	if scores.Coverage < 70 {
		alerts = append(alerts, QualityAlert{
			Severity:       "warning",
			Component:      "coverage",
			Message:        "Low test coverage detected",
			Impact:         "medium",
			ActionRequired: "Increase test coverage for critical paths",
		})
	}

	if scores.Duplication < 70 {
		alerts = append(alerts, QualityAlert{
			Severity:       "warning",
			Component:      "duplication",
			Message:        "Code duplication detected",
			Impact:         "medium",
			ActionRequired: "Consolidate duplicated code",
		})
	}

	// Info alerts
	if scores.Maintainability < 80 {
		alerts = append(alerts, QualityAlert{
			Severity:       "info",
			Component:      "maintainability",
			Message:        "Maintainability could be improved",
			Impact:         "low",
			ActionRequired: "Consider maintainability improvements",
		})
	}

	return alerts
}

// generateKeyMetrics creates key metrics for the dashboard
func (qr *QualityReporter) generateKeyMetrics(
	scores ComponentScores,
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) []KeyMetric {
	return []KeyMetric{
		{
			Name:        "Average Complexity",
			Value:       complexity.AverageComplexity,
			Unit:        "score",
			Target:      10.0,
			Status:      qr.getMetricStatus(complexity.AverageComplexity, 10.0, false),
			Description: "Average cyclomatic complexity across all functions",
		},
		{
			Name:        "Code Duplication",
			Value:       duplication.DuplicationRatio * 100,
			Unit:        "%",
			Target:      5.0,
			Status:      qr.getMetricStatus(duplication.DuplicationRatio*100, 5.0, false),
			Description: "Percentage of duplicated code",
		},
		{
			Name:        "Technical Debt Hours",
			Value:       technicalDebt.TotalDebtHours,
			Unit:        "hours",
			Target:      40.0,
			Status:      qr.getMetricStatus(technicalDebt.TotalDebtHours, 40.0, false),
			Description: "Estimated hours to fix technical debt",
		},
		{
			Name:        "Test Coverage",
			Value:       coverage.EstimatedCoverage,
			Unit:        "%",
			Target:      80.0,
			Status:      qr.getMetricStatus(coverage.EstimatedCoverage, 80.0, true),
			Description: "Estimated test coverage percentage",
		},
		{
			Name:        "Performance Score",
			Value:       performance.OverallScore,
			Unit:        "score",
			Target:      85.0,
			Status:      qr.getMetricStatus(performance.OverallScore, 85.0, true),
			Description: "Overall performance analysis score",
		},
		{
			Name:        "Maintainability Index",
			Value:       maintainability.AverageIndex,
			Unit:        "score",
			Target:      75.0,
			Status:      qr.getMetricStatus(maintainability.AverageIndex, 75.0, true),
			Description: "Average maintainability index",
		},
	}
}

// getMetricStatus determines the status of a metric compared to its target
func (qr *QualityReporter) getMetricStatus(value, target float64, higherIsBetter bool) string {
	var ratio float64
	if higherIsBetter {
		ratio = value / target
	} else {
		ratio = target / value
	}

	switch {
	case ratio >= 1.0:
		return "excellent"
	case ratio >= 0.8:
		return "good"
	case ratio >= 0.6:
		return "fair"
	default:
		return "poor"
	}
}

// generateProgressIndicators creates progress indicators for quality goals
func (qr *QualityReporter) generateProgressIndicators(scores ComponentScores) []ProgressIndicator {
	targets := map[string]float64{
		"Overall Quality": 85.0,
		"Code Complexity": 80.0,
		"Technical Debt":  75.0,
		"Test Coverage":   80.0,
		"Performance":     85.0,
		"Maintainability": 80.0,
	}

	currentValues := map[string]float64{
		"Overall Quality": qr.calculateOverallScore(scores),
		"Code Complexity": scores.Complexity,
		"Technical Debt":  scores.TechnicalDebt,
		"Test Coverage":   scores.Coverage,
		"Performance":     scores.Performance,
		"Maintainability": scores.Maintainability,
	}

	var indicators []ProgressIndicator
	for goal, target := range targets {
		current := currentValues[goal]
		progress := (current / target) * 100
		if progress > 100 {
			progress = 100
		}

		status := "in_progress"
		timeline := "4-8 weeks"

		if progress >= 100 {
			status = "completed"
			timeline = "achieved"
		} else if progress < 50 {
			status = "needs_attention"
			timeline = "8-12 weeks"
		}

		indicators = append(indicators, ProgressIndicator{
			Goal:     goal,
			Current:  current,
			Target:   target,
			Progress: progress,
			Timeline: timeline,
			Status:   status,
		})
	}

	return indicators
} // generateRecommendations creates actionable recommendations from all analysis results
func (qr *QualityReporter) generateRecommendations(
	complexity *ComplexityMetrics,
	duplication *DuplicationMetrics,
	technicalDebt *TechnicalDebtMetrics,
	coverage *CoverageMetrics,
	performance *PerformanceMetrics,
	maintainability *MaintainabilityMetrics,
) []QualityRecommendation {
	var recommendations []QualityRecommendation

	// Generate complexity-based recommendations
	recommendations = append(recommendations, qr.generateComplexityRecommendations(complexity)...)

	// Generate duplication-based recommendations
	recommendations = append(recommendations, qr.generateDuplicationRecommendations(duplication)...)

	// Generate technical debt recommendations
	recommendations = append(recommendations, qr.generateDebtRecommendations(technicalDebt)...)

	// Generate coverage recommendations
	recommendations = append(recommendations, qr.generateCoverageRecommendations(coverage)...)

	// Generate performance recommendations
	recommendations = append(recommendations, qr.generatePerformanceRecommendations(performance)...)

	// Generate maintainability recommendations
	recommendations = append(recommendations, qr.generateMaintainabilityRecommendations(maintainability)...)

	return recommendations
}

// generateComplexityRecommendations creates recommendations based on complexity analysis
func (qr *QualityReporter) generateComplexityRecommendations(complexity *ComplexityMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	// High complexity functions
	for _, funcMetric := range complexity.FunctionMetrics {
		if funcMetric.CyclomaticValue > 15 {
			category := qr.categorizeByComplexity(funcMetric.CyclomaticValue)
			linesOfCode := funcMetric.EndLine - funcMetric.StartLine + 1
			effort := qr.estimateRefactoringEffort(funcMetric.CyclomaticValue, linesOfCode)

			recommendations = append(recommendations, QualityRecommendation{
				ID:          fmt.Sprintf("COMPLEX-%d", id),
				Title:       fmt.Sprintf("Refactor high complexity function: %s", funcMetric.Name),
				Description: fmt.Sprintf("Function has cyclomatic complexity of %d (threshold: 15)", funcMetric.CyclomaticValue),
				Category:    category,
				Priority:    qr.determinePriority(float64(funcMetric.CyclomaticValue), 15, 25),
				Impact:      qr.determineImpact(float64(funcMetric.CyclomaticValue), 15),
				Effort:      qr.determineEffortLevel(effort),
				EffortHours: effort,
				ROI:         qr.calculateROI(effort, float64(funcMetric.CyclomaticValue)),
				Component:   "complexity",
				Files:       []string{funcMetric.FilePath},
				Actions: []RecommendationAction{
					{
						Type:           "refactor",
						Description:    "Extract methods to reduce complexity",
						EstimatedHours: effort * 0.6,
					},
					{
						Type:           "test",
						Description:    "Add unit tests for new extracted methods",
						EstimatedHours: effort * 0.4,
					},
				},
				Benefits: []string{
					"Improved code readability and maintainability",
					"Easier debugging and testing",
					"Reduced risk of bugs",
				},
				Risks: []string{
					"Potential introduction of bugs during refactoring",
					"Temporary code instability",
				},
				Dependencies: []string{},
				Timeline:     qr.estimateTimeline(effort),
			})
			id++
		}
	}

	return recommendations
}

// generateDuplicationRecommendations creates recommendations for code duplication
func (qr *QualityReporter) generateDuplicationRecommendations(duplication *DuplicationMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	// Process exact duplicates first (highest priority)
	for _, duplicate := range duplication.ExactDuplicates {
		if len(duplicate.Instances) > 1 && duplicate.LineCount > 10 {
			effort := qr.estimateDuplicationFixEffort(duplicate.LineCount, len(duplicate.Instances))

			recommendations = append(recommendations, QualityRecommendation{
				ID:          fmt.Sprintf("DUP-%d", id),
				Title:       fmt.Sprintf("Consolidate duplicated code: %d lines", duplicate.LineCount),
				Description: fmt.Sprintf("Found %d instances of duplicated code (%d lines each)", len(duplicate.Instances), duplicate.LineCount),
				Category:    CategoryQuickWins,
				Priority:    qr.determinePriority(float64(duplicate.LineCount), 10, 50),
				Impact:      qr.determineImpact(float64(duplicate.LineCount*len(duplicate.Instances)), 100),
				Effort:      qr.determineEffortLevel(effort),
				EffortHours: effort,
				ROI:         qr.calculateROI(effort, float64(duplicate.LineCount*len(duplicate.Instances))),
				Component:   "duplication",
				Files:       qr.extractFilesFromInstances(duplicate.Instances),
				Actions: []RecommendationAction{
					{
						Type:           "extract",
						Description:    "Extract common functionality into shared utility",
						EstimatedHours: effort * 0.7,
					},
					{
						Type:           "refactor",
						Description:    "Update all locations to use extracted utility",
						EstimatedHours: effort * 0.3,
					},
				},
				Benefits: []string{
					"Reduced maintenance burden",
					"Consistent behavior across codebase",
					"Improved code reusability",
				},
				Risks: []string{
					"Risk of breaking existing functionality",
					"Need for comprehensive testing",
				},
				Dependencies: []string{},
				Timeline:     qr.estimateTimeline(effort),
			})
			id++
		}
	}

	return recommendations
}

// generateDebtRecommendations creates technical debt reduction recommendations
func (qr *QualityReporter) generateDebtRecommendations(debt *TechnicalDebtMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	// Sort files by debt score for prioritization
	type fileDebt struct {
		filename string
		score    float64
		hours    float64
	}

	var fileDebts []fileDebt
	for filename, fileMetrics := range debt.FileDebtScores {
		fileDebts = append(fileDebts, fileDebt{
			filename: filename,
			score:    fileMetrics.OverallScore,
			hours:    fileMetrics.DebtHours,
		})
	}

	sort.Slice(fileDebts, func(i, j int) bool {
		return fileDebts[i].score < fileDebts[j].score // Lower score = more debt
	})

	// Generate recommendations for high-debt files
	for i, fd := range fileDebts {
		if i >= 10 || fd.score > 60 { // Top 10 or score > 60
			break
		}

		category := qr.categorizeDebtByScore(fd.score)

		recommendations = append(recommendations, QualityRecommendation{
			ID:          fmt.Sprintf("DEBT-%d", id),
			Title:       fmt.Sprintf("Reduce technical debt in %s", fd.filename),
			Description: fmt.Sprintf("File has technical debt score of %.1f (estimated %.1f hours)", fd.score, fd.hours),
			Category:    category,
			Priority:    qr.determinePriority(100-fd.score, 40, 70),
			Impact:      qr.determineImpact(100-fd.score, 40),
			Effort:      qr.determineEffortLevel(fd.hours),
			EffortHours: fd.hours,
			ROI:         qr.calculateROI(fd.hours, 100-fd.score),
			Component:   "technical_debt",
			Files:       []string{fd.filename},
			Actions: []RecommendationAction{
				{
					Type:           "refactor",
					Description:    "Address code smells and anti-patterns",
					EstimatedHours: fd.hours * 0.6,
				},
				{
					Type:           "test",
					Description:    "Add comprehensive tests",
					EstimatedHours: fd.hours * 0.4,
				},
			},
			Benefits: []string{
				"Improved code maintainability",
				"Reduced development time for future changes",
				"Lower risk of bugs",
			},
			Risks: []string{
				"Potential regression if not properly tested",
				"Short-term development slowdown",
			},
			Dependencies: []string{},
			Timeline:     qr.estimateTimeline(fd.hours),
		})
		id++
	}

	return recommendations
}

// generateCoverageRecommendations creates test coverage recommendations
func (qr *QualityReporter) generateCoverageRecommendations(coverage *CoverageMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	// Generate recommendations for untested functions
	for _, funcAnalysis := range coverage.FunctionAnalysis {
		if funcAnalysis.TestabilityScore < 70 && funcAnalysis.RiskLevel == "high" {
			effort := qr.estimateTestingEffort(funcAnalysis.TestingDifficulty, len(funcAnalysis.RequiredMocks))

			recommendations = append(recommendations, QualityRecommendation{
				ID:          fmt.Sprintf("TEST-%d", id),
				Title:       fmt.Sprintf("Add tests for %s", funcAnalysis.Name),
				Description: fmt.Sprintf("High-priority function with low testability score (%.1f)", funcAnalysis.TestabilityScore),
				Category:    CategoryStrategicImprovements,
				Priority:    PriorityHigh,
				Impact:      ImpactHigh,
				Effort:      qr.determineEffortLevel(effort),
				EffortHours: effort,
				ROI:         qr.calculateTestingROI(effort, funcAnalysis.TestabilityScore),
				Component:   "coverage",
				Files:       []string{funcAnalysis.FilePath},
				Actions: []RecommendationAction{
					{
						Type:           "test",
						Description:    "Create comprehensive unit tests",
						EstimatedHours: effort * 0.8,
					},
					{
						Type:           "mock",
						Description:    "Set up required mocks and test doubles",
						EstimatedHours: effort * 0.2,
					},
				},
				Benefits: []string{
					"Improved code reliability",
					"Easier refactoring with confidence",
					"Better regression detection",
				},
				Risks: []string{
					"Initial time investment",
					"Maintenance overhead for tests",
				},
				Dependencies: []string{},
				Timeline:     qr.estimateTimeline(effort),
			})
			id++
		}
	}

	return recommendations
}

// generatePerformanceRecommendations creates performance optimization recommendations
func (qr *QualityReporter) generatePerformanceRecommendations(performance *PerformanceMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	for _, antiPattern := range performance.AntiPatterns {
		if antiPattern.Severity == "high" || antiPattern.Severity == "critical" {
			effort := qr.estimatePerformanceFixEffort(antiPattern.Type, antiPattern.Impact.Score)
			category := qr.categorizePerformanceIssue(antiPattern.Severity)

			recommendations = append(recommendations, QualityRecommendation{
				ID:          fmt.Sprintf("PERF-%d", id),
				Title:       fmt.Sprintf("Fix %s anti-pattern", antiPattern.Type),
				Description: antiPattern.Description,
				Category:    category,
				Priority:    qr.mapSeverityToPriority(antiPattern.Severity),
				Impact:      qr.determineImpact(antiPattern.Impact.Score, 70),
				Effort:      qr.determineEffortLevel(effort),
				EffortHours: effort,
				ROI:         qr.calculatePerformanceROI(effort, antiPattern.Impact.Score),
				Component:   "performance",
				Files:       []string{antiPattern.FilePath},
				Actions: []RecommendationAction{
					{
						Type:           "optimize",
						Description:    fmt.Sprintf("Fix %s performance anti-pattern: %s", antiPattern.Type, antiPattern.Impact.Description),
						EstimatedHours: effort,
					},
				},
				Benefits: []string{
					"Improved application performance",
					"Better user experience",
					"Reduced resource consumption",
				},
				Risks: []string{
					"Potential complexity increase",
					"Need for performance testing",
				},
				Dependencies: []string{},
				Timeline:     qr.estimateTimeline(effort),
			})
			id++
		}
	}

	return recommendations
}

// generateMaintainabilityRecommendations creates maintainability improvement recommendations
func (qr *QualityReporter) generateMaintainabilityRecommendations(maintainability *MaintainabilityMetrics) []QualityRecommendation {
	var recommendations []QualityRecommendation
	id := 1

	// Sort files by maintainability index (lowest first)
	type fileMaintainability struct {
		filename string
		index    float64
	}

	var fileMaintainabilities []fileMaintainability
	for filename, fileMetrics := range maintainability.FileMetrics {
		fileMaintainabilities = append(fileMaintainabilities, fileMaintainability{
			filename: filename,
			index:    fileMetrics.OverallIndex,
		})
	}

	sort.Slice(fileMaintainabilities, func(i, j int) bool {
		return fileMaintainabilities[i].index < fileMaintainabilities[j].index
	})

	// Generate recommendations for low-maintainability files
	for i, fm := range fileMaintainabilities {
		if i >= 5 || fm.index > 70 { // Top 5 or index > 70
			break
		}

		effort := qr.estimateMaintainabilityImprovement(fm.index)

		recommendations = append(recommendations, QualityRecommendation{
			ID:          fmt.Sprintf("MAINT-%d", id),
			Title:       fmt.Sprintf("Improve maintainability of %s", fm.filename),
			Description: fmt.Sprintf("Low maintainability index: %.1f", fm.index),
			Category:    CategoryLongTermGoals,
			Priority:    qr.determinePriority(100-fm.index, 30, 50),
			Impact:      ImpactMedium,
			Effort:      qr.determineEffortLevel(effort),
			EffortHours: effort,
			ROI:         qr.calculateMaintenanceROI(effort, fm.index),
			Component:   "maintainability",
			Files:       []string{fm.filename},
			Actions: []RecommendationAction{
				{
					Type:           "refactor",
					Description:    "Improve code structure and documentation",
					EstimatedHours: effort * 0.7,
				},
				{
					Type:           "document",
					Description:    "Add comprehensive documentation",
					EstimatedHours: effort * 0.3,
				},
			},
			Benefits: []string{
				"Easier maintenance and modifications",
				"Improved developer productivity",
				"Better code understanding",
			},
			Risks: []string{
				"Initial time investment",
				"Potential for introducing bugs",
			},
			Dependencies: []string{},
			Timeline:     qr.estimateTimeline(effort),
		})
		id++
	}

	return recommendations
}

// rankAndLimitRecommendations sorts recommendations by priority and limits the count
func (qr *QualityReporter) rankAndLimitRecommendations(recommendations []QualityRecommendation) []QualityRecommendation {
	// Sort by ROI (descending), then by Impact, then by Priority
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].ROI != recommendations[j].ROI {
			return recommendations[i].ROI > recommendations[j].ROI
		}

		if recommendations[i].Impact != recommendations[j].Impact {
			return qr.impactToScore(recommendations[i].Impact) > qr.impactToScore(recommendations[j].Impact)
		}

		return qr.priorityToScore(recommendations[i].Priority) > qr.priorityToScore(recommendations[j].Priority)
	})

	// Limit to max recommendations
	if len(recommendations) > qr.config.MaxRecommendations {
		recommendations = recommendations[:qr.config.MaxRecommendations]
	}

	return recommendations
} // Helper methods for categorization and estimation

// categorizeByComplexity categorizes recommendations based on complexity levels
func (qr *QualityReporter) categorizeByComplexity(complexity int) RecommendationCategory {
	switch {
	case complexity > 25:
		return CategoryCriticalFixes
	case complexity > 20:
		return CategoryStrategicImprovements
	case complexity > 15:
		return CategoryQuickWins
	default:
		return CategoryLongTermGoals
	}
}

// categorizeDebtByScore categorizes technical debt recommendations
func (qr *QualityReporter) categorizeDebtByScore(score float64) RecommendationCategory {
	switch {
	case score < 30:
		return CategoryCriticalFixes
	case score < 50:
		return CategoryStrategicImprovements
	case score < 70:
		return CategoryQuickWins
	default:
		return CategoryLongTermGoals
	}
}

// categorizePerformanceIssue categorizes performance issues
func (qr *QualityReporter) categorizePerformanceIssue(severity string) RecommendationCategory {
	switch severity {
	case "critical":
		return CategoryCriticalFixes
	case "high":
		return CategoryStrategicImprovements
	case "medium":
		return CategoryQuickWins
	default:
		return CategoryLongTermGoals
	}
}

// Effort estimation methods

// estimateRefactoringEffort estimates hours needed to refactor complex code
func (qr *QualityReporter) estimateRefactoringEffort(complexity int, linesOfCode int) float64 {
	baseHours := float64(complexity-10) * 0.5                   // Base time for complexity reduction
	sizeMultiplier := math.Max(1.0, float64(linesOfCode)/100.0) // Size adjustment

	effort := baseHours * sizeMultiplier

	// Apply bounds
	if effort < 2.0 {
		effort = 2.0
	}
	if effort > 40.0 {
		effort = 40.0
	}

	return math.Round(effort*4) / 4 // Round to nearest 0.25
}

// estimateDuplicationFixEffort estimates hours to fix code duplication
func (qr *QualityReporter) estimateDuplicationFixEffort(lines int, instances int) float64 {
	baseHours := float64(lines) * 0.1                           // Base time per line
	instanceMultiplier := math.Max(1.0, float64(instances)*0.3) // More instances = more complexity

	effort := baseHours * instanceMultiplier

	// Apply bounds
	if effort < 1.0 {
		effort = 1.0
	}
	if effort > 24.0 {
		effort = 24.0
	}

	return math.Round(effort*4) / 4
}

// estimateTestingEffort estimates hours needed for testing
func (qr *QualityReporter) estimateTestingEffort(complexity string, mocksRequired int) float64 {
	var baseHours float64
	switch complexity {
	case "high":
		baseHours = 8.0
	case "medium":
		baseHours = 4.0
	case "low":
		baseHours = 2.0
	default:
		baseHours = 3.0
	}

	mockMultiplier := 1.0 + (float64(mocksRequired) * 0.5)
	effort := baseHours * mockMultiplier

	// Apply bounds
	if effort > 32.0 {
		effort = 32.0
	}

	return math.Round(effort*4) / 4
}

// estimatePerformanceFixEffort estimates hours to fix performance issues
func (qr *QualityReporter) estimatePerformanceFixEffort(antiPatternType string, impact float64) float64 {
	var baseHours float64

	switch antiPatternType {
	case "nested_loops", "sync_in_async":
		baseHours = 6.0
	case "n_plus_one_query", "memory_leak":
		baseHours = 8.0
	case "large_bundle", "unnecessary_rerenders":
		baseHours = 4.0
	default:
		baseHours = 5.0
	}

	impactMultiplier := 1.0 + (impact / 100.0)
	effort := baseHours * impactMultiplier

	if effort > 24.0 {
		effort = 24.0
	}

	return math.Round(effort*4) / 4
}

// estimateMaintainabilityImprovement estimates hours to improve maintainability
func (qr *QualityReporter) estimateMaintainabilityImprovement(index float64) float64 {
	improvementNeeded := 80.0 - index // Target maintainability index of 80
	baseHours := improvementNeeded * 0.3

	if baseHours < 2.0 {
		baseHours = 2.0
	}
	if baseHours > 20.0 {
		baseHours = 20.0
	}

	return math.Round(baseHours*4) / 4
}

// Priority and impact determination methods

// determinePriority determines priority based on value and thresholds
func (qr *QualityReporter) determinePriority(value, lowThreshold, highThreshold float64) Priority {
	switch {
	case value >= highThreshold:
		return PriorityCritical
	case value >= lowThreshold:
		return PriorityHigh
	case value >= lowThreshold*0.5:
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// determineImpact determines impact level based on value and threshold
func (qr *QualityReporter) determineImpact(value, threshold float64) ImpactLevel {
	ratio := value / threshold
	switch {
	case ratio >= 2.0:
		return ImpactHigh
	case ratio >= 1.0:
		return ImpactMedium
	default:
		return ImpactLow
	}
}

// determineEffortLevel determines effort level based on hours
func (qr *QualityReporter) determineEffortLevel(hours float64) EffortLevel {
	switch {
	case hours >= 16.0:
		return EffortHigh
	case hours >= 4.0:
		return EffortMedium
	default:
		return EffortLow
	}
}

// mapSeverityToPriority maps severity strings to priority levels
func (qr *QualityReporter) mapSeverityToPriority(severity string) Priority {
	switch severity {
	case "critical":
		return PriorityCritical
	case "high":
		return PriorityHigh
	case "medium":
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// ROI calculation methods

// calculateROI calculates return on investment for recommendations
func (qr *QualityReporter) calculateROI(effort, benefit float64) float64 {
	if effort <= 0 {
		return 0
	}

	// Simple ROI calculation: benefit/cost ratio
	roi := benefit / effort
	return math.Round(roi*100) / 100
}

// calculateTestingROI calculates ROI for testing recommendations
func (qr *QualityReporter) calculateTestingROI(effort, testabilityScore float64) float64 {
	benefit := (100.0 - testabilityScore) * 2.0 // Higher benefit for lower testability
	return qr.calculateROI(effort, benefit)
}

// calculatePerformanceROI calculates ROI for performance improvements
func (qr *QualityReporter) calculatePerformanceROI(effort, impact float64) float64 {
	benefit := impact * 1.5 // Performance improvements have high business value
	return qr.calculateROI(effort, benefit)
}

// calculateMaintenanceROI calculates ROI for maintainability improvements
func (qr *QualityReporter) calculateMaintenanceROI(effort, maintainabilityIndex float64) float64 {
	benefit := (100.0 - maintainabilityIndex) * 1.2
	return qr.calculateROI(effort, benefit)
}

// Utility methods

// impactToScore converts impact level to numeric score for sorting
func (qr *QualityReporter) impactToScore(impact ImpactLevel) float64 {
	switch impact {
	case ImpactHigh:
		return 3.0
	case ImpactMedium:
		return 2.0
	case ImpactLow:
		return 1.0
	default:
		return 0.0
	}
}

// priorityToScore converts priority level to numeric score for sorting
func (qr *QualityReporter) priorityToScore(priority Priority) float64 {
	switch priority {
	case PriorityCritical:
		return 4.0
	case PriorityHigh:
		return 3.0
	case PriorityMedium:
		return 2.0
	case PriorityLow:
		return 1.0
	default:
		return 0.0
	}
}

// estimateTimeline estimates timeline based on effort hours
func (qr *QualityReporter) estimateTimeline(hours float64) string {
	switch {
	case hours <= 4:
		return "1-2 days"
	case hours <= 16:
		return "3-5 days"
	case hours <= 40:
		return "1-2 weeks"
	default:
		return "2-4 weeks"
	}
}

// extractFilesFromInstances extracts unique file paths from duplication instances
func (qr *QualityReporter) extractFilesFromInstances(instances []DuplicationInstance) []string {
	fileMap := make(map[string]bool)
	for _, instance := range instances {
		fileMap[instance.FilePath] = true
	}

	var files []string
	for filename := range fileMap {
		files = append(files, filename)
	}

	sort.Strings(files)
	return files
}

// generateRoadmap creates a quality improvement roadmap with milestones
func (qr *QualityReporter) generateRoadmap(recommendations []QualityRecommendation, scores ComponentScores) QualityRoadmap {
	timeframeWeeks := qr.config.RoadmapTimeframe

	// Group recommendations into phases
	phases := qr.createImprovementPhases(recommendations)

	// Create milestones
	milestones := qr.createMilestones(phases, timeframeWeeks)

	// Calculate resource planning
	resourcePlan := qr.calculateResourcePlan(recommendations)

	// Generate risk assessment
	risks := qr.generateRoadmapRisks()

	// Define success metrics
	successMetrics := qr.defineSuccessMetrics(scores)

	return QualityRoadmap{
		Overview:       qr.generateRoadmapOverview(recommendations, phases),
		TimeframeWeeks: timeframeWeeks,
		Milestones:     milestones,
		Phases:         phases,
		ResourcePlan:   resourcePlan,
		RiskAssessment: risks,
		SuccessMetrics: successMetrics,
	}
}

// createImprovementPhases groups recommendations into logical phases
func (qr *QualityReporter) createImprovementPhases(recommendations []QualityRecommendation) []ImprovementPhase {
	// Group recommendations by category
	categoryMap := make(map[RecommendationCategory][]QualityRecommendation)
	for _, rec := range recommendations {
		categoryMap[rec.Category] = append(categoryMap[rec.Category], rec)
	}

	var phases []ImprovementPhase

	// Phase 1: Quick Wins (weeks 1-2)
	if quickWins := categoryMap[CategoryQuickWins]; len(quickWins) > 0 {
		phases = append(phases, qr.createPhase("Quick Wins", "2 weeks", quickWins, map[string]float64{
			"duplication": 15.0,
			"complexity":  10.0,
		}))
	}

	// Phase 2: Critical Fixes (weeks 3-6)
	if criticalFixes := categoryMap[CategoryCriticalFixes]; len(criticalFixes) > 0 {
		phases = append(phases, qr.createPhase("Critical Fixes", "4 weeks", criticalFixes, map[string]float64{
			"technical_debt": 25.0,
			"performance":    20.0,
			"complexity":     15.0,
		}))
	}

	// Phase 3: Strategic Improvements (weeks 7-10)
	if strategic := categoryMap[CategoryStrategicImprovements]; len(strategic) > 0 {
		phases = append(phases, qr.createPhase("Strategic Improvements", "4 weeks", strategic, map[string]float64{
			"coverage":        20.0,
			"maintainability": 15.0,
			"technical_debt":  10.0,
		}))
	}

	// Phase 4: Long-term Goals (weeks 11-12)
	if longTerm := categoryMap[CategoryLongTermGoals]; len(longTerm) > 0 {
		phases = append(phases, qr.createPhase("Long-term Goals", "2 weeks", longTerm, map[string]float64{
			"maintainability": 10.0,
			"coverage":        5.0,
		}))
	}

	return phases
}

// createPhase creates an improvement phase from recommendations
func (qr *QualityReporter) createPhase(name, duration string, recommendations []QualityRecommendation, expectedImpact map[string]float64) ImprovementPhase {
	var focus []string
	var recommendationIds []string
	var totalHours, devHours, qaHours, reviewHours float64
	specialistsMap := make(map[string]bool)

	for _, rec := range recommendations {
		focus = append(focus, rec.Component)
		recommendationIds = append(recommendationIds, rec.ID)
		totalHours += rec.EffortHours

		// Distribute hours (rough estimation)
		devHours += rec.EffortHours * 0.7
		qaHours += rec.EffortHours * 0.2
		reviewHours += rec.EffortHours * 0.1

		// Determine specialists needed
		switch rec.Component {
		case "performance":
			specialistsMap["Performance Engineer"] = true
		case "technical_debt":
			specialistsMap["Senior Developer"] = true
		case "coverage":
			specialistsMap["QA Engineer"] = true
		}
	}

	// Remove duplicates from focus
	focusMap := make(map[string]bool)
	var uniqueFocus []string
	for _, f := range focus {
		if !focusMap[f] {
			focusMap[f] = true
			uniqueFocus = append(uniqueFocus, f)
		}
	}

	var specialists []string
	for specialist := range specialistsMap {
		specialists = append(specialists, specialist)
	}

	return ImprovementPhase{
		Name:            name,
		Duration:        duration,
		Focus:           uniqueFocus,
		Recommendations: recommendationIds,
		ExpectedImpact:  expectedImpact,
		ResourceNeeds: PhaseResourceNeeds{
			DeveloperHours: devHours,
			QAHours:        qaHours,
			ReviewHours:    reviewHours,
			Specialists:    specialists,
		},
	}
} // createMilestones creates quality improvement milestones
func (qr *QualityReporter) createMilestones(phases []ImprovementPhase, timeframeWeeks int) []QualityMilestone {
	var milestones []QualityMilestone
	currentWeek := 0
	now := time.Now()

	for i, phase := range phases {
		var duration int
		switch phase.Duration {
		case "2 weeks":
			duration = 2
		case "4 weeks":
			duration = 4
		default:
			duration = 3
		}

		currentWeek += duration
		targetDate := now.AddDate(0, 0, currentWeek*7)

		milestone := QualityMilestone{
			Name:            fmt.Sprintf("Milestone %d: %s Complete", i+1, phase.Name),
			TargetDate:      targetDate,
			Description:     fmt.Sprintf("Complete all %s initiatives", phase.Name),
			Goals:           qr.createMilestoneGoals(phase),
			Deliverables:    qr.createMilestoneDeliverables(phase),
			SuccessCriteria: qr.createSuccessCriteria(phase),
			Dependencies:    qr.createMilestoneDependencies(i, phases),
			EstimatedHours:  phase.ResourceNeeds.DeveloperHours + phase.ResourceNeeds.QAHours + phase.ResourceNeeds.ReviewHours,
		}

		milestones = append(milestones, milestone)
	}

	return milestones
}

// createMilestoneGoals creates goals for a milestone
func (qr *QualityReporter) createMilestoneGoals(phase ImprovementPhase) []string {
	var goals []string

	for component, improvement := range phase.ExpectedImpact {
		goals = append(goals, fmt.Sprintf("Improve %s score by %.1f points", component, improvement))
	}

	return goals
}

// createMilestoneDeliverables creates deliverables for a milestone
func (qr *QualityReporter) createMilestoneDeliverables(phase ImprovementPhase) []string {
	var deliverables []string

	switch phase.Name {
	case "Quick Wins":
		deliverables = []string{
			"Consolidated duplicate code",
			"Refactored high-complexity functions",
			"Updated documentation",
		}
	case "Critical Fixes":
		deliverables = []string{
			"Fixed critical performance issues",
			"Reduced technical debt in priority files",
			"Implemented security improvements",
		}
	case "Strategic Improvements":
		deliverables = []string{
			"Increased test coverage for critical components",
			"Improved system architecture",
			"Enhanced code maintainability",
		}
	case "Long-term Goals":
		deliverables = []string{
			"Comprehensive documentation update",
			"Final code quality improvements",
			"Quality gate implementation",
		}
	}

	return deliverables
}

// createSuccessCriteria creates success criteria for a milestone
func (qr *QualityReporter) createSuccessCriteria(phase ImprovementPhase) []string {
	var criteria []string

	for component, improvement := range phase.ExpectedImpact {
		criteria = append(criteria, fmt.Sprintf("%s improvement of %.1f%% achieved", component, improvement))
	}

	criteria = append(criteria, "All phase recommendations completed")
	criteria = append(criteria, "No regression in existing quality metrics")
	criteria = append(criteria, "Code review approval obtained")

	return criteria
}

// createMilestoneDependencies creates dependencies between milestones
func (qr *QualityReporter) createMilestoneDependencies(phaseIndex int, phases []ImprovementPhase) []string {
	var dependencies []string

	if phaseIndex > 0 {
		dependencies = append(dependencies, fmt.Sprintf("Completion of %s phase", phases[phaseIndex-1].Name))
	}

	// Add specific dependencies based on phase type
	switch phases[phaseIndex].Name {
	case "Critical Fixes":
		dependencies = append(dependencies, "Team availability for intensive refactoring")
	case "Strategic Improvements":
		dependencies = append(dependencies, "Infrastructure setup for testing")
	case "Long-term Goals":
		dependencies = append(dependencies, "Stakeholder review and approval")
	}

	return dependencies
}

// calculateResourcePlan calculates overall resource requirements
func (qr *QualityReporter) calculateResourcePlan(recommendations []QualityRecommendation) ResourcePlan {
	var totalDeveloperHours, totalQAHours, totalReviewHours float64
	skillsMap := make(map[string]bool)

	for _, rec := range recommendations {
		totalDeveloperHours += rec.EffortHours * 0.7
		totalQAHours += rec.EffortHours * 0.2
		totalReviewHours += rec.EffortHours * 0.1

		// Determine required skills
		switch rec.Component {
		case "complexity":
			skillsMap["Refactoring"] = true
			skillsMap["Code Architecture"] = true
		case "performance":
			skillsMap["Performance Optimization"] = true
			skillsMap["Profiling"] = true
		case "coverage":
			skillsMap["Test Automation"] = true
			skillsMap["Mock Development"] = true
		case "technical_debt":
			skillsMap["Legacy Code Maintenance"] = true
		}
	}

	var skillsNeeded []string
	for skill := range skillsMap {
		skillsNeeded = append(skillsNeeded, skill)
	}

	// Estimate team size (assuming 40 hours per week per person)
	totalHours := totalDeveloperHours + totalQAHours + totalReviewHours
	teamSize := int(math.Ceil(totalHours / (40.0 * float64(qr.config.RoadmapTimeframe))))
	if teamSize < 2 {
		teamSize = 2 // Minimum team size
	}

	// Estimate cost (rough calculation at $100/hour average)
	estimatedCost := totalHours * 100.0

	// Calculate duration
	duration := fmt.Sprintf("%d weeks", qr.config.RoadmapTimeframe)

	return ResourcePlan{
		TotalDeveloperHours: totalDeveloperHours,
		TotalQAHours:        totalQAHours,
		TotalReviewHours:    totalReviewHours,
		EstimatedCost:       estimatedCost,
		TeamSize:            teamSize,
		Duration:            duration,
		SkillsNeeded:        skillsNeeded,
	}
}

// generateRoadmapRisks generates risk assessment for the improvement roadmap
func (qr *QualityReporter) generateRoadmapRisks() []RoadmapRisk {
	return []RoadmapRisk{
		{
			Description: "Refactoring introduces new bugs",
			Probability: "medium",
			Impact:      "high",
			Mitigation:  "Comprehensive testing and gradual rollout",
			Owner:       "Development Team",
		},
		{
			Description: "Resource availability constraints",
			Probability: "high",
			Impact:      "medium",
			Mitigation:  "Flexible timeline and cross-training",
			Owner:       "Project Manager",
		},
		{
			Description: "Scope creep during improvements",
			Probability: "medium",
			Impact:      "medium",
			Mitigation:  "Clear requirements and change control process",
			Owner:       "Technical Lead",
		},
		{
			Description: "Performance regression during refactoring",
			Probability: "low",
			Impact:      "high",
			Mitigation:  "Performance monitoring and benchmarking",
			Owner:       "Performance Engineer",
		},
		{
			Description: "Stakeholder resistance to changes",
			Probability: "low",
			Impact:      "medium",
			Mitigation:  "Clear communication and benefits demonstration",
			Owner:       "Product Manager",
		},
	}
}

// defineSuccessMetrics defines success criteria for quality improvements
func (qr *QualityReporter) defineSuccessMetrics(scores ComponentScores) []SuccessMetric {
	return []SuccessMetric{
		{
			Name:        "Overall Quality Score",
			Current:     qr.calculateOverallScore(scores),
			Target:      85.0,
			Unit:        "score",
			Timeline:    fmt.Sprintf("%d weeks", qr.config.RoadmapTimeframe),
			Measurement: "Automated quality analysis",
		},
		{
			Name:        "Code Complexity",
			Current:     scores.Complexity,
			Target:      80.0,
			Unit:        "score",
			Timeline:    "6 weeks",
			Measurement: "Cyclomatic complexity analysis",
		},
		{
			Name:        "Technical Debt",
			Current:     scores.TechnicalDebt,
			Target:      75.0,
			Unit:        "score",
			Timeline:    "8 weeks",
			Measurement: "Technical debt scoring algorithm",
		},
		{
			Name:        "Test Coverage",
			Current:     scores.Coverage,
			Target:      80.0,
			Unit:        "percentage",
			Timeline:    "10 weeks",
			Measurement: "Test coverage analysis",
		},
		{
			Name:        "Performance Score",
			Current:     scores.Performance,
			Target:      85.0,
			Unit:        "score",
			Timeline:    "6 weeks",
			Measurement: "Performance anti-pattern detection",
		},
		{
			Name:        "Code Duplication",
			Current:     scores.Duplication,
			Target:      90.0,
			Unit:        "score",
			Timeline:    "4 weeks",
			Measurement: "Duplication detection algorithm",
		},
	}
}

// generateRoadmapOverview creates an overview description for the roadmap
func (qr *QualityReporter) generateRoadmapOverview(recommendations []QualityRecommendation, phases []ImprovementPhase) string {
	totalRecommendations := len(recommendations)
	totalPhases := len(phases)

	var totalHours float64
	for _, rec := range recommendations {
		totalHours += rec.EffortHours
	}

	return fmt.Sprintf(
		"Comprehensive quality improvement plan addressing %d recommendations across %d phases. "+
			"Estimated effort: %.1f hours over %d weeks. "+
			"Focus areas include complexity reduction, technical debt remediation, and test coverage improvement.",
		totalRecommendations, totalPhases, totalHours, qr.config.RoadmapTimeframe,
	)
}

// generateExecutiveSummary creates executive summary for stakeholders
func (qr *QualityReporter) generateExecutiveSummary(
	overallScore float64,
	qualityGrade string,
	scores ComponentScores,
	recommendations []QualityRecommendation,
) *ExecutiveSummary {
	// Key findings
	keyFindings := qr.generateKeyFindings(scores, recommendations)

	// Critical issues
	criticalIssues := qr.identifyCriticalIssues(scores, recommendations)

	// Top recommendations
	topRecommendations := qr.selectTopRecommendations(recommendations, 5)

	// Business impact
	businessImpact := qr.assessBusinessImpact(scores, recommendations)

	// Investment summary
	investment := qr.calculateInvestmentSummary(recommendations)

	// Expected outcomes
	outcomes := qr.defineExpectedOutcomes(scores)

	// Next steps
	nextSteps := qr.defineNextSteps(recommendations)

	// Overall assessment
	assessment := qr.generateOverallAssessment(overallScore, qualityGrade, scores)

	return &ExecutiveSummary{
		OverallAssessment:  assessment,
		KeyFindings:        keyFindings,
		CriticalIssues:     criticalIssues,
		TopRecommendations: topRecommendations,
		BusinessImpact:     businessImpact,
		InvestmentRequired: investment,
		ExpectedOutcomes:   outcomes,
		NextSteps:          nextSteps,
	}
}

// generateKeyFindings identifies key findings from the analysis
func (qr *QualityReporter) generateKeyFindings(scores ComponentScores, recommendations []QualityRecommendation) []string {
	var findings []string

	// Analyze scores for insights
	if scores.TechnicalDebt < 60 {
		findings = append(findings, fmt.Sprintf("High technical debt detected (score: %.1f) requiring immediate attention", scores.TechnicalDebt))
	}

	if scores.Complexity < 70 {
		findings = append(findings, fmt.Sprintf("Code complexity concerns identified (score: %.1f) affecting maintainability", scores.Complexity))
	}

	if scores.Coverage < 70 {
		findings = append(findings, fmt.Sprintf("Test coverage gaps detected (score: %.1f) increasing quality risk", scores.Coverage))
	}

	if scores.Performance < 75 {
		findings = append(findings, fmt.Sprintf("Performance issues identified (score: %.1f) impacting user experience", scores.Performance))
	}

	// Analyze recommendation patterns
	categoryCount := make(map[RecommendationCategory]int)
	for _, rec := range recommendations {
		categoryCount[rec.Category]++
	}

	if categoryCount[CategoryCriticalFixes] > 3 {
		findings = append(findings, fmt.Sprintf("%d critical fixes required for system stability", categoryCount[CategoryCriticalFixes]))
	}

	if categoryCount[CategoryQuickWins] > 5 {
		findings = append(findings, fmt.Sprintf("%d quick wins available for immediate quality improvements", categoryCount[CategoryQuickWins]))
	}

	return findings
}

// identifyCriticalIssues identifies critical issues requiring immediate attention
func (qr *QualityReporter) identifyCriticalIssues(scores ComponentScores, recommendations []QualityRecommendation) []string {
	var issues []string

	// Score-based critical issues
	if scores.TechnicalDebt < 40 {
		issues = append(issues, "Critical technical debt levels may impact delivery velocity")
	}

	if scores.Performance < 50 {
		issues = append(issues, "Severe performance issues affecting user experience")
	}

	if scores.Complexity > 85 { // High complexity score means low complexity (inverse)
		// This is good, no issue
	} else if scores.Complexity < 50 {
		issues = append(issues, "Extremely high code complexity hindering maintenance")
	}

	// Recommendation-based critical issues
	criticalCount := 0
	for _, rec := range recommendations {
		if rec.Priority == PriorityCritical {
			criticalCount++
		}
	}

	if criticalCount > 5 {
		issues = append(issues, fmt.Sprintf("%d critical-priority recommendations require immediate action", criticalCount))
	}

	return issues
}

// selectTopRecommendations selects the top N recommendations by ROI and impact
func (qr *QualityReporter) selectTopRecommendations(recommendations []QualityRecommendation, count int) []string {
	if len(recommendations) > count {
		recommendations = recommendations[:count] // Already sorted by ROI in rankAndLimitRecommendations
	}

	var topRecs []string
	for _, rec := range recommendations {
		topRecs = append(topRecs, fmt.Sprintf("%s (ROI: %.1fx, %s impact)", rec.Title, rec.ROI, rec.Impact))
	}

	return topRecs
} // assessBusinessImpact evaluates business implications of quality issues
func (qr *QualityReporter) assessBusinessImpact(scores ComponentScores, recommendations []QualityRecommendation) BusinessImpact {
	overallScore := qr.calculateOverallScore(scores)

	// Determine risk level
	var riskLevel string
	switch {
	case overallScore >= 80:
		riskLevel = "Low"
	case overallScore >= 60:
		riskLevel = "Medium"
	case overallScore >= 40:
		riskLevel = "High"
	default:
		riskLevel = "Critical"
	}

	// Calculate maintenance cost impact
	maintenanceCost := qr.calculateMaintenanceCostImpact(scores)

	// Assess development velocity impact
	var velocityImpact string
	if scores.Complexity < 60 || scores.TechnicalDebt < 50 {
		velocityImpact = "Significantly reduced due to code complexity and technical debt"
	} else if scores.Complexity < 75 || scores.TechnicalDebt < 70 {
		velocityImpact = "Moderately impacted by quality issues"
	} else {
		velocityImpact = "Minimal impact on development speed"
	}

	// Calculate technical debt cost
	var totalDebtHours float64
	for _, rec := range recommendations {
		if rec.Category == CategoryTechnicalDebtReduction {
			totalDebtHours += rec.EffortHours
		}
	}
	technicalDebtCost := totalDebtHours * 100.0 // $100/hour estimate

	// Assess quality risk
	var qualityRisk string
	if scores.Coverage < 60 || scores.Performance < 60 {
		qualityRisk = "High risk of production issues and customer impact"
	} else if scores.Coverage < 75 || scores.Performance < 75 {
		qualityRisk = "Medium risk of quality issues"
	} else {
		qualityRisk = "Low risk with current quality practices"
	}

	// Assess customer impact
	var customerImpact string
	if scores.Performance < 60 {
		customerImpact = "Poor performance likely affecting user satisfaction and retention"
	} else if scores.Performance < 75 {
		customerImpact = "Performance issues may impact user experience"
	} else {
		customerImpact = "Minimal direct customer impact from current quality issues"
	}

	return BusinessImpact{
		RiskLevel:           riskLevel,
		MaintenanceCost:     maintenanceCost,
		DevelopmentVelocity: velocityImpact,
		TechnicalDebtCost:   technicalDebtCost,
		QualityRisk:         qualityRisk,
		CustomerImpact:      customerImpact,
	}
}

// calculateMaintenanceCostImpact estimates maintenance cost impact
func (qr *QualityReporter) calculateMaintenanceCostImpact(scores ComponentScores) float64 {
	// Base maintenance cost multiplier based on quality scores
	complexityMultiplier := 1.0 + ((100.0 - scores.Complexity) / 100.0)
	debtMultiplier := 1.0 + ((100.0 - scores.TechnicalDebt) / 100.0)
	maintainabilityMultiplier := 1.0 + ((100.0 - scores.Maintainability) / 100.0)

	// Calculate impact as percentage increase in maintenance cost
	impactPercentage := (complexityMultiplier + debtMultiplier + maintainabilityMultiplier - 3.0) * 100.0

	return math.Round(impactPercentage*10) / 10
}

// calculateInvestmentSummary provides cost/benefit analysis
func (qr *QualityReporter) calculateInvestmentSummary(recommendations []QualityRecommendation) InvestmentSummary {
	var totalHours, totalROI float64

	for _, rec := range recommendations {
		totalHours += rec.EffortHours
		totalROI += rec.ROI
	}

	estimatedCost := totalHours * 100.0 // $100/hour average

	// Calculate expected savings (rough estimate based on maintenance cost reduction)
	expectedSavings := estimatedCost * 2.0 // Assume 2x return over time

	// Calculate payback period
	var paybackPeriod string
	if totalROI > 0 {
		avgROI := totalROI / float64(len(recommendations))
		switch {
		case avgROI >= 3.0:
			paybackPeriod = "3-6 months"
		case avgROI >= 2.0:
			paybackPeriod = "6-12 months"
		case avgROI >= 1.0:
			paybackPeriod = "12-18 months"
		default:
			paybackPeriod = "18+ months"
		}
	} else {
		paybackPeriod = "Not quantifiable"
	}

	// Calculate overall ROI
	overallROI := 0.0
	if estimatedCost > 0 {
		overallROI = (expectedSavings - estimatedCost) / estimatedCost * 100.0
	}

	return InvestmentSummary{
		TotalInvestmentHours: totalHours,
		EstimatedCost:        estimatedCost,
		ExpectedSavings:      expectedSavings,
		PaybackPeriod:        paybackPeriod,
		ROI:                  overallROI,
	}
}

// defineExpectedOutcomes describes expected results from improvements
func (qr *QualityReporter) defineExpectedOutcomes(scores ComponentScores) []ExpectedOutcome {
	var outcomes []ExpectedOutcome

	// Complexity outcomes
	if scores.Complexity < 75 {
		outcomes = append(outcomes, ExpectedOutcome{
			Area:       "Code Complexity",
			Current:    "High complexity hindering maintenance",
			Improved:   "Simplified, maintainable codebase",
			Timeline:   "6-8 weeks",
			Confidence: "High",
		})
	}

	// Technical debt outcomes
	if scores.TechnicalDebt < 70 {
		outcomes = append(outcomes, ExpectedOutcome{
			Area:       "Technical Debt",
			Current:    "Accumulating technical debt",
			Improved:   "Reduced debt with sustainable practices",
			Timeline:   "8-10 weeks",
			Confidence: "Medium",
		})
	}

	// Performance outcomes
	if scores.Performance < 80 {
		outcomes = append(outcomes, ExpectedOutcome{
			Area:       "Performance",
			Current:    "Performance bottlenecks present",
			Improved:   "Optimized performance and responsiveness",
			Timeline:   "4-6 weeks",
			Confidence: "High",
		})
	}

	// Coverage outcomes
	if scores.Coverage < 80 {
		outcomes = append(outcomes, ExpectedOutcome{
			Area:       "Test Coverage",
			Current:    "Insufficient test coverage",
			Improved:   "Comprehensive test suite with high coverage",
			Timeline:   "10-12 weeks",
			Confidence: "Medium",
		})
	}

	// Overall quality outcome
	outcomes = append(outcomes, ExpectedOutcome{
		Area:       "Overall Quality",
		Current:    fmt.Sprintf("Quality score: %.1f", qr.calculateOverallScore(scores)),
		Improved:   "Target quality score: 85+",
		Timeline:   fmt.Sprintf("%d weeks", qr.config.RoadmapTimeframe),
		Confidence: "High",
	})

	return outcomes
}

// defineNextSteps provides immediate next steps
func (qr *QualityReporter) defineNextSteps(recommendations []QualityRecommendation) []string {
	var nextSteps []string

	// Priority-based next steps
	criticalCount := 0
	quickWinCount := 0

	for _, rec := range recommendations {
		if rec.Priority == PriorityCritical {
			criticalCount++
		}
		if rec.Category == CategoryQuickWins {
			quickWinCount++
		}
	}

	if criticalCount > 0 {
		nextSteps = append(nextSteps, fmt.Sprintf("Address %d critical-priority issues immediately", criticalCount))
	}

	if quickWinCount > 0 {
		nextSteps = append(nextSteps, fmt.Sprintf("Implement %d quick wins in the next 2 weeks", quickWinCount))
	}

	nextSteps = append(nextSteps, "Review and approve the quality improvement roadmap")
	nextSteps = append(nextSteps, "Allocate development resources for quality improvements")
	nextSteps = append(nextSteps, "Establish quality gates and monitoring processes")
	nextSteps = append(nextSteps, "Schedule regular quality assessment reviews")

	return nextSteps
}

// generateOverallAssessment creates overall quality assessment
func (qr *QualityReporter) generateOverallAssessment(overallScore float64, qualityGrade string, scores ComponentScores) string {
	assessment := fmt.Sprintf("The codebase has an overall quality score of %.1f (%s grade). ", overallScore, qualityGrade)

	// Identify strongest and weakest areas
	componentMap := map[string]float64{
		"complexity":      scores.Complexity,
		"duplication":     scores.Duplication,
		"technical debt":  scores.TechnicalDebt,
		"coverage":        scores.Coverage,
		"performance":     scores.Performance,
		"maintainability": scores.Maintainability,
	}

	var strongest, weakest string
	var highestScore, lowestScore float64
	first := true

	for component, score := range componentMap {
		if first {
			strongest = component
			weakest = component
			highestScore = score
			lowestScore = score
			first = false
		} else {
			if score > highestScore {
				strongest = component
				highestScore = score
			}
			if score < lowestScore {
				weakest = component
				lowestScore = score
			}
		}
	}

	assessment += fmt.Sprintf("The strongest area is %s (%.1f), while %s (%.1f) requires the most attention. ",
		strongest, highestScore, weakest, lowestScore)

	// Add recommendation based on grade
	switch qualityGrade {
	case "Excellent":
		assessment += "Continue current practices and focus on maintaining quality standards."
	case "Good":
		assessment += "Minor improvements recommended to achieve excellence."
	case "Fair":
		assessment += "Moderate quality improvements needed to reduce technical risk."
	default: // Poor
		assessment += "Significant quality improvements required to ensure project success."
	}

	return assessment
}

// generateTrendAnalysis creates trend analysis if enabled
func (qr *QualityReporter) generateTrendAnalysis(scores ComponentScores) *QualityTrend {
	// For MVP implementation, create mock trend data
	// In production, this would analyze historical data

	now := time.Now()

	// Create sample historical data points (last 4 weeks)
	var historicalData []HistoricalDataPoint
	for i := 4; i > 0; i-- {
		timestamp := now.AddDate(0, 0, -i*7)

		// Simulate slightly improving trends
		improvement := float64(4-i) * 2.0
		historicalScores := ComponentScores{
			Complexity:      math.Max(0, scores.Complexity-improvement),
			Duplication:     math.Max(0, scores.Duplication-improvement),
			TechnicalDebt:   math.Max(0, scores.TechnicalDebt-improvement),
			Coverage:        math.Max(0, scores.Coverage-improvement),
			Performance:     math.Max(0, scores.Performance-improvement),
			Maintainability: math.Max(0, scores.Maintainability-improvement),
		}

		historicalData = append(historicalData, HistoricalDataPoint{
			Timestamp: timestamp,
			Scores:    historicalScores,
			Events:    []QualityEvent{},
		})
	}

	// Add current data point
	historicalData = append(historicalData, HistoricalDataPoint{
		Timestamp: now,
		Scores:    scores,
		Events:    []QualityEvent{},
	})

	// Create component trends
	componentTrends := map[string]TrendDirection{
		"complexity":      {Direction: "improving", Velocity: 1.5, Confidence: 0.8, Significance: "medium"},
		"duplication":     {Direction: "stable", Velocity: 0.2, Confidence: 0.9, Significance: "low"},
		"technical_debt":  {Direction: "improving", Velocity: 2.0, Confidence: 0.7, Significance: "high"},
		"coverage":        {Direction: "improving", Velocity: 1.8, Confidence: 0.8, Significance: "medium"},
		"performance":     {Direction: "stable", Velocity: 0.5, Confidence: 0.9, Significance: "low"},
		"maintainability": {Direction: "improving", Velocity: 1.2, Confidence: 0.7, Significance: "medium"},
	}

	// Overall trend
	overallTrend := TrendDirection{
		Direction:    "improving",
		Velocity:     1.4,
		Confidence:   0.8,
		Significance: "medium",
	}

	// Create predictions
	predictions := []TrendPrediction{
		{
			Component:   "overall",
			TimeFrame:   "4 weeks",
			Prediction:  qr.calculateOverallScore(scores) + 8.0,
			Confidence:  0.8,
			Assumptions: []string{"Continued improvement efforts", "No major architectural changes"},
		},
	}

	return &QualityTrend{
		Period:           "4 weeks",
		OverallTrend:     overallTrend,
		ComponentTrends:  componentTrends,
		HistoricalData:   historicalData,
		TrendPredictions: predictions,
		SeasonalPatterns: []SeasonalPattern{},
		InflectionPoints: []InflectionPoint{},
	}
}
