package Client

import (
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/util/AMF"
	"syscall"
)

type Client struct {
	conn *net.TCPConn

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

	ApplicationName string
	InstanceName    string

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
	Result func()
	Status func()
}

func New(conn *net.TCPConn, id string) (*Client, error) {
	if conn == nil || id == "" {
		return nil, syscall.EINVAL
	}

	var client Client
	client.ID = id
	client.ReadAccess = "/"
	client.WriteAccess = "/"

	client.AddEventListener("checkBandwidth", client.CheckBandwidth, 0)
	client.AddEventListener("getStats", client.GetStats, 0)

	return &client, nil
}

func (this *Client) Recv() {

}

func (this *Client) Call(methodName string, resultObj *Responder, args ...*AMF.AMFValue) bool {
	return true
}

func (this *Client) Ping() {

}

func (this *Client) CheckBandwidth() {

}

func (this *Client) GetStats() *stats {
	return &this.stats
}
