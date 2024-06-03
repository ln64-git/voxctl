package player

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/faiface/beep"
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
	initialized     bool
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

	if len(ap.audioQueue) == 0 {
		// Entire queue is finished, signal completion
		ap.isAudioPlaying = false
		select {
		case <-ap.doneChannel:
		default:
			close(ap.doneChannel)
		}
		ap.mutex.Unlock()
		return
	}

	audioData := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]

	audioReader := bytes.NewReader(audioData)
	audioReadCloser := io.NopCloser(audioReader)

	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		log.Errorf("Error decoding audio data: %v", err)
		ap.mutex.Unlock()
		ap.playNextAudioChunkIfAvailable()
		return
	}
	defer audioStreamer.Close()

	if !ap.initialized {
		ap.audioFormat = format
		err = speaker.Init(ap.audioFormat.SampleRate, ap.audioFormat.SampleRate.N(time.Second/10))
		if err != nil {
			log.Errorf("Error initializing speaker: %v", err)
			ap.mutex.Unlock()
			ap.playNextAudioChunkIfAvailable()
			return
		}
		ap.initialized = true
	}

	ap.audioController = &beep.Ctrl{Streamer: audioStreamer, Paused: false}
	ap.mutex.Unlock()

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)

	speaker.Play(beep.Seq(ap.audioController, beep.Callback(func() {
		waitGroup.Done()
	})))

	waitGroup.Wait()
	ap.playNextAudioChunkIfAvailable()
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
		speaker.Lock()
		ap.audioController.Paused = true
		speaker.Unlock()
	}
}

func (ap *AudioPlayer) Resume() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		speaker.Lock()
		ap.audioController.Paused = false
		speaker.Unlock()
	}
}

func (ap *AudioPlayer) Stop() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Silence the audio controller
	if ap.audioController != nil {
		speaker.Lock()
		ap.audioController.Paused = true
		speaker.Unlock()
	}

	// Clear the audio queue and reset state
	ap.audioQueue = nil
	ap.isAudioPlaying = false
}

func (ap *AudioPlayer) WaitForCompletion() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	// Check if audio is already finished
	if !ap.isAudioPlaying {
		return
	}

	// Wait on the done channel to signal completion
	<-ap.doneChannel
}

func (ap *AudioPlayer) IsPlaying() bool {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	return ap.isAudioPlaying
}

func (ap *AudioPlayer) SetIsPlaying(isPlaying bool) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	ap.audioController.Paused = isPlaying
}

func (ap *AudioPlayer) Clear() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	ap.audioQueue = nil
	ap.audioController = nil
	ap.audioFormat = beep.Format{}
	ap.isAudioPlaying = false
	ap.initialized = false

	// Ensure the channel is closed only once
	select {
	case <-ap.doneChannel:
		// Channel is already closed
	default:
		close(ap.doneChannel)
	}

	ap.doneChannel = make(chan struct{})
}
