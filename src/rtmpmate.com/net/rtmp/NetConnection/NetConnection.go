package NetConnection

import (
	"container/list"
	"fmt"
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/net/rtmp/Chunk"
	"rtmpmate.com/net/rtmp/Chunk/States"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
	"strconv"
	"syscall"
)

var farID int = 0

type NetConnection struct {
	conn          *net.TCPConn
	chunks        list.List
	farChunkSize  int
	nearChunkSize int

	Agent             string
	Application       string
	AudioCodecs       float64
	AudioSampleAccess string
	Connected         bool
	FarID             string
	ID                string
	Instance          string
	IP                string
	NearID            string
	ObjectEncoding    byte
	PageURL           string
	Protocol          string
	ProtocolVersion   string
	ReadAccess        string
	Referrer          string
	Secure            bool
	URI               string
	VideoCodecs       float64
	VideoSampleAccess string
	VirtualKey        string
	WriteAccess       string

	stats
	events.EventDispatcher
}

type stats struct {
	statsToAdmin

	pingRTT int

	audioQueueBytes int
	videoQueueBytes int
	soQueueBytes    int
	dataQueueBytes  int

	droppedAudioBytes int
	droppedVideoBytes int

	audioQueueMsgs int
	videoQueueMsgs int
	soQueueMsgs    int
	dataQueueMsgs  int

	droppedAudioMsgs int
	droppedVideoMsgs int
}

type statsToAdmin struct {
	connectTime float64

	bytesIn  int
	bytesOut int

	msgIn      int
	msgOut     int
	msgDropped int
}

type Responder struct {
	ID     float64
	Result func()
	Status func()
}

func New(conn *net.TCPConn) (*NetConnection, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	farID++

	var nc NetConnection
	nc.conn = conn
	nc.farChunkSize = 128
	nc.ID = strconv.Itoa(farID)
	nc.ReadAccess = "/"
	nc.WriteAccess = "/"

	nc.AddEventListener("checkBandwidth", nc.CheckBandwidth, 0)
	nc.AddEventListener("getStats", nc.GetStats, 0)

	return &nc, nil
}

func (this *NetConnection) Read(size int, once bool) ([]byte, error) {
	var b = make([]byte, size)
	var err error

	for n, pos := 0, 0; pos < size; {
		n, err = this.conn.Read(b[pos:])
		if err != nil {
			return nil, err
		}

		if once {
			return b[:n], nil
		}

		pos += n
	}

	return b, nil
}

func (this *NetConnection) Write(b []byte) (int, error) {
	return this.conn.Write(b)
}

func (this *NetConnection) WaitRequest() error {
	var b = make([]byte, 4096)

	for {
		n, err := this.conn.Read(b)
		if err != nil {
			return err
		}

		err = this.parseChunk(b[:n], n)
		if err != nil {
			return err
		}
	}
}

/*func (this *NetConnection) requestHandler(b []byte, size int) error {
	var l list.List
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: "level"})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: Level.STATUS})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: "code"})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: Code.NETCONNECTION_CONNECT_SUCCESS})

	var info AMF.AMFValue
	info.Type = Types.OBJECT
	info.Data = l

	return this.Call(Command.RESULT, &Responder{ID: 1}, &AMF.AMFValue{Type: Types.OBJECT, Data: list.List{}}, &info)
}*/

