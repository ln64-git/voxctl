package elevenLabs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

const (
	apiEndpoint = "https://api.elevenlabs.io/v1/text-to-speech"
)

type VoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
	Style           float64 `json:"style,omitempty"`
	UseSpeakerBoost bool    `json:"use_speaker_boost,omitempty"`
}

type PronunciationDictionaryLocator struct {
	PronunciationDictionaryID string `json:"pronunciation_dictionary_id"`
	VersionID                 string `json:"version_id"`
}

type SynthesizeRequest struct {
	Text                            string                           `json:"text"`
	ModelID                         string                           `json:"model_id"`
	VoiceSettings                   VoiceSettings                    `json:"voice_settings"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	Seed                            int                              `json:"seed,omitempty"`
	PreviousText                    string                           `json:"previous_text,omitempty"`
	NextText                        string                           `json:"next_text,omitempty"`
	PreviousRequestIDs              []string                         `json:"previous_request_ids,omitempty"`
	NextRequestIDs                  []string                         `json:"next_request_ids,omitempty"`
}

func SynthesizeSpeech(subscriptionKey, voiceID, text string, voiceSettings VoiceSettings) ([]byte, error) {
	log.Infof("subscriptionKey: %s", subscriptionKey)
	log.Infof("voiceID: %s", voiceID)
	log.Infof("text: %s", text)
	log.Infof("voiceSettings: %+v", voiceSettings)

	requestBody := SynthesizeRequest{
		Text:          text,
		ModelID:       "eleven_monolingual_v1", // Ensure this is the correct model ID
		VoiceSettings: voiceSettings,
		// Add other optional fields if needed
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := fmt.Sprintf("%s/%s", apiEndpoint, voiceID)
	headers := map[string]string{
		"xi-api-key":   subscriptionKey,
		"Accept":       "audio/mpeg",
		"Content-Type": "application/json",
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read the response body for more details
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status: %s, body: %s", resp.Status, string(errorBody))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return audioData, nil
}
