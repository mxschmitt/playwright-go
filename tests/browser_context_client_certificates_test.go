package playwright_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// gotoSettled navigates to url, retrying while a concurrent navigation keeps
// interrupting it. Since v1.61 a client-cert handshake abort triggers an async
// navigation to the error page (and an automatic reload of the target), which
// races with the immediately-following navigation to the matching origin.
func gotoSettled(t *testing.T, p playwright.Page, url string) {
	t.Helper()
	var err error
	for i := 0; i < 20; i++ {
		_, err = p.Goto(url)
		if err == nil {
			return
		}
		if !strings.Contains(err.Error(), "is interrupted by another navigation") {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, err)
}

// requireClientCertHandshakeError asserts that err is the TLS-handshake abort a
// browser reports when it presents no client certificate to a server that
// requires one (v1.61+ scopes the cert to its matching origin). The wording
// varies by browser and platform, so match any of the known variants.
func requireClientCertHandshakeError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	msg := err.Error()
	variants := []string{
		"net::ERR_BAD_SSL_CLIENT_AUTH_CERT", // chromium
		"SSL_ERROR_UNKNOWN",                 // firefox
		"Certificate is required",           // webkit (linux/macos)
		"Failure when receiving data",       // webkit (windows)
	}
	for _, v := range variants {
		if strings.Contains(msg, v) {
			return
		}
	}
	require.Failf(t, "unexpected client-cert handshake error",
		"error %q did not match any known TLS-handshake abort variant", msg)
}

func NewTLSServerRequireClientCert(t *testing.T) *httptest.Server {
	t.Helper()
	certPath := Asset("client-certificates/server/server_cert.pem")
	keyPath := Asset("client-certificates/server/server_key.pem")
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	require.NoError(t, err)
	ca, err := os.ReadFile(Asset("client-certificates/server/server_cert.pem"))
	require.NoError(t, err)
	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(ca)
	require.True(t, ok)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := []byte(fmt.Sprintf(`<div data-testid="message">Hello %s, your certificate was issued by %s!</div>`,
			r.TLS.PeerCertificates[0].Subject.CommonName, r.TLS.PeerCertificates[0].Issuer.CommonName))
		_, err := w.Write(body)
		require.NoError(t, err)
	}))
	// ts.EnableHTTP2 = true
	ts.TLS = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert, // Uses the go standard client certificate verification method
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
	}
	ts.StartTLS()
	return ts
}

