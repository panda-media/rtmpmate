package Command

import (
	"rtmpmate.com/util/AMF"
)

const (
	CONNECT       = "connect"
	CLOSE         = "close"
	CREATE_STREAM = "createStream"
	_RESULT       = "_result"
	_ERROR        = "_error"

	PLAY          = "play"
	PLAY2         = "play2"
	DELETE_STREAM = "deleteStream"
	CLOSE_STREAM  = "closeStream"
	RECEIVE_AUDIO = "receiveAudio"
	RECEIVE_VIDEO = "receiveVideo"
	PUBLISH       = "publish"
	SEEK          = "seek"
	PAUSE         = "pause"
	ON_STATUS     = "onStatus"
)

type Command struct {
	Name          AMF.AMFString
	TransactionID float64
	fields        []AMF.AMFValue
}

func New(name string, id float64) (*Command, error) {
	var cmd Command
	cmd.Name = AMF.AMFString{name, 0}
	cmd.TransactionID = id
	cmd.fields = make([]AMF.AMFValue, 0, 5)

	return &cmd, nil
}

func (this *Command) Append(AMFType byte, data interface{}) error {
	var field AMF.AMFValue
	field.Type = AMFType
	field.Data = data

	this.fields = append(this.fields, field)

	return nil
}
