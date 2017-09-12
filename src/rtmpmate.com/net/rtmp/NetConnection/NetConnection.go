package NetConnection

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"net"
	"regexp"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/events/ServerEvent"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/Application"
	"rtmpmate.com/net/rtmp/Chunk"
	"rtmpmate.com/net/rtmp/Chunk/CSIDs"
	"rtmpmate.com/net/rtmp/Chunk/States"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/AggregateMessage"
	"rtmpmate.com/net/rtmp/Message/AudioMessage"
	"rtmpmate.com/net/rtmp/Message/BandwidthMessage"
	"rtmpmate.com/net/rtmp/Message/CommandMessage"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/DataMessage"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/Message/UserControlMessage"
	EventTypes "rtmpmate.com/net/rtmp/Message/UserControlMessage/Event/Types"
	"rtmpmate.com/net/rtmp/Message/VideoMessage"
	"rtmpmate.com/net/rtmp/Responder"
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
	bwLimitType       byte
	conn              *net.TCPConn
	chunks            list.List
	farAckWindowSize  uint32
	farChunkSize      int
	nearAckWindowSize uint32
	nearChunkSize     int
	responders        map[int]*Responder.Responder

	Agent             string
	App               *Application.Application
	AppName           string
	AudioCodecs       uint64
	AudioSampleAccess string
	Connected         bool
	FarID             string
	InstName          string
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

	bytesIn  uint32
	bytesOut uint32

	msgIn      int
	msgOut     int
	msgDropped int
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
	nc.responders = make(map[int]*Responder.Responder)

	nc.FarID = strconv.Itoa(farID)
	nc.InstName = "_definst_"
	nc.ObjectEncoding = AMF.AMF0
	nc.ReadAccess = "/"
	nc.WriteAccess = "/"
	nc.AudioSampleAccess = "/"
	nc.VideoSampleAccess = "/"

	nc.AddEventListener(CommandEvent.CONNECT, nc.onConnect, 0)

	return &nc, nil
}

func (this *NetConnection) Read(b []byte) (int, error) {
	return this.conn.Read(b)
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

	this.bytesOut += uint32(h.Length)

	return h.Length, nil
}

func (this *NetConnection) WaitRequest() error {
	var b = make([]byte, 14+4096)

	for {
		n, err := this.conn.Read(b)
		if err != nil {
			return err
		}

		this.bytesIn += uint32(n)

		err = this.parseChunk(b[:n], n)
		if err != nil {
			return err
		}
	}
}

