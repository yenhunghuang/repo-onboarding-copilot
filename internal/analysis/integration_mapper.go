package analysis

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IntegrationType represents different types of integrations
type IntegrationType string

const (
	DatabaseIntegration    IntegrationType = "database"
	APIIntegration         IntegrationType = "api"
	ServiceIntegration     IntegrationType = "service"
	AuthIntegration        IntegrationType = "auth"
	PaymentIntegration     IntegrationType = "payment"
	StorageIntegration     IntegrationType = "storage"
	AnalyticsIntegration   IntegrationType = "analytics"
	MessagingIntegration   IntegrationType = "messaging"
	CloudIntegration       IntegrationType = "cloud"
	UnknownIntegration     IntegrationType = "unknown"
)

// SecurityRiskLevel represents the security risk assessment
type SecurityRiskLevel string

const (
	LowRisk      SecurityRiskLevel = "low"
	MediumRisk   SecurityRiskLevel = "medium"
	HighRisk     SecurityRiskLevel = "high"
	CriticalRisk SecurityRiskLevel = "critical"
)

// IntegrationPoint represents an external integration point
type IntegrationPoint struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         IntegrationType        `json:"type"`
	Protocol     string                 `json:"protocol"`
	Endpoint     string                 `json:"endpoint"`
	FilePath     string                 `json:"file_path"`
	LineNumber   int                    `json:"line_number"`
	SecurityRisk SecurityRiskLevel      `json:"security_risk"`
	RiskReasons  []string               `json:"risk_reasons"`
	Environment  string                 `json:"environment"`
	Credentials  CredentialInfo         `json:"credentials"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CredentialInfo represents credential usage information
type CredentialInfo struct {
	UsesEnvVars     bool     `json:"uses_env_vars"`
	UsesHardcoded   bool     `json:"uses_hardcoded"`
	UsesConfigFile  bool     `json:"uses_config_file"`
	CredentialTypes []string `json:"credential_types"`
	SecurityIssues  []string `json:"security_issues"`
}

// IntegrationMapper maps and analyzes external integration points
type IntegrationMapper struct {
	integrations    []IntegrationPoint
	riskAssessments map[string]SecurityRiskLevel
	componentId     *ComponentIdentifier
	dataFlowAn      *DataFlowAnalyzer
}

// NewIntegrationMapper creates a new integration mapper instance
func NewIntegrationMapper(ci *ComponentIdentifier, dfa *DataFlowAnalyzer) *IntegrationMapper {
	return &IntegrationMapper{
		integrations:    make([]IntegrationPoint, 0),
		riskAssessments: make(map[string]SecurityRiskLevel),
		componentId:     ci,
		dataFlowAn:      dfa,
	}
}

// MapIntegrationPoints performs comprehensive integration point mapping
func (im *IntegrationMapper) MapIntegrationPoints(filePath string, content string) error {
	// Scan for different types of integrations
	if err := im.scanDatabaseConnections(filePath, content); err != nil {
		return err
	}
	
	if err := im.scanAPIEndpoints(filePath, content); err != nil {
		return err
	}
	
	if err := im.scanServiceIntegrations(filePath, content); err != nil {
		return err
	}
	
	if err := im.scanAuthenticationSystems(filePath, content); err != nil {
		return err
	}
	
	if err := im.scanEnvironmentVariables(filePath, content); err != nil {
		return err
	}

	return nil
}

// scanDatabaseConnections identifies database connection patterns
func (im *IntegrationMapper) scanDatabaseConnections(filePath, content string) error {
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		// MongoDB connections
		if im.containsMongoConnection(line) {
			integration := im.createMongoIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// PostgreSQL connections
		if im.containsPostgreSQLConnection(line) {
			integration := im.createPostgreSQLIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// MySQL connections
		if im.containsMySQLConnection(line) {
			integration := im.createMySQLIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// Redis connections
		if im.containsRedisConnection(line) {
			integration := im.createRedisIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
	}

	return nil
}

// scanAPIEndpoints identifies external API integrations
func (im *IntegrationMapper) scanAPIEndpoints(filePath, content string) error {
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		// Priority order: GraphQL > WebSocket > HTTP API
		// This prevents double-detection when multiple patterns match
		
		if im.containsGraphQLRequest(line) {
			// GraphQL endpoints (highest priority)
			integration := im.createGraphQLIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		} else if im.containsWebSocketConnection(line) {
			// WebSocket connections
			integration := im.createWebSocketIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		} else if im.containsHTTPRequest(line) {
			// HTTP/HTTPS API calls (lowest priority)
			integration := im.createAPIIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
	}

	return nil
}

// scanServiceIntegrations identifies third-party service integrations
func (im *IntegrationMapper) scanServiceIntegrations(filePath, content string) error {
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		// AWS services
		if im.containsAWSService(line) {
			integration := im.createAWSIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// Google Cloud services
		if im.containsGCPService(line) {
			integration := im.createGCPIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// Payment services (Stripe, PayPal)
		if im.containsPaymentService(line) {
			integration := im.createPaymentIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// Analytics services
		if im.containsAnalyticsService(line) {
			integration := im.createAnalyticsIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// Messaging services
		if im.containsMessagingService(line) {
			integration := im.createMessagingIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
	}

	return nil
}

// scanAuthenticationSystems identifies authentication integrations
func (im *IntegrationMapper) scanAuthenticationSystems(filePath, content string) error {
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		// OAuth providers
		if im.containsOAuthProvider(line) {
			integration := im.createOAuthIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// JWT handling
		if im.containsJWTUsage(line) {
			integration := im.createJWTIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
		
		// LDAP/Active Directory
		if im.containsLDAPConnection(line) {
			integration := im.createLDAPIntegration(filePath, content, line, lineNum)
			im.integrations = append(im.integrations, integration)
		}
	}

	return nil
}

// scanEnvironmentVariables identifies environment variable dependencies
func (im *IntegrationMapper) scanEnvironmentVariables(filePath, content string) error {
	lines := strings.Split(content, "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "process.env") || strings.Contains(line, "os.Getenv") {
			envVar := im.extractEnvironmentVariable(line)
			if envVar != "" {
				integration := im.createEnvironmentIntegration(filePath, line, lineNum, envVar)
				im.integrations = append(im.integrations, integration)
			}
		}
	}

	return nil
}

// Pattern detection methods

func (im *IntegrationMapper) containsMongoConnection(line string) bool {
	return strings.Contains(line, "mongodb://") || 
		   strings.Contains(line, "mongodb+srv://") ||
		   strings.Contains(line, "MongoClient") ||
		   strings.Contains(line, "mongoose.connect")
}

func (im *IntegrationMapper) containsPostgreSQLConnection(line string) bool {
	return strings.Contains(line, "postgresql://") || 
		   strings.Contains(line, "postgres://") ||
		   strings.Contains(line, "pg.Client") ||
		   strings.Contains(line, "new Pool")
}

func (im *IntegrationMapper) containsMySQLConnection(line string) bool {
	return strings.Contains(line, "mysql://") ||
		   strings.Contains(line, "mysql.createConnection") ||
		   strings.Contains(line, "mysql2")
}

func (im *IntegrationMapper) containsRedisConnection(line string) bool {
	return strings.Contains(line, "redis://") ||
		   strings.Contains(line, "createClient") ||
		   strings.Contains(line, "new Redis")
}

func (im *IntegrationMapper) containsHTTPRequest(line string) bool {
	// Check for direct HTTP/HTTPS URLs with HTTP methods
	hasURL := strings.Contains(line, "http://") || strings.Contains(line, "https://")
	hasMethod := strings.Contains(line, "fetch") || strings.Contains(line, "axios") ||
		strings.Contains(line, "request") || strings.Contains(line, "get") ||
		strings.Contains(line, "post")
	
	if hasURL && hasMethod {
		return true
	}
	
	// Also detect HTTP method calls that might use variables for URLs
	// Common patterns: axios.get(url), fetch(url), request.post(url)
	return (strings.Contains(line, "axios.get(") || strings.Contains(line, "axios.post(") ||
		strings.Contains(line, "axios.put(") || strings.Contains(line, "axios.delete(") ||
		strings.Contains(line, "fetch(") || 
		strings.Contains(line, "request.get(") || strings.Contains(line, "request.post(") ||
		strings.Contains(line, "request.put(") || strings.Contains(line, "request.delete(")) &&
		!strings.Contains(line, "://") // Avoid double-detection with direct URL patterns
}

func (im *IntegrationMapper) containsGraphQLRequest(line string) bool {
	// Only detect GraphQL client instantiations with endpoints, not individual queries
	return (strings.Contains(line, "apollo") && strings.Contains(line, "new ")) ||
		   (strings.Contains(line, "graphql") && strings.Contains(line, "://"))
}

func (im *IntegrationMapper) containsWebSocketConnection(line string) bool {
	// Detect actual WebSocket connections and instantiations, not just imports
	return strings.Contains(line, "ws://") || 
		   strings.Contains(line, "wss://") ||
		   (strings.Contains(line, "WebSocket") && strings.Contains(line, "new ")) ||
		   (strings.Contains(line, "socket.io") && !strings.Contains(line, "import"))
}

func (im *IntegrationMapper) containsAWSService(line string) bool {
	return strings.Contains(line, "aws-sdk") ||
		   strings.Contains(line, ".amazonaws.com") ||
		   strings.Contains(line, "AWS.") ||
		   strings.Contains(line, "S3") ||
		   strings.Contains(line, "DynamoDB") ||
		   strings.Contains(line, "Lambda")
}

func (im *IntegrationMapper) containsGCPService(line string) bool {
	return strings.Contains(line, "google-cloud") ||
		   strings.Contains(line, ".googleapis.com") ||
		   strings.Contains(line, "firebase") ||
		   strings.Contains(line, "firestore")
}

func (im *IntegrationMapper) containsPaymentService(line string) bool {
	return strings.Contains(line, "stripe") ||
		   strings.Contains(line, "paypal") ||
		   strings.Contains(line, "payment") ||
		   strings.Contains(line, "checkout")
}

func (im *IntegrationMapper) containsAnalyticsService(line string) bool {
	return strings.Contains(line, "analytics") ||
		   strings.Contains(line, "google-analytics") ||
		   strings.Contains(line, "gtag") ||
		   strings.Contains(line, "ga-gtag") ||
		   strings.Contains(line, "mixpanel") ||
		   strings.Contains(line, "amplitude")
}

func (im *IntegrationMapper) containsMessagingService(line string) bool {
	return strings.Contains(line, "kafka") ||
		   strings.Contains(line, "rabbitmq") ||
		   strings.Contains(line, "sqs") ||
		   strings.Contains(line, "pubsub")
}

func (im *IntegrationMapper) containsOAuthProvider(line string) bool {
	return strings.Contains(line, "oauth") ||
		   strings.Contains(line, "google") && strings.Contains(line, "auth") ||
		   strings.Contains(line, "facebook") && strings.Contains(line, "auth") ||
		   strings.Contains(line, "github") && strings.Contains(line, "auth")
}

func (im *IntegrationMapper) containsJWTUsage(line string) bool {
	return strings.Contains(line, "jwt") ||
		   strings.Contains(line, "jsonwebtoken") ||
		   strings.Contains(line, "bearer")
}

func (im *IntegrationMapper) containsLDAPConnection(line string) bool {
	return strings.Contains(line, "ldap") ||
		   strings.Contains(line, "active directory") ||
		   strings.Contains(line, "AD")
}

// Integration creation methods

func (im *IntegrationMapper) createMongoIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractConnectionString(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "mongodb", lineNum),
		Name:         "MongoDB Database",
		Type:         DatabaseIntegration,
		Protocol:     "MongoDB",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessDatabaseRisk(credentials, endpoint),
		RiskReasons:  im.getDatabaseRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"database_type": "mongodb",
			"source_line":   line,
		},
	}
}

func (im *IntegrationMapper) createPostgreSQLIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractConnectionString(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "postgresql", lineNum),
		Name:         "PostgreSQL Database",
		Type:         DatabaseIntegration,
		Protocol:     "PostgreSQL",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessDatabaseRisk(credentials, endpoint),
		RiskReasons:  im.getDatabaseRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"database_type": "postgresql",
			"source_line":   line,
		},
	}
}

func (im *IntegrationMapper) createMySQLIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractConnectionString(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "mysql", lineNum),
		Name:         "MySQL Database",
		Type:         DatabaseIntegration,
		Protocol:     "MySQL",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessDatabaseRisk(credentials, endpoint),
		RiskReasons:  im.getDatabaseRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"database_type": "mysql",
			"source_line":   line,
		},
	}
}

func (im *IntegrationMapper) createRedisIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractConnectionString(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "redis", lineNum),
		Name:         "Redis Cache",
		Type:         StorageIntegration,
		Protocol:     "Redis",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessCacheRisk(credentials, endpoint),
		RiskReasons:  im.getCacheRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"storage_type": "redis",
			"source_line":  line,
		},
	}
}

func (im *IntegrationMapper) createAPIIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractAPIEndpoint(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "api", lineNum),
		Name:         "External API",
		Type:         APIIntegration,
		Protocol:     im.detectAPIProtocol(endpoint),
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessAPIRisk(credentials, endpoint),
		RiskReasons:  im.getAPIRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"api_type":    "rest",
			"source_line": line,
		},
	}
}

func (im *IntegrationMapper) createGraphQLIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractAPIEndpoint(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "graphql", lineNum),
		Name:         "GraphQL API",
		Type:         APIIntegration,
		Protocol:     "GraphQL",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessAPIRisk(credentials, endpoint),
		RiskReasons:  im.getAPIRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"api_type":    "graphql",
			"source_line": line,
		},
	}
}

func (im *IntegrationMapper) createWebSocketIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractWebSocketEndpoint(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "websocket", lineNum),
		Name:         "WebSocket Connection",
		Type:         ServiceIntegration,
		Protocol:     "WebSocket",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessWebSocketRisk(credentials, endpoint),
		RiskReasons:  im.getWebSocketRiskReasons(credentials, endpoint),
		Environment:  im.detectEnvironment(endpoint),
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"connection_type": "websocket",
			"source_line":     line,
		},
	}
}

func (im *IntegrationMapper) createAWSIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	service := im.extractAWSService(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "aws", lineNum),
		Name:         "AWS " + service,
		Type:         CloudIntegration,
		Protocol:     "AWS SDK",
		Endpoint:     service,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessCloudRisk(credentials, service),
		RiskReasons:  im.getCloudRiskReasons(credentials, service),
		Environment:  "cloud",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"cloud_provider": "aws",
			"service":        service,
			"source_line":    line,
		},
	}
}

func (im *IntegrationMapper) createGCPIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	service := im.extractGCPService(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "gcp", lineNum),
		Name:         "Google Cloud " + service,
		Type:         CloudIntegration,
		Protocol:     "GCP SDK",
		Endpoint:     service,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessCloudRisk(credentials, service),
		RiskReasons:  im.getCloudRiskReasons(credentials, service),
		Environment:  "cloud",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"cloud_provider": "gcp",
			"service":        service,
			"source_line":    line,
		},
	}
}

func (im *IntegrationMapper) createPaymentIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	provider := im.extractPaymentProvider(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "payment", lineNum),
		Name:         provider + " Payment",
		Type:         PaymentIntegration,
		Protocol:     "HTTPS",
		Endpoint:     provider,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: CriticalRisk, // Payment integrations are always high risk
		RiskReasons:  []string{"Handles sensitive payment data", "PCI compliance required", "Financial transactions"},
		Environment:  "external",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"payment_provider": provider,
			"source_line":      line,
		},
	}
}

func (im *IntegrationMapper) createAnalyticsIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	provider := im.extractAnalyticsProvider(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "analytics", lineNum),
		Name:         provider + " Analytics",
		Type:         AnalyticsIntegration,
		Protocol:     "HTTPS",
		Endpoint:     provider,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: MediumRisk, // Analytics may collect user data
		RiskReasons:  []string{"Collects user behavior data", "Privacy considerations"},
		Environment:  "external",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"analytics_provider": provider,
			"source_line":        line,
		},
	}
}

func (im *IntegrationMapper) createMessagingIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	service := im.extractMessagingService(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "messaging", lineNum),
		Name:         service + " Messaging",
		Type:         MessagingIntegration,
		Protocol:     im.detectMessagingProtocol(service),
		Endpoint:     service,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessMessagingRisk(credentials, service),
		RiskReasons:  im.getMessagingRiskReasons(credentials, service),
		Environment:  "external",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"messaging_service": service,
			"source_line":       line,
		},
	}
}

func (im *IntegrationMapper) createOAuthIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	provider := im.extractOAuthProvider(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "oauth", lineNum),
		Name:         provider + " OAuth",
		Type:         AuthIntegration,
		Protocol:     "OAuth 2.0",
		Endpoint:     provider,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: HighRisk, // Auth integrations are high risk
		RiskReasons:  []string{"Handles user authentication", "Access to user data", "Token management"},
		Environment:  "external",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"auth_provider": provider,
			"source_line":   line,
		},
	}
}

func (im *IntegrationMapper) createJWTIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "jwt", lineNum),
		Name:         "JWT Authentication",
		Type:         AuthIntegration,
		Protocol:     "JWT",
		Endpoint:     "local",
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: HighRisk,
		RiskReasons:  []string{"Token-based authentication", "Secret key management", "Token expiration handling"},
		Environment:  "local",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"auth_type":   "jwt",
			"source_line": line,
		},
	}
}

func (im *IntegrationMapper) createLDAPIntegration(filePath, content, line string, lineNum int) IntegrationPoint {
	endpoint := im.extractLDAPEndpoint(line)
	credentials := im.analyzeCredentials(content)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "ldap", lineNum),
		Name:         "LDAP/Active Directory",
		Type:         AuthIntegration,
		Protocol:     "LDAP",
		Endpoint:     endpoint,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: HighRisk,
		RiskReasons:  []string{"Directory service access", "User credential verification", "Domain authentication"},
		Environment:  "internal",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"auth_type":   "ldap",
			"source_line": line,
		},
	}
}

func (im *IntegrationMapper) createEnvironmentIntegration(filePath, line string, lineNum int, envVar string) IntegrationPoint {
	credentials := im.analyzeCredentials(line)
	
	return IntegrationPoint{
		ID:           im.generateID(filePath, "env", lineNum),
		Name:         "Environment Variable: " + envVar,
		Type:         UnknownIntegration,
		Protocol:     "Environment",
		Endpoint:     envVar,
		FilePath:     filePath,
		LineNumber:   lineNum + 1,
		SecurityRisk: im.assessEnvironmentVarRisk(envVar),
		RiskReasons:  im.getEnvironmentVarRiskReasons(envVar),
		Environment:  "config",
		Credentials:  credentials,
		Metadata: map[string]interface{}{
			"env_var":     envVar,
			"source_line": line,
		},
	}
}

// Helper methods for extraction and analysis

func (im *IntegrationMapper) generateID(filePath, integrationType string, lineNum int) string {
	return strings.ReplaceAll(filePath, "/", "_") + "_" + integrationType + "_" + fmt.Sprintf("%d", lineNum)
}

func (im *IntegrationMapper) extractConnectionString(line string) string {
	// Extract connection strings from various patterns
	if strings.Contains(line, "://") {
		start := strings.Index(line, "://")
		if start != -1 {
			// Find the start of the connection string
			lineStart := start
			for lineStart > 0 && line[lineStart-1] != '"' && line[lineStart-1] != '\'' && line[lineStart-1] != '`' {
				lineStart--
			}
			
			// Find the end of the connection string
			end := start + 3
			for end < len(line) && line[end] != '"' && line[end] != '\'' && line[end] != '`' && line[end] != ' ' && line[end] != ')' {
				end++
			}
			
			return line[lineStart:end]
		}
	}
	
	return "unknown"
}

