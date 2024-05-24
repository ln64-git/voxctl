package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ln64-git/voxctl/external/azure"
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
		inputReq, err := processSpeechRequest(r)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Logger.Printf("%v", err)
			return
		}

		err = speech.ProcessSpeech(*inputReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
			log.Logger.Printf("Failed to play audio: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Logger.Printf("Speech synthesized and added to the queue")
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenReq, err := processSpeechRequest(r)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			log.Logger.Printf("%v", err)
			return
		}

		log.Logger.Printf("Received token request: %v", tokenReq.Text)

		var sentences []string
		var currentSentence string
		for i, char := range tokenReq.Text {
			if char == ',' || char == '.' || char == '!' || char == '?' {
				sentences = append(sentences, currentSentence)
				currentSentence = ""
				log.Logger.Printf("Parsed sentence: %s", sentences[len(sentences)-1])
			} else {
				currentSentence += string(char)
				if i == len(tokenReq.Text)-1 {
					sentences = append(sentences, currentSentence)
					log.Logger.Printf("Parsed sentence: %s", currentSentence)
				}
			}
		}

		for _, sentence := range sentences {
			log.Logger.Printf("Synthesizing sentence: %s", sentence)
			audioData, err := azure.SynthesizeSpeech(state.AzureSubscriptionKey, state.AzureRegion, sentence, tokenReq.Gender, tokenReq.VoiceName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to synthesize speech: %v", err), http.StatusInternalServerError)
				log.Logger.Printf("Failed to synthesize speech: %v", err)
				return
			}
			log.Logger.Printf("Playing synthesized audio for sentence: %s", sentence)
			state.AudioPlayer.Play(audioData)
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
