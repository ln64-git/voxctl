package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ln64-git/voxctl/client/internal/speech"
)

type Handler struct {
	speechService speech.Service
}

func NewHandler(speechService speech.Service) *Handler {
	return &Handler{
		speechService: speechService,
	}
}

func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	h.speechService.Play(payload.Text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) {
	h.speechService.Pause()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	h.speechService.Resume()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	h.speechService.Stop()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
