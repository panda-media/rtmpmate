package ProgressEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
)

const (
	PROGRESS = "progress"
)

type ProgressEvent struct {
	Event.Event
	BytesLoaded int
	BytesTotal  int
}

func New(Type string, Target interface{}, BytesLoaded int, BytesTotal int) *ProgressEvent {
	return &ProgressEvent{Event.Event{Type, Target}, BytesLoaded, BytesTotal}
}

func (this *ProgressEvent) Clone() *ProgressEvent {
	return &ProgressEvent{Event.Event{this.Type, this}, this.BytesLoaded, this.BytesTotal}
}

func (this *ProgressEvent) ToString() string {
	return fmt.Sprintf("[ProgressEvent type=%s bytesLoaded=%d bytesTotal=%d]", this.Type, this.BytesLoaded, this.BytesTotal)
}