func (im *IntegrationMapper) extractAPIEndpoint(line string) string {
	// Extract API endpoints from fetch/axios calls
	patterns := []string{"http://", "https://"}
	
	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			start := strings.Index(line, pattern)
			end := start + len(pattern)
			
			// Find the end of the URL
			for end < len(line) && line[end] != '"' && line[end] != '\'' && line[end] != '`' && line[end] != ' ' && line[end] != ')' {
				end++
			}
			
			return line[start:end]
		}
	}
	
	return "unknown"
}

func (im *IntegrationMapper) extractWebSocketEndpoint(line string) string {
	patterns := []string{"ws://", "wss://"}
	
	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			start := strings.Index(line, pattern)
			end := start + len(pattern)
			
			for end < len(line) && line[end] != '"' && line[end] != '\'' && line[end] != ' ' {
				end++
			}
			
			return line[start:end]
		}
	}
	
	return "websocket"
}

func (im *IntegrationMapper) extractAWSService(line string) string {
	services := []string{"S3", "DynamoDB", "Lambda", "RDS", "SQS", "SNS", "CloudWatch"}
	
	for _, service := range services {
		if strings.Contains(strings.ToUpper(line), service) {
			return service
		}
	}
	
	return "AWS"
}

func (im *IntegrationMapper) extractGCPService(line string) string {
	if strings.Contains(line, "firestore") {
		return "Firestore"
	}
	if strings.Contains(line, "firebase") {
		return "Firebase"
	}
	if strings.Contains(line, "cloud-storage") {
		return "Cloud Storage"
	}
	
	return "Google Cloud"
}

