package audio

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type AudioPlayer struct {
	audioQueue      [][]byte
	mutex           sync.Mutex
	audioController *beep.Ctrl
	doneChannel     chan struct{}
	audioFormat     beep.Format
	isAudioPlaying  bool
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		audioQueue:  make([][]byte, 0),
		doneChannel: make(chan struct{}),
	}
}

func (ap *AudioPlayer) Play(audioData []byte) {
	if ap == nil {
		log.Error("AudioPlayer is nil")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic: %v", r)
		}
	}()

	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	ap.audioQueue = append(ap.audioQueue, audioData)

	if !ap.isAudioPlaying {
		ap.isAudioPlaying = true
		go ap.playNextAudioChunk()
	}
}

func (ap *AudioPlayer) playNextAudioChunk() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) == 0 {
		ap.isAudioPlaying = false
		close(ap.doneChannel)
		return
	}

	audioData := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]

	audioReader := bytes.NewReader(audioData)
	audioReadCloser := io.NopCloser(audioReader)

	// Determine the format of the audio data
	var audioStreamer beep.StreamSeekCloser
	var format beep.Format
	var err error

	if isWAV(audioData) {
		audioStreamer, format, err = wav.Decode(audioReadCloser)
	} else {
		audioStreamer, format, err = mp3.Decode(audioReadCloser)
	}

	if err != nil {
		log.Errorf("Error decoding audio data: %v", err)
		ap.playNextAudioChunkIfAvailable()
		return
	}
	defer audioStreamer.Close()

	if ap.audioFormat == (beep.Format{}) {
		ap.audioFormat = format
		err = speaker.Init(ap.audioFormat.SampleRate, ap.audioFormat.SampleRate.N(time.Second/10))
		if err != nil {
			log.Errorf("Error initializing speaker: %v", err)
			ap.playNextAudioChunkIfAvailable()
			return
		}
	}

	ap.audioController = &beep.Ctrl{Streamer: audioStreamer, Paused: false}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	speaker.Play(beep.Seq(ap.audioController, beep.Callback(func() {
		waitGroup.Done()
	})))

	go func() {
		waitGroup.Wait()
		ap.playNextAudioChunkIfAvailable()
	}()
}

func (ap *AudioPlayer) playNextAudioChunkIfAvailable() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) > 0 {
		ap.isAudioPlaying = true
		go ap.playNextAudioChunk()
	} else {
		ap.isAudioPlaying = false
	}
}

func (ap *AudioPlayer) Pause() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		ap.audioController.Paused = true
	}
}

func (ap *AudioPlayer) Resume() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		ap.audioController.Paused = false
	}
}

func (ap *AudioPlayer) Stop() {
	speaker.Lock()
	defer speaker.Unlock()

	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		if closer, ok := ap.audioController.Streamer.(io.Closer); ok {
			closer.Close()
		}
		ap.doneChannel <- struct{}{}
		ap.isAudioPlaying = false
		ap.audioQueue = nil
	}
}

func (ap *AudioPlayer) WaitForCompletion() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if !ap.isAudioPlaying {
		return
	}

	<-ap.doneChannel
}

// isWAV checks if the audio data is in WAV format.
func isWAV(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "RIFF"
}
