package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ln64-git/sandbox/internal/config"
	"github.com/ln64-git/sandbox/internal/log"
	"github.com/ln64-git/sandbox/internal/server"
)

func main() {
	// Initialize the logger
	err := log.InitLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer log.Logger.Writer()
	log.Logger.Println("main - Program Started")

	// Get the configuration
	cfg, err := config.GetConfig()
	if err != nil {
		log.Logger.Printf("Failed to get configuration: %v\n", err)
		return
	}

	// Define the port
	port := 8080

	// Check if the server is already running
	isRunning := server.CheckServerRunning(port)
	if isRunning {
		log.Logger.Printf("Server is already running on port %d. Connecting to the existing server...\n", port)
		server.ConnectToServer(port)
	} else {
		// Start the server
		go server.StartServer(port, cfg.AzureSubscriptionKey, cfg.AzureRegion)
	}

	// Block main from exiting
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Println("main - Program Exiting")
}
