package handlers

import (
	"net/http"

	"github.com/ln64-git/voxctl/internal/features/scribe"
	"github.com/ln64-git/voxctl/internal/state"
)

func HandleScribeStart(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scribe.ScribeStart(state)
	w.WriteHeader(http.StatusOK)
}

func HandleScribeStop(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	scribe.ScribeStop(state)
	w.WriteHeader(http.StatusOK)
}

func HandleScribeToggle(w http.ResponseWriter, r *http.Request, state *state.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if state.ScribeConfig.ScribeStatus {
		scribe.ScribeStop(state)
	} else {
		scribe.ScribeStart(state)
	}
	w.WriteHeader(http.StatusOK)
}
