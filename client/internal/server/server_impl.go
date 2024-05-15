package server

// ServerStatus represents the status of the server
type ServerStatus struct {
	Launched bool
	Port     int
	Error    error
}

// NewServerStatus creates a new ServerStatus with the given port
func NewServerStatus(port int) ServerStatus {
	return ServerStatus{
		Launched: false,
		Port:     port,
		Error:    nil,
	}
}

// WithError sets the error in the ServerStatus
func (s ServerStatus) WithError(err error) ServerStatus {
	return ServerStatus{
		Launched: s.Launched,
		Port:     s.Port,
		Error:    err,
	}
}

// WithLaunched sets the Launched field in the ServerStatus
func (s ServerStatus) WithLaunched(launched bool) ServerStatus {
	return ServerStatus{
		Launched: launched,
		Port:     s.Port,
		Error:    s.Error,
	}
}
