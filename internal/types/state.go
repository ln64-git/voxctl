package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
)

// State struct to hold program state
type AppState struct {
	ClientPort  int
	ClientInput string

	ServerStatusRequested bool
	ServerQuitRequested   bool
	ServerPauseRequested  bool
	ServerStopRequested   bool

	VoiceService                   string
	ElevenLabsSubscriptionKey      string
	ElevenLabsVoiceModelID         string
	ElevenLabsVoiceStability       float64
	ElevenLabsVoiceSimilarityBoost float64
	ElevenLabsVoiceStyle           float64
	ElevenLabsVoiceUseSpeakerBoost bool

	AzureSubscriptionKey string
	AzureRegion          string
	AzureVoiceGender     string
	AzureVoiceName       string

	GoogleSubscriptionKey string
	GoogleLanguageCode    string
	GoogleVoiceName       string

	AudioPlayer          *audio.AudioPlayer
	ServerAlreadyRunning bool
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
