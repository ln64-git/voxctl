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
)

func main() {
	// Parse command-line flags
	flagsConfig := parseFlags()

	settingsConfig := config.GetConfig()

	// Populate state from configuration
	initializeAppState(&flagsConfig, settingsConfig)

	// Check if server is already running
	if !flagsConfig.ServerAlreadyRunning {
		flagsConfig.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(flagsConfig)
		time.Sleep(35 * time.Millisecond)
	} else {
		resp, err := server.ConnectToServer(flagsConfig.ClientPort)
		if err == nil {
			resp.Body.Close()
		}
	}

	// Process request and exit on quit flag
	processRequest(flagsConfig)
	if flagsConfig.ServerQuitRequested {
		log.Info("Quit flag requested, Program Exiting")
		return
	}
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
		speechReq := speech.SpeechRequest{
			Text: state.ClientInput,
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

func parseFlags() types.AppState {
	return types.AppState{
		ClientPort:            *flag.Int("port", 8080, "Port number to connect or serve"),
		ClientInput:           *flag.String("input", "", "Input text to play"),
		ServerStatusRequested: *flag.Bool("status", false, "Request info"),
		ServerQuitRequested:   *flag.Bool("quit", false, "Exit application after request"),
		ServerPauseRequested:  *flag.Bool("pause", false, "Pause audio playback"),
		ServerStopRequested:   *flag.Bool("stop", false, "Stop audio playback"),
	}
}

func initializeAppState(state *types.AppState, configData map[string]interface{}) {
	state.VoiceService = config.GetStringOrDefault(configData, "VoiceService", "")
	state.ElevenLabsSubscriptionKey = config.GetStringOrDefault(configData, "ElevenLabsSubscriptionKey", "")
	state.ElevenLabsVoiceModelID = config.GetStringOrDefault(configData, "ElevenLabsVoiceModelID", "eleven_monolingual_v1")
	state.ElevenLabsVoiceStability = config.GetFloat64OrDefault(configData, "ElevenLabsVoiceStability", 0.5)
	state.ElevenLabsVoiceSimilarityBoost = config.GetFloat64OrDefault(configData, "ElevenLabsVoiceSimilarityBoost", 0.5)
	state.ElevenLabsVoiceStyle = config.GetFloat64OrDefault(configData, "ElevenLabsVoiceStyle", 0.5)
	state.ElevenLabsVoiceUseSpeakerBoost = config.GetBoolOrDefault(configData, "ElevenLabsVoiceUseSpeakerBoost", false)

	state.AzureSubscriptionKey = config.GetStringOrDefault(configData, "AzureSubscriptionKey", "")
	state.AzureRegion = config.GetStringOrDefault(configData, "AzureRegion", "eastus")
	state.AzureVoiceGender = config.GetStringOrDefault(configData, "AzureVoiceGender", "Female")
	state.AzureVoiceName = config.GetStringOrDefault(configData, "AzureVoiceName", "en-US-JennyNeural")

	state.GoogleSubscriptionKey = config.GetStringOrDefault(configData, "GoogleSubscriptionKey", "")
	state.GoogleLanguageCode = config.GetStringOrDefault(configData, "GoogleLanguageCode", "en-US")
	state.GoogleVoiceName = config.GetStringOrDefault(configData, "GoogleVoiceName", "en-US-Wavenet-D")

	state.ServerAlreadyRunning = server.CheckServerRunning(state.ClientPort)
}
