package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/audio"
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

			audioData, err := azure.SynthesizeSpeech(azureSubscriptionKey, azureRegion, req.Text, req.Gender, req.VoiceName)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to synthesize speech: %v", err), http.StatusInternalServerError)
				return
			}

			err = audio.PlayAudio(audioData)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to play audio: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Speech synthesized and played successfully")
		})
		addr := ":" + strconv.Itoa(port)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			state.SetStatus(fmt.Sprintf("Failed to start server: %v", err))
		}
		state.SetStatus("Ready")
	}()
}
