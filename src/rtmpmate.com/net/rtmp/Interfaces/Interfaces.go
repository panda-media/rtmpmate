package Interfaces

import (
	"io"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Responder"
	"rtmpmate.com/util/AMF"
)

type IEventDispatcher interface {
	AddEventListener(event string, handler interface{}, count int)
	RemoveEventListener(event string, handler interface{})
	HasEventListener(event string) bool
	DispatchEvent(event interface{})
}

type INetConnection interface {
	io.ReadWriteCloser

	Connect(uri string, args ...*AMF.AMFValue) error
	CreateStream() error
	Call(method string, res *Responder.Responder, args ...*AMF.AMFValue) error
	WriteByChunk(b []byte, csid int, h *Message.Header) (int, error)

	GetAppName() string
	GetInstName() string
	GetFarID() string

	IEventDispatcher
}

type IStream interface {
	io.Closer

	Play(name string, start float64, length float64, reset bool) error
	Record(mode string, maxDuration int, maxSize int) error
	Send(handler string, args ...*AMF.AMFValue) error
	Stop() error
	Clear() error

	IEventDispatcher
}

type INetStream interface {
	io.Closer

	Attach(stream *IStream) error
	Play(name string) error
	Pause() error
	Resume() error
	ReceiveAudio(flag bool) error
	ReceiveVideo(flag bool) error
	Seek(offset float64) error
	Send(handler string, args ...*AMF.AMFValue) error
	Publish(name string, t string) error
	Stop() error
	Dispose() error

	IEventDispatcher
}
