package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/request"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/types"
)

func main() {
	// Parse command-line flags
	flags := config.ParseFlags()

	// Retrieve configuration
	configData := config.LoadConfig("voxctl.json")

	// Initialize application state
	state := initializeAppState(flags, configData)

	// Check and start server
	server.HandleServerState(&state)

	// Process user request
	request.ProcessRequest(&state)

	// Handle graceful shutdown
	handleShutdown()
}

func initializeAppState(flags *types.Flags, configData map[string]interface{}) types.AppState {
	return types.AppState{
		Port:                  *flags.Port,
		ServerAlreadyRunning:  server.CheckServerRunning(*flags.Port),
		ConversationMode:      *flags.Convo,
		SpeakText:             *flags.SpeakText,
		ChatText:              *flags.ChatText,
		StatusRequest:         *flags.Status,
		StopRequest:           *flags.Stop,
		ClearRequest:          *flags.Clear,
		QuitRequest:           *flags.Quit,
		PauseRequest:          *flags.Pause,
		ResumeRequest:         *flags.Resume,
		TogglePlaybackRequest: *flags.TogglePlayback,
		ScribeStartRequest:    *flags.ScribeStart,
		ScribeStopRequest:     *flags.ScribeStop,
		ScribeToggleRequest:   *flags.ScribeToggle,
		ScribeTextChan:        make(chan string),
		VoskModelPath:         config.GetStringOrDefault(configData, "VoskModelPath", ""),
		AzureSubscriptionKey:  config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:           config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:      config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:        config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		OllamaModel:           config.GetStringOrDefault(configData, "OllamaModel", "llama3"),
		OllamaPreface:         config.GetStringOrDefault(configData, "OllamaPreface", ""),
	}
}

func handleShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}
