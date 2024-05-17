package main

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/model"
)

func main() {
	action, input := "", ""
	port := 8080

	if len(os.Args) >= 2 {
		action = os.Args[1]
	}

	if action == "serve" && len(os.Args) >= 3 {
		if portNum, err := strconv.Atoi(os.Args[2]); err == nil {
			port = portNum
		}
	} else if action == "play" && len(os.Args) >= 3 {
		input = os.Args[2]
	}

	initialModel := model.InitialModel(action, input, port)
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
