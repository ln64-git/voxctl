package main

import (
	"fmt"

	"github.com/ln64-git/sandbox/internal/log"
)

func main() {
	// Initialize the logger
	err := log.InitLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer log.Logger.Writer()
	log.Logger.Println("main - Program Started")

	// input := flag.String("play", "", "Input text to play")
	// port := flag.Int("port", 8080, "Port number to connect or serve")
	// quit := flag.Bool("quit", false, "Exit application after request")
	// pause := flag.Bool("pause", false, "Pause audio playback")
	// stop := flag.Bool("stop", false, "Stop audio playback")

	// Wait for the audio to finish playing
	// audioPlayer.WaitForCompletion()

}

// audioPlayer := audio.NewAudioPlayer()

// audioData, err := os.ReadFile("public/sample.wav")
// if err != nil {
// 	logger.Logger.Printf("Failed to read audio file: %v\n", err)
// 	return
// }

// audioPlayer.Play(audioData)
