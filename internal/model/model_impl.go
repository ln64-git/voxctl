package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type model struct {
	userAction           string
	userInput            string
	azureSubscriptionKey string
	azureRegion          string
	azureVoiceGender     string
	azureVoiceName       string
	status               string
	err                  error
}

func InitialModel(action, input string) model {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	voiceGender := os.Getenv("VOICE_GENDER")
	voiceName := os.Getenv("VOICE_NAME")
	subscriptionKey := os.Getenv("AZURE_SUBSCRIPTION_KEY")
	region := os.Getenv("AZURE_REGION")

	return model{
		userAction:           action,
		userInput:            input,
		azureSubscriptionKey: subscriptionKey,
		azureRegion:          region,
		azureVoiceGender:     voiceGender,
		azureVoiceName:       voiceName,
		status:               "Ready",
		err:                  nil,
	}
}
