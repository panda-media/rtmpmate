package Application

import (
	"fmt"
	"rtmpmate.com/events"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/ServerEvent"
	"rtmpmate.com/net/rtmp/Application/Instance"
	"rtmpmate.com/net/rtmp/Interfaces"
	"sync"
	"syscall"
)

var (
	apps    map[string]*Application
	appsMtx sync.RWMutex
)

func init() {
	apps = make(map[string]*Application)
}

type Application struct {
	Name string

	instances    map[string]*Instance.Instance
	instancesMtx sync.RWMutex

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

	// Total number of instances that have been loaded since the application started.
	// This property does not represent the total number of active instances loaded. To get the number of active
	// instances loaded, subtract the value of total_instances_unloaded from total_instances_loaded.
	totalInstancesLoaded   int
	totalInstancesUnloaded int // Total number of instances that have been unloaded since the application started.

	totalConnects    int // Total number of socket connections to the application since the application was started.
	totalDisconnects int // Total number of disconnections from the application since the application was started.

	accepted        int // Total number of connection attempts accepted by this application.
	rejected        int // Total number of connection attempts rejected by this application.
	connected       int // Total number of connections currently active.
	normalConnects  int // Total number of normal connections.
	virtualConnects int // Total number of connections through a remote edge.
	adminConnects   int // Total number of administrator connections.
	debugConnects   int // Total number of debug connections.

	swfVerificationAttempts int
	swfVerificationMatches  int
	swfVerificationFailures int
}

func Get(name string) (*Application, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	appsMtx.Lock()

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Failed to DispatchEvent: %v.\n", err)
			appsMtx.Unlock()
		}
	}()

	app, ok := apps[name]
	if ok {
		appsMtx.Unlock()
		return app, nil
	}

	var newApp Application
	newApp.Name = name
	newApp.instances = make(map[string]*Instance.Instance)

	newApp.AddEventListener(ServerEvent.CONNECT, newApp.onConnect, 0)
	newApp.AddEventListener(ServerEvent.PUBLISH, newApp.onPublish, 0)
	newApp.AddEventListener(ServerEvent.UNPUBLISH, newApp.onUnpublish, 0)
	newApp.AddEventListener(ServerEvent.DISCONNECT, newApp.onDisconnect, 0)
	newApp.AddEventListener(CommandEvent.GET_STATS, newApp.onGetStats, 0)

	apps[name] = &newApp
	appsMtx.Unlock()

	newApp.onStart()

	return &newApp, nil
}

func (this *Application) Accept(nc *Interfaces.INetConnection) error {
	return nil
}

func (this *Application) Reject(nc *Interfaces.INetConnection) error {
	return nil
}

func (this *Application) Disconnect(nc *Interfaces.INetConnection) error {
	return nil
}

func (this *Application) GetStream(instName string, streamName string, start float64) (Interfaces.IStream, error) {
	var stream Interfaces.IStream
	var err error

	this.instancesMtx.Lock()

	inst, ok := this.instances[instName]
	if ok {
		stream, err = inst.GetStream(streamName, start)
	} else {
		stream = nil
		err = fmt.Errorf("instance (name=%s) not found", instName)
	}

	this.instancesMtx.Unlock()

	return stream, err
}

func (this *Application) GC() {

}

func (this *Application) Shutdown() {
	this.instancesMtx.Lock()

	for _, inst := range this.instances {
		inst.Unload()
		delete(this.instances, inst.Name)
	}

	this.instancesMtx.Unlock()
}

func (this *Application) onStart() {

}

func (this *Application) onConnect(e *ServerEvent.ServerEvent) {
	fmt.Printf("Application.onConnect: id=%s.\n", e.Client.GetFarID())

	this.instancesMtx.Lock()

	instName := e.Client.GetInstName()
	inst, ok := this.instances[instName]
	if ok == false {
		inst, _ = Instance.New(instName)
		this.instances[instName] = inst
	}
	inst.DispatchEvent(ServerEvent.New(ServerEvent.CONNECT, inst, e.Client, nil))

	this.connected++

	this.instancesMtx.Unlock()
}

func (this *Application) onPublish(e *ServerEvent.ServerEvent) {

}

func (this *Application) onUnpublish(e *ServerEvent.ServerEvent) {

}

func (this *Application) onDisconnect(e *ServerEvent.ServerEvent) {
	fmt.Printf("Application.onDisconnect: id=%s.\n", e.Client.GetFarID())

	this.instancesMtx.Lock()

	instName := e.Client.GetInstName()
	inst, ok := this.instances[instName]
	if ok {
		inst.DispatchEvent(ServerEvent.New(ServerEvent.DISCONNECT, inst, e.Client, nil))
	}

	this.connected--

	this.instancesMtx.Unlock()
}

func (this *Application) onGetStats(e *CommandEvent.CommandEvent) {

}

func (this *Application) onStop() {

}
