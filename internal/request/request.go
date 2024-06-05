package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/function/speak"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func ProcessRequest(state *types.AppState) {
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

	log.Info("processChatRequest - INIT")
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