func (this *NetConnection) parseChunk(b []byte, size int) error {
	c := this.getUncompleteChunk()

	for i := 0; i < size; i++ {
		//tmp := uint32(b[i])
		//fmt.Printf("b[%d] = 0x%02x\n", i, tmp)

		switch c.State {
		case States.START:
			c.Fmt = (b[i] >> 6) & 0xFF
			c.CSID = uint32(b[i]) & 0x3F

			this.extendsFromPrecedingChunk(c)
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
			c.State = States.DATA
			fallthrough
		case States.DATA:
			n := c.MessageLength - c.Data.Len()
			if n > size-i {
				n = size - i
			}
			if n > this.farChunkSize-c.Loaded {
				n = this.farChunkSize - c.Loaded
				c.Loaded = 0
				c.State = States.START
			} else {
				c.Loaded += n
			}

			_, err := c.Data.Write(b[i : i+n])
			if err != nil {
				return err
			}

			i += n - 1

			if c.Data.Len() < c.MessageLength {
				// c.State = States.DATA
			} else if c.Data.Len() == c.MessageLength {
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
	if c.MessageTypeID != 0x08 && c.MessageTypeID != 0x09 {
		fmt.Printf("onMessage: 0x%02x.\n", c.MessageTypeID)
	}

	b := c.Data.Bytes()
	size := c.Data.Len()

	switch c.MessageTypeID {
	case Types.SET_CHUNK_SIZE:
		this.farChunkSize = int(binary.BigEndian.Uint32(b) & 0x7FFFFFFF)
		fmt.Printf("Set farChunkSize to %d.\n", this.farChunkSize)

	case Types.ABORT:
		csid := binary.BigEndian.Uint32(b)
		fmt.Printf("Abort chunk stream %d.\n", csid)

		element := this.chunks.Back()
		if element != nil {
			c := element.Value.(*Chunk.Chunk)
			if c.State != States.COMPLETE && c.CSID == csid {
				this.chunks.Remove(element)
				fmt.Printf("Removed uncomplete chunk %d.\n", csid)
			}
		}

	case Types.ACK:
		sequenceNumber := binary.BigEndian.Uint32(b)
		fmt.Printf("Sequence Number: %d, Bytes out: %d.\n", sequenceNumber, this.bytesOut)

		if sequenceNumber != this.bytesOut {
			fmt.Printf("Should I close the connection?\n")
		}

	case Types.USER_CONTROL:
		m, _ := UserControlMessage.New()
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onUserControl(m)
		if err != nil {
			return err
		}

	case Types.ACK_WINDOW_SIZE:
		this.farAckWindowSize = binary.BigEndian.Uint32(b)
		fmt.Printf("Set farAckWindowSize to %d.\n", this.farAckWindowSize)

	case Types.BANDWIDTH:
		m, _ := BandwidthMessage.New()
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onBandwidth(m)
		if err != nil {
			return err
		}

	case Types.EDGE:
		// TODO:

	case Types.AUDIO:
		m, _ := AudioMessage.New()
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(AudioEvent.New(AudioEvent.DATA, this, m))

	case Types.VIDEO:
		m, _ := VideoMessage.New()
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(VideoEvent.New(VideoEvent.DATA, this, m))

	case Types.AMF3_DATA:
		fallthrough
	case Types.DATA:
		m, _ := DataMessage.New(this.ObjectEncoding)
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		this.DispatchEvent(DataFrameEvent.New(m.Handler, this, m.Key, m.Data))

	case Types.AMF3_SHARED_OBJECT:
		fallthrough
	case Types.SHARED_OBJECT:
		// TODO:

	case Types.AMF3_COMMAND:
		fallthrough
	case Types.COMMAND:
		m, _ := CommandMessage.New(this.ObjectEncoding)
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		if m.CommandObject != nil {
			encoding, _ := m.CommandObject.Get("objectEncoding")
			if encoding != nil && encoding.Data.(float64) != 0 {
				this.ObjectEncoding = AMF.AMF3
				m.Type = Types.AMF3_COMMAND
			}
		}

		err = this.onCommand(m)
		if err != nil {
			return err
		}

	case Types.AGGREGATE:
		m, _ := AggregateMessage.New()
		m.Header.Timestamp = c.Timestamp
		m.Header.StreamID = c.MessageStreamID

		err := m.Parse(b, 0, size)
		if err != nil {
			return err
		}

		err = this.onAggregate(m)
		if err != nil {
			return err
		}

	default:
	}

	return nil
}

func (this *NetConnection) onUserControl(m *UserControlMessage.UserControlMessage) error {
	fmt.Printf("onUserControl: type=%d.\n", m.Event.Type)

	switch m.Event.Type {
	case EventTypes.STREAM_BEGIN:
		streamID := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Stream Begin: id=%d.\n", streamID)

	case EventTypes.STREAM_EOF:
		streamID := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Stream EOF: id=%d.\n", streamID)

	case EventTypes.STREAM_DRY:
		streamID := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Stream Dry: id=%d.\n", streamID)

	case EventTypes.SET_BUFFER_LENGTH:
		streamID := binary.BigEndian.Uint32(m.Event.Data)
		bufferLength := binary.BigEndian.Uint32(m.Event.Data[4:])
		fmt.Printf("Set Buffer Length: id=%d, len=%dms.\n", streamID, bufferLength)

	case EventTypes.STREAM_IS_RECORDED:
		streamID := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Stream is Recorded: id=%d.\n", streamID)

	case EventTypes.PING_REQUEST:
		timestamp := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Ping Request: timestamp=%d.\n", timestamp)

	case EventTypes.PING_RESPONSE:
		timestamp := binary.BigEndian.Uint32(m.Event.Data)
		fmt.Printf("Ping Response: timestamp=%d.\n", timestamp)

	default:
	}

	return nil
}

func (this *NetConnection) onBandwidth(m *BandwidthMessage.BandwidthMessage) error {
	fmt.Printf("onBandwidth: ack=%d, limit=%d.\n", m.AckWindowSize, m.LimitType)

	this.nearAckWindowSize = m.AckWindowSize
	this.bwLimitType = m.LimitType

	return nil
}

func (this *NetConnection) onCommand(m *CommandMessage.CommandMessage) error {
	fmt.Printf("onCommand: name=%s.\n", m.Name)

	var encoder AMF.Encoder

	if this.HasEventListener(m.Name) {
		this.DispatchEvent(CommandEvent.New(m.Name, this, m, &encoder))
	} else {
		// Should not return error, this might be an user call
		fmt.Printf("No handler found of event \"%s\".\n", m.Name)
		return nil
	}

	m.Length = encoder.Len()
	if m.Length == 0 {
		return nil
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	_, err = this.WriteByChunk(b, CSIDs.COMMAND, &m.Header)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetConnection) onConnect(e *CommandEvent.CommandEvent) {
	if this.Connected {
		fmt.Printf("Already connected.\n")
		return
	}

	// Init properties
	app, _ := e.Message.CommandObject.Get("app")
	if app != nil {
		this.AppName = app.Data.(string)
	}

	tcUrl, _ := e.Message.CommandObject.Get("tcUrl")
	if tcUrl != nil {
		arr := urlRe.FindStringSubmatch(tcUrl.Data.(string))
		if arr != nil {
			instName := arr[len(arr)-1]
			if instName != "" {
				this.InstName = instName
			}
		}
	}

	// Encode response
	var command, level, code string

	if this.ReadAccess == "/" || this.ReadAccess == "/"+this.AppName {
		this.App, _ = Application.Get(this.AppName)

		this.AddEventListener(CommandEvent.CLOSE, this.onClose, 0)
		this.AddEventListener(CommandEvent.RESULT, this.onResult, 0)
		this.AddEventListener(CommandEvent.ERROR, this.onError, 0)
		this.AddEventListener(CommandEvent.CHECK_BANDWIDTH, this.onCheckBandwidth, 0)
		this.AddEventListener(CommandEvent.GET_STATS, this.onGetStats, 0)

		command = Commands.RESULT
		level = Level.STATUS
		code = Code.NETCONNECTION_CONNECT_SUCCESS
	} else {
		command = Commands.ERROR
		level = Level.ERROR
		code = Code.NETCONNECTION_CONNECT_REJECTED
	}

	e.Encoder.EncodeString(command)
	e.Encoder.EncodeNumber(1)

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
	e.Encoder.EncodeObject(&prop)

	var data list.List
	data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "version",
		Data: "5,0,3,3029",
	})

	info, _ := this.GetInfoObject(level, code, "Connection succeeded")
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.DOUBLE,
		Key:  "objectEncoding",
		Data: float64(this.ObjectEncoding),
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.ECMA_ARRAY,
		Key:  "data",
		Data: data,
	})
	e.Encoder.EncodeObject(info)

	// TODO: reject
	this.Connected = true
	this.App.DispatchEvent(ServerEvent.New(ServerEvent.CONNECT, this.App, this, nil))
}

