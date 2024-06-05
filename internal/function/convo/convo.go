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
		state.AudioPlayer.Stop()
		log.Info("SpeakText - Stop -")
		state.SpeakText = ""
	case "pause":
		state.AudioPlayer.Pause()
		log.Info("SpeakText - Pause -")
		state.SpeakText = ""
	case "resume":
		state.AudioPlayer.Resume()
		log.Info("SpeakText - Resume -")
		state.SpeakText = ""
	case "clear":
		state.AudioPlayer.Clear()
		log.Info("SpeakText - Clear -")
		state.SpeakText = ""
	default:
		go func() {
			ollamaReq := ollama.OllamaRequest{
				Model:   state.OllamaModel,
				Prompt:  state.SpeakText,
				Preface: state.OllamaPreface,
			}
			body, err := json.Marshal(ollamaReq)
			if err != nil {
				logrus.Errorf("Failed to marshal Ollama request: %v", err)
				return
			}

			log.Infof("SpeakText - %s -", state.SpeakText)
			state.SpeakText = ""

			req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(state.Port)+"/chat", strings.NewReader(string(body)))
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
