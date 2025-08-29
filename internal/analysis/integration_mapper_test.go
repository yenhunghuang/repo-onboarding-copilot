package analysis

import (
	"strings"
	"testing"
)

func TestIntegrationMapper_ScanDatabaseConnections(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	tests := []struct {
		name             string
		filePath         string
		content          string
		expectedType     IntegrationType
		expectedMinRisk  SecurityRiskLevel
		expectedProtocol string
	}{
		{
			name:     "MongoDB Connection with Environment Variables",
			filePath: "/src/config/database.js",
			content: `
const mongoose = require('mongoose');

const connectDB = async () => {
  try {
    await mongoose.connect(process.env.MONGODB_URI, {
      useNewUrlParser: true,
      useUnifiedTopology: true
    });
    console.log('MongoDB connected');
  } catch (error) {
    console.error('Database connection failed:', error);
  }
};

module.exports = connectDB;
			`,
			expectedType:     DatabaseIntegration,
			expectedMinRisk:  LowRisk,
			expectedProtocol: "MongoDB",
		},
		{
			name:     "PostgreSQL Connection with Hardcoded Credentials",
			filePath: "/src/config/postgres.js",
			content: `
const { Pool } = require('pg');

const pool = new Pool({
  user: 'admin',
  host: 'localhost',
  database: 'myapp',
  password: 'secret123',
  port: 5432,
});

module.exports = pool;
			`,
			expectedType:     DatabaseIntegration,
			expectedMinRisk:  CriticalRisk,
			expectedProtocol: "PostgreSQL",
		},
		{
			name:     "MySQL Connection String",
			filePath: "/src/config/mysql.js",
			content: `
const mysql = require('mysql2');

const connection = mysql.createConnection({
  host: process.env.DB_HOST,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
  database: process.env.DB_NAME
});

connection.connect();
			`,
			expectedType:     DatabaseIntegration,
			expectedMinRisk:  LowRisk,
			expectedProtocol: "MySQL",
		},
		{
			name:     "Redis Connection",
			filePath: "/src/config/redis.js",
			content: `
const redis = require('redis');

const client = redis.createClient({
  url: process.env.REDIS_URL || 'redis://localhost:6379'
});

client.on('error', (err) => console.log('Redis Client Error', err));
			`,
			expectedType:     StorageIntegration,
			expectedMinRisk:  LowRisk,
			expectedProtocol: "Redis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := im.scanDatabaseConnections(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("scanDatabaseConnections() error = %v", err)
			}

			integrations := im.GetIntegrationPoints()
			found := false
			for _, integration := range integrations {
				if integration.Type == tt.expectedType && integration.FilePath == tt.filePath {
					found = true

					if integration.Protocol != tt.expectedProtocol {
						t.Errorf("Expected protocol %s, got %s", tt.expectedProtocol, integration.Protocol)
					}

					// Check security risk level
					if tt.expectedMinRisk == CriticalRisk && integration.SecurityRisk != CriticalRisk {
						t.Errorf("Expected critical risk for hardcoded credentials")
					}

					// Check that risk reasons are provided
					if len(integration.RiskReasons) == 0 {
						t.Error("Expected risk reasons to be provided")
					}

					// Check credentials analysis
					if tt.name == "PostgreSQL Connection with Hardcoded Credentials" {
						if !integration.Credentials.UsesHardcoded {
							t.Error("Expected hardcoded credentials detection")
						}
					}

					break
				}
			}

			if !found {
				t.Errorf("Expected to find %s integration", tt.expectedType)
			}
		})
	}
}

