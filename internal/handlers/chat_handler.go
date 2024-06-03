package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/function/chat"
	"github.com/ln64-git/voxctl/internal/types"
)

func HandleChatRequest(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	var req ollama.OllamaRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		log.Errorf("Failed to decode request body: %v", err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	ollamaReq := &req

	finalPrompt := ollamaReq.Preface + ollamaReq.Prompt

	tokenChan, err := ollama.GetOllamaTokenResponse(ollamaReq.Model, finalPrompt)
	if err != nil {
		log.Errorf("Failed to get Ollama token response: %v", err)
		http.Error(w, "Failed to get Ollama token response", http.StatusInternalServerError)
		return
	}

	sentenceChan := make(chan string)
	go chat.SegmentTextFromChannel(tokenChan, sentenceChan)

	go func() {
		for sentence := range sentenceChan {
			audioData, err := azure.SynthesizeSpeech(state.AzureSubscriptionKey, state.AzureRegion, sentence, state.AzureVoiceGender, state.AzureVoiceName)
			if err != nil {
				log.Errorf("Failed to synthesize speech: %v", err)
				return
			}
			state.AudioPlayer.Play(audioData)
		}
	}()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
