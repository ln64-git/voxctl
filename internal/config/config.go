package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	AzureSubscriptionKey string `json:"azure_subscription_key"`
	AzureRegion          string `json:"azure_region"`
	VoiceGender          string `json:"voice_gender"`
	VoiceName            string `json:"voice_name"`
}

func GetConfig() (Config, error) {
	var cfg Config

	// Get the user's home directory
	user, err := user.Current()
	if err != nil {
		return cfg, fmt.Errorf("error getting user's home directory: %v", err)
	}

	// Construct the path to the configuration file
	configFile := filepath.Join(user.HomeDir, "voxctl.json")

	// Load the configuration from the JSON file
	err = readConfig(configFile, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("error reading configuration: %v", err)
	}
	return cfg, nil
}

func readConfig(configFile string, cfg *Config) error {
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
