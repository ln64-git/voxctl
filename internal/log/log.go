package log

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() error {
	// Create the logs directory if it doesn't exist
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		return err
	}

	// Open the log file
	logFile, err := os.OpenFile("logs/server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Create the logger
	Logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	return nil
}
