package ServerEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Interfaces"
)

const (
	CONNECT    = "connect"
	PUBLISH    = "publish"
	UNPUBLISH  = "unpublish"
	DISCONNECT = "disconnect"
)

type ServerEvent struct {
	Event.Event
	Client Interfaces.INetConnection
	Stream Interfaces.IStream
}

func New(Type string, Target interface{}, client Interfaces.INetConnection, stream Interfaces.IStream) *ServerEvent {
	return &ServerEvent{Event.Event{Type, Target}, client, stream}
}

func (this *ServerEvent) Clone() *ServerEvent {
	return &ServerEvent{Event.Event{this.Type, this}, this.Client, this.Stream}
}

func (this *ServerEvent) ToString() string {
	return fmt.Sprintf("[ServerEvent type=%s client=%v stream=%v]", this.Type, this.Client, this.Stream)
}
