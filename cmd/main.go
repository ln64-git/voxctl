package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func main() {

	// Parse command-line flags
	flagPort := flag.Int("port", 8080, "Port number to connect or serve")
	flagInput := flag.String("input", "", "Input text to play")
	flagStatus := flag.Bool("status", false, "Request info")
	flagQuit := flag.Bool("quit", false, "Exit application after request")
	flagPause := flag.Bool("pause", false, "Pause audio playback")
	flagStop := flag.Bool("stop", false, "Stop audio playback")
	flag.Parse()

	// Retrieve configuration
	configName := "voxctl.json"
	configData, err := config.GetConfig(configName)
	if err != nil {
		logrus.Fatalf("failed to load configuration: %v", err)
	}

	// Populate state from configuration
	state := types.AppState{
		ClientPort: *flagPort,
		// Token:                 *flagToken,
		ClientInput:           *flagInput,
		ServerStatusRequested: *flagStatus,
		ServerQuitRequested:   *flagQuit,
		ServerPauseRequested:  *flagPause,
		ServerStopRequested:   *flagStop,
		AzureSubscriptionKey:  config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:           config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:      config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:        config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		ServerAlreadyRunning:  server.CheckServerRunning(*flagPort),
	}

	// Check if server is already running
	if !server.CheckServerRunning(state.ClientPort) {
		state.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(state)
		time.Sleep(35 * time.Millisecond)
	} else {
		resp, err := server.ConnectToServer(state.ClientPort)
		if err != nil {
			log.Errorf("Failed to connect to the existing server on port %d: %v", state.ClientPort, err)
		} else {
			log.Infof("Connected to the existing server on port %d. Status: %s", state.ClientPort, resp.Status)
			resp.Body.Close()
		}
	}

	processRequest(state)
	if state.ServerQuitRequested {
		log.Info("Quit flag requested, Program Exiting")
		return
	}

	// Handle OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}

func processRequest(state types.AppState) {
	client := &http.Client{}

	switch {
	case state.ServerStatusRequested:
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.ClientPort))
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.ClientInput != "":
		// log.Info(state.Input)
		speechReq := speech.SpeechRequest{
			Text:      state.ClientInput,
			Gender:    state.AzureVoiceGender,
			VoiceName: state.AzureVoiceName,
		}
		body := bytes.NewBufferString(speechReq.SpeechRequestToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", state.ClientPort), "application/json", body)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.ServerPauseRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/pause", state.ClientPort), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.ServerStopRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/stop", state.ClientPort), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}
}
