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

type AudioPlayer struct {
	mutex              sync.Mutex
	audioController    *beep.Ctrl
	doneChannel        chan struct{}
	audioFormat        beep.Format
	initialized        bool
	audioEntries       []models.AudioEntry
	audioEntriesUpdate chan []models.AudioEntry
}

func NewAudioPlayer(audioEntriesUpdate chan []models.AudioEntry) *AudioPlayer {
	return &AudioPlayer{
		doneChannel:        make(chan struct{}),
		audioEntriesUpdate: audioEntriesUpdate,
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
				go ap.playNextAudioEntry()
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (ap *AudioPlayer) playNextAudioEntry() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(ap.audioEntries) == 0 {
		return
	}

	entry := ap.audioEntries[0]
	ap.audioEntries = ap.audioEntries[1:]

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
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	speaker.Play(beep.Seq(ap.audioController, beep.Callback(func() {
		waitGroup.Done()
	})))

	// Wait for the audio to finish playing
	waitGroup.Wait()

	// Continue playing the next audio entry if there are more
	if len(ap.audioEntries) > 0 {
		go ap.playNextAudioEntry()
	} else {
		// No more audio entries, mark as not playing
		ap.audioController.Paused = false
	}
}

func (ap *AudioPlayer) WaitForCompletion() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController.Paused {
		return
	}

	<-ap.doneChannel
}

func (ap *AudioPlayer) IsPlaying() bool {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	return !ap.audioController.Paused
}

func (ap *AudioPlayer) Pause() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		log.Info("AudioPlayer - Pause - Paused")
		ap.audioController.Paused = true
	}
}

func (ap *AudioPlayer) Resume() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		log.Info("AudioPlayer - Resume - Resumed")
		ap.audioController.Paused = false
	}
}

func (ap *AudioPlayer) Stop() {
	ap.mutex.Lock()
	log.Info("AudioPlayer - Stop - Stopped")
	defer ap.mutex.Unlock()
}

func (ap *AudioPlayer) Clear() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if ap.audioController != nil {
		ap.audioController.Paused = true
		ap.audioController = nil
		ap.audioFormat = beep.Format{}
		ap.initialized = false
	}

	close(ap.doneChannel)
	ap.doneChannel = make(chan struct{})
}
