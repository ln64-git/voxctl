package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
)

// State struct to hold program state
type AppState struct {
	Port                 int
	OllamaPort           int
	OllamaModel          string
	OllamaInput          string
	OllamaPreface        string
	SpeechInput          string
	StatusRequested      bool
	QuitRequested        bool
	PauseRequested       bool
	StopRequested        bool
	AzureSubscriptionKey string
	AzureRegion          string
	VoiceGender          string
	VoiceName            string
	AudioPlayer          *audio.AudioPlayer
	ServerAlreadyRunning bool
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
