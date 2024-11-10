package config

import (
	"encoding/json"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

// GetConfig retrieves the configuration from a JSON file in the user's home directory.
func GetConfig() map[string]interface{} {
	var cfg map[string]interface{}

	// Get current user's home directory
	user, err := user.Current()
	if err != nil {
		log.Fatalf("error getting user's home directory: %v", err)
	}

	configName := "voxctl.json"

	// Construct file path
	configFile := filepath.Join(user.HomeDir, configName)

	// Check if the configuration file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("configuration file does not exist: %s", configFile)
	} else if err != nil {
		log.Fatalf("error checking configuration file: %v", err)
	}

	// Load configuration
	err = readConfig(configFile, &cfg)
	if err != nil {
		log.Fatalf("error reading configuration: %v", err)
	}

	return cfg
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
