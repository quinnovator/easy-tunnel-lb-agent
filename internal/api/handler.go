// Package api provides the HTTP API handlers and models for the easy-tunnel-lb-agent.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/tunnel"
	"github.com/quinnovator/easy-tunnel-lb-agent/internal/utils"
	"github.com/rs/zerolog"
)

// Handler handles HTTP requests for the tunnel API
type Handler struct {
	tunnelManager *tunnel.Manager
	logger        *zerolog.Logger
	startTime     time.Time
	version       string
}

// NewHandler creates a new API handler
func NewHandler(tunnelManager *tunnel.Manager, version string) *Handler {
	return &Handler{
		tunnelManager: tunnelManager,
		logger:        utils.GetLogger(),
		startTime:     time.Now(),
		version:      version,
	}
}

// RegisterRoutes registers the API routes with the given router
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/new-tunnel", h.handleCreateTunnel)
	mux.HandleFunc("/api/remove-tunnel", h.handleRemoveTunnel)
	mux.HandleFunc("/api/status", h.handleStatus)
}

func (h *Handler) handleCreateTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TunnelID == "" || req.Hostname == "" || req.TargetPort <= 0 {
		h.sendError(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create the tunnel
	tunnelInfo, err := h.tunnelManager.CreateTunnel(
		req.TunnelID,
		req.Hostname,
		req.TargetPort,
		req.WireGuardPublicKey,
		req.Metadata,
	)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := CreateTunnelResponse{
		TunnelID:       tunnelInfo.ID,
		PublicEndpoint: tunnelInfo.PublicEndpoint,
	}

	// Add WireGuard config if available
	if tunnelInfo.WireGuardConfig != nil {
		resp.WireGuardConfig = &WireGuardConfig{
			PublicKey:  tunnelInfo.WireGuardConfig.PublicKey,
			PrivateKey: tunnelInfo.WireGuardConfig.PrivateKey,
			ServerIP:   tunnelInfo.WireGuardConfig.ServerIP,
			ClientIP:   tunnelInfo.WireGuardConfig.ClientIP,
			Port:       tunnelInfo.WireGuardConfig.Port,
		}
	}

	h.sendJSON(w, resp, http.StatusCreated)
}

func (h *Handler) handleRemoveTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RemoveTunnelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TunnelID == "" {
		h.sendError(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	if err := h.tunnelManager.RemoveTunnel(req.TunnelID); err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, RemoveTunnelResponse{
		Success: true,
		Message: "Tunnel removed successfully",
	}, http.StatusOK)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnels := h.tunnelManager.GetAllTunnels()
	
	h.sendJSON(w, StatusResponse{
		Status:     "healthy",
		Version:    h.version,
		Uptime:     time.Since(h.startTime).String(),
		NumTunnels: len(tunnels),
	}, http.StatusOK)
}

// Helper functions for sending responses

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode JSON response")
	}
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	h.sendJSON(w, ErrorResponse{
		Error:   http.StatusText(status),
		Code:    status,
		Details: message,
	}, status)
} 