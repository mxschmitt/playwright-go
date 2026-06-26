package playwright_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertSlicesEqual(t *testing.T, expected []interface{}, cb func() (interface{}, error)) {
	t.Helper()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		actual, err := cb()
		require.NoError(t, err)
		require.ElementsMatch(t, expected, actual)
	}, 5*time.Second, 200*time.Millisecond)
}

func setupWS(t *testing.T, target playwright.Page, port string, protocol string) {
	t.Helper()

	_, err := target.Goto("about:blank")
	require.NoError(t, err)
	_, err = target.Evaluate(`
	({ port, binaryType }) => {
    window.log = [];
    window.ws = new WebSocket('ws://localhost:' + port + '/ws');
    window.ws.binaryType = binaryType;
    window.ws.addEventListener('open', () => window.log.push('open'));
    window.ws.addEventListener('close', event => window.log.push(`+"`close code=${event.code} reason=${event.reason} wasClean=${event.wasClean}`"+`));
    window.ws.addEventListener('error', event => window.log.push(`+"`error`"+`));
    window.ws.addEventListener('message', async event => {
      let data;
      console.log(event);
      if (typeof event.data === 'string')
        data = event.data;
      else if (event.data instanceof Blob)
        data = 'blob:' + await event.data.text();
      else
        data = 'arraybuffer:' + await (new Blob([event.data])).text();
      window.log.push(`+"`message: data=${data} origin=${event.origin} lastEventId=${event.lastEventId}`"+`);
    });
    window.wsOpened = new Promise(f => window.ws.addEventListener('open', () => f()));
  }`, map[string]interface{}{"port": port, "binaryType": protocol})
	require.NoError(t, err)
}

func TestShouldWorkWithWSClose(t *testing.T) {
	BeforeEach(t)

	wsRouteChan := make(chan playwright.WebSocketRoute, 1)

	handleWS := func(ws playwright.WebSocketRoute) {
		_, _ = ws.ConnectToServer()
		wsRouteChan <- ws
	}

	require.NoError(t, page.RouteWebSocket(regexp.MustCompile(".*"), handleWS))

	wsConnChan := server.WaitForWebSocketConnection()
	setupWS(t, page, server.PORT, "blob")
	<-wsConnChan

	wsRoute := <-wsRouteChan
	wsRoute.Send("hello")

	assertSlicesEqual(t, []interface{}{"open", "message: data=hello origin=ws://localhost:" + server.PORT + " lastEventId="}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	closedError := make(chan *websocket.CloseError, 1)
	server.OnceWebSocketClose(func(err *websocket.CloseError) {
		closedError <- err
	})

	wsRoute.Close(playwright.WebSocketRouteCloseOptions{
		Code:   playwright.Int(3009),
		Reason: playwright.String("oops"),
	})

	assertSlicesEqual(t,
		[]interface{}{
			"open",
			"message: data=hello origin=ws://localhost:" + server.PORT + " lastEventId=",
			"close code=3009 reason=oops wasClean=true",
		},
		func() (interface{}, error) {
			return page.Evaluate(`window.log`)
		})

	result := <-closedError
	require.Equal(t, "3009, oops", fmt.Sprintf("%d, %s", result.Code, result.Reason))
}

func TestShouldPatterMatch(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.RouteWebSocket(regexp.MustCompile(".*/ws$"), func(ws playwright.WebSocketRoute) {
		go func() {
			_, _ = ws.ConnectToServer()
		}()
	}))
	require.NoError(t, page.RouteWebSocket("**/mock-ws", func(ws playwright.WebSocketRoute) {
		ws.OnMessage(func(i interface{}) {
			ws.Send("mock-response")
		})
	}))

	wsConnChan := server.WaitForWebSocketConnection()
	_, err := page.Goto("about:blank")
	require.NoError(t, err)

	_, err = page.Evaluate(`
  async ({ port }) => {
        window.log = [];
        window.ws1 = new WebSocket('ws://localhost:' + port + '/ws');
        window.ws1.addEventListener('message', event => window.log.push(`+"`ws1:${event.data}`"+`));
        window.ws2 = new WebSocket('ws://localhost:' + port + '/something/something/mock-ws');
        window.ws2.addEventListener('message', event => window.log.push(`+"`ws2:${event.data}`"+`));
        await Promise.all([
            new Promise(f => window.ws1.addEventListener('open', f)),
            new Promise(f => window.ws2.addEventListener('open', f)),
        ]);
    }
	`, map[string]interface{}{"port": server.PORT})
	require.NoError(t, err)

	<-wsConnChan
	server.OnWebSocketMessage(func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte) {
		err := c.Write(r.Context(), websocket.MessageText, []byte("response"))
		t.Log(err)
	})

	_, err = page.Evaluate(`window.ws1.send('request')`)
	require.NoError(t, err)
	assertSlicesEqual(t, []interface{}{"ws1:response"}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	_, err = page.Evaluate(`window.ws2.send('request')`)
	require.NoError(t, err)
	assertSlicesEqual(t, []interface{}{"ws1:response", "ws2:mock-response"}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})
}

