// Package sandbox provides network isolation management for secure repository analysis
// with controlled network access patterns and comprehensive monitoring.
package sandbox

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/yenhunghuang/repo-onboarding-copilot/pkg/logger"
)

// NetworkIsolationMode represents different levels of network isolation
type NetworkIsolationMode string

const (
	// NetworkModeNone completely disables network access
	NetworkModeNone NetworkIsolationMode = "none"
	// NetworkModeGitOnly allows network access only for Git operations
	NetworkModeGitOnly NetworkIsolationMode = "git-only"
	// NetworkModeControlled allows specific hosts for controlled operations
	NetworkModeControlled NetworkIsolationMode = "controlled"
)

// NetworkConfig represents network isolation configuration
type NetworkConfig struct {
	// Isolation settings
	Mode         NetworkIsolationMode `json:"mode"`
	AllowedHosts []string             `json:"allowed_hosts"`
	AllowedPorts []int                `json:"allowed_ports"`

	// Git-specific settings
	GitTimeout        time.Duration `json:"git_timeout"`
	AllowGitProtocols []string      `json:"allowed_git_protocols"`

	// Monitoring settings
	LogConnections   bool `json:"log_connections"`
	MonitorBandwidth bool `json:"monitor_bandwidth"`

	// Security settings
	BlockPrivateIPs bool `json:"block_private_ips"`
	RequireHTTPS    bool `json:"require_https"`
}

// NetworkMonitor tracks network activity and enforces policies
type NetworkMonitor struct {
	config      *NetworkConfig
	auditLogger *logger.Logger

	// Statistics
	ConnectionsAllowed int64
	ConnectionsBlocked int64
	BytesTransferred   int64
}

// NetworkConnection represents a monitored network connection
type NetworkConnection struct {
	SourceIP      string
	DestIP        string
	DestPort      int
	Protocol      string
	Timestamp     time.Time
	BytesIn       int64
	BytesOut      int64
	Allowed       bool
	BlockedReason string
}

// NewNetworkMonitor creates a new network isolation monitor
func NewNetworkMonitor(auditLogger *logger.Logger) (*NetworkMonitor, error) {
	if auditLogger == nil {
		return nil, fmt.Errorf("audit logger cannot be nil")
	}

	config := &NetworkConfig{
		Mode:              NetworkModeGitOnly,
		AllowedHosts:      []string{"github.com", "gitlab.com", "bitbucket.org"},
		AllowedPorts:      []int{80, 443, 22}, // HTTP, HTTPS, SSH
		GitTimeout:        5 * time.Minute,
		AllowGitProtocols: []string{"https", "ssh"},
		LogConnections:    true,
		MonitorBandwidth:  true,
		BlockPrivateIPs:   true,
		RequireHTTPS:      true,
	}

	nm := &NetworkMonitor{
		config:      config,
		auditLogger: auditLogger,
	}

	auditLogger.WithFields(map[string]interface{}{
		"operation":         "network_monitor_created",
		"isolation_mode":    string(config.Mode),
		"allowed_hosts":     len(config.AllowedHosts),
		"allowed_ports":     len(config.AllowedPorts),
		"log_connections":   config.LogConnections,
		"monitor_bandwidth": config.MonitorBandwidth,
		"timestamp":         time.Now().Unix(),
	}).Info("Network isolation monitor initialized")

	return nm, nil
}

// ValidateNetworkAccess checks if a network connection should be allowed
func (nm *NetworkMonitor) ValidateNetworkAccess(destURL string, operation string) (*NetworkConnection, error) {
	startTime := time.Now()

	conn := &NetworkConnection{
		Timestamp: startTime,
		Protocol:  "unknown",
		Allowed:   false,
	}

	// Parse URL to extract host and port
	parsedURL, err := url.Parse(destURL)
	if err != nil {
		conn.BlockedReason = fmt.Sprintf("invalid URL format: %v", err)
		nm.logNetworkEvent(conn, operation, false)
		nm.ConnectionsBlocked++
		return conn, fmt.Errorf("invalid destination URL: %w", err)
	}

	conn.DestIP = parsedURL.Hostname()
	conn.Protocol = parsedURL.Scheme

	// Determine port
	port := parsedURL.Port()
	if port == "" {
		switch parsedURL.Scheme {
		case "http":
			conn.DestPort = 80
		case "https":
			conn.DestPort = 443
		case "ssh":
			conn.DestPort = 22
		default:
			conn.BlockedReason = "unknown protocol or port"
			nm.logNetworkEvent(conn, operation, false)
			nm.ConnectionsBlocked++
			return conn, fmt.Errorf("unknown protocol or port for %s", destURL)
		}
	} else {
		if _, err := fmt.Sscanf(port, "%d", &conn.DestPort); err != nil {
			conn.BlockedReason = fmt.Sprintf("invalid port: %s", port)
			nm.logNetworkEvent(conn, operation, false)
			nm.ConnectionsBlocked++
			return conn, fmt.Errorf("invalid port: %s", port)
		}
	}

	// Apply network isolation rules
	allowed, reason := nm.evaluateNetworkPolicy(conn, operation)
	conn.Allowed = allowed
	conn.BlockedReason = reason

	if allowed {
		nm.ConnectionsAllowed++
	} else {
		nm.ConnectionsBlocked++
	}

	nm.logNetworkEvent(conn, operation, allowed)
	return conn, nil
}

