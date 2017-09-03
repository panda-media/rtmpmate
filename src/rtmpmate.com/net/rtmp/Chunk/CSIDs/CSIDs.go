package CSIDs

import ()

const (
	PROTOCOL_CONTROL = 0x02
	COMMAND          = 0x03
	COMMAND_2        = 0x04 // onStatus(NetStream.Play.Reset)
	STREAM           = 0x05
	VIDEO            = 0x06
	AUDIO            = 0x07
	AV               = 0x08
)
