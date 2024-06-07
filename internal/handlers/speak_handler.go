package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/features/speak"
	"github.com/ln64-git/voxctl/internal/state"
)

func HandleSpeakText(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	// Process the Azure speech request
	var speechReq speak.AzureSpeechRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&speechReq)
	if err != nil {
		log.Errorf("Failed to process speech request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Read the text using the processed request
	err = speak.SpeakText(speechReq, state)
	if err != nil {
		log.Errorf("Failed to process speech: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
