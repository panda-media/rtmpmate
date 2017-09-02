package CommandMessage

import (
	"container/list"
	"fmt"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
)

type CommandMessage struct {
	Message.Header
	Name          AMF.AMFString
	TransactionID float64

	CommandObject  *AMF.AMFObject
	Arguments      *AMF.AMFObject
	StreamName     *AMF.AMFString
	Start          float64
	Duration       float64
	Reset          bool
	Parameters     *AMF.AMFObject
	StreamID       float64
	Flag           bool
	PublishingName *AMF.AMFString
	PublishingType *AMF.AMFString
	MilliSeconds   float64
	Pause          bool
}

func New(version byte) (*CommandMessage, error) {
	var msg CommandMessage

	if version == AMF.AMF0 {
		msg.Type = Types.COMMAND
	} else {
		msg.Type = Types.AMF3_COMMAND
	}

	return &msg, nil
}

func (this *CommandMessage) Parse(b []byte, offset int, size int) error {
	v, err := AMF.DecodeValue(b, offset, size-offset)
	if err != nil {
		return err
	}

	offset += v.Cost
	this.Name.Data = v.Data.(string)
	this.Name.Cost = v.Cost

	v, err = AMF.DecodeValue(b, offset, size-offset)
	if err != nil {
		return err
	}

	offset += v.Cost
	this.TransactionID = v.Data.(float64)

	switch this.Name.Data {
	// NetConnection Commands
	case Commands.CONNECT:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		if v.Type == AMFTypes.OBJECT {
			this.CommandObject = &AMF.AMFObject{AMF.AMFHash{v.Hash}, v.Data.(list.List), v.Cost, v.Ended}
		}

		v, _ = AMF.DecodeValue(b, offset, size-offset)
		if v != nil {
			offset += v.Cost
			this.Arguments = &AMF.AMFObject{AMF.AMFHash{v.Hash}, v.Data.(list.List), v.Cost, v.Ended}
		}

	case Commands.CLOSE:
		return fmt.Errorf("peer sends a close command")

	case Commands.CREATE_STREAM:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		if v.Type == AMFTypes.OBJECT {
			this.CommandObject = &AMF.AMFObject{AMF.AMFHash{v.Hash}, v.Data.(list.List), v.Cost, v.Ended}
		}

	// NetStream Commands
	case Commands.PLAY:
	case Commands.PLAY2:
	case Commands.DELETE_STREAM:
	case Commands.CLOSE_STREAM:
	case Commands.RECEIVE_AUDIO:
	case Commands.RECEIVE_VIDEO:
	case Commands.PUBLISH:
	case Commands.SEEK:
	case Commands.PAUSE:
	default:
	}

	return nil
}
