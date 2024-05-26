package input

import (
	"encoding/json"

	"github.com/ln64-git/voxctl/internal/speech"
)

func SanitizeInput(requestBody string) (string, error) {
	var req speech.SpeechRequest
	err := json.Unmarshal([]byte(requestBody), &req)
	if err != nil {
		return "", err
	}

	return req.Text, nil
}
