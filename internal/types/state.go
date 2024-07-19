package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
)

// State struct to hold program state
type AppState struct {
	ClientPort                int
	ClientInput               string
	ServerStatusRequested     bool
	ServerQuitRequested       bool
	ServerPauseRequested      bool
	ServerStopRequested       bool
	VoiceService              string
	ElevenLabsSubscriptionKey string
	ElevenLabsRegion          string
	ElevenLabsGender          string
	ElevenLabsVoice           string
	AzureSubscriptionKey      string
	AzureRegion               string
	AzureVoiceGender          string
	AzureVoiceName            string
	AudioPlayer               *audio.AudioPlayer
	ServerAlreadyRunning      bool
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