func TestIntegrationMapper_ScanAPIEndpoints(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		content       string
		expectedCount int
		expectedRisk  SecurityRiskLevel
	}{
		{
			name:     "HTTPS API Calls",
			filePath: "/src/services/apiService.js",
			content: `
import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'https://api.example.com';

export const userService = {
  async getUsers() {
    const response = await axios.get(API_BASE_URL + '/users');
    return response.data;
  },
  
  async createUser(userData) {
    const response = await fetch('https://api.users.com/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + process.env.API_TOKEN
      },
      body: JSON.stringify(userData)
    });
    return response.json();
  }
};
			`,
			expectedCount: 2,
			expectedRisk:  LowRisk,
		},
		{
			name:     "Insecure HTTP API Call",
			filePath: "/src/services/insecureService.js",
			content: `
export const fetchData = async () => {
  const response = await fetch('http://api.example.com/data');
  return response.json();
};
			`,
			expectedCount: 1,
			expectedRisk:  MediumRisk,
		},
		{
			name:     "GraphQL API",
			filePath: "/src/graphql/client.js",
			content: `
import { ApolloClient, InMemoryCache, gql } from '@apollo/client';

const client = new ApolloClient({
  uri: 'https://api.graphql.example.com/graphql',
  cache: new InMemoryCache()
});

const GET_USERS = gql` + "`" + `
  query GetUsers {
    users {
      id
      name
      email
    }
  }
` + "`" + `;

export { client, GET_USERS };
			`,
			expectedCount: 1,
			expectedRisk:  LowRisk,
		},
		{
			name:     "WebSocket Connection",
			filePath: "/src/services/websocket.js",
			content: `
import io from 'socket.io-client';

const socket = io('wss://realtime.example.com', {
  transports: ['websocket'],
  upgrade: true
});

socket.on('connect', () => {
  console.log('Connected to WebSocket');
});

export default socket;
			`,
			expectedCount: 1,
			expectedRisk:  LowRisk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh IntegrationMapper for each test to avoid accumulation
			ci := NewComponentIdentifier()
			dfa := NewDataFlowAnalyzer(ci)
			im := NewIntegrationMapper(ci, dfa)

			err := im.scanAPIEndpoints(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("scanAPIEndpoints() error = %v", err)
			}

			integrations := im.GetIntegrationsByType(APIIntegration)
			if len(integrations) == 0 && tt.name == "WebSocket Connection" {
				integrations = im.GetIntegrationsByType(ServiceIntegration)
			}

			fileIntegrations := 0
			for _, integration := range integrations {
				if integration.FilePath == tt.filePath {
					fileIntegrations++
				}
			}

			if fileIntegrations != tt.expectedCount {
				t.Errorf("Expected %d API integrations, got %d", tt.expectedCount, fileIntegrations)
			}

			// Check security risk for insecure connections
			if tt.name == "Insecure HTTP API Call" {
				found := false
				for _, integration := range integrations {
					if integration.FilePath == tt.filePath && integration.SecurityRisk == MediumRisk {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected medium risk for HTTP connection")
				}
			}
		})
	}
}

func TestIntegrationMapper_ScanServiceIntegrations(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	tests := []struct {
		name         string
		filePath     string
		content      string
		expectedType IntegrationType
		expectedRisk SecurityRiskLevel
	}{
		{
			name:     "AWS S3 Integration",
			filePath: "/src/services/s3Service.js",
			content: `
const AWS = require('aws-sdk');

const s3 = new AWS.S3({
  accessKeyId: process.env.AWS_ACCESS_KEY_ID,
  secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
  region: process.env.AWS_REGION
});

const uploadFile = async (file, bucketName) => {
  const params = {
    Bucket: bucketName,
    Key: file.name,
    Body: file.data,
    ContentType: file.mimetype
  };
  
  return s3.upload(params).promise();
};

module.exports = { uploadFile };
			`,
			expectedType: CloudIntegration,
			expectedRisk: MediumRisk,
		},
		{
			name:     "Stripe Payment Integration",
			filePath: "/src/services/paymentService.js",
			content: `
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);

const createPaymentIntent = async (amount, currency = 'usd') => {
  try {
    const paymentIntent = await stripe.paymentIntents.create({
      amount: amount * 100, // Convert to cents
      currency: currency,
      metadata: { integration_check: 'accept_a_payment' }
    });
    
    return paymentIntent;
  } catch (error) {
    throw new Error('Payment failed: ' + error.message);
  }
};

module.exports = { createPaymentIntent };
			`,
			expectedType: PaymentIntegration,
			expectedRisk: CriticalRisk,
		},
		{
			name:     "Google Analytics Integration",
			filePath: "/src/utils/analytics.js",
			content: `
import { gtag } from 'ga-gtag';

const GA_TRACKING_ID = process.env.REACT_APP_GA_TRACKING_ID;

export const trackEvent = (eventName, parameters) => {
  gtag('event', eventName, {
    event_category: parameters.category,
    event_label: parameters.label,
    value: parameters.value
  });
};

export const trackPageView = (page) => {
  gtag('config', GA_TRACKING_ID, {
    page_path: page
  });
};
			`,
			expectedType: AnalyticsIntegration,
			expectedRisk: MediumRisk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := im.scanServiceIntegrations(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("scanServiceIntegrations() error = %v", err)
			}

			integrations := im.GetIntegrationsByType(tt.expectedType)
			found := false
			for _, integration := range integrations {
				if integration.FilePath == tt.filePath {
					found = true

					if integration.SecurityRisk != tt.expectedRisk {
						t.Errorf("Expected %s risk, got %s", tt.expectedRisk, integration.SecurityRisk)
					}

					// Payment integrations should always be critical risk
					if tt.expectedType == PaymentIntegration && integration.SecurityRisk != CriticalRisk {
						t.Error("Payment integrations should always be critical risk")
					}

					break
				}
			}

			if !found {
				t.Errorf("Expected to find %s integration", tt.expectedType)
			}
		})
	}
}

