// Package main is the entry point for the easy-tunnel-lb-agent.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/quinnovator/easy-tunnel-lb-agent/internal/api"
	"github.com/quinnovator/easy-tunnel-lb-agent/internal/config"
	"github.com/quinnovator/easy-tunnel-lb-agent/internal/loadbalancer"
	"github.com/quinnovator/easy-tunnel-lb-agent/internal/tunnel"
	"github.com/quinnovator/easy-tunnel-lb-agent/internal/utils"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "", "path to config file (not implemented yet)")
	logLevel := flag.String("log-level", "info", "log level (debug, info, warn, error)")
	flag.Parse()

	// Initialize logger
	utils.InitLogger(*logLevel)
	logger := utils.GetLogger()

	// TODO: Implement config file loading
	if *configFile != "" {
		logger.Warn().Str("config_file", *configFile).Msg("Config file loading not implemented yet, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create tunnel manager
	tunnelManager := tunnel.NewManager(cfg.MaxTunnels)

	// Create router and load balancer
	lbConfig := &loadbalancer.Config{
		HTTPPort: cfg.PublicPort,
		TCPPort:  cfg.PublicPort + 1,
		TLSConfig: &loadbalancer.TLSConfig{
			CertFile: cfg.TLSCertPath,
			KeyFile:  cfg.TLSKeyPath,
		},
	}

	router := loadbalancer.NewRouter(lbConfig)
	lb := loadbalancer.NewLoadBalancer(router, lbConfig)

	// Create API handler
	apiHandler := api.NewHandler(tunnelManager, version)
	apiMux := http.NewServeMux()
	apiHandler.RegisterRoutes(apiMux)

	// Create API server
	apiServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort),
		Handler: apiMux,
	}

	// Start the load balancer
	if err := lb.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start load balancer")
	}

	// Start API server
	go func() {
		logger.Info().
			Str("address", apiServer.Addr).
			Msg("Starting API server")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("API server failed")
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down servers...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown API server
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("API server forced to shutdown")
	}

	// Stop load balancer
	if err := lb.Stop(); err != nil {
		logger.Error().Err(err).Msg("Failed to stop load balancer")
	}

	logger.Info().Msg("Servers stopped")
} 