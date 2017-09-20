package DASHMuxer

import (
	"rtmpmate.com/muxer/FMP4Muxer"
)

type DASHMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New() (*DASHMuxer, error) {
	var m DASHMuxer

	err := m.Init("DASHMuxer")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *DASHMuxer) Init(t string) error {
	err := this.FMP4Muxer.Init(t)
	if err != nil {
		return err
	}

	return nil
}

func (this *DASHMuxer) GetMPD() ([]byte, error) {
	return this.Slicer.MPD.GetMPDXML()
}
