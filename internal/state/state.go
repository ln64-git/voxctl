package state

import (
	"fmt"
	"net"
	"time"

	"github.com/ln64-git/voxctl/config"
	"github.com/ln64-git/voxctl/internal/audio/audioplayer"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
	"github.com/ln64-git/voxctl/internal/flags"
	"github.com/ln64-git/voxctl/internal/models"
)

// AppState holds the overall application state.
type AppState struct {
	ServerConfig     ServerConfig
	AudioConfig      AudioConfig
	ScribeConfig     ScribeConfig
	AzureConfig      AzureConfig
	OllamaConfig     OllamaConfig
	ConversationMode bool
	SpeakText        string
	ChatText         string
}

// ServerConfig holds the server-related configuration.
type ServerConfig struct {
	Port                 int
	ServerRunning        bool
	ServerAlreadyRunning bool
}

// AudioConfig holds the audio-related configuration.
type AudioConfig struct {
	AudioPlayer        *audioplayer.AudioPlayer
	AudioEntries       []models.AudioEntry
	AudioEntriesUpdate chan []models.AudioEntry
}

// ScribeConfig holds the scribe-related configuration.
type ScribeConfig struct {
	ScribeText       string
	ScribeStatus     bool
	VoskModelPath    string
	ScribeTextChan   chan string
	SpeechRecognizer *vosk.SpeechRecognizer
}

// AzureConfig holds the Azure-related configuration.
type AzureConfig struct {
	SubscriptionKey string
	Region          string
	VoiceGender     string
	VoiceName       string
}

// OllamaConfig holds the Ollama-related configuration.
type OllamaConfig struct {
	Model   string
	Preface string
}

// CheckServerRunning checks if the server is already running on the specified port.
func CheckServerRunning(port int) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func InitializeAppState(flags *flags.Flags, configData map[string]interface{}) AppState {
	return AppState{
		ServerConfig: ServerConfig{
			Port:                 *flags.Port,
			ServerRunning:        false,
			ServerAlreadyRunning: CheckServerRunning(*flags.Port),
		},
		AudioConfig: AudioConfig{
			AudioPlayer:        &audioplayer.AudioPlayer{},
			AudioEntries:       []models.AudioEntry{},
			AudioEntriesUpdate: make(chan []models.AudioEntry),
		},
		ScribeConfig: ScribeConfig{
			ScribeTextChan: make(chan string),
			VoskModelPath:  config.GetStringOrDefault(configData, "VoskModelPath", ""),
		},
		AzureConfig: AzureConfig{
			SubscriptionKey: config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
			Region:          config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
			VoiceGender:     config.GetStringOrDefault(configData, "VoiceGender", "Female"),
			VoiceName:       config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		},
		OllamaConfig: OllamaConfig{
			Model:   config.GetStringOrDefault(configData, "OllamaModel", "llama3"),
			Preface: config.GetStringOrDefault(configData, "OllamaPreface", ""),
		},
		ConversationMode: *flags.Convo,
		SpeakText:        *flags.SpeakText,
		ChatText:         *flags.ChatText,
	}
}
