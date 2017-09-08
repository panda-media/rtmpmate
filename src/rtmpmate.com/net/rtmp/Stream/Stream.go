package Stream

import (
	"container/list"
	"rtmpmate.com/net/rtmp/Stream/RecordModes"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type Stream struct {
	ID                 int
	Name               string
	Type               string
	Chunks             list.List
	Time               float64 // at when to init
	BufferTime         float64
	CurrentTime        float64
	Duration           float64
	MaxQueueDelay      float64 // ms
	MaxQueueSize       int
	Pause              bool
	PublishQueryString string
}

func New(id int, name string, t string) (*Stream, error) {
	if id < 0 {
		return nil, syscall.EINVAL
	}

	var s Stream
	s.ID = id
	s.Name = name
	s.Type = t

	return &s, nil
}

func (this *Stream) Play(name string, start float64, length float64, reset bool) error {
	return syscall.EINVAL
}

func (this *Stream) Record(mode string, maxDuration int, maxSize int) error {
	switch mode {
	case RecordModes.RECORD:
	case RecordModes.APPEND:
	case RecordModes.STOP:
	default:
		return syscall.EINVAL
	}

	return nil
}

func (this *Stream) Send(handler string, args ...*AMF.AMFValue) error {
	return syscall.EINVAL
}

func (this *Stream) Stop() error {
	return syscall.EINVAL
}

func (this *Stream) Clear() error {
	return syscall.EINVAL
}
