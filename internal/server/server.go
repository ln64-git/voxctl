package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/audio/audioplayer"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
	"github.com/ln64-git/voxctl/internal/features/scribe"
	"github.com/ln64-git/voxctl/internal/handlers"
	"github.com/ln64-git/voxctl/internal/state"
	"github.com/sirupsen/logrus"
)

// AppStatus holds the status of the application server
type AppStatus struct {
	Port                 int  `json:"port"`
	ServerAlreadyRunning bool `json:"serverAlreadyRunning"`
	ScribeStatus         bool `json:"toggleSpeechStatus"`
}

func StartServer(state *state.AppState) {
	port := state.ServerConfig.Port
	log.Infof("Starting server on port %d", port)

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
		handleAudioRequest(w, r, state.AudioConfig.AudioPlayer.Stop)
	})

	http.HandleFunc("/clear", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioConfig.AudioPlayer.Clear)
	})

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioConfig.AudioPlayer.Pause)
	})

	http.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		handleAudioRequest(w, r, state.AudioConfig.AudioPlayer.Resume)
	})

	http.HandleFunc("/toggle_playback", func(w http.ResponseWriter, r *http.Request) {
		handleTogglePlayback(w, r, state)
	})

	http.HandleFunc("/exit_server", func(w http.ResponseWriter, r *http.Request) {
		handleExitRequest(w, r)
	})

	go startHTTPServer(port)

	state.ServerConfig.ServerRunning = true

	if state.ConversationMode {
		log.Info("Conversation Mode Enabled: Starting Speech Recognition")
		err := state.ScribeConfig.SpeechRecognizer.Start(state.ScribeConfig.ScribeTextChan)
		state.ScribeConfig.ScribeStatus = true
		if err != nil {
			logrus.Errorf("Error starting speech recognizer: %v", err)
		}
	}

	// Start ScribeText in its own goroutine
	go scribe.ScribeText(state)

	// Initialize and start the AudioPlayer in its own goroutine
	state.AudioConfig.AudioPlayer = audioplayer.NewAudioPlayer(state.AudioConfig.AudioEntriesUpdate)
	go state.AudioConfig.AudioPlayer.Start()
}

func handleAudioRequest(w http.ResponseWriter, r *http.Request, controlFunc func()) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	controlFunc()
	w.WriteHeader(http.StatusOK)
}

func handleTogglePlayback(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if state.AudioConfig.AudioPlayer.IsPlaying() {
		state.AudioConfig.AudioPlayer.Pause()
	} else {
		state.AudioConfig.AudioPlayer.Resume()
	}
	w.WriteHeader(http.StatusOK)
}

func handleStatus(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := AppStatus{
		Port:                 state.ServerConfig.Port,
		ServerAlreadyRunning: state.ServerConfig.ServerAlreadyRunning,
		ScribeStatus:         state.ScribeConfig.ScribeStatus,
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

func initializeSpeechRecognizer(state *state.AppState) {
	recognizer, err := vosk.NewSpeechRecognizer(state.ScribeConfig.VoskModelPath)
	if err != nil {
		logrus.Errorf("Failed to initialize Vosk speech recognizer: %v", err)
	} else {
		state.ScribeConfig.SpeechRecognizer = recognizer // Assigning the pointer directly
	}
}

func HandleServerState(appState *state.AppState) {
	if !state.CheckServerRunning(appState.ServerConfig.Port) {
		go StartServer(appState)
		time.Sleep(100 * time.Millisecond) // Initial sleep to give server some time to start
	} else {
		resp, err := ConnectToServer(appState.ServerConfig.Port)
		if err != nil {
			log.Errorf("Failed to connect to the existing server on port %d: %v", appState.ServerConfig.Port, err)
		} else {
			log.Infof("Connected to the existing server on port %d. Status: %s", appState.ServerConfig.Port, resp.Status)
			go func() {
				os.Exit(0)
			}()
			resp.Body.Close()
		}
	}
}

// handleExitRequest handles the /exit endpoint to gracefully shutdown the server.
func handleExitRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	go func() {
		os.Exit(0)
	}()
}
