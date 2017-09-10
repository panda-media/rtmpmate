package events

import (
	"fmt"
	"reflect"
)

type listener struct {
	handler interface{}
	count   int
}

type EventDispatcher struct {
	listeners map[string][]listener
}

func (this *EventDispatcher) AddEventListener(event string, handler interface{}, count int) {
	if this.listeners == nil {
		this.listeners = make(map[string][]listener)
	}

	listeners := this.listeners[event]

	if listeners == nil {
		listeners = make([]listener, 0, 8)
	}

	listeners = append(listeners, listener{handler, count})
	this.listeners[event] = listeners

	//fmt.Printf("Added event: %s, len=%d, cap=%d.\n", event, len(listeners), cap(listeners))
}

func (this *EventDispatcher) RemoveEventListener(event string, handler interface{}) {
	listeners := this.listeners[event]

	if listeners == nil {
		return
	}

	if handler == nil {
		listeners = listeners[0:0]
		this.listeners[event] = listeners
		return
	}

	for i, listener := range listeners {
		p1 := reflect.ValueOf(listener.handler).Pointer()
		p2 := reflect.ValueOf(handler).Pointer()
		if p1 == p2 {
			copy(listeners[i:], listeners[i+1:])
			this.listeners[event] = listeners[:len(listeners)-1]
			break
		}
	}

	fmt.Printf("Removed event: %s, len: %d, cap: %d.\n", event, len(listeners), cap(listeners))
}

func (this *EventDispatcher) HasEventListener(event string) bool {
	listeners, _ := this.listeners[event]

	if listeners == nil || len(listeners) == 0 {
		return false
	}

	return true
}

func (this *EventDispatcher) DispatchEvent(event interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Failed to DispatchEvent: %v.\n", err)
		}
	}()

	var eventValue = reflect.ValueOf(event)
	var eventElem = eventValue.Elem()
	var eventType = eventElem.FieldByName("Type").String()

	listeners := this.listeners[eventType]
	if listeners == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Failed to handle %s: %v.\n", eventElem.MethodByName("ToString").Call(nil), err)
		}
	}()

	var eventIn = []reflect.Value{eventValue}

	for i, listener := range listeners {
		if listener.count > 0 {
			listener.count--

			if listener.count == 0 {
				copy(listeners[i:], listeners[i+1:])
				listeners = listeners[:len(listeners)-1]
				this.listeners[eventType] = listeners
			}
		}

		if listener.handler != nil {
			reflect.ValueOf(listener.handler).Call(eventIn)
		}
	}
}
