package speech

type SpeechService interface {
	GetSpeechResponse(text, apiKey, region, voiceGender, voiceName string) error
	Pause() error
	Resume() error
	Stop() error
}

type Service struct {
	speechService SpeechService
}

func NewService(speechService SpeechService) *Service {
	return &Service{
		speechService: speechService,
	}
}

func (s *Service) Play(text, apiKey, region, voiceGender, voiceName string) error {
	return s.speechService.GetSpeechResponse(text, apiKey, region, voiceGender, voiceName)
}

func (s *Service) Pause() error {
	return s.speechService.Pause()
}

func (s *Service) Resume() error {
	return s.speechService.Resume()
}

func (s *Service) Stop() error {
	return s.speechService.Stop()
}
