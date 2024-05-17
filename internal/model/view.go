package model

import (
	"fmt"
)

func (m model) View() string {
	s := fmt.Sprintf("Server running on port %d\n", m.userPort)
	s += fmt.Sprintf("Text to synthesize: %s\n\n", m.userInput)
	s += fmt.Sprintf("Voice Gender: %s\n", m.azureVoiceGender)
	s += fmt.Sprintf("Voice Name: %s\n\n", m.azureVoiceName)
	s += fmt.Sprintf("Status: %s\n", m.status)
	if m.err != nil {
		s += fmt.Sprintf("\nError: %v\n", m.err)
	} else {
		s += "\nPress 'ctrl+c' to quit.\n"
	}

	return s
}
