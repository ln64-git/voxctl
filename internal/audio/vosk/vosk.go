package vosk

import (
	"fmt"
	"sync"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
)

type SpeechRecognizer struct {
	model      *vosk.VoskModel
	recognizer *vosk.VoskRecognizer
	stream     *portaudio.Stream
	stopChan   chan bool
	mu         sync.Mutex // to synchronize start and stop
}

var portAudioInitialized = false
var portAudioInitMu sync.Mutex

func NewSpeechRecognizer(modelPath string) (*SpeechRecognizer, error) {
	portAudioInitMu.Lock()
	defer portAudioInitMu.Unlock()

	if !portAudioInitialized {
		err := portaudio.Initialize()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize PortAudio: %v", err)
		}
		portAudioInitialized = true
	}

	// Load Vosk model
	model, err := vosk.NewModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vosk model: %v", err)
	}

	// Create Vosk recognizer
	recognizer, err := vosk.NewRecognizer(model, 16000)
	if err != nil {
		model.Free()
		return nil, fmt.Errorf("failed to create Vosk recognizer: %v", err)
	}

	return &SpeechRecognizer{
		model:      model,
		recognizer: recognizer,
		stopChan:   make(chan bool),
	}, nil
}

func (sr *SpeechRecognizer) Start(resultChan chan<- string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	stream, err := portaudio.OpenDefaultStream(1, 0, 16000, 0, sr.audioCallback(resultChan))
	if err != nil {
		return fmt.Errorf("failed to open PortAudio stream: %v", err)
	}

	sr.stream = stream

	err = sr.stream.Start()
	if err != nil {
		return fmt.Errorf("failed to start PortAudio stream: %v", err)
	}

	go func() {
		<-sr.stopChan // Wait until stop is called
		sr.mu.Lock()
		sr.stream.Stop()
		sr.stream.Close()
		sr.mu.Unlock()
	}()

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
			// } else {
			// 	partialResult := sr.recognizer.PartialResult()
			// 	resultChan <- partialResult
		}
	}
}

func (sr *SpeechRecognizer) Stop() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.stopChan <- true
	sr.recognizer.Reset()
}

// Call this function before exiting the program to properly terminate PortAudio
func TerminatePortAudio() {
	portAudioInitMu.Lock()
	defer portAudioInitMu.Unlock()

	if portAudioInitialized {
		portaudio.Terminate()
		portAudioInitialized = false
	}
}
