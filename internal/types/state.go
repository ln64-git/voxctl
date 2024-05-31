package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
)

// State struct to hold program state
type AppState struct {
	Port                    int
	UserInput               string
	AudioPlayer             *audio.AudioPlayer
	ServerAlreadyRunning    bool
	StatusRequested         bool
	StopRequested           bool
	ClearRequested          bool
	PauseRequested          bool
	ResumeRequested         bool
	TogglePlaybackRequested bool
	QuitRequested           bool
	AzureSpeechRequest      bool
	AzureSubscriptionKey    string
	AzureRegion             string
	AzureVoiceGender        string
	AzureVoiceName          string
	OllamaPort              int
	OllamaModel             string
	OllamaPreface           string
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
