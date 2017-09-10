package VideoEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
)

const (
	DATA = "videodata"
)

type VideoEvent struct {
	Event.Event
	Data *VideoMessage.VideoMessage
}

func New(Type string, Target interface{}, data *VideoMessage.VideoMessage) *VideoEvent {
	return &VideoEvent{Event.Event{Type, Target}, data}
}

func (this *VideoEvent) Clone() *VideoEvent {
	return &VideoEvent{Event.Event{this.Type, this}, this.Data}
}

func (this *VideoEvent) ToString() string {
	return fmt.Sprintf("[VideoEvent type=%s data=%v]", this.Type, this.Data)
}
