package playwright_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestWebSocketShouldWork(t *testing.T) {
	BeforeEach(t)

	server.SendOnWebSocketConnection(websocket.MessageText, []byte("incoming"))

	value, err := page.Evaluate(`port => {
        let cb;
        const result = new Promise(f => cb = f);
        const ws = new WebSocket('ws://localhost:' + port + '/ws');
        ws.addEventListener('message', data => { ws.close(); cb(data.data); });
        return result;
	}`, server.PORT)
	require.NoError(t, err)
	require.Equal(t, "incoming", value)
}

func TestWebSocketShouldEmitCloseEvents(t *testing.T) {
	BeforeEach(t)

	server.SendOnWebSocketConnection(websocket.MessageText, []byte("incoming"))

	ws, err := page.ExpectWebSocket(func() error {
		_, err := page.Evaluate(`port => {
					const ws = new WebSocket('ws://localhost:' + port + '/ws');
					ws.addEventListener('message', data => { ws.close() });
        }`, server.PORT)
		return err
	}, playwright.PageExpectWebSocketOptions{
		Timeout: playwright.Float(1000),
	})
	require.NoError(t, err)
	ws.OnClose(func(w playwright.WebSocket) {
		require.Equal(t, w.URL(), fmt.Sprintf("ws://localhost:%s/ws", server.PORT))
	})
	require.Equal(t, ws.URL(), fmt.Sprintf("ws://localhost:%s/ws", server.PORT))

	require.Eventually(t, func() bool { return ws.IsClosed() }, 1*time.Second, 10*time.Millisecond)
}

func TestWebSocketShouldEmitFrameEvents(t *testing.T) {
	BeforeEach(t)

	server.SendOnWebSocketConnection(websocket.MessageText, []byte("incoming"))
	server.OnceWebSocketMessage(func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte) {
		if bytes.Equal(msg, []byte("echo-text")) {
			_ = c.Write(r.Context(), msgType, []byte("text"))
		}
	})

	sent := [][]byte{}
	received := [][]byte{}

	page.OnWebSocket(func(ws playwright.WebSocket) {
		ws.OnFrameSent(func(payload []byte) {
			sent = append(sent, payload)
		})
		ws.OnFrameReceived(func(payload []byte) {
			received = append(received, payload)
		})
	})
	ws, err := page.ExpectWebSocket(func() error {
		_, err := page.Evaluate(`port => {
						let count = 0;
            const ws = new WebSocket('ws://localhost:' + port + '/ws');
            ws.addEventListener('open', () => {
                ws.send('echo-text');
            });
						ws.addEventListener('message', data => { count++; if (count >= 2) { ws.close() } });
        }`, server.PORT)
		return err
	})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return ws.IsClosed() }, 1*time.Second, 10*time.Millisecond)

	require.Equal(t, sent, [][]byte{[]byte("echo-text")})
	require.Equal(t, received, [][]byte{[]byte("incoming"), []byte("text")})
}

func TestWebSocketShouldEmitBinaryFrameEvents(t *testing.T) {
	BeforeEach(t)

	server.SendOnWebSocketConnection(websocket.MessageText, []byte("incoming"))
	server.OnWebSocketMessage(func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte) {
		if bytes.Equal(msg, []byte("echo-bin")) {
			_ = c.Write(r.Context(), websocket.MessageBinary, []byte{4, 2})
		}
	})

	sent := [][]byte{}
	received := [][]byte{}

	page.Once("websocket", func(ws playwright.WebSocket) {
		ws.OnFrameSent(func(payload []byte) {
			sent = append(sent, payload)
		})
		ws.OnFrameReceived(func(payload []byte) {
			received = append(received, payload)
		})
	})
	ws, err := page.ExpectWebSocket(func() error {
		_, err := page.Evaluate(`port => {
						let count = 0;
            const ws = new WebSocket('ws://localhost:' + port + '/ws');
            ws.addEventListener('open', () => {
                const binary = new Uint8Array(5);
                for (let i = 0; i < 5; ++i)
                    binary[i] = i;
                ws.send(binary);
                ws.send('echo-bin');
            });
						ws.addEventListener('message', data => { count++; if (count >= 2) { ws.close() } });
        }`, server.PORT)
		return err
	})
	require.NoError(t, err)
	require.Eventually(t, func() bool { return ws.IsClosed() }, 1*time.Second, 10*time.Millisecond)

	require.Equal(t, [][]byte{{0, 1, 2, 3, 4}, []byte("echo-bin")}, sent)
	require.Equal(t, [][]byte{[]byte("incoming"), {4, 2}}, received)
}

func TestWebSocketShouldRejectWaitForEventOnCloseAndError(t *testing.T) {
	BeforeEach(t)

	ws, err := page.ExpectWebSocket(func() error {
		_, err := page.Evaluate(`port => {
            ws = new WebSocket('ws://localhost:' + port + '/ws');
        }`, server.PORT)
		return err
	})
	require.NoError(t, err)
	// event may have been generated before interception
	_, _ = ws.WaitForEvent("framereceived", playwright.WebSocketWaitForEventOptions{
		Timeout: playwright.Float(1000),
	})
	_, err = ws.ExpectEvent("framesent", func() error {
		_, err := page.Evaluate(`window.ws.close()`)
		return err
	}, playwright.WebSocketExpectEventOptions{
		Timeout: playwright.Float(1000),
		Predicate: func(ev interface{}) bool {
			return true
		},
	})
	require.ErrorContains(t, err, "websocket closed")
}

func TestWebSocketShouldEmitErrorEvent(t *testing.T) {
	BeforeEach(t)

	chanMsg := make(chan string, 1)
	page.OnWebSocket(func(ws playwright.WebSocket) {
		ws.OnSocketError(func(err string) {
			chanMsg <- err
		})
	})
	_, err := page.Evaluate(`port => {
			ws = new WebSocket('ws://localhost:' + port + '/bogus-ws');
		}`, server.PORT)
	require.NoError(t, err)
	msg := <-chanMsg
	// Since v1.61 Firefox reports the same ": 404" error as the other browsers
	// instead of "CLOSE_ABNORMAL".
	require.Contains(t, msg, ": 404")
}
