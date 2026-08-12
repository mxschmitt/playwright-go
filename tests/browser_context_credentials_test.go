package playwright_test

import (
	"path/filepath"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserContextExposesCredentialsProperty(t *testing.T) {
	BeforeEach(t)

	require.NotNil(t, context.Credentials())
	// The same instance is returned on each access.
	require.Same(t, context.Credentials(), context.Credentials())
}

func TestBrowserContextInstallCreateGetDeleteCredentials(t *testing.T) {
	BeforeEach(t)

	// WebAuthn requires a secure context; the test server's localhost origin qualifies.
	_, err := page.Goto(server.CROSS_PROCESS_PREFIX+"/empty.html", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	require.NoError(t, err)

	creds := context.Credentials()
	require.NoError(t, creds.Install())

	created, err := creds.Create("localhost")
	require.NoError(t, err)
	require.Equal(t, "localhost", created.RpId)
	require.NotEmpty(t, created.Id)

	list, err := creds.Get()
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, created.Id, list[0].Id)

	require.NoError(t, creds.Delete(created.Id))
	list, err = creds.Get()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestStorageStateRoundTripWebAuthnCredentials(t *testing.T) {
	BeforeEach(t)
	require.NoError(t, context.AddCookies([]playwright.OptionalCookie{{
		Name:  "session",
		Value: "cookie-value",
		URL:   playwright.String(server.PREFIX),
	}}))
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	_, err = page.Evaluate(`() => localStorage.setItem("roll-key", "roll-value")`)
	require.NoError(t, err)

	// Seed a virtual credential.
	require.NoError(t, context.Credentials().Install())
	cred, err := context.Credentials().Create("example.com")
	require.NoError(t, err)
	require.NotEmpty(t, cred.Id)
	withoutCredentials, err := context.StorageState()
	require.NoError(t, err)
	require.Empty(t, withoutCredentials.Credentials, "credentials must be opt-in")
	require.NotEmpty(t, withoutCredentials.Cookies)
	require.NotEmpty(t, withoutCredentials.Origins)

	state, err := context.StorageState(playwright.BrowserContextStorageStateOptions{
		Credentials: playwright.Bool(true),
	})
	require.NoError(t, err)
	require.NotEmpty(t, state.Credentials)
	require.Equal(t, cred.Id, state.Credentials[0].Id)

	// In-memory round-trip via OptionalStorageState.
	opt := state.ToOptionalStorageState()
	require.NotEmpty(t, opt.Credentials)

	ctx2, err := browser.NewContext(playwright.BrowserNewContextOptions{
		StorageState: opt,
	})
	require.NoError(t, err)
	defer ctx2.Close() //nolint:errcheck
	restored, err := ctx2.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored)
	require.Equal(t, cred.Id, restored[0].Id)

	// Path round-trip.
	path := filepath.Join(t.TempDir(), "state.json")
	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{
		Credentials: playwright.Bool(true),
		Path:        playwright.String(path),
	})
	require.NoError(t, err)
	ctx3, err := browser.NewContext(playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(path),
	})
	require.NoError(t, err)
	defer ctx3.Close() //nolint:errcheck
	restored2, err := ctx3.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored2)

	// SetStorageState must restore and subsequently clear credentials.
	ctx4, err := browser.NewContext()
	require.NoError(t, err)
	defer ctx4.Close() //nolint:errcheck
	require.NoError(t, ctx4.SetStorageState(path))
	restored3, err := ctx4.Credentials().Get()
	require.NoError(t, err)
	require.NotEmpty(t, restored3)
	withoutCredentialsPath := filepath.Join(t.TempDir(), "state-without-credentials.json")
	_, err = context.StorageState(playwright.BrowserContextStorageStateOptions{
		Path: playwright.String(withoutCredentialsPath),
	})
	require.NoError(t, err)
	require.NoError(t, ctx4.SetStorageState(withoutCredentialsPath))
	restored3, err = ctx4.Credentials().Get()
	require.NoError(t, err)
	require.Empty(t, restored3)

	// APIRequestContext strips credentials while preserving cookies/origins.
	req, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageState: state,
	})
	require.NoError(t, err)
	reqState, err := req.StorageState()
	require.NoError(t, err)
	require.Empty(t, reqState.Credentials)
	require.Equal(t, state.Cookies, reqState.Cookies)
	require.Equal(t, state.Origins, reqState.Origins)
	require.NoError(t, req.Dispose())

	reqFromFile, err := pw.Request.NewContext(playwright.APIRequestNewContextOptions{
		StorageStatePath: playwright.String(path),
	})
	require.NoError(t, err)
	reqFileState, err := reqFromFile.StorageState()
	require.NoError(t, err)
	require.Empty(t, reqFileState.Credentials)
	require.Equal(t, state.Cookies, reqFileState.Cookies)
	require.Equal(t, state.Origins, reqFileState.Origins)
	require.NoError(t, reqFromFile.Dispose())
}
