package FMP4Muxer

import (
	"fmt"
	"github.com/panda-media/muxer-fmp4/codec/H264"
	"github.com/panda-media/muxer-fmp4/dashSlicer"
	"io"
	"os"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/muxer"
	"strconv"
)

type FMP4Muxer struct {
	muxer.Muxer
	Slicer          *dashSlicer.DASHSlicer
	MaxBufferLength int
	MaxBufferTime   int
	LowLatency      bool
	Record          bool
}

func New() (*FMP4Muxer, error) {
	var m FMP4Muxer

	err := m.Init("FMP4Muxer")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *FMP4Muxer) Init(t string) error {
	err := this.Muxer.Init(t)
	if err != nil {
		return err
	}

	this.Slicer, err = dashSlicer.NEWSlicer(0, 8, 5, this)
	if err != nil {
		return err
	}

	this.Record = true

	return nil
}

func (this *FMP4Muxer) VideoHeaderGenerated(videoHeader []byte) {
	fmt.Printf("VideoHeaderGenerated\n")

	name := "../www/initV.m4s"
	this.Save(name, videoHeader)
}

func (this *FMP4Muxer) VideoSegmentGenerated(videoSegment []byte, timestamp int64, duration int) {
	fmt.Printf("VideoSegmentGenerated\n")

	name := strconv.Itoa(int(timestamp))
	this.Save(name, videoSegment)
}

func (this *FMP4Muxer) AudioHeaderGenerated(audioHeader []byte) {
	fmt.Printf("AudioHeaderGenerated\n")

	name := "../www/initA.m4s"
	this.Save(name, audioHeader)
}

func (this *FMP4Muxer) AudioSegmentGenerated(audioSegment []byte, timestamp int64, duration int) {
	fmt.Printf("AudioSegmentGenerated\n")

	name := strconv.Itoa(int(timestamp))
	this.Save(name, audioSegment)
}

func (this *FMP4Muxer) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {
	fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))

	this.DataFrames[e.Key] = e.Data
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.SET_DATA_FRAME, this, e.Key, e.Data))
}

func (this *FMP4Muxer) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {
	delete(this.DataFrames, e.Key)
	this.DispatchEvent(DataFrameEvent.New(DataFrameEvent.CLEAR_DATA_FRAME, this, e.Key, e.Data))
}

func (this *FMP4Muxer) onAudio(e *AudioEvent.AudioEvent) {
	this.Slicer.AddAACFrame(e.Message.Payload[2:])

	this.LastAudioTimestamp = e.Message.Timestamp
	this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, e.Message))
}

func (this *FMP4Muxer) onVideo(e *VideoEvent.VideoEvent) {
	if b := e.Message.Payload; b[0] == 0x17 && b[1] == 0 {
		avc, err := H264.DecodeAVC(b[5:])
		if err != nil {
			fmt.Printf("Failed to decode AVC: %v.\n", err)
			return
		}

		for e := avc.SPS.Front(); e != nil; e = e.Next() {
			nal := make([]byte, 3+len(e.Value.([]byte)))
			nal[0] = 0
			nal[1] = 0
			nal[2] = 1
			copy(nal[3:], e.Value.([]byte))
			this.Slicer.AddH264Nals(nal)
		}

		for e := avc.PPS.Front(); e != nil; e = e.Next() {
			nal := make([]byte, 3+len(e.Value.([]byte)))
			nal[0] = 0
			nal[1] = 0
			nal[2] = 1
			copy(nal[3:], e.Value.([]byte))
			this.Slicer.AddH264Nals(nal)
		}
	} else {
		for i := 5; i < len(b); /* void */ {
			size := int(b[i]) << 24
			size |= int(b[i+1]) << 16
			size |= int(b[i+2]) << 8
			size |= int(b[i+3]) << 0
			i += 4

			nal := make([]byte, 3+size)
			nal[0] = 0
			nal[1] = 0
			nal[2] = 1
			copy(nal[3:], b[i:i+size])
			i += size

			this.Slicer.AddH264Nals(nal)
		}
	}

	this.LastVideoTimestamp = e.Message.Timestamp
	this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, e.Message))
}

func (this *FMP4Muxer) Save(name string, data []byte) error {
	var (
		f   *os.File
		err error
	)

	if _, err = os.Stat(name); os.IsNotExist(err) {
		f, err = os.Create(name)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(f, string(data))

	return err
}

func (this *FMP4Muxer) EndOfStream(explain string) {
	this.Muxer.EndOfStream(explain)
}
