// Package loadbalancer provides load balancing functionality for the easy-tunnel-lb-agent.
package loadbalancer

import (
	"fmt"
	"sync"
)

// Router manages the routing table for tunnels
type Router struct {
	mu            sync.RWMutex
	hostMap       map[string]*Target
	portMap       map[int]*Target
	config        *Config
}

// Target represents a tunnel endpoint
type Target struct {
	ID   string
	IP   string
	Port int
}

// NewRouter creates a new router instance
func NewRouter(config *Config) *Router {
	return &Router{
		hostMap: make(map[string]*Target),
		portMap: make(map[int]*Target),
		config:  config,
	}
}

// AddRoute adds a new route to the routing table
func (r *Router) AddRoute(tunnelID string, hostname string, ip string, port int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	target := &Target{
		ID:   tunnelID,
		IP:   ip,
		Port: port,
	}

	// Check if hostname is already in use
	if _, exists := r.hostMap[hostname]; exists {
		return fmt.Errorf("hostname %s is already in use", hostname)
	}

	// Add to host map
	r.hostMap[hostname] = target

	// Optionally add to port map if port-based routing is needed
	if port > 0 {
		if _, exists := r.portMap[port]; exists {
			return fmt.Errorf("port %d is already in use", port)
		}
		r.portMap[port] = target
	}

	return nil
}

// RemoveRoute removes a route from the routing table
func (r *Router) RemoveRoute(tunnelID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from host map
	for hostname, target := range r.hostMap {
		if target.ID == tunnelID {
			delete(r.hostMap, hostname)
		}
	}

	// Remove from port map
	for port, target := range r.portMap {
		if target.ID == tunnelID {
			delete(r.portMap, port)
		}
	}
}

// GetTunnelByHost returns the target for a given hostname
func (r *Router) GetTunnelByHost(hostname string) (*Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	target, exists := r.hostMap[hostname]
	if !exists {
		return nil, fmt.Errorf("no tunnel found for hostname: %s", hostname)
	}

	return target, nil
}

// GetTunnelByPort returns the target for a given port
func (r *Router) GetTunnelByPort(port int) (*Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	target, exists := r.portMap[port]
	if !exists {
		return nil, fmt.Errorf("no tunnel found for port: %d", port)
	}

	return target, nil
}

// ListRoutes returns all active routes
func (r *Router) ListRoutes() map[string]*Target {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make(map[string]*Target)
	for hostname, target := range r.hostMap {
		routes[hostname] = target
	}

	return routes
} 