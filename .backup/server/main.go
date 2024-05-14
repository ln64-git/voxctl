package main

import (
	"fmt"

	"github.com/ln64-git/voxctl/server/internal/handler"
	"github.com/ln64-git/voxctl/server/internal/server"
	"github.com/ln64-git/voxctl/server/internal/speech"
)

func main() {
	speechService := speech.NewService()
	handler := handler.NewHandler(speechService)

	status := server.Start(handler)
	if status.Launched {
		fmt.Printf("Server started on port %d\n", status.Port)
	} else if status.Error != nil {
		fmt.Printf("Failed to start server: %v\n", status.Error)
	}

	// Keep the server running until interrupted
	select {}
}
