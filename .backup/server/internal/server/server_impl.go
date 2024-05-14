package server

type ServerStatus struct {
	Launched bool
	Port     int
	Error    error
}
