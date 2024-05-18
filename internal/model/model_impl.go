package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/types"
)

type model struct {
	userRequest          bool
	userInput            string
	userPort             int
	azureSubscriptionKey string
	azureRegion          string
	azureVoiceGender     string
	azureVoiceName       string
	err                  error
	state                *types.State
}

func InitialModel(input string, port int) model {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	voiceGender := os.Getenv("VOICE_GENDER")
	voiceName := os.Getenv("VOICE_NAME")
	subscriptionKey := os.Getenv("AZURE_SUBSCRIPTION_KEY")
	region := os.Getenv("AZURE_REGION")

	userRequest := false

	audioPlayer := audio.NewAudioPlayer()
	state := &types.State{AudioPlayer: audioPlayer, Status: "Starting..."}

	go server.StartServer(port, subscriptionKey, region, state)

	return model{
		userRequest:          userRequest,
		userInput:            input,
		userPort:             port,
		azureSubscriptionKey: subscriptionKey,
		azureRegion:          region,
		azureVoiceGender:     voiceGender,
		azureVoiceName:       voiceName,
		err:                  nil,
		state:                state,
	}
}
