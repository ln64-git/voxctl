package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/function/clipboard"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func HandleSpeakStart(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	go func() {
		err := state.SpeechRecognizer.Start(state.SpeakTextChan)
		if err != nil {
			logrus.Errorf("Error during speech recognition: %v", err)
		}
	}()
	log.Infof("SpeechInput Starting")
	state.SpeakStatus = true
	w.WriteHeader(http.StatusOK)
}

func HandleSpeakStop(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	go func() {
		state.SpeechRecognizer.Stop()
		clipboard.CopyToClipboard(state.SpeakText)
		state.SpeakText = ""
	}()
	log.Infof("SpeechInput Stopped")
	state.SpeakStatus = false
	w.WriteHeader(http.StatusOK)
}

func HandleSpeakToggle(w http.ResponseWriter, r *http.Request, state *types.AppState) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if state.SpeakStatus {
		HandleSpeakStop(w, r, state)
	} else {
		HandleSpeakStart(w, r, state)
	}
}
