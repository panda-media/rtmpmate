package SourceBuffer

import (
	"bytes"
	"rtmpmate.com/muxer/TimeRanges"
)

type SourceBuffer struct {
	MimeType string
	Buffered *TimeRanges.TimeRanges
	Data     bytes.Buffer
	Active   bool
}

func New(mime string) (*SourceBuffer, error) {
	trs, err := TimeRanges.New()
	if err != nil {
		return nil, err
	}

	var sb SourceBuffer
	sb.MimeType = mime
	sb.Buffered = trs
	sb.Active = true

	return &sb, nil
}

func (this *SourceBuffer) Abort() error {
	return nil
}

func (this *SourceBuffer) AppendBuffer(b []byte) error {
	this.Data.Write(b)
	return nil
}

func (this *SourceBuffer) Remove(start float64, end float64) error {
	return nil
}