func TestIntegrationMapper_ScanAuthenticationSystems(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	tests := []struct {
		name         string
		filePath     string
		content      string
		expectedAuth string
		expectedRisk SecurityRiskLevel
	}{
		{
			name:     "Google OAuth Integration",
			filePath: "/src/auth/googleAuth.js",
			content: `
import { GoogleAuth } from 'google-auth-library';

const auth = new GoogleAuth({
  keyFile: process.env.GOOGLE_SERVICE_ACCOUNT_KEY,
  scopes: ['https://www.googleapis.com/auth/cloud-platform']
});

export const authenticateGoogle = async () => {
  const authClient = await auth.getClient();
  return authClient;
};
			`,
			expectedAuth: "Google",
			expectedRisk: HighRisk,
		},
		{
			name:     "JWT Token Handling",
			filePath: "/src/auth/jwt.js",
			content: `
const jwt = require('jsonwebtoken');

const generateToken = (payload) => {
  return jwt.sign(payload, process.env.JWT_SECRET, {
    expiresIn: '24h'
  });
};

const verifyToken = (token) => {
  return jwt.verify(token, process.env.JWT_SECRET);
};

module.exports = { generateToken, verifyToken };
			`,
			expectedAuth: "JWT",
			expectedRisk: HighRisk,
		},
		{
			name:     "LDAP Authentication",
			filePath: "/src/auth/ldap.js",
			content: `
const ldap = require('ldapjs');

const client = ldap.createClient({
  url: 'ldap://company.local:389'
});

const authenticate = (username, password) => {
  return new Promise((resolve, reject) => {
    client.bind(username, password, (err) => {
      if (err) {
        reject(err);
      } else {
        resolve(true);
      }
    });
  });
};

module.exports = { authenticate };
			`,
			expectedAuth: "LDAP",
			expectedRisk: HighRisk,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := im.scanAuthenticationSystems(tt.filePath, tt.content)
			if err != nil {
				t.Fatalf("scanAuthenticationSystems() error = %v", err)
			}

			integrations := im.GetIntegrationsByType(AuthIntegration)
			found := false
			for _, integration := range integrations {
				if integration.FilePath == tt.filePath {
					found = true

					if integration.SecurityRisk != tt.expectedRisk {
						t.Errorf("Expected %s risk, got %s", tt.expectedRisk, integration.SecurityRisk)
					}

					// Check that authentication integrations are properly categorized
					if !strings.Contains(integration.Name, tt.expectedAuth) {
						t.Errorf("Expected auth type %s in name %s", tt.expectedAuth, integration.Name)
					}

					break
				}
			}

			if !found {
				t.Error("Expected to find authentication integration")
			}
		})
	}
}

