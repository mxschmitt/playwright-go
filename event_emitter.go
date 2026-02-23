package playwright

import (
	"math"
	"reflect"
	"slices"
	"sync"
)

type EventEmitter interface {
	Emit(name string, payload ...any) bool
	ListenerCount(name string) int
	On(name string, handler any)
	Once(name string, handler any)
	RemoveListener(name string, handler any)
	RemoveListeners(name string)
}

type (
	eventEmitter struct {
		eventsMutex sync.Mutex
		events      map[string]*eventRegister
		hasInit     bool
	}
	eventRegister struct {
		sync.Mutex
		listeners []listener
	}
	listener struct {
		handler any
		once    bool
	}
)

func NewEventEmitter() EventEmitter {
	return &eventEmitter{}
}

func (e *eventEmitter) Emit(name string, payload ...any) (hasListener bool) {
	e.eventsMutex.Lock()
	e.init()

	evt, ok := e.events[name]
	if !ok {
		e.eventsMutex.Unlock()
		return
	}
	e.eventsMutex.Unlock()
	return evt.callHandlers(payload...) > 0
}

func (e *eventEmitter) Once(name string, handler any) {
	e.addEvent(name, handler, true)
}

func (e *eventEmitter) On(name string, handler any) {
	e.addEvent(name, handler, false)
}

func (e *eventEmitter) RemoveListener(name string, handler any) {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	e.init()

	if evt, ok := e.events[name]; ok {
		evt.Lock()
		defer evt.Unlock()
		evt.removeHandler(handler)
	}
}

func (e *eventEmitter) RemoveListeners(name string) {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	e.init()
	delete(e.events, name)
}

// ListenerCount count the listeners by name, count all if name is empty
func (e *eventEmitter) ListenerCount(name string) int {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	e.init()

	if name != "" {
		evt, ok := e.events[name]
		if !ok {
			return 0
		}
		return evt.count()
	}

	count := 0
	for key := range e.events {
		count += e.events[key].count()
	}

	return count
}

func (e *eventEmitter) addEvent(name string, handler any, once bool) {
	e.eventsMutex.Lock()
	defer e.eventsMutex.Unlock()
	e.init()

	if _, ok := e.events[name]; !ok {
		e.events[name] = &eventRegister{
			listeners: make([]listener, 0),
		}
	}
	e.events[name].addHandler(handler, once)
}

func (e *eventEmitter) init() {
	if !e.hasInit {
		e.events = make(map[string]*eventRegister, 0)
		e.hasInit = true
	}
}

func (er *eventRegister) addHandler(handler any, once bool) {
	er.Lock()
	defer er.Unlock()
	er.listeners = append(er.listeners, listener{handler: handler, once: once})
}

func (er *eventRegister) count() int {
	er.Lock()
	defer er.Unlock()
	return len(er.listeners)
}

func (er *eventRegister) removeHandler(handler any) {
	handlerPtr := reflect.ValueOf(handler).Pointer()

	er.listeners = slices.DeleteFunc(er.listeners, func(l listener) bool {
		return reflect.ValueOf(l.handler).Pointer() == handlerPtr
	})
}

func (er *eventRegister) callHandlers(payloads ...any) int {
	payloadV := make([]reflect.Value, 0)

	for _, p := range payloads {
		payloadV = append(payloadV, reflect.ValueOf(p))
	}

	handle := func(l listener) {
		handlerV := reflect.ValueOf(l.handler)
		handlerV.Call(payloadV[:int(math.Min(float64(handlerV.Type().NumIn()), float64(len(payloadV))))])
	}

	er.Lock()
	defer er.Unlock()
	count := len(er.listeners)
	for _, l := range er.listeners {
		if l.once {
			defer er.removeHandler(l.handler)
		}
		handle(l)
	}
	return count
}
