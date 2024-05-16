package audio

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func PlayAudio(audioData []byte) error {
	// Create a bytes.Reader from the audio data
	audioReader := bytes.NewReader(audioData)

	// Create an io.ReadCloser from the bytes.Reader
	audioReadCloser := io.NopCloser(audioReader)

	// Decode the WAV data
	audioStreamer, format, err := wav.Decode(audioReadCloser)
	if err != nil {
		return fmt.Errorf("failed to decode WAV data: %v", err)
	}
	defer audioStreamer.Close()

	// Initialize the speaker
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("failed to initialize speaker: %v", err)
	}

	// Play the audio
	done := make(chan struct{})
	speaker.Play(beep.Seq(audioStreamer, beep.Callback(func() {
		close(done)
	})))

	// Wait for the audio to finish playing
	<-done

	return nil
}
