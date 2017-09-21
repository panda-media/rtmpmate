package rtmp

import (
	"container/list"
	"rtmpmate.com/util/AMF"
	AMFTypes "rtmpmate.com/util/AMF/Types"
)

const (
	APPLICATIONS = "applications/"
)

var (
	FMSProperties AMF.AMFObject
	FMSVersion    list.List
)

func init() {
	FMSProperties.Init()
	FMSProperties.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "fmsVer",
		Data: "FMS/5,0,3,3029",
	})
	FMSProperties.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.DOUBLE,
		Key:  "capabilities",
		Data: float64(255),
	})
	FMSProperties.Data.PushBack(&AMF.AMFValue{
		Type: AMFTypes.DOUBLE,
		Key:  "mode",
		Data: float64(1),
	})

	FMSVersion.PushBack(&AMF.AMFValue{
		Type: AMFTypes.STRING,
		Key:  "version",
		Data: "5,0,3,3029",
	})
}
