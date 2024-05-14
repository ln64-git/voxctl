package model

type model struct {
	choices      []string
	cursor       int
	selected     int
	messages     []string
	textInput    string
	inputFocused bool
}

func InitialModel() model {
	return model{
		choices:  []string{"serve", "play", "pause", "resume", "clear", "stop", "input"},
		selected: -1,
	}
}
