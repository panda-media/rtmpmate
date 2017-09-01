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
	Loaded   uint32
	State    byte
}

type BasicHeader struct {
	Fmt  byte   // 2 bits
	CSID uint32 // 6 | 14 | 22 bits
}

type MessageHeader struct {
	Timestamp       uint32
	MessageLength   uint32 // 3 bytes
	MessageTypeID   byte
	MessageStreamID uint32
}

func New(extended bool) (*Chunk, error) {
	var chunk Chunk
	chunk.Extended = extended
	chunk.State = States.START

	return &chunk, nil
}
