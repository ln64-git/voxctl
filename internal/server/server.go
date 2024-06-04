package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
	"github.com/ln64-git/voxctl/internal/function/scribe"
	"github.com/ln64-git/voxctl/internal/handlers"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func StartServer(state *types.AppState) {
	port := state.Port
	log.Infof("Starting server on port %d", port)

	// Initialize Vosk speech recognizer
	initializeSpeechRecognizer(state)

	http.HandleFunc("/scribe_start", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleScribeStart(w, r, state)
	})

	http.HandleFunc("/scribe_stop", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleScribeStop(w, r, state)
	})

	http.HandleFunc("/scribe_toggle", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleScribeToggle(w, r, state)
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleChatRequest(w, r, state)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		handleStatus(w, r, state)
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioPlayer.Stop)
	})

	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioPlayer.Clear)
	})

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioPlayer.Pause)
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioPlayer.Resume)
	})

	http.HandleFunc("/toggle_playback", func(w http.ResponseWriter, r *http.Request) {
		handleTogglePlayback(w, r, state)
	})

	// Start the HTTP server in a separate goroutine
	go startHTTPServer(port)

	// If Conversation Mode then Start Speech Recognition
	if state.ConversationMode {
		log.Info("Conversation Mode Enabled: Starting Speech Recognition")
		err := state.SpeechRecognizer.Start(state.SpeechTextChan)
		state.ScribeStatus = true
		if err != nil {
			logrus.Errorf("Error starting speech recognizer: %v", err)
		}
	}

	go scribe.ScribeText(state)
}

func handleAudioRequest(w http.ResponseWriter, r *http.Request, controlFunc func()) {
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

func handleStatus(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := types.AppStatusState{
		Port:                 state.Port,
		ServerAlreadyRunning: state.ServerAlreadyRunning,
		ScribeStatus:         state.ScribeStatus,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Errorf("Failed to encode status response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func startHTTPServer(port int) {
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
func initializeSpeechRecognizer(state *types.AppState) {
	recognizer, err := vosk.NewSpeechRecognizer(state.VoskModelPath)
	if err != nil {
		logrus.Errorf("Failed to initialize Vosk speech recognizer: %v", err)
	} else {
		state.SpeechRecognizer = recognizer // Assigning the pointer directly
	}
}
