package models

type AudioEntry struct {
	AudioData   []byte
	SegmentText string
	FullText    []string
	ChatQuery   string
}
