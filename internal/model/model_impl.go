package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type model struct {
	textInput       string
	voiceGender     string
	voiceName       string
	status          string
	err             error
	subscriptionKey string
	region          string
}

func InitialModel() model {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	voiceGender := os.Getenv("VOICE_GENDER")
	voiceName := os.Getenv("VOICE_NAME")
	subscriptionKey := os.Getenv("AZURE_SUBSCRIPTION_KEY")
	region := os.Getenv("AZURE_REGION")

	return model{
		textInput:       "",
		voiceGender:     voiceGender,
		voiceName:       voiceName,
		status:          "Ready",
		err:             nil,
		subscriptionKey: subscriptionKey,
		region:          region,
	}
}
