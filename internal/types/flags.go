package types

// Flags represents the command-line flags used by the application.
type Flags struct {
	Port           *int    // Port number to connect or serve
	SpeakText      *string // User input for speech or Ollama requests
	Convo          *bool   // User input for speech or Ollama requests
	ScribeStart    *bool   // Start listening for Speech input
	ScribeStop     *bool   // Stop listening for Speech input
	ScribeToggle   *bool   // Toggle listening for Speech input
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
