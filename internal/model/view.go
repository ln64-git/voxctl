package model

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	defaultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a4a4a4"))

	statusReadyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#40b772"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4c4c4c"))
)

func (m model) View() string {
	var s string

	s += ("Server running on port " + fmt.Sprintf("%d", m.userPort) + "\n")
	s += ("Text to synthesize: " + m.userInput + "\n\n")
	s += ("Voice Gender: " + m.azureVoiceGender + "\n")
	s += ("Voice Name: " + m.azureVoiceName + "\n\n")

	status := m.state.GetStatus()
	if status == "Ready" {
		s += "Status: " + statusReadyStyle.Render("Ready") + "\n"
	} else {
		s += defaultStyle.Render("Status: " + status + "\n")
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render("Error: "+fmt.Sprintf("%v", m.err)) + "\n"
	} else {
		s += "\n" + footerStyle.Render("Press 'ctrl+c' to quit.") + "\n"
	}

	return s
}
