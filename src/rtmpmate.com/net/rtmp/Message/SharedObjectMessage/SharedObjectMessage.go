package SharedObjectMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
)

// Not sure about this.
type SharedObjectMessage struct {
	Message.Header
	Name    AMF.AMFString
	Version byte
	Flags   uint8
	Events  list.List
}

func New(encoding byte) (*SharedObjectMessage, error) {
	var m SharedObjectMessage

	if encoding == AMF.AMF0 {
		m.Type = Types.DATA
	} else {
		m.Type = Types.AMF3_DATA
	}

	return &m, nil
}
