package CommandMessage

import (
	"container/list"
	"fmt"
	"math"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
)

type CommandMessage struct {
	Message.Header
	Name          AMF.AMFString
	TransactionID uint64

	CommandObject  *AMF.AMFObject
	Arguments      *AMF.AMFObject
	StreamName     *AMF.AMFString
	Start          float64
	Duration       float64
	Reset          bool
	Parameters     *AMF.AMFObject
	StreamID       uint64
	Flag           bool
	PublishingName *AMF.AMFString
	PublishingType *AMF.AMFString
	MilliSeconds   float64
	Pause          bool
}

func New(encoding byte) (*CommandMessage, error) {
	var m CommandMessage

	if encoding == AMF.AMF0 {
		m.Type = Types.COMMAND
	} else {
		m.Type = Types.AMF3_COMMAND
	}

	return &m, nil
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
	this.TransactionID = math.Float64bits(v.Data.(float64))

	switch this.Name.Data {
	// NetConnection Commands
	case Commands.CONNECT:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Data:    v.Data.(list.List),
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

		v, _ = AMF.DecodeValue(b, offset, size-offset)
		if v != nil {
			offset += v.Cost
			this.Arguments = &AMF.AMFObject{
				AMFHash: AMF.AMFHash{v.Hash},
				Data:    v.Data.(list.List),
				Cost:    v.Cost,
				Ended:   v.Ended,
			}
		}

	case Commands.CLOSE:
		fmt.Printf("Parsing command %s.\n", this.Name.Data)

	case Commands.CREATE_STREAM:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

		if v.Type == AMFTypes.OBJECT {
			this.CommandObject.Data = v.Data.(list.List)
		}

	// NetStream Commands
	case Commands.PLAY:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.StreamName = &AMF.AMFString{
			Data: v.Data.(string),
			Cost: v.Cost,
		}

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Start = v.Data.(float64)

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Duration = v.Data.(float64)

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Reset = v.Data.(bool)

	case Commands.PLAY2:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Parameters = &AMF.AMFObject{
			AMFHash: AMF.AMFHash{v.Hash},
			Data:    v.Data.(list.List),
			Cost:    v.Cost,
			Ended:   v.Ended,
		}

	case Commands.DELETE_STREAM:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.StreamID = math.Float64bits(v.Data.(float64))

	case Commands.CLOSE_STREAM:
		fmt.Printf("Parsing command %s.\n", this.Name.Data)

	case Commands.RECEIVE_AUDIO:
		fallthrough
	case Commands.RECEIVE_VIDEO:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Flag = v.Data.(bool)

	case Commands.PUBLISH:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.PublishingName = &AMF.AMFString{
			Data: v.Data.(string),
			Cost: v.Cost,
		}

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.PublishingType = &AMF.AMFString{
			Data: v.Data.(string),
			Cost: v.Cost,
		}

	case Commands.SEEK:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.MilliSeconds = v.Data.(float64)

	case Commands.PAUSE:
		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.CommandObject = nil // v.Type == Types.NULL

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.Pause = v.Data.(bool)

		v, err = AMF.DecodeValue(b, offset, size-offset)
		if err != nil {
			return err
		}

		offset += v.Cost
		this.MilliSeconds = v.Data.(float64)

	default:
	}

	return nil
}
