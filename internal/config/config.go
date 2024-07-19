package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

// GetBoolOrDefault retrieves a boolean value from the configuration map, or returns a default value if the key is not present or the value is not a boolean.
func GetBoolOrDefault(cfg map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := cfg[key]; ok {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

func GetFloat64OrDefault(configData map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := configData[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case string:
			if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
				return floatValue
			}
		default:
			log.Printf("Warning: Key %s is not a float64 or string. Using default value.", key)
		}
	}
	return defaultValue
}
