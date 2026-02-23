package playwright

type cdpSessionImpl struct {
	channelOwner
}

func (c *cdpSessionImpl) Detach() error {
	_, err := c.channel.Send("detach")
	return err
}

func (c *cdpSessionImpl) Send(method string, params map[string]any) (any, error) {
	result, err := c.channel.Send("send", map[string]any{
		"method": method,
		"params": params,
	})
	if err != nil {
		return nil, err
	}

	return result, err
}

func (c *cdpSessionImpl) onEvent(params map[string]any) {
	c.Emit(params["method"].(string), params["params"])
}

func newCDPSession(parent *channelOwner, objectType string, guid string, initializer map[string]any) *cdpSessionImpl {
	bt := &cdpSessionImpl{}

	bt.createChannelOwner(bt, parent, objectType, guid, initializer)

	bt.channel.On("event", func(params map[string]any) {
		bt.onEvent(params)
	})

	return bt
}
