package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/types"
)

type PlayRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r PlayRequest) ToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func StartServer(port int, azureSubscriptionKey, azureRegion string, state *types.State) {
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

			err = parseAndPlay(req, azureSubscriptionKey, azureRegion, state)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Speech synthesized and added to the queue")
		})
		addr := ":" + strconv.Itoa(port)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			state.SetStatus(fmt.Sprintf("Failed to start server: %v", err))
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
			return fmt.Errorf("failed to synthesize speech for sentence '%s': %v", sentence, err)
		}

		if len(audioData) == 0 {
			return fmt.Errorf("empty audio data received from Azure for sentence '%s'", sentence)
		}

		if state.AudioPlayer != nil {
			state.AudioPlayer.Play(audioData)
		} else {
			return fmt.Errorf("AudioPlayer not initialized")
		}
	}

	return nil
}
