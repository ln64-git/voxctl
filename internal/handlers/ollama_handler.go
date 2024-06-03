package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/ln64-git/voxctl/internal/utils/read"
)

func HandleOllamaRequest(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	ollamaReq, err := processOllamaRequest(r)
	if err != nil {
		log.Errorf("Failed to process Ollama request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	finalPrompt := ollamaReq.Preface + ollamaReq.Prompt

	tokenChan, err := ollama.GetOllamaTokenResponse(ollamaReq.Model, finalPrompt)
	if err != nil {
		log.Errorf("Failed to get Ollama token response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sentenceChan := make(chan string)
	go read.SegmentTextFromChannel(tokenChan, sentenceChan)

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

func processOllamaRequest(r *http.Request) (*ollama.OllamaRequest, error) {
	var req ollama.OllamaRequest
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	log.Infof("Raw request body: %s", string(bodyBytes))

	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request body: %v", err)
	}

	return &req, nil
}
