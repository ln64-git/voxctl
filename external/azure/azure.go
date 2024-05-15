package azure

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Service struct {
	apiKey      string
	region      string
	voiceGender string
	voiceName   string
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetSpeechResponse(text, apiKey, region, voiceGender, voiceName string) ([]byte, error) {
	s.apiKey = apiKey
	s.region = region
	s.voiceGender = voiceGender
	s.voiceName = voiceName

	tokenURL := fmt.Sprintf("https://%s.api.cognitive.microsoft.com/sts/v1.0/issueToken", s.region)
	ttsURL := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", s.region)

	// Get the access token
	tokenResp, err := http.Post(tokenURL, "", bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get access token: status code %d", tokenResp.StatusCode)
	}

	accessToken, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read access token: %v", err)
	}

	// Make the text-to-speech request
	body := fmt.Sprintf(`<speak version='1.0' xml:lang='en-US'><voice xml:lang='en-US' xml:gender='%s' name='%s'>%s</voice></speak>`, s.voiceGender, s.voiceName, text)

	req, err := http.NewRequest("POST", ttsURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS request: %v", err)
	}
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "audio-16khz-128kbitrate-mono-mp3")
	req.Header.Set("Authorization", "Bearer "+string(accessToken))

	client := &http.Client{}
	ttsResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make TTS request: %v", err)
	}
	defer ttsResp.Body.Close()

	if ttsResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS request failed: status code %d", ttsResp.StatusCode)
	}

	audioContent, err := io.ReadAll(ttsResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio content: %v", err)
	}

	return audioContent, nil
}

func (s *Service) Pause() error {
	// Implement pause functionality
	return nil
}

func (s *Service) Resume() error {
	// Implement resume functionality
	return nil
}

func (s *Service) Stop() error {
	// Implement stop functionality
	return nil
}
