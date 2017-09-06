package Stream

import (
	"container/list"
	"syscall"

	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/Stream/RecordModes"
	"rtmpmate.com/util/AMF"
)

type Stream struct {
	Name               string
	Chunks             list.List
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

//Send invokes a remote method on client-side NetStream object
func (stream *Stream) Send(handlerName string, args ...*AMF.AMFValue) error {
	return syscall.EINVAL
}

//Clear del a recorded Stream file from the server
func (stream *Stream) Clear(streamName string) (bool, error) {
	return false, syscall.EINVAL
}

func (this *Stream) Destroy() error {
	return syscall.EINVAL
}

//Flush can flushes a stream
func (stream *Stream) Flush() error {
	return syscall.EINVAL
}

//GetOnMetaData returns an object containing
//the named stream or video file
func (stream *Stream) GetOnMetaData() error {
	return syscall.EINVAL
}

//Length return the length of a recorded stream in seconds
func (stream *Stream) Length() error {
	return syscall.EINVAL
}

//SetBufferTime set the len of the msg queue
func (stream *Stream) SetBufferTime() error {
	return syscall.EINVAL
}

//SetVirtualPath set the virtual dir path for video stream playback
func (stream *Stream) SetVirtualPath() error {
	return syscall.EINVAL
}

//Size return the size of a recorded stream in bytes
func (stream *Stream) Size() error {
	return syscall.EINVAL
}
