package FMP4Muxer

import (
	"rtmpmate.com/muxer"
)

type FMP4Muxer struct {
	muxer.Muxer
}

func New() (*FMP4Muxer, error) {
	var m FMP4Muxer
	m.Init()

	return &m, nil
}

func (this *FMP4Muxer) Init() error {
	this.Muxer.Init()

	return nil
}
