package DASHMuxer

import (
	"rtmpmate.com/muxer/FMP4Muxer"
)

type DASHMuxer struct {
	FMP4Muxer.FMP4Muxer
}

func New() (*DASHMuxer, error) {
	var m DASHMuxer
	return &m, nil
}
