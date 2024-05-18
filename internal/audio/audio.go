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
	audioQueue [][]byte
	mutex      sync.Mutex
	ctrl       *beep.Ctrl
	done       chan struct{}
	format     beep.Format
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		audioQueue: make([][]byte, 0),
		done:       make(chan struct{}),
	}
}

func (ap *AudioPlayer) Play(audioData []byte) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	ap.audioQueue = append(ap.audioQueue, audioData)

	if len(ap.audioQueue) == 1 {
		go ap.playNext()
	}
}

func (ap *AudioPlayer) playNext() {
	ap.mutex.Lock()
	audioData := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]
	ap.mutex.Unlock()

	audioReader := bytes.NewReader(audioData)
	audioReadCloser := io.NopCloser(audioReader)

	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		fmt.Printf("Failed to decode WAV data: %v\n", err)
		ap.playNextIfAvailable()
		return
	}
	defer audioStreamer.Close()

	if ap.format == (beep.Format{}) {
		ap.format = format
		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			fmt.Printf("Failed to initialize speaker: %v\n", err)
			ap.playNextIfAvailable()
			return
		}
	}

	ap.ctrl = &beep.Ctrl{Streamer: audioStreamer, Paused: false}
	speaker.Play(beep.Seq(ap.ctrl, beep.Callback(func() {
		ap.playNextIfAvailable()
	})))

	<-ap.done
}

func (ap *AudioPlayer) playNextIfAvailable() {
	ap.mutex.Lock()
	if len(ap.audioQueue) > 0 {
		go ap.playNext()
	}
	ap.mutex.Unlock()
}

func (ap *AudioPlayer) Pause() {
	if ap.ctrl != nil {
		ap.ctrl.Paused = true
	}
}

func (ap *AudioPlayer) Resume() {
	if ap.ctrl != nil {
		ap.ctrl.Paused = false
	}
}

func (ap *AudioPlayer) Stop() {
	speaker.Lock()
	if ap.ctrl != nil {
		if closer, ok := ap.ctrl.Streamer.(io.Closer); ok {
			closer.Close()
		}
		ap.done <- struct{}{}
	}
	speaker.Unlock()
}
