package UserControlMessage

import (
	"rtmpmate.com/net/rtmp/Message"
)

type UserControlMessage struct {
	Message.Header

	Type    uint16
	Payload []byte
}
