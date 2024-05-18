package types

import (
	"sync"

	"github.com/ln64-git/voxctl/internal/audio"
)

type State struct {
	AudioPlayer *audio.AudioPlayer
	Status      string
	mu          sync.Mutex
}

func (s *State) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

func (s *State) GetStatus() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Status
}
