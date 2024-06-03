package convo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func HandleConversation(state *types.AppState) {
	log.Info("HandleConversation - state.SpeakText:")
	log.Info(state.SpeakText)

	switch strings.TrimSpace(state.SpeakText) {
	case "stop":
		state.AudioPlayer.Stop()
	case "pause":
		state.AudioPlayer.Pause()
	case "resume":
		state.AudioPlayer.Resume()
	case "clear":
		state.AudioPlayer.Clear()
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

			req, err := http.NewRequest("POST", "http://localhost:"+strconv.Itoa(state.Port)+"/ollama", strings.NewReader(string(body)))
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
			state.SpeakText = ""
		}()
	}
}
