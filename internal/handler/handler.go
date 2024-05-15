package handler

import (
	"net/http"
	"os"

	"github.com/ln64-git/voxctl/internal/speech"
)

type Handler struct {
	speechService *speech.Service
}

func NewHandler(speechService *speech.Service) *Handler {
	return &Handler{
		speechService: speechService,
	}
}

func (h *Handler) Play(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("text")
	apiKey := r.FormValue("apiKey")
	region := r.FormValue("region")
	voiceGender := r.FormValue("voiceGender")
	voiceName := r.FormValue("voiceName")

	if apiKey == "" {
		apiKey = os.Getenv("AZURE_API_KEY")
	}
	if region == "" {
		region = "eastus"
	}
	if voiceGender == "" {
		voiceGender = "Female"
	}
	if voiceName == "" {
		voiceName = "en-US-JennyNeural"
	}

	err := h.speechService.Play(text, apiKey, region, voiceGender, voiceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Pause(w http.ResponseWriter, r *http.Request) {
	err := h.speechService.Pause()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Resume(w http.ResponseWriter, r *http.Request) {
	err := h.speechService.Resume()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Stop(w http.ResponseWriter, r *http.Request) {
	err := h.speechService.Stop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
