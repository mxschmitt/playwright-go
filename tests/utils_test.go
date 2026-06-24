package playwright_test

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/coder/websocket"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

var (
	server *testServer
	utils  *testUtils
)

func init() {
	if err := mime.AddExtensionType(".js", "application/javascript"); err != nil {
		log.Fatalf("could not add mime extension type: %v", err)
	}
}

func Asset(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get cwd: %v", err)
	}
	return filepath.Join(cwd, "assets", path)
}

func newTestServer(tls ...bool) *testServer {
	ts := &testServer{
		eventEmitter:        playwright.NewEventEmitter(),
		mux:                 http.NewServeMux(),
		routes:              make(map[string]http.HandlerFunc),
		requestSubscriberes: make(map[string][]chan *http.Request),
	}
	ts.mux.HandleFunc("/ws", ts.wsHandler)
	ts.mux.HandleFunc("/", ts.handler)
	if len(tls) > 0 && tls[0] {
		ts.testServer = httptest.NewTLSServer(ts.mux)
	} else {
		ts.testServer = httptest.NewServer(ts.mux)
	}
	ts.PREFIX = ts.testServer.URL
	ts.EMPTY_PAGE = ts.testServer.URL + "/empty.html"
	ts.CROSS_PROCESS_PREFIX = strings.Replace(ts.testServer.URL, "127.0.0.1", "localhost", 1)
	ts.PORT = ts.testServer.URL[strings.LastIndex(ts.testServer.URL, ":")+1:]
	return ts
}

type testServer struct {
	sync.Mutex
	eventEmitter         playwright.EventEmitter
	mux                  *http.ServeMux
	testServer           *httptest.Server
	routes               map[string]http.HandlerFunc
	requestSubscriberes  map[string][]chan *http.Request
	PREFIX               string
	EMPTY_PAGE           string
	CROSS_PROCESS_PREFIX string
	WS_PREFIX            string
	PORT                 string
}

func (t *testServer) AfterEach() {
	t.Lock()
	defer t.Unlock()
	t.routes = make(map[string]http.HandlerFunc)
	t.requestSubscriberes = make(map[string][]chan *http.Request)
	t.eventEmitter.RemoveListeners("connection")
	t.eventEmitter.RemoveListeners("message")
	t.eventEmitter.RemoveListeners("close")
	t.testServer.CloseClientConnections()
}

func (t *testServer) CloseClientConnections() {
	t.testServer.CloseClientConnections()
}

func (t *testServer) SetTLSConfig(config *tls.Config) {
	t.testServer.TLS = config
}

type wsConnection struct {
	Conn *websocket.Conn
	Req  *http.Request
}

func (c *wsConnection) SendMessage(msgType websocket.MessageType, data []byte) {
	err := c.Conn.Write(c.Req.Context(), msgType, data)
	if err != nil {
		log.Println("testServer: could not write ws message:", err)
		return
	}
}

func (t *testServer) OnceWebSocketConnection(handler func(c *websocket.Conn, r *http.Request)) {
	t.eventEmitter.Once("connection", handler)
}

func (t *testServer) OnWebSocketClose(handler func(err *websocket.CloseError)) {
	t.eventEmitter.On("close", handler)
}

func (t *testServer) OnceWebSocketClose(handler func(err *websocket.CloseError)) {
	t.eventEmitter.Once("close", handler)
}

func (t *testServer) OnWebSocketMessage(handler func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte)) {
	t.eventEmitter.On("message", handler)
}

func (t *testServer) OnceWebSocketMessage(handler func(c *websocket.Conn, r *http.Request, msgType websocket.MessageType, msg []byte)) {
	t.eventEmitter.Once("message", handler)
}

func (t *testServer) SendOnWebSocketConnection(msgType websocket.MessageType, data []byte) {
	t.OnceWebSocketConnection(func(c *websocket.Conn, r *http.Request) {
		err := c.Write(r.Context(), msgType, data)
		if err != nil {
			log.Println("testServer: could not write ws message:", err)
			return
		}
	})
}

func (t *testServer) WaitForWebSocketConnection() <-chan *wsConnection {
	channel := make(chan *wsConnection)
	t.OnceWebSocketConnection(func(c *websocket.Conn, r *http.Request) {
		channel <- &wsConnection{Conn: c, Req: r}
		close(channel)
	})
	return channel
}

func (t *testServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println("testServer: could not upgrade ws connection:", err)
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "") //nolint:errcheck

	t.eventEmitter.Emit("connection", c, r)

	for {
		typ, message, err := c.Read(r.Context())
		if err != nil {
			closeErr := new(websocket.CloseError)
			if errors.As(err, closeErr) {
				t.eventEmitter.Emit("close", closeErr)
			}
			switch websocket.CloseStatus(err) {
			case websocket.StatusNormalClosure, websocket.StatusGoingAway, websocket.StatusNoStatusRcvd:
			default:
				log.Println("testServer: could not read ws message:", err)
			}
			break
		}
		t.eventEmitter.Emit("message", c, r, typ, message)
	}
}

func (t *testServer) handler(w http.ResponseWriter, r *http.Request) {
	t.Lock()
	defer t.Unlock()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("testServer: Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if handlers, ok := t.requestSubscriberes[r.URL.Path]; ok {
		for _, handler := range handlers {
			handler <- r
		}
	}
	if route, ok := t.routes[r.URL.Path]; ok {
		route(w, r)
		return
	}
	w.Header().Add("Cache-Control", "no-cache, no-store")
	http.FileServer(http.Dir("assets")).ServeHTTP(w, r)
}

