package speech

type Service interface {
	Play(text string)
	Pause()
	Resume()
	Stop()
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Play(text string) {
	// Implement the logic to play the provided text
}

func (s *service) Pause() {
	// Implement the logic to pause the current playback
}

func (s *service) Resume() {
	// Implement the logic to resume the paused playback
}

func (s *service) Stop() {
	// Implement the logic to stop the current playback
}
