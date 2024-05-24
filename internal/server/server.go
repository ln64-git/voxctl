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

	"github.com/ln64-git/sandbox/internal/log"
	"github.com/ln64-git/sandbox/internal/speech"
)

type PlayRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r PlayRequest) ToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func StartServer(port int, azureSubscriptionKey, azureRegion string) {
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

		err := speech.ParseAndPlay(req, azureSubscriptionKey, azureRegion)
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
		// Handle pause request
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Audio paused")
	})

	http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		// Handle stop request
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Audio stopped")
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
	if pr, ok := v.(*PlayRequest); ok {
		pr.Text = strings.TrimSpace(pr.Text)
	}

	return nil
}
