package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ln64-git/voxctl/internal/handler"
	"github.com/ln64-git/voxctl/internal/speech"
)

func Start() error {
	speechService := speech.NewService()
	handler := handler.NewHandler(speechService)
	r := mux.NewRouter()

	r.HandleFunc("/play", handler.Play).Methods("POST")
	r.HandleFunc("/pause", handler.Pause).Methods("POST")
	r.HandleFunc("/resume", handler.Resume).Methods("POST")
	r.HandleFunc("/stop", handler.Stop).Methods("POST")

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
