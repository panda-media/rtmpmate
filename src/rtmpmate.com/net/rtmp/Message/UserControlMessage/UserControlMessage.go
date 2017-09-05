package UserControlMessage

import (
	"encoding/binary"
	"fmt"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage/Event"
)

type UserControlMessage struct {
	Message.Header
	Event Event.Event
}

func New() (*UserControlMessage, error) {
	var m UserControlMessage
	m.Type = Types.USER_CONTROL

	return &m, nil
}

func (this *UserControlMessage) Parse(b []byte, offset int, size int) error {
	if size < 6 {
		return fmt.Errorf("data (size=%d) not enough", size)
	}

	this.Event.Type = binary.BigEndian.Uint16(b[offset : offset+2])
	offset += 2

	this.Event.Data = b[offset:]

	return nil
}
