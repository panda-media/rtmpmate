package HLSMuxer

import (
	"rtmpmate.com/muxer/FMP4Muxer"
	"syscall"
)

type HLSMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New() (*HLSMuxer, error) {
	var m HLSMuxer

	err := m.Init("HLSMuxer")
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (this *HLSMuxer) Init(t string) error {
	err := this.FMP4Muxer.Init(t)
	if err != nil {
		return err
	}

	return nil
}

func (this *HLSMuxer) GetM3U8() ([]byte, error) {
	return nil, syscall.EINVAL
}
