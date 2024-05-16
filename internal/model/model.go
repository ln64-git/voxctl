package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/external/azure"
	"github.com/ln64-git/voxctl/internal/audio"
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "enter":
			if m.userInput != "" {
				m.status = "Synthesizing..."
				return m, tea.Cmd(func() tea.Msg {
					audioData, err := azure.SynthesizeSpeech(m.azureSubscriptionKey, m.azureRegion, m.userInput, m.azureVoiceGender, m.azureVoiceName)
					if err != nil {
						return errMsg{err}
					}
					return synthMsg{audioData}
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
		m.status = "Error"
		m.err = msg.err
	case synthMsg:
		m.status = "Playing"
		return m, tea.Cmd(func() tea.Msg {
			err := audio.PlayAudio(msg.audioData)
			if err != nil {
				return errMsg{err}
			}
			return playedMsg{}
		})
	case playedMsg:
		m.status = "Ready"
		m.userInput = ""
	}

	if m.userAction == "play" && m.status == "Ready" {
		m.status = "Synthesizing..."
		return m, tea.Cmd(func() tea.Msg {
			audioData, err := azure.SynthesizeSpeech(m.azureSubscriptionKey, m.azureRegion, m.userInput, m.azureVoiceGender, m.azureVoiceName)
			if err != nil {
				return errMsg{err}
			}
			return synthMsg{audioData}
		})
	}
	return m, nil
}

type errMsg struct{ err error }
type synthMsg struct{ audioData []byte }
type playedMsg struct{}
