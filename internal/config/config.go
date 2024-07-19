package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// GetConfig retrieves the configuration from a JSON file in the user's home directory.
func GetConfig(configName string) (map[string]interface{}, error) {
	var cfg map[string]interface{}

	// Get current user's home directory
	user, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting user's home directory: %v", err)
	}

	// Construct file path
	configFile := filepath.Join(user.HomeDir, configName)

	// Load configuration
	err = readConfig(configFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration: %v", err)
	}

	return cfg, nil
}

// readConfig reads and unmarshals the configuration file.
func readConfig(configFile string, cfg *map[string]interface{}) error {
	// Read file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	// Unmarshal JSON
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}

// GetStringOrDefault retrieves a string value from the configuration map, or returns a default value if the key is not present or the value is not a string.
func GetStringOrDefault(cfg map[string]interface{}, key string, defaultValue string) string {
	if value, ok := cfg[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}
