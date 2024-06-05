package chat

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/external/ollama"
	"github.com/ln64-git/voxctl/internal/audio/audioplayer"
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
	go segmentTextFromChannel(tokenChan, sentenceChan)

	var audioEntry []audioplayer.AudioEntry
	var fullText []string

	go func() {
		for sentence := range sentenceChan {
			audioData, err := azure.SynthesizeSpeech(state.AzureConfig.SubscriptionKey, state.AzureConfig.Region, sentence, state.AzureConfig.VoiceGender, state.AzureConfig.VoiceName)
			if err != nil {
				log.Errorf("Failed to synthesize speech: %v", err)
				return
			}
			fullText = append(fullText, sentence)
			audioEntry = append(audioEntry, audioplayer.AudioEntry{
				AudioData:   audioData,
				SegmentText: sentence,
				FullText:    fullText,
				ChatQuery:   req.Prompt,
			})
			state.AudioConfig.AudioEntries = append(state.AudioConfig.AudioEntries, audioEntry...)
		}
	}()
}

func segmentTextFromChannel(tokenChan <-chan string, sentenceChan chan<- string) {
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
