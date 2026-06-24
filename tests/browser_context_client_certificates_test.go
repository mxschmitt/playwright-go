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

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

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

		resp, err := page.Goto(strings.Replace(tlsServer.URL, "127.0.0.1", "localhost", 1))
		if tlsServer.EnableHTTP2 {
			require.ErrorContains(t, err, "net::ERR_CONNECTION_CLOSED")
		} else {
			require.NoError(t, err)
			require.False(t, resp.Ok()) // status code 503, client didn't provide a certificate due to origin mismatch
		}

		_, err = page.Goto(tlsServer.URL)
		require.NoError(t, err)
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

		resp, err := page2.Goto(strings.Replace(tlsServer.URL, "127.0.0.1", "localhost", 1))
		if tlsServer.EnableHTTP2 {
			require.ErrorContains(t, err, "net::ERR_CONNECTION_CLOSED")
		} else {
			require.NoError(t, err)
			require.False(t, resp.Ok()) // status code 503, client didn't provide a certificate due to origin mismatch
		}

		_, err = page2.Goto(tlsServer.URL)
		require.NoError(t, err)
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