func (im *IntegrationMapper) extractPaymentProvider(line string) string {
	if strings.Contains(strings.ToLower(line), "stripe") {
		return "Stripe"
	}
	if strings.Contains(strings.ToLower(line), "paypal") {
		return "PayPal"
	}
	if strings.Contains(strings.ToLower(line), "square") {
		return "Square"
	}
	
	return "Payment Provider"
}

func (im *IntegrationMapper) extractAnalyticsProvider(line string) string {
	lowerLine := strings.ToLower(line)
	if strings.Contains(lowerLine, "google-analytics") || strings.Contains(lowerLine, "gtag") || strings.Contains(lowerLine, "ga-gtag") {
		return "Google Analytics"
	}
	if strings.Contains(lowerLine, "mixpanel") {
		return "Mixpanel"
	}
	if strings.Contains(lowerLine, "amplitude") {
		return "Amplitude"
	}
	
	return "Analytics"
}

func (im *IntegrationMapper) extractMessagingService(line string) string {
	if strings.Contains(strings.ToLower(line), "kafka") {
		return "Kafka"
	}
	if strings.Contains(strings.ToLower(line), "rabbitmq") {
		return "RabbitMQ"
	}
	if strings.Contains(strings.ToLower(line), "sqs") {
		return "AWS SQS"
	}
	
	return "Messaging"
}

