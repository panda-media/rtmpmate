package RTMPListener

import (
	"fmt"
	"net"
	"rtmpmate.com/net/rtmp/Handshaker"
	"strconv"
)

type RTMPListener struct {
	tcpln   *net.TCPListener
	network string
	port    int
	exiting bool
}

func New() (*RTMPListener, error) {
	var ln RTMPListener
	return &ln, nil
}

func (this *RTMPListener) Listen(network string, port int) {
	if network == "" {
		network = "tcp4"
	}

	if port == 0 {
		port = 1935
	}

	address := strconv.Itoa(port)

	tcpaddr, err := net.ResolveTCPAddr(network, ":"+address)
	if err != nil {
		fmt.Printf("Failed to ResolveTCPAddr: %v.\n", err)
		return
	}

	tcpln, err := net.ListenTCP(network, tcpaddr)
	if err != nil {
		fmt.Printf("Failed to listen on port %d: %v.\n", port, err)
		return
	}

	this.tcpln = tcpln

	for this.exiting == false {
		conn, err := tcpln.AcceptTCP()
		if err != nil {
			fmt.Printf("Failed to accept RTMP connection: %v.\n", err)
			continue
		}

		go this.connHandler(conn)
	}

	fmt.Printf("%v exiting...\n", this)
}

func (this *RTMPListener) connHandler(conn *net.TCPConn) {
	shaker, err := Handshaker.New(conn)
	if err != nil {
		fmt.Printf("Failed to create Handshaker: %v.\n", err)
		return
	}

	err = shaker.Shake()
	if err != nil {
		fmt.Printf("Failed to complete handshake: %v.\n", err)
	} else {
		err = shaker.Conn.WaitRequest()
		if err != nil {
			fmt.Printf("Failed to wait request: %v.\n", err)
		}
	}

	err = shaker.Conn.Close()
	if err != nil {
		fmt.Printf("Failed to close client: %v.\n", err)
	}

	fmt.Printf("Closed client: id=%s.\n", shaker.Conn.ID)
}
