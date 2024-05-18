package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/audio"
)

type PlayRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r PlayRequest) ToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func StartServer(port int, azureSubscriptionKey, azureRegion string) <-chan string {
	statusCh := make(chan string)

	go func() {
		http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			var req PlayRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}

			audioData, err := azure.SynthesizeSpeech(azureSubscriptionKey, azureRegion, req.Text, req.Gender, req.VoiceName)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			err = audio.PlayAudio(audioData)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Speech synthesized and played successfully")
		})

		addr := ":" + strconv.Itoa(port)

		err := http.ListenAndServe(addr, nil)
		if err != nil {
			statusCh <- fmt.Sprintf("Failed to start server: %v", err)
			return
		}

		statusCh <- fmt.Sprintf("Server started successfully on %s", addr)
	}()

	return statusCh
}
