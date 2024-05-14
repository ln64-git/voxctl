package server

import (
	"context"
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
	server        *http.Server
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

	// Create a new HTTP server
	server = &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	// Start the HTTP server
	go func() {
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

	serverRunning = false
	return nil
}
