package Instance

import (
	"fmt"
	"rtmpmate.com/events"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/ServerEvent"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Stream"
	"sync"
)

type Instance struct {
	Name string

	clients    map[string]Interfaces.INetConnection
	clientsMtx sync.RWMutex
	streams    map[string]Interfaces.IStream
	streamsMtx sync.RWMutex

	statsToAdmin
	events.EventDispatcher
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
	inst.clients = make(map[string]Interfaces.INetConnection)
	inst.streams = make(map[string]Interfaces.IStream)

	inst.AddEventListener(ServerEvent.CONNECT, inst.onConnect, 0)
	inst.AddEventListener(ServerEvent.DISCONNECT, inst.onDisconnect, 0)

	return &inst, nil
}

func (this *Instance) GetStream(name string, start float64) (Interfaces.IStream, error) {
	var err error = nil

	this.streamsMtx.Lock()

	stream, ok := this.streams[name]
	if ok == false {
		if start == -2 {
			stream, err = Stream.New(name)
			if stream != nil {
				this.streams[name] = stream
			}
		}
	}

	if stream == nil {
		err = fmt.Errorf("stream (name=%s) not found", name)
	}

	this.streamsMtx.Unlock()

	return stream, err
}

func (this *Instance) Unload() {
	this.clientsMtx.Lock()

	for _, nc := range this.clients {
		this.RemoveEventListener(ServerEvent.CONNECT, this.onConnect)
		this.RemoveEventListener(ServerEvent.DISCONNECT, this.onDisconnect)
		nc.Close()

		delete(this.clients, nc.GetFarID())
	}

	this.clientsMtx.Unlock()
}

func (this *Instance) onConnect(e *ServerEvent.ServerEvent) {
	this.clientsMtx.Lock()

	farID := e.Client.GetFarID()
	this.clients[farID] = e.Client

	this.connected++

	this.clientsMtx.Unlock()
}

func (this *Instance) onDisconnect(e *ServerEvent.ServerEvent) {
	this.clientsMtx.Lock()

	farID := e.Client.GetFarID()
	delete(this.clients, farID)

	this.connected--

	this.clientsMtx.Unlock()
}

func (this *Instance) onGetStats(e *CommandEvent.CommandEvent) {

}
