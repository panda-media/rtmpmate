package FLVMuxer

import (
	"fmt"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/format/FLV"
	"rtmpmate.com/muxer"
	"rtmpmate.com/util/AMF"
)

type FLVMuxer struct {
	muxer.Muxer
	lastVideoTimestamp uint32
}

func New() (*FLVMuxer, error) {
	var m FLVMuxer
	m.Init()

	return &m, nil
}

func (this *FLVMuxer) Init() error {
	this.Muxer.Init()

	b, _ := FLV.GetFileHeader()
	this.Data.Write(b)

	return nil
}

func (this *FLVMuxer) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {
	fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))

	var encoder AMF.Encoder
	encoder.EncodeString(e.Key)
	encoder.EncodeECMAArray(e.Data)
	data, _ := encoder.Encode()

	b, _ := FLV.Format(0x12, encoder.Len(), 0, data)
	this.Data.Write(b)

	this.DataFrames[e.Key] = e.Data
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.SET_DATA_FRAME, this, e.Key, e.Data))
}

func (this *FLVMuxer) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {
	delete(this.DataFrames, e.Key)
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.CLEAR_DATA_FRAME, this, e.Key, e.Data))
}

func (this *FLVMuxer) onAudio(e *AudioEvent.AudioEvent) {
	b, _ := FLV.Format(0x08, e.Message.Length, int(e.Message.Timestamp), e.Message.Payload)
	this.Data.Write(b)

	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *FLVMuxer) onVideo(e *VideoEvent.VideoEvent) {
	b, _ := FLV.Format(0x09, e.Message.Length, int(e.Message.Timestamp), e.Message.Payload)
	this.Data.Write(b)

	this.lastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *FLVMuxer) EndOfStream(explain string) {
	b, _ := FLV.Format(0x09, 5, int(this.lastVideoTimestamp), []byte{
		0x17, 0x02, 0x00, 0x00, 0x00,
	})
	this.Data.Write(b)

	this.Muxer.EndOfStream(explain)
}
