package Client

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/net/rtmp/Chunk"
	"rtmpmate.com/net/rtmp/Chunk/States"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/util/AMF"
	"strconv"
	"syscall"
)

var clientID int

type Client struct {
	conn      *net.TCPConn
	chunks    list.List
	chunkSize int

	Application string
	Instance    string

	Agent             string
	AudioSampleAccess string
	IP                string
	PageURL           string
	Protocol          string
	ProtocolVersion   string
	ReadAccess        string
	Referrer          string
	Secure            bool
	URI               string
	VideoSampleAccess string
	VirtualKey        string
	WriteAccess       string

	stats
	events.EventDispatcher
}

type stats struct {
	ID      string
	pingRTT int

	statsToAdmin

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

func New(conn *net.TCPConn) (*Client, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	clientID++

	var client Client
	client.conn = conn
	client.ID = strconv.Itoa(clientID)
	client.ReadAccess = "/"
	client.WriteAccess = "/"

	client.AddEventListener("checkBandwidth", client.CheckBandwidth, 0)
	client.AddEventListener("getStats", client.GetStats, 0)

	return &client, nil
}

func (this *Client) Read(size int, once bool) ([]byte, error) {
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

func (this *Client) Write(b []byte) (int, error) {
	return this.conn.Write(b)
}

func (this *Client) WaitRequest() error {
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

/*func (this *Client) requestHandler(b []byte, size int) error {
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

func (this *Client) parseChunk(b []byte, size int) error {
	c := this.getUncompleteChunk()

	for i := 0; i < size; i++ {
		tmp := uint32(b[i])
		fmt.Printf("b[%d] = 0x%02x\n", i, tmp)

		switch c.State {
		case States.START:
			c.Fmt = (b[i] >> 6) & 0xFF
			c.CSID = uint32(b[i]) & 0x3F
			c.State = States.FMT

		case States.FMT:
			if c.CSID == 0 {
				c.CSID = uint32(b[i]) + 64
				c.State = States.CSID_1
			} else if c.CSID == 1 {
				c.CSID = uint32(b[i])
				c.State = States.CSID_0
			} else {
				if c.Fmt == 3 {
					if c.Extended {
						c.Timestamp = uint32(b[i]) << 24
						c.State = States.EXTENDED_TIMESTAMP_0
					} else {
						i--
						c.State = States.EXTENDED_TIMESTAMP_3
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
					c.State = States.EXTENDED_TIMESTAMP_3
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
				}
				c.State = States.EXTENDED_TIMESTAMP_0
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
				}
				c.State = States.EXTENDED_TIMESTAMP_0
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
			}
			c.State = States.EXTENDED_TIMESTAMP_0

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
			need := int(c.MessageLength - c.Loaded)
			remain := size - i

			var n int
			if need >= remain {
				n = remain
			} else {
				n = need
			}

			_, err := c.Data.Write(b[i : i+n])
			if err != nil {
				return err
			}

			c.Loaded += uint32(n)

			if c.Loaded < c.MessageLength {
				c.State = States.START
			} else if c.Loaded == c.MessageLength {
				c.State = States.COMPLETE
				this.parseMessage(c)

				if i < size {
					c = this.getUncompleteChunk()
				}
			} else {
				return fmt.Errorf("Failed to parse chunk: [3].")
			}

			i += n - 1 // avoid i++ of the for loop

		default:
			return fmt.Errorf("Failed to parse chunk: [4].")
		}
	}

	return nil
}

func (this *Client) parseMessage(chunk *Chunk.Chunk) {
	b := chunk.Data.Bytes()
	m, _ := Message.New()
	m.Type = b[0]
	m.Length = uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	m.Timestamp = binary.BigEndian.Uint32(b[4:8])
	m.StreamID = uint32(b[8])<<16 | uint32(b[9])<<8 | uint32(b[10])
}

func (this *Client) Call(methodName string, resultObj *Responder, args ...*AMF.AMFValue) error {
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

func (this *Client) Ping() {

}

func (this *Client) CheckBandwidth() {

}

func (this *Client) GetStats() *stats {
	return &this.stats
}

func (this *Client) Close() error {
	return this.conn.Close()
}

func (this *Client) getUncompleteChunk() *Chunk.Chunk {
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
