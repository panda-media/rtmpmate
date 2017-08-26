package UserControlMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage/Event"
)

type UserControlMessage struct {
	Message.Header
	Event Event.Event
}

func New() (*UserControlMessage, error) {
	var msg UserControlMessage
	msg.Type = Types.USER_CONTROL

	return &msg, nil
}
