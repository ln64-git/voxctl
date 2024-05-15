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

func (s *Service) GetSpeechResponse(text, apiKey, region, voiceGender, voiceName string) error {
	s.apiKey = apiKey
	s.region = region
	s.voiceGender = voiceGender
	s.voiceName = voiceName

	tokenURL := fmt.Sprintf("https://%s.api.cognitive.microsoft.com/sts/v1.0/issueToken", s.region)
	ttsURL := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", s.region)

	// Get the access token
	tokenResp, err := http.Post(tokenURL, "", bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}
	defer tokenResp.Body.Close()

	accessToken, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return err
	}

	// Make the text-to-speech request
	body := fmt.Sprintf(`<speak version='1.0' xml:lang='en-US'><voice xml:lang='en-US' xml:gender='%s' name='%s'>%s</voice></speak>`, s.voiceGender, s.voiceName, text)

	req, err := http.NewRequest("POST", ttsURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("Authorization", "Bearer "+string(accessToken))

	client := &http.Client{}
	ttsResp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer ttsResp.Body.Close()

	audioContent, err := io.ReadAll(ttsResp.Body)
	if err != nil {
		return err
	}

	// Here, you can save the audioContent or stream it to the client
	// For now, we'll just print it to the console
	fmt.Printf("Azure speech response:\n%s\n", audioContent)

	return nil
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
