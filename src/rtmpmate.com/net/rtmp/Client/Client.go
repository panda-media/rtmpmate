package Client

import (
	"container/list"
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/net/rtmp/Message/Command"
	"rtmpmate.com/util/AMF"
	"rtmpmate.com/util/AMF/Types"
	"strconv"
	"syscall"
)

var index int

type Client struct {
	conn *net.TCPConn

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

	index++

	var client Client
	client.conn = conn
	client.ID = strconv.Itoa(index)
	client.ReadAccess = "/"
	client.WriteAccess = "/"

	client.AddEventListener("checkBandwidth", client.CheckBandwidth, 0)
	client.AddEventListener("getStats", client.GetStats, 0)

	return &client, nil
}

func (this *Client) Read(size int, once bool) ([]byte, error) {
	var data = make([]byte, size)
	var err error

	for n, pos := 0, 0; pos < size; {
		n, err = this.conn.Read(data[pos:])
		if err != nil {
			return nil, err
		}

		if once {
			return data[:n], nil
		}

		pos += n
	}

	return data, nil
}

func (this *Client) Write(b []byte) (int, error) {
	return this.conn.Write(b)
}

func (this *Client) WaitRequest() error {
	var data = make([]byte, 4096)

	for {
		n, err := this.conn.Read(data)
		if err != nil {
			return err
		}

		err = this.requestHandler(data[:n])
		if err != nil {
			return err
		}
	}
}

func (this *Client) requestHandler(data []byte) error {
	var l list.List
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: "level"})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: Level.STATUS})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: "code"})
	l.PushBack(&AMF.AMFValue{Type: Types.STRING, Data: Code.NETCONNECTION_CONNECT_SUCCESS})

	var info AMF.AMFValue
	info.Type = Types.OBJECT
	info.Data = l

	return this.Call(Command.RESULT, &Responder{ID: 1}, &AMF.AMFValue{Type: Types.OBJECT, Data: list.List{}}, &info)
}

func (this *Client) Call(methodName string, resultObj *Responder, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(methodName)
	encoder.EncodeNumber(resultObj.ID)

	for _, v := range args {
		encoder.EncodeValue(v)
	}

	data, err := encoder.Encode()
	if err != nil {
		return err
	}

	this.Write(data)

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
