package types

import (
	"github.com/ln64-git/voxctl/internal/audio/player"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
)

// State struct to hold program state

type AppState struct {
	Port                  int
	AudioPlayer           *player.AudioPlayer
	AudioEntries          []player.AudioEntry
	ServerAlreadyRunning  bool
	ConversationMode      bool
	SpeakText             string
	ChatText              string
	ScribeText            string
	StatusRequest         bool
	StopRequest           bool
	ClearRequest          bool
	PauseRequest          bool
	ResumeRequest         bool
	TogglePlaybackRequest bool
	QuitRequest           bool
	ScribeStartRequest    bool
	ScribeStopRequest     bool
	ScribeToggleRequest   bool
	ScribeStatus          bool
	VoskModelPath         string
	ScribeTextChan        chan string
	SpeechRecognizer      *vosk.SpeechRecognizer
	AzureSubscriptionKey  string
	AzureRegion           string
	AzureVoiceGender      string
	AzureVoiceName        string
	OllamaPort            int
	OllamaModel           string
	OllamaPreface         string
}

type SpeechRequest struct {
	Text      string `json:"text"`
	Gender    string `json:"gender"`
	VoiceName string `json:"voiceName"`
}
