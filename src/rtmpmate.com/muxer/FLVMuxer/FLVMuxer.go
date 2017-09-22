package FLVMuxer

import (
	"fmt"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/format/FLV"
	"rtmpmate.com/muxer"
	MuxerTypes "rtmpmate.com/muxer/Types"
	"rtmpmate.com/util/AMF"
)

type FLVMuxer struct {
	muxer.Muxer
	Record bool
}

func New(dir string, name string) (*FLVMuxer, error) {
	var m FLVMuxer

	err := m.Init(dir, name, MuxerTypes.FLV)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *FLVMuxer) Init(dir string, name string, t string) error {
	err := this.Muxer.Init(dir, name, t)
	if err != nil {
		return err
	}

	b, _ := FLV.GetFileHeader()
	this.Data.Write(b)

	return nil
}

func (this *FLVMuxer) onSetDataFrame(e *DataEvent.DataEvent) {
	fmt.Printf("FLVMuxer.%s: %s\n", e.Message.Key, e.Message.Data.ToString(0))

	if this.Record {
		var encoder AMF.Encoder
		encoder.EncodeString(e.Message.Key)
		encoder.EncodeECMAArray(e.Message.Data)
		data, _ := encoder.Encode()

		b, _ := FLV.Format(0x12, encoder.Len(), 0, data)
		this.Data.Write(b)
	}

	this.DataFrames[e.Message.Key] = e.Message
	this.DispatchEvent(DataEvent.New(DataEvent.SET_DATA_FRAME, this, e.Message))
}

func (this *FLVMuxer) onClearDataFrame(e *DataEvent.DataEvent) {
	delete(this.DataFrames, e.Message.Key)
	this.DispatchEvent(DataEvent.New(DataEvent.CLEAR_DATA_FRAME, this, e.Message))
}

func (this *FLVMuxer) onAudio(e *AudioEvent.AudioEvent) {
	if this.Record {
		b, _ := FLV.Format(0x08, e.Message.Length, int(e.Message.Timestamp), e.Message.Payload)
		this.Data.Write(b)
	}

	this.LastAudioTimestamp = e.Message.Timestamp
	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *FLVMuxer) onVideo(e *VideoEvent.VideoEvent) {
	if this.Record {
		b, _ := FLV.Format(0x09, e.Message.Length, int(e.Message.Timestamp), e.Message.Payload)
		this.Data.Write(b)
	}

	this.LastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *FLVMuxer) EndOfStream(explain string) {
	if this.Record {
		b, _ := FLV.Format(0x09, 5, int(this.LastVideoTimestamp), []byte{
			0x17, 0x02, 0x00, 0x00, 0x00,
		})
		this.Data.Write(b)
	}

	this.Muxer.EndOfStream(explain)
}
