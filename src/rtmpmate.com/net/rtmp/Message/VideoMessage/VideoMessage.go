package VideoMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type VideoMessage struct {
	Message.Header
	Payload []byte
}

func New() (*VideoMessage, error) {
	var msg VideoMessage
	msg.Type = Types.VIDEO

	return &msg, nil
}
