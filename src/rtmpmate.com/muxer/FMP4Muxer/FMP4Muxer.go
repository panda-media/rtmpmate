package FMP4Muxer

import (
	"rtmpmate.com/muxer"
)

type FMP4Muxer struct {
	muxer.Muxer
}

func New() (*FMP4Muxer, error) {
	var m FMP4Muxer
	return &m, nil
}

func (this *FMP4Muxer) IsTypeSupported(mime string) bool {
	return true
}
