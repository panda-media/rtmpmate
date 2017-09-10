package NetStream

import (
	"fmt"
	"math"
	"rtmpmate.com/events"
	"rtmpmate.com/events/AudioEvent"
	"rtmpmate.com/events/CommandEvent"
	"rtmpmate.com/events/DataFrameEvent"
	"rtmpmate.com/events/NetStatusEvent"
	"rtmpmate.com/events/NetStatusEvent/Code"
	"rtmpmate.com/events/NetStatusEvent/Level"
	"rtmpmate.com/events/VideoEvent"
	"rtmpmate.com/net/rtmp/Chunk/CSIDs"
	"rtmpmate.com/net/rtmp/Message"
	"rtmpmate.com/net/rtmp/Message/CommandMessage/Commands"
	"rtmpmate.com/net/rtmp/Message/Types"
	"rtmpmate.com/net/rtmp/NetConnection"
	"rtmpmate.com/net/rtmp/Stream"
	StreamTypes "rtmpmate.com/net/rtmp/Stream/Types"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
)

type NetStream struct {
	nc     *NetConnection.NetConnection
	stream *Stream.Stream

	events.EventDispatcher
}

func New(nc *NetConnection.NetConnection) (*NetStream, error) {
	var ns NetStream
	ns.nc = nc

	nc.AddEventListener(CommandEvent.CREATE_STREAM, ns.onCreateStream, 0)
	nc.AddEventListener(CommandEvent.PLAY, ns.onPlay, 0)
	nc.AddEventListener(CommandEvent.PLAY2, ns.onPlay2, 0)
	nc.AddEventListener(CommandEvent.DELETE_STREAM, ns.onDeleteStream, 0)
	nc.AddEventListener(CommandEvent.CLOSE_STREAM, ns.onCloseStream, 0)
	nc.AddEventListener(CommandEvent.RECEIVE_AUDIO, ns.onReceiveAV, 0)
	nc.AddEventListener(CommandEvent.RECEIVE_VIDEO, ns.onReceiveAV, 0)
	nc.AddEventListener(CommandEvent.PUBLISH, ns.onPublish, 0)
	nc.AddEventListener(CommandEvent.SEEK, ns.onSeek, 0)
	nc.AddEventListener(CommandEvent.PAUSE, ns.onPause, 0)

	nc.AddEventListener(DataFrameEvent.SET_DATA_FRAME, ns.onSetDataFrame, 0)
	nc.AddEventListener(DataFrameEvent.CLEAR_DATA_FRAME, ns.onClearDataFrame, 0)
	nc.AddEventListener(AudioEvent.DATA, ns.onAudio, 0)
	nc.AddEventListener(VideoEvent.DATA, ns.onVideo, 0)

	return &ns, nil
}

