package DASHMuxer

import (
	"rtmpmate.com/muxer/FMP4Muxer"
)

const (
	MPD_FILENAME = "manifest"
)

type DASHMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New(dir string, name string) (*DASHMuxer, error) {
	var m DASHMuxer

	err := m.Init(dir, name, "DASHMuxer")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *DASHMuxer) Init(dir string, name string, t string) error {
	err := this.FMP4Muxer.Init(dir, name, t)
	if err != nil {
		return err
	}

	return nil
}

func (this *DASHMuxer) GetMPD() ([]byte, error) {
	return this.Slicer.MPD.GetMPDXML()
}
