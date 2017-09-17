package muxer

import (
	"fmt"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/ErrorEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/muxer/SourceBuffer"
	AACTypes "rtmpmate.com/muxer/codec/AAC/Types"
	"rtmpmate.com/muxer/codec/AudioFormats"
	H264Types "rtmpmate.com/muxer/codec/H264/Types"
	"rtmpmate.com/muxer/codec/VideoCodecs"
	"rtmpmate.com/net/rtmp/Interfaces"
	"syscall"
)

type IMuxer interface {
	Source(src Interfaces.IStream) error
	AddSourceBuffer(mime string) (*SourceBuffer.SourceBuffer, error)
	RemoveSourceBuffer(sb *SourceBuffer.SourceBuffer) error
	IsTypeSupported(mime string) bool
	EndOfStream(explain string)

	Interfaces.IEventDispatcher
}

type Muxer struct {
	src           Interfaces.IStream
	SourceBuffers map[string]*SourceBuffer.SourceBuffer
	Duration      float64
	endOfStream   bool

	events.EventDispatcher
}

func New() (*Muxer, error) {
	var m Muxer
	return &m, nil
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

func (this *Muxer) AddSourceBuffer(mime string) (*SourceBuffer.SourceBuffer, error) {
	sb, err := SourceBuffer.New(mime)
	if err != nil {
		return nil, err
	}

	this.SourceBuffers[mime] = sb

	return sb, nil
}

func (this *Muxer) RemoveSourceBuffer(sb *SourceBuffer.SourceBuffer) error {
	delete(this.SourceBuffers, sb.MimeType)
	return nil
}

func (this *Muxer) IsTypeSupported(mime string) bool {
	return true
}

func (this *Muxer) EndOfStream(explain string) {
	this.endOfStream = true
}

func (this *Muxer) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {

}

func (this *Muxer) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {

}

func (this *Muxer) onAudio(e *AudioEvent.AudioEvent) {
	var sb *SourceBuffer.SourceBuffer
	var err error
	var ok bool

	if e.Message.Format == AudioFormats.AAC {
		if e.Message.DataType == AACTypes.SPECIFIC_CONFIG {
			sb, err = this.AddSourceBuffer("audio/mp4; codecs=\"mp4a.40.2\"")
			if err != nil {
				this.DispatchEvent(ErrorEvent.New(ErrorEvent.ERROR, this,
					fmt.Errorf("Failed to AddSourceBuffer")))
				return
			}
		} else {
			sb, ok = this.SourceBuffers["audio/mp4; codecs=\"mp4a.40.2\""]
			if ok == false {
				this.DispatchEvent(ErrorEvent.New(ErrorEvent.ERROR, this,
					fmt.Errorf("SourceBuffer not found")))
				return
			}
		}

		sb.AppendBuffer(e.Message.Payload)
	}
}

func (this *Muxer) onVideo(e *VideoEvent.VideoEvent) {
	var sb *SourceBuffer.SourceBuffer
	var err error
	var ok bool

	if e.Message.Codec == VideoCodecs.AVC {
		if e.Message.DataType == H264Types.SEQUENCE_HEADER {
			sb, err = this.AddSourceBuffer("video/mp4; codecs=\"avc1.640028\"")
			if err != nil {
				this.DispatchEvent(ErrorEvent.New(ErrorEvent.ERROR, this,
					fmt.Errorf("Failed to AddSourceBuffer")))
				return
			}
		} else {
			sb, ok = this.SourceBuffers["video/mp4; codecs=\"avc1.640028\""]
			if ok == false {
				this.DispatchEvent(ErrorEvent.New(ErrorEvent.ERROR, this,
					fmt.Errorf("SourceBuffer not found")))
				return
			}
		}

		sb.AppendBuffer(e.Message.Payload)
	}
}
