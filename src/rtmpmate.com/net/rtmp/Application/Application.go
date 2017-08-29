package Application

import (
	"fmt"
	"rtmpmate.com/net/rtmp/Application/Instance"
	"rtmpmate.com/net/rtmp/Client"
	"rtmpmate.com/net/rtmp/Stream"
	"rtmpmate.com/util/AMF"
	"sync"
	"syscall"
)

type Application struct {
	Name         string
	instances    map[string]*Instance.Instance
	instancesMtx sync.RWMutex

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

var apps map[string]*Application
var appsMtx sync.RWMutex

func init() {
	apps = make(map[string]*Application)
}

func New(name string) (*Application, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	var app Application
	app.Name = name
	app.instances = make(map[string]*Instance.Instance)

	return &app, nil
}

func Get(name string) (*Application, error) {
	if name == "" {
		return nil, syscall.EINVAL
	}

	appsMtx.Lock()

	var app, ok = apps[name]
	if ok == false {
		app, err := New(name)
		if err != nil {
			fmt.Printf("Failed to create application \"%s\": %v.\n", name, err)
			return nil, err
		}

		apps[name] = app
	}

	appsMtx.Unlock()

	return app, nil
}

func (this *Application) GetInstance(name string) (*Instance.Instance, error) {
	if name == "" {
		name = "_definst_"
	}

	this.instancesMtx.Lock()

	var inst, ok = this.instances[name]
	if ok == false {
		inst, err := Instance.New(name)
		if err != nil {
			fmt.Printf("Failed to create instance \"%s\" of application \"%s\": %v.\n", this.Name, name, err)
			return nil, err
		}

		this.instances[name] = inst
	}

	this.instancesMtx.Unlock()

	return inst, nil
}

func (this *Application) AcceptConnection(client *Client.Client) {
	inst, err := this.GetInstance(client.Instance)
	if err != nil {
		fmt.Printf("Failed to get instance \"%s\" of application \"%s\": %v.\n",
			client.Application, client.Instance, err)
		return
	}

	inst.OnConnect(client)

	client.Call("onStatus", nil, nil)
}

func (this *Application) RejectConnection(client *Client.Client, description string, errObj *AMF.AMFObject) {

}

func (this *Application) RedirectConnection(client *Client.Client, url string, description string, errObj *AMF.AMFObject) {

}

func (this *Application) GetStats() *stats {
	return &this.stats
}

func (this *Application) Disconnect(client *Client.Client) {

}

func (this *Application) GC() {

}

func (this *Application) Shutdown() bool {
	return true
}

func OnStart() {

}

func OnConnect(client *Client.Client, args []interface{}) {
	app, err := Get(client.Application)
	if err != nil {
		fmt.Printf("Failed to get application \"%s\": %v.\n", client.Application, err)
		return
	}

	app.AcceptConnection(client)
}

func OnPublish(client *Client.Client, stream *Stream.Stream) {

}

func OnUnpublish(client *Client.Client, stream *Stream.Stream) {

}

func OnDisconnect(client *Client.Client) {

}

func OnStop() {

}
