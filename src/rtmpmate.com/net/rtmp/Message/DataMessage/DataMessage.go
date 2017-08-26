package DataMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
)

type DataMessage struct {
	Message.Header
	Data []byte
}

func New(version byte) (*DataMessage, error) {
	var msg DataMessage

	if version == AMF.AMF0 {
		msg.Type = Types.DATA
	} else {
		msg.Type = Types.AMF3_DATA
	}

	return &msg, nil
}
