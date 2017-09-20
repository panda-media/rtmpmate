package muxer

import (
	"bytes"
	"fmt"
	AACTypes "rtmpmate.com/codec/AAC/Types"
	"rtmpmate.com/codec/AudioFormats"
	H264Types "rtmpmate.com/codec/H264/Types"
	"rtmpmate.com/codec/VideoCodecs"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/NetStatusEvent"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type Muxer struct {
	Type               string
	DataFrames         map[string]*AMF.AMFObject
	InitAudio          *AudioMessage.AudioMessage
	InitVideo          *VideoMessage.VideoMessage
	Data               bytes.Buffer
	LastAudioTimestamp uint32
	LastVideoTimestamp uint32
	src                Interfaces.IStream
	endOfStream        bool

	events.EventDispatcher
}

func New() (*Muxer, error) {
	var m Muxer

	err := m.Init("Muxer")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *Muxer) Init(t string) error {
	this.Type = t
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

	meta := this.src.GetDataFrame("onMetaData")
	if meta != nil {
		this.DataFrames["onMetaData"] = meta
		this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.SET_DATA_FRAME, this, "onMetaData", meta))
	}

	return nil
}

func (this *Muxer) Unlink(src Interfaces.IStream) error {
	src.RemoveEventListener(DataFrameEvent.SET_DATA_FRAME, this.onSetDataFrame)
	src.RemoveEventListener(DataFrameEvent.CLEAR_DATA_FRAME, this.onClearDataFrame)
	src.RemoveEventListener(AudioEvent.DATA, this.onAudio)
	src.RemoveEventListener(VideoEvent.DATA, this.onVideo)
	this.src = nil

	return nil
}

func (this *Muxer) IsTypeSupported(mime string) bool {
	return true
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
	fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))

	this.DataFrames[e.Key] = e.Data
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.SET_DATA_FRAME, this, e.Key, e.Data))
}

func (this *Muxer) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {
	delete(this.DataFrames, e.Key)
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.CLEAR_DATA_FRAME, this, e.Key, e.Data))
}

func (this *Muxer) onAudio(e *AudioEvent.AudioEvent) {
	if this.InitAudio == nil {
		if e.Message.Format == AudioFormats.AAC && e.Message.DataType == AACTypes.SPECIFIC_CONFIG {
			this.InitAudio = e.Message
		}
	}

	this.LastAudioTimestamp = e.Message.Timestamp
	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *Muxer) onVideo(e *VideoEvent.VideoEvent) {
	if this.InitVideo == nil {
		if e.Message.Codec == VideoCodecs.AVC && e.Message.DataType == H264Types.SEQUENCE_HEADER {
			this.InitVideo = e.Message
		}
	}

	this.LastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *Muxer) EndOfStream(explain string) {
	this.endOfStream = true
	this.DispatchEvent(NetStatusEvent.New(NetStatusEvent.NET_STATUS, this, map[string]interface{}{
		"level":       Level.STATUS,
		"code":        Code.NETSTREAM_PLAY_STOP,
		"description": "play stop",
	}))
}
