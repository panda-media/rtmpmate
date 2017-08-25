package NetConnection

import (
	"net"
	"rtmpmate.com/events"
	"rtmpmate.com/net/rtmp/Client"
	"syscall"
)

type NetConnection struct {
	Client *Client.Client
	conn   *net.TCPConn
	events.EventDispatcher
}

func New(conn *net.TCPConn) (*NetConnection, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	var nc NetConnection
	nc.conn = conn

	return &nc, nil
}

func (this *NetConnection) Call(methodName string, resp *Client.Responder, args []interface{}) bool {
	return true
}
