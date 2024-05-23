package audio

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type AudioPlayer struct {
	audioQueue      [][]byte
	mutex           sync.Mutex
	audioController *beep.Ctrl
	audioFormat     beep.Format
	isAudioPlaying  bool
	MPRISController *MPRISController
}

func NewAudioPlayer() *AudioPlayer {
	ap := &AudioPlayer{
		audioQueue: make([][]byte, 0),
	}

	mprisController := NewMPRISController(ap)
	if mprisController == nil {
		fmt.Println("Failed to initialize MPRIS controller")
		return nil
	}
	ap.MPRISController = mprisController

	return ap
}

func (ap *AudioPlayer) Play() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) == 0 && !ap.isAudioPlaying {
		return
	}

	if ap.isAudioPlaying {
		ap.audioController.Paused = false
		return
	}

	ap.isAudioPlaying = true
	go ap.playNextAudioChunk()
}

func (ap *AudioPlayer) playNextAudioChunk() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) == 0 {
		ap.isAudioPlaying = false
		return
	}

	audioData := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]

	audioReader := bytes.NewReader(audioData)
	audioReadCloser := io.NopCloser(audioReader)

	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		fmt.Printf("Failed to decode WAV data: %v\n", err)
		return
	}
	defer audioStreamer.Close()

	if ap.audioFormat == (beep.Format{}) {
		ap.audioFormat = format
		if err := speaker.Init(ap.audioFormat.SampleRate, ap.audioFormat.SampleRate.N(time.Second/10)); err != nil {
			fmt.Printf("Failed to initialize speaker: %v\n", err)
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
		ap.playNextAudioChunk()
	}()
}

func (ap *AudioPlayer) Pause() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		ap.audioController.Paused = true
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
		ap.isAudioPlaying = false
		ap.audioQueue = nil
	}
}

func (ap *AudioPlayer) AddToQueue(audioData []byte) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	ap.audioQueue = append(ap.audioQueue, audioData)
	if !ap.isAudioPlaying {
		ap.isAudioPlaying = true
		go ap.playNextAudioChunk()
	}
}