func (this *NetConnection) parseChunk(b []byte, size int) error {
	c := this.getUncompleteChunk()

	for i := 0; i < size; i++ {
		//tmp := uint32(b[i])
		//fmt.Printf("b[%d] = 0x%02x\n", i, tmp)

		switch c.State {
		case States.START:
			c.Fmt = (b[i] >> 6) & 0xFF
			c.CSID = uint32(b[i]) & 0x3F
			c.State = States.FMT

		case States.FMT:
			switch c.CSID {
			case 0:
				c.CSID = uint32(b[i]) + 64
				c.State = States.CSID_1
			case 1:
				c.CSID = uint32(b[i])
				c.State = States.CSID_0
			default:
				if c.Fmt == 3 {
					if c.Extended {
						c.Timestamp = uint32(b[i]) << 24
						c.State = States.EXTENDED_TIMESTAMP_0
					} else {
						i--
						c.State = States.DATA
					}
				} else {
					c.Timestamp = uint32(b[i]) << 16
					c.State = States.TIMESTAMP_0
				}
			}

		case States.CSID_0:
			c.CSID |= uint32(b[i]) << 8
			c.CSID += 64
			c.State = States.CSID_1

		case States.CSID_1:
			if c.Fmt == 3 {
				if c.Extended {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					i--
					c.State = States.DATA
				}
			} else {
				c.Timestamp = uint32(b[i]) << 16
				c.State = States.TIMESTAMP_0
			}

		case States.TIMESTAMP_0:
			c.Timestamp |= uint32(b[i]) << 8
			c.State = States.TIMESTAMP_1

		case States.TIMESTAMP_1:
			c.Timestamp |= uint32(b[i])
			c.State = States.TIMESTAMP_2

		case States.TIMESTAMP_2:
			if c.Fmt == 0 || c.Fmt == 1 {
				c.MessageLength = uint32(b[i]) << 16
				c.State = States.MESSAGE_LENGTH_0
			} else if c.Fmt == 2 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					i--
					c.State = States.DATA
				}
			} else {
				// Should not happen
				return fmt.Errorf("Failed to parse chunk: [1].")
			}

		case States.MESSAGE_LENGTH_0:
			c.MessageLength |= uint32(b[i]) << 8
			c.State = States.MESSAGE_LENGTH_1

		case States.MESSAGE_LENGTH_1:
			c.MessageLength |= uint32(b[i])
			c.State = States.MESSAGE_LENGTH_2

		case States.MESSAGE_LENGTH_2:
			c.MessageTypeID = b[i]
			c.State = States.MESSAGE_TYPE_ID

		case States.MESSAGE_TYPE_ID:
			if c.Fmt == 0 {
				c.MessageStreamID = uint32(b[i]) << 24
				c.State = States.MESSAGE_STREAM_ID_0
			} else if c.Fmt == 1 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					i--
					c.State = States.DATA
				}
			} else {
				// Should not happen
				return fmt.Errorf("Failed to parse chunk: [2].")
			}

		case States.MESSAGE_STREAM_ID_0:
			c.MessageStreamID |= uint32(b[i]) << 16
			c.State = States.MESSAGE_STREAM_ID_1

		case States.MESSAGE_STREAM_ID_1:
			c.MessageStreamID |= uint32(b[i]) << 8
			c.State = States.MESSAGE_STREAM_ID_2

		case States.MESSAGE_STREAM_ID_2:
			c.MessageStreamID |= uint32(b[i])
			c.State = States.MESSAGE_STREAM_ID_3

		case States.MESSAGE_STREAM_ID_3:
			if c.Timestamp == 0xFFFFFF {
				c.Timestamp = uint32(b[i]) << 24
				c.State = States.EXTENDED_TIMESTAMP_0
			} else {
				i--
				c.State = States.DATA
			}

		case States.EXTENDED_TIMESTAMP_0:
			c.Timestamp |= uint32(b[i]) << 16
			c.State = States.EXTENDED_TIMESTAMP_1

		case States.EXTENDED_TIMESTAMP_1:
			c.Timestamp |= uint32(b[i]) << 8
			c.State = States.EXTENDED_TIMESTAMP_2

		case States.EXTENDED_TIMESTAMP_2:
			c.Timestamp |= uint32(b[i])
			c.State = States.EXTENDED_TIMESTAMP_3

		case States.EXTENDED_TIMESTAMP_3:
			fallthrough
		case States.DATA:
			var n int = int(c.MessageLength) - c.Data.Len()
			if n > size-i {
				n = size - i
			}
			if n > this.farChunkSize {
				n = this.farChunkSize
			}

			_, err := c.Data.Write(b[i : i+n])
			if err != nil {
				return err
			}

			i += n - 1

			if c.Data.Len() < int(c.MessageLength) {
				c.State = States.START
			} else if c.Data.Len() == int(c.MessageLength) {
				c.State = States.COMPLETE

				err := this.parseMessage(c)
				if err != nil {
					return err
				}

				if i < size-1 {
					c = this.getUncompleteChunk()
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [3].")
			}

		default:
			return fmt.Errorf("Failed to parse chunk: [4].")
		}
	}

	return nil
}

