package types

import (
	"github.com/ln64-git/voxctl/internal/audio/player"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
)

// State struct to hold program state
type AppState struct {
	Port                  int
	ReadText              string
	AudioPlayer           *player.AudioPlayer
	ServerAlreadyRunning  bool
	ConversationMode      bool
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
	SpeakTextChan         chan string
	SpeechRecognizer      vosk.SpeechRecognizer
	SpeakText             string
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
