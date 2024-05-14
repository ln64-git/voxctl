package model

import (
	"fmt"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ln64-git/voxctl/internal/server"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
	messages []string
}

func InitialModel() model {
	return model{
		choices:  []string{"serve", "play", "stop", "pause", "resume", "clear"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

			choice := m.choices[m.cursor]
			switch choice {
			case "serve":
				go func() {
					err := server.Start()
					if err != nil {
						m.messages = append(m.messages, fmt.Sprintf("Server error: %v", err))
					}
				}()
				m.messages = append(m.messages, "Server started")

			case "play", "stop", "pause", "resume":
				go func() {
					resp, err := http.Post(fmt.Sprintf("http://localhost:3000/%s", choice), "application/json", nil)
					if err != nil {
						m.messages = append(m.messages, fmt.Sprintf("Request error: %v", err))
						return
					}
					defer resp.Body.Close()
				}()
				m.messages = append(m.messages, fmt.Sprintf("%s request sent", choice))

			case "clear":
				m.messages = nil
			}
		}
	}

	return m, nil
}

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
		sb.WriteString(outputStyle.Render("  " + msg + "  "))
	} else {
		sb.WriteString(outputStyle.Render("   Select Option   "))
	}

	sb.WriteString(outputStyle.Render("\n\n"))

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

	sb.WriteString("\n" + footerStyle.Render("Press 'q' to quit."))

	return sb.String()
}
