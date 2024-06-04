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
	"github.com/ln64-git/voxctl/internal/audio/player"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/function/speak"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	flags := parseFlags()

	// Retrieve configuration
	configData := loadConfig("voxctl.json")

	// Initialize application state
	state := initializeAppState(flags, configData)

	// Check and start server
	handleServerState(&state)

	// Process user request
	processRequest(&state)

	// Handle graceful shutdown
	handleShutdown()
}

func parseFlags() *types.Flags {
	flags := &types.Flags{
		Port:           flag.Int("port", 8080, "Port number to connect or serve"),
		Convo:          flag.Bool("convo", false, "Start Conversation Mode"),
		SpeakText:      flag.String("speak", "", "User input for speech or ollama requests"),
		ChatText:       flag.String("chat", "", "Chat with AI through text"),
		ScribeStart:    flag.Bool("scribe_start", false, "Start listening for Speech input"),
		ScribeStop:     flag.Bool("scribe_stop", false, "Stop listening for Speech input"),
		ScribeToggle:   flag.Bool("scribe_toggle", false, "Toggle listening for Speech input"),
		Status:         flag.Bool("status", false, "Request info"),
		Stop:           flag.Bool("stop", false, "Stop audio playback"),
		Clear:          flag.Bool("clear", false, "Clear playback"),
		Quit:           flag.Bool("quit", false, "Exit application after request"),
		Pause:          flag.Bool("pause", false, "Pause audio playback"),
		Resume:         flag.Bool("resume", false, "Resume audio playback"),
		TogglePlayback: flag.Bool("playback_toggle", false, "Toggle audio playback"),
	}
	flag.Parse()
	// If no flag arguments are provided, set Convo to true
	if flag.NArg() == 0 {
		flags.Convo = new(bool)
		*flags.Convo = true
	}
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
		SpeechTextChan:        make(chan string),
		VoskModelPath:         config.GetStringOrDefault(configData, "VoskModelPath", ""),
		AzureSubscriptionKey:  config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:           config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:      config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:        config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		OllamaModel:           config.GetStringOrDefault(configData, "OllamaModel", "llama3"),
		OllamaPreface:         config.GetStringOrDefault(configData, "OllamaPreface", ""),
	}
}

func handleServerState(state *types.AppState) {
	if !server.CheckServerRunning(state.Port) {
		state.AudioPlayer = player.NewAudioPlayer()
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
}

func handleShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}

func processRequest(state *types.AppState) {
	client := &http.Client{}

	switch {
	case state.ScribeStartRequest:
		sendPostRequest(client, state.Port, "/scribe_start")

	case state.ScribeStopRequest:
		sendPostRequest(client, state.Port, "/scribe_stop")

	case state.ScribeToggleRequest:
		sendPostRequest(client, state.Port, "/scribe_toggle")

	case state.ChatText != "":
		processChatRequest(client, state)

	case state.SpeakText != "":
		processSpeakRequest(client, state)

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

func processChatRequest(client *http.Client, state *types.AppState) {
	ollamaReq := ollama.OllamaRequest{
		Model:   state.OllamaModel,
		Prompt:  state.ChatText,
		Preface: state.OllamaPreface,
	}
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		logrus.Errorf("Failed to marshal Ollama request: %v", err)
		return
	}

	log.Info("processChatRequest - INTI")
	log.Info(ollamaReq)

	resp, err := client.Post(fmt.Sprintf("http://localhost:%d/chat", state.Port), "text/plain", bytes.NewBuffer(body))
	if err != nil {
		logrus.Errorf("Failed to send Ollama request: %v", err)
		return
	}
	defer resp.Body.Close()
}

func processSpeakRequest(client *http.Client, state *types.AppState) {
	speechReq := speak.AzureSpeechRequest{
		Text:      state.SpeakText,
		Gender:    state.AzureVoiceGender,
		VoiceName: state.AzureVoiceName,
	}
	body := bytes.NewBufferString(speechReq.AzureRequestToJSON())
	resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", state.Port), "application/json", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func processStatusRequest(client *http.Client, state *types.AppState) {
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
