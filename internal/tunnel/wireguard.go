// Package tunnel provides tunnel management functionality for the easy-tunnel-lb-agent.
package tunnel

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/utils"
	"github.com/rs/zerolog"
)

// WireGuardManager manages WireGuard interfaces and peers
type WireGuardManager struct {
	mu           sync.RWMutex
	logger       *zerolog.Logger
	interfaceName string
	basePort     int
	ipNet        *net.IPNet
	nextIP       net.IP
}

// NewWireGuardManager creates a new WireGuard manager
func NewWireGuardManager() *WireGuardManager {
	logger := utils.GetLogger()
	_, ipNet, _ := net.ParseCIDR("10.10.0.0/16")
	nextIP := net.ParseIP("10.10.0.1")

	return &WireGuardManager{
		logger:       logger,
		interfaceName: "wg0",
		basePort:     51820,
		ipNet:        ipNet,
		nextIP:       nextIP,
	}
}

// SetupPeer creates a new WireGuard peer
func (w *WireGuardManager) SetupPeer(id string, publicKey string) (*WireGuardConfig, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Generate private/public key pair for the server
	privKey, err := w.generatePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	pubKey, err := w.generatePublicKey(privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %v", err)
	}

	// Allocate IP for the peer
	peerIP := w.allocateIP()
	if peerIP == nil {
		return nil, fmt.Errorf("failed to allocate IP for peer")
	}

	config := &WireGuardConfig{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		ServerIP:   w.nextIP.String(),
		ClientIP:   peerIP.String(),
		Port:       w.basePort,
	}

	// Add the peer to WireGuard interface
	if err := w.addPeer(publicKey, peerIP); err != nil {
		return nil, fmt.Errorf("failed to add WireGuard peer: %v", err)
	}

	w.logger.Info().
		Str("peer_id", id).
		Str("peer_ip", peerIP.String()).
		Msg("Added WireGuard peer")

	return config, nil
}

// RemovePeer removes a WireGuard peer
func (w *WireGuardManager) RemovePeer(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cmd := exec.Command("wg", "set", w.interfaceName, "peer", id, "remove")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove WireGuard peer: %v", err)
	}

	w.logger.Info().
		Str("peer_id", id).
		Msg("Removed WireGuard peer")

	return nil
}

// Helper functions

func (w *WireGuardManager) generatePrivateKey() (string, error) {
	cmd := exec.Command("wg", "genkey")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (w *WireGuardManager) generatePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (w *WireGuardManager) allocateIP() net.IP {
	// Simple IP allocation strategy: increment the last octet
	ip := make(net.IP, len(w.nextIP))
	copy(ip, w.nextIP)
	
	// Increment the IP
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}

	// Check if the IP is still in our subnet
	if !w.ipNet.Contains(ip) {
		return nil
	}

	w.nextIP = ip
	return ip
}

func (w *WireGuardManager) addPeer(publicKey string, peerIP net.IP) error {
	cmd := exec.Command("wg", "set", w.interfaceName,
		"peer", publicKey,
		"allowed-ips", peerIP.String()+"/32")
	return cmd.Run()
} 