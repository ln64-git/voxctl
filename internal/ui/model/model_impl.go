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
		choices:  []string{"serve", "play", "stop", "pause", "resume", "clear", "input"},
		selected: -1,
	}
}
