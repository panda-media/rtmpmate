package Message

import ()

type Message struct {
	Header
	Payload []byte
}

type Header struct {
	Type      byte
	Length    [3]byte
	Timestamp uint
	StreamID  [3]byte
}
