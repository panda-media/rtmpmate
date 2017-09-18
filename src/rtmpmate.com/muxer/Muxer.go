package muxer

import (
	"bytes"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type IMuxer interface {
	Source(src Interfaces.IStream) error
	IsTypeSupported(mime string) bool
	EndOfStream(explain string)

	GetDataFrame(name string) *AMF.AMFObject
	GetInitAudio() *AudioMessage.AudioMessage
	GetInitVideo() *VideoMessage.VideoMessage

	Interfaces.IEventDispatcher
}

type Muxer struct {
	src         Interfaces.IStream
	DataFrames  map[string]*AMF.AMFObject
	InitAudio   *AudioMessage.AudioMessage
	InitVideo   *VideoMessage.VideoMessage
	Data        bytes.Buffer
	endOfStream bool

	events.EventDispatcher
}

func New() (*Muxer, error) {
	var m Muxer
	m.Init()

	return &m, nil
}

func (this *Muxer) Init() error {
	this.DataFrames = make(map[string]*AMF.AMFObject)

	return nil
}

func (this *Muxer) Source(src Interfaces.IStream) error {
	if src == nil {
		return syscall.EINVAL
	}

	this.src = src
	this.src.AddEventListener(DataFrameEvent.SET_DATA_FRAME, this.onSetDataFrame, 0)
	this.src.AddEventListener(DataFrameEvent.CLEAR_DATA_FRAME, this.onClearDataFrame, 0)
	this.src.AddEventListener(AudioEvent.DATA, this.onAudio, 0)
	this.src.AddEventListener(VideoEvent.DATA, this.onVideo, 0)

	return nil
}

func (this *Muxer) IsTypeSupported(mime string) bool {
	return true
}

func (this *Muxer) EndOfStream(explain string) {
	this.endOfStream = true
}

func (this *Muxer) GetDataFrame(name string) *AMF.AMFObject {
	data, _ := this.DataFrames[name]
	return data
}

func (this *Muxer) GetInitAudio() *AudioMessage.AudioMessage {
	return this.InitAudio
}

func (this *Muxer) GetInitVideo() *VideoMessage.VideoMessage {
	return this.InitVideo
}

func (this *Muxer) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {

}

func (this *Muxer) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {

}

func (this *Muxer) onAudio(e *AudioEvent.AudioEvent) {

}

func (this *Muxer) onVideo(e *VideoEvent.VideoEvent) {

}
