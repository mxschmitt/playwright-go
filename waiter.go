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
			return
		}
		matches, predicateErr := callPredicate(predicates[0], ev)
		if predicateErr != nil {
			w.reject(predicateErr)
			return
		}
		if matches {
			w.reject(err)
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
		if isNilPredicate(predicate) {
			w.fulfilled.Store(true)
			if len(ev) == 1 {
				evChan <- ev[0]
			} else {
				evChan <- nil
			}
			return
		}
		matches, err := callPredicate(predicate, ev)
		if err != nil {
			w.reject(err)
			return
		}
		if matches {
			w.fulfilled.Store(true)
			evChan <- ev[0]
		}
	}
}

// isNilPredicate reports whether no predicate was supplied. reflect.Value.IsNil
// panics on kinds that cannot be nil, so the kind is checked before asking.
func isNilPredicate(predicate any) bool {
	if predicate == nil {
		return true
	}
	v := reflect.ValueOf(predicate)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	default:
		return false
	}
}

// callPredicate invokes a user supplied event predicate. Predicates for the
// generic ExpectEvent/WaitForEvent APIs arrive as `any` because the event
// argument type differs per event, so the signature has to be validated here
// rather than by the compiler. Upstream catches a throwing predicate and
// rejects the wait (client/waiter.ts), so a mismatch is reported as an error
// for the caller instead of panicking inside the event dispatch goroutine.
func callPredicate(predicate any, ev []any) (matches bool, err error) {
	v := reflect.ValueOf(predicate)
	if v.Kind() != reflect.Func {
		return false, fmt.Errorf("waiter: predicate must be a function, but got %T", predicate)
	}
	t := v.Type()
	if t.NumIn() != 1 || t.NumOut() != 1 || t.Out(0).Kind() != reflect.Bool {
		return false, fmt.Errorf("waiter: predicate must be a func(event) bool, but got %T", predicate)
	}
	// Both func(T) bool and the variadic func(...T) bool are used as predicates;
	// reflect.Call packs the single event into the variadic slice for us.
	wantType := t.In(0)
	if t.IsVariadic() {
		wantType = wantType.Elem()
	}
	if len(ev) == 0 {
		return false, fmt.Errorf("waiter: predicate %T got no event argument", predicate)
	}
	arg := reflect.ValueOf(ev[0])
	if !arg.IsValid() {
		// The event carried an untyped nil; pass the zero value of the
		// parameter, mirroring upstream calling the predicate with undefined.
		arg = reflect.Zero(wantType)
	} else if !arg.Type().AssignableTo(wantType) {
		return false, fmt.Errorf("waiter: predicate %T cannot be called with an event of type %T", predicate, ev[0])
	}
	defer func() {
		// A panic raised by the predicate body must not tear down the event
		// dispatch goroutine either; surface it to the caller like upstream.
		if r := recover(); r != nil {
			err = fmt.Errorf("waiter: predicate panicked: %v", r)
		}
	}()
	return v.Call([]reflect.Value{arg})[0].Bool(), nil
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
