package handlers

import (
	"net/http"

	"github.com/ln64-git/voxctl/internal/function/scribe"
	"github.com/ln64-git/voxctl/internal/types"
)

func HandleScribeStart(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scribe.ScribeStart(state)
	w.WriteHeader(http.StatusOK)
}

func HandleScribeStop(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scribe.ScribeStop(state)
	w.WriteHeader(http.StatusOK)
}

func HandleScribeToggle(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if state.ScribeStatus {
		scribe.ScribeStop(state)
	} else {
		scribe.ScribeStart(state)
	}
	w.WriteHeader(http.StatusOK)
}
