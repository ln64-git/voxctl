package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
)

func StartServer(state types.AppState) {
	port := state.Port
	log.Infof("Starting server on port %d", port)

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		speechReq, err := processSpeechRequest(r)
		if err != nil {
			log.Errorf("%v", err)
			return
		}

		err = speech.ProcessSpeech(*speechReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
		if err != nil {
			log.Errorf("%v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/ollama", func(w http.ResponseWriter, r *http.Request) {
		ollamaReq, err := processOllamaRequest(r)
		if err != nil {
			log.Errorf("%v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var finalPrompt = ollamaReq.Preface + ollamaReq.Prompt

		tokenChan, err := ollama.GetOllamaTokenResponse(ollamaReq.Model, finalPrompt)
		if err != nil {
			log.Errorf("%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sentenceChan := make(chan string)
		go speech.SegmentTextFromChannel(tokenChan, sentenceChan)

		go func() {
			for sentence := range sentenceChan {
				audioData, err := azure.SynthesizeSpeech(state.AzureSubscriptionKey, state.AzureRegion, sentence, state.AzureVoiceGender, state.AzureVoiceName)
				if err != nil {
					log.Errorf("%v", err)
					return
				}
				state.AudioPlayer.Play(audioData)
			}
		}()

		w.Header().Set("Content-Type", "text/plain")
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
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

	addr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Errorf("%v", err)
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

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}

func processSpeechRequest(r *http.Request) (*speech.SpeechRequest, error) {
	var req speech.SpeechRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}