func (s *testServer) SetBasicAuth(path, username, password string) {
	s.SetRoute(path, func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != username || p != password {
			w.Header().Add("WWW-Authenticate", "Basic") // needed or playwright will do not send auth header
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}
	})
}

func (s *testServer) SetRoute(path string, f http.HandlerFunc) {
	s.Lock()
	defer s.Unlock()
	s.routes[path] = f
}

func (s *testServer) SetRedirect(from, to string) {
	s.SetRoute(from, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusFound)
	})
}

func (s *testServer) WaitForRequestChan(path string) <-chan *http.Request {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.requestSubscriberes[path]; !ok {
		s.requestSubscriberes[path] = make([]chan *http.Request, 0)
	}
	channel := make(chan *http.Request, 1)
	s.requestSubscriberes[path] = append(s.requestSubscriberes[path], channel)
	return channel
}

func Map[T any, R any](vs []T, f func(T) R) []R {
	vsm := make([]R, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// ChanToSlice reads amount of values from the channel, returns them as a slice
func ChanToSlice[T any](ch chan T, amount int) []T {
	data := make([]T, 0)
	for i := 0; i < amount; i++ {
		data = append(data, <-ch)
	}
	return data
}

type syncSlice[T any] struct {
	sync.Mutex
	slice []T
}

func (s *syncSlice[T]) Append(v T) {
	s.Lock()
	defer s.Unlock()
	s.slice = append(s.slice, v)
}

func (s *syncSlice[T]) Get() []T {
	s.Lock()
	defer s.Unlock()
	return s.slice
}

func (s *syncSlice[T]) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.slice)
}

func newSyncSlice[T any]() *syncSlice[T] {
	return &syncSlice[T]{
		slice: make([]T, 0),
	}
}

type testUtils struct{}

func (t *testUtils) AttachFrame(page playwright.Page, frameId string, url string) (playwright.Frame, error) {
	handle, err := page.EvaluateHandle(`async ({ frame_id, url }) => {
		const frame = document.createElement('iframe');
		frame.src = url;
		frame.id = frame_id;
		document.body.appendChild(frame);
		await new Promise(x => frame.onload = x);
		return frame;
	}`, map[string]interface{}{
		"frame_id": frameId,
		"url":      url,
	})
	if err != nil {
		return nil, err
	}
	elem := handle.AsElement()
	if elem == nil {
		return nil, errors.New("frame not found")
	}
	return elem.ContentFrame()
}

func (t *testUtils) DetachFrame(page playwright.Page, frameId string) error {
	_, err := page.Evaluate(`frame_id => document.getElementById(frame_id).remove()`, frameId)
	return err
}

func (tu *testUtils) DumpFrames(frame playwright.Frame, indentation string) []string {
	desc := strings.Replace(frame.URL(), server.PREFIX, "http://localhost:<PORT>", 1)
	if frame.Name() != "" {
		desc = fmt.Sprintf("%s (%s)", desc, frame.Name())
	}
	result := []string{
		indentation + desc,
	}
	sortedFrames := frame.ChildFrames()
	sort.SliceStable(sortedFrames, func(i, j int) bool {
		return (sortedFrames[i].URL() + sortedFrames[i].Name()) < (sortedFrames[j].URL() + sortedFrames[j].Name())
	})
	for _, f := range sortedFrames {
		result = append(result, tu.DumpFrames(f, "    "+indentation)...)
	}
	return result
}

func (tu *testUtils) VerifyViewport(t *testing.T, page playwright.Page, width, height int) {
	require.Equal(t, page.ViewportSize().Width, width)
	require.Equal(t, page.ViewportSize().Height, height)
	innerWidth, err := page.Evaluate("window.innerWidth")
	require.NoError(t, err)
	require.Equal(t, innerWidth, width)
	innerHeight, err := page.Evaluate("window.innerHeight")
	require.NoError(t, err)
	require.Equal(t, innerHeight, height)
}

func (tu *testUtils) AssertEval(t *testing.T, page playwright.Page, script string, expected interface{}) {
	result, err := page.Evaluate(script)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func (tu *testUtils) AssertResult(t *testing.T, fn func() (interface{}, error), expected interface{}) {
	result, err := fn()
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close() //nolint:errcheck
	return l.Addr().(*net.TCPAddr).Port, nil
}

func readFromZip(zipFile string, fileName string) ([]byte, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer r.Close() //nolint:errcheck

	for _, f := range r.File {
		if f.Name == fileName {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close() //nolint:errcheck

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, rc)
			if err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("file %s not found in %s", fileName, zipFile)
}

func getFileLastModifiedTimeMs(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return 0, fmt.Errorf("%s is a directory", path)
	}
	return info.ModTime().UnixMilli(), nil
}

func mustGetHostname(ref string) *string {
	u, err := url.Parse(ref)
	if err != nil {
		return nil
	}
	return playwright.String(u.Hostname())
}

func chromiumVersionLessThan(a, b string) bool {
	left := strings.Split(a, ".")
	right := strings.Split(b, ".")
	for i := 0; i < len(left) && i < len(right); i++ {
		if left[i] != right[i] {
			return left[i] < right[i]
		}
	}
	return len(left) < len(right)
}
