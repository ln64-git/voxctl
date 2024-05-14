package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/ln64-git/voxctl/internal/handler"
	"github.com/ln64-git/voxctl/internal/speech"
)

var (
	serverRunning bool
	serverLock    sync.Mutex
)

func Start() error {
	// Lock to ensure thread safety while checking and updating serverRunning status
	serverLock.Lock()
	defer serverLock.Unlock()

	// Check if the server is already running
	if serverRunning {
		return fmt.Errorf("server is already running")
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

	// Start the HTTP server
	go func() {
		err := http.ListenAndServe(":3000", r)
		if err != nil {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	// Update serverRunning status
	serverRunning = true

	return nil
}
