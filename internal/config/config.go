// Package config provides configuration management for the easy-tunnel-lb-agent.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServerConfig holds all configuration for the server agent
type ServerConfig struct {
	// API Server settings
	APIPort     int
	APIHost     string
	APIBasePath string

	// Public Load Balancer settings
	PublicPort int
	PublicHost string
	
	// TLS Configuration
	TLSCertPath string
	TLSKeyPath  string

	// Tunnel settings
	MaxTunnels int

	// Logging
	LogLevel string

	// Server shutdown timeout
	ShutdownTimeout time.Duration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*ServerConfig, error) {
	config := &ServerConfig{
		APIPort:     getEnvInt("API_PORT", 8080),
		APIHost:     getEnvStr("API_HOST", "0.0.0.0"),
		APIBasePath: getEnvStr("API_BASE_PATH", "/api"),
		PublicPort:  getEnvInt("PUBLIC_PORT", 443),
		PublicHost:  getEnvStr("PUBLIC_HOST", "0.0.0.0"),
		TLSCertPath: getEnvStr("TLS_CERT_PATH", ""),
		TLSKeyPath:  getEnvStr("TLS_KEY_PATH", ""),
		MaxTunnels:  getEnvInt("MAX_TUNNELS", 100),
		LogLevel:    getEnvStr("LOG_LEVEL", "info"),
		ShutdownTimeout: time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_SECONDS", 30)) * time.Second,
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate checks if the configuration is valid
func (c *ServerConfig) validate() error {
	if c.APIPort <= 0 || c.APIPort > 65535 {
		return fmt.Errorf("invalid API port: %d", c.APIPort)
	}

	if c.PublicPort <= 0 || c.PublicPort > 65535 {
		return fmt.Errorf("invalid public port: %d", c.PublicPort)
	}

	// If TLS is configured, both cert and key must be provided
	if (c.TLSCertPath != "" && c.TLSKeyPath == "") || (c.TLSCertPath == "" && c.TLSKeyPath != "") {
		return fmt.Errorf("both TLS certificate and key must be provided")
	}

	return nil
}

// Helper functions to get environment variables
func getEnvStr(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
} 