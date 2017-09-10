package AudioEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
)

const (
	DATA = "audiodata"
)

type AudioEvent struct {
	Event.Event
	Data *AudioMessage.AudioMessage
}

func New(Type string, Target interface{}, data *AudioMessage.AudioMessage) *AudioEvent {
	return &AudioEvent{Event.Event{Type, Target}, data}
}

func (this *AudioEvent) Clone() *AudioEvent {
	return &AudioEvent{Event.Event{this.Type, this}, this.Data}
}

func (this *AudioEvent) ToString() string {
	return fmt.Sprintf("[AudioEvent type=%s data=%v]", this.Type, this.Data)
}
