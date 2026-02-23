package playwright

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testEventNameFoobar = "foobar"
	testEventNameReject = "reject"
	testEventPayload    = "payload data"
)

func TestWaiterWaitForEvent(t *testing.T) {
	timeout := 500.0
	emitter := &eventEmitter{}
	waiter := newWaiter().WithTimeout(timeout)
	_, err := waiter.Wait()
	require.Error(t, err)
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	go func() {
		emitter.Emit(testEventNameFoobar, testEventPayload)
		emitter.Emit(testEventNameFoobar, "2")
		emitter.Emit(testEventNameFoobar, "3")
	}()
	result, err := waiter.Wait()
	require.NoError(t, err)
	require.Equal(t, result, testEventPayload)
}

func TestWaiterWaitForEventWithPredicate(t *testing.T) {
	timeout := 500.0
	emitter := &eventEmitter{}
	waiter := newWaiter().WithTimeout(timeout)
	waiter.WaitForEvent(emitter, testEventNameFoobar, func(payload any) bool {
		content, ok := payload.(string)
		if ok && content == testEventPayload {
			return true
		}
		return false
	})
	go func() {
		emitter.Emit(testEventNameFoobar, "1")
		emitter.Emit(testEventNameFoobar, testEventPayload)
		emitter.Emit(testEventNameFoobar, "3")
	}()
	result, err := waiter.Wait()
	require.NoError(t, err)
	require.Equal(t, result, testEventPayload)
}

func TestWaiterRejectOnTimeout(t *testing.T) {
	timeout := 300.0
	emitter := &eventEmitter{}
	waiter := newWaiter().WithTimeout(timeout)
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	go func() {
		time.Sleep(time.Duration(timeout*2) * time.Millisecond)
		emitter.Emit(testEventNameFoobar, testEventPayload)
	}()
	result, err := waiter.Wait()
	require.ErrorContains(t, err, fmt.Sprintf("Timeout %.2fms exceeded.", timeout))
	require.Nil(t, result)
}

func TestWaiterRejectOnEvent(t *testing.T) {
	errCause := fmt.Errorf("reject on event")
	errPredicate := fmt.Errorf("payload on event")
	emitter := &eventEmitter{}
	waiter := newWaiter().RejectOnEvent(emitter, testEventNameReject, errCause)
	waiter.RejectOnEvent(emitter, testEventNameFoobar, errPredicate, func(payload any) bool {
		content, ok := payload.(string)
		if ok && content == "testEventPayload" {
			return true
		}
		return false
	})
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	require.Equal(t, 1, emitter.ListenerCount(testEventNameReject))
	go func() {
		emitter.Emit(testEventNameReject)
		emitter.Emit(testEventNameFoobar, testEventPayload)
		emitter.Emit(testEventNameFoobar, "3")
	}()
	result, err := waiter.Wait()
	require.EqualError(t, err, errCause.Error())
	require.Equal(t, 0, emitter.ListenerCount(testEventNameReject))
	require.Nil(t, result)
}

func TestWaiterRejectOnEventWithPredicate(t *testing.T) {
	errCause := fmt.Errorf("reject on event")
	errPredicate := fmt.Errorf("payload on event")
	emitter := &eventEmitter{}
	waiter := newWaiter().RejectOnEvent(emitter, testEventNameReject, errCause)
	waiter.RejectOnEvent(emitter, testEventNameFoobar, errPredicate, func(payload any) bool {
		content, ok := payload.(string)
		if ok && content == testEventPayload {
			return true
		}
		return false
	})
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	require.Equal(t, 1, emitter.ListenerCount(testEventNameReject))
	go func() {
		emitter.Emit(testEventNameFoobar, testEventPayload)
		emitter.Emit(testEventNameReject)
		emitter.Emit(testEventNameFoobar, "3")
	}()
	result, err := waiter.Wait()
	require.EqualError(t, err, errPredicate.Error())
	require.Equal(t, 0, emitter.ListenerCount(testEventNameReject))
	require.Nil(t, result)
}

