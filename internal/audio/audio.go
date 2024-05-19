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
	isPlaying  bool
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

	if !ap.isPlaying {
		ap.isPlaying = true
		go ap.playNext()
	}
}

func (ap *AudioPlayer) playNext() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) == 0 {
		ap.isPlaying = false
		return
	}

	audioData := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]

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
		err = speaker.Init(ap.format.SampleRate, ap.format.SampleRate.N(time.Second/10))
		if err != nil {
			fmt.Printf("Failed to initialize speaker: %v\n", err)
			ap.playNextIfAvailable()
			return
		}
	}

	ap.ctrl = &beep.Ctrl{Streamer: audioStreamer, Paused: false}

	var wg sync.WaitGroup
	wg.Add(1)
	speaker.Play(beep.Seq(ap.ctrl, beep.Callback(func() {
		wg.Done()
	})))

	go func() {
		wg.Wait()
		ap.playNextIfAvailable()
	}()
}

func (ap *AudioPlayer) playNextIfAvailable() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioQueue) > 0 {
		ap.isPlaying = true
		go ap.playNext()
	} else {
		ap.isPlaying = false
	}
}

func (ap *AudioPlayer) Pause() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.ctrl != nil {
		ap.ctrl.Paused = true
	}
}

func (ap *AudioPlayer) Resume() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.ctrl != nil {
		ap.ctrl.Paused = false
	}
}

func (ap *AudioPlayer) Stop() {
	speaker.Lock()
	defer speaker.Unlock()

	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.ctrl != nil {
		if closer, ok := ap.ctrl.Streamer.(io.Closer); ok {
			closer.Close()
		}
		ap.done <- struct{}{}
		ap.isPlaying = false
		ap.audioQueue = nil
	}
}
