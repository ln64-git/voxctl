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
