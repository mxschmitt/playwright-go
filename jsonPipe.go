package playwright

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type jsonPipe struct {
	channelOwner
	// mu guards the queue and closed flag.
	mu sync.Mutex
	// queue holds messages received from the outer connection that have not yet
	// been consumed by Poll(). It is unbounded on purpose: the outer connection's
	// dispatch goroutine must never block while delivering a message here,
	// otherwise it cannot deliver the reply that an in-flight Send() is waiting
	// for, which deadlocks the whole connection (see the streaming upload path).
	queue  []*message
	closed bool
	// signal wakes up a Poll() that is waiting for a message or for close.
	signal chan struct{}
}

func (j *jsonPipe) Send(message map[string]any) error {
	_, err := j.channel.Send("send", map[string]any{
		"message": message,
	})
	return err
}

func (j *jsonPipe) Close() error {
	_, err := j.channel.Send("close")
	return err
}

func (j *jsonPipe) Poll() (*message, error) {
	for {
		j.mu.Lock()
		if len(j.queue) > 0 {
			msg := j.queue[0]
			j.queue = j.queue[1:]
			j.mu.Unlock()
			return msg, nil
		}
		if j.closed {
			j.mu.Unlock()
			return nil, errors.New("jsonPipe closed")
		}
		j.mu.Unlock()
		<-j.signal
	}
}

// enqueue appends a message and wakes a waiting Poll(). It never blocks, so it
// is safe to call from the outer connection's dispatch goroutine.
func (j *jsonPipe) enqueue(msg *message) {
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		return
	}
	j.queue = append(j.queue, msg)
	j.mu.Unlock()
	j.wake()
}

func (j *jsonPipe) markClosed() {
	j.mu.Lock()
	if j.closed {
		j.mu.Unlock()
		return
	}
	j.closed = true
	j.mu.Unlock()
	j.wake()
}

func (j *jsonPipe) wake() {
	select {
	case j.signal <- struct{}{}:
	default:
	}
}

func newJsonPipe(parent *channelOwner, objectType string, guid string, initializer map[string]any) *jsonPipe {
	j := &jsonPipe{
		signal: make(chan struct{}, 1),
	}
	j.createChannelOwner(j, parent, objectType, guid, initializer)
	j.channel.On("message", func(ev map[string]any) {
		var msg message
		m, err := json.Marshal(ev["message"])
		if err == nil {
			err = json.Unmarshal(m, &msg)
		}
		if err != nil {
			msg = message{
				Error: &struct {
					Error Error "json:\"error\""
				}{
					Error: Error{
						Name:    "Error",
						Message: fmt.Sprintf("jsonPipe: could not decode message: %s", err.Error()),
					},
				},
			}
		}
		// Enqueue without blocking the dispatch goroutine, while preserving
		// message ordering. A bounded channel here could fill up and stall the
		// dispatch goroutine, deadlocking any in-flight Send() awaiting a reply.
		j.enqueue(&msg)
	})
	j.channel.Once("closed", func() {
		j.Emit("closed")
		j.markClosed()
	})
	return j
}
