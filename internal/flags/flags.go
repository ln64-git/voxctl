package flags

import (
	"flag"
)

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

func ParseFlags() *Flags {
	flags := &Flags{
		Port:           flag.Int("port", 8080, "Port number to connect or serve"),
		Convo:          flag.Bool("convo", false, "Start Conversation Mode"),
		SpeakText:      flag.String("speak", "", "User input for speech or ollama requests"),
		ChatText:       flag.String("chat", "", "Chat with AI through text"),
		ScribeStart:    flag.Bool("scribe_start", false, "Start listening for Speech input"),
		ScribeStop:     flag.Bool("scribe_stop", false, "Stop listening for Speech input"),
		ScribeToggle:   flag.Bool("scribe_toggle", false, "Toggle listening for Speech input"),
		Status:         flag.Bool("status", false, "Request info"),
		Stop:           flag.Bool("stop", false, "Stop audio playback"),
		Clear:          flag.Bool("clear", false, "Clear playback"),
		Quit:           flag.Bool("quit", false, "Exit application after request"),
		Pause:          flag.Bool("pause", false, "Pause audio playback"),
		Resume:         flag.Bool("resume", false, "Resume audio playback"),
		TogglePlayback: flag.Bool("playback_toggle", false, "Toggle audio playback"),
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flags.Convo = new(bool)
		*flags.Convo = true
	}
	return flags
}
