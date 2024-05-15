package speech

import (
	"bytes"
	"io"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type SpeechService interface {
	GetSpeechResponse(text, apiKey, region, voiceGender, voiceName string) ([]byte, error)
	Pause() error
	Resume() error
	Stop() error
}

type Service struct {
	speechService SpeechService
	audioStream   beep.StreamSeekCloser
	ctrl          *beep.Ctrl
}

func NewService(speechService SpeechService) *Service {
	return &Service{
		speechService: speechService,
	}
}

func (s *Service) Play(text, apiKey, region, voiceGender, voiceName string) error {
	audioContent, err := s.speechService.GetSpeechResponse(text, apiKey, region, voiceGender, voiceName)
	if err != nil {
		return err
	}

	audioReader := bytes.NewReader(audioContent)
	audioReadCloser := io.NopCloser(audioReader)
	audioStreamer, format, err := mp3.Decode(audioReadCloser)
	if err != nil {
		return err
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}

	s.audioStream = audioStreamer
	s.ctrl = &beep.Ctrl{Streamer: beep.Loop(-1, audioStreamer)}

	speaker.Play(s.ctrl)

	return nil
}

func (s *Service) Pause() error {
	speaker.Lock()
	s.ctrl.Paused = true
	speaker.Unlock()
	return nil
}

func (s *Service) Resume() error {
	speaker.Lock()
	s.ctrl.Paused = false
	speaker.Unlock()
	return nil
}

func (s *Service) Stop() error {
	speaker.Clear()
	s.audioStream.Close()
	return nil
}
