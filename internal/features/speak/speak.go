package speak

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/models"
	"github.com/ln64-git/voxctl/internal/state"
)

// AzureSpeechRequest represents a request to synthesize speech.
type AzureSpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

// SpeakText processes the speech request by synthesizing and playing the speech.
func SpeakText(req AzureSpeechRequest, state *state.AppState) error {
	segments := segmentText(req.Text)
	var fullText []string

	for _, segment := range segments {
		audioData, err := azure.SynthesizeSpeech(state.AzureConfig.SubscriptionKey, state.AzureConfig.Region, segment, req.Gender, req.VoiceName)
		if err != nil {
			log.Errorf("Failed to synthesize speech: %v", err)
			return err
		}
		fullText = append(fullText, segment)
		audioEntry := models.AudioEntry{
			AudioData:   audioData,
			SegmentText: segment,
			FullText:    fullText,
		}
		state.AudioConfig.AudioEntriesUpdate <- []models.AudioEntry{audioEntry}
		log.Infof("Speech processed: %s", segment) // Example log message
	}

	return nil
}

// segmentText splits text into segments based on punctuation.
func segmentText(text string) []string {
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

// sanitizeText sanitizes and escapes text for JSON compatibility.
func sanitizeText(input string) string {
	// Replace newlines, carriage returns, and tabs with a space
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\t", " ")

	// Replace multiple spaces with a single space
	input = strings.Join(strings.Fields(input), " ")

	// Escape double quotes
	input = strings.ReplaceAll(input, `"`, `\"`)

	return input
}

// AzureRequestToJSON converts a AzureSpeechRequest to a JSON string.
func (r AzureSpeechRequest) AzureRequestToJSON() string {
	sanitizedText := sanitizeText(r.Text)
	return fmt.Sprintf(`{"text":"%s","gender":"%s","voiceName":"%s"}`, sanitizedText, r.Gender, r.VoiceName)
}
