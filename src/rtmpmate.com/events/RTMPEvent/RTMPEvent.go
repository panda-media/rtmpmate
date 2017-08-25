package RTMPEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
)

const (
	NET_STATUS = "netStatus"
)

type RTMPEvent struct {
	Event.Event
	Info *map[string]interface{}
}

func New(Type string, Target interface{}, Info *map[string]interface{}) *RTMPEvent {
	return &RTMPEvent{Event.Event{Type, Target}, Info}
}

func (this *RTMPEvent) Clone() *RTMPEvent {
	return &RTMPEvent{Event.Event{this.Type, this}, this.Info}
}

func (this *RTMPEvent) ToString() string {
	return fmt.Sprintf("[RTMPEvent type=%s info=%v]", this.Type, this.Info)
}
