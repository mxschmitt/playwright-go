package playwright

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/coder/websocket"
)

// wsMaxMessageBytes mirrors the Playwright server's own limit: file payloads
// travel base64-encoded inside a single protocol message.
const wsMaxMessageBytes = 256 * 1024 * 1024

// serverVersionPattern extracts the server's major.minor from the body of its
// HTTP 428 version-mismatch refusal.
var serverVersionPattern = regexp.MustCompile(`server version: v(\d+\.\d+)`)

// wsTransport speaks the Playwright protocol directly to a remote Playwright
// server (BrowserType.launchServer / `playwright run-server`) over a WebSocket.
// Unlike pipeTransport there is no local Node.js driver: every protocol message
// is one JSON text frame.
type wsTransport struct {
	conn      *websocket.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
}

// remoteUserAgent is how the Playwright server learns which protocol version
// this client speaks. It refuses a client whose major.minor after
// "Playwright/" differs from its own, but skips the check entirely for a
// User-Agent it cannot parse, so this value is never caller-overridable.
func remoteUserAgent() string {
	return fmt.Sprintf("Playwright/%s (%s; %s) go/%s", playwrightCliVersion, runtime.GOARCH, runtime.GOOS, runtime.Version())
}

// playwrightMajorMinor returns the "1.62" part of a "1.62.1" version.
func playwrightMajorMinor(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

func dialWSTransport(ctx context.Context, endpoint string, headers map[string]string) (*wsTransport, error) {
	header := http.Header{}
	for k, v := range headers {
		if http.CanonicalHeaderKey(k) == "User-Agent" {
			continue
		}
		header.Set(k, v)
	}
	header.Set("User-Agent", remoteUserAgent())
	conn, resp, err := websocket.Dial(ctx, endpoint, &websocket.DialOptions{
		HTTPHeader:      header,
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		var body []byte
		if resp != nil && resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
		}
		if resp != nil && resp.StatusCode == http.StatusPreconditionRequired {
			return nil, versionMismatchError(endpoint, body)
		}
		if len(body) > 0 {
			return nil, fmt.Errorf("could not connect to %s: %w\n%s", endpoint, err, body)
		}
		return nil, fmt.Errorf("could not connect to %s: %w", endpoint, err)
	}
	conn.SetReadLimit(wsMaxMessageBytes)
	tctx, cancel := context.WithCancel(context.Background())
	return &wsTransport{conn: conn, ctx: tctx, cancel: cancel}, nil
}

func versionMismatchError(endpoint string, body []byte) error {
	client := playwrightMajorMinor(playwrightCliVersion)
	match := serverVersionPattern.FindSubmatch(body)
	if match == nil {
		return fmt.Errorf("%w: the server at %s refused this client, which speaks Playwright %s\n%s", ErrVersionMismatch, endpoint, client, body)
	}
	server := string(match[1])
	return fmt.Errorf("%w: the server at %s runs Playwright %s but this client speaks Playwright %s; use the playwright-go release built for Playwright %s (tag v0.%sxx.x)",
		ErrVersionMismatch, endpoint, server, client, server, strings.SplitN(server, ".", 2)[1])
}

func (t *wsTransport) Send(msg map[string]any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("wsTransport: could not marshal json: %w", err)
	}
	if os.Getenv("DEBUGP") != "" {
		fmt.Fprintf(os.Stdout, "\x1b[32mSEND>\x1b[0m\n%s\n", data) //nolint:errcheck
	}
	return t.conn.Write(t.ctx, websocket.MessageText, data)
}

func (t *wsTransport) Poll() (*message, error) {
	_, data, err := t.conn.Read(t.ctx)
	if err != nil {
		return nil, fmt.Errorf("websocket closed: %w", err)
	}
	if os.Getenv("DEBUGP") != "" {
		fmt.Fprintf(os.Stdout, "\x1b[33mRECV>\x1b[0m\n%s\n", data) //nolint:errcheck
	}
	msg := &message{}
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("could not decode json: %w", err)
	}
	return msg, nil
}

func (t *wsTransport) Close() error {
	var err error
	t.closeOnce.Do(func() {
		err = t.conn.Close(websocket.StatusNormalClosure, "")
		t.cancel()
		var ce websocket.CloseError
		if errors.As(err, &ce) || errors.Is(err, context.Canceled) {
			err = nil
		}
	})
	return err
}
