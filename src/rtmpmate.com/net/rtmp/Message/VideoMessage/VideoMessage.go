package VideoMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
)

type VideoMessage struct {
	Message.Header
	FrameType byte // 0xF0
	Codec     byte // 0x0F
	DataType  byte
	Payload   []byte
}

func New() (*VideoMessage, error) {
	var m VideoMessage
	m.Type = Types.VIDEO

	return &m, nil
}

func (this *VideoMessage) Parse(b []byte, offset int, size int) error {
	this.Length = size

	pos := 0
	tmp := b[offset+pos]
	this.FrameType = (tmp >> 4) & 0x0F
	this.Codec = tmp & 0x0F

	pos++
	this.DataType = b[offset+pos]

	this.Payload = b[offset : offset+size]

	return nil
}
