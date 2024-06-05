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
)

type AudioEntry struct {
	AudioData   []byte
	SegmentText string
	FullText    []string
	ChatQuery   string
}

type AudioPlayer struct {
	audioQueue      []AudioEntry
	mutex           sync.Mutex
	audioController *beep.Ctrl
	doneChannel     chan struct{}
	audioFormat     beep.Format
	isAudioPlaying  bool
	initialized     bool
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		audioQueue:  make([]AudioEntry, 0),
		doneChannel: make(chan struct{}),
	}
}

func (ap *AudioPlayer) Start() {
	go func() {
		for {
			ap.playNextAudioEntry()
		}
	}()
}

func (ap *AudioPlayer) PlayAudioEntries(entries []AudioEntry) {
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

	ap.audioQueue = append(ap.audioQueue, entries...)

	if !ap.isAudioPlaying {
		ap.isAudioPlaying = true
	}
}

func (ap *AudioPlayer) playNextAudioEntry() {
	// log.Info("playNextAudioEntry called")
	ap.mutex.Lock()

	if len(ap.audioQueue) == 0 {
		ap.isAudioPlaying = false
		ap.mutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		return
	}

	entry := ap.audioQueue[0]
	ap.audioQueue = ap.audioQueue[1:]

	ap.mutex.Unlock()

	audioReader := bytes.NewReader(entry.AudioData)
	audioReadCloser := io.NopCloser(audioReader)

	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		log.Errorf("Error decoding audio data: %v", err)
		ap.playNextAudioEntry()
		return
	}
	defer audioStreamer.Close()

	if !ap.initialized {
		ap.audioFormat = format
		err = speaker.Init(ap.audioFormat.SampleRate, ap.audioFormat.SampleRate.N(time.Second/10))
		if err != nil {
			log.Errorf("Error initializing speaker: %v", err)
			ap.playNextAudioEntry()
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

	waitGroup.Wait()
	ap.playNextAudioEntry()
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

	if ap.audioController != nil {
		speaker.Lock()
		ap.audioController.Paused = true
		speaker.Unlock()
	}

	ap.audioQueue = nil
	ap.isAudioPlaying = false
}

func (ap *AudioPlayer) WaitForCompletion() {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	if !ap.isAudioPlaying {
		return
	}

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

	select {
	case <-ap.doneChannel:
	default:
		close(ap.doneChannel)
	}

	ap.doneChannel = make(chan struct{})
}
