package Instance

import (
	"rtmpmate.com/net/rtmp/Client"
	"rtmpmate.com/net/rtmp/Stream"
	"sync"
)

type Instance struct {
	Name       string
	clients    map[string]*Client.Client
	clientsMtx sync.RWMutex
	streams    map[string]*Stream.Stream
	streamsMtx sync.RWMutex

	statsToAdmin
}

type stats struct {
	bytesIn  int
	bytesOut int

	msgIn      int
	msgOut     int
	msgDropped int
}

type statsToAdmin struct {
	stats

	launchTime float64 // time the application started.
	upTime     float64 // time, in seconds, the application has been running.

	totalConnects    int // Total number of socket connections to this instance since the instance was started.
	totalDisconnects int // Total number of socket disconnections from this instance since the instance was started.

	accepted        int // Total number of connection attempts accepted by this instance.
	rejected        int // Total number of connection attempts rejected by this instance.
	connected       int // Total number of connections currently active.
	normalConnects  int // Total number of normal connections.
	virtualConnects int // Total number of connections through a remote edge.
	adminConnects   int // Total number of administrator connections.
	debugConnects   int // Total number of debug connections.

	pid   int  // The pid of the core process running the instance.
	debug bool // true if a debug session is enabled, otherwise false.

	swfVerificationAttempts int
	swfVerificationMatches  int
	swfVerificationFailures int
}

func New(name string) (*Instance, error) {
	if name == "" {
		name = "_definst_"
	}

	var inst Instance
	inst.Name = name
	inst.clients = make(map[string]*Client.Client)
	inst.streams = make(map[string]*Stream.Stream)

	return &inst, nil
}

func (this *Instance) GetStats() *stats {
	return &this.stats
}

func (this *Instance) OnConnect(client *Client.Client) {
	this.clientsMtx.Lock()
	this.clients[client.ID] = client
	this.clientsMtx.Unlock()
}

func (this *Instance) OnDisconnect(client *Client.Client) {
	this.clientsMtx.Lock()
	this.clients[client.ID] = nil
	this.clientsMtx.Unlock()
}
