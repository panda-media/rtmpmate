package Body

import (
	"rtmpmate.com/net/rtmp/Message"
)

type Body struct {
	Message.Header
	Data []byte
	Size uint
}
