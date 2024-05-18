package types

import "sync"

type State struct {
	Status string
	mu     sync.Mutex
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
