package sandbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

func TestNewNetworkMonitor(t *testing.T) {
	tests := []struct {
		name        string
		auditLogger *logger.Logger
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid logger",
			auditLogger: logger.New(),
			wantErr:     false,
		},
		{
			name:        "nil logger",
			auditLogger: nil,
			wantErr:     true,
			errMsg:      "audit logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm, err := NewNetworkMonitor(tt.auditLogger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, nm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, nm)

				// Verify default configuration
				config := nm.GetNetworkConfig()
				assert.Equal(t, NetworkModeGitOnly, config.Mode)
				assert.Contains(t, config.AllowedHosts, "github.com")
				assert.Contains(t, config.AllowedPorts, 443)
				assert.True(t, config.LogConnections)
				assert.True(t, config.MonitorBandwidth)
				assert.True(t, config.BlockPrivateIPs)
				assert.True(t, config.RequireHTTPS)
			}
		})
	}
}

func TestValidateNetworkAccess(t *testing.T) {
	auditLogger := logger.New()
	nm, err := NewNetworkMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name        string
		destURL     string
		operation   string
		wantAllowed bool
		wantReason  string
	}{
		{
			name:        "allowed GitHub HTTPS",
			destURL:     "https://github.com/user/repo.git",
			operation:   "git-clone",
			wantAllowed: true,
			wantReason:  "",
		},
		{
			name:        "blocked HTTP on GitHub",
			destURL:     "http://github.com/user/repo.git",
			operation:   "git-clone",
			wantAllowed: false,
			wantReason:  "protocol 'http' not allowed, HTTPS required",
		},
		{
			name:        "blocked unknown host",
			destURL:     "https://malicious.com/repo.git",
			operation:   "git-clone",
			wantAllowed: false,
			wantReason:  "host 'malicious.com' not in allowed list",
		},
		{
			name:        "blocked non-git operation",
			destURL:     "https://github.com/user/repo.git",
			operation:   "web-request",
			wantAllowed: false,
			wantReason:  "operation 'web-request' not allowed in git-only mode",
		},
		{
			name:        "invalid URL format",
			destURL:     "not-a-valid-url",
			operation:   "git-clone",
			wantAllowed: false,
			wantReason:  "unknown protocol or port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := nm.ValidateNetworkAccess(tt.destURL, tt.operation)

			if tt.wantAllowed {
				assert.NoError(t, err)
				assert.True(t, conn.Allowed)
				assert.Empty(t, conn.BlockedReason)
			} else {
				assert.False(t, conn.Allowed)
				assert.Contains(t, conn.BlockedReason, tt.wantReason)
			}
		})
	}
}

func TestNetworkIsolationModes(t *testing.T) {
	auditLogger := logger.New()
	nm, err := NewNetworkMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name        string
		mode        NetworkIsolationMode
		destURL     string
		operation   string
		wantAllowed bool
	}{
		{
			name:        "none mode blocks everything",
			mode:        NetworkModeNone,
			destURL:     "https://github.com/user/repo.git",
			operation:   "git-clone",
			wantAllowed: false,
		},
		{
			name:        "git-only allows git operations",
			mode:        NetworkModeGitOnly,
			destURL:     "https://github.com/user/repo.git",
			operation:   "git-clone",
			wantAllowed: true,
		},
		{
			name:        "controlled mode allows configured access",
			mode:        NetworkModeControlled,
			destURL:     "https://github.com/user/repo.git",
			operation:   "web-request",
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update network config
			config := nm.GetNetworkConfig()
			config.Mode = tt.mode
			err := nm.SetNetworkConfig(config)
			require.NoError(t, err)

			conn, _ := nm.ValidateNetworkAccess(tt.destURL, tt.operation)
			assert.Equal(t, tt.wantAllowed, conn.Allowed, "Expected allowed=%v for mode %s", tt.wantAllowed, tt.mode)
		})
	}
}

func TestValidateNetworkConfig(t *testing.T) {
	auditLogger := logger.New()
	nm, err := NewNetworkMonitor(auditLogger)
	require.NoError(t, err)

	tests := []struct {
		name    string
		config  *NetworkConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &NetworkConfig{
				Mode:              NetworkModeGitOnly,
				AllowedHosts:      []string{"github.com"},
				AllowedPorts:      []int{443},
				GitTimeout:        5 * time.Minute,
				AllowGitProtocols: []string{"https"},
				LogConnections:    true,
				MonitorBandwidth:  true,
				BlockPrivateIPs:   true,
				RequireHTTPS:      true,
			},
			wantErr: false,
		},
		{
			name: "invalid isolation mode",
			config: &NetworkConfig{
				Mode:              "invalid-mode",
				GitTimeout:        5 * time.Minute,
				AllowGitProtocols: []string{"https"},
			},
			wantErr: true,
			errMsg:  "invalid isolation mode",
		},
		{
			name: "invalid git timeout",
			config: &NetworkConfig{
				Mode:              NetworkModeGitOnly,
				GitTimeout:        -1 * time.Minute,
				AllowGitProtocols: []string{"https"},
			},
			wantErr: true,
			errMsg:  "git timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := nm.SetNetworkConfig(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetNetworkStatistics(t *testing.T) {
	auditLogger := logger.New()
	nm, err := NewNetworkMonitor(auditLogger)
	require.NoError(t, err)

	// Perform some network access validations to generate statistics
	nm.ValidateNetworkAccess("https://github.com/user/repo.git", "git-clone") // Should be allowed
	nm.ValidateNetworkAccess("http://github.com/user/repo.git", "git-clone")  // Should be blocked

	stats := nm.GetNetworkStatistics()

	assert.Contains(t, stats, "connections_allowed")
	assert.Contains(t, stats, "connections_blocked")
	assert.Contains(t, stats, "isolation_mode")
	assert.Equal(t, int64(1), stats["connections_allowed"])
	assert.Equal(t, int64(1), stats["connections_blocked"])
}
