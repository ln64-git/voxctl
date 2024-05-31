package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/ln64-git/voxctl/internal/vosk"
	"github.com/sirupsen/logrus"
)

func main() {

	// Parse command-line flags
	flagPort := flag.Int("port", 8080, "Port number to connect or serve")
	flagUserInput := flag.String("input", "", "User input for speech or ollama requests")
	flagSpeak := flag.Bool("speak", false, "Listen for Speech input")
	flagStatus := flag.Bool("status", false, "Request info")
	flagStop := flag.Bool("stop", false, "Stop audio playback")
	flagClear := flag.Bool("clear", false, "Clear playback")
	flagQuit := flag.Bool("quit", false, "Exit application after request")
	flagPause := flag.Bool("pause", false, "Pause audio playback")
	flagResume := flag.Bool("resume", false, "Ollama model to use")
	flagTogglePlayback := flag.Bool("toggle_playback", false, "Ollama model to use")
	flagOllamaRequest := flag.Bool("ollama", false, "Request ollama querry")
	flagOllamaModel := flag.String("ollama_model", "", "Ollama model to use")
	flagOllamaPreface := flag.String("ollama_preface", "", "Preface text for the Ollama prompt")
	flagOllamaPort := flag.Int("ollama_port", 0, "input for ollama")
	flag.Parse()

	// Retrieve configuration
	configName := "voxctl.json"
	configData, err := config.GetConfig(configName)
	if err != nil {
		logrus.Fatalf("failed to load configuration: %v", err)
	}

	// Populate state from configuration
	var ollamaModel string
	if *flagOllamaModel == "" {
		ollamaModel = config.GetStringOrDefault(configData, "OllamaModel", "")
	} else {
		ollamaModel = *flagOllamaModel
	}

	// Populate state from configuration
	state := types.AppState{
		Port:                    *flagPort,
		UserInput:               *flagUserInput,
		ServerAlreadyRunning:    server.CheckServerRunning(*flagPort),
		StatusRequested:         *flagStatus,
		StopRequested:           *flagStop,
		ClearRequested:          *flagClear,
		QuitRequested:           *flagQuit,
		PauseRequested:          *flagPause,
		ResumeRequested:         *flagResume,
		TogglePlaybackRequested: *flagTogglePlayback,
		SpeakRequest:            *flagSpeak,
		VoskModelPath:           config.GetStringOrDefault(configData, "VoskModelPath", ""),
		AzureSubscriptionKey:    config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:             config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:        config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:          config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		OllamaRequest:           *flagOllamaRequest,
		OllamaPort:              *flagOllamaPort,
		OllamaModel:             ollamaModel,
		OllamaPreface:           *flagOllamaPreface,
	}

	// Check if server is already running
	if !server.CheckServerRunning(state.Port) {
		state.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(state)
		time.Sleep(35 * time.Millisecond)
	} else {
		resp, err := server.ConnectToServer(state.Port)
		if err != nil {
			log.Errorf("Failed to connect to the existing server on port %d: %v", state.Port, err)
		} else {
			log.Infof("Connected to the existing server on port %d. Status: %s", state.Port, resp.Status)
			resp.Body.Close()
		}
	}

	processRequest(state)
	if state.QuitRequested {
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
	case state.SpeakRequest:
		recognizer, err := vosk.NewSpeechRecognizer(state.VoskModelPath)
		if err != nil {
			logrus.Errorf("Failed to initialize Vosk speech recognizer: %v", err)
			return
		}
		resultChan := make(chan string)

		go func() {
			err := recognizer.Start(resultChan)
			if err != nil {
				logrus.Errorf("Error during speech recognition: %v", err)
			}
		}()
		defer recognizer.Stop()

	case state.UserInput != "" && state.OllamaRequest:
		ollamaReq := ollama.OllamaRequest{
			Model:   state.OllamaModel,
			Prompt:  state.UserInput,
			Preface: state.OllamaPreface,
		}
		body, err := json.Marshal(ollamaReq)
		if err != nil {
			logrus.Errorf("Failed to marshal Ollama request: %v", err)
			return
		}

		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/ollama", state.Port), "text/plain", bytes.NewBuffer(body))
		if err != nil {
			logrus.Errorf("Failed to send Ollama request: %v", err)
			return
		}
		defer resp.Body.Close()

	case state.UserInput != "":
		speechReq := speech.SpeechRequest{
			Text:      state.UserInput,
			Gender:    state.AzureVoiceGender,
			VoiceName: state.AzureVoiceName,
		}
		body := bytes.NewBufferString(speechReq.SpeechRequestToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", state.Port), "application/json", body)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.StatusRequested:
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.Port))
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.StopRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/stop", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.ClearRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/clear", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.PauseRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/pause", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.ResumeRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/resume", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.TogglePlaybackRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/toggle_playback", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	}
}
