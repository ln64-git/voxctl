package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	serverRunning bool
	serverLock    sync.Mutex
)

// Start launches the Rust server in a new process
func Start(port int) ServerStatus {
	serverLock.Lock()
	defer serverLock.Unlock()

	if serverRunning {
		return NewServerStatus(port)
	}

	if port == 0 {
		port = 3000
	}

	// Get the absolute path to the server's Cargo.toml file
	serverDir, err := os.Getwd()
	if err != nil {
		return NewServerStatus(port).WithError(err)
	}
	cargoTomlPath := filepath.Join(serverDir, "/../../../server/Cargo.toml")

	cmd := exec.Command("cargo", "run", "-q", "--manifest-path", cargoTomlPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return NewServerStatus(port).WithError(err)
	}

	serverRunning = true
	return NewServerStatus(port).WithLaunched(true)
}

// Stop stops the Rust server
func Stop() error {
	serverLock.Lock()
	defer serverLock.Unlock()

	if !serverRunning {
		return nil
	}

	// TODO: Implement the logic to stop the server
	serverRunning = false
	return nil
}