func (this *NetConnection) onClose(e *CommandEvent.CommandEvent) {
	this.Close()
}

func (this *NetConnection) onResult(e *CommandEvent.CommandEvent) {

}

func (this *NetConnection) onError(e *CommandEvent.CommandEvent) {

}

func (this *NetConnection) onCheckBandwidth(e *Event.Event) {

}

func (this *NetConnection) onGetStats(e *Event.Event) {

}

func (this *NetConnection) onAggregate(m *AggregateMessage.AggregateMessage) error {
	fmt.Printf("onAggregate: id=%d, timstamp=%d, length=%d.\n", m.StreamID, m.Timestamp, m.Length)
	return nil
}

func (this *NetConnection) setChunkSize(size int) error {
	var encoder AMF.Encoder
	encoder.AppendInt32(int32(size), false)

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h = Message.Header{
		Type:      Types.SET_CHUNK_SIZE,
		Length:    encoder.Len(),
		Timestamp: 0,
		StreamID:  0,
	}

	_, err = this.WriteByChunk(b, CSIDs.PROTOCOL_CONTROL, &h)
	if err != nil {
		return err
	}

	this.nearChunkSize = size

	return nil
}

func (this *NetConnection) abort() error {
	return nil
}

func (this *NetConnection) sendAckSequence(num int) error {
	return nil
}

func (this *NetConnection) sendUserControl(event uint16) error {
	return nil
}

func (this *NetConnection) setAckWindowSize(size uint32) error {
	this.nearAckWindowSize = size
	return nil
}

func (this *NetConnection) Connect(uri string, args ...*AMF.AMFValue) error {
	return nil
}

func (this *NetConnection) CreateStream() error {
	return nil
}

func (this *NetConnection) Call(method string, res *Responder.Responder, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(method)
	encoder.EncodeNumber(float64(res.ID))

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

func (this *NetConnection) Close() error {
	this.Connected = false
	err := this.conn.Close()

	this.DispatchEvent(Event.New(Event.CLOSE, this))
	this.App.DispatchEvent(ServerEvent.New(ServerEvent.DISCONNECT, this.App, this, nil))

	return err
}

func (this *NetConnection) GetAppName() string {
	return this.AppName
}

func (this *NetConnection) GetInstName() string {
	return this.InstName
}

func (this *NetConnection) GetFarID() string {
	return this.FarID
}

func (this *NetConnection) GetInfoObject(level string, code string, description string) (*AMF.AMFObject, error) {
	var info AMF.AMFObject
	info.Init()

	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "level",
		Data: level,
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "code",
		Data: code,
	})
	info.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "description",
		Data: description,
	})

	return &info, nil
}

func (this *NetConnection) getUncompleteChunk() *Chunk.Chunk {
	for e := this.chunks.Back(); e != nil; e = e.Prev() {
		c := e.Value.(*Chunk.Chunk)
		if c.State != States.COMPLETE {
			return c
		}

		break
	}

	c, _ := Chunk.New()
	this.chunks.PushBack(c)

	return c
}

func (this *NetConnection) extendsFromPrecedingChunk(c *Chunk.Chunk) {
	if c.Fmt != 1 && c.Fmt != 2 {
		return
	}

	for e, checking := this.chunks.Back(), false; e != nil; e = e.Prev() {
		b := e.Value.(*Chunk.Chunk)
		if b.CSID != c.CSID {
			continue
		}

		if checking == false {
			checking = true
			continue
		}

		if c.Fmt >= 1 {
			c.MessageStreamID = b.MessageStreamID
		}
		if c.Fmt >= 2 {
			c.MessageLength = b.MessageLength
			c.MessageTypeID = b.MessageTypeID
		}

		break
	}
}