func TestIntegrationMapper_ScanEnvironmentVariables(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	content := `
const config = {
  port: process.env.PORT || 3000,
  databaseUrl: process.env.DATABASE_URL,
  apiKey: process.env.API_KEY,
  secretKey: process.env.SECRET_KEY,
  debugMode: process.env.DEBUG === 'true',
  stripePublishableKey: process.env.STRIPE_PUBLISHABLE_KEY
};

module.exports = config;
	`

	err := im.scanEnvironmentVariables("/src/config/config.js", content)
	if err != nil {
		t.Fatalf("scanEnvironmentVariables() error = %v", err)
	}

	integrations := im.GetIntegrationPoints()
	envVarIntegrations := 0
	highRiskEnvVars := 0

	for _, integration := range integrations {
		if integration.Type == UnknownIntegration && integration.Protocol == "Environment" {
			envVarIntegrations++

			if integration.SecurityRisk == HighRisk {
				highRiskEnvVars++
			}
		}
	}

	if envVarIntegrations < 5 { // Should detect at least 5 env vars
		t.Errorf("Expected at least 5 environment variable integrations, got %d", envVarIntegrations)
	}

	// Variables with sensitive names should be high risk
	if highRiskEnvVars < 2 { // API_KEY, SECRET_KEY should be high risk
		t.Errorf("Expected at least 2 high risk environment variables, got %d", highRiskEnvVars)
	}
}

func TestIntegrationMapper_SecurityAssessment(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	// Add some test integrations with different risk levels
	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "test1",
		Name:         "Safe API",
		SecurityRisk: LowRisk,
		Credentials: CredentialInfo{
			UsesEnvVars: true,
		},
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "test2",
		Name:         "Risky Database",
		SecurityRisk: CriticalRisk,
		Credentials: CredentialInfo{
			UsesHardcoded:  true,
			SecurityIssues: []string{"Hardcoded password"},
		},
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "test3",
		Name:         "Medium Risk API",
		SecurityRisk: MediumRisk,
	})

	assessment := im.GetSecurityAssessment()

	// Check overall risk assessment
	if overallRisk, exists := assessment["overall_risk"]; !exists {
		t.Error("Expected overall_risk in assessment")
	} else if overallRisk != "medium" && overallRisk != "high" {
		t.Errorf("Expected medium or high overall risk, got %s", overallRisk)
	}

	// Check high risk percentage
	if percentage, exists := assessment["high_risk_percentage"]; !exists {
		t.Error("Expected high_risk_percentage in assessment")
	} else if percentage.(float64) <= 0 {
		t.Error("Expected non-zero high risk percentage")
	}

	// Check recommendations
	if recommendations, exists := assessment["recommendations"]; !exists {
		t.Error("Expected recommendations in assessment")
	} else {
		recList := recommendations.([]string)
		if len(recList) == 0 {
			t.Error("Expected security recommendations")
		}

		// Should recommend fixing hardcoded credentials
		hasCredentialRec := false
		for _, rec := range recList {
			if strings.Contains(rec, "hardcoded") {
				hasCredentialRec = true
				break
			}
		}
		if !hasCredentialRec {
			t.Error("Expected recommendation about hardcoded credentials")
		}
	}
}

func TestIntegrationMapper_GetIntegrationStats(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	// Add test integrations
	im.integrations = append(im.integrations, IntegrationPoint{
		Type:         DatabaseIntegration,
		SecurityRisk: LowRisk,
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		Type:         APIIntegration,
		SecurityRisk: HighRisk,
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		Type:         PaymentIntegration,
		SecurityRisk: CriticalRisk,
	})

	stats := im.GetIntegrationStats()

	// Check total count
	if total, exists := stats["total_integrations"]; !exists || total != 3 {
		t.Errorf("Expected 3 total integrations, got %v", total)
	}

	// Check type counts
	if byType, exists := stats["by_type"]; !exists {
		t.Error("Expected by_type statistics")
	} else {
		typeCounts := byType.(map[IntegrationType]int)
		if typeCounts[DatabaseIntegration] != 1 {
			t.Errorf("Expected 1 database integration, got %d", typeCounts[DatabaseIntegration])
		}
		if typeCounts[APIIntegration] != 1 {
			t.Errorf("Expected 1 API integration, got %d", typeCounts[APIIntegration])
		}
		if typeCounts[PaymentIntegration] != 1 {
			t.Errorf("Expected 1 payment integration, got %d", typeCounts[PaymentIntegration])
		}
	}

	// Check high risk count
	if highRiskCount, exists := stats["high_risk_count"]; !exists || highRiskCount != 2 {
		t.Errorf("Expected 2 high risk integrations, got %v", highRiskCount)
	}
}

