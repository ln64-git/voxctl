package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/model"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <action> <input>")
		os.Exit(1)
	}

	action := os.Args[1]
	input := os.Args[2]

	initialModel := model.InitialModel(action, input)

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
