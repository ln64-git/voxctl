package convo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/state"
	"github.com/sirupsen/logrus"
)

func HandleConversation(state *state.AppState) {
	switch strings.TrimSpace(state.SpeakText) {
	case "stop":
		state.AudioConfig.AudioPlayer.Stop()
		log.Info("HandleConversation - Stop - Called")
		state.SpeakText = ""
	case "pause":
		state.AudioConfig.AudioPlayer.Pause()
		log.Info("HandleConversation - Pause - Called")
		state.SpeakText = ""
	case "resume":
		state.AudioConfig.AudioPlayer.Resume()
		log.Info("HandleConversation - Resume - Called")
		state.SpeakText = ""
	case "clear":
		state.AudioConfig.AudioPlayer.Clear()
		log.Info("HandleConversation - Clear - Called")
		state.SpeakText = ""
	default:
		go func() {
			ollamaReq := ollama.OllamaRequest{
				Model:   state.OllamaConfig.Model,
				Prompt:  state.SpeakText,
				Preface: state.OllamaConfig.Preface,
			}
			body, err := json.Marshal(ollamaReq)
			if err != nil {
				logrus.Errorf("Failed to marshal Ollama request: %v", err)
				return
			}

			log.Infof("HandleConversation - %s -", state.SpeakText)
			state.SpeakText = ""

			req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(state.ServerConfig.Port)+"/chat", strings.NewReader(string(body)))
			if err != nil {
				logrus.Errorf("Error creating request: %v", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				logrus.Errorf("Error making request: %v", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				logrus.Errorf("Request failed with status: %v", resp.Status)
				return
			}
		}()
	}
}
