package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ln64-git/voxctl/client/internal/speech"
	"github.com/ln64-git/voxctl/server/internal/handler"
)

var (
	serverRunning bool
	serverLock    sync.Mutex
	server        *http.Server
	wg            sync.WaitGroup
)

func Start() ServerStatus {
	serverLock.Lock()
	defer serverLock.Unlock()

	if serverRunning {
		return ServerStatus{
			Launched: false,
			Port:     3000,
			Error:    nil,
		}
	}

	// Initialize speech service and handler
	speechService := speech.NewService()
	handler := handler.NewHandler(speechService)

	// Create a new router
	r := mux.NewRouter()

	// Register handler functions
	r.HandleFunc("/play", handler.Play).Methods("POST")
	r.HandleFunc("/pause", handler.Pause).Methods("POST")
	r.HandleFunc("/resume", handler.Resume).Methods("POST")
	r.HandleFunc("/stop", handler.Stop).Methods("POST")

	// Create a new HTTP server
	server = &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	// Start the HTTP server in a new goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	// Update serverRunning status
	serverRunning = true

	return ServerStatus{
		Launched: true,
		Port:     3000,
		Error:    nil,
	}
}

func Stop() error {
	serverLock.Lock()
	defer serverLock.Unlock()

	if !serverRunning {
		return nil
	}

	err := server.Shutdown(context.Background())
	if err != nil {
		return err
	}

	// Wait for the server to stop
	wg.Wait()

	serverRunning = false
	return nil
}
