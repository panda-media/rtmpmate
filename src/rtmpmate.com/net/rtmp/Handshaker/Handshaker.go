package Handshaker

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"rtmpmate.com/net/rtmp"
	"rtmpmate.com/net/rtmp/Client"
	"rtmpmate.com/net/rtmp/Handshaker/Types"
	"syscall"
)

const (
	PACKET_SIZE = 1536
	DIGEST_SIZE = 32
	VERSION     = 0x5033029
)

type Handshaker struct {
	Client *Client.Client
	mode   uint8
}

func New(conn *net.TCPConn) (*Handshaker, error) {
	if conn == nil {
		return nil, syscall.EINVAL
	}

	client, err := Client.New(conn)
	if err != nil {
		fmt.Printf("Failed to create client: %v.\n", err)
		return nil, err
	}

	var shaker Handshaker
	shaker.Client = client
	shaker.mode = Types.SIMPLE

	return &shaker, nil
}

func (this *Handshaker) Shake() error {
	data, err := this.Client.Read(1+PACKET_SIZE, false)
	if err != nil {
		return err
	}

	if data[0] != 0x03 {
		return fmt.Errorf("Invalid handshake version: %d.\n", data[0])
	}

	zero := binary.BigEndian.Uint32(data[5:9])
	if zero == 0 {
		err = this.simpleHandshake(data[1:])
	} else {
		err = this.complexHandshake(data[1:])
	}

	if err != nil {
		return err
	}

	// Handshake done
	fmt.Printf("Handshake done: id=%s.\n", this.Client.ID)

	return nil
}

func (this *Handshaker) simpleHandshake(c1 []byte) error {
	// S0 + S1
	s01 := make([]byte, 1+PACKET_SIZE)

	s01[0] = 0x03
	binary.BigEndian.PutUint32(s01[1:5], 0)
	binary.BigEndian.PutUint32(s01[5:9], 0)

	for i := 9; i <= PACKET_SIZE; i++ {
		s01[i] = byte(rand.Int() % 256)
	}

	n, err := this.Client.Write(s01)
	if err != nil {
		return err
	}

	if n != 1+PACKET_SIZE {
		return fmt.Errorf("Uncomplete to send S0 + S1: %d.\n", n)
	}

	// s2
	n, err = this.Client.Write(c1)
	if err != nil {
		return err
	}

	if n != PACKET_SIZE {
		return fmt.Errorf("Uncomplete to send S2: %d.\n", n)
	}

	// C2
	c2, err := this.Client.Read(PACKET_SIZE, false)
	if err != nil {
		return err
	}

	for i := 0; i < PACKET_SIZE; i++ {
		if c2[i] != s01[i+1] {
			return fmt.Errorf("Packet C2 & S1 not match.\n")
		}
	}

	return nil
}

func (this *Handshaker) complexHandshake(c1 []byte) error {
	return nil
}

func (this *Handshaker) getComplexS1() []byte {
	s1 := make([]byte, PACKET_SIZE)
	binary.BigEndian.PutUint32(s1[0:4], 0)
	binary.BigEndian.PutUint32(s1[4:8], VERSION)

	for i := 8; i < PACKET_SIZE; i++ {
		s1[i] = byte(rand.Int() % 256)
	}

	return s1
}

func (this *Handshaker) getComplexS2() []byte {
	size := PACKET_SIZE - DIGEST_SIZE
	s2 := make([]byte, size)

	for i := 0; i < size; i++ {
		s2[i] = byte(rand.Int() % 256)
	}

	return s2
}

func (this *Handshaker) checkComplexC1(c1 []byte) (int, []byte, []byte) {
	challenge, digest, err := this.checkComplexC1Scheme(c1, 1)
	if err == nil {
		return 0, challenge, digest
	}

	fmt.Printf("Failed to checkComplexC1Scheme: %v. Keep trying...", err)

	challenge, digest, err = this.checkComplexC1Scheme(c1, 0)
	if err == nil {
		return 1, challenge, digest
	}

	fmt.Printf("Failed to checkComplexC1Scheme: %v.", err)

	return -1, challenge, digest
}

func (this *Handshaker) checkComplexC1Scheme(c1 []byte, scheme int) ([]byte, []byte, error) {
	digestOffset := this.getDigestOffset(c1, scheme)
	challengeOffset := this.getDHOffset(c1, scheme)

	if digestOffset == -1 || challengeOffset == -1 {
		return nil, nil, syscall.EINVAL
	}

	size := PACKET_SIZE - DIGEST_SIZE
	data := make([]byte, size)
	digest := make([]byte, DIGEST_SIZE)

	copy(digest, c1[digestOffset:DIGEST_SIZE+digestOffset])
	if digestOffset != 0 {
		copy(data, c1[:digestOffset])
		copy(data[digestOffset:], c1[DIGEST_SIZE+digestOffset:])
	} else {
		copy(data, c1)
	}

	hash := hmac.New(sha256.New, rtmp.FP_KEY[:30])
	hash.Write(data)

	tmp := hash.Sum(nil)
	if bytes.Compare(tmp, digest) != 0 {
		return nil, nil, syscall.EINVAL
	}

	challenge := make([]byte, 128)
	copy(challenge, c1[challengeOffset:challengeOffset+128])

	return challenge, digest, nil
}

func (this *Handshaker) getDigestOffset(c1 []byte, scheme int) int {
	var offset int

	if scheme == 0 {
		offset = int(c1[8]) + int(c1[9]) + int(c1[10]) + int(c1[11])
		offset = (offset % 728) + 8 + 4

		if offset+32 > 1536 {
			return -1
		}

		return offset
	} else if scheme == 1 {
		offset = int(c1[772]) + int(c1[773]) + int(c1[774]) + int(c1[775])
		offset = (offset % 728) + 772 + 4

		if offset+32 > 1536 {
			return -1
		}

		return offset
	}

	return -1
}

func (this *Handshaker) getDHOffset(c1 []byte, scheme int) int {
	var offset int

	if scheme == 0 {
		offset = int(c1[1532]) + int(c1[1533]) + int(c1[1534]) + int(c1[1535])
		offset = (offset % 632) + 772

		if offset+128 > 1536 {
			return -1
		}

		return offset
	} else if scheme == 1 {
		offset = int(c1[768]) + int(c1[769]) + int(c1[770]) + int(c1[771])
		offset = (offset % 632) + 8

		if offset+128 > 1536 {
			return -1
		}

		return offset
	}

	return -1
}
