package vosk

import (
	"fmt"
	"log"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
)

type SpeechRecognizer struct {
	model      *vosk.VoskModel
	recognizer *vosk.VoskRecognizer
	stream     *portaudio.Stream
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

	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open PortAudio stream: %v", err)
	}

	return &SpeechRecognizer{
		model:      model,
		recognizer: recognizer,
		stream:     stream,
	}, nil
}

func (sr *SpeechRecognizer) Start(resultChan chan<- string) error {
	err := sr.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start PortAudio stream: %v", err)
	}

	buffer := make([]int16, 16000)
	byteBuffer := make([]byte, len(buffer)*2)

	for {
		err := sr.stream.Read()
		if err != nil {
			return fmt.Errorf("failed to read from PortAudio stream: %v", err)
		}

		// Convert int16 buffer to byte buffer
		for i, v := range buffer {
			byteBuffer[2*i] = byte(v)
			byteBuffer[2*i+1] = byte(v >> 8)
		}

		if sr.recognizer.AcceptWaveform(byteBuffer) > 0 {
			result := sr.recognizer.Result()
			log.Println("Result:", result)
			resultChan <- result
		} else {
			partial := sr.recognizer.PartialResult()
			log.Println("Partial Result:", partial)
		}
	}
}

func (sr *SpeechRecognizer) Stop() {
	sr.stream.Stop()
	sr.stream.Close()
	sr.recognizer.Free()
	sr.model.Free()
	portaudio.Terminate()
}
