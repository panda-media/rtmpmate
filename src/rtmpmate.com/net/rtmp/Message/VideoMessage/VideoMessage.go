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
	var m VideoMessage
	m.Type = Types.VIDEO

	return &m, nil
}

func (this *VideoMessage) Parse(b []byte, offset int, size int) error {
	this.Length = size
	this.Payload = b[offset : offset+size]

	return nil
}
