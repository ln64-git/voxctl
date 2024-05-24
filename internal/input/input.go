package input

import (
	"encoding/json"

	"github.com/ln64-git/voxctl/internal/log"
	"github.com/ln64-git/voxctl/internal/speech"
)

func SanitizeInput(requestBody string) (string, error) {
	var req speech.PlayRequest
	err := json.Unmarshal([]byte(requestBody), &req)
	if err != nil {
		return "", err
	}

	log.Logger.Printf("text: %s", req.Text)
	return req.Text, nil
}
