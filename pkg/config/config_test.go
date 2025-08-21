package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name         string
		configData   string
		expectError  bool
		validateFunc func(*testing.T, *Config)
	}{
		{
			name:        "load with empty file path",
			configData:  "",
			expectError: false,
			validateFunc: func(t *testing.T, c *Config) {
				// Should load with defaults
				assert.Equal(t, "repo-onboarding-copilot", c.App.Name)
				assert.Equal(t, "info", c.Logging.Level)
				assert.Equal(t, 2048, c.Security.MaxURLLength)
				assert.Contains(t, c.Security.AllowedSchemes, "https")
			},
		},
		{
			name: "load valid config",
			configData: `
app:
  name: "test-app"
  version: "2.0.0"
  debug: true
logging:
  level: "debug"
  format: "json"
security:
  max_url_length: 1024
  allowed_schemes: ["https", "ssh"]
  enable_sanitization: true
`,
			expectError: false,
			validateFunc: func(t *testing.T, c *Config) {
				assert.Equal(t, "test-app", c.App.Name)
				assert.Equal(t, "2.0.0", c.App.Version)
				assert.True(t, c.App.Debug)
				assert.Equal(t, "debug", c.Logging.Level)
				assert.Equal(t, 1024, c.Security.MaxURLLength)
				assert.Equal(t, []string{"https", "ssh"}, c.Security.AllowedSchemes)
			},
		},
		{
			name: "invalid yaml",
			configData: `
app:
  name: "test
  invalid yaml
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configFile string
			
			if tt.configData != "" {
				// Create temporary config file
				tmpDir := t.TempDir()
				configFile = filepath.Join(tmpDir, "test-config.yaml")
				err := os.WriteFile(configFile, []byte(tt.configData), 0644)
				require.NoError(t, err)
			}

			config, err := Load(configFile)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.validateFunc != nil {
					tt.validateFunc(t, config)
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{},
			expectError: false,
		},
		{
			name: "empty app name",
			config: func() *Config {
				c := &Config{}
				c.setDefaults()
				c.App.Name = ""
				return c
			}(),
			expectError: true,
		},
		{
			name: "invalid max url length",
			config: func() *Config {
				c := &Config{}
				c.setDefaults()
				c.Security.MaxURLLength = 0
				return c
			}(),
			expectError: true,
		},
		{
			name: "empty allowed schemes",
			config: func() *Config {
				c := &Config{}
				c.setDefaults()
				c.Security.AllowedSchemes = []string{}
				return c
			}(),
			expectError: true,
		},
		{
			name: "invalid logging level",
			config: func() *Config {
				c := &Config{}
				c.setDefaults()
				c.Logging.Level = "invalid"
				return c
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.setDefaults()
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}