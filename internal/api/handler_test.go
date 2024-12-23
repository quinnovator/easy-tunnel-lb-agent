package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/tunnel"
)

func TestNewHandler(t *testing.T) {
	tunnelManager := tunnel.NewManager(10)
	version := "test-version"

	handler := NewHandler(tunnelManager, version)

	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}

	if handler.tunnelManager != tunnelManager {
		t.Error("Expected handler to store tunnel manager reference")
	}

	if handler.version != version {
		t.Errorf("Expected version %s, got %s", version, handler.version)
	}
}

func TestHandleCreateTunnel(t *testing.T) {
	tunnelManager := tunnel.NewManager(10)
	handler := NewHandler(tunnelManager, "test")

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Valid tunnel creation",
			method: http.MethodPost,
			requestBody: CreateTunnelRequest{
				TunnelID:    "test-1",
				Hostname:    "test.example.com",
				TargetPort:  8080,
				Metadata:    map[string]string{"env": "test"},
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp CreateTunnelResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.TunnelID != "test-1" {
					t.Errorf("Expected tunnel ID test-1, got %s", resp.TunnelID)
				}
			},
		},
		{
			name:           "Invalid method",
			method:         http.MethodGet,
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Code != http.StatusMethodNotAllowed {
					t.Errorf("Expected error code %d, got %d", http.StatusMethodNotAllowed, resp.Code)
				}
			},
		},
		{
			name:   "Invalid request body",
			method: http.MethodPost,
			requestBody: map[string]string{
				"invalid": "request",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Code != http.StatusBadRequest {
					t.Errorf("Expected error code %d, got %d", http.StatusBadRequest, resp.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.requestBody != nil {
				if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/new-tunnel", &body)
			w := httptest.NewRecorder()

			handler.handleCreateTunnel(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestHandleRemoveTunnel(t *testing.T) {
	tunnelManager := tunnel.NewManager(10)
	handler := NewHandler(tunnelManager, "test")

	// Create a test tunnel first
	_, err := tunnelManager.CreateTunnel("test-1", "test.example.com", 8080, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Valid tunnel removal",
			method: http.MethodPost,
			requestBody: RemoveTunnelRequest{
				TunnelID: "test-1",
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp RemoveTunnelResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if !resp.Success {
					t.Error("Expected success to be true")
				}
			},
		},
		{
			name:   "Non-existent tunnel",
			method: http.MethodPost,
			requestBody: RemoveTunnelRequest{
				TunnelID: "non-existent",
			},
			expectedStatus: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Code != http.StatusInternalServerError {
					t.Errorf("Expected error code %d, got %d", http.StatusInternalServerError, resp.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.requestBody != nil {
				if err := json.NewEncoder(&body).Encode(tt.requestBody); err != nil {
					t.Fatalf("Failed to encode request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/remove-tunnel", &body)
			w := httptest.NewRecorder()

			handler.handleRemoveTunnel(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

func TestHandleStatus(t *testing.T) {
	tunnelManager := tunnel.NewManager(10)
	version := "test-version"
	handler := NewHandler(tunnelManager, version)

	// Create some test tunnels
	_, err := tunnelManager.CreateTunnel("test-1", "test1.example.com", 8080, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}
	_, err = tunnelManager.CreateTunnel("test-2", "test2.example.com", 8081, "", nil)
	if err != nil {
		t.Fatalf("Failed to create test tunnel: %v", err)
	}

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid status request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp StatusResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Status != "healthy" {
					t.Errorf("Expected status healthy, got %s", resp.Status)
				}
				if resp.Version != version {
					t.Errorf("Expected version %s, got %s", version, resp.Version)
				}
				if resp.NumTunnels != 2 {
					t.Errorf("Expected 2 tunnels, got %d", resp.NumTunnels)
				}
				if resp.Uptime == "" {
					t.Error("Expected non-empty uptime")
				}
			},
		},
		{
			name:           "Invalid method",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Code != http.StatusMethodNotAllowed {
					t.Errorf("Expected error code %d, got %d", http.StatusMethodNotAllowed, resp.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/status", nil)
			w := httptest.NewRecorder()

			handler.handleStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
} 