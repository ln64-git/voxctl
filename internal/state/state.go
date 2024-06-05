package state

import (
	"fmt"
	"net"
	"time"

	"github.com/ln64-git/voxctl/config"
	"github.com/ln64-git/voxctl/internal/audio/player"
	"github.com/ln64-git/voxctl/internal/audio/vosk"
	"github.com/ln64-git/voxctl/internal/flags"
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
		Port:                  *flags.Port,
		ServerAlreadyRunning:  CheckServerRunning(*flags.Port),
		ConversationMode:      *flags.Convo,
		SpeakText:             *flags.SpeakText,
		ChatText:              *flags.ChatText,
		StatusRequest:         *flags.Status,
		StopRequest:           *flags.Stop,
		ClearRequest:          *flags.Clear,
		QuitRequest:           *flags.Quit,
		PauseRequest:          *flags.Pause,
		ResumeRequest:         *flags.Resume,
		TogglePlaybackRequest: *flags.TogglePlayback,
		ScribeStartRequest:    *flags.ScribeStart,
		ScribeStopRequest:     *flags.ScribeStop,
		ScribeToggleRequest:   *flags.ScribeToggle,
		ScribeTextChan:        make(chan string),
		VoskModelPath:         config.GetStringOrDefault(configData, "VoskModelPath", ""),
		AzureSubscriptionKey:  config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:           config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		AzureVoiceGender:      config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		AzureVoiceName:        config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		OllamaModel:           config.GetStringOrDefault(configData, "OllamaModel", "llama3"),
		OllamaPreface:         config.GetStringOrDefault(configData, "OllamaPreface", ""),
	}
}
