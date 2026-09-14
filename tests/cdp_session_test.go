package playwright_test

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCDPSessionSend(t *testing.T) {
	BeforeEach(t)

	cdpSession, err := browser.NewBrowserCDPSession()
	if isChromium {
		require.NoError(t, err)
		result, err := cdpSession.Send("Target.getTargets", nil)
		require.NoError(t, err)
		targetInfos := result.(map[string]interface{})["targetInfos"].([]interface{})
		require.GreaterOrEqual(t, len(targetInfos), 1)
	} else {
		require.Error(t, err)
	}
}

func TestCDPSessionOn(t *testing.T) {
	BeforeEach(t)

	cdpSession, err := page.Context().NewCDPSession(page)
	if isChromium {
		require.NoError(t, err)
		_, err = cdpSession.Send("Console.enable", nil)
		require.NoError(t, err)
		cdpSession.On("Console.messageAdded", func(params map[string]interface{}) {
			require.NotNil(t, params)
		})
		_, err = page.Evaluate(`console.log("hello")`)
		require.NoError(t, err)
		require.NoError(t, cdpSession.Detach())
	} else {
		require.Error(t, err)
	}
}

// Chrome's Page.downloadWillBegin / Page.downloadProgress payloads carry a
// download guid that is not a Playwright channel. A CDPSession with Page.enable
// relays those maps as events; they must not tear down the driver connection.
func TestCDPSessionDownloadDoesNotCloseConnection(t *testing.T) {
	BeforeEach(t)

	cdpSession, err := page.Context().NewCDPSession(page)
	if !isChromium {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)

	server.SetRoute("/downloadWithFilename", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment; filename=file.txt")
		if _, err := w.Write([]byte("foobar")); err != nil {
			log.Printf("could not write: %v", err)
		}
	})

	began := make(chan map[string]any, 1)
	cdpSession.On("Page.downloadWillBegin", func(params map[string]any) {
		began <- params
	})
	_, err = cdpSession.Send("Page.enable", nil)
	require.NoError(t, err)
	_, err = cdpSession.Send("Page.setDownloadBehavior", map[string]any{
		"behavior":     "allow",
		"downloadPath": t.TempDir(),
	})
	require.NoError(t, err)

	require.NoError(t, page.SetContent(
		fmt.Sprintf(`<a href="%s/downloadWithFilename">download</a>`, server.PREFIX),
	))
	download, err := page.ExpectDownload(func() error {
		return page.Locator("a").Click()
	})
	require.NoError(t, err)
	require.Equal(t, "file.txt", download.SuggestedFilename())

	select {
	case params := <-began:
		require.Equal(t, "file.txt", params["suggestedFilename"])
		require.NotEmpty(t, params["guid"])
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Page.downloadWillBegin")
	}

	ctx, err := browser.NewContext()
	require.NoError(t, err, "connection must survive an unbound CDP download guid")
	require.NoError(t, ctx.Close())
}

func TestCDPSessionDetach(t *testing.T) {
	BeforeEach(t)

	cdpSession, err := browser.NewBrowserCDPSession()
	if isChromium {
		require.NoError(t, err)
		require.NoError(t, cdpSession.Detach())
	} else {
		require.Error(t, err)
	}
}
