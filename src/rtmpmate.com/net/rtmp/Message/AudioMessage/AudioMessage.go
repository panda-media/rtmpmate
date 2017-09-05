package AudioMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type AudioMessage struct {
	Message.Header
	Payload []byte
}

func New() (*AudioMessage, error) {
	var m AudioMessage
	m.Type = Types.AUDIO

	return &m, nil
}

func (this *AudioMessage) Parse(b []byte, offset int, size int) error {
	this.Length = size
	this.Payload = b[offset : offset+size]

	return nil
}
