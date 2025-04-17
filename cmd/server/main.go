package main

import (
	"flag"
	"os"

	"github.com/hewenyu/gin-pkg/internal/app"
	"github.com/hewenyu/gin-pkg/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config/default.yaml", "path to configuration file")
	debugMode := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	// Configure logger
	if *debugMode {
		logger.SetDefaultLogger(logger.NewDefaultLogger(os.Stdout, logger.DebugLevel))
	}

	// Create new application
	application, err := app.NewApp(*configPath)
	if err != nil {
		logger.Fatalf("Failed to create application: %v", err)
	}

	// Initialize application
	if err := application.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}

	// Ensure resources are cleaned up
	defer application.Cleanup()

	// Run application
	if err := application.Run(); err != nil {
		logger.Fatalf("Application error: %v", err)
	}
}
