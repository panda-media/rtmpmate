package AudioMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AudioMessage struct {
	Message.Header
	Data []byte
}

func New() (*AudioMessage, error) {
	var msg AudioMessage
	msg.Type = Types.AUDIO

	return &msg, nil
}