func (this *NetConnection) parseMessage(c *Chunk.Chunk) error {
	b := c.Data.Bytes()
	size := c.Data.Len()

	switch c.MessageTypeID {
	case Types.SET_CHUNK_SIZE:
	case Types.ABORT:
	case Types.ACK:
	case Types.USER_CONTROL:
	case Types.ACK_SIZE:
	case Types.BANDWIDTH:
	case Types.EDGE:
	case Types.AUDIO:
	case Types.VIDEO:
	case Types.AMF3_DATA:
	case Types.AMF3_SHARED_OBJECT:
	case Types.AMF3_COMMAND:
	case Types.DATA:
	case Types.SHARED_OBJECT:
	case Types.COMMAND:
		m, _ := CommandMessage.New(AMF.AMF0)
		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onCommand(m)
		if err != nil {
			return err
		}

	case Types.AGGREGATE:
		/*m, _ := Message.New()
		m.Type = b[0]
		m.Length = uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		m.Timestamp = binary.BigEndian.Uint32(b[4:8])
		m.StreamID = uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])*/

	default:
	}

	return nil
}

func (this *NetConnection) onCommand(m *CommandMessage.CommandMessage) error {
	switch m.Name.Data {
	// NetConnection Commands
	case Commands.CONNECT:
		if this.Connected {
			return fmt.Errorf("already connected")
		}

		var encoder AMF.Encoder
		encoder.EncodeString(Commands.RESULT)
		encoder.EncodeNumber(1)

		var prop AMF.AMFObject
		prop.Init()
		prop.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.STRING, Key: "fmsVer", Data: "FMS/5,0,3,3029"})
		prop.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.DOUBLE, Key: "capabilities", Data: float64(255)})
		prop.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.DOUBLE, Key: "mode", Data: float64(1)})
		encoder.EncodeObject(&prop)

		var info AMF.AMFObject
		info.Init()
		info.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.STRING, Key: "level", Data: Level.STATUS})
		info.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.STRING, Key: "code", Data: Code.NETCONNECTION_CONNECT_SUCCESS})
		info.Data.PushBack(&AMF.AMFValue{Type: AMFTypes.STRING, Key: "description", Data: "Connection connected."})
		encoder.EncodeObject(&info)

		b, err := encoder.Encode()
		if err != nil {
			return err
		}

		_, err = this.Write(b)
		if err != nil {
			return err
		}

		this.Connected = true

	case Commands.CLOSE:
	case Commands.CREATE_STREAM:

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

func (this *NetConnection) Call(methodName string, resultObj *Responder, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(methodName)
	encoder.EncodeNumber(resultObj.ID)

	for _, v := range args {
		encoder.EncodeValue(v)
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	this.Write(b)

	return nil
}

func (this *NetConnection) Ping() {

}

func (this *NetConnection) CheckBandwidth() {

}

func (this *NetConnection) GetStats() *stats {
	return &this.stats
}

func (this *NetConnection) Close() error {
	return this.conn.Close()
}

func (this *NetConnection) getUncompleteChunk() *Chunk.Chunk {
	var chunk *Chunk.Chunk
	var extended bool

	element := this.chunks.Back()
	if element != nil {
		chunk = element.Value.(*Chunk.Chunk)

		if chunk.State != States.COMPLETE {
			return chunk
		}

		extended = chunk.Extended
	}

	chunk, _ = Chunk.New(extended)
	this.chunks.PushBack(chunk)

	return chunk
}
