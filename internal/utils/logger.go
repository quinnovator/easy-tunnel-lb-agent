// Package utils provides utility functions for the easy-tunnel-lb-agent.
package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger with the specified log level
func InitLogger(level string) {
	// Parse the log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	// Create a console writer
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Set global logger
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &log.Logger
} 