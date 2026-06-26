package playwright

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"sync/atomic"
)

type webSocketRouteImpl struct {
	channelOwner
	connected       *atomic.Bool
	server          WebSocketRoute
	onPageMessage   func(any)
	onPageClose     func(code *int, reason *string)
	onServerMessage func(any)
	onServerClose   func(code *int, reason *string)
}

func newWebSocketRoute(parent *channelOwner, objectType string, guid string, initializer map[string]any) *webSocketRouteImpl {
	route := &webSocketRouteImpl{
		connected: &atomic.Bool{},
	}
	route.createChannelOwner(route, parent, objectType, guid, initializer)
	route.markAsInternalType()

	route.server = newServerWebSocketRoute(route)

	route.channel.On("messageFromPage", func(event map[string]any) {
		msg, err := untransformWebSocketMessage(event)
		if err != nil {
			panic(fmt.Errorf("Could not decode WebSocket message: %w", err))
		}
		if route.onPageMessage != nil {
			route.onPageMessage(msg)
		} else if route.connected.Load() {
			go route.channel.SendNoReply("sendToServer", event)
		}
	})

	route.channel.On("messageFromServer", func(event map[string]any) {
		msg, err := untransformWebSocketMessage(event)
		if err != nil {
			panic(fmt.Errorf("Could not decode WebSocket message: %w", err))
		}
		if route.onServerMessage != nil {
			route.onServerMessage(msg)
		} else {
			go route.channel.SendNoReply("sendToPage", event)
		}
	})

	route.channel.On("closePage", func(event map[string]any) {
		if route.onPageClose != nil {
			var code *int
			if v, ok := event["code"]; ok && v != nil {
				i := int(v.(float64))
				code = &i
			}
			var reason *string
			if v, ok := event["reason"]; ok && v != nil {
				s := v.(string)
				reason = &s
			}
			route.onPageClose(code, reason)
		} else {
			go route.channel.SendNoReply("closeServer", event)
		}
	})

	route.channel.On("closeServer", func(event map[string]any) {
		if route.onServerClose != nil {
			var code *int
			if v, ok := event["code"]; ok && v != nil {
				i := int(v.(float64))
				code = &i
			}
			var reason *string
			if v, ok := event["reason"]; ok && v != nil {
				s := v.(string)
				reason = &s
			}
			route.onServerClose(code, reason)
		} else {
			go route.channel.SendNoReply("closePage", event)
		}
	})

	return route
}

func (r *webSocketRouteImpl) Close(options ...WebSocketRouteCloseOptions) {
	r.channel.SendNoReply("closePage", options, map[string]any{"wasClean": true})
}

func (r *webSocketRouteImpl) ConnectToServer() (WebSocketRoute, error) {
	if r.connected.Load() {
		return nil, fmt.Errorf("Already connected to the server")
	}
	r.connected.Store(true)
	r.channel.SendNoReply("connect")
	return r.server, nil
}

func (r *webSocketRouteImpl) OnClose(handler func(code *int, reason *string)) {
	r.onPageClose = handler
}

func (r *webSocketRouteImpl) OnMessage(handler func(any)) {
	r.onPageMessage = handler
}

func (r *webSocketRouteImpl) Send(message any) {
	data, err := transformWebSocketMessage(message)
	if err != nil {
		panic(fmt.Errorf("Could not encode WebSocket message: %w", err))
	}
	go r.channel.SendNoReply("sendToPage", data)
}

func (r *webSocketRouteImpl) URL() string {
	return r.initializer["url"].(string)
}

func (r *webSocketRouteImpl) Protocols() ([]string, error) {
	protocols := r.initializer["protocols"].([]any)
	result := make([]string, len(protocols))
	for i, p := range protocols {
		result[i] = p.(string)
	}
	return result, nil
}

func (r *webSocketRouteImpl) afterHandle() {
	if r.connected.Load() {
		return
	}
	// Ensure that websocket is "open" and can send messages without an actual server connection.
	// If this happens after the page has been closed, ignore the error (matches upstream).
	_, _ = r.channel.Send("ensureOpened")
}

type serverWebSocketRouteImpl struct {
	webSocketRoute *webSocketRouteImpl
}

func newServerWebSocketRoute(route *webSocketRouteImpl) *serverWebSocketRouteImpl {
	return &serverWebSocketRouteImpl{webSocketRoute: route}
}

func (s *serverWebSocketRouteImpl) OnMessage(handler func(any)) {
	s.webSocketRoute.onServerMessage = handler
}

func (s *serverWebSocketRouteImpl) OnClose(handler func(code *int, reason *string)) {
	s.webSocketRoute.onServerClose = handler
}

func (s *serverWebSocketRouteImpl) ConnectToServer() (WebSocketRoute, error) {
	return nil, fmt.Errorf("ConnectToServer must be called on the page-side WebSocketRoute")
}

func (s *serverWebSocketRouteImpl) URL() string {
	return s.webSocketRoute.URL()
}

func (s *serverWebSocketRouteImpl) Protocols() ([]string, error) {
	return s.webSocketRoute.Protocols()
}

func (s *serverWebSocketRouteImpl) Close(options ...WebSocketRouteCloseOptions) {
	go s.webSocketRoute.channel.SendNoReply("closeServer", options, map[string]any{"wasClean": true})
}

func (s *serverWebSocketRouteImpl) Send(message any) {
	data, err := transformWebSocketMessage(message)
	if err != nil {
		panic(fmt.Errorf("Could not encode WebSocket message: %w", err))
	}
	go s.webSocketRoute.channel.SendNoReply("sendToServer", data)
}

func transformWebSocketMessage(message any) (map[string]any, error) {
	data := map[string]any{}
	switch v := message.(type) {
	case []byte:
		data["isBase64"] = true
		data["message"] = base64.StdEncoding.EncodeToString(v)
	case string:
		data["isBase64"] = false
		data["message"] = v
	default:
		return nil, fmt.Errorf("Unsupported message type: %T", v)
	}
	return data, nil
}

func untransformWebSocketMessage(data map[string]any) (any, error) {
	if data["isBase64"].(bool) {
		return base64.StdEncoding.DecodeString(data["message"].(string))
	}
	return data["message"], nil
}

type webSocketRouteHandler struct {
	matcher *urlMatcher
	handler func(WebSocketRoute)
}

func newWebSocketRouteHandler(matcher *urlMatcher, handler func(WebSocketRoute)) *webSocketRouteHandler {
	return &webSocketRouteHandler{matcher: matcher, handler: handler}
}

func (h *webSocketRouteHandler) Handle(route WebSocketRoute) {
	h.handler(route)
	route.(*webSocketRouteImpl).afterHandle()
}

func (h *webSocketRouteHandler) Matches(wsURL string) bool {
	return h.matcher.Matches(wsURL)
}

func prepareWebSocketRouteHandlerInterceptionPatterns(handlers []*webSocketRouteHandler) []map[string]any {
	patterns := []map[string]any{}
	all := false
	for _, handler := range handlers {
		switch handler.matcher.raw.(type) {
		case *regexp.Regexp:
			pattern, flags := convertRegexp(handler.matcher.raw.(*regexp.Regexp))
			patterns = append(patterns, map[string]any{
				"regexSource": pattern,
				"regexFlags":  flags,
			})
		case string:
			patterns = append(patterns, map[string]any{
				"glob": handler.matcher.raw.(string),
			})
		default:
			all = true
		}
	}
	if all {
		return []map[string]any{
			{
				"glob": "**/*",
			},
		}
	}
	return patterns
}
