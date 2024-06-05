package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/request"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/state"
	"github.com/ln64-git/voxctl/pkg/flags"
)

func main() {
	// Parse command-line flags
	flagValues := flags.ParseFlags()

	// Retrieve configuration
	configData := config.LoadConfig("voxctl.json")

	// Initialize application state
	appState := state.InitializeAppState(flagValues, configData)

	// Check and start server
	server.HandleServerState(&appState)

	// Process user request
	request.ProcessRequest(&appState)

	// Handle graceful shutdown
	handleShutdown()
}

func handleShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}
