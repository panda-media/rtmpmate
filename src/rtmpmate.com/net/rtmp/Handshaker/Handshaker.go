package Handshaker

import (
	"fmt"
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/net/rtmp/Client"
	"rtmpmate.com/net/rtmp/Handshaker/Types"
	"syscall"
)

type Handshaker struct {
	Client *Client.Client
	mode   uint8

	events.EventDispatcher
}

func New(conn *net.TCPConn) (*Handshaker, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	client, err := Client.New(conn, "123")
	if err != nil {
		fmt.Printf("Failed to create client: %v.\n", err)
		return nil, err
	}

	var shaker Handshaker
	shaker.Client = client
	shaker.mode = Types.SIMPLE

	return &shaker, nil
}

func (this *Handshaker) Shake() {
	this.Client.ApplicationName = "live"
	this.Client.InstanceName = "_definst_"

	// Handshake done
	this.DispatchEvent(Event.New(Event.COMPLETE, this))
}
