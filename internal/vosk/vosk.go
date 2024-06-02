package vosk

import (
	"fmt"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
)

type SpeechRecognizer struct {
	model      *vosk.VoskModel
	recognizer *vosk.VoskRecognizer
	stream     *portaudio.Stream
	stopChan   chan bool
}

func NewSpeechRecognizer(modelPath string) (*SpeechRecognizer, error) {
	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vosk model: %v", err)
	}

	recognizer, err := vosk.NewRecognizer(model, 16000)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vosk recognizer: %v", err)
	}

	err = portaudio.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PortAudio: %v", err)
	}

	return &SpeechRecognizer{
		model:      model,
		recognizer: recognizer,
		stopChan:   make(chan bool),
	}, nil
}

func (sr *SpeechRecognizer) Start(resultChan chan<- string) error {
	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, 0, sr.audioCallback(resultChan))
	if err != nil {
		return fmt.Errorf("failed to open PortAudio stream: %v", err)
	}

	sr.stream = stream

	err = sr.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start PortAudio stream: %v", err)
	}

	<-sr.stopChan // Wait until stop is called
	return nil
}

func (sr *SpeechRecognizer) audioCallback(resultChan chan<- string) func([]int16) {
	return func(input []int16) {
		byteBuffer := make([]byte, len(input)*2)
		for i, v := range input {
			byteBuffer[2*i] = byte(v)
			byteBuffer[2*i+1] = byte(v >> 8)
		}
		if sr.recognizer.AcceptWaveform(byteBuffer) > 0 {
			result := sr.recognizer.Result()
			resultChan <- result
		} else {
			partialResult := sr.recognizer.PartialResult()
			resultChan <- partialResult
		}
	}
}

func (sr *SpeechRecognizer) Stop() {
	sr.stopChan <- true
	sr.stream.Stop()
	sr.stream.Close()
	sr.recognizer.Free()
	sr.model.Free()
	portaudio.Terminate()
}
