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
		var audioEntries []models.AudioEntry
		for {
			select {
			case newEntries := <-ap.audioEntriesUpdate:
				log.Info("New Entry added")
				audioEntries = append(audioEntries, newEntries...)
			default:
				if len(audioEntries) > 0 {
					ap.playNextAudioEntry(audioEntries)
					audioEntries = audioEntries[1:]
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

func (ap *AudioPlayer) playNextAudioEntry(audioEntries []models.AudioEntry) {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if len(audioEntries) == 0 {
		ap.audioController.Paused = false
		return
	}

	entry := audioEntries[0]

	log.Info("Current Entry is - ")
	log.Info(entry.SegmentText)
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

	// Remove the first entry from the slice to advance to the next one
	audioEntries = audioEntries[1:]

	// Continue playing the next audio entry if there are more
	if len(audioEntries) > 0 {
		ap.playNextAudioEntry(audioEntries)
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
	ap.mutex.Lock()
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
