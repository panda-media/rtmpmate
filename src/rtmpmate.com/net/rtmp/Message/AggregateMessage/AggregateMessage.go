package AggregateMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/AggregateMessage/Body"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AggregateMessage struct {
	Message.Header
	Body []Body.Body
}

func New(version byte) (*AggregateMessage, error) {
	var msg AggregateMessage
	msg.Type = Types.AGGREGATE

	return &msg, nil
}
