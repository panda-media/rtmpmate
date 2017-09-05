package DataMessage

import (
	"container/list"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
)

type DataMessage struct {
	Message.Header
	Key  string
	Data *AMF.AMFObject
}

func New(encoding byte) (*DataMessage, error) {
	var m DataMessage

	if encoding == AMF.AMF0 {
		m.Type = Types.DATA
	} else {
		m.Type = Types.AMF3_DATA
	}

	return &m, nil
}

func (this *DataMessage) Parse(b []byte, offset int, size int) error {
	k, err := AMF.DecodeString(b, offset, size-offset)
	if err != nil {
		return err
	}

	offset += k.Cost
	this.Key = k.Data

	v, err := AMF.DecodeValue(b, offset, size-offset)
	if err != nil {
		return err
	}

	offset += v.Cost
	this.Data = &AMF.AMFObject{
		AMFHash: AMF.AMFHash{v.Hash},
		Cost:    v.Cost,
		Ended:   v.Ended,
	}

	if v.Type == AMFTypes.OBJECT || v.Type == AMFTypes.ECMA_ARRAY || v.Type == AMFTypes.STRICT_ARRAY {
		this.Data.Data = v.Data.(list.List)
	}

	return nil
}