func (im *IntegrationMapper) extractOAuthProvider(line string) string {
	if strings.Contains(strings.ToLower(line), "google") {
		return "Google"
	}
	if strings.Contains(strings.ToLower(line), "facebook") {
		return "Facebook"
	}
	if strings.Contains(strings.ToLower(line), "github") {
		return "GitHub"
	}
	
	return "OAuth Provider"
}

func (im *IntegrationMapper) extractLDAPEndpoint(line string) string {
	if strings.Contains(line, "ldap://") {
		return im.extractConnectionString(line)
	}
	
	return "ldap://unknown"
}

func (im *IntegrationMapper) extractEnvironmentVariable(line string) string {
	// Extract environment variable names
	patterns := []string{"process.env.", "os.Getenv("}
	
	for _, pattern := range patterns {
		if strings.Contains(line, pattern) {
			start := strings.Index(line, pattern) + len(pattern)
			if pattern == "os.Getenv(" {
				start++ // Skip opening quote
			}
			
			end := start
			for end < len(line) && line[end] != '"' && line[end] != '\'' && line[end] != ')' && line[end] != ']' {
				end++
			}
			
			return line[start:end]
		}
	}
	
	return ""
}

func (im *IntegrationMapper) analyzeCredentials(content string) CredentialInfo {
	cred := CredentialInfo{
		CredentialTypes: make([]string, 0),
		SecurityIssues:  make([]string, 0),
	}
	
	// Check for environment variables
	if strings.Contains(content, "process.env") || strings.Contains(content, "os.Getenv") {
		cred.UsesEnvVars = true
		cred.CredentialTypes = append(cred.CredentialTypes, "environment_variables")
	}
	
	// Check for hardcoded credentials (look for quoted strings with credentials)
	if strings.Contains(content, "password") && (strings.Contains(content, "'") || strings.Contains(content, "\"")) {
		if !strings.Contains(content, "process.env") {
			// Check if password appears to be hardcoded (in quotes)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "password") && !strings.Contains(line, "process.env") {
					// Look for patterns like: password: 'value', password: "value", password='value', password="value"
					if (strings.Contains(line, "password:") && (strings.Contains(line, "'") || strings.Contains(line, "\""))) ||
					   (strings.Contains(line, "password=") && (strings.Contains(line, "'") || strings.Contains(line, "\""))) ||
					   (strings.Contains(line, "password ") && (strings.Contains(line, "'") || strings.Contains(line, "\""))) {
						cred.UsesHardcoded = true
						cred.CredentialTypes = append(cred.CredentialTypes, "hardcoded_password")
						cred.SecurityIssues = append(cred.SecurityIssues, "Hardcoded password detected")
						break
					}
				}
			}
		}
	}
	
	// Check for API keys
	if strings.Contains(strings.ToLower(content), "apikey") || strings.Contains(strings.ToLower(content), "api_key") {
		cred.CredentialTypes = append(cred.CredentialTypes, "api_key")
		if !strings.Contains(content, "process.env") {
			cred.UsesHardcoded = true
			cred.SecurityIssues = append(cred.SecurityIssues, "Hardcoded API key detected")
		}
	}
	
	// Check for tokens
	if strings.Contains(strings.ToLower(content), "token") {
		cred.CredentialTypes = append(cred.CredentialTypes, "token")
		if !strings.Contains(content, "process.env") {
			cred.UsesHardcoded = true
			cred.SecurityIssues = append(cred.SecurityIssues, "Hardcoded token detected")
		}
	}
	
	return cred
}

