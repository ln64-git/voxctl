package types

import (
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/vosk"
)

// State struct to hold program state
type AppState struct {
	Port                  int
	UserInput             string
	AudioPlayer           *audio.AudioPlayer
	ServerAlreadyRunning  bool
	StatusRequest         bool
	StopRequest           bool
	ClearRequest          bool
	PauseRequest          bool
	ResumeRequest         bool
	TogglePlaybackRequest bool
	QuitRequest           bool
	StartSpeechRequest    bool
	StopSpeechRequest     bool
	ToggleSpeechRequest   bool
	ToggleSpeechStatus    bool
	VoskModelPath         string
	SpeechInputChan       chan string
	SpeechRecognizer      vosk.SpeechRecognizer
	SpeechInput           string
	AzureSubscriptionKey  string
	AzureRegion           string
	AzureVoiceGender      string
	AzureVoiceName        string
	OllamaRequest         bool
	OllamaPort            int
	OllamaModel           string
	OllamaPreface         string
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
