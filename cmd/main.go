package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
	"github.com/sirupsen/logrus"
)

func main() {

	// Parse command-line flags
	flagPort := flag.Int("port", 8080, "Port number to connect or serve")
	flagToken := flag.String("token", "", "Process input stream token")
	flagInput := flag.String("input", "", "Input text to play")
	flagStatus := flag.Bool("status", false, "Request info")
	flagQuit := flag.Bool("quit", false, "Exit application after request")
	flagPause := flag.Bool("pause", false, "Pause audio playback")
	flagStop := flag.Bool("stop", false, "Stop audio playback")
	flag.Parse()

	// Retrieve configuration
	configName := "voxctl.json"
	configData, err := config.GetConfig(configName)
	if err != nil {
		logrus.Fatalf("failed to load configuration: %v", err)
	}

	// Populate state from configuration
	state := types.AppState{
		Port:                 *flagPort,
		Token:                *flagToken,
		Input:                *flagInput,
		StatusRequested:      *flagStatus,
		QuitRequested:        *flagQuit,
		PauseRequested:       *flagPause,
		StopRequested:        *flagStop,
		AzureSubscriptionKey: config.GetStringOrDefault(configData, "AzureSubscriptionKey", ""),
		AzureRegion:          config.GetStringOrDefault(configData, "AzureRegion", "eastus"),
		VoiceGender:          config.GetStringOrDefault(configData, "VoiceGender", "Female"),
		VoiceName:            config.GetStringOrDefault(configData, "VoiceName", "en-US-JennyNeural"),
		ServerAlreadyRunning: server.CheckServerRunning(*flagPort),
	}

	// Check if server is already running
	if !server.CheckServerRunning(state.Port) {
		state.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(state)
		time.Sleep(35 * time.Millisecond)
	} else {
		resp, err := server.ConnectToServer(state.Port)
		if err != nil {
			log.Errorf("Failed to connect to the existing server on port %d: %v", state.Port, err)
		} else {
			log.Infof("Connected to the existing server on port %d. Status: %s", state.Port, resp.Status)
			resp.Body.Close()
		}
	}

	processRequest(state)
	if state.QuitRequested {
		log.Info("Quit flag requested, Program Exiting")
		return
	}

	// Handle OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infof("Program Exiting")
}

func processRequest(state types.AppState) {
	client := &http.Client{}

	switch {
	case state.StatusRequested:
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.Port))
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.Input != "":
		// log.Info(state.Input)
		speechReq := speech.SpeechRequest{
			Text:      state.Input,
			Gender:    state.VoiceGender,
			VoiceName: state.VoiceName,
		}
		body := bytes.NewBufferString(speechReq.SpeechRequestToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", state.Port), "application/json", body)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.PauseRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/pause", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.StopRequested:
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/stop", state.Port), "", nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

	case state.Token != "":
		pipeReader, pipeWriter := io.Pipe()
		go func() {
			defer pipeWriter.Close()
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				var resp types.OllamaResponse
				err := json.Unmarshal([]byte(line), &resp)
				if err != nil {
					continue
				}
				if resp.Done {
					break
				}
				_, err = pipeWriter.Write([]byte(resp.Response))
				if err != nil {
					break
				}
			}
		}()

		tokenText, err := io.ReadAll(pipeReader)
		if err != nil {
			return
		}

		tokenReq := speech.SpeechRequest{
			Text:      string(tokenText),
			Gender:    state.VoiceGender,
			VoiceName: state.VoiceName,
		}
		body := bytes.NewBufferString(tokenReq.SpeechRequestToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/token", state.Port), "application/json", body)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}
}