func (im *IntegrationMapper) detectEnvironment(endpoint string) string {
	if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
		return "development"
	}
	if strings.Contains(endpoint, "staging") || strings.Contains(endpoint, "test") {
		return "staging"
	}
	if strings.Contains(endpoint, "prod") || strings.Contains(endpoint, ".com") || strings.Contains(endpoint, ".net") {
		return "production"
	}
	
	return "unknown"
}

func (im *IntegrationMapper) detectAPIProtocol(endpoint string) string {
	if strings.HasPrefix(endpoint, "https://") {
		return "HTTPS"
	}
	if strings.HasPrefix(endpoint, "http://") {
		return "HTTP"
	}
	
	return "Unknown"
}

func (im *IntegrationMapper) detectMessagingProtocol(service string) string {
	switch strings.ToLower(service) {
	case "kafka":
		return "Kafka Protocol"
	case "rabbitmq":
		return "AMQP"
	case "aws sqs":
		return "AWS SQS"
	default:
		return "Unknown"
	}
}

// Risk assessment methods

func (im *IntegrationMapper) assessDatabaseRisk(credentials CredentialInfo, endpoint string) SecurityRiskLevel {
	risk := LowRisk
	
	if credentials.UsesHardcoded {
		risk = CriticalRisk
	} else if len(credentials.SecurityIssues) > 0 {
		risk = HighRisk
	} else if strings.Contains(endpoint, "localhost") {
		risk = LowRisk
	} else {
		risk = MediumRisk
	}
	
	return risk
}

