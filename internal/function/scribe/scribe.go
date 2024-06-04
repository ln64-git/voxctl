package scribe

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/function/convo"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func ScribeText(state *types.AppState) {
	for result := range state.ScribeTextChan {
		var textResult types.TextResponse
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

func ScribeStart(state *types.AppState) {
	go func() {
		err := state.SpeechRecognizer.Start(state.ScribeTextChan)
		if err != nil {
			logrus.Errorf("Error during speech recognition: %v", err)
		}
	}()
	log.Infof("SpeechInput Starting")
	state.ScribeStatus = true
}

func ScribeStop(state *types.AppState) {
	go func() {
		state.SpeechRecognizer.Stop()
	}()
	log.Infof("SpeechInput Stopped")
	state.ScribeStatus = false
}
