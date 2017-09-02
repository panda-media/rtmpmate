package Instance

import (
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/Stream"
	"sync"
)

type Instance struct {
	Name           string
	connections    map[string]*NetConnection.NetConnection
	connectionsMtx sync.RWMutex
	streams        map[string]*Stream.Stream
	streamsMtx     sync.RWMutex

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
	inst.connections = make(map[string]*NetConnection.NetConnection)
	inst.streams = make(map[string]*Stream.Stream)

	return &inst, nil
}

func (this *Instance) GetStats() *stats {
	return &this.stats
}

func (this *Instance) OnConnect(nc *NetConnection.NetConnection) {
	this.connectionsMtx.Lock()
	this.connections[nc.ID] = nc
	this.connectionsMtx.Unlock()
}

func (this *Instance) OnDisconnect(nc *NetConnection.NetConnection) {
	this.connectionsMtx.Lock()
	this.connections[nc.ID] = nil
	this.connectionsMtx.Unlock()
}