func (im *IntegrationMapper) getDatabaseRiskReasons(credentials CredentialInfo, endpoint string) []string {
	reasons := make([]string, 0)
	
	if credentials.UsesHardcoded {
		reasons = append(reasons, "Hardcoded database credentials")
	}
	
	if strings.Contains(endpoint, "localhost") {
		reasons = append(reasons, "Development database")
	} else {
		reasons = append(reasons, "External database connection")
	}
	
	if len(credentials.SecurityIssues) > 0 {
		reasons = append(reasons, credentials.SecurityIssues...)
	}
	
	return reasons
}

func (im *IntegrationMapper) assessCacheRisk(credentials CredentialInfo, endpoint string) SecurityRiskLevel {
	return im.assessDatabaseRisk(credentials, endpoint) // Similar logic
}

func (im *IntegrationMapper) getCacheRiskReasons(credentials CredentialInfo, endpoint string) []string {
	return im.getDatabaseRiskReasons(credentials, endpoint)
}

func (im *IntegrationMapper) assessAPIRisk(credentials CredentialInfo, endpoint string) SecurityRiskLevel {
	risk := LowRisk
	
	if credentials.UsesHardcoded {
		risk = HighRisk
	} else if strings.HasPrefix(endpoint, "http://") {
		risk = MediumRisk // HTTP is insecure
	} else if strings.HasPrefix(endpoint, "https://") {
		risk = LowRisk
	} else {
		risk = MediumRisk
	}
	
	return risk
}

