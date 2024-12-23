// Package tunnel provides tunnel management functionality for the easy-tunnel-lb-agent.
package tunnel

import (
	"fmt"
	"sync"
	"time"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/utils"
	"github.com/rs/zerolog"
)

// TunnelInfo represents information about a single tunnel
type TunnelInfo struct {
	ID              string
	Hostname        string
	TargetPort      int
	PublicEndpoint  string
	Created         time.Time
	LastActive      time.Time
	WireGuardConfig *WireGuardConfig
	Metadata        map[string]string
}

// WireGuardConfig contains WireGuard-specific configuration
type WireGuardConfig struct {
	PublicKey  string
	PrivateKey string
	ServerIP   string
	ClientIP   string
	Port       int
}

// Manager handles the lifecycle of tunnels
type Manager struct {
	tunnels    map[string]*TunnelInfo
	mu         sync.RWMutex
	maxTunnels int
	logger     *zerolog.Logger
	wg         *WireGuardManager
}

// NewManager creates a new tunnel manager
func NewManager(maxTunnels int) *Manager {
	logger := utils.GetLogger()
	return &Manager{
		tunnels:    make(map[string]*TunnelInfo),
		maxTunnels: maxTunnels,
		logger:     logger,
		wg:         NewWireGuardManager(),
	}
}

// CreateTunnel creates a new tunnel with the given configuration
func (m *Manager) CreateTunnel(id, hostname string, targetPort int, wgPubKey string, metadata map[string]string) (*TunnelInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we've reached the maximum number of tunnels
	if len(m.tunnels) >= m.maxTunnels {
		return nil, fmt.Errorf("maximum number of tunnels (%d) reached", m.maxTunnels)
	}

	// Check if tunnel ID already exists
	if _, exists := m.tunnels[id]; exists {
		return nil, fmt.Errorf("tunnel with ID %s already exists", id)
	}

	tunnel := &TunnelInfo{
		ID:         id,
		Hostname:   hostname,
		TargetPort: targetPort,
		Created:    time.Now(),
		LastActive: time.Now(),
		Metadata:   metadata,
	}

	// If WireGuard public key is provided, set up WireGuard
	if wgPubKey != "" {
		wgConfig, err := m.wg.SetupPeer(id, wgPubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to setup WireGuard peer: %v", err)
		}
		tunnel.WireGuardConfig = wgConfig
	}

	m.tunnels[id] = tunnel
	m.logger.Info().
		Str("tunnel_id", id).
		Str("hostname", hostname).
		Int("target_port", targetPort).
		Msg("Created new tunnel")

	return tunnel, nil
}

// RemoveTunnel removes an existing tunnel
func (m *Manager) RemoveTunnel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return fmt.Errorf("tunnel with ID %s not found", id)
	}

	// If it's a WireGuard tunnel, remove the peer
	if tunnel.WireGuardConfig != nil {
		if err := m.wg.RemovePeer(id); err != nil {
			m.logger.Error().
				Err(err).
				Str("tunnel_id", id).
				Msg("Failed to remove WireGuard peer")
		}
	}

	delete(m.tunnels, id)
	m.logger.Info().
		Str("tunnel_id", id).
		Msg("Removed tunnel")

	return nil
}

// GetTunnel retrieves information about a specific tunnel
func (m *Manager) GetTunnel(id string) (*TunnelInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnel, exists := m.tunnels[id]
	if !exists {
		return nil, fmt.Errorf("tunnel with ID %s not found", id)
	}

	return tunnel, nil
}

// GetTunnelByHostname retrieves a tunnel by its hostname
func (m *Manager) GetTunnelByHostname(hostname string) (*TunnelInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tunnel := range m.tunnels {
		if tunnel.Hostname == hostname {
			return tunnel, nil
		}
	}

	return nil, fmt.Errorf("no tunnel found for hostname %s", hostname)
}

// UpdateLastActive updates the last active timestamp for a tunnel
func (m *Manager) UpdateLastActive(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tunnel, exists := m.tunnels[id]; exists {
		tunnel.LastActive = time.Now()
	}
}

// GetAllTunnels returns a list of all active tunnels
func (m *Manager) GetAllTunnels() []*TunnelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tunnels := make([]*TunnelInfo, 0, len(m.tunnels))
	for _, tunnel := range m.tunnels {
		tunnels = append(tunnels, tunnel)
	}

	return tunnels
} 