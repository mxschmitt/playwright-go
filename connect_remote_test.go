package playwright

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// serverVersionMismatchBody is the 428 body a Playwright 1.63 server sends to a
// 1.62 client, captured from a real server.
const serverVersionMismatchBody = `╔════════════════════════════════════════════════════╗
║ Playwright version mismatch:                       ║
║   - server version: v1.63                          ║
║   - client version: v1.62                          ║
║                                                    ║
║ If you are using VSCode extension, restart VSCode. ║
║                                                    ║
║ If you are connecting to a remote service,         ║
║ keep your local Playwright version in sync         ║
║ with the remote service version.                   ║
║                                                    ║
║ <3 Playwright Team                                 ║
╚════════════════════════════════════════════════════╝`

// TestRemoteUserAgentPassesServerVersionCheck mirrors the server's
// userAgentVersionMatchesErrorMessage: it only checks a User-Agent matching
// this pattern, comparing its major.minor with the server's own.
func TestRemoteUserAgentPassesServerVersionCheck(t *testing.T) {
	match := regexp.MustCompile(`^Playwright/(\d+\.\d+\.\d+)`).FindStringSubmatch(remoteUserAgent())
	require.NotNil(t, match, "the server skips its version check for a User-Agent it cannot parse")
	require.Equal(t, playwrightCliVersion, match[1])
	require.Equal(t, playwrightMajorMinor(playwrightCliVersion), strings.Join(strings.Split(match[1], ".")[:2], "."))
}

func refusingServer(t *testing.T, status int, body string) (*httptest.Server, <-chan http.Header) {
	t.Helper()
	received := make(chan http.Header, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case received <- r.Header.Clone():
		default:
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server, received
}

func TestConnectRemoteSendsVersionAndCallerHeaders(t *testing.T) {
	server, received := refusingServer(t, http.StatusPreconditionRequired, serverVersionMismatchBody)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ConnectRemote(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), RemoteConnectOptions{
		Browser: "firefox",
		Headers: map[string]string{
			"Authorization": "Bearer token",
			"user-agent":    "Playwright/1.50.0 (spoofed)",
		},
	})
	require.Error(t, err)
	headers := <-received
	require.Equal(t, remoteUserAgent(), headers.Get("User-Agent"), "callers must not be able to override the version check")
	require.Equal(t, "Bearer token", headers.Get("Authorization"))
	require.Equal(t, "firefox", headers.Get("X-Playwright-Browser"))
}

func TestConnectRemoteVersionMismatch(t *testing.T) {
	server, _ := refusingServer(t, http.StatusPreconditionRequired, serverVersionMismatchBody)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ConnectRemote(ctx, "ws"+strings.TrimPrefix(server.URL, "http"))
	require.ErrorIs(t, err, ErrVersionMismatch)
	require.Contains(t, err.Error(), "runs Playwright 1.63 but this client speaks Playwright "+playwrightMajorMinor(playwrightCliVersion))
	require.Contains(t, err.Error(), "tag v0.63xx.x")
}

func TestConnectRemoteVersionMismatchUnparseableBody(t *testing.T) {
	server, _ := refusingServer(t, http.StatusPreconditionRequired, "precondition failed")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ConnectRemote(ctx, "ws"+strings.TrimPrefix(server.URL, "http"))
	require.ErrorIs(t, err, ErrVersionMismatch)
	require.Contains(t, err.Error(), "precondition failed")
}

func TestConnectRemoteOtherHandshakeFailure(t *testing.T) {
	server, _ := refusingServer(t, http.StatusBadGateway, "Unable to connect to browser")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ConnectRemote(ctx, "ws"+strings.TrimPrefix(server.URL, "http"))
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrVersionMismatch))
	require.Contains(t, err.Error(), "Unable to connect to browser")
}
