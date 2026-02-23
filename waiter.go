package playwright

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type (
	waiter struct {
		mu        sync.Mutex
		timeout   float64
		fulfilled atomic.Bool
		listeners []eventListener
		errChan   chan error
		waitFunc  func() (any, error)
	}
	eventListener struct {
		emitter EventEmitter
		event   string
		handler any
	}
)

// RejectOnEvent sets the Waiter to return an error when an event occurs (and the predicate returns true)
func (w *waiter) RejectOnEvent(emitter EventEmitter, event string, err error, predicates ...any) *waiter {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.waitFunc != nil {
		w.reject(fmt.Errorf("waiter: call RejectOnEvent before WaitForEvent"))
		return w
	}
	handler := func(ev ...any) {
		if w.fulfilled.Load() {
			return
		}
		if len(predicates) == 0 {
			w.reject(err)
		} else {
			result := reflect.ValueOf(predicates[0]).Call([]reflect.Value{reflect.ValueOf(ev[0])})
			if result[0].Bool() {
				w.reject(err)
			}
		}
	}
	emitter.On(event, handler)
	w.listeners = append(w.listeners, eventListener{
		emitter: emitter,
		event:   event,
		handler: handler,
	})
	return w
}

// WithTimeout sets timeout, in milliseconds, for the waiter. 0 means no timeout.
func (w *waiter) WithTimeout(timeout float64) *waiter {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.waitFunc != nil {
		w.reject(fmt.Errorf("waiter: please set timeout before WaitForEvent"))
		return w
	}
	w.timeout = timeout
	return w
}

// WaitForEvent sets the Waiter to return when an event occurs (and the predicate returns true)
func (w *waiter) WaitForEvent(emitter EventEmitter, event string, predicate any) *waiter {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.waitFunc != nil {
		w.reject(fmt.Errorf("waiter: WaitForEvent can only be called once"))
		return w
	}
	evChan := make(chan any, 1)
	handler := w.createHandler(evChan, predicate)
	ctx, cancel := context.WithCancel(context.Background())
	if w.timeout != 0 {
		timeout := w.timeout
		go func() {
			select {
			case <-time.After(time.Duration(timeout) * time.Millisecond):
				err := fmt.Errorf("%w:Timeout %.2fms exceeded.", ErrTimeout, timeout)
				w.reject(err)
				return
			case <-ctx.Done():
				return
			}
		}()
	}

	emitter.On(event, handler)
	w.listeners = append(w.listeners, eventListener{
		emitter: emitter,
		event:   event,
		handler: handler,
	})

	w.waitFunc = func() (any, error) {
		var (
			err error
			val any
		)
		select {
		case err = <-w.errChan:
			break
		case val = <-evChan:
			break
		}
		cancel()
		w.mu.Lock()
		defer w.mu.Unlock()
		for _, l := range w.listeners {
			l.emitter.RemoveListener(l.event, l.handler)
		}
		close(evChan)
		if err != nil {
			return nil, err
		}
		return val, nil
	}
	return w
}

// Wait waits for the waiter to return. It needs to call WaitForEvent once first.
func (w *waiter) Wait() (any, error) {
	if w.waitFunc == nil {
		return nil, fmt.Errorf("waiter: call WaitForEvent first")
	}
	return w.waitFunc()
}

// RunAndWait waits for the waiter to return after calls func.
func (w *waiter) RunAndWait(cb func() error) (any, error) {
	if w.waitFunc == nil {
		return nil, fmt.Errorf("waiter: call WaitForEvent first")
	}
	if cb != nil {
		if err := cb(); err != nil {
			w.errChan <- err
		}
	}
	return w.waitFunc()
}

func (w *waiter) createHandler(evChan chan<- any, predicate any) func(...any) {
	return func(ev ...any) {
		if w.fulfilled.Load() {
			return
		}
		if predicate == nil || reflect.ValueOf(predicate).IsNil() {
			w.fulfilled.Store(true)
			if len(ev) == 1 {
				evChan <- ev[0]
			} else {
				evChan <- nil
			}
		} else {
			result := reflect.ValueOf(predicate).Call([]reflect.Value{reflect.ValueOf(ev[0])})
			if result[0].Bool() {
				w.fulfilled.Store(true)
				evChan <- ev[0]
			}
		}
	}
}

func (w *waiter) reject(err error) {
	w.fulfilled.Store(true)
	w.errChan <- err
}

func newWaiter() *waiter {
	w := &waiter{
		// receive both event timeout err and callback err
		// but just return event timeout err
		errChan: make(chan error, 2),
	}
	return w
}
