package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/model"
)

func main() {

	input := flag.String("play", "", "Input text to play")
	port := flag.Int("port", 8080, "Port number to connect or serve")
	quit := flag.Bool("quit", false, "Exit application after request")

	flag.Parse()

	initialModel := model.InitialModel(*input, *port, *quit)
	p := tea.NewProgram(initialModel)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
