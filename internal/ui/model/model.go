package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/server"
)

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "up":
			if !m.inputFocused && m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if !m.inputFocused && m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.choices)-1 {
				m.inputFocused = !m.inputFocused
			} else if !m.inputFocused {
				m.selected = m.cursor

				choice := m.choices[m.selected]
				switch choice {
				case "serve":
					m.ServeHandler()
				case "play":
					m.PlayHandler()
				case "stop", "pause", "resume":
					m.ControlHandler(choice)
				case "clear":
					m.ClearHandler()
				}
			}
		default:
			if m.inputFocused {
				if msg.String() == "backspace" {
					if len(m.textInput) > 0 {
						m.textInput = m.textInput[:len(m.textInput)-1]
					}
				} else {
					m.textInput += msg.String()
				}
			}
		}
	}
	return m, nil
}

func (m *model) ServeHandler() {
	statusChan := make(chan server.ServerStatus)
	go func() {
		status := server.Start()
		statusChan <- status
	}()

	status := <-statusChan
	if status.Launched {
		m.messages = append(m.messages, fmt.Sprintf("Server successfully launched on port %d", status.Port))
	} else if status.Error != nil {
		m.messages = append(m.messages, fmt.Sprintf("Server error: %v", status.Error))
	} else {
		m.messages = append(m.messages, fmt.Sprintf("Server already running on port %d", status.Port))
	}
}

func (m *model) PlayHandler() {
	messageChan := make(chan string)
	go func() {
		payload := map[string]string{"text": m.textInput}
		jsonPayload, _ := json.Marshal(payload)

		resp, err := http.Post("http://localhost:3000/play", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			messageChan <- fmt.Sprintf("Request error: %v", err)
			close(messageChan)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			messageChan <- "play request successful"
		} else {
			messageChan <- fmt.Sprintf("Play request failed, status code: %d", resp.StatusCode)
		}
		close(messageChan)
	}()
	for msg := range messageChan {
		m.messages = append(m.messages, msg)
	}
}

func (m *model) ControlHandler(choice string) {
	messageChan := make(chan string)
	go func() {
		resp, err := http.Post(fmt.Sprintf("http://localhost:3000/%s", choice), "application/json", nil)
		if err != nil {
			messageChan <- fmt.Sprintf("Request error: %v", err)
			close(messageChan)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			messageChan <- fmt.Sprintf("%s request successful", choice)
		} else {
			messageChan <- fmt.Sprintf("%s request failed, status code: %d", choice, resp.StatusCode)
		}
		close(messageChan)
	}()
	for msg := range messageChan {
		m.messages = append(m.messages, msg)
	}
}

func (m *model) ClearHandler() {
	m.messages = nil
	m.textInput = ""
}
