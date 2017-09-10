package rtmp

import (
	"fmt"
	"net"
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/NetStream"
)

type RTMP struct {
	nc *NetConnection.NetConnection
	ns *NetStream.NetStream
}

func New(conn *net.TCPConn) (*RTMP, error) {
	nc, err := NetConnection.New(conn)
	if err != nil {
		fmt.Printf("Failed to create NetConnection: %v.\n", err)
		return nil, err
	}

	ns, err := NetStream.New(nc)
	if err != nil {
		fmt.Printf("Failed to create NetStream: %v.\n", err)
		return nil, err
	}

	var r RTMP
	r.nc = nc
	r.ns = ns

	return &r, nil
}

func (this *RTMP) WaitRequest() error {
	return this.nc.WaitRequest()
}
