package DataFrameEvent

import (
	"fmt"
	"rtmpmate.com/events/Event"
	"rtmpmate.com/util/AMF"
)

const (
	SET_DATA_FRAME   = "@setDataFrame"
	CLEAR_DATA_FRAME = "@clearDataFrame"
)

type DataFrameEvent struct {
	Event.Event
	Key  string
	Data *AMF.AMFObject
}

func New(Type string, Target interface{}, key string, data *AMF.AMFObject) *DataFrameEvent {
	return &DataFrameEvent{Event.Event{Type, Target}, key, data}
}

func (this *DataFrameEvent) Clone() *DataFrameEvent {
	return &DataFrameEvent{Event.Event{this.Type, this}, this.Key, this.Data}
}

func (this *DataFrameEvent) ToString() string {
	return fmt.Sprintf("[DataFrameEvent type=%s key=%s data=%v]", this.Type, this.Key, this.Data)
}
