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
	})

	http.HandleFunc("/stop_speech", func(w http.ResponseWriter, r *http.Request) {
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
	})

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		speechReq, err := processSpeechRequest(r)
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
	})

	http.HandleFunc("/ollama", func(w http.ResponseWriter, r *http.Request) {
		ollamaReq, err := processOllamaRequest(r)
		if err != nil {
			log.Errorf("Failed to process Ollama request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var finalPrompt = ollamaReq.Preface + ollamaReq.Prompt

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
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := types.AppStatusState{
			Port:                 state.Port,
			ServerAlreadyRunning: state.ServerAlreadyRunning,
			ToggleSpeechStatus:   state.ToggleSpeechStatus, // Assuming this field exists and is being tracked
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(status)
		if err != nil {
			log.Errorf("Failed to encode status response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Stop()
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("AudioPlayer not initialized")
		}
	})

	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Clear()
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("AudioPlayer not initialized")
		}
	})

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Pause()
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("AudioPlayer not initialized")
		}
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Resume()
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("AudioPlayer not initialized")
		}
	})

	http.HandleFunc("/toggle_playback", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			if state.AudioPlayer.IsPlaying() {
				state.AudioPlayer.Pause()
			} else {
				state.AudioPlayer.Resume()
			}
			w.WriteHeader(http.StatusOK)
		} else {
			log.Error("AudioPlayer not initialized")
		}
	})

	// Start the HTTP server in a separate goroutine
	go func() {
		addr := ":" + strconv.Itoa(port)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Errorf("%v", err)
		}
	}()

	// Process speech input from the channel
	for result := range state.SpeechInputChan {
		var textResult types.TextResponse
		err := json.Unmarshal([]byte(result), &textResult)
		if err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}
		text := strings.TrimSpace(textResult.Text)
		if text != "" {
			state.SpeechInput += text
		}
	}
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

	// Log the raw request body
	log.Infof("Raw request body: %s", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}

func processSpeechRequest(r *http.Request) (*speech.AzureSpeechRequest, error) {
	var req speech.AzureSpeechRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	// Log the raw request body
	log.Infof("Raw request body: %s", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}
