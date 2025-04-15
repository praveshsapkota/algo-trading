package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/user/algo-trading/internal/backtester"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "configs/.env", "Path to configuration file")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(*configFile); err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		logger.SetLevel(logLevel)
	}

	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		logger.Info("Received termination signal, shutting down...")
		cancel()
	}()

	// Initialize backtester service
	service, err := backtester.NewBacktesterService(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		logger,
	)
	if err != nil {
		logger.Fatalf("Failed to initialize backtester service: %v", err)
	}

	// Start HTTP server
	port := os.Getenv("BACKTESTER_PORT")
	if port == "" {
		port = "8005" // Default port
	}

	// Setup routes
	service.SetupRoutes()

	// Start the server
	go func() {
		logger.Infof("Starting backtester service on port %s", port)
		if err := service.Start(":" + port); err != nil {
			logger.Errorf("Server error: %v", err)
			cancel()
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	logger.Info("Backtester service shutting down")
}