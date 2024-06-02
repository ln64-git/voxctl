package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/clipboard"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func StartServer(state types.AppState) {
	port := state.Port
	log.Infof("Starting server on port %d", port)

	http.HandleFunc("/start_speech", func(w http.ResponseWriter, r *http.Request) {
		handleStartSpeech(w, r, &state)
	})

	http.HandleFunc("/stop_speech", func(w http.ResponseWriter, r *http.Request) {
		handleStopSpeech(w, r, &state)
	})

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		handleAzureInput(w, r, &state)
	})

	http.HandleFunc("/ollama", func(w http.ResponseWriter, r *http.Request) {
		handleOllamaRequest(w, r, &state)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		handleStatus(w, r, &state)
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		handleAudioControl(w, r, state.AudioPlayer.Stop)
	})

	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		handleAudioControl(w, r, state.AudioPlayer.Clear)
	})

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		handleAudioControl(w, r, state.AudioPlayer.Pause)
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		handleAudioControl(w, r, state.AudioPlayer.Resume)
	})

	http.HandleFunc("/toggle_playback", func(w http.ResponseWriter, r *http.Request) {
		handleTogglePlayback(w, r, &state)
	})

	// Start the HTTP server in a separate goroutine
	go startHTTPServer(port)

	// If Conversation Mode then Start Speech Recognition
	if state.ConversationMode {
		log.Info("Conversation Mode Enabled: Starting Speech Recognition")
		err := state.SpeechRecognizer.Start(state.SpeechInputChan)
		if err != nil {
			logrus.Errorf("Error starting speech recognizer: %v", err)
		}
	}

	// Process speech input from the channel in a separate goroutine
	go processSpeechInput(&state)
}

func handleStartSpeech(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	go func() {
		err := state.SpeechRecognizer.Start(state.SpeechInputChan)
		if err != nil {
			logrus.Errorf("Error during speech recognition: %v", err)
		}
	}()
	log.Infof("SpeechInput Starting")
	state.ToggleSpeechStatus = true
	w.WriteHeader(http.StatusOK)
}

func handleStopSpeech(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	go func() {
		state.SpeechRecognizer.Stop()
		clipboard.CopyToClipboard(state.SpeechInput)
		state.SpeechInput = ""
	}()
	log.Infof("SpeechInput Stopped")
	state.ToggleSpeechStatus = false
	w.WriteHeader(http.StatusOK)
}

func handleAzureInput(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	speechReq, err := processAzureRequest(r)
	if err != nil {
		log.Errorf("Failed to process speech request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = speech.ProcessSpeech(*speechReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
	if err != nil {
		log.Errorf("Failed to process speech: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleOllamaRequest(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	ollamaReq, err := processOllamaRequest(r)
	if err != nil {
		log.Errorf("Failed to process Ollama request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	finalPrompt := ollamaReq.Preface + ollamaReq.Prompt

	tokenChan, err := ollama.GetOllamaTokenResponse(ollamaReq.Model, finalPrompt)
	if err != nil {
		log.Errorf("Failed to get Ollama token response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sentenceChan := make(chan string)
	go speech.SegmentTextFromChannel(tokenChan, sentenceChan)

	go func() {
		for sentence := range sentenceChan {
			audioData, err := azure.SynthesizeSpeech(state.AzureSubscriptionKey, state.AzureRegion, sentence, state.AzureVoiceGender, state.AzureVoiceName)
			if err != nil {
				log.Errorf("Failed to synthesize speech: %v", err)
				return
			}
			state.AudioPlayer.Play(audioData)
		}
	}()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func handleStatus(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := types.AppStatusState{
		Port:                 state.Port,
		ServerAlreadyRunning: state.ServerAlreadyRunning,
		ToggleSpeechStatus:   state.ToggleSpeechStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Errorf("Failed to encode status response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func handleAudioControl(w http.ResponseWriter, r *http.Request, controlFunc func()) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	controlFunc()
	w.WriteHeader(http.StatusOK)
}

func handleTogglePlayback(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if state.AudioPlayer.IsPlaying() {
		state.AudioPlayer.Pause()
	} else {
		state.AudioPlayer.Resume()
	}
	w.WriteHeader(http.StatusOK)
}

func startHTTPServer(port int) {
	addr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Errorf("%v", err)
	}
}

func processSpeechInput(state *types.AppState) {
	for result := range state.SpeechInputChan {
		log.Info(result)

		var textResult types.TextResponse
		err := json.Unmarshal([]byte(result), &textResult)
		if err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}
		text := strings.TrimSpace(textResult.Text)
		if text != "" {
			log.Infof("SpeechInputChan: %s", text)

			state.SpeechInput += text + " "
			// Handle conversation mode
			if state.ConversationMode && len(strings.Fields(state.SpeechInput)) >= 3 {
				handleConversation(state)
			}
		}
	}
}

func handleConversation(state *types.AppState) {
	log.Info("handleConversation Called")
	go func() {
		ollamaReq := ollama.OllamaRequest{
			Model:   state.OllamaModel,
			Prompt:  state.SpeechInput,
			Preface: state.OllamaPreface,
		}
		body, err := json.Marshal(ollamaReq)
		if err != nil {
			logrus.Errorf("Failed to marshal Ollama request: %v", err)
			return
		}

		req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(state.Port)+"/ollama", strings.NewReader(string(body)))
		if err != nil {
			logrus.Errorf("Error creating request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logrus.Errorf("Error making request: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logrus.Errorf("Request failed with status: %v", resp.Status)
			return
		}
		state.SpeechInput = ""
	}()
}

// CheckServerRunning checks if the server is already running on the specified port.
func CheckServerRunning(port int) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// ConnectToServer connects to the server on the specified port and returns the response or an error.
func ConnectToServer(port int) (*http.Response, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/status", port))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the server: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func processOllamaRequest(r *http.Request) (*ollama.OllamaRequest, error) {
	var req ollama.OllamaRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	log.Infof("Raw request body: %s", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}

func processAzureRequest(r *http.Request) (*speech.AzureSpeechRequest, error) {
	var req speech.AzureSpeechRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	log.Infof("Raw request body: %s", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}