func TestWaiterReturnErrorWhenMisuse(t *testing.T) {
	emitter := &eventEmitter{}
	waiter := newWaiter()
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	waiter.WithTimeout(500)
	_, err := waiter.Wait()
	require.ErrorContains(t, err, "please set timeout before WaitForEvent")

	waiter = newWaiter()
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	waiter.WaitForEvent(emitter, testEventNameFoo, nil)
	_, err = waiter.Wait()
	require.ErrorContains(t, err, "WaitForEvent can only be called once")

	waiter = newWaiter()
	waiter.WaitForEvent(emitter, testEventNameFoobar, nil)
	waiter.RejectOnEvent(emitter, testEventNameFoo, nil)
	_, err = waiter.Wait()
	require.ErrorContains(t, err, "call RejectOnEvent before WaitForEvent")
}

func TestWaiterDeadlockForErrChanCapIs1AndCallbackErr(t *testing.T) {
	// deadlock happen on waiter timeout before callback return err
	waiterTimeout := 200.0
	callbackTimeout := time.Duration(waiterTimeout+200.0) * time.Millisecond

	mockCallbackErr := errors.New("mock callback error")

	emitter := &eventEmitter{}
	w := &waiter{
		// just receive event timeout err or callback err
		errChan: make(chan error, 1),
	}

	callbackOverCh := make(chan struct{})
	callbackErrCh := make(chan error)
	isAfterWaiterRunAndWaitExecuted := atomic.Bool{}
	go func() {
		_, err := w.WithTimeout(waiterTimeout).WaitForEvent(emitter, "", nil).RunAndWait(func() error {
			time.Sleep(callbackTimeout)
			close(callbackOverCh)
			// block for this err, for waiter.errChan has cache event timeout err
			return mockCallbackErr
		})

		isAfterWaiterRunAndWaitExecuted.Store(true)
		callbackErrCh <- err
	}()

	// ensure waiter timeout
	<-callbackOverCh
	// give some time but never enough
	time.Sleep(200 * time.Millisecond)

	// Originally it was executed, but because waiter.errChan is currently caching the waiter timeout error,
	// the callback error is blocked (because waitFunc has not been executed yet,
	// waiter.errChan has not started receiving).
	require.False(t, isAfterWaiterRunAndWaitExecuted.Load())

	// if not receive waiter timeout error, isAfterWaiterRunAndWaitExecuted should be always false
	err1 := <-w.errChan
	require.ErrorIs(t, err1, ErrTimeout)

	// for w.errChan cache is empty, callback error is sent and received, and then return it
	err2 := <-callbackErrCh
	require.ErrorIs(t, err2, mockCallbackErr)
	require.True(t, isAfterWaiterRunAndWaitExecuted.Load())
}

func TestWaiterHasNotDeadlockForErrChanCapBiggerThan1AndCallbackErr(t *testing.T) {
	// deadlock happen on waiter timeout before callback return err
	waiterTimeout := 100.0
	callbackTimeout := time.Duration(waiterTimeout+100.0) * time.Millisecond

	mockCallbackErr := errors.New("mock callback error")

	emitter := &eventEmitter{}
	w := newWaiter()

	callbackOverCh := make(chan struct{})
	callbackErrCh := make(chan error)
	isAfterWaiterRunAndWaitExecuted := atomic.Bool{}
	go func() {
		_, err := w.WithTimeout(waiterTimeout).WaitForEvent(emitter, "", nil).RunAndWait(func() error {
			time.Sleep(callbackTimeout)
			close(callbackOverCh)
			return mockCallbackErr
		})
		isAfterWaiterRunAndWaitExecuted.Store(true)
		callbackErrCh <- err
	}()

	// ensure waiter timeout
	<-callbackOverCh

	// for waiter.errChan cap is 2(greater than 1), so it will not block(deadlock)
	require.Eventually(t,
		func() bool { return isAfterWaiterRunAndWaitExecuted.Load() }, 100*time.Millisecond, 10*time.Microsecond)

	// the first err still is waiter timeout, and is returned
	err1 := <-w.errChan
	require.ErrorIs(t, err1, mockCallbackErr)
	err2 := <-callbackErrCh
	require.ErrorIs(t, err2, ErrTimeout)
}
