package Chunk

import (
	"bytes"
	"rtmpmate.com/net/rtmp/Chunk/States"
)

type Chunk struct {
	BasicHeader
	MessageHeader
	Extended bool
	Data     bytes.Buffer
	State    byte
}

type BasicHeader struct {
	Fmt  byte   // 2 bits
	CSID uint32 // 6 | 14 | 22 bits
}

type MessageHeader struct {
	Timestamp       uint32
	MessageLength   int // 3 bytes
	MessageTypeID   byte
	MessageStreamID uint32
}

func New(extended bool) (*Chunk, error) {
	var c Chunk
	c.Extended = extended
	c.State = States.START

	return &c, nil
}
