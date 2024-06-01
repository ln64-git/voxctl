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
	flags := parseFlags()

	// Retrieve configuration
	configData := loadConfig("voxctl.json")

	// Initialize application state
	state := initializeAppState(flags, configData)

	// Initialize Vosk speech recognizer
	initializeSpeechRecognizer(&state)

	// Check and start server
	handleServerState(&state)

	// Process user request
	processRequest(state)

	// Handle graceful shutdown
	handleShutdown()
}

func parseFlags() *types.Flags {
	flags := &types.Flags{
		Port:           flag.Int("port", 8080, "Port number to connect or serve"),
		UserInput:      flag.String("input", "", "User input for speech or ollama requests"),
		SpeakStart:     flag.Bool("speak_start", false, "Start listening for Speech input"),
		SpeakStop:      flag.Bool("speak_stop", false, "Stop listening for Speech input"),
		SpeakToggle:    flag.Bool("speak_toggle", false, "Toggle listening for Speech input"),
		Status:         flag.Bool("status", false, "Request info"),
		Stop:           flag.Bool("stop", false, "Stop audio playback"),
		Clear:          flag.Bool("clear", false, "Clear playback"),
		Quit:           flag.Bool("quit", false, "Exit application after request"),
		Pause:          flag.Bool("pause", false, "Pause audio playback"),
		Resume:         flag.Bool("resume", false, "Resume audio playback"),
		TogglePlayback: flag.Bool("toggle_playback", false, "Toggle audio playback"),
		OllamaRequest:  flag.Bool("ollama", false, "Request Ollama query"),
		OllamaModel:    flag.String("ollama_model", "", "Ollama model to use"),
		OllamaPreface:  flag.String("ollama_preface", "", "Preface text for the Ollama prompt"),
		OllamaPort:     flag.Int("ollama_port", 0, "Input for Ollama"),
	}
	flag.Parse()
	return flags
}

func loadConfig(configName string) map[string]interface{} {
	configData, err := config.GetConfig(configName)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}
	return configData
}

func initializeAppState(flags *types.Flags, configData map[string]interface{}) types.AppState {
	ollamaModel := config.GetStringOrDefault(configData, "OllamaModel", "")
	if *flags.OllamaModel != "" {
		ollamaModel = *flags.OllamaModel
	}

	return types.AppState{
		Port:                  *flags.Port,
		UserInput:             *flags.UserInput,
		ServerAlreadyRunning:  server.CheckServerRunning(*flags.Port),
		StatusRequest:         *flags.Status,
		StopRequest:           *flags.Stop,
		ClearRequest:          *flags.Clear,
		QuitRequest:           *flags.Quit,
		PauseRequest:          *flags.Pause,
		ResumeRequest:         *flags.Resume,
		TogglePlaybackRequest: *flags.TogglePlayback,
		StartSpeechRequest:    *flags.SpeakStart,
		StopSpeechRequest:     *flags.SpeakStop,
		ToggleSpeechRequest:   *flags.SpeakToggle,
		SpeechInputChan:       make(chan string),
		VoskModelPath:         config.GetStringOrDefault(configData, "VoskModelPath", ""),
		AzureSubscriptionKey:  config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:           config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:      config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:        config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		OllamaRequest:         *flags.OllamaRequest,
		OllamaPort:            *flags.OllamaPort,
		OllamaModel:           ollamaModel,
		OllamaPreface:         *flags.OllamaPreface,
	}
}

func initializeSpeechRecognizer(state *types.AppState) {
	recognizer, err := vosk.NewSpeechRecognizer(state.VoskModelPath)
	if err != nil {
		logrus.Errorf("Failed to initialize Vosk speech recognizer: %v", err)
	} else {
		state.SpeechRecognizer = *recognizer
	}
}

func handleServerState(state *types.AppState) {
	if !server.CheckServerRunning(state.Port) {
		state.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(*state)
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
}

func handleShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}

func processRequest(state types.AppState) {
	client := &http.Client{}

	switch {
	case state.StartSpeechRequest:
		sendPostRequest(client, state.Port, "/start_speech")

	case state.StopSpeechRequest:
		sendPostRequest(client, state.Port, "/stop_speech")

	case state.ToggleSpeechRequest:
		toggleSpeechRecognition(client, state)

	case state.UserInput != "" && state.OllamaRequest:
		processOllamaRequest(client, state)

	case state.UserInput != "":
		processAzureRequest(client, state)

	case state.StatusRequest:
		processStatusRequest(client, state)

	case state.StopRequest:
		sendPostRequest(client, state.Port, "/stop")

	case state.ClearRequest:
		sendPostRequest(client, state.Port, "/clear")

	case state.PauseRequest:
		sendPostRequest(client, state.Port, "/pause")

	case state.ResumeRequest:
		sendPostRequest(client, state.Port, "/resume")

	case state.TogglePlaybackRequest:
		sendPostRequest(client, state.Port, "/toggle_playback")
	}
}

func sendPostRequest(client *http.Client, port int, endpoint string) {
	resp, err := client.Post(fmt.Sprintf("http://localhost:%d%s", port, endpoint), "", nil)
	if err != nil {
		log.Errorf("Failed to send POST request to %s: %v", endpoint, err)
		return
	}
	defer resp.Body.Close()
}

func toggleSpeechRecognition(client *http.Client, state types.AppState) {
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.Port))
	if err != nil {
		log.Errorf("Failed to get status: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	var status types.AppStatusState
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		log.Errorf("Failed to decode JSON response: %v", err)
		return
	}

	if status.ToggleSpeechStatus {
		sendPostRequest(client, state.Port, "/stop_speech")
	} else {
		sendPostRequest(client, state.Port, "/start_speech")
	}
}

func processOllamaRequest(client *http.Client, state types.AppState) {
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
}

func processAzureRequest(client *http.Client, state types.AppState) {
	speechReq := speech.AzureSpeechRequest{
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
}

func processStatusRequest(client *http.Client, state types.AppState) {
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.Port))
	if err != nil {
		log.Errorf("Failed to get status: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	var status types.AppStatusState
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		log.Errorf("Failed to decode JSON response: %v", err)
		return
	}
	if status.ServerAlreadyRunning {
		fmt.Println("Server is already running")
	}
}