func (this *NetStream) Attach(src *Stream.Stream) error {
	this.stream = src
	//this.stream.AddEventListener("onMetaData", this.onMetaData, 0)
	this.stream.AddEventListener(DataFrameEvent.SET_DATA_FRAME, this.sendDataFrame, 0)
	this.stream.AddEventListener(DataFrameEvent.CLEAR_DATA_FRAME, this.clearDataFrame, 0)
	this.stream.AddEventListener(AudioEvent.DATA, this.sendAudio, 0)
	this.stream.AddEventListener(VideoEvent.DATA, this.sendVideo, 0)

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

func (this *NetStream) Send(handler string, args ...*AMF.AMFValue) error {
	var encoder AMF.Encoder
	encoder.EncodeString(handler)

	for _, v := range args {
		encoder.EncodeValue(v)
	}

	b, err := encoder.Encode()
	if err != nil {
		return err
	}

	var h Message.Header
	if this.nc.ObjectEncoding == AMF.AMF0 {
		h.Type = Types.DATA
	} else {
		h.Type = Types.AMF3_DATA
	}
	h.Length = encoder.Len()
	h.Timestamp = 0
	h.StreamID = uint32(this.stream.ID)

	_, err = this.nc.WriteByChunk(b, CSIDs.COMMAND_2, &h)
	if err != nil {
		return err
	}

	return nil
}

func (this *NetStream) sendDataFrame(e *DataFrameEvent.DataFrameEvent) error {
	fmt.Printf("Sending %s...\n", e.Type)

	return this.Send(e.Type, &AMF.AMFValue{
		Type: AMFTypes.STRING,
		Data: e.Key,
	}, &AMF.AMFValue{
		Type: AMFTypes.OBJECT,
		Data: e.Data.Data,
	})
}

func (this *NetStream) clearDataFrame(e *DataFrameEvent.DataFrameEvent) error {
	return this.Send(e.Type, &AMF.AMFValue{
		Type: AMFTypes.STRING,
		Data: e.Key,
	})
}

func (this *NetStream) sendAudio(e *AudioEvent.AudioEvent) error {
	_, err := this.nc.Write(e.Data.Payload)
	return err
}

func (this *NetStream) sendVideo(e *VideoEvent.VideoEvent) error {
	_, err := this.nc.Write(e.Data.Payload)
	return err
}

func (this *NetStream) Stop() error {
	return nil
}

func (this *NetStream) Dispose() error {
	return nil
}

func (this *NetStream) onCreateStream(e *CommandEvent.CommandEvent) {
	var command, code, description string

	if this.nc.ReadAccess == "/" || this.nc.ReadAccess == "/"+this.nc.AppName {
		this.stream, _ = Stream.New(this.nc.FarID)
		if this.stream != nil {
			this.stream.ID = 1 // ID 0 is used as NetConnection
			this.stream.Type = StreamTypes.IDLE
			this.Attach(this.stream)

			command = Commands.RESULT
		} else {
			command = Commands.ERROR
			code = Code.NETSTREAM_FAILED
			description = "Internal error"
		}
	} else {
		// TODO: Test on AMS
		command = Commands.ERROR
		code = Code.NETSTREAM_PLAY_FAILED
		description = "No read access"
	}

	e.Encoder.EncodeString(command)
	e.Encoder.EncodeNumber(math.Float64frombits(e.Message.TransactionID))
	e.Encoder.EncodeNull()

	if command == Commands.RESULT {
		e.Encoder.EncodeNumber(float64(this.stream.ID))
		return
	}

	// TODO: Test on AMS
	info, _ := this.nc.GetInfoObject(Level.ERROR, code, description)
	e.Encoder.EncodeObject(info)
}

func (this *NetStream) onPlay(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if this.nc.ReadAccess == "/" || this.nc.ReadAccess == "/"+this.nc.AppName {
		stream, _ := this.nc.App.GetStream(this.nc.InstName, e.Message.StreamName, e.Message.Start)
		if stream != nil {
			this.stream.Name = e.Message.StreamName
			this.stream.Type = StreamTypes.PLAYING_LIVE
			this.stream.Source(stream.(*Stream.Stream))

			if e.Message.Reset {
				info, _ := this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_RESET, "Play reset")

				e.Encoder.EncodeString(Commands.ON_STATUS)
				e.Encoder.EncodeNumber(0)
				e.Encoder.EncodeNull()
				e.Encoder.EncodeObject(info)

				e.Message.Length = e.Encoder.Len()

				b, _ := e.Encoder.Encode()
				this.nc.WriteByChunk(b, CSIDs.COMMAND, &e.Message.Header)

				e.Encoder.Reset()
			}

			info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_START, "Play start")
		} else {
			info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PLAY_STREAMNOTFOUND, "Stream not found")
		}
	} else {
		info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PLAY_FAILED, "No read access")
	}

	e.Encoder.EncodeString(Commands.ON_STATUS)
	e.Encoder.EncodeNumber(0)
	e.Encoder.EncodeNull()
	e.Encoder.EncodeObject(info)
}

func (this *NetStream) onPlay2(e *CommandEvent.CommandEvent) {

}

func (this *NetStream) onDeleteStream(e *CommandEvent.CommandEvent) {
	this.stream = nil
}

func (this *NetStream) onCloseStream(e *CommandEvent.CommandEvent) {
	this.stream = nil
}

func (this *NetStream) onReceiveAV(e *CommandEvent.CommandEvent) {
	if e.Message.Name == CommandEvent.RECEIVE_AUDIO && this.stream.ReceiveAudio ||
		e.Message.Name == CommandEvent.RECEIVE_VIDEO && this.stream.ReceiveVideo {
		return
	}

	if e.Message.Flag {
		info, _ := this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_SEEK_NOTIFY, "Seek notify")

		e.Encoder.EncodeString(Commands.ON_STATUS)
		e.Encoder.EncodeNumber(0)
		e.Encoder.EncodeNull()
		e.Encoder.EncodeObject(info)

		e.Message.Length = e.Encoder.Len()
		b, _ := e.Encoder.Encode()
		this.nc.WriteByChunk(b, CSIDs.COMMAND, &e.Message.Header)

		e.Encoder.Reset()

		info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PLAY_START, "Play start")

		e.Encoder.EncodeString(Commands.ON_STATUS)
		e.Encoder.EncodeNumber(0)
		e.Encoder.EncodeNull()
		e.Encoder.EncodeObject(info)
	}
}

