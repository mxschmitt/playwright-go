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

// protocolCallOptions describes transport-level behavior for a single
// protocol call. timeoutAware is intentionally independent from timeout: a
// timeout-aware protocol method must strip a public `params.timeout` even when
// the caller deliberately omits metadata.timeout.
type protocolCallOptions struct {
	timeout      *float64
	timeoutAware bool
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
	return c.send(method, protocolCallOptions{}, options...)
}

// SendWithTimeout sends a protocol method with the call timeout carried in
// metadata.timeout (Playwright ≥1.62). timeout may be a pointer to zero
// (unlimited); a nil timeout omits metadata.timeout. Any "timeout" key present
// in the transformed params is stripped so it is not double-sent as a param.
func (c *channel) SendWithTimeout(method string, timeout *float64, options ...any) (any, error) {
	return c.send(method, protocolCallOptions{timeout: timeout, timeoutAware: true}, options...)
}

func (c *channel) send(method string, callOptions protocolCallOptions, options ...any) (any, error) {
	return c.connection.WrapAPICall(func() (any, error) {
		result, err := c.innerSend(method, callOptions, options...).GetResultValue()
		if err != nil {
			return nil, err
		}
		// GUIDs are now always eagerly resolved in connection.Dispatch
		return result, nil
	}, c.owner.isInternalType)
}

func (c *channel) SendReturnAsDict(method string, options ...any) (map[string]any, error) {
	return c.sendReturnAsDict(method, protocolCallOptions{}, options...)
}

// SendReturnAsDictWithTimeout is the timeout-aware form of SendReturnAsDict.
func (c *channel) SendReturnAsDictWithTimeout(method string, timeout *float64, options ...any) (map[string]any, error) {
	return c.sendReturnAsDict(method, protocolCallOptions{timeout: timeout, timeoutAware: true}, options...)
}

func (c *channel) sendReturnAsDict(method string, callOptions protocolCallOptions, options ...any) (map[string]any, error) {
	ret, err := c.connection.WrapAPICall(func() (any, error) {
		result, err := c.innerSend(method, callOptions, options...).GetResult()
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

func (c *channel) innerSend(method string, callOptions protocolCallOptions, options ...any) *protocolCallback {
	if err := c.connection.err.Get(); err != nil {
		c.connection.err.Set(nil)
		pc := newProtocolCallback(c.connection, false, c.connection.abort)
		pc.SetError(err)
		return pc
	}
	params := transformOptions(options...)
	if callOptions.timeoutAware {
		// Timeout-aware boundary: call timeout travels in metadata, not params.
		delete(params, "timeout")
	}
	return c.connection.sendMessageToServer(c.owner, method, params, false, callOptions.timeout)
}

// SendNoReply ignores return value and errors
// almost equivalent to `send(...).catch(() => {})`
func (c *channel) SendNoReply(method string, options ...any) {
	c.innerSendNoReply(method, c.owner.isInternalType, protocolCallOptions{}, options...)
}

func (c *channel) SendNoReplyInternal(method string, options ...any) {
	c.innerSendNoReply(method, true, protocolCallOptions{}, options...)
}

// SendNoReplyInternalWithTimeout is a fire-and-forget send that still carries
// metadata.timeout when needed (e.g. some internal driver ops).
func (c *channel) SendNoReplyInternalWithTimeout(method string, timeout *float64, options ...any) {
	c.innerSendNoReply(method, true, protocolCallOptions{timeout: timeout, timeoutAware: true}, options...)
}

func (c *channel) innerSendNoReply(method string, isInternal bool, callOptions protocolCallOptions, options ...any) {
	params := transformOptions(options...)
	if callOptions.timeoutAware {
		delete(params, "timeout")
	}
	_, err := c.connection.WrapAPICall(func() (any, error) {
		return c.connection.sendMessageToServer(c.owner, method, params, true, callOptions.timeout).GetResult()
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
