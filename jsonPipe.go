package playwright

import (
	"encoding/json"
	"errors"
	"fmt"
)

type jsonPipe struct {
	channelOwner
	msgChan chan *message
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
	msg := <-j.msgChan
	if msg == nil {
		return nil, errors.New("jsonPipe closed")
	}
	return msg, nil
}

func newJsonPipe(parent *channelOwner, objectType string, guid string, initializer map[string]any) *jsonPipe {
	j := &jsonPipe{
		msgChan: make(chan *message, 10),
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
		// Send directly to maintain message ordering - the channel buffer prevents blocking
		// Previously used a goroutine which could cause out-of-order delivery
		defer func() {
			// Recover from panic if channel is closed
			_ = recover()
		}()
		j.msgChan <- &msg
	})
	j.channel.Once("closed", func() {
		j.Emit("closed")
		close(j.msgChan)
	})
	return j
}
