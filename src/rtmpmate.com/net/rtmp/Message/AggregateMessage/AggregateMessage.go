package AggregateMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AggregateMessage struct {
	Message.Header
	Body list.List
}

func New(version byte) (*AggregateMessage, error) {
	var m AggregateMessage
	m.Type = Types.AGGREGATE

	return &m, nil
}
