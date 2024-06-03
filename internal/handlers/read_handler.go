package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/ln64-git/voxctl/internal/utils/read"
)

func HandleReadText(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	speechReq, err := read.ProcessAzureRequest(r)
	if err != nil {
		log.Errorf("Failed to process speech request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = read.ReadText(*speechReq, state.AzureSubscriptionKey, state.AzureRegion, state.AudioPlayer)
	if err != nil {
		log.Errorf("Failed to process speech: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
