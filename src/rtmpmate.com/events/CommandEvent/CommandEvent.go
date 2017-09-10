package CommandEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
	"rtmpmate.com/util/AMF"
)

const (
	CONNECT       = "connect"
	CLOSE         = "close"
	CREATE_STREAM = "createStream"
	RESULT        = "_result"
	ERROR         = "_error"

	PLAY          = "play"
	PLAY2         = "play2"
	DELETE_STREAM = "deleteStream"
	CLOSE_STREAM  = "closeStream"
	RECEIVE_AUDIO = "receiveAudio"
	RECEIVE_VIDEO = "receiveVideo"
	PUBLISH       = "publish"
	SEEK          = "seek"
	PAUSE         = "pause"
	ON_STATUS     = "onStatus"

	CHECK_BANDWIDTH = "checkBandwidth"
	GET_STATS       = "getStats"
)

type CommandEvent struct {
	Event.Event
	Message *CommandMessage.CommandMessage
	Encoder *AMF.Encoder
}

func New(Type string, Target interface{}, m *CommandMessage.CommandMessage, encoder *AMF.Encoder) *CommandEvent {
	return &CommandEvent{Event.Event{Type, Target}, m, encoder}
}

func (this *CommandEvent) Clone() *CommandEvent {
	return &CommandEvent{Event.Event{this.Type, this}, this.Message, this.Encoder}
}

func (this *CommandEvent) ToString() string {
	return fmt.Sprintf("[CommandEvent type=%s message=%v encoder=%v]", this.Type, this.Message, this.Encoder)
}
