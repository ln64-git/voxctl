package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/types"
)

var logger *log.Logger

type PlayRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r PlayRequest) ToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func StartServer(port int, azureSubscriptionKey, azureRegion string, state *types.State) {
	err := initLogger()
	if err != nil {
		state.SetStatus(fmt.Sprintf("Failed to initialize logger: %v", err))
		return
	}

	logger.Printf("Starting server on port %d", port)

	go func() {
		http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
			var req PlayRequest
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				logger.Printf("Failed to read request body: %v", err)
				return
			}

			// Remove all extra characters
			bodyString := strings.ReplaceAll(string(bodyBytes), "\n", "")
			bodyString = strings.ReplaceAll(bodyString, "\t", "")

			err = json.Unmarshal([]byte(bodyString), &req)
			if err != nil {
				http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
				logger.Printf("Failed to decode request body: %v", err)
				return
			}

			// Trim leading and trailing whitespace from the text
			req.Text = strings.TrimSpace(req.Text)

			// Log the request
			logger.Printf("Received POST request to /play with text: %s", req.Text)

			err = parseAndPlay(req, azureSubscriptionKey, azureRegion, state)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
				logger.Printf("Failed to play audio: %v", err)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Speech synthesized and added to the queue")
			logger.Printf("Speech synthesized and added to the queue")
		})

		http.HandleFunc("/pause", func(w http.ResponseWriter, r *http.Request) {
			if state.AudioPlayer != nil {
				state.AudioPlayer.Pause()
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Audio playback paused")
				logger.Print("Audio playback paused")
			} else {
				http.Error(w, "AudioPlayer not initialized", http.StatusInternalServerError)
				logger.Print("AudioPlayer not initialized")
			}
		})

		http.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
			if state.AudioPlayer != nil {
				state.AudioPlayer.Stop()
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "Audio playback stopped")
				logger.Print("Audio playback stopped")
			} else {
				http.Error(w, "AudioPlayer not initialized", http.StatusInternalServerError)
				logger.Print("AudioPlayer not initialized")
			}
		})

		addr := ":" + strconv.Itoa(port)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			state.SetStatus(fmt.Sprintf("Failed to start server: %v", err))
			logger.Printf("Failed to start server: %v", err)
		}

		state.SetStatus("Ready")
	}()
}

func parseAndPlay(req PlayRequest, azureSubscriptionKey, azureRegion string, state *types.State) error {
	var sentences []string
	var currentSentence string

	for i, char := range req.Text {
		if char == ',' {
			sentences = append(sentences, currentSentence)
			currentSentence = ""
		} else {
			currentSentence += string(char)
			if i == len(req.Text)-1 {
				sentences = append(sentences, currentSentence)
			}
		}
	}

	for _, sentence := range sentences {
		audioData, err := azure.SynthesizeSpeech(azureSubscriptionKey, azureRegion, sentence, req.Gender, req.VoiceName)
		if err != nil {
			logger.Printf("Failed to synthesize speech for sentence '%s': %v", sentence, err)
			return fmt.Errorf("failed to synthesize speech for sentence '%s': %v", sentence, err)
		}

		if len(audioData) == 0 {
			logger.Printf("Empty audio data received from Azure for sentence '%s'", sentence)
			return fmt.Errorf("empty audio data received from Azure for sentence '%s'", sentence)
		}

		if state.AudioPlayer != nil {
			state.AudioPlayer.AddToQueue(audioData)
			state.AudioPlayer.Play()
		} else {
			logger.Print("AudioPlayer not initialized")
			return fmt.Errorf("AudioPlayer not initialized")
		}
	}

	return nil
}

func initLogger() error {
	// Create the logs directory if it doesn't exist
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		return err
	}

	// Open the log file
	logFile, err := os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Create the logger
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	return nil
}
