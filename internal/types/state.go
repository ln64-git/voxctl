package types

import "github.com/ln64-git/voxctl/internal/audio"

// State struct to hold program state
type AppState struct {
	Port                 int
	Input                string
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
