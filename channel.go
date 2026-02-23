package playwright

import (
	"encoding/json"
	"fmt"
)

type channel struct {
	eventEmitter
	guid       string
	connection *connection
	owner      *channelOwner // to avoid type conversion
	object     any           // retain type info (for fromChannel needed)
}

func (c *channel) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"guid": c.guid,
	})
}

// for catch errors of route handlers etc.
func (c *channel) CreateTask(fn func()) {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				err, ok := e.(error)
				if ok {
					c.connection.err.Set(err)
				} else {
					c.connection.err.Set(fmt.Errorf("%v", e))
				}
			}
		}()
		fn()
	}()
}

func (c *channel) Send(method string, options ...any) (any, error) {
	return c.connection.WrapAPICall(func() (any, error) {
		result, err := c.innerSend(method, options...).GetResultValue()
		if err != nil {
			return nil, err
		}
		// GUIDs are now always eagerly resolved in connection.Dispatch
		return result, nil
	}, c.owner.isInternalType)
}

func (c *channel) SendReturnAsDict(method string, options ...any) (map[string]any, error) {
	ret, err := c.connection.WrapAPICall(func() (any, error) {
		result, err := c.innerSend(method, options...).GetResult()
		if err != nil {
			return nil, err
		}
		// GUIDs are now always eagerly resolved in connection.Dispatch
		return result, nil
	}, c.owner.isInternalType)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return make(map[string]any), nil
	}
	return ret.(map[string]any), nil
}

func (c *channel) innerSend(method string, options ...any) *protocolCallback {
	if err := c.connection.err.Get(); err != nil {
		c.connection.err.Set(nil)
		pc := newProtocolCallback(false, c.connection.abort)
		pc.SetError(err)
		return pc
	}
	params := transformOptions(options...)
	return c.connection.sendMessageToServer(c.owner, method, params, false)
}

// SendNoReply ignores return value and errors
// almost equivalent to `send(...).catch(() => {})`
func (c *channel) SendNoReply(method string, options ...any) {
	c.innerSendNoReply(method, c.owner.isInternalType, options...)
}

func (c *channel) SendNoReplyInternal(method string, options ...any) {
	c.innerSendNoReply(method, true, options...)
}

func (c *channel) innerSendNoReply(method string, isInternal bool, options ...any) {
	params := transformOptions(options...)
	_, err := c.connection.WrapAPICall(func() (any, error) {
		return c.connection.sendMessageToServer(c.owner, method, params, true).GetResult()
	}, isInternal)
	if err != nil {
		// ignore error actively, log only for debug
		logger.Error("SendNoReply failed", "error", err)
	}
}

func newChannel(owner *channelOwner, object any) *channel {
	channel := &channel{
		connection: owner.connection,
		guid:       owner.guid,
		owner:      owner,
		object:     object,
	}
	return channel
}
