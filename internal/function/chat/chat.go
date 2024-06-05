package chat

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/models"
	"github.com/ln64-git/voxctl/internal/state"
)

func ProcessChat(state *state.AppState, req *ollama.OllamaRequest) {
	finalPrompt := req.Preface + req.Prompt

	tokenChan, err := ollama.GetOllamaTokenResponse(req.Model, finalPrompt)
	if err != nil {
		log.Errorf("Failed to get Ollama token response: %v", err)
		return
	}

	sentenceChan := make(chan string)
	go buildSentences(tokenChan, sentenceChan)

	for sentence := range sentenceChan {
		go func(sentence string) {
			audioData, err := azure.SynthesizeSpeech(state.AzureConfig.SubscriptionKey, state.AzureConfig.Region, sentence, state.AzureConfig.VoiceGender, state.AzureConfig.VoiceName)
			if err != nil {
				log.Errorf("Failed to synthesize speech: %v", err)
				return
			}

			// Create a new audio entry for the sentence
			audioEntry := models.AudioEntry{
				AudioData:   audioData,
				SegmentText: sentence,
				FullText:    append([]string{}, sentence), // Copy the sentence to FullText
				ChatQuery:   req.Prompt,
			}

			// Send the audio entry to AudioEntriesUpdate channel
			state.AudioConfig.AudioEntriesUpdate <- []models.AudioEntry{audioEntry}
		}(sentence)
	}
}

func buildSentences(tokenChan <-chan string, sentenceChan chan<- string) {
	defer close(sentenceChan)
	var builder strings.Builder

	for token := range tokenChan {
		// Remove newline characters from token
		token = strings.ReplaceAll(token, "\n", "")

		builder.WriteString(token)
		if strings.ContainsAny(token, ",.!?") {
			sentence := builder.String()
			sentenceChan <- sentence
			builder.Reset()
		}
	}
}
