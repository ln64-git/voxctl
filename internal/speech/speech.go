// speech/speech.go
package speech

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/elevenLabs"
	"github.com/ln64-git/voxctl/internal/types"
)

// SpeechRequest represents a request to synthesize speech.
type SpeechRequest struct {
	Text string `json:"text"`
}

// SanitizeInput removes unwanted characters from a string.
func SanitizeInput(input string) string {
	// Replace newlines, carriage returns, and tabs with a space
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\t", " ")

	// Replace multiple spaces with a single space
	input = strings.Join(strings.Fields(input), " ")

	return input
}

// SpeechRequestToJSON converts a SpeechRequest to a JSON string.
func (r SpeechRequest) SpeechRequestToJSON() string {
	sanitizedText := SanitizeInput(r.Text)
	return fmt.Sprintf(`{"text":"%s"}`, sanitizedText)
}

// ProcessSpeech processes the speech request by synthesizing and playing the speech.
func ProcessSpeech(req SpeechRequest, state types.AppState) error {
	sanitizedText := SanitizeInput(req.Text)
	segments := getSegmentedText(sanitizedText)
	if state.VoiceService == "ElevenLabs" {
		for _, segment := range segments {
			audioData, err := elevenLabs.SynthesizeSpeech(state.ElevenLabsSubscriptionKey, state.ElevenLabsRegion, segment, state.ElevenLabsGender, state.ElevenLabsVoice)
			if err != nil {
				log.Errorf("%s", err)
				return err
			}
			state.AudioPlayer.Play(audioData)
			log.Infof("Speech processed: %s", segment) // Example log message
		}
	} else if state.VoiceService == "Azure" {
		for _, segment := range segments {
			audioData, err := azure.SynthesizeSpeech(state.AzureSubscriptionKey, state.AzureRegion, segment, state.AzureVoiceGender, state.AzureVoiceName)
			if err != nil {
				log.Errorf("%s", err)
				return err
			}
			state.AudioPlayer.Play(audioData)
			log.Infof("Speech processed: %s", segment) // Example log message
		}
	} else {
		log.Info("No Subscription Key found in ~/voxctl.json")
	}
	return nil
}

// getSegmentedText splits text into segments based on punctuation.
func getSegmentedText(text string) []string {
	var sentences []string
	var currentSentence string
	for i, char := range text {
		if char == ',' || char == '.' || char == '!' || char == '?' {
			if currentSentence != "" {
				sentences = append(sentences, currentSentence)
				currentSentence = ""
			}
		} else {
			currentSentence += string(char)
			if i == len(text)-1 && currentSentence != "" {
				sentences = append(sentences, currentSentence)
			}
		}
	}
	return sentences
}