// evaluateNetworkPolicy applies network isolation rules
func (nm *NetworkMonitor) evaluateNetworkPolicy(conn *NetworkConnection, operation string) (bool, string) {
	// Check isolation mode
	switch nm.config.Mode {
	case NetworkModeNone:
		return false, "network access completely disabled"

	case NetworkModeGitOnly:
		// Only allow Git operations
		if operation != "git-clone" && operation != "git-fetch" && operation != "git-pull" {
			return false, fmt.Sprintf("operation '%s' not allowed in git-only mode", operation)
		}

	case NetworkModeControlled:
		// Allow controlled access based on configuration
		break

	default:
		return false, fmt.Sprintf("unknown network isolation mode: %s", nm.config.Mode)
	}

	// Check if HTTPS is required
	if nm.config.RequireHTTPS && conn.Protocol != "https" && conn.Protocol != "ssh" {
		return false, fmt.Sprintf("protocol '%s' not allowed, HTTPS required", conn.Protocol)
	}

	// Check allowed protocols for Git operations
	if strings.HasPrefix(operation, "git-") {
		protocolAllowed := false
		for _, allowedProto := range nm.config.AllowGitProtocols {
			if conn.Protocol == allowedProto {
				protocolAllowed = true
				break
			}
		}
		if !protocolAllowed {
			return false, fmt.Sprintf("Git protocol '%s' not in allowed list", conn.Protocol)
		}
	}

	// Check allowed hosts
	hostAllowed := false
	for _, allowedHost := range nm.config.AllowedHosts {
		if strings.Contains(conn.DestIP, allowedHost) || conn.DestIP == allowedHost {
			hostAllowed = true
			break
		}
	}
	if !hostAllowed {
		return false, fmt.Sprintf("host '%s' not in allowed list", conn.DestIP)
	}

	// Check allowed ports
	portAllowed := false
	for _, allowedPort := range nm.config.AllowedPorts {
		if conn.DestPort == allowedPort {
			portAllowed = true
			break
		}
	}
	if !portAllowed {
		return false, fmt.Sprintf("port %d not in allowed list", conn.DestPort)
	}

	// Check for private IP addresses if blocked
	if nm.config.BlockPrivateIPs && nm.isPrivateIP(conn.DestIP) {
		return false, "private IP addresses are blocked"
	}

	return true, ""
}

// isPrivateIP checks if an IP address is in a private range
func (nm *NetworkMonitor) isPrivateIP(hostname string) bool {
	// First resolve hostname to IP if necessary
	ip := net.ParseIP(hostname)
	if ip == nil {
		// Try to resolve hostname
		ips, err := net.LookupIP(hostname)
		if err != nil || len(ips) == 0 {
			return false // Can't resolve, let other rules handle it
		}
		ip = ips[0]
	}

	// Check private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // RFC 1918
		"172.16.0.0/12",  // RFC 1918
		"192.168.0.0/16", // RFC 1918
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// logNetworkEvent logs network connection attempts
func (nm *NetworkMonitor) logNetworkEvent(conn *NetworkConnection, operation string, allowed bool) {
	if !nm.config.LogConnections {
		return
	}

	fields := map[string]interface{}{
		"operation":     "network_connection_attempt",
		"git_operation": operation,
		"dest_host":     conn.DestIP,
		"dest_port":     conn.DestPort,
		"protocol":      conn.Protocol,
		"allowed":       allowed,
		"timestamp":     conn.Timestamp.Unix(),
	}

	if !allowed {
		fields["blocked_reason"] = conn.BlockedReason
	}

	if allowed {
		nm.auditLogger.WithFields(fields).Info("Network connection allowed")
	} else {
		nm.auditLogger.WithFields(fields).Warn("Network connection blocked")
	}
}

// CreateNetworkNamespace creates an isolated network namespace for containers
func (nm *NetworkMonitor) CreateNetworkNamespace(ctx context.Context, namespaceName string) error {
	startTime := time.Now()

	nm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "network_namespace_creation_start",
		"namespace_name": namespaceName,
		"timestamp":      startTime.Unix(),
	}).Info("Creating isolated network namespace")

	// Create network namespace
	cmd := exec.CommandContext(ctx, "ip", "netns", "add", namespaceName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		nm.auditLogger.WithFields(map[string]interface{}{
			"operation":      "network_namespace_creation_failure",
			"namespace_name": namespaceName,
			"error":          err.Error(),
			"output":         string(output),
			"execution_time": time.Since(startTime).Seconds(),
			"timestamp":      time.Now().Unix(),
		}).Error("Failed to create network namespace")
		return fmt.Errorf("failed to create network namespace %s: %w", namespaceName, err)
	}

	nm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "network_namespace_creation_success",
		"namespace_name": namespaceName,
		"execution_time": time.Since(startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Info("Network namespace created successfully")

	return nil
}

