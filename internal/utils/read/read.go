package read

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/audio/player"
)

// AzureSpeechRequest represents a request to synthesize speech.
type AzureSpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}

// ReadText processes the speech request by synthesizing and playing the speech.
func ReadText(req AzureSpeechRequest, azureSubscriptionKey, azureRegion string, audioPlayer *player.AudioPlayer) error {
	segments := segmentText(req.Text)
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

func ProcessAzureRequest(r *http.Request) (*AzureSpeechRequest, error) {
	var speechReq AzureSpeechRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&speechReq)
	if err != nil {
		return nil, err
	}
	return &speechReq, nil
}

func SegmentTextFromChannel(tokenChan <-chan string, sentenceChan chan<- string) {
	defer close(sentenceChan)
	var builder strings.Builder

	for token := range tokenChan {
		builder.WriteString(token)
		if strings.ContainsAny(token, ",.!?") {
			sentence := builder.String()
			sentenceChan <- sentence
			builder.Reset()
		}
	}
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
