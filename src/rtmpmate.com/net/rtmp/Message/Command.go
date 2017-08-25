package Message

import (
	"rtmpmate.com/util/AMF"
)

type Command struct {
	Name           AMF.AMFString
	TransactionID  float64
	Command        AMF.AMFObject
	Infomation     AMF.AMFObject
	StreamID       float64
	StreamName     AMF.AMFString
	Start          float64
	Duration       float64
	Reset          bool
	Flag           bool
	PublishingName AMF.AMFString
	PublishingType AMF.AMFString
	MilliSeconds   float64
}
