package model

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/types"
)

type config struct {
	AzureSubscriptionKey string `json:"azure_subscription_key"`
	AzureRegion          string `json:"azure_region"`
	VoiceGender          string `json:"voice_gender"`
	VoiceName            string `json:"voice_name"`
}

type model struct {
	userPause            bool
	userStop             bool
	userQuit             bool
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

func InitialModel(input string, port int, quit bool, pause bool, stop bool) model {
	// Get the user's home directory
	user, err := user.Current()
	if err != nil {
		fmt.Println("Error getting user's home directory:", err)
		// Use a fallback directory or handle the error
	}

	// Construct the path to the configuration file
	configFile := filepath.Join(user.HomeDir, ".config", "voxctl", "config.json")

	// Load the configuration from the JSON file
	var cfg config
	err = readConfig(configFile, &cfg)
	if err != nil {
		fmt.Println("Error reading configuration:", err)
		// Use fallback values or handle the error
	}

	var userRequest bool
	if input != "" {
		userRequest = true
	} else {
		userRequest = false
	}

	audioPlayer := audio.NewAudioPlayer()
	state := &types.State{AudioPlayer: audioPlayer, Status: "Starting..."}

	go server.StartServer(port, cfg.AzureSubscriptionKey, cfg.AzureRegion, state)

	return model{
		userPause:            pause,
		userStop:             stop,
		userQuit:             quit,
		userRequest:          userRequest,
		userInput:            input,
		userPort:             port,
		azureSubscriptionKey: cfg.AzureSubscriptionKey,
		azureRegion:          cfg.AzureRegion,
		azureVoiceGender:     cfg.VoiceGender,
		azureVoiceName:       cfg.VoiceName,
		err:                  nil,
		state:                state,
	}
}

func readConfig(configFile string, cfg *config) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}
