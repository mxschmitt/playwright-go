package playwright_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mxschmitt/playwright-go"
)

func newContextWithHttpCredentials(t *testing.T, credentials ...playwright.HttpCredentials) playwright.Page {
	t.Helper()
	_, page := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{
		HttpCredentialsList: credentials,
	})
	return page
}

func TestBrowserContextHttpCredentialsListWithASingleCredential(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user", "pass")
	page := newContextWithHttpCredentials(t, playwright.HttpCredentials{Username: "user", Password: "pass"})
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
}

func TestBrowserContextHttpCredentialsListWithMultipleOrigins(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user1", "pass1")
	page := newContextWithHttpCredentials(t,
		playwright.HttpCredentials{Username: "user1", Password: "pass1", Origin: playwright.String(server.PREFIX)},
		playwright.HttpCredentials{Username: "user2", Password: "pass2", Origin: playwright.String(server.CROSS_PROCESS_PREFIX)},
	)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	// The other origin gets its own, wrong credentials.
	response, err = page.Goto(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
}

func TestBrowserContextHttpCredentialsListFallsBackToCredentialsWithoutOrigin(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user", "pass")
	page := newContextWithHttpCredentials(t,
		playwright.HttpCredentials{Username: "user2", Password: "pass2", Origin: playwright.String(server.CROSS_PROCESS_PREFIX)},
		playwright.HttpCredentials{Username: "user", Password: "pass"},
	)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	// The first matching entry has wrong credentials for this origin.
	response, err = page.Goto(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
}

func TestBrowserContextHttpCredentialsListUsesTheFirstMatchingCredential(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user", "pass")
	page := newContextWithHttpCredentials(t,
		playwright.HttpCredentials{Username: "wrong", Password: "wrong"},
		playwright.HttpCredentials{Username: "user", Password: "pass", Origin: playwright.String(server.PREFIX)},
	)
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
}

func TestBrowserContextHttpCredentialsListTakesPrecedenceOverHttpCredentials(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user", "pass")
	_, page := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{
		HttpCredentials:     &playwright.HttpCredentials{Username: "wrong", Password: "wrong"},
		HttpCredentialsList: []playwright.HttpCredentials{{Username: "user", Password: "pass"}},
	})
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
}

func TestBrowserNewPageHttpCredentialsList(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user", "pass")
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		HttpCredentialsList: []playwright.HttpCredentials{{Username: "user", Password: "pass"}},
	})
	require.NoError(t, err)
	defer page.Close() //nolint:errcheck
	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
}

func TestBrowserContextRequestSupportsMultipleHttpCredentials(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user1", "pass1")
	ctx, _ := newBrowserContextAndPage(t, playwright.BrowserNewContextOptions{
		HttpCredentialsList: []playwright.HttpCredentials{
			{Username: "user1", Password: "pass1", Origin: playwright.String(server.PREFIX)},
			{Username: "user2", Password: "pass2", Origin: playwright.String(server.CROSS_PROCESS_PREFIX)},
		},
	})
	response, err := ctx.Request().Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	// The other origin gets its own, wrong credentials.
	response, err = ctx.Request().Get(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
}

func TestGlobalFetchSupportsMultipleHttpCredentials(t *testing.T) {
	BeforeEach(t)
	server.SetBasicAuth("/empty.html", "user1", "pass1")
	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		HttpCredentialsList: []playwright.HttpCredentials{
			{Username: "user1", Password: "pass1", Origin: playwright.String(server.PREFIX)},
			{Username: "user2", Password: "pass2", Origin: playwright.String(server.CROSS_PROCESS_PREFIX)},
		},
	})
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck
	response, err := request.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	// The other origin gets its own, wrong credentials.
	response, err = request.Get(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.Equal(t, 401, response.Status())
}

func TestGlobalFetchSupportsHttpCredentialsSendWithMultipleHttpCredentials(t *testing.T) {
	BeforeEach(t)
	request, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		HttpCredentialsList: []playwright.HttpCredentials{
			{Username: "user1", Password: "pass1", Origin: playwright.String(server.PREFIX), Send: playwright.HttpCredentialsSendAlways},
			{Username: "user2", Password: "pass2", Origin: playwright.String(server.CROSS_PROCESS_PREFIX), Send: playwright.HttpCredentialsSendUnauthorized},
		},
	})
	require.NoError(t, err)
	defer request.Dispose() //nolint:errcheck

	requests := server.WaitForRequestChan("/empty.html")
	response, err := request.Get(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	require.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("user1:pass1")), (<-requests).Header.Get("Authorization"))

	response, err = request.Get(server.CROSS_PROCESS_PREFIX + "/empty.html")
	require.NoError(t, err)
	require.Equal(t, 200, response.Status())
	// This origin uses send: unauthorized, so credentials are not sent proactively.
	require.Empty(t, (<-requests).Header.Get("Authorization"))
}
