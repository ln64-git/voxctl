package types

// Flags represents the command-line flags used by the application.
type Flags struct {
	Port           *int    // Port number to connect or serve
	ReadText       *string // User input for speech or Ollama requests
	Convo          *bool   // User input for speech or Ollama requests
	SpeakStart     *bool   // Start listening for Speech input
	SpeakStop      *bool   // Stop listening for Speech input
	SpeakToggle    *bool   // Toggle listening for Speech input
	Status         *bool   // Request info
	Stop           *bool   // Stop audio playback
	Clear          *bool   // Clear playback
	Quit           *bool   // Exit application after request
	Pause          *bool   // Pause audio playback
	Resume         *bool   // Resume audio playback
	TogglePlayback *bool   // Toggle audio playback
	ChatText       *string // Request Ollama query
	OllamaModel    *string // Ollama model to use
	OllamaPreface  *string // Preface text for the Ollama prompt
	OllamaPort     *int    // Input for Ollama
}
