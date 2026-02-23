package playwright

import (
	"encoding/base64"
	"errors"
)

type webSocketImpl struct {
	channelOwner
	isClosed bool
	page     *pageImpl
}

func (ws *webSocketImpl) URL() string {
	return ws.initializer["url"].(string)
}

func newWebsocket(parent *channelOwner, objectType string, guid string, initializer map[string]any) *webSocketImpl {
	ws := &webSocketImpl{}
	ws.createChannelOwner(ws, parent, objectType, guid, initializer)
	ws.page = fromChannel(parent.channel).(*pageImpl)
	ws.channel.On("close", func() {
		ws.Lock()
		ws.isClosed = true
		ws.Unlock()
		ws.Emit("close", ws)
	})
	ws.channel.On(
		"frameSent",
		func(params map[string]any) {
			ws.onFrameSent(params["opcode"].(float64), params["data"].(string))
		},
	)
	ws.channel.On(
		"frameReceived",
		func(params map[string]any) {
			ws.onFrameReceived(params["opcode"].(float64), params["data"].(string))
		},
	)
	ws.channel.On(
		"socketError",
		func(params map[string]any) {
			ws.Emit("socketerror", params["error"])
		},
	)
	return ws
}

func (ws *webSocketImpl) onFrameSent(opcode float64, data string) {
	if opcode == 2 {
		payload, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			logger.Error("could not decode WebSocket.onFrameSent payload", "error", err)
			return
		}
		ws.Emit("framesent", payload)
	} else {
		ws.Emit("framesent", []byte(data))
	}
}

func (ws *webSocketImpl) onFrameReceived(opcode float64, data string) {
	if opcode == 2 {
		payload, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			logger.Error("could not decode WebSocket.onFrameReceived payload", "error", err)
			return
		}
		ws.Emit("framereceived", payload)
	} else {
		ws.Emit("framereceived", []byte(data))
	}
}

func (ws *webSocketImpl) ExpectEvent(event string, cb func() error, options ...WebSocketExpectEventOptions) (any, error) {
	return ws.expectEvent(event, cb, options...)
}

func (ws *webSocketImpl) WaitForEvent(event string, options ...WebSocketWaitForEventOptions) (any, error) {
	if len(options) == 1 {
		option := WebSocketExpectEventOptions(options[0])
		return ws.expectEvent(event, nil, option)
	} else {
		return ws.expectEvent(event, nil)
	}
}

func (ws *webSocketImpl) expectEvent(event string, cb func() error, options ...WebSocketExpectEventOptions) (any, error) {
	var predicate any = nil
	timeout := ws.page.timeoutSettings.Timeout()
	if len(options) == 1 {
		if options[0].Timeout != nil {
			timeout = *options[0].Timeout
		}
		if options[0].Predicate != nil {
			predicate = options[0].Predicate
		}
	}
	waiter := newWaiter().WithTimeout(timeout)
	if event != "close" {
		waiter.RejectOnEvent(ws, "close", errors.New("websocket closed"))
	}
	if event != "socketerror" {
		waiter.RejectOnEvent(ws, "socketerror", errors.New("websocket error"))
	}
	waiter.RejectOnEvent(ws.page, "close", errors.New("page closed"))
	if cb == nil {
		return waiter.WaitForEvent(ws, event, predicate).Wait()
	} else {
		return waiter.WaitForEvent(ws, event, predicate).RunAndWait(cb)
	}
}

func (ws *webSocketImpl) IsClosed() bool {
	ws.RLock()
	defer ws.RUnlock()
	return ws.isClosed
}

func (ws *webSocketImpl) OnClose(fn func(WebSocket)) {
	ws.On("close", fn)
}

func (ws *webSocketImpl) OnFrameReceived(fn func(payload []byte)) {
	ws.On("framereceived", fn)
}

func (ws *webSocketImpl) OnFrameSent(fn func(payload []byte)) {
	ws.On("framesent", fn)
}

func (ws *webSocketImpl) OnSocketError(fn func(string)) {
	ws.On("socketerror", fn)
}
