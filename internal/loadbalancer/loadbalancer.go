// Package loadbalancer provides load balancing functionality for the easy-tunnel-lb-agent.
package loadbalancer

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/utils"
	"github.com/rs/zerolog"
)

// LoadBalancer handles the routing of incoming requests to appropriate tunnels
type LoadBalancer struct {
	router     *Router
	logger     *zerolog.Logger
	httpServer *http.Server
	tcpServer  net.Listener
	mu         sync.RWMutex
}

// Config holds the configuration for the load balancer
type Config struct {
	HTTPPort  int
	TCPPort   int
	TLSConfig *TLSConfig
}

// TLSConfig holds TLS certificate configuration
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// NewLoadBalancer creates a new load balancer instance
func NewLoadBalancer(router *Router, config *Config) *LoadBalancer {
	logger := utils.GetLogger()
	return &LoadBalancer{
		router: router,
		logger: logger,
	}
}

// Start starts the load balancer
func (lb *LoadBalancer) Start() error {
	// Start HTTP server
	if err := lb.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}

	// Start TCP server
	if err := lb.startTCPServer(); err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}

	return nil
}

// Stop gracefully stops the load balancer
func (lb *LoadBalancer) Stop() error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Stop HTTP server
	if lb.httpServer != nil {
		if err := lb.httpServer.Close(); err != nil {
			lb.logger.Error().Err(err).Msg("Failed to stop HTTP server")
		}
	}

	// Stop TCP server
	if lb.tcpServer != nil {
		if err := lb.tcpServer.Close(); err != nil {
			lb.logger.Error().Err(err).Msg("Failed to stop TCP server")
		}
	}

	return nil
}

func (lb *LoadBalancer) startHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.handleHTTPRequest)

	lb.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", lb.router.config.HTTPPort),
		Handler: mux,
	}

	go func() {
		if err := lb.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lb.logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	return nil
}

func (lb *LoadBalancer) startTCPServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", lb.router.config.TCPPort))
	if err != nil {
		return err
	}

	lb.tcpServer = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" {
					return // Server is shutting down
				}
				lb.logger.Error().Err(err).Msg("Failed to accept TCP connection")
				continue
			}
			go lb.handleTCPConnection(conn)
		}
	}()

	return nil
}

func (lb *LoadBalancer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	host := r.Host

	// Find the target tunnel based on the hostname
	target, err := lb.router.GetTunnelByHost(host)
	if err != nil {
		lb.logger.Error().
			Err(err).
			Str("host", host).
			Msg("No tunnel found for host")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	// Create the reverse proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf("%s:%d", target.IP, target.Port)
			req.Host = host
		},
	}

	// Forward the request
	proxy.ServeHTTP(w, r)

	lb.logger.Info().
		Str("host", host).
		Str("tunnel_id", target.ID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Dur("duration", time.Since(start)).
		Msg("Handled HTTP request")
}

func (lb *LoadBalancer) handleTCPConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Get the original destination (this is where you'd implement port-based routing)
	target, err := lb.router.GetTunnelByPort(clientConn.LocalAddr().(*net.TCPAddr).Port)
	if err != nil {
		lb.logger.Error().
			Err(err).
			Int("port", clientConn.LocalAddr().(*net.TCPAddr).Port).
			Msg("No tunnel found for port")
		return
	}

	// Connect to the backend
	backendConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", target.IP, target.Port))
	if err != nil {
		lb.logger.Error().
			Err(err).
			Str("tunnel_id", target.ID).
			Msg("Failed to connect to backend")
		return
	}
	defer backendConn.Close()

	// Start proxying in both directions
	go lb.proxy(clientConn, backendConn)
	lb.proxy(backendConn, clientConn)
}

func (lb *LoadBalancer) proxy(dst net.Conn, src net.Conn) {
	buffer := make([]byte, 32*1024)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}
		_, err = dst.Write(buffer[:n])
		if err != nil {
			return
		}
	}
} 