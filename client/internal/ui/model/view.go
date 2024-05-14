package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	outputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#acacac")).Background(lipgloss.Color("#283460")).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#02af78")).Bold(true)
	optionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bababa"))
	selectedOptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#646cd4"))
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6a6a6a"))

	var sb strings.Builder

	// Render the latest output message
	if len(m.messages) > 0 {
		msg := m.messages[len(m.messages)-1] // Get the last message
		sb.WriteString(outputStyle.Render(" " + msg + " "))
	} else {
		sb.WriteString(outputStyle.Render(" Select Option "))
	}
	if m.textInput != "" {
		sb.WriteString(optionStyle.Render("\n" + m.textInput))
	} else {
		sb.WriteString(outputStyle.Render("\n"))
	}
	sb.WriteString(outputStyle.Render("\n"))

	// Render choices with cursor and styles
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render("-")
			choice = selectedOptionStyle.Render(choice)
		} else {
			choice = optionStyle.Render(choice)
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	sb.WriteString("\n" + footerStyle.Render("Press 'ctrl+c' to exit."))

	return sb.String()
}