func (im *IntegrationMapper) getAPIRiskReasons(credentials CredentialInfo, endpoint string) []string {
	reasons := make([]string, 0)
	
	if credentials.UsesHardcoded {
		reasons = append(reasons, "Hardcoded API credentials")
	}
	
	if strings.HasPrefix(endpoint, "http://") {
		reasons = append(reasons, "Insecure HTTP connection")
	}
	
	if len(credentials.SecurityIssues) > 0 {
		reasons = append(reasons, credentials.SecurityIssues...)
	}
	
	return reasons
}

func (im *IntegrationMapper) assessWebSocketRisk(credentials CredentialInfo, endpoint string) SecurityRiskLevel {
	if strings.HasPrefix(endpoint, "ws://") {
		return MediumRisk // Insecure WebSocket
	}
	
	return LowRisk // WSS is secure
}

func (im *IntegrationMapper) getWebSocketRiskReasons(credentials CredentialInfo, endpoint string) []string {
	reasons := make([]string, 0)
	
	if strings.HasPrefix(endpoint, "ws://") {
		reasons = append(reasons, "Insecure WebSocket connection")
	}
	
	return reasons
}

func (im *IntegrationMapper) assessCloudRisk(credentials CredentialInfo, service string) SecurityRiskLevel {
	if credentials.UsesHardcoded {
		return CriticalRisk
	}
	
	return MediumRisk // Cloud services need proper IAM
}

func (im *IntegrationMapper) getCloudRiskReasons(credentials CredentialInfo, service string) []string {
	reasons := make([]string, 0)
	
	if credentials.UsesHardcoded {
		reasons = append(reasons, "Hardcoded cloud credentials")
	}
	
	reasons = append(reasons, "Requires proper IAM configuration")
	reasons = append(reasons, "Cloud service access")
	
	return reasons
}

func (im *IntegrationMapper) assessMessagingRisk(credentials CredentialInfo, service string) SecurityRiskLevel {
	return MediumRisk // Messaging services typically need authentication
}

func (im *IntegrationMapper) getMessagingRiskReasons(credentials CredentialInfo, service string) []string {
	return []string{"Message queue access", "Potential data exposure"}
}

func (im *IntegrationMapper) assessEnvironmentVarRisk(envVar string) SecurityRiskLevel {
	sensitive := []string{"password", "secret", "key", "token", "credential"}
	
	envVarLower := strings.ToLower(envVar)
	for _, keyword := range sensitive {
		if strings.Contains(envVarLower, keyword) {
			return HighRisk
		}
	}
	
	return LowRisk
}