func TestRouteWebSocketShouldWorkWithServer(t *testing.T) {
	BeforeEach(t)

	wsRouteChan := make(chan playwright.WebSocketRoute, 1)

	handleWS := func(ws playwright.WebSocketRoute) {
		server, err := ws.ConnectToServer()
		require.NoError(t, err)

		ws.OnMessage(func(message interface{}) {
			msg := message.(string)
			switch msg {
			case "to-respond":
				ws.Send("response")
				return
			case "to-block":
				return
			case "to-modify":
				server.Send("modified")
				return
			default:
				server.Send(message)
			}
		})

		server.OnMessage(func(message interface{}) {
			msg := message.(string)
			switch msg {
			case "to-block":
				return
			case "to-modify":
				ws.Send("modified")
				return
			default:
				ws.Send(message)
			}
		})

		server.Send("fake")
		wsRouteChan <- ws
	}

	log := newSyncSlice[string]()

	server.OnWebSocketMessage(func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte) {
		log.Append(fmt.Sprintf("message: %s", msg))
	})
	server.OnWebSocketClose(func(err *websocket.CloseError) {
		log.Append(fmt.Sprintf("close: code=%d reason=%s", err.Code, err.Reason))
	})

	require.NoError(t, page.RouteWebSocket(regexp.MustCompile(".*"), handleWS))

	wsConnChan := server.WaitForWebSocketConnection()

	setupWS(t, page, server.PORT, "blob")
	ws := <-wsConnChan
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		require.ElementsMatch(t, []string{"message: fake"}, log.Get())
	}, 5*time.Second, 200*time.Millisecond)

	ws.SendMessage(websocket.MessageText, []byte("to-modify"))
	ws.SendMessage(websocket.MessageText, []byte("to-block"))
	ws.SendMessage(websocket.MessageText, []byte("pass-server"))

	assertSlicesEqual(t, []interface{}{
		"open",
		"message: data=modified origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=pass-server origin=ws://localhost:" + server.PORT + " lastEventId=",
	}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	_, err := page.Evaluate(`
	() => {
			window.ws.send('to-respond');
			window.ws.send('to-modify');
			window.ws.send('to-block');
			window.ws.send('pass-client');
	}`)
	require.NoError(t, err)

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		require.ElementsMatch(t, []string{"message: fake", "message: modified", "message: pass-client"}, log.Get())
	}, 5*time.Second, 200*time.Millisecond)

	assertSlicesEqual(t, []interface{}{
		"open",
		"message: data=modified origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=pass-server origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
	}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	route := <-wsRouteChan
	route.Send("another")
	assertSlicesEqual(t, []interface{}{
		"open",
		"message: data=modified origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=pass-server origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=another origin=ws://localhost:" + server.PORT + " lastEventId=",
	}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	_, err = page.Evaluate(`
	() => {
			window.ws.send('pass-client-2');
	}`)
	require.NoError(t, err)

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		require.ElementsMatch(t, []string{"message: fake", "message: modified", "message: pass-client", "message: pass-client-2"}, log.Get())
	}, 5*time.Second, 200*time.Millisecond)

	_, err = page.Evaluate(`
	() => {
			window.ws.close(3009, 'problem');
	}`)
	require.NoError(t, err)

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		require.ElementsMatch(t, []string{
			"message: fake",
			"message: modified",
			"message: pass-client",
			"message: pass-client-2",
			"close: code=3009 reason=problem",
		}, log.Get())
	}, 5*time.Second, 200*time.Millisecond)
}

func TestRouteWebSocketShouldWorkWithoutServer(t *testing.T) {
	BeforeEach(t)

	wsRouteChan := make(chan playwright.WebSocketRoute, 1)

	handleWS := func(ws playwright.WebSocketRoute) {
		ws.OnMessage(func(message interface{}) {
			if message.(string) == "to-respond" {
				ws.Send("response")
			}
		})

		wsRouteChan <- ws
	}

	require.NoError(t, page.RouteWebSocket(regexp.MustCompile(".*"), handleWS))
	setupWS(t, page, server.PORT, "blob")

	_, err := page.Evaluate(`
	async () => {
			await window.wsOpened;
			window.ws.send('to-respond');
			window.ws.send('to-block');
			window.ws.send('to-respond');
	}`)
	require.NoError(t, err)

	assertSlicesEqual(t, []interface{}{
		"open",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
	}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})

	route := <-wsRouteChan
	route.Send("another")
	// wait for the message to be processed
	time.Sleep(100 * time.Millisecond)
	route.Close(playwright.WebSocketRouteCloseOptions{
		Code:   playwright.Int(3008),
		Reason: playwright.String("oops"),
	})
	assertSlicesEqual(t, []interface{}{
		"open",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=response origin=ws://localhost:" + server.PORT + " lastEventId=",
		"message: data=another origin=ws://localhost:" + server.PORT + " lastEventId=",
		"close code=3008 reason=oops wasClean=true",
	}, func() (interface{}, error) {
		return page.Evaluate(`window.log`)
	})
}
