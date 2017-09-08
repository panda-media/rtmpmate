package Instance

import (
	"sync"
	"syscall"
)

type Instance struct {
	Name string

	connections    map[string]interface{}
	connectionsMtx sync.RWMutex
	streams        map[string]interface{}
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

	launchTime uint64 // time the application started.
	upTime     uint64 // time, in seconds, the application has been running.

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
	inst.connections = make(map[string]interface{})
	inst.streams = make(map[string]interface{})

	return &inst, nil
}

func (this *Instance) GetConnection(id string) (interface{}, error) {
	if id == "" {
		return nil, syscall.EINVAL
	}

	this.connectionsMtx.Lock()
	nc, _ := this.connections[id]
	this.connectionsMtx.Unlock()

	return nc, nil
}

func (this *Instance) GetStream(name string, start float64) (interface{}, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	this.streamsMtx.Lock()
	s, _ := this.streams[name]
	this.streamsMtx.Unlock()

	return s, nil
}

func (this *Instance) AddConnection(id string, nc interface{}) {
	this.connectionsMtx.Lock()
	this.connections[id] = nc
	this.connectionsMtx.Unlock()
}

func (this *Instance) RemoveConnection(id string) {
	this.connectionsMtx.Lock()
	this.connections[id] = nil
	this.connectionsMtx.Unlock()
}

func (this *Instance) AddStream(name string, s interface{}) {
	this.streamsMtx.Lock()
	this.streams[name] = s
	this.streamsMtx.Unlock()
}

func (this *Instance) RemoveStream(name string) {
	this.streamsMtx.Lock()
	this.streams[name] = nil
	this.streamsMtx.Unlock()
}

func (this *Instance) GetStats() *stats {
	return &this.stats
}
