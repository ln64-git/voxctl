package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/log"
)

// State struct to hold program state
type AppState struct {
	Port                 int
	Token                string
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
	Logger               *log.Logger
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
