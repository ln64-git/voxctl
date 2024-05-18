package model

import (
	"fmt"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/server"
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.state.SetStatus("Ready")
	if m.userRequest {
		if m.userInput != "" {
			m.sendPlayRequest()
			// m.userInput = ""
			// m.userRequest = false
			return m, tea.Quit
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.userInput != "" {
				m.state.SetStatus("Synthesizing...")
				return m, tea.Cmd(func() tea.Msg {
					return m.sendPlayRequest()
				})
			}
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.userInput) > 0 {
				m.userInput = m.userInput[:len(m.userInput)-1]
			}
		default:
			m.userInput += msg.String()
		}
	case errMsg:
		m.state.SetStatus("Error")
		m.err = msg.err
		return m, nil
	case playedMsg:
		m.state.SetStatus("Ready")
		m.userInput = ""
		return m, nil
	}

	return m, nil
}

func (m model) sendPlayRequest() tea.Msg {
	req := server.PlayRequest{
		Text:      m.userInput,
		Gender:    m.azureVoiceGender,
		VoiceName: m.azureVoiceName,
	}
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/play", m.userPort), "application/json", strings.NewReader(req.ToJSON()))
	if err != nil {
		return errMsg{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errMsg{fmt.Errorf("server returned status code %d", resp.StatusCode)}
	}

	return playedMsg{}
}

type errMsg struct{ err error }
type playedMsg struct{}