func (this *NetStream) onPublish(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if this.nc.WriteAccess == "/" || this.nc.WriteAccess == "/"+this.nc.AppName {
		stream, _ := this.nc.App.GetStream(this.nc.InstName, e.Message.PublishingName, -2)
		if stream == nil {
			info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_FAILED, "Internal error")
		} else if stream.(*Stream.Stream).Type == StreamTypes.PUBLISHING {
			info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_PUBLISH_BADNAME, "Publish bad name")
		} else {
			this.stream.Type = StreamTypes.PUBLISHING
			this.stream.Sink(stream.(*Stream.Stream))

			info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PUBLISH_START, "Publish start")
		}
	} else {
		// TODO: Test on AMS
		info, _ = this.nc.GetInfoObject(Level.ERROR, "No write access", "No write access")
	}

	e.Encoder.EncodeString(Commands.ON_STATUS)
	e.Encoder.EncodeNumber(0)
	e.Encoder.EncodeNull()
	e.Encoder.EncodeObject(info)
}

func (this *NetStream) onSeek(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if this.stream.Type == StreamTypes.PLAYING_VOD {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.stream.Duration {
			info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_SEEK_NOTIFY, "Seek notify")
		} else {
			info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_SEEK_INVALIDTIME, "Seek invalid time")
		}
	} else {
		info, _ = this.nc.GetInfoObject(Level.ERROR, Code.NETSTREAM_SEEK_FAILED, "Seek failed")
	}

	e.Encoder.EncodeString(Commands.ON_STATUS)
	e.Encoder.EncodeNumber(0)
	e.Encoder.EncodeNull()
	e.Encoder.EncodeObject(info)
}

func (this *NetStream) onPause(e *CommandEvent.CommandEvent) {
	var info *AMF.AMFObject

	if e.Message.Pause {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.stream.Duration {
			this.stream.Pause = e.Message.Pause
			this.stream.CurrentTime = e.Message.MilliSeconds

			info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_PAUSE_NOTIFY, "Pause notify")
		} else {
			info, _ = this.nc.GetInfoObject(Level.ERROR, "Pause invalid time", "Pause invalid time")
		}
	} else {
		if e.Message.MilliSeconds >= 0 && e.Message.MilliSeconds <= this.stream.Duration {
			this.stream.Pause = e.Message.Pause
			this.stream.CurrentTime = e.Message.MilliSeconds

			info, _ = this.nc.GetInfoObject(Level.STATUS, Code.NETSTREAM_UNPAUSE_NOTIFY, "Unpause notify")
		} else {
			info, _ = this.nc.GetInfoObject(Level.ERROR, "Unpause invalid time", "Unpause invalid time")
		}
	}

	e.Encoder.EncodeString(Commands.ON_STATUS)
	e.Encoder.EncodeNumber(0)
	e.Encoder.EncodeNull()
	e.Encoder.EncodeObject(info)
}

func (this *NetStream) onStatus(e *NetStatusEvent.NetStatusEvent) {

}

func (this *NetStream) onSetDataFrame(e *DataFrameEvent.DataFrameEvent) {
	this.stream.DispatchEvent(DataFrameEvent.New(e.Type, this, e.Key, e.Data))
}

func (this *NetStream) onClearDataFrame(e *DataFrameEvent.DataFrameEvent) {
	this.stream.DispatchEvent(DataFrameEvent.New(e.Type, this, e.Key, e.Data))
}

func (this *NetStream) onAudio(e *AudioEvent.AudioEvent) {
	this.stream.DispatchEvent(AudioEvent.New(e.Type, this, e.Data))
}

func (this *NetStream) onVideo(e *VideoEvent.VideoEvent) {
	this.stream.DispatchEvent(VideoEvent.New(e.Type, this, e.Data))
}

func (this *NetStream) onMetaData(e *DataFrameEvent.DataFrameEvent) {
	//fmt.Printf("%s: %s\n", e.Key, e.Data.ToString(0))
}