// DeleteNetworkNamespace removes an isolated network namespace
func (nm *NetworkMonitor) DeleteNetworkNamespace(ctx context.Context, namespaceName string) error {
	startTime := time.Now()

	nm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "network_namespace_deletion_start",
		"namespace_name": namespaceName,
		"timestamp":      startTime.Unix(),
	}).Info("Deleting network namespace")

	// Delete network namespace
	cmd := exec.CommandContext(ctx, "ip", "netns", "del", namespaceName)
	output, err := cmd.CombinedOutput()

	if err != nil {
		nm.auditLogger.WithFields(map[string]interface{}{
			"operation":      "network_namespace_deletion_failure",
			"namespace_name": namespaceName,
			"error":          err.Error(),
			"output":         string(output),
			"execution_time": time.Since(startTime).Seconds(),
			"timestamp":      time.Now().Unix(),
		}).Error("Failed to delete network namespace")
		return fmt.Errorf("failed to delete network namespace %s: %w", namespaceName, err)
	}

	nm.auditLogger.WithFields(map[string]interface{}{
		"operation":      "network_namespace_deletion_success",
		"namespace_name": namespaceName,
		"execution_time": time.Since(startTime).Seconds(),
		"timestamp":      time.Now().Unix(),
	}).Info("Network namespace deleted successfully")

	return nil
}

// GetNetworkStatistics returns current network monitoring statistics
func (nm *NetworkMonitor) GetNetworkStatistics() map[string]interface{} {
	return map[string]interface{}{
		"connections_allowed": nm.ConnectionsAllowed,
		"connections_blocked": nm.ConnectionsBlocked,
		"bytes_transferred":   nm.BytesTransferred,
		"isolation_mode":      string(nm.config.Mode),
		"allowed_hosts":       nm.config.AllowedHosts,
		"allowed_ports":       nm.config.AllowedPorts,
		"timestamp":           time.Now().Unix(),
	}
}

// SetNetworkConfig updates the network configuration
func (nm *NetworkMonitor) SetNetworkConfig(config *NetworkConfig) error {
	if config == nil {
		return fmt.Errorf("network config cannot be nil")
	}

	// Validate configuration
	if err := nm.validateNetworkConfig(config); err != nil {
		return fmt.Errorf("invalid network configuration: %w", err)
	}

	nm.config = config

	nm.auditLogger.WithFields(map[string]interface{}{
		"operation":         "network_config_updated",
		"isolation_mode":    string(config.Mode),
		"allowed_hosts":     len(config.AllowedHosts),
		"allowed_ports":     len(config.AllowedPorts),
		"git_timeout":       config.GitTimeout.Seconds(),
		"log_connections":   config.LogConnections,
		"monitor_bandwidth": config.MonitorBandwidth,
		"timestamp":         time.Now().Unix(),
	}).Info("Network configuration updated")

	return nil
}

// validateNetworkConfig validates network configuration parameters
func (nm *NetworkMonitor) validateNetworkConfig(config *NetworkConfig) error {
	// Validate isolation mode
	validModes := []NetworkIsolationMode{NetworkModeNone, NetworkModeGitOnly, NetworkModeControlled}
	modeValid := false
	for _, mode := range validModes {
		if config.Mode == mode {
			modeValid = true
			break
		}
	}
	if !modeValid {
		return fmt.Errorf("invalid isolation mode: %s", config.Mode)
	}

	// Validate timeout
	if config.GitTimeout <= 0 {
		return fmt.Errorf("git timeout must be positive, got: %v", config.GitTimeout)
	}

	// Validate allowed protocols
	validProtocols := []string{"http", "https", "ssh", "git"}
	for _, proto := range config.AllowGitProtocols {
		protoValid := false
		for _, validProto := range validProtocols {
			if proto == validProto {
				protoValid = true
				break
			}
		}
		if !protoValid {
			return fmt.Errorf("invalid Git protocol: %s", proto)
		}
	}

	// Validate ports
	for _, port := range config.AllowedPorts {
		if port <= 0 || port > 65535 {
			return fmt.Errorf("invalid port number: %d", port)
		}
	}

	return nil
}

// GetNetworkConfig returns the current network configuration
func (nm *NetworkMonitor) GetNetworkConfig() *NetworkConfig {
	return nm.config
}
