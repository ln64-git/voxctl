package scribe

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/function/convo"
	"github.com/ln64-git/voxctl/internal/state"
	"github.com/sirupsen/logrus"
)

type TextResponse struct {
	Text string `json:"text"`
}

func ScribeText(state *state.AppState) {
	for result := range state.ScribeTextChan {
		var textResult TextResponse
		err := json.Unmarshal([]byte(result), &textResult)
		if err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			continue
		}
		text := strings.TrimSpace(textResult.Text)
		if text != "" {
			state.SpeakText += text + " "
			if state.ConversationMode {
				convo.HandleConversation(state)
			}
		}
	}
}

func ScribeStart(state *state.AppState) {
	go func() {
		err := state.SpeechRecognizer.Start(state.ScribeTextChan)
		if err != nil {
			logrus.Errorf("Error during speech recognition: %v", err)
		}
	}()
	log.Infof("SpeechInput Starting")
	state.ScribeStatus = true
}

func ScribeStop(state *state.AppState) {
	go func() {
		state.SpeechRecognizer.Stop()
	}()
	log.Infof("SpeechInput Stopped")
	state.ScribeStatus = false
}
