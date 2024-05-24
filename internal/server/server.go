package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ln64-git/voxctl/internal/input"
	"github.com/ln64-git/voxctl/internal/log"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
)

func StartServer(state types.AppState) {
	port := state.Port
	log.Logger.Printf("Starting server on port %d", port)

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Server is running")
	})

	http.HandleFunc("/input", func(w http.ResponseWriter, r *http.Request) {
		playReq, err := processInputRequest(r)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Logger.Printf("%v", err)
			return
		}

		// Pass the AudioPlayer as a pointer
		err = speech.ParseAndPlay(*playReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
			log.Logger.Printf("Failed to play audio: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Logger.Printf("Speech synthesized and added to the queue")
	})

	http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Pause()
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Audio paused")
		} else {
			http.Error(w, "AudioPlayer not initialized", http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if state.AudioPlayer != nil {
			state.AudioPlayer.Stop()
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Audio stopped")
		} else {
			http.Error(w, "AudioPlayer not initialized", http.StatusInternalServerError)
		}
	})

	addr := ":" + strconv.Itoa(port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Logger.Printf("Failed to start server: %v", err)
	}
}

func CheckServerRunning(port int) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func ConnectToServer(port int) {
	log.Logger.Printf("Connected to the server on port %d.\n", port)
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/status", port))
	if err != nil {
		log.Logger.Printf("Failed to connect to the server: %v\n", err)
		return
	}
	defer resp.Body.Close()
	log.Logger.Printf("Server response: %s\n", resp.Status)
}

func processInputRequest(r *http.Request) (*speech.PlayRequest, error) {
	var req speech.PlayRequest

	// Read the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	// Unmarshal the request body into the PlayRequest struct
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	// Parse the text from the request body
	text, err := input.SanitizeInput(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse text from request: %v", err)
	}

	// Log the request
	log.Logger.Printf("Received POST request to /play with text: %s", text)

	playReq := speech.PlayRequest{
		Text:      text,
		Gender:    req.Gender,
		VoiceName: req.VoiceName,
	}

	return &playReq, nil
}
