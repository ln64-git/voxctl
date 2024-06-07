package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/features/chat"
	"github.com/ln64-git/voxctl/internal/state"
)

func HandleChatRequest(w http.ResponseWriter, r *http.Request, state *state.AppState) {
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

	// Delegate chat processing to a separate function
	chat.ProcessChat(state, &req)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
