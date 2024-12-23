package tunnel

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	maxTunnels := 10
	manager := NewManager(maxTunnels)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.maxTunnels != maxTunnels {
		t.Errorf("Expected maxTunnels to be %d, got %d", maxTunnels, manager.maxTunnels)
	}

	if manager.tunnels == nil {
		t.Error("Expected non-nil tunnels map")
	}
}

func TestCreateTunnel(t *testing.T) {
	manager := NewManager(2)
	
	tests := []struct {
		name        string
		id          string
		hostname    string
		targetPort  int
		wgPubKey    string
		metadata    map[string]string
		shouldError bool
	}{
		{
			name:       "Valid tunnel without WireGuard",
			id:         "test-1",
			hostname:   "test1.example.com",
			targetPort: 8080,
			metadata: map[string]string{
				"env": "test",
			},
			shouldError: false,
		},
		{
			name:       "Duplicate tunnel ID",
			id:         "test-1",
			hostname:   "test2.example.com",
			targetPort: 8081,
			shouldError: true,
		},
		{
			name:       "Valid second tunnel",
			id:         "test-2",
			hostname:   "test2.example.com",
			targetPort: 8081,
			shouldError: false,
		},
		{
			name:       "Exceeds max tunnels",
			id:         "test-3",
			hostname:   "test3.example.com",
			targetPort: 8082,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tunnel, err := manager.CreateTunnel(tt.id, tt.hostname, tt.targetPort, tt.wgPubKey, tt.metadata)

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

			if tunnel.ID != tt.id {
				t.Errorf("Expected tunnel ID %s, got %s", tt.id, tunnel.ID)
			}

			if tunnel.Hostname != tt.hostname {
				t.Errorf("Expected hostname %s, got %s", tt.hostname, tunnel.Hostname)
			}

			if tunnel.TargetPort != tt.targetPort {
				t.Errorf("Expected target port %d, got %d", tt.targetPort, tunnel.TargetPort)
			}

			// Check metadata
			for k, v := range tt.metadata {
				if tunnel.Metadata[k] != v {
					t.Errorf("Expected metadata %s=%s, got %s", k, v, tunnel.Metadata[k])
				}
			}
		})
	}
}

func TestGetTunnel(t *testing.T) {
	manager := NewManager(10)
	
	// Create a test tunnel
	testID := "test-1"
	testHostname := "test.example.com"
	testPort := 8080
	
	_, err := manager.CreateTunnel(testID, testHostname, testPort, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}

	// Test getting existing tunnel
	tunnel, err := manager.GetTunnel(testID)
	if err != nil {
		t.Errorf("Unexpected error getting tunnel: %v", err)
	}
	if tunnel.ID != testID {
		t.Errorf("Expected tunnel ID %s, got %s", testID, tunnel.ID)
	}

	// Test getting non-existent tunnel
	_, err = manager.GetTunnel("non-existent")
	if err == nil {
		t.Error("Expected error getting non-existent tunnel, got nil")
	}
}

func TestRemoveTunnel(t *testing.T) {
	manager := NewManager(10)
	
	// Create a test tunnel
	testID := "test-1"
	testHostname := "test.example.com"
	testPort := 8080
	
	_, err := manager.CreateTunnel(testID, testHostname, testPort, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}

	// Test removing existing tunnel
	err = manager.RemoveTunnel(testID)
	if err != nil {
		t.Errorf("Unexpected error removing tunnel: %v", err)
	}

	// Verify tunnel was removed
	_, err = manager.GetTunnel(testID)
	if err == nil {
		t.Error("Expected error getting removed tunnel, got nil")
	}

	// Test removing non-existent tunnel
	err = manager.RemoveTunnel("non-existent")
	if err == nil {
		t.Error("Expected error removing non-existent tunnel, got nil")
	}
}

func TestGetTunnelByHostname(t *testing.T) {
	manager := NewManager(10)
	
	// Create test tunnels
	tunnels := []struct {
		id       string
		hostname string
		port     int
	}{
		{"test-1", "test1.example.com", 8080},
		{"test-2", "test2.example.com", 8081},
	}

	for _, tt := range tunnels {
		_, err := manager.CreateTunnel(tt.id, tt.hostname, tt.port, "", nil)
		if err != nil {
			t.Fatalf("Failed to create test tunnel: %v", err)
		}
	}

	// Test getting existing tunnels by hostname
	for _, tt := range tunnels {
		tunnel, err := manager.GetTunnelByHostname(tt.hostname)
		if err != nil {
			t.Errorf("Unexpected error getting tunnel by hostname: %v", err)
			continue
		}
		if tunnel.ID != tt.id {
			t.Errorf("Expected tunnel ID %s, got %s", tt.id, tunnel.ID)
		}
	}

	// Test getting non-existent tunnel
	_, err := manager.GetTunnelByHostname("non-existent.example.com")
	if err == nil {
		t.Error("Expected error getting non-existent tunnel by hostname, got nil")
	}
}

func TestUpdateLastActive(t *testing.T) {
	manager := NewManager(10)
	
	// Create a test tunnel
	testID := "test-1"
	testHostname := "test.example.com"
	testPort := 8080
	
	tunnel, err := manager.CreateTunnel(testID, testHostname, testPort, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}

	// Record the initial last active time
	initialLastActive := tunnel.LastActive

	// Wait a short time
	time.Sleep(time.Millisecond * 100)

	// Update last active
	manager.UpdateLastActive(testID)

	// Get the tunnel again
	updatedTunnel, err := manager.GetTunnel(testID)
	if err != nil {
		t.Fatalf("Failed to get updated tunnel: %v", err)
	}

	// Verify last active was updated
	if !updatedTunnel.LastActive.After(initialLastActive) {
		t.Error("Expected LastActive to be updated")
	}
}

func TestGetAllTunnels(t *testing.T) {
	manager := NewManager(10)
	
	// Create test tunnels
	tunnels := []struct {
		id       string
		hostname string
		port     int
	}{
		{"test-1", "test1.example.com", 8080},
		{"test-2", "test2.example.com", 8081},
		{"test-3", "test3.example.com", 8082},
	}

	for _, tt := range tunnels {
		_, err := manager.CreateTunnel(tt.id, tt.hostname, tt.port, "", nil)
		if err != nil {
			t.Fatalf("Failed to create test tunnel: %v", err)
		}
	}

	// Get all tunnels
	allTunnels := manager.GetAllTunnels()

	// Verify the number of tunnels
	if len(allTunnels) != len(tunnels) {
		t.Errorf("Expected %d tunnels, got %d", len(tunnels), len(allTunnels))
	}

	// Verify each tunnel exists in the result
	for _, tt := range tunnels {
		found := false
		for _, tunnel := range allTunnels {
			if tunnel.ID == tt.id {
				found = true
				if tunnel.Hostname != tt.hostname {
					t.Errorf("Expected hostname %s, got %s", tt.hostname, tunnel.Hostname)
				}
				if tunnel.TargetPort != tt.port {
					t.Errorf("Expected port %d, got %d", tt.port, tunnel.TargetPort)
				}
				break
			}
		}
		if !found {
			t.Errorf("Tunnel %s not found in results", tt.id)
		}
	}
} 