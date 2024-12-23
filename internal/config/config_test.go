package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"API_PORT",
		"API_HOST",
		"API_BASE_PATH",
		"PUBLIC_PORT",
		"PUBLIC_HOST",
		"TLS_CERT_PATH",
		"TLS_KEY_PATH",
		"MAX_TUNNELS",
		"LOG_LEVEL",
		"SHUTDOWN_TIMEOUT_SECONDS",
	}

	for _, env := range envVars {
		if value, exists := os.LookupEnv(env); exists {
			originalEnv[env] = value
			os.Unsetenv(env)
		}
	}

	// Restore environment after test
	defer func() {
		for env, value := range originalEnv {
			os.Setenv(env, value)
		}
	}()

	// Test default configuration
	t.Run("Default Configuration", func(t *testing.T) {
		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load default config: %v", err)
		}

		// Check default values
		if config.APIPort != 8080 {
			t.Errorf("Expected default API port 8080, got %d", config.APIPort)
		}
		if config.APIHost != "0.0.0.0" {
			t.Errorf("Expected default API host 0.0.0.0, got %s", config.APIHost)
		}
		if config.APIBasePath != "/api" {
			t.Errorf("Expected default API base path /api, got %s", config.APIBasePath)
		}
		if config.PublicPort != 443 {
			t.Errorf("Expected default public port 443, got %d", config.PublicPort)
		}
		if config.MaxTunnels != 100 {
			t.Errorf("Expected default max tunnels 100, got %d", config.MaxTunnels)
		}
		if config.LogLevel != "info" {
			t.Errorf("Expected default log level info, got %s", config.LogLevel)
		}
		if config.ShutdownTimeout != 30*time.Second {
			t.Errorf("Expected default shutdown timeout 30s, got %v", config.ShutdownTimeout)
		}
	})

	// Test custom configuration
	t.Run("Custom Configuration", func(t *testing.T) {
		customConfig := map[string]string{
			"API_PORT":                 "9090",
			"API_HOST":                 "127.0.0.1",
			"API_BASE_PATH":            "/custom",
			"PUBLIC_PORT":              "8443",
			"PUBLIC_HOST":              "example.com",
			"TLS_CERT_PATH":            "/path/to/cert.pem",
			"TLS_KEY_PATH":             "/path/to/key.pem",
			"MAX_TUNNELS":              "50",
			"LOG_LEVEL":                "debug",
			"SHUTDOWN_TIMEOUT_SECONDS": "60",
		}

		// Set custom environment variables
		for key, value := range customConfig {
			os.Setenv(key, value)
		}

		config, err := LoadConfig()
		if err != nil {
			t.Fatalf("Failed to load custom config: %v", err)
		}

		// Check custom values
		if config.APIPort != 9090 {
			t.Errorf("Expected API port 9090, got %d", config.APIPort)
		}
		if config.APIHost != "127.0.0.1" {
			t.Errorf("Expected API host 127.0.0.1, got %s", config.APIHost)
		}
		if config.APIBasePath != "/custom" {
			t.Errorf("Expected API base path /custom, got %s", config.APIBasePath)
		}
		if config.PublicPort != 8443 {
			t.Errorf("Expected public port 8443, got %d", config.PublicPort)
		}
		if config.PublicHost != "example.com" {
			t.Errorf("Expected public host example.com, got %s", config.PublicHost)
		}
		if config.TLSCertPath != "/path/to/cert.pem" {
			t.Errorf("Expected TLS cert path /path/to/cert.pem, got %s", config.TLSCertPath)
		}
		if config.TLSKeyPath != "/path/to/key.pem" {
			t.Errorf("Expected TLS key path /path/to/key.pem, got %s", config.TLSKeyPath)
		}
		if config.MaxTunnels != 50 {
			t.Errorf("Expected max tunnels 50, got %d", config.MaxTunnels)
		}
		if config.LogLevel != "debug" {
			t.Errorf("Expected log level debug, got %s", config.LogLevel)
		}
		if config.ShutdownTimeout != 60*time.Second {
			t.Errorf("Expected shutdown timeout 60s, got %v", config.ShutdownTimeout)
		}
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *ServerConfig
		shouldError bool
	}{
		{
			name: "Valid configuration",
			config: &ServerConfig{
				APIPort:     8080,
				PublicPort:  443,
				MaxTunnels:  100,
				LogLevel:    "info",
			},
			shouldError: false,
		},
		{
			name: "Invalid API port",
			config: &ServerConfig{
				APIPort:     -1,
				PublicPort:  443,
				MaxTunnels:  100,
				LogLevel:    "info",
			},
			shouldError: true,
		},
		{
			name: "Invalid public port",
			config: &ServerConfig{
				APIPort:     8080,
				PublicPort:  70000,
				MaxTunnels:  100,
				LogLevel:    "info",
			},
			shouldError: true,
		},
		{
			name: "Missing TLS key",
			config: &ServerConfig{
				APIPort:     8080,
				PublicPort:  443,
				MaxTunnels:  100,
				LogLevel:    "info",
				TLSCertPath: "/path/to/cert.pem",
			},
			shouldError: true,
		},
		{
			name: "Missing TLS cert",
			config: &ServerConfig{
				APIPort:     8080,
				PublicPort:  443,
				MaxTunnels:  100,
				LogLevel:    "info",
				TLSKeyPath:  "/path/to/key.pem",
			},
			shouldError: true,
		},
		{
			name: "Valid TLS configuration",
			config: &ServerConfig{
				APIPort:     8080,
				PublicPort:  443,
				MaxTunnels:  100,
				LogLevel:    "info",
				TLSCertPath: "/path/to/cert.pem",
				TLSKeyPath:  "/path/to/key.pem",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.shouldError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestGetEnvHelpers(t *testing.T) {
	// Test getEnvStr
	t.Run("getEnvStr", func(t *testing.T) {
		key := "TEST_ENV_STR"
		defaultVal := "default"
		customVal := "custom"

		// Test default value
		if val := getEnvStr(key, defaultVal); val != defaultVal {
			t.Errorf("Expected default value %s, got %s", defaultVal, val)
		}

		// Test custom value
		os.Setenv(key, customVal)
		if val := getEnvStr(key, defaultVal); val != customVal {
			t.Errorf("Expected custom value %s, got %s", customVal, val)
		}
		os.Unsetenv(key)
	})

	// Test getEnvInt
	t.Run("getEnvInt", func(t *testing.T) {
		key := "TEST_ENV_INT"
		defaultVal := 123
		customVal := 456

		// Test default value
		if val := getEnvInt(key, defaultVal); val != defaultVal {
			t.Errorf("Expected default value %d, got %d", defaultVal, val)
		}

		// Test custom value
		os.Setenv(key, "456")
		if val := getEnvInt(key, defaultVal); val != customVal {
			t.Errorf("Expected custom value %d, got %d", customVal, val)
		}

		// Test invalid value
		os.Setenv(key, "invalid")
		if val := getEnvInt(key, defaultVal); val != defaultVal {
			t.Errorf("Expected default value %d for invalid input, got %d", defaultVal, val)
		}
		os.Unsetenv(key)
	})
} 