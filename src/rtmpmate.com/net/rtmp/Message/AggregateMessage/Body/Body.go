package Body

import (
	"rtmpmate.com/net/rtmp/Message"
)

type Body struct {
	Message.Header
	Data []byte
	Size uint
}

func (this *Body) New() (*Body, error) {
	var b Body
	return &b, nil
}