func TestClientCerts(t *testing.T) {
	if isWebKit && runtime.GOOS == "darwin" {
		t.Skip("WebKit does not proxy localhost on macOS")
	}
	tlsServer := NewTLSServerRequireClientCert(t)
	defer tlsServer.Close()

	t.Run("should throw with untrusted client certs", func(t *testing.T) {
		BeforeEach(t)

		request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
			ClientCertificates: []playwright.ClientCertificate{
				{
					Origin:   tlsServer.URL,
					CertPath: playwright.String(Asset("client-certificates/client/self-signed/cert.pem")),
					KeyPath:  playwright.String(Asset("client-certificates/client/self-signed/key.pem")),
				},
			},
		})
		require.NoError(t, err)
		_, err = request.Get(tlsServer.URL)
		require.Error(t, err)
		require.Regexp(t, `alert (unknown ca|bad certificate)`, err.Error()) // go v1.19-1.20 fails with "bad certificate"

		require.NoError(t, request.Dispose())
	})

	t.Run("should work with new context", func(t *testing.T) {
		BeforeEach(t, playwright.BrowserNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
			ClientCertificates: []playwright.ClientCertificate{
				{
					Origin:   tlsServer.URL,
					CertPath: playwright.String(Asset("client-certificates/client/trusted/cert.pem")),
					KeyPath:  playwright.String(Asset("client-certificates/client/trusted/key.pem")),
				},
			},
		})

		// Since v1.61 the client-certificate interceptor only presents the cert for
		// the matching origin (upstream socksClientCertificatesInterceptor rewrite).
		// Navigating to the mismatched "localhost" origin therefore sends no cert and
		// the RequireAndVerifyClientCert server aborts the TLS handshake.
		_, err := page.Goto(strings.Replace(tlsServer.URL, "127.0.0.1", "localhost", 1))
		if tlsServer.EnableHTTP2 {
			require.ErrorContains(t, err, "net::ERR_CONNECTION_CLOSED")
		} else {
			requireClientCertHandshakeError(t, err)
		}

		gotoSettled(t, page, tlsServer.URL)
		content, err := page.GetByTestId("message").TextContent()
		require.NoError(t, err)
		require.Equal(t, "Hello Alice, your certificate was issued by localhost!", content)
	})

	t.Run("should work with new persistent context", func(t *testing.T) {
		BeforeEach(t)

		context2, err := browserType.LaunchPersistentContext(
			t.TempDir(),
			playwright.BrowserTypeLaunchPersistentContextOptions{
				IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
				ClientCertificates: []playwright.ClientCertificate{
					{
						Origin:   tlsServer.URL,
						CertPath: playwright.String(Asset("client-certificates/client/trusted/cert.pem")),
						KeyPath:  playwright.String(Asset("client-certificates/client/trusted/key.pem")),
					},
				},
			},
		)
		require.NoError(t, err)
		defer context2.Close() //nolint:errcheck
		page2, err := context2.NewPage()
		require.NoError(t, err)

		// Since v1.61 a mismatched origin sends no cert; the server aborts the handshake.
		_, err = page2.Goto(strings.Replace(tlsServer.URL, "127.0.0.1", "localhost", 1))
		if tlsServer.EnableHTTP2 {
			require.ErrorContains(t, err, "net::ERR_CONNECTION_CLOSED")
		} else {
			requireClientCertHandshakeError(t, err)
		}

		gotoSettled(t, page2, tlsServer.URL)
		content, err := page2.GetByTestId("message").TextContent()
		require.NoError(t, err)
		require.Equal(t, "Hello Alice, your certificate was issued by localhost!", content)
	})

	t.Run("should work with global apirequestcontext", func(t *testing.T) {
		BeforeEach(t)

		request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
			ClientCertificates: []playwright.ClientCertificate{
				{
					Origin:   tlsServer.URL,
					CertPath: playwright.String(Asset("client-certificates/client/trusted/cert.pem")),
					KeyPath:  playwright.String(Asset("client-certificates/client/trusted/key.pem")),
				},
			},
		})
		require.NoError(t, err)
		resp, err := request.Get(tlsServer.URL)
		require.NoError(t, err)
		require.True(t, resp.Ok())

		body, err := resp.Body()
		require.NoError(t, err)
		require.Contains(t, string(body), "Hello Alice, your certificate was issued by localhost!")
		require.NoError(t, request.Dispose())
	})

	t.Run("should work with pfx", func(t *testing.T) {
		BeforeEach(t)

		request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
			ClientCertificates: []playwright.ClientCertificate{
				{
					Origin:     tlsServer.URL,
					PfxPath:    playwright.String(Asset("client-certificates/client/trusted/cert.pfx")),
					Passphrase: playwright.String("secure"),
				},
			},
		})
		require.NoError(t, err)
		resp, err := request.Get(tlsServer.URL)
		require.NoError(t, err)
		require.Equal(t, 200, resp.Status())
		body, err := resp.Body()
		require.NoError(t, err)
		require.Contains(t, string(body), "Hello Alice, your certificate was issued by localhost!")
		require.NoError(t, request.Dispose())
	})

	t.Run("should pass with matching certificates when passing as content", func(t *testing.T) {
		certContent, err := os.ReadFile(Asset("client-certificates/client/trusted/cert.pem"))
		require.NoError(t, err)
		keyContent, err := os.ReadFile(Asset("client-certificates/client/trusted/key.pem"))
		require.NoError(t, err)

		BeforeEach(t, playwright.BrowserNewContextOptions{
			IgnoreHttpsErrors: playwright.Bool(true), // TODO: Remove this once we can pass a custom CA.
			ClientCertificates: []playwright.ClientCertificate{
				{
					Origin: tlsServer.URL,
					Cert:   certContent,
					Key:    keyContent,
				},
			},
		})

		_, err = page.Goto(tlsServer.URL)
		require.NoError(t, err)
		content, err := page.GetByTestId("message").TextContent()
		require.NoError(t, err)
		require.Equal(t, "Hello Alice, your certificate was issued by localhost!", content)
	})
}
