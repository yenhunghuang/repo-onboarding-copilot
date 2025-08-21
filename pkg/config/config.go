// Package config provides configuration management for the application.
// It handles loading and validation of YAML configuration files for
// different environments (development, production, testing).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration structure
type Config struct {
	// Application settings
	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Debug   bool   `yaml:"debug"`
	} `yaml:"app"`

	// Logging configuration
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	// Security settings
	Security struct {
		MaxURLLength    int      `yaml:"max_url_length"`
		AllowedSchemes  []string `yaml:"allowed_schemes"`
		EnableSanitization bool  `yaml:"enable_sanitization"`
	} `yaml:"security"`
}

// Load loads configuration from the specified file
func Load(configFile string) (*Config, error) {
	// Set default values
	config := &Config{}
	config.setDefaults()

	// Read configuration file
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// LoadFromEnv loads configuration based on environment
func LoadFromEnv(env string) (*Config, error) {
	configFile := filepath.Join("configs", fmt.Sprintf("%s.yaml", env))
	return Load(configFile)
}

// setDefaults sets default configuration values
func (c *Config) setDefaults() {
	c.App.Name = "repo-onboarding-copilot"
	c.App.Version = "1.0.0"
	c.App.Debug = false
	
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	
	c.Security.MaxURLLength = 2048
	c.Security.AllowedSchemes = []string{"http", "https", "git", "ssh"}
	c.Security.EnableSanitization = true
}

// Validate validates the configuration settings
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name cannot be empty")
	}
	
	if c.Security.MaxURLLength <= 0 {
		return fmt.Errorf("security.max_url_length must be positive")
	}
	
	if len(c.Security.AllowedSchemes) == 0 {
		return fmt.Errorf("security.allowed_schemes cannot be empty")
	}
	
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}
	
	return nil
}