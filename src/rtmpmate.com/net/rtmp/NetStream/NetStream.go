package NetStream

import (
	"rtmpmate.com/events/NetStatusEvent"
	"rtmpmate.com/net/rtmp/Interfaces"
	"rtmpmate.com/net/rtmp/Stream"
)

type NetStream struct {
	nc     *Interfaces.INetConnection
	Stream *Stream.Stream
}

func New(nc *Interfaces.INetConnection) (*NetStream, error) {
	var ns NetStream
	ns.nc = nc

	return &ns, nil
}

func (this *NetStream) Attach(stream *Stream.Stream) error {
	this.Stream = stream
	return nil
}

func (this *NetStream) Play(name string) error {
	return nil
}

func (this *NetStream) Pause() error {
	return nil
}

func (this *NetStream) Resume() error {
	return nil
}

func (this *NetStream) ReceiveAudio(flag bool) error {
	return nil
}

func (this *NetStream) ReceiveVideo(flag bool) error {
	return nil
}

func (this *NetStream) Seek(offset float64) error {
	return nil
}

func (this *NetStream) Publish(name string, t string) error {
	return nil
}

func (this *NetStream) Stop() error {
	return nil
}

func (this *NetStream) Dispose() error {
	return nil
}

func (this *NetStream) onStatus(e *NetStatusEvent.NetStatusEvent) {

}
