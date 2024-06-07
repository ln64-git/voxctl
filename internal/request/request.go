package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/flags"
	"github.com/ln64-git/voxctl/internal/features/speak"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/state"
	"github.com/sirupsen/logrus"
)

func ProcessRequest(appState *state.AppState, flagState *flags.Flags) {
	client := &http.Client{}

	switch {
	case *flagState.ScribeStart:
		sendPostRequest(client, appState.ServerConfig.Port, "/scribe_start")

	case *flagState.ScribeStop:
		sendPostRequest(client, appState.ServerConfig.Port, "/scribe_stop")

	case *flagState.ScribeToggle:
		sendPostRequest(client, appState.ServerConfig.Port, "/scribe_toggle")

	case appState.ChatText != "":
		processChatRequest(client, appState)

	case appState.SpeakText != "":
		processSpeakRequest(client, appState)

	case *flagState.Status:
		processStatusRequest(client, appState)

	case *flagState.Stop:
		sendPostRequest(client, appState.ServerConfig.Port, "/stop")

	case *flagState.Clear:
		sendPostRequest(client, appState.ServerConfig.Port, "/clear")

	case *flagState.Pause:
		sendPostRequest(client, appState.ServerConfig.Port, "/pause")

	case *flagState.Resume:
		sendPostRequest(client, appState.ServerConfig.Port, "/resume")

	case *flagState.TogglePlayback:
		sendPostRequest(client, appState.ServerConfig.Port, "/toggle_playback")

	case *flagState.ExitServer:
		sendPostRequest(client, appState.ServerConfig.Port, "/exit_server")
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

func processChatRequest(client *http.Client, appState *state.AppState) {
	ollamaReq := ollama.OllamaRequest{
		Model:   appState.OllamaConfig.Model,
		Prompt:  appState.ChatText,
		Preface: appState.OllamaConfig.Preface,
	}
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		logrus.Errorf("Failed to marshal Ollama request: %v", err)
		return
	}

	log.Info("processChatRequest - INIT")
	log.Info(ollamaReq)

	resp, err := client.Post(fmt.Sprintf("http://localhost:%d/chat", appState.ServerConfig.Port), "text/plain", bytes.NewBuffer(body))
	if err != nil {
		logrus.Errorf("Failed to send Ollama request: %v", err)
		return
	}
	defer resp.Body.Close()
}

func processSpeakRequest(client *http.Client, appState *state.AppState) {
	speechReq := speak.AzureSpeechRequest{
		Text:      appState.SpeakText,
		Gender:    appState.AzureConfig.VoiceGender,
		VoiceName: appState.AzureConfig.VoiceName,
	}
	body := bytes.NewBufferString(speechReq.AzureRequestToJSON())
	resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", appState.ServerConfig.Port), "application/json", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func processStatusRequest(client *http.Client, appState *state.AppState) {
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", appState.ServerConfig.Port))
	if err != nil {
		log.Errorf("Failed to get status: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	var status server.AppStatus
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		log.Errorf("Failed to decode JSON response: %v", err)
		return
	}
	if status.ServerAlreadyRunning {
		fmt.Println("Server is already running")
	}
}
