// Package api provides the HTTP API handlers and models for the easy-tunnel-lb-agent.
package api

// CreateTunnelRequest represents the request payload for creating a new tunnel
type CreateTunnelRequest struct {
	// Unique identifier for the tunnel
	TunnelID string `json:"tunnel_id"`
	
	// The hostname to route traffic to (e.g., service.example.com)
	Hostname string `json:"hostname"`
	
	// The target port on the tunnel endpoint
	TargetPort int `json:"target_port"`
	
	// Optional: WireGuard public key if using WireGuard tunnels
	WireGuardPublicKey string `json:"wireguard_public_key,omitempty"`
	
	// Optional: Additional metadata for the tunnel
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CreateTunnelResponse represents the response for a successful tunnel creation
type CreateTunnelResponse struct {
	// The tunnel ID that was created
	TunnelID string `json:"tunnel_id"`
	
	// The assigned public hostname or IP for the tunnel
	PublicEndpoint string `json:"public_endpoint"`
	
	// WireGuard configuration if applicable
	WireGuardConfig *WireGuardConfig `json:"wireguard_config,omitempty"`
}

// WireGuardConfig contains WireGuard-specific configuration
type WireGuardConfig struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
	ServerIP   string `json:"server_ip"`
	ClientIP   string `json:"client_ip"`
	Port       int    `json:"port"`
}

// RemoveTunnelRequest represents the request payload for removing a tunnel
type RemoveTunnelRequest struct {
	TunnelID string `json:"tunnel_id"`
}

// RemoveTunnelResponse represents the response for a successful tunnel removal
type RemoveTunnelResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
}

// StatusResponse represents the response for the status endpoint
type StatusResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Uptime    string `json:"uptime"`
	NumTunnels int   `json:"num_tunnels"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
} 