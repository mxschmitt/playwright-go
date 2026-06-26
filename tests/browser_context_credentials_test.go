package playwright_test

import (
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
