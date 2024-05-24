package speech

import (
	"fmt"

	"github.com/ln64-git/sandbox/external/azure"
	"github.com/ln64-git/sandbox/internal/audio"
	"github.com/ln64-git/sandbox/internal/log"
)

type PlayRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r PlayRequest) ToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func ParseAndPlay(req PlayRequest, azureSubscriptionKey, azureRegion string, audioPlayer *audio.AudioPlayer) error {
	var sentences []string
	var currentSentence string

	for i, char := range req.Text {
		if char == ',' {
			sentences = append(sentences, currentSentence)
			currentSentence = ""
		} else {
			currentSentence += string(char)
			if i == len(req.Text)-1 {
				sentences = append(sentences, currentSentence)
			}
		}
	}

	for _, sentence := range sentences {
		audioData, err := azure.SynthesizeSpeech(azureSubscriptionKey, azureRegion, sentence, req.Gender, req.VoiceName)
		if err != nil {
			log.Logger.Printf("Failed to synthesize speech: %v", err)
			return err
		}
		audioPlayer.Play(audioData)
	}

	return nil
}
