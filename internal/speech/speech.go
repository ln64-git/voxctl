// speech/speech.go
package speech

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/audio"
)

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

func (r SpeechRequest) SpeechRequestToJSON() string {
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, r.Text, r.Gender, r.VoiceName)
}

func ProcessSpeech(req SpeechRequest, azureSubscriptionKey, azureRegion string, audioPlayer *audio.AudioPlayer) error {
	segments := getSegmentedText(req.Text)
	for _, segment := range segments {
		audioData, err := azure.SynthesizeSpeech(azureSubscriptionKey, azureRegion, segment, req.Gender, req.VoiceName)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}

		audioPlayer.Play(audioData)
		log.Infof("Speech processed: %s", segment) // Example log message
	}
	return nil
}

func getSegmentedText(text string) []string {
	var sentences []string
	var currentSentence string

	for i, char := range text {
		if char == ',' || char == '.' || char == '!' || char == '?' {
			sentences = append(sentences, currentSentence)
			currentSentence = ""
		} else {
			currentSentence += string(char)
			if i == len(text)-1 {
				sentences = append(sentences, currentSentence)
			}
		}
	}

	return sentences
}
