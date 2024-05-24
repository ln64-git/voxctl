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

	"github.com/ln64-git/voxctl/internal/audio"
	"github.com/ln64-git/voxctl/internal/config"
	"github.com/ln64-git/voxctl/internal/log"
	"github.com/ln64-git/voxctl/internal/server"
	"github.com/ln64-git/voxctl/internal/speech"
	"github.com/ln64-git/voxctl/internal/types"
)

func main() {
	err := log.InitLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer log.Logger.Writer()
	log.Logger.Println("main - Program Started")

	flagPort := flag.Int("port", 8080, "Port number to connect or serve")
	flagToken := flag.String("token", "", "Process input stream token")
	flagInput := flag.String("input", "", "Input text to play")
	flagStatus := flag.Bool("status", false, "Request info")
	flagQuit := flag.Bool("quit", false, "Exit application after request")
	flagPause := flag.Bool("pause", false, "Pause audio playback")
	flagStop := flag.Bool("stop", false, "Stop audio playback")
	flag.Parse()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Logger.Printf("Failed to get configuration: %v\n", err)
		return
	}

	serverAlreadyRunning := server.CheckServerRunning(*flagPort)

	state := types.AppState{
		Port:                 *flagPort,
		Token:                *flagToken,
		Input:                *flagInput,
		StatusRequested:      *flagStatus,
		QuitRequested:        *flagQuit,
		PauseRequested:       *flagPause,
		StopRequested:        *flagStop,
		AzureSubscriptionKey: cfg.AzureSubscriptionKey,
		AzureRegion:          cfg.AzureRegion,
		VoiceGender:          cfg.VoiceGender,
		VoiceName:            cfg.VoiceName,
		ServerAlreadyRunning: serverAlreadyRunning,
	}

	if !serverAlreadyRunning {
		state.AudioPlayer = audio.NewAudioPlayer()
		go server.StartServer(state)
	} else {
		log.Logger.Printf("Server is already running on port %d. Connecting to the existing server...\n", state.Port)
		server.ConnectToServer(state.Port)
	}

	processRequest(state)
	if state.QuitRequested {
		log.Logger.Println("Quit flag requested, Program Exiting")
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Println("Program Exiting")
}

func processRequest(state types.AppState) {
	client := &http.Client{}

	switch {
	case state.StatusRequested:
		log.Logger.Println("Status requested.")
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/status", state.Port))
		if err != nil {
			log.Logger.Printf("Failed to get status: %v\n", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		log.Logger.Printf("Status response: %s\n", string(body))

	case state.Input != "":
		log.Logger.Println("Input requested.")
		playReq := speech.SpeechRequest{
			Text:      state.Input,
			Gender:    state.VoiceGender,
			VoiceName: state.VoiceName,
		}
		body := bytes.NewBufferString(playReq.ToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/input", state.Port), "application/json", body)
		if err != nil {
			log.Logger.Printf("Failed to send input request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		log.Logger.Printf("Input response: %s\n", resp.Status)

	case state.PauseRequested:
		log.Logger.Println("Pause requested.")
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/pause", state.Port), "", nil)
		if err != nil {
			log.Logger.Printf("Failed to send pause request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		log.Logger.Printf("Pause response: %s\n", resp.Status)

	case state.StopRequested:
		log.Logger.Println("Stop requested.")
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/stop", state.Port), "", nil)
		if err != nil {
			log.Logger.Printf("Failed to send stop request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		log.Logger.Printf("Stop response: %s\n", resp.Status)

	case state.Token != "":
		log.Logger.Println("Token input detected.")
		pipeReader, pipeWriter := io.Pipe()
		go func() {
			defer pipeWriter.Close()
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				var resp types.OllamaResponse
				err := json.Unmarshal([]byte(line), &resp)
				if err != nil {
					log.Logger.Printf("Failed to unmarshal ollama response: %v\n", err)
					continue
				}
				if resp.Done {
					break
				}
				_, err = pipeWriter.Write([]byte(resp.Response))
				if err != nil {
					log.Logger.Printf("Failed to write to pipe: %v\n", err)
					break
				}
			}
		}()

		tokenText, err := io.ReadAll(pipeReader)
		if err != nil {
			log.Logger.Printf("Failed to read piped input: %v\n", err)
			return
		}

		tokenReq := speech.SpeechRequest{
			Text:      string(tokenText),
			Gender:    state.VoiceGender,
			VoiceName: state.VoiceName,
		}
		body := bytes.NewBufferString(tokenReq.ToJSON())
		resp, err := client.Post(fmt.Sprintf("http://localhost:%d/token", state.Port), "application/json", body)
		if err != nil {
			log.Logger.Printf("Failed to send token request: %v\n", err)
			return
		}
		defer resp.Body.Close()
		log.Logger.Printf("Token response: %s\n", resp.Status)
	}
}
