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

	http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		var req speech.PlayRequest
		if err := parseJSONRequest(r, &req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Logger.Printf("Failed to decode request body: %v", err)
			return
		}

		// Log the request
		log.Logger.Printf("Received POST request to /play with text: %s", req.Text)

		playReq := speech.PlayRequest{
			Text:      req.Text,
			Gender:    req.Gender,
			VoiceName: req.VoiceName,
		}

		// Pass the AudioPlayer as a pointer
		err := speech.ParseAndPlay(playReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
			log.Logger.Printf("Failed to play audio: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Speech synthesized and added to the queue")
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

func parseJSONRequest(r *http.Request, v interface{}) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	// Remove all extra characters
	bodyString := strings.ReplaceAll(string(bodyBytes), "\n", "")
	bodyString = strings.ReplaceAll(bodyString, "\t", "")

	err = json.Unmarshal([]byte(bodyString), v)
	if err != nil {
		return fmt.Errorf("failed to decode request body: %v", err)
	}

	// Trim leading and trailing whitespace from text fields
	if pr, ok := v.(*speech.PlayRequest); ok {
		pr.Text = strings.TrimSpace(pr.Text)
	}

	return nil
}
