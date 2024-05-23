package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ln64-git/voxctl/internal/model"
)

func main() {
	input := flag.String("play", "", "Input text to play")
	port := flag.Int("port", 8080, "Port number to connect or serve")
	quit := flag.Bool("quit", false, "Exit application after request")
	pause := flag.Bool("pause", false, "Pause audio playback")
	stop := flag.Bool("stop", false, "Stop audio playback")
	flag.Parse()

	initialModel := model.InitialModel(*input, *port, *quit)

	if *pause {
		pauseAudio(*port)
	}

	if *stop {
		stopAudio(*port)
	}

	if !*pause && !*stop {
		p := tea.NewProgram(initialModel)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}

func pauseAudio(port int) {
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/pause", port), "application/json", nil)
	if err != nil {
		fmt.Println("Error pausing audio:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(body))
}

func stopAudio(port int) {
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/stop", port), "application/json", nil)
	if err != nil {
		fmt.Println("Error stopping audio:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(body))
}