func (im *IntegrationMapper) getEnvironmentVarRiskReasons(envVar string) []string {
	reasons := make([]string, 0)
	
	sensitive := []string{"password", "secret", "key", "token", "credential"}
	envVarLower := strings.ToLower(envVar)
	
	for _, keyword := range sensitive {
		if strings.Contains(envVarLower, keyword) {
			reasons = append(reasons, "Contains sensitive information: "+keyword)
		}
	}
	
	if len(reasons) == 0 {
		reasons = append(reasons, "Environment variable dependency")
	}
	
	return reasons
}

// Analysis result methods

func (im *IntegrationMapper) GetIntegrationPoints() []IntegrationPoint {
	return im.integrations
}

func (im *IntegrationMapper) GetIntegrationsByType(integrationType IntegrationType) []IntegrationPoint {
	filtered := make([]IntegrationPoint, 0)
	
	for _, integration := range im.integrations {
		if integration.Type == integrationType {
			filtered = append(filtered, integration)
		}
	}
	
	return filtered
}

func (im *IntegrationMapper) GetHighRiskIntegrations() []IntegrationPoint {
	highRisk := make([]IntegrationPoint, 0)
	
	for _, integration := range im.integrations {
		if integration.SecurityRisk == HighRisk || integration.SecurityRisk == CriticalRisk {
			highRisk = append(highRisk, integration)
		}
	}
	
	return highRisk
}

func (im *IntegrationMapper) GetIntegrationStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Count by type
	typeCounts := make(map[IntegrationType]int)
	riskCounts := make(map[SecurityRiskLevel]int)
	
	for _, integration := range im.integrations {
		typeCounts[integration.Type]++
		riskCounts[integration.SecurityRisk]++
	}
	
	stats["total_integrations"] = len(im.integrations)
	stats["by_type"] = typeCounts
	stats["by_risk"] = riskCounts
	stats["high_risk_count"] = riskCounts[HighRisk] + riskCounts[CriticalRisk]
	
	return stats
}

func (im *IntegrationMapper) GetSecurityAssessment() map[string]interface{} {
	assessment := make(map[string]interface{})
	
	highRiskIntegrations := im.GetHighRiskIntegrations()
	totalIntegrations := len(im.integrations)
	
	if totalIntegrations > 0 {
		riskPercentage := float64(len(highRiskIntegrations)) / float64(totalIntegrations) * 100
		assessment["high_risk_percentage"] = riskPercentage
		
		if riskPercentage > 50 {
			assessment["overall_risk"] = "high"
		} else if riskPercentage > 25 {
			assessment["overall_risk"] = "medium"
		} else {
			assessment["overall_risk"] = "low"
		}
	} else {
		assessment["overall_risk"] = "none"
		assessment["high_risk_percentage"] = 0.0
	}
	
	// Collect all security issues
	allIssues := make([]string, 0)
	for _, integration := range im.integrations {
		if len(integration.Credentials.SecurityIssues) > 0 {
			allIssues = append(allIssues, integration.Credentials.SecurityIssues...)
		}
	}
	
	assessment["security_issues"] = allIssues
	assessment["recommendations"] = im.generateSecurityRecommendations()
	
	return assessment
}

func (im *IntegrationMapper) generateSecurityRecommendations() []string {
	recommendations := make([]string, 0)
	
	// Check for hardcoded credentials
	hasHardcoded := false
	hasInsecureConnections := false
	
	for _, integration := range im.integrations {
		if integration.Credentials.UsesHardcoded {
			hasHardcoded = true
		}
		
		if strings.HasPrefix(integration.Endpoint, "http://") {
			hasInsecureConnections = true
		}
	}
	
	if hasHardcoded {
		recommendations = append(recommendations, "Move hardcoded credentials to environment variables")
		recommendations = append(recommendations, "Implement proper secrets management")
	}
	
	if hasInsecureConnections {
		recommendations = append(recommendations, "Replace HTTP connections with HTTPS")
		recommendations = append(recommendations, "Enable TLS encryption for all external connections")
	}
	
	if len(im.integrations) > 10 {
		recommendations = append(recommendations, "Consider API gateway for managing multiple integrations")
	}
	
	return recommendations
}

func (im *IntegrationMapper) ExportToJSON() ([]byte, error) {
	result := map[string]interface{}{
		"integration_points":   im.integrations,
		"statistics":           im.GetIntegrationStats(),
		"security_assessment":  im.GetSecurityAssessment(),
	}
	
	return json.MarshalIndent(result, "", "  ")
}