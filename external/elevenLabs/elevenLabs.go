package elevenLabs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const (
	apiEndpoint = "https://api.elevenlabs.io/v1/text-to-speech"
)

func SynthesizeSpeech(subscriptionKey, region, text, voiceGender, voiceName string) ([]byte, error) {
	ssml := generateSSML(text, voiceGender, voiceName)

	url := fmt.Sprintf("%s/%s/synthesize", apiEndpoint, region)
	headers := map[string]string{
		"Authorization": "Bearer " + subscriptionKey,
		"Content-Type":  "application/ssml+xml",
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(ssml)))
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
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return audioData, nil
}

func generateSSML(text, voiceGender, voiceName string) string {
	return fmt.Sprintf(`<speak version='1.0' xml:lang='en-US'>
                            <voice xml:lang='en-US' xml:gender='%s' name='%s'>
                                %s
                            </voice>
                        </speak>`, voiceGender, voiceName, text)
}
