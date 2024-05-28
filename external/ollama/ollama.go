package ollama

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OllamaRequest struct {
	Model   string
	Prompt  string
	Preface string
}

// OllamaResponse represents the structure of the response received from the Ollama API.
type OllamaResponse struct {
	Model      string `json:"model"`
	CreatedAt  string `json:"created_at"`
	Response   string `json:"response"`
	Done       bool   `json:"done"`
	DoneReason string `json:"done_reason,omitempty"`
}

// GetOllamaTokenResponse generates a response token from Ollama.
// returns a channel that streams the response tokens and any encountered error.
func GetOllamaTokenResponse(model string, prompt string, port ...int) (<-chan string, error) {
	// Set default port if not provided
	var portValue int
	if len(port) > 0 {
		portValue = port[0]
	} else {
		portValue = 11434
	}

	// Prepare the request URL and payload
	url := fmt.Sprintf("http://localhost:%d/api/generate", portValue)
	payload := strings.NewReader(fmt.Sprintf(`{"model": "%s","prompt":"%s"}`, model, prompt))

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	// Set the request headers
	req.Header.Add("Content-Type", "application/json")

	// Execute the HTTP request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Channel to stream response tokens
	tokenChan := make(chan string)

	// Goroutine to process the response body and stream tokens
	go func() {
		defer res.Body.Close()
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			var response OllamaResponse
			if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
				close(tokenChan)
				return
			}
			tokenChan <- response.Response
			if response.Done {
				break
			}
		}
		close(tokenChan)
	}()

	return tokenChan, nil
}
