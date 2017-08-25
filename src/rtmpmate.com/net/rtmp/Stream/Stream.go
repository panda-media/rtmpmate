package Stream

import (
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/Stream/RecordModes"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type Stream struct {
	Name               string
	Live               bool
	Time               int
	BufferTime         int
	MaxQueueDelay      int // ms
	MaxQueueSize       int
	PublishQueryString string
	ReceiveAudio       bool
	ReceiveVideo       bool
}

func (this *Stream) Get(name string) (*Stream, error) {
	return nil, syscall.EINVAL
}

/*
	startTime:  -2: live -> vod -> create
	            -1: live -> create
	           >=0: vod  -> ignore
	             default -2
	length: -1: all
	         0: first video frame
	        >0: seconds
	         default -1
*/
func (this *Stream) Play(streamName string, startTime int, length int, reset bool, remoteConnection *NetConnection.NetConnection, virtualKey string) error {
	return syscall.EINVAL
}

func (this *Stream) Record(mode uint8, maxDuration int, maxSize int) error {
	switch mode {
	case RecordModes.STOP:

	case RecordModes.RECORD:

	case RecordModes.APPEND:

	default:
		return syscall.EINVAL
	}

	return nil
}

func (this *Stream) Send(handlerName string, args ...*AMF.AMFValue) error {
	return syscall.EINVAL
}

func (this *Stream) Clear() error {
	return syscall.EINVAL
}

func (this *Stream) Destroy() error {
	return syscall.EINVAL
}
