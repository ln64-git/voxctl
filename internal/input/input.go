// input/input.go
package input

import (
	"strings"
)

// ParseTextFromRequest extracts the text field from a JSON request body
func SanitizeInput(requestBody string) (string, error) {
	// Remove all extra characters
	bodyString := strings.ReplaceAll(requestBody, "\n", "")
	bodyString = strings.ReplaceAll(bodyString, "\t", "")

	// Trim leading and trailing whitespace
	text := strings.TrimSpace(bodyString)

	return text, nil
}
