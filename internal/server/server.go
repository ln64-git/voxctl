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

type ServerStatus struct {
	Launched bool
	Port     int
	Error    error
}

func Start() ServerStatus {
	serverLock.Lock()
	defer serverLock.Unlock()

	if serverRunning {
		return ServerStatus{
			Launched: false,
			Port:     3000, // or whatever the port is
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

	// Start the HTTP server
	go func() {
		err := http.ListenAndServe(":3000", r)
		if err != nil {
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
