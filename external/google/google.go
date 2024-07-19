package google

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

const (
	apiEndpoint = "https://texttospeech.googleapis.com/v1/text:synthesize"
)

type SynthesizeRequest struct {
	Input struct {
		Text string `json:"text"`
	} `json:"input"`
	Voice struct {
		LanguageCode string `json:"languageCode"`
		Name         string `json:"name,omitempty"`
	} `json:"voice"`
	AudioConfig struct {
		AudioEncoding string `json:"audioEncoding"`
	} `json:"audioConfig"`
}

type SynthesizeResponse struct {
	AudioContent string `json:"audioContent"`
}

func SynthesizeSpeech(apiKey, text, languageCode, voiceName string) ([]byte, error) {
	log.Infof("apiKey: %s", apiKey)
	log.Infof("text: %s", text)
	log.Infof("languageCode: %s", languageCode)
	log.Infof("voiceName: %s", voiceName)

	requestBody := SynthesizeRequest{
		Input: struct {
			Text string `json:"text"`
		}{
			Text: text,
		},
		Voice: struct {
			LanguageCode string `json:"languageCode"`
			Name         string `json:"name,omitempty"`
		}{
			LanguageCode: languageCode,
			Name:         voiceName,
		},
		AudioConfig: struct {
			AudioEncoding string `json:"audioEncoding"`
		}{
			AudioEncoding: "MP3",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := fmt.Sprintf("%s?key=%s", apiEndpoint, apiKey)
	headers := map[string]string{
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
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status: %s, body: %s", resp.Status, string(errorBody))
	}

	// Read and decode the response
	var synthesizeResponse SynthesizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&synthesizeResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	audioData, err := base64.StdEncoding.DecodeString(synthesizeResponse.AudioContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio content: %v", err)
	}

	return audioData, nil
}
