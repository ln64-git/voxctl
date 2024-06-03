package types

// AppStatusState holds the status of the application server
type AppStatusState struct {
	Port                 int  `json:"port"`
	ServerAlreadyRunning bool `json:"serverAlreadyRunning"`
	SpeakStatus          bool `json:"toggleSpeechStatus"`
}
