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
	if m.userPause {
		m.state.SetStatus("Ready")
		return m, handlePause(m)
	}
	if m.userStop {
		m.state.SetStatus("Ready")
		return m, handleStop(m)
	}

	if m.userRequest {
		if m.userInput != "" {
			m.sendPlayRequest()
			if m.userQuit {
				return m, tea.Quit
			}
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
	case pausedMsg:
		m.state.SetStatus("Paused")
		return m, nil
	case stoppedMsg:
		m.state.SetStatus("Stopped")
		return m, nil
	}

	return m, nil
}

func handlePause(m model) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		m.sendPauseRequest()
		return tea.Quit
	})
}

func handleStop(m model) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		m.sendStopRequest()
		return tea.Quit
	})
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

func (m model) sendPauseRequest() tea.Msg {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/pause", m.userPort), nil)
	if err != nil {
		return errMsg{err}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errMsg{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errMsg{fmt.Errorf("server returned status code %d", resp.StatusCode)}
	}

	return pausedMsg{}
}

func (m model) sendStopRequest() tea.Msg {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/stop", m.userPort), nil)
	if err != nil {
		return errMsg{err}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errMsg{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errMsg{fmt.Errorf("server returned status code %d", resp.StatusCode)}
	}

	return stoppedMsg{}
}

type pausedMsg struct{}
type stoppedMsg struct{}
type errMsg struct{ err error }
type playedMsg struct{}
