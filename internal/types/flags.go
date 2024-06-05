package types

// Flags represents the command-line flags used by the application.

type Flags struct {
	Port           *int
	Convo          *bool
	SpeakText      *string
	ChatText       *string
	ScribeStart    *bool
	ScribeStop     *bool
	ScribeToggle   *bool
	Status         *bool
	Stop           *bool
	Clear          *bool
	Quit           *bool
	Pause          *bool
	Resume         *bool
	TogglePlayback *bool
}
