package playwright

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testEventName    = "foobar"
	testEventNameFoo = "foo"
	testEventNameBar = "bar"
)

func TestEventEmitterListenerCount(t *testing.T) {
	handler := &eventEmitter{}
	wasCalled := make(chan any, 1)
	myHandler := func(payload ...any) {
		wasCalled <- payload[0]
	}
	require.Nil(t, handler.events[testEventNameFoo])
	handler.On(testEventNameFoo, myHandler)
	require.Equal(t, 1, handler.ListenerCount(testEventNameFoo))
	handler.Once(testEventNameFoo, myHandler)
	require.Equal(t, 2, handler.ListenerCount(testEventNameFoo))
	require.Nil(t, handler.events[testEventNameBar])
	handler.Once(testEventNameBar, myHandler)
	require.Equal(t, 1, handler.ListenerCount(testEventNameBar))
	require.Equal(t, 3, handler.ListenerCount(""))
}

func TestEventEmitterOn(t *testing.T) {
	handler := &eventEmitter{}
	wasCalled := make(chan any, 1)
	require.Nil(t, handler.events[testEventName])
	handler.On(testEventName, func(payload ...any) {
		wasCalled <- payload[0]
	})
	require.Equal(t, 1, handler.ListenerCount(testEventName))
	value := 123
	handler.Emit(testEventName, value)
	result := <-wasCalled
	require.Equal(t, 1, handler.ListenerCount(testEventName))
	require.Equal(t, result.(int), value)
}

func TestEventEmitterOnce(t *testing.T) {
	handler := &eventEmitter{}
	wasCalled := make(chan any, 1)
	require.Nil(t, handler.events[testEventName])
	handler.Once(testEventName, func(payload ...any) {
		wasCalled <- payload[0]
	})
	require.Equal(t, 1, handler.ListenerCount(testEventName))
	value := 123
	handler.Emit(testEventName, value)
	result := <-wasCalled
	require.Equal(t, result.(int), value)
	require.Equal(t, 0, handler.ListenerCount(testEventName))
}

func TestEventEmitterRemove(t *testing.T) {
	handler := &eventEmitter{}
	wasCalled := make(chan any, 1)
	require.Nil(t, handler.events[testEventName])
	myHandler := func(payload ...any) {
		wasCalled <- payload[0]
	}
	handler.On(testEventName, myHandler)
	require.Equal(t, 1, handler.ListenerCount(testEventName))
	value := 123
	handler.Emit(testEventName, value)
	result := <-wasCalled
	require.Equal(t, 1, handler.ListenerCount(testEventName))
	require.Equal(t, result.(int), value)
	handler.Once(testEventName, myHandler)
	handler.RemoveListener(testEventName, myHandler)
	require.Equal(t, 0, handler.ListenerCount(testEventName))
}

func TestEventEmitterRemoveEmpty(t *testing.T) {
	handler := &eventEmitter{}
	handler.RemoveListener(testEventName, func(...any) {})
	require.Equal(t, 0, handler.ListenerCount(testEventName))
}

func TestEventEmitterRemoveKeepExisting(t *testing.T) {
	handler := &eventEmitter{}
	handler.On(testEventName, func(...any) {})
	handler.Once(testEventName, func(...any) {})
	handler.RemoveListener("abc123", func(...any) {})
	handler.RemoveListener(testEventName, func(...any) {})
	require.Equal(t, 2, handler.ListenerCount(testEventName))
}

func TestEventEmitterOnLessArgsAcceptingReceiver(t *testing.T) {
	handler := &eventEmitter{}
	wasCalled := make(chan bool, 1)
	require.Nil(t, handler.events[testEventName])
	handler.Once(testEventName, func(ev ...any) {
		wasCalled <- true
	})
	handler.Emit(testEventName)
	<-wasCalled
}

// TestEventEmitterHandlerCanReenter verifies that a handler may call back into
// the same event's registry (here, registering another listener) while it is
// running. This only works if Emit does not hold the per-event lock across
// handler execution; otherwise the re-entrant call would deadlock.
func TestEventEmitterHandlerCanReenter(t *testing.T) {
	handler := &eventEmitter{}
	done := make(chan bool, 1)
	handler.On(testEventName, func(...any) {
		// Re-enter the emitter from within the handler.
		handler.On(testEventNameFoo, func(...any) {})
		_ = handler.ListenerCount(testEventName)
		done <- true
	})

	go handler.Emit(testEventName)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler that re-enters the emitter deadlocked: lock held across handler execution")
	}
	require.Equal(t, 1, handler.ListenerCount(testEventNameFoo))
}

// TestEventEmitterRemoveDistinctClosures guards against a regression where
// RemoveListener compared handlers by their code pointer
// (reflect.Value.Pointer()), which is shared by every closure created from the
// same function literal. Removing one such closure then removed all its
// siblings, silently unsubscribing unrelated listeners. This reproduced as
// concurrent ExpectResponse waiters timing out once any one of them completed
// (playwright-go issue #323).
//
// It uses the real waiter.createHandler, whose closures are exactly the ones
// that collided under the old comparison.
func TestEventEmitterRemoveDistinctClosures(t *testing.T) {
	handler := &eventEmitter{}
	matchAll := func(any) bool { return true }

	var waiters []*waiter
	for i := 0; i < 3; i++ {
		w := newWaiter()
		h := w.createHandler(make(chan any, 1), matchAll)
		handler.On(testEventName, h)
		w.listeners = append(w.listeners, eventListener{emitter: handler, event: testEventName, handler: h})
		waiters = append(waiters, w)
	}
	require.Equal(t, 3, handler.ListenerCount(testEventName))

	// Removing the first waiter's handler must leave the other two registered.
	handler.RemoveListener(testEventName, waiters[0].listeners[0].handler)
	require.Equal(t, 2, handler.ListenerCount(testEventName))
}

// TestEventEmitterOnceRunsExactlyOnce verifies a one-shot listener fires once
// even though removal now happens before handler invocation.
func TestEventEmitterOnceRunsExactlyOnce(t *testing.T) {
	handler := &eventEmitter{}
	var calls int32
	handler.Once(testEventName, func(...any) {
		atomic.AddInt32(&calls, 1)
	})
	handler.Emit(testEventName)
	handler.Emit(testEventName)
	require.Equal(t, int32(1), atomic.LoadInt32(&calls))
	require.Equal(t, 0, handler.ListenerCount(testEventName))
}
