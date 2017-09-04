package NetConnection

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"regexp"
	"rtmpmate.com/events"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/net/rtmp/Chunk"
	"rtmpmate.com/net/rtmp/Chunk/CSIDs"
	"rtmpmate.com/net/rtmp/Chunk/States"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
	"strconv"
	"syscall"
)

var (
	farID    int = 0
	urlRe, _     = regexp.Compile("^(rtmp[es]*)://([a-z0-9.-]+)(:([0-9]+))?/([a-z0-9.-_]+)(/([a-z0-9.-_]*))?$")
)

type NetConnection struct {
	conn          *net.TCPConn
	chunks        list.List
	farChunkSize  int
	nearChunkSize int

	Agent             string
	Application       string
	AudioCodecs       uint64
	AudioSampleAccess string
	Connected         bool
	FarID             string
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
	VideoCodecs       uint64
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
	ID     uint64
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
	nc.nearChunkSize = 128

	nc.FarID = strconv.Itoa(farID)
	nc.Instance = "_definst_"
	nc.ObjectEncoding = AMF.AMF0
	nc.AudioSampleAccess = "/"
	nc.VideoSampleAccess = "/"

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

func (this *NetConnection) WriteByChunk(b []byte, csid int, h *Message.Header) (int, error) {
	if h.Length < 2 {
		return 0, fmt.Errorf("chunk data (len=%d) not enough", h.Length)
	}

	var c Chunk.Chunk
	c.Fmt = 0

	for i := 0; i < h.Length; /* void */ {
		if csid <= 63 {
			c.Data.WriteByte((c.Fmt << 6) | byte(csid))
		} else if csid <= 319 {
			c.Data.WriteByte((c.Fmt << 6) | 0x00)
			c.Data.WriteByte(byte(csid - 64))
		} else if csid <= 65599 {
			tmp := uint16(csid)
			c.Data.WriteByte((c.Fmt << 6) | 0x01)
			err := binary.Write(&c.Data, binary.LittleEndian, &tmp)
			if err != nil {
				return i, err
			}
		} else {
			return i, fmt.Errorf("chunk size (%d) out of range", h.Length)
		}

		if c.Fmt <= 2 {
			if h.Timestamp >= 0xFFFFFF {
				c.Data.Write([]byte{0xFF, 0xFF, 0xFF})
			} else {
				c.Data.Write([]byte{
					byte(h.Timestamp>>16) & 0xFF,
					byte(h.Timestamp>>8) & 0xFF,
					byte(h.Timestamp>>0) & 0xFF,
				})
			}
		}
		if c.Fmt <= 1 {
			c.Data.Write([]byte{
				byte(h.Length>>16) & 0xFF,
				byte(h.Length>>8) & 0xFF,
				byte(h.Length>>0) & 0xFF,
			})
			c.Data.WriteByte(h.Type)
		}
		if c.Fmt == 0 {
			binary.Write(&c.Data, binary.LittleEndian, &h.StreamID)
		}

		// Extended Timestamp
		if h.Timestamp >= 0xFFFFFF {
			binary.Write(&c.Data, binary.BigEndian, &h.Timestamp)
		}

		// Chunk Data
		n := h.Length - i
		if n > this.nearChunkSize {
			n = this.nearChunkSize
		}

		_, err := c.Data.Write(b[i : i+n])
		if err != nil {
			return i, err
		}

		//fmt.Println(c.Data.Bytes())

		i += n

		if i < h.Length {
			switch h.Type {
			default:
				c.Fmt = 3
			}
		} else if i == h.Length {
			cs := c.Data.Bytes()
			_, err = this.Write(cs)
			if err != nil {
				return i, err
			}

			/*size := len(cs)
			for x := 0; x < size; x += 16 {
				fmt.Printf("\n")

				for y := 0; y < int(math.Min(float64(size-x), 16)); y++ {
					fmt.Printf("%02x ", cs[x+y])

					if y == 7 {
						fmt.Printf(" ")
					}
				}
			}*/
		} else {
			return i, fmt.Errorf("wrote too much")
		}
	}

	return h.Length, nil
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

			if c.Fmt == 3 && c.Extended == false {
				c.State = States.DATA
			} else {
				c.State = States.FMT
			}

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
						return fmt.Errorf("Failed to parse chunk: [1].")
					}
				} else {
					c.Timestamp = uint32(b[i]) << 16
					c.State = States.TIMESTAMP_0
				}
			}

		case States.CSID_0:
			c.CSID |= uint32(b[i]) << 8
			c.CSID += 64

			if c.Fmt == 3 && c.Extended == false {
				c.State = States.DATA
			} else {
				c.State = States.CSID_1
			}

		case States.CSID_1:
			if c.Fmt == 3 {
				if c.Extended {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [2].")
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

			if c.Fmt == 2 && c.Timestamp != 0xFFFFFF {
				c.State = States.DATA
			} else {
				c.State = States.TIMESTAMP_2
			}

		case States.TIMESTAMP_2:
			if c.Fmt == 0 || c.Fmt == 1 {
				c.MessageLength = int(b[i]) << 16
				c.State = States.MESSAGE_LENGTH_0
			} else if c.Fmt == 2 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [3].")
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [4].")
			}

		case States.MESSAGE_LENGTH_0:
			c.MessageLength |= int(b[i]) << 8
			c.State = States.MESSAGE_LENGTH_1

		case States.MESSAGE_LENGTH_1:
			c.MessageLength |= int(b[i])
			c.State = States.MESSAGE_LENGTH_2

		case States.MESSAGE_LENGTH_2:
			c.MessageTypeID = b[i]

			if c.Fmt == 1 && c.Timestamp != 0xFFFFFF {
				c.State = States.DATA
			} else {
				c.State = States.MESSAGE_TYPE_ID
			}

		case States.MESSAGE_TYPE_ID:
			if c.Fmt == 0 {
				c.MessageStreamID = uint32(b[i])
				c.State = States.MESSAGE_STREAM_ID_0
			} else if c.Fmt == 1 {
				if c.Timestamp == 0xFFFFFF {
					c.Timestamp = uint32(b[i]) << 24
					c.State = States.EXTENDED_TIMESTAMP_0
				} else {
					return fmt.Errorf("Failed to parse chunk: [5].")
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [6].")
			}

		case States.MESSAGE_STREAM_ID_0:
			c.MessageStreamID |= uint32(b[i]) << 8
			c.State = States.MESSAGE_STREAM_ID_1

		case States.MESSAGE_STREAM_ID_1:
			c.MessageStreamID |= uint32(b[i]) << 16
			c.State = States.MESSAGE_STREAM_ID_2

		case States.MESSAGE_STREAM_ID_2:
			c.MessageStreamID |= uint32(b[i]) << 24
			if c.Timestamp == 0xFFFFFF {
				c.State = States.MESSAGE_STREAM_ID_3
			} else {
				c.State = States.DATA
			}

		case States.MESSAGE_STREAM_ID_3:
			if c.Timestamp == 0xFFFFFF {
				c.Timestamp = uint32(b[i]) << 24
				c.State = States.EXTENDED_TIMESTAMP_0
			} else {
				return fmt.Errorf("Failed to parse chunk: [7].")
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
			n := int(c.MessageLength) - c.Data.Len()
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
				return fmt.Errorf("Failed to parse chunk: [8].")
			}

		default:
			return fmt.Errorf("Failed to parse chunk: [9].")
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
		m, _ := CommandMessage.New(this.ObjectEncoding)
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		encoding, _ := m.CommandObject.Get("objectEncoding")
		if encoding != nil && encoding.Data.(float64) != 0 {
			this.ObjectEncoding = AMF.AMF3
			m.Type = Types.AMF3_COMMAND
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
	var encoder AMF.Encoder

	switch m.Name.Data {
	// NetConnection Commands
	case Commands.CONNECT:
		if this.Connected {
			return fmt.Errorf("already connected")
		}

		app, _ := m.CommandObject.Get("app")
		if app != nil {
			this.Application = app.Data.(string)
		}

		tcUrl, _ := m.CommandObject.Get("tcUrl")
		if tcUrl != nil {
			arr := urlRe.FindStringSubmatch(tcUrl.Data.(string))
			if arr != nil {
				inst := arr[len(arr)-1]
				if inst != "" {
					this.Instance = inst
				}
			}
		}

		encoder.EncodeString(Commands.RESULT)
		encoder.EncodeNumber(1)

		var prop AMF.AMFObject
		prop.Init()
		prop.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.STRING,
			Key:  "fmsVer",
			Data: "FMS/5,0,3,3029",
		})
		prop.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.DOUBLE,
			Key:  "capabilities",
			Data: float64(255),
		})
		prop.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.DOUBLE,
			Key:  "mode",
			Data: float64(1),
		})
		encoder.EncodeObject(&prop)

		var info AMF.AMFObject
		info.Init()
		info.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.STRING,
			Key:  "level",
			Data: Level.STATUS,
		})
		info.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.STRING,
			Key:  "code",
			Data: Code.NETCONNECTION_CONNECT_SUCCESS,
		})
		info.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.STRING,
			Key:  "description",
			Data: "Connection succeeded.",
		})
		info.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.DOUBLE,
			Key:  "objectEncoding",
			Data: float64(this.ObjectEncoding),
		})
		var data list.List
		data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.STRING,
			Key:  "version",
			Data: "5,0,3,3029",
		})
		info.Data.PushBack(&AMF.AMFValue{
			Type: AMFTypes.ECMA_ARRAY,
			Key:  "data",
			Data: data,
		})
		encoder.EncodeObject(&info)

		this.Connected = true

	case Commands.CLOSE:
	case Commands.CREATE_STREAM:
		var err error
		if (this.ReadAccess == "/" || this.ReadAccess == "/"+this.Application) &&
			(this.WriteAccess == "/" || this.WriteAccess == "/"+this.Application) {
			encoder.EncodeString(Commands.RESULT)
		} else {
			err = fmt.Errorf("Access denied.")
			encoder.EncodeString(Commands.ERROR)
		}

		encoder.EncodeNumber(math.Float64frombits(m.TransactionID))
		encoder.EncodeNull()

		if err != nil {
			var info AMF.AMFObject
			info.Init()
			info.Data.PushBack(&AMF.AMFValue{
				Type: AMFTypes.STRING,
				Key:  "level",
				Data: Level.ERROR,
			})
			info.Data.PushBack(&AMF.AMFValue{
				Type: AMFTypes.STRING,
				Key:  "code",
				Data: err.Error(),
			})
			encoder.EncodeObject(&info)
		} else {
			encoder.EncodeNumber(math.Float64frombits(m.StreamID))
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

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	m.Length = len(b)

	_, err = this.WriteByChunk(b, CSIDs.COMMAND, &m.Header)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetConnection) Call(methodName string, resultObj *Responder, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(methodName)
	encoder.EncodeNumber(float64(resultObj.ID))

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
