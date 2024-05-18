package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/ln64-git/voxctl/internal/server"
)

type model struct {
	userInput            string
	userPort             int
	azureSubscriptionKey string
	azureRegion          string
	azureVoiceGender     string
	azureVoiceName       string
	status               string
	err                  error
	statusCh             <-chan string
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

	statusCh := server.StartServer(port, subscriptionKey, region)

	return model{
		userInput:            input,
		userPort:             port,
		azureSubscriptionKey: subscriptionKey,
		azureRegion:          region,
		azureVoiceGender:     voiceGender,
		azureVoiceName:       voiceName,
		status:               "Server starting...",
		err:                  nil,
		statusCh:             statusCh,
	}
}
