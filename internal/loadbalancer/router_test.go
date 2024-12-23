package loadbalancer

import (
	"testing"
)

func TestNewRouter(t *testing.T) {
	config := &Config{
		HTTPPort: 8080,
		TCPPort:  8081,
	}

	router := NewRouter(config)

	if router == nil {
		t.Fatal("Expected non-nil router")
	}

	if router.config != config {
		t.Error("Expected router to store config reference")
	}

	if router.hostMap == nil {
		t.Error("Expected non-nil hostMap")
	}

	if router.portMap == nil {
		t.Error("Expected non-nil portMap")
	}
}

func TestAddRoute(t *testing.T) {
	router := NewRouter(&Config{})

	tests := []struct {
		name        string
		tunnelID    string
		hostname    string
		ip          string
		port        int
		shouldError bool
	}{
		{
			name:        "Valid route",
			tunnelID:    "test-1",
			hostname:    "test1.example.com",
			ip:          "10.0.0.1",
			port:        8080,
			shouldError: false,
		},
		{
			name:        "Duplicate hostname",
			tunnelID:    "test-2",
			hostname:    "test1.example.com",
			ip:          "10.0.0.2",
			port:        8081,
			shouldError: true,
		},
		{
			name:        "Duplicate port",
			tunnelID:    "test-3",
			hostname:    "test3.example.com",
			ip:          "10.0.0.3",
			port:        8080,
			shouldError: true,
		},
		{
			name:        "Valid route with different hostname and port",
			tunnelID:    "test-4",
			hostname:    "test4.example.com",
			ip:          "10.0.0.4",
			port:        8082,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := router.AddRoute(tt.tunnelID, tt.hostname, tt.ip, tt.port)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify host mapping
			target, err := router.GetTunnelByHost(tt.hostname)
			if err != nil {
				t.Errorf("Failed to get tunnel by hostname: %v", err)
				return
			}

			if target.ID != tt.tunnelID {
				t.Errorf("Expected tunnel ID %s, got %s", tt.tunnelID, target.ID)
			}

			if target.IP != tt.ip {
				t.Errorf("Expected IP %s, got %s", tt.ip, target.IP)
			}

			if target.Port != tt.port {
				t.Errorf("Expected port %d, got %d", tt.port, target.Port)
			}

			// Verify port mapping
			target, err = router.GetTunnelByPort(tt.port)
			if err != nil {
				t.Errorf("Failed to get tunnel by port: %v", err)
				return
			}

			if target.ID != tt.tunnelID {
				t.Errorf("Expected tunnel ID %s, got %s", tt.tunnelID, target.ID)
			}
		})
	}
}

func TestRemoveRoute(t *testing.T) {
	router := NewRouter(&Config{})

	// Add a test route
	tunnelID := "test-1"
	hostname := "test.example.com"
	ip := "10.0.0.1"
	port := 8080

	err := router.AddRoute(tunnelID, hostname, ip, port)
	if err != nil {
		t.Fatalf("Failed to add test route: %v", err)
	}

	// Remove the route
	router.RemoveRoute(tunnelID)

	// Verify route was removed from hostname mapping
	_, err = router.GetTunnelByHost(hostname)
	if err == nil {
		t.Error("Expected error getting removed tunnel by hostname, got nil")
	}

	// Verify route was removed from port mapping
	_, err = router.GetTunnelByPort(port)
	if err == nil {
		t.Error("Expected error getting removed tunnel by port, got nil")
	}
}

func TestGetTunnelByHost(t *testing.T) {
	router := NewRouter(&Config{})

	// Add test routes
	routes := []struct {
		tunnelID string
		hostname string
		ip       string
		port     int
	}{
		{"test-1", "test1.example.com", "10.0.0.1", 8080},
		{"test-2", "test2.example.com", "10.0.0.2", 8081},
	}

	for _, r := range routes {
		err := router.AddRoute(r.tunnelID, r.hostname, r.ip, r.port)
		if err != nil {
			t.Fatalf("Failed to add test route: %v", err)
		}
	}

	// Test getting existing routes
	for _, r := range routes {
		target, err := router.GetTunnelByHost(r.hostname)
		if err != nil {
			t.Errorf("Unexpected error getting tunnel by hostname: %v", err)
			continue
		}

		if target.ID != r.tunnelID {
			t.Errorf("Expected tunnel ID %s, got %s", r.tunnelID, target.ID)
		}

		if target.IP != r.ip {
			t.Errorf("Expected IP %s, got %s", r.ip, target.IP)
		}

		if target.Port != r.port {
			t.Errorf("Expected port %d, got %d", r.port, target.Port)
		}
	}

	// Test getting non-existent route
	_, err := router.GetTunnelByHost("non-existent.example.com")
	if err == nil {
		t.Error("Expected error getting non-existent tunnel by hostname, got nil")
	}
}

func TestGetTunnelByPort(t *testing.T) {
	router := NewRouter(&Config{})

	// Add test routes
	routes := []struct {
		tunnelID string
		hostname string
		ip       string
		port     int
	}{
		{"test-1", "test1.example.com", "10.0.0.1", 8080},
		{"test-2", "test2.example.com", "10.0.0.2", 8081},
	}

	for _, r := range routes {
		err := router.AddRoute(r.tunnelID, r.hostname, r.ip, r.port)
		if err != nil {
			t.Fatalf("Failed to add test route: %v", err)
		}
	}

	// Test getting existing routes
	for _, r := range routes {
		target, err := router.GetTunnelByPort(r.port)
		if err != nil {
			t.Errorf("Unexpected error getting tunnel by port: %v", err)
			continue
		}

		if target.ID != r.tunnelID {
			t.Errorf("Expected tunnel ID %s, got %s", r.tunnelID, target.ID)
		}

		if target.IP != r.ip {
			t.Errorf("Expected IP %s, got %s", r.ip, target.IP)
		}

		if target.Port != r.port {
			t.Errorf("Expected port %d, got %d", r.port, target.Port)
		}
	}

	// Test getting non-existent route
	_, err := router.GetTunnelByPort(9999)
	if err == nil {
		t.Error("Expected error getting non-existent tunnel by port, got nil")
	}
}

func TestListRoutes(t *testing.T) {
	router := NewRouter(&Config{})

	// Add test routes
	routes := []struct {
		tunnelID string
		hostname string
		ip       string
		port     int
	}{
		{"test-1", "test1.example.com", "10.0.0.1", 8080},
		{"test-2", "test2.example.com", "10.0.0.2", 8081},
		{"test-3", "test3.example.com", "10.0.0.3", 8082},
	}

	for _, r := range routes {
		err := router.AddRoute(r.tunnelID, r.hostname, r.ip, r.port)
		if err != nil {
			t.Fatalf("Failed to add test route: %v", err)
		}
	}

	// Get all routes
	allRoutes := router.ListRoutes()

	// Verify the number of routes
	if len(allRoutes) != len(routes) {
		t.Errorf("Expected %d routes, got %d", len(routes), len(allRoutes))
	}

	// Verify each route exists in the result
	for _, r := range routes {
		target, exists := allRoutes[r.hostname]
		if !exists {
			t.Errorf("Route for hostname %s not found in results", r.hostname)
			continue
		}

		if target.ID != r.tunnelID {
			t.Errorf("Expected tunnel ID %s, got %s", r.tunnelID, target.ID)
		}

		if target.IP != r.ip {
			t.Errorf("Expected IP %s, got %s", r.ip, target.IP)
		}

		if target.Port != r.port {
			t.Errorf("Expected port %d, got %d", r.port, target.Port)
		}
	}
} 