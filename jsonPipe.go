package playwright

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type jsonPipe struct {
	channelOwner
	// mu guards queue and closed; cond (built on mu) lets Poll wait for either a
	// new message or close without busy-waiting.
	mu     sync.Mutex
	cond   *sync.Cond
	queue  []*message
	closed bool
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
	j.mu.Lock()
	defer j.mu.Unlock()
	for len(j.queue) == 0 && !j.closed {
		j.cond.Wait()
	}
	if len(j.queue) > 0 {
		msg := j.queue[0]
		j.queue = j.queue[1:]
		return msg, nil
	}
	return nil, errors.New("jsonPipe closed")
}

// enqueue appends a message and wakes any waiting Poll(). It never blocks (the
// queue is unbounded), so it is safe to call from the outer connection's
// dispatch goroutine: that goroutine must never block delivering a message
// here, otherwise it cannot deliver the reply an in-flight Send() is waiting
// for, which would deadlock the whole connection (see the streaming upload
// path). Ordering is preserved since the single dispatch goroutine is the only
// appender.
func (j *jsonPipe) enqueue(msg *message) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return
	}
	j.queue = append(j.queue, msg)
	j.cond.Broadcast()
}

func (j *jsonPipe) markClosed() {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.closed {
		return
	}
	j.closed = true
	j.cond.Broadcast()
}

func newJsonPipe(parent *channelOwner, objectType string, guid string, initializer map[string]any) *jsonPipe {
	j := &jsonPipe{}
	j.cond = sync.NewCond(&j.mu)
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
		j.enqueue(&msg)
	})
	j.channel.Once("closed", func() {
		j.Emit("closed")
		j.markClosed()
	})
	return j
}