func TestIntegrationMapper_GetHighRiskIntegrations(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	// Add test integrations
	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "low_risk",
		SecurityRisk: LowRisk,
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "high_risk",
		SecurityRisk: HighRisk,
	})

	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "critical_risk",
		SecurityRisk: CriticalRisk,
	})

	highRisk := im.GetHighRiskIntegrations()

	if len(highRisk) != 2 {
		t.Errorf("Expected 2 high risk integrations, got %d", len(highRisk))
	}

	// Verify both high and critical risk are included
	foundHigh := false
	foundCritical := false

	for _, integration := range highRisk {
		if integration.ID == "high_risk" {
			foundHigh = true
		} else if integration.ID == "critical_risk" {
			foundCritical = true
		}
	}

	if !foundHigh {
		t.Error("Expected to find high risk integration")
	}
	if !foundCritical {
		t.Error("Expected to find critical risk integration")
	}
}

func TestIntegrationMapper_ComprehensiveScanning(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	// Complex file with multiple integration types
	content := `
const express = require('express');
const mongoose = require('mongoose');
const stripe = require('stripe')(process.env.STRIPE_SECRET_KEY);
const AWS = require('aws-sdk');
const jwt = require('jsonwebtoken');

// Database connection
mongoose.connect(process.env.MONGODB_URI);

// AWS S3 setup
const s3 = new AWS.S3({
  accessKeyId: process.env.AWS_ACCESS_KEY_ID,
  secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY,
  region: 'us-east-1'
});

// External API calls
const fetchExternalData = async () => {
  const response = await fetch('https://api.external.com/data', {
    headers: {
      'Authorization': 'Bearer ' + process.env.API_TOKEN
    }
  });
  return response.json();
};

// Payment processing
const processPayment = async (amount) => {
  const paymentIntent = await stripe.paymentIntents.create({
    amount: amount * 100,
    currency: 'usd'
  });
  return paymentIntent;
};

// JWT authentication
const generateToken = (user) => {
  return jwt.sign({ userId: user.id }, process.env.JWT_SECRET, {
    expiresIn: '1h'
  });
};

module.exports = { fetchExternalData, processPayment, generateToken };
	`

	err := im.MapIntegrationPoints("/src/app.js", content)
	if err != nil {
		t.Fatalf("MapIntegrationPoints() error = %v", err)
	}

	integrations := im.GetIntegrationPoints()

	// Should detect multiple types of integrations
	expectedTypes := []IntegrationType{
		DatabaseIntegration,
		CloudIntegration,
		APIIntegration,
		PaymentIntegration,
		AuthIntegration,
	}

	foundTypes := make(map[IntegrationType]bool)
	for _, integration := range integrations {
		foundTypes[integration.Type] = true
	}

	for _, expectedType := range expectedTypes {
		if !foundTypes[expectedType] {
			t.Errorf("Expected to find %s integration", expectedType)
		}
	}

	// Should have at least one high-risk integration (payment)
	highRiskIntegrations := im.GetHighRiskIntegrations()
	if len(highRiskIntegrations) == 0 {
		t.Error("Expected at least one high-risk integration")
	}

	// Payment integration should be critical risk
	paymentIntegrations := im.GetIntegrationsByType(PaymentIntegration)
	if len(paymentIntegrations) > 0 {
		if paymentIntegrations[0].SecurityRisk != CriticalRisk {
			t.Error("Expected payment integration to be critical risk")
		}
	}
}

func TestIntegrationMapper_ExportToJSON(t *testing.T) {
	ci := NewComponentIdentifier()
	dfa := NewDataFlowAnalyzer(ci)
	im := NewIntegrationMapper(ci, dfa)

	// Add test data
	im.integrations = append(im.integrations, IntegrationPoint{
		ID:           "test",
		Name:         "Test Integration",
		Type:         APIIntegration,
		SecurityRisk: MediumRisk,
		RiskReasons:  []string{"Test reason"},
	})

	jsonData, err := im.ExportToJSON()
	if err != nil {
		t.Fatalf("ExportToJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON export")
	}

	// Verify it contains expected keys
	jsonStr := string(jsonData)
	expectedKeys := []string{"integration_points", "statistics", "security_assessment"}

	for _, key := range expectedKeys {
		if !strings.Contains(jsonStr, key) {
			t.Errorf("Expected JSON to contain key: %s", key)
		}
	}
}
