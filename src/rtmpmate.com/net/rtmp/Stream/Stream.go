package Stream

import (
	"container/list"
	"fmt"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/Stream/RecordModes"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type Stream struct {
	ID                 int
	Name               string
	Type               string
	Chunks             list.List
	DataFrames         map[string]*AMF.AMFObject
	Time               float64 // at when to init
	BufferTime         float64
	CurrentTime        float64
	Duration           float64
	MaxQueueDelay      float64 // ms
	MaxQueueSize       int
	Pause              bool
	PublishQueryString string
	ReceiveAudio       bool
	ReceiveVideo       bool

	events.EventDispatcher
}

func New(name string) (*Stream, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	var s Stream
	s.Name = name
	s.DataFrames = make(map[string]*AMF.AMFObject)

	return &s, nil
}

func (this *Stream) Source(src *Stream) error {
	if src == nil {
		return syscall.EINVAL
	}

	src.AddEventListener(DataFrameEvent.SET_DATA_FRAME, this.onSetDataFrame, 0)
	src.AddEventListener(DataFrameEvent.CLEAR_DATA_FRAME, this.onClearDataFrame, 0)
	src.AddEventListener(AudioEvent.DATA, this.onAudio, 0)
	src.AddEventListener(VideoEvent.DATA, this.onVideo, 0)

	return nil
}

func (this *Stream) Sink(to *Stream) error {
	if to == nil {
		return syscall.EINVAL
	}

	to.Source(this)

	return nil
}

func (this *Stream) Play(name string, start float64, length float64, reset bool) error {
	return nil
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
	return nil
}

func (this *Stream) Stop() error {
	return nil
}

func (this *Stream) Clear() error {
	this.Chunks.Init()
	return nil
}

func (this *Stream) Close() error {
	return nil
}

func (this *Stream) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {
	fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))

	this.DataFrames[e.Key] = e.Data
	this.DispatchEvent(DataFrameEvent.New(e.Type, this, e.Key, e.Data))
}

func (this *Stream) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {
	delete(this.DataFrames, e.Key)
	this.DispatchEvent(DataFrameEvent.New(e.Type, this, e.Key, e.Data))
}

func (this *Stream) onAudio(e *AudioEvent.AudioEvent) {
	this.DispatchEvent(AudioEvent.New(e.Type, this, e.Data))
}

func (this *Stream) onVideo(e *VideoEvent.VideoEvent) {
	this.DispatchEvent(VideoEvent.New(e.Type, this, e.Data))
}
