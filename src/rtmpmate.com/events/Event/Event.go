package Event

import (
	"fmt"
)

const (
	CANCEL   = "cancel"
	CHANGE   = "change"
	CLEAR    = "clear"
	CLOSE    = "close"
	COMPLETE = "complete"
	CONNECT  = "connect"
	RESIZE   = "resize"
)

type Event struct {
	Type   string
	Target interface{} // A reference to the object that dispatched the event.
}

func New(Type string, Target interface{}) *Event {
	return &Event{Type, Target}
}

func (this *Event) Clone() *Event {
	return &Event{this.Type, this}
}

func (this *Event) ToString() string {
	return fmt.Sprintf("[Event type=%s]", this.Type)
}
