package SharedObjectMessage

import (
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/SharedObjectMessage/Event"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
)

// Not sure about this.
type SharedObjectMessage struct {
	Message.Header
	Name    AMF.AMFString
	Version byte
	Flags   uint8
	Events  []Event.Event
}

func New(version byte) (*SharedObjectMessage, error) {
	var msg SharedObjectMessage

	if version == AMF.AMF0 {
		msg.Type = Types.DATA
	} else {
		msg.Type = Types.AMF3_DATA
	}

	return &msg, nil
}
