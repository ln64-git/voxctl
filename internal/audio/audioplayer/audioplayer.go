package audioplayer

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/ln64-git/voxctl/internal/models"
)

type ControlCommand int

const (
	ControlPause ControlCommand = iota
	ControlResume
	ControlStop
)

type AudioPlayer struct {
	mutex              sync.Mutex
	audioController    *beep.Ctrl
	doneChannel        chan struct{}
	audioFormat        beep.Format
	initialized        bool
	audioEntries       []models.AudioEntry
	audioEntriesUpdate chan []models.AudioEntry
	controlChannel     chan ControlCommand
	playing            bool
	activeChatQuery    string
}

func NewAudioPlayer(audioEntriesUpdate chan []models.AudioEntry) *AudioPlayer {
	return &AudioPlayer{
		doneChannel:        make(chan struct{}),
		audioEntriesUpdate: audioEntriesUpdate,
		controlChannel:     make(chan ControlCommand),
		playing:            false,
	}
}

func (ap *AudioPlayer) Start() {
	go func() {
		for {
			select {
			case newEntries := <-ap.audioEntriesUpdate:
				ap.mutex.Lock()
				ap.audioEntries = append(ap.audioEntries, newEntries...)
				ap.mutex.Unlock()
			default:
				ap.mutex.Lock()
				if !ap.playing && len(ap.audioEntries) > 0 {
					ap.playing = true
					ap.mutex.Unlock()
					ap.playNextAudioEntry()
				} else {
					ap.mutex.Unlock()
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

func (ap *AudioPlayer) handleControlCommand(cmd ControlCommand) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	switch cmd {
	case ControlPause:
		if ap.audioController != nil {
			ap.audioController.Paused = true
		}
	case ControlResume:
		if ap.audioController != nil {
			ap.audioController.Paused = false
		}
	case ControlStop:
		if ap.audioController != nil {
			ap.audioController = nil
			ap.playing = false
			ap.audioEntries = []models.AudioEntry{} // Clear all entries when stopped
		}
	}
}

func (ap *AudioPlayer) playNextAudioEntry() {
	ap.mutex.Lock()
	if len(ap.audioEntries) == 0 {
		ap.playing = false
		ap.mutex.Unlock()
		return
	}

	entry := ap.audioEntries[0]
	ap.audioEntries = ap.audioEntries[1:]
	ap.activeChatQuery = entry.ChatQuery
	ap.mutex.Unlock()

	log.Infof("playNextAudioEntry - %s -", entry.SegmentText)

	audioReader := bytes.NewReader(entry.AudioData)
	audioReadCloser := io.NopCloser(audioReader)
	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		log.Errorf("Error decoding audio data: %v", err)
		return
	}
	defer audioStreamer.Close()

	if !ap.initialized {
		ap.audioFormat = format
		err = speaker.Init(ap.audioFormat.SampleRate, ap.audioFormat.SampleRate.N(time.Second/10))
		if err != nil {
			log.Errorf("Error initializing speaker: %v", err)
			return
		}
		ap.initialized = true
	}

	ap.audioController = &beep.Ctrl{Streamer: audioStreamer, Paused: false}
	done := make(chan struct{})

	go func() {
		speaker.Play(beep.Seq(ap.audioController, beep.Callback(func() {
			close(done)
		})))
	}()

	for {
		select {
		case <-done:
			ap.mutex.Lock()
			if len(ap.audioEntries) > 0 {
				ap.mutex.Unlock()
				ap.playNextAudioEntry()
			} else {
				ap.playing = false
				ap.mutex.Unlock()
			}
			return
		case cmd := <-ap.controlChannel:
			ap.handleControlCommand(cmd)
		}
	}
}

func (ap *AudioPlayer) IsPlaying() bool {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	return ap.playing && ap.audioController != nil && !ap.audioController.Paused
}

func (ap *AudioPlayer) Pause() {
	ap.controlChannel <- ControlPause
}

func (ap *AudioPlayer) Resume() {
	ap.controlChannel <- ControlResume
}

func (ap *AudioPlayer) Stop() {
	ap.controlChannel <- ControlStop
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	if ap.audioController != nil {
		ap.audioController.Paused = true
		ap.audioController = nil
	}
	ap.audioEntries = []models.AudioEntry{}
	ap.playing = false
	ap.initialized = false
	ap.activeChatQuery = ""
	close(ap.doneChannel)
	ap.doneChannel = make(chan struct{})
}

func (ap *AudioPlayer) Clear() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		ap.audioController.Paused = true
		ap.audioController = nil
		ap.audioFormat = beep.Format{}
		ap.initialized = false
		ap.playing = false
	}

	close(ap.doneChannel)
	ap.doneChannel = make(chan struct{})
	ap.activeChatQuery = ""
}
